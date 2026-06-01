// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tokenvalidator implements the token-validator plugin (PRD §6.5 plugin a).
//
// # v1 — PLG-3 (commit 442cece)
//
// Single-issuer, RS256/JWKS + HS256 test mode. Loads from jwks_url + audience
// (singular) config keys. Returns 403 on any validation failure.
//
// # v2 — PLG-3b (ADR PI2-yaa-0007)
//
// Additive extension. Activated when any of the following config keys are
// present: issuers, audiences, algorithms, clock_skew_seconds, required_claims,
// propagate_claims, token, on_failure, max_token_bytes.
//
// Amendments on top of v1:
//
//  1. Multi-issuer + per-issuer JWKS pool (issuers: list).
//  2. Algorithm allowlist enforced before signature verification; "none" forbidden.
//  3. Multi-audience list (token aud must match ≥1).
//  4. Clock-skew tolerance applied to exp / nbf / iat.
//  5. Required-claims non-empty check.
//  6. Propagate-claims contract: all | allowlist.
//  7. Configurable token header + scheme.
//  8. RFC-correct status codes via on_failure map (401 default suite).
//  9. Token-size cap checked before parse.
//
// Backwards compatibility: a v1 config (jwks_url + audience) loads cleanly in
// v2 mode (backward-compat shim) with a WARN at boot. The shim preserves v1
// status-code defaults (403) so existing deployments are not broken.
//
// Registration: init() → plugin.Register(&TokenValidator{}) per ADR PI2-yaa-0001 §3.
package tokenvalidator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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

// ── failureCodes ──────────────────────────────────────────────────────────────

// failureCodes holds the HTTP status code returned for each failure mode.
// All codes default to 401 in v2 mode (RFC 7235); operators may override.
type failureCodes struct {
	MissingToken         int
	InvalidSignature     int
	Expired              int
	NotYetValid          int
	UnknownIssuer        int
	AudienceMismatch     int
	RequiredClaimMissing int
	DisallowedAlgorithm  int
	OversizedToken       int
	JWKSUnavailable      int
}

// defaultFailureCodes returns the v2 defaults: 401 for credential failures,
// 400 for oversized token, 503 for JWKS unreachable.
func defaultFailureCodes() failureCodes {
	return failureCodes{
		MissingToken:         http.StatusUnauthorized,
		InvalidSignature:     http.StatusUnauthorized,
		Expired:              http.StatusUnauthorized,
		NotYetValid:          http.StatusUnauthorized,
		UnknownIssuer:        http.StatusUnauthorized,
		AudienceMismatch:     http.StatusUnauthorized,
		RequiredClaimMissing: http.StatusUnauthorized,
		DisallowedAlgorithm:  http.StatusUnauthorized,
		OversizedToken:       http.StatusBadRequest,
		JWKSUnavailable:      http.StatusServiceUnavailable,
	}
}

// defaultFailureCodesV1 returns v1-compatible defaults (403 for credential
// failures). Used by the v1-shim path to preserve backwards compatibility.
func defaultFailureCodesV1() failureCodes {
	return failureCodes{
		MissingToken:         http.StatusForbidden,
		InvalidSignature:     http.StatusForbidden,
		Expired:              http.StatusForbidden,
		NotYetValid:          http.StatusForbidden,
		UnknownIssuer:        http.StatusForbidden,
		AudienceMismatch:     http.StatusForbidden,
		RequiredClaimMissing: http.StatusForbidden,
		DisallowedAlgorithm:  http.StatusForbidden,
		OversizedToken:       http.StatusBadRequest,
		JWKSUnavailable:      http.StatusServiceUnavailable,
	}
}

