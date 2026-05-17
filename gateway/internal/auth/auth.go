// Package auth provides JWT bearer validation middleware for the yaagents gateway.
//
// Two validation modes per ADR PI1-yaa-0001 §3:
//   - HS256: symmetric secret via GATEWAY_JWT_SECRET (dev/demo default)
//   - RS256: public-key via GATEWAY_JWT_JWKS_URL with cached JWKS (production)
//
// JWKS takes precedence when both env vars are set (warn-logged).
// Missing or invalid bearer token → 401 application/vnd.yaagents.error+json.
package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
)

// Validator validates a raw JWT string and returns parsed Claims on success.
type Validator interface {
	Validate(tokenStr string) (Claims, error)
}

// Claims holds the parsed JWT fields the gateway uses for RBAC and context
// propagation (tenant/actor injection in WI-1yaa.GW-3; role check in GW-4).
type Claims struct {
	// Subject is the "sub" JWT claim.
	Subject string
	// Roles is the "roles" JWT claim (string array).
	Roles []string
	// Raw holds all parsed map claims for upstream header injection.
	Raw map[string]interface{}
}

// contextKey is unexported to prevent collision with other packages.
type contextKey int

// ClaimsKey is the context key under which validated Claims are stored by Middleware.
// Consumers: GW-3 (tenant/actor extraction), GW-4 (RBAC check).
const ClaimsKey contextKey = 0

// Middleware wraps next with JWT bearer validation.
// On auth failure it writes 401 application/vnd.yaagents.error+json and
// does not call next. On success it stores Claims in the request context.
func Middleware(v Validator, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, ok := extractBearer(r)
			if !ok {
				writeAuthError(w, r, "MISSING_TOKEN",
					"authorization header with Bearer token is required")
				return
			}
			claims, err := v.Validate(tokenStr)
			if err != nil {
				log.Warn("auth rejected",
					slog.String("path", r.URL.Path),
					slog.String("error", err.Error()))
				writeAuthError(w, r, "INVALID_TOKEN", "token validation failed")
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearer pulls the token string from "Authorization: Bearer <token>".
func extractBearer(r *http.Request) (string, bool) {
	after, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || after == "" {
		return "", false
	}
	return after, true
}

// writeAuthError emits a 401 vendor-error body, propagating or generating
// correlation and request IDs for the trace block.
func writeAuthError(w http.ResponseWriter, r *http.Request, code, msg string) {
	corrID := r.Header.Get("X-Correlation-ID")
	if corrID == "" {
		corrID = newUUID()
	}
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = newUUID()
	}
	response.WriteError(w, http.StatusUnauthorized, response.ErrorBody{
		Type:    "error",
		Code:    code,
		Message: msg,
		Trace: response.Trace{
			CorrelationID: corrID,
			RequestID:     reqID,
		},
	})
}

// newUUID returns a random UUID v4 string.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// extractClaims converts jwt.MapClaims into the gateway's Claims struct.
// Shared by HS256Validator and JWKSValidator.
func extractClaims(token *jwt.Token) (Claims, error) {
	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, fmt.Errorf("unexpected claims type %T", token.Claims)
	}

	sub, _ := mc["sub"].(string)

	var roles []string
	if rv, exists := mc["roles"]; exists {
		if arr, ok := rv.([]interface{}); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}

	raw := make(map[string]interface{}, len(mc))
	for k, v := range mc {
		raw[k] = v
	}

	return Claims{Subject: sub, Roles: roles, Raw: raw}, nil
}
