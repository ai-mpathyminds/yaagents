// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tokenvalidator implements the token-validator plugin (PRD §6.5 plugin a).
//
// This is the PLG-3 full implementation (Sprint 2), superseding the PI1-yaa bridge
// that delegated to gateway/internal/auth. It is self-contained per ADR PI2-yaa-0005
// Decision 1 (inline JWKS re-implementation; portfolio/packages/go/auth-jwks/ not
// yet extracted at A-3 verification).
//
// Two validation modes:
//
//	JWKS (production): RS256 JWT validated via a remote JWKS endpoint.
//	  Config: jwks_url, cache_ttl_seconds (default 600).
//	  Stale-while-revalidate: on JWKS refresh failure, stale keys are used
//	  and a warn is emitted — requests are not hard-failed.
//
//	HS256 (dev/test): symmetric HMAC-SHA256.
//	  Config: test_mode: true, jwt_secret (required when test_mode is true).
//
// Always-on: Init returns a non-nil error (gateway exit 1) if enabled: false
// (ADR PI2-yaa-0001 §5 defence-in-depth; the gateway core also asserts this).
//
// Validation failure: 403 application/vnd.yaagents.error+json with trace.correlationId
// populated from reqctx. next is NOT called.
//
// Registration: init() → plugin.Register(&TokenValidator{}) — import side-effect
// wiring per ADR PI2-yaa-0001 §3.
package tokenvalidator

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&TokenValidator{})
}

// TokenValidator is the PLG-3 full token-validator plugin.
// Zero value is invalid; always call Init before Handler.
type TokenValidator struct {
	// HS256 test mode.
	testMode bool
	hs256Key []byte

	// RS256 / JWKS mode.
	jwksVal *jwksValidator

	// Optional audience claim validation (skipped when empty).
	audience string
}

// Name returns the canonical plugin identifier.
func (tv *TokenValidator) Name() string { return "token-validator" }

// Init validates configuration and wires the appropriate validation strategy.
//
// Returns non-nil error (gateway exit 1) when:
//   - enabled is false — always-on plugin cannot be disabled
//   - test_mode is true but jwt_secret is empty
//   - neither jwks_url nor test_mode+jwt_secret is provided
func (tv *TokenValidator) Init(cfg plugin.PluginConfig) error {
	if !cfg.GetBool("enabled") {
		return fmt.Errorf("token-validator cannot be disabled (always-on per ADR PI2-yaa-0001 §5)")
	}

	testMode := cfg.GetBool("test_mode")
	jwksURL := cfg.GetString("jwks_url")
	jwtSecret := cfg.GetString("jwt_secret")
	audience := cfg.GetString("audience")

	ttlSecs := cfg.GetInt("cache_ttl_seconds")
	if ttlSecs <= 0 {
		ttlSecs = 600
	}

	switch {
	case testMode:
		if jwtSecret == "" {
			return fmt.Errorf("token-validator: test_mode is true but jwt_secret is empty")
		}
		tv.testMode = true
		tv.hs256Key = []byte(jwtSecret)
	case jwksURL != "":
		tv.jwksVal = newJWKSValidator(jwksURL, time.Duration(ttlSecs)*time.Second)
	default:
		return fmt.Errorf("token-validator: neither jwks_url nor test_mode+jwt_secret is configured; " +
			"set GATEWAY_JWT_JWKS_URL or GATEWAY_JWT_SECRET+test_mode")
	}

	tv.audience = audience
	return nil
}

// Handler returns the JWT validation middleware.
//
// On validation failure the middleware writes 403 application/vnd.yaagents.error+json
// with trace.correlationId populated from reqctx (or the X-Correlation-ID header as
// fallback) and returns without calling next.
//
// On success, actor subject, roles, and tenant ID are stored in the request context
// via reqctx and next is called.
func (tv *TokenValidator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := extractBearer(r)
		if !ok {
			writeForbidden(w, r, "MISSING_TOKEN",
				"Authorization header with Bearer token is required")
			return
		}

		mc, err := tv.validate(tokenStr)
		if err != nil {
			slog.Warn("token-validator: validation failed",
				slog.String("path", r.URL.Path),
				slog.String("error", err.Error()))
			writeForbidden(w, r, "INVALID_TOKEN", "token validation failed")
			return
		}

		// Propagate parsed claims into request context for downstream middleware.
		ctx := reqctx.WithActorSubject(r.Context(), claimStr(mc, "sub"))
		ctx = reqctx.WithActorRoles(ctx, claimStrSlice(mc, "roles"))
		if tid := claimStr(mc, "tenant_id"); tid != "" {
			ctx = reqctx.WithTenantID(ctx, tid)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Shutdown is a no-op; the plugin holds no persistent background resources.
func (tv *TokenValidator) Shutdown(_ context.Context) error { return nil }

// validate dispatches to the configured validation strategy.
func (tv *TokenValidator) validate(tokenStr string) (jwt.MapClaims, error) {
	if tv.testMode {
		return validateHS256(tokenStr, tv.hs256Key, tv.audience)
	}
	return tv.jwksVal.validate(tokenStr, tv.audience)
}

// validateHS256 parses and verifies a HS256-signed JWT.
func validateHS256(tokenStr string, key []byte, audience string) (jwt.MapClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	}
	if audience != "" {
		opts = append(opts, jwt.WithAudience(audience))
	}
	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("expected HS256, got %q", t.Header["alg"])
			}
			return key, nil
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type %T", token.Claims)
	}
	return mc, nil
}

// writeForbidden writes a 403 vendor-error body.
// Correlation and request IDs are read from reqctx first, then from headers.
func writeForbidden(w http.ResponseWriter, r *http.Request, code, msg string) {
	corrID := reqctx.CorrelationID(r.Context())
	if corrID == "" {
		corrID = r.Header.Get("X-Correlation-ID")
	}
	reqID := reqctx.RequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get("X-Request-ID")
	}
	response.WriteError(w, http.StatusForbidden, response.ErrorBody{
		Type:    "forbidden",
		Code:    code,
		Message: msg,
		Trace: response.Trace{
			CorrelationID: corrID,
			RequestID:     reqID,
		},
	})
}

// extractBearer extracts the token from "Authorization: Bearer <token>".
func extractBearer(r *http.Request) (string, bool) {
	after, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || after == "" {
		return "", false
	}
	return after, true
}

// claimStr returns the string value of a MapClaim key, or "" if absent or wrong type.
func claimStr(mc jwt.MapClaims, key string) string {
	s, _ := mc[key].(string)
	return s
}

// claimStrSlice converts a MapClaim array value to []string.
// JWT libraries represent JSON arrays as []interface{}; non-string elements are skipped.
func claimStrSlice(mc jwt.MapClaims, key string) []string {
	rv, ok := mc[key]
	if !ok {
		return nil
	}
	arr, ok := rv.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