// parseFailureCodes overlays on_failure config values on top of defaults.
func parseFailureCodes(m map[string]any, def failureCodes) failureCodes {
	get := func(key string, d int) int {
		v, ok := m[key]
		if !ok {
			return d
		}
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		default:
			return d
		}
	}
	return failureCodes{
		MissingToken:         get("missing_token", def.MissingToken),
		InvalidSignature:     get("invalid_signature", def.InvalidSignature),
		Expired:              get("expired", def.Expired),
		NotYetValid:          get("not_yet_valid", def.NotYetValid),
		UnknownIssuer:        get("unknown_issuer", def.UnknownIssuer),
		AudienceMismatch:     get("audience_mismatch", def.AudienceMismatch),
		RequiredClaimMissing: get("required_claim_missing", def.RequiredClaimMissing),
		DisallowedAlgorithm:  get("disallowed_algorithm", def.DisallowedAlgorithm),
		OversizedToken:       get("oversized_token", def.OversizedToken),
		JWKSUnavailable:      get("jwks_unavailable", def.JWKSUnavailable),
	}
}

// ── TokenValidator ────────────────────────────────────────────────────────────

// TokenValidator is the PLG-3 / PLG-3b full token-validator plugin.
// Zero value is invalid; always call Init before Handler.
type TokenValidator struct {
	// ── v1 fields (preserved for backward compat and test access) ──────────

	// testMode is true when the plugin is in HS256 dev/test mode.
	testMode bool
	hs256Key []byte

	// jwksVal is the v1 single-issuer JWKS validator. Set by v1 config form
	// (jwks_url key). Also populated by the v1 compat shim in v2 mode so that
	// existing tests that access tv.jwksVal.hitCount() continue to work.
	jwksVal *jwksValidator

	// audience is the v1 single-audience string (skip validation when empty).
	audience string

	// ── v2 fields (PLG-3b / ADR PI2-yaa-0007) ──────────────────────────────

	// v2mode is true when any v2-specific config key is present. Controls
	// which handler branch executes.
	v2mode bool

	// issuersPool is the per-issuer JWKS validator pool (v2 multi-issuer).
	// Nil when v2mode is false.
	issuersPool *jwksPool

	// algorithms is the allowlist checked before signature verification.
	// Set to ["HS256"] in test_mode; ["RS256","ES256"] by default otherwise.
	algorithms []string

	// audiences is the v2 multi-audience list. Empty = skip aud validation.
	audiences []string

	// clockSkew is the tolerance applied to exp / nbf / iat checks.
	clockSkew time.Duration

	// requiredClaims lists claim names that must be present and non-empty.
	requiredClaims []string

	// propagateMode controls which validated claims land in the request context.
	// "all" (default) or "allowlist".
	propagateMode   string
	propagateClaims []string // used when propagateMode == "allowlist"

	// tokenHeader and tokenScheme configure token extraction (defaults:
	// "Authorization" and "Bearer"). scheme == "" means no prefix is stripped.
	tokenHeader string
	tokenScheme string

	// onFailure holds per-failure HTTP status codes.
	onFailure failureCodes

	// maxTokenBytes is the token size cap checked before parse. 0 = unlimited.
	maxTokenBytes int
}

// Name returns the canonical plugin identifier.
func (tv *TokenValidator) Name() string { return "token-validator" }

// Init validates configuration and wires the appropriate validation strategy.
//
// v2 mode is activated when any of these config keys are present:
// issuers, audiences, algorithms, clock_skew_seconds, required_claims,
// propagate_claims, token, on_failure, max_token_bytes.
//
// v1 config (jwks_url + audience singular) is accepted in both modes:
// when v2 mode is active it triggers the v1-compat shim + WARN log.
func (tv *TokenValidator) Init(cfg plugin.PluginConfig) error {
	if !cfg.GetBool("enabled") {
		return fmt.Errorf("token-validator cannot be disabled (always-on per ADR PI2-yaa-0001 §5)")
	}

	raw := cfg.Raw()
	tv.v2mode = isV2Config(raw)

	if tv.v2mode {
		return tv.initV2(cfg, raw)
	}
	return tv.initV1(cfg)
}

// isV2Config returns true when any v2-exclusive config key is present.
func isV2Config(raw map[string]any) bool {
	for _, k := range []string{
		"issuers", "audiences", "algorithms", "clock_skew_seconds",
		"required_claims", "propagate_claims", "token", "on_failure", "max_token_bytes",
	} {
		if _, ok := raw[k]; ok {
			return true
		}
	}
	return false
}

// initV1 is the original v1 Init logic (unchanged). Called when no v2 config
// keys are present.
func (tv *TokenValidator) initV1(cfg plugin.PluginConfig) error {
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

// initV2 parses and validates all v2 config keys. Accepts a v1-compat shim
// (jwks_url + audience singular) and emits a WARN when the shim fires.
func (tv *TokenValidator) initV2(cfg plugin.PluginConfig, raw map[string]any) error {
	testMode := cfg.GetBool("test_mode")
	jwtSecret := cfg.GetString("jwt_secret")

	// ── (8) on_failure codes — parse first (used by other init steps) ─────
	failDefaults := defaultFailureCodes()

	// ── (1) Multi-issuer / v1-compat shim ─────────────────────────────────
	var v1Shim bool
	if testMode {
		if jwtSecret == "" {
			return fmt.Errorf("token-validator: test_mode is true but jwt_secret is empty")
		}
		tv.testMode = true
		tv.hs256Key = []byte(jwtSecret)
	} else {
		issuers, err := parseIssuers(raw)
		if err != nil {
			return fmt.Errorf("token-validator: issuers: %w", err)
		}
		if len(issuers) == 0 {
			// v1 compat shim: fall back to jwks_url
			jwksURL := cfg.GetString("jwks_url")
			if jwksURL == "" {
				return fmt.Errorf("token-validator: issuers list is empty and test_mode is false; " +
					"set issuers or test_mode+jwt_secret")
			}
			ttlSecs := cfg.GetInt("cache_ttl_seconds")
			if ttlSecs <= 0 {
				ttlSecs = 600
			}
			slog.Warn("token-validator config uses v1 single-issuer form; " +
				"please migrate to issuers: list per ADR PI2-yaa-0007")
			tv.jwksVal = newJWKSValidator(jwksURL, time.Duration(ttlSecs)*time.Second)
			tv.issuersPool = &jwksPool{
				entries: []issuerEntry{{issuer: "", validator: tv.jwksVal}},
			}
			v1Shim = true
			// v1 shim defaults to 403 suite unless on_failure overrides
			failDefaults = defaultFailureCodesV1()
		} else {
			pool := &jwksPool{entries: make([]issuerEntry, 0, len(issuers))}
			for _, ic := range issuers {
				if ic.issuer == "" {
					return fmt.Errorf("token-validator: issuers[].issuer must be non-empty in v2 mode")
				}
				if _, err := url.ParseRequestURI(ic.jwksURL); err != nil {
					return fmt.Errorf("token-validator: issuers[].jwks_url %q is not a valid URL: %w",
						ic.jwksURL, err)
				}
				pool.entries = append(pool.entries, issuerEntry{
					issuer:    ic.issuer,
					validator: newJWKSValidator(ic.jwksURL, time.Duration(ic.ttl)*time.Second),
				})
			}
			tv.issuersPool = pool
		}
	}

	// ── (2) Algorithm allowlist ────────────────────────────────────────────
	if testMode {
		tv.algorithms = []string{"HS256"}
	} else {
		algs := cfg.GetStringSlice("algorithms")
		if len(algs) == 0 {
			algs = []string{"RS256", "ES256"} // default
		}
		for _, a := range algs {
			if strings.EqualFold(a, "none") {
				return fmt.Errorf("token-validator: algorithm \"none\" is forbidden (ADR PI2-yaa-0007)")
			}
		}
		tv.algorithms = algs
	}

	// ── (3) Multi-audience ─────────────────────────────────────────────────
	if auds := cfg.GetStringSlice("audiences"); len(auds) > 0 {
		tv.audiences = auds
	} else if !v1Shim {
		// leave tv.audiences nil → audience validation skipped
	} else {
		// v1 shim: promote singular audience to slice
		if aud := cfg.GetString("audience"); aud != "" {
			tv.audiences = []string{aud}
		}
	}

	// ── (4) Clock skew ─────────────────────────────────────────────────────
	skewSecs := cfg.GetInt("clock_skew_seconds")
	if skewSecs == 0 {
		skewSecs = 60 // default 60 s
	}
	if skewSecs < 0 || skewSecs > 600 {
		return fmt.Errorf("token-validator: clock_skew_seconds must be 0..600, got %d", skewSecs)
	}
	tv.clockSkew = time.Duration(skewSecs) * time.Second

	// ── (5) Required claims ────────────────────────────────────────────────
	if rc := cfg.GetStringSlice("required_claims"); len(rc) > 0 {
		tv.requiredClaims = rc
	} else {
		tv.requiredClaims = []string{"sub"} // default
	}

	// ── (6) Propagate-claims contract ──────────────────────────────────────
	tv.propagateMode = "all" // default
	if pcRaw, ok := raw["propagate_claims"].(map[string]any); ok {
		mode, _ := pcRaw["mode"].(string)
		if mode != "" {
			tv.propagateMode = mode
		}
		if tv.propagateMode != "all" && tv.propagateMode != "allowlist" {
			return fmt.Errorf("token-validator: propagate_claims.mode must be \"all\" or \"allowlist\", got %q",
				tv.propagateMode)
		}
		if tv.propagateMode == "allowlist" {
			claims, err := stringSliceFromAny(pcRaw["claims"])
			if err != nil || len(claims) == 0 {
				return fmt.Errorf("token-validator: propagate_claims.mode=allowlist requires non-empty claims list")
			}
			tv.propagateClaims = claims
		}
	}

	// ── (7) Configurable token header ──────────────────────────────────────
	tv.tokenHeader = "Authorization" // default
	tv.tokenScheme = "Bearer"        // default
	if tokenRaw, ok := raw["token"].(map[string]any); ok {
		if h, ok := tokenRaw["header"].(string); ok && h != "" {
			tv.tokenHeader = h
		}
		if tv.tokenHeader == "" {
			return fmt.Errorf("token-validator: token.header must not be empty")
		}
		if s, ok := tokenRaw["scheme"].(string); ok {
			tv.tokenScheme = s // empty scheme = no prefix strip
		}
	}

	// ── (9) Token size cap ─────────────────────────────────────────────────
	maxBytes := cfg.GetInt("max_token_bytes")
	if maxBytes == 0 {
		maxBytes = 8192 // default
	}
	if maxBytes < 0 || maxBytes > 65536 {
		return fmt.Errorf("token-validator: max_token_bytes must be 1..65536, got %d", maxBytes)
	}
	tv.maxTokenBytes = maxBytes

	// ── (8) on_failure codes ───────────────────────────────────────────────
	if fcRaw, ok := raw["on_failure"].(map[string]any); ok {
		tv.onFailure = parseFailureCodes(fcRaw, failDefaults)
	} else {
		tv.onFailure = failDefaults
	}

	return nil
}

// issuerCfg is an intermediate struct used only during Init parsing.
type issuerCfg struct {
	issuer  string
	jwksURL string
	ttl     int
}

// parseIssuers reads the "issuers" key from the raw config map.
// Returns nil (not an error) when the key is absent.
func parseIssuers(raw map[string]any) ([]issuerCfg, error) {
	v, ok := raw["issuers"]
	if !ok {
		return nil, nil
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}
	out := make([]issuerCfg, 0, len(items))
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("[%d] must be a map", i)
		}
		iss, _ := m["issuer"].(string)
		jwksURL, _ := m["jwks_url"].(string)
		if jwksURL == "" {
			return nil, fmt.Errorf("[%d]: jwks_url is required", i)
		}
		ttl := 600
		if t, ok := m["jwks_cache_ttl_seconds"]; ok {
			switch n := t.(type) {
			case int:
				ttl = n
			case int64:
				ttl = int(n)
			case float64:
				ttl = int(n)
			}
		}
		out = append(out, issuerCfg{issuer: iss, jwksURL: jwksURL, ttl: ttl})
	}
	return out, nil
}

// stringSliceFromAny converts a raw any value ([]any or []string) to []string.
func stringSliceFromAny(v any) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	switch typed := v.(type) {
	case []string:
		return typed, nil
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("element %v is not a string", item)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected []string, got %T", v)
	}
}

// ── Handler ───────────────────────────────────────────────────────────────────

// Handler returns the JWT validation middleware.
// Routes to v2 handler when tv.v2mode is true; otherwise falls through to v1.
func (tv *TokenValidator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tv.v2mode {
			tv.handleV2(next, w, r)
		} else {
			tv.handleV1(next, w, r)
		}
	})
}

// handleV1 is the original v1 handler — unchanged from PLG-3.
func (tv *TokenValidator) handleV1(next http.Handler, w http.ResponseWriter, r *http.Request) {
	tokenStr, ok := extractBearer(r)
	if !ok {
		writeForbidden(w, r, "MISSING_TOKEN", "Authorization header with Bearer token is required")
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

	ctx := reqctx.WithActorSubject(r.Context(), claimStr(mc, "sub"))
	ctx = reqctx.WithActorRoles(ctx, claimStrSlice(mc, "roles"))
	if tid := claimStr(mc, "tenant_id"); tid != "" {
		ctx = reqctx.WithTenantID(ctx, tid)
	}
	ctx = reqctx.WithJWTClaims(ctx, map[string]any(mc))
	next.ServeHTTP(w, r.WithContext(ctx))
}

// handleV2 implements the PLG-3b handler with all 9 amendments.
func (tv *TokenValidator) handleV2(next http.Handler, w http.ResponseWriter, r *http.Request) {
	// (7) Extract token from configurable header + scheme.
	tokenStr, ok := tv.extractTokenV2(r)
	if !ok {
		tv.writeV2Error(w, r, tv.onFailure.MissingToken, "MISSING_TOKEN", "token is required")
		return
	}

	// (9) Token size cap — BEFORE parse.
	if tv.maxTokenBytes > 0 && len(tokenStr) > tv.maxTokenBytes {
		slog.Warn("token-validator: oversized token rejected",
			slog.Int("size", len(tokenStr)),
			slog.Int("max", tv.maxTokenBytes))
		tv.writeV2Error(w, r, tv.onFailure.OversizedToken, "OVERSIZED_TOKEN",
			fmt.Sprintf("token exceeds maximum size of %d bytes", tv.maxTokenBytes))
		return
	}

	// (2) Algorithm pre-check — BEFORE signature verify.
	if !tv.testMode {
		alg := peekAlgorithm(tokenStr)
		if alg == "" {
			tv.writeV2Error(w, r, tv.onFailure.InvalidSignature, "INVALID_TOKEN",
				"malformed JWT: cannot decode header")
			return
		}
		allowed := false
		for _, a := range tv.algorithms {
			if a == alg {
				allowed = true
				break
			}
		}
		if !allowed {
			slog.Warn("token-validator: disallowed algorithm",
				slog.String("alg", alg), slog.String("path", r.URL.Path))
			tv.writeV2Error(w, r, tv.onFailure.DisallowedAlgorithm, "DISALLOWED_ALGORITHM",
				fmt.Sprintf("algorithm %q is not allowed", alg))
			return
		}
	}

	var (
		mc  jwt.MapClaims
		err error
	)

	if tv.testMode {
		// HS256 test mode: no issuer matching; use clock skew.
		mc, err = validateHS256V2(tokenStr, tv.hs256Key, tv.clockSkew)
	} else {
		// (1) Multi-issuer: read iss claim, find matching JWKS validator.
		iss := peekIssuer(tokenStr)
		v := tv.issuersPool.findValidator(iss)
		if v == nil {
			slog.Warn("token-validator: unknown issuer",
				slog.String("iss", iss), slog.String("path", r.URL.Path))
			tv.writeV2Error(w, r, tv.onFailure.UnknownIssuer, "UNKNOWN_ISSUER",
				fmt.Sprintf("issuer %q is not configured", iss))
			return
		}
		mc, err = v.validateWith(tokenStr, tv.algorithms, tv.clockSkew)
	}

	if err != nil {
		slog.Warn("token-validator: validation failed",
			slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		if isJWKSUnavailable(err) {
			tv.writeV2Error(w, r, tv.onFailure.JWKSUnavailable, "JWKS_UNAVAILABLE",
				"JWKS service unavailable; try again later")
			return
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			tv.writeV2Error(w, r, tv.onFailure.Expired, "TOKEN_EXPIRED", "token has expired")
			return
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			tv.writeV2Error(w, r, tv.onFailure.NotYetValid, "TOKEN_NOT_YET_VALID",
				"token is not yet valid")
			return
		}
		tv.writeV2Error(w, r, tv.onFailure.InvalidSignature, "INVALID_TOKEN",
			"token validation failed")
		return
	}

	// (3) Multi-audience check.
	if len(tv.audiences) > 0 {
		if !checkAudiences(mc, tv.audiences) {
			tv.writeV2Error(w, r, tv.onFailure.AudienceMismatch, "AUDIENCE_MISMATCH",
				"token audience does not match configured audiences")
			return
		}
	}

	// (5) Required-claims check.
	if missing := checkRequiredClaims(mc, tv.requiredClaims); missing != "" {
		tv.writeV2Error(w, r, tv.onFailure.RequiredClaimMissing, "REQUIRED_CLAIM_MISSING",
			fmt.Sprintf("required claim %q is missing or empty", missing))
		return
	}

	// (6) Propagate claims to request context.
	ctx := propagateClaimsToCtx(r.Context(), mc, tv.propagateMode, tv.propagateClaims)
	// Also wire the v1-compat reqctx fields (sub, roles, tenant_id).
	ctx = reqctx.WithActorSubject(ctx, claimStr(mc, "sub"))
	ctx = reqctx.WithActorRoles(ctx, claimStrSlice(mc, "roles"))
	if tid := claimStr(mc, "tenant_id"); tid != "" {
		ctx = reqctx.WithTenantID(ctx, tid)
	}

	next.ServeHTTP(w, r.WithContext(ctx))
}

// Shutdown is a no-op; the plugin holds no persistent background resources.
func (tv *TokenValidator) Shutdown(_ context.Context) error { return nil }

// ── v1 helpers (unchanged from PLG-3) ────────────────────────────────────────

// validate dispatches to the configured v1 validation strategy.
func (tv *TokenValidator) validate(tokenStr string) (jwt.MapClaims, error) {
	if tv.testMode {
		return validateHS256(tokenStr, tv.hs256Key, tv.audience)
	}
	return tv.jwksVal.validate(tokenStr, tv.audience)
}

// validateHS256 parses and verifies a HS256-signed JWT (v1).
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

// writeForbidden writes a 403 vendor-error body (v1 — unchanged).
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

// extractBearer extracts the token from "Authorization: Bearer <token>" (v1).
func extractBearer(r *http.Request) (string, bool) {
	after, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || after == "" {
		return "", false
	}
	return after, true
}

// ── v2 helpers (PLG-3b) ───────────────────────────────────────────────────────

// extractTokenV2 reads the token from the configurable header + scheme.
func (tv *TokenValidator) extractTokenV2(r *http.Request) (string, bool) {
	val := r.Header.Get(tv.tokenHeader)
	if val == "" {
		return "", false
	}
	if tv.tokenScheme == "" {
		return val, true
	}
	after, ok := strings.CutPrefix(val, tv.tokenScheme+" ")
	if !ok || after == "" {
		return "", false
	}
	return after, true
}

// writeV2Error writes a vendor-error body with the supplied status code.
func (tv *TokenValidator) writeV2Error(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	corrID := reqctx.CorrelationID(r.Context())
	if corrID == "" {
		corrID = r.Header.Get("X-Correlation-ID")
	}
	reqID := reqctx.RequestID(r.Context())
	if reqID == "" {
		reqID = r.Header.Get("X-Request-ID")
	}
	response.WriteError(w, status, response.ErrorBody{
		Type:    "error",
		Code:    code,
		Message: msg,
		Trace: response.Trace{
			CorrelationID: corrID,
			RequestID:     reqID,
		},
	})
}

// validateHS256V2 parses and verifies a HS256-signed JWT with clock-skew leeway.
func validateHS256V2(tokenStr string, key []byte, skew time.Duration) (jwt.MapClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	}
	if skew > 0 {
		opts = append(opts, jwt.WithLeeway(skew))
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

// peekAlgorithm extracts the "alg" header from a JWT without signature
// verification. Returns "" when the token is malformed or the header cannot
// be decoded.
func peekAlgorithm(tokenStr string) string {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	hdr, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	var header map[string]any
	if err := json.Unmarshal(hdr, &header); err != nil {
		return ""
	}
	alg, _ := header["alg"].(string)
	return alg
}

// peekIssuer extracts the "iss" claim from a JWT payload without signature
// verification. Returns "" when the token is malformed.
func peekIssuer(tokenStr string) string {
	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	iss, _ := claims["iss"].(string)
	return iss
}

// propagateClaimsToCtx stores validated JWT claims in the request context
// according to propagate_claims.mode.
func propagateClaimsToCtx(ctx context.Context, mc jwt.MapClaims, mode string, allow []string) context.Context {
	if mode == "allowlist" {
		filtered := make(map[string]any, len(allow))
		for _, k := range allow {
			if v, ok := mc[k]; ok {
				filtered[k] = v
			}
		}
		return reqctx.WithJWTClaims(ctx, filtered)
	}
	// "all" (default)
	return reqctx.WithJWTClaims(ctx, map[string]any(mc))
}

// checkAudiences returns true when the token's aud claim contains at least one
// value from the configured audiences list. RFC 7519: aud may be a single
// string or an array of strings.
func checkAudiences(mc jwt.MapClaims, audiences []string) bool {
	aud, ok := mc["aud"]
	if !ok {
		return false
	}
	var tokenAuds []string
	switch v := aud.(type) {
	case string:
		tokenAuds = []string{v}
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok {
				tokenAuds = append(tokenAuds, s)
			}
		}
	case []string:
		tokenAuds = v
	}
	for _, ta := range tokenAuds {
		for _, ca := range audiences {
			if ta == ca {
				return true
			}
		}
	}
	return false
}

// checkRequiredClaims returns the first claim name that is missing or empty,
// or "" if all required claims are present and non-empty.
func checkRequiredClaims(mc jwt.MapClaims, required []string) string {
	for _, name := range required {
		v, ok := mc[name]
		if !ok {
			return name
		}
		if s, isStr := v.(string); isStr && s == "" {
			return name
		}
	}
	return ""
}

// ── v1 shared helpers (used by both handlers) ─────────────────────────────────

// claimStr returns the string value of a MapClaim key, or "" if absent or wrong type.
func claimStr(mc jwt.MapClaims, key string) string {
	s, _ := mc[key].(string)
	return s
}

// claimStrSlice converts a MapClaim array value to []string.
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
