package auth

import (
	"errors"
	"log/slog"
	"os"
)

// NewValidator creates the appropriate Validator from environment variables.
//
// Precedence per ADR PI1-yaa-0001 §3:
//   - GATEWAY_JWT_JWKS_URL set (alone or alongside GATEWAY_JWT_SECRET): RS256 via JWKS
//   - Only GATEWAY_JWT_SECRET set: HS256 symmetric
//   - Neither set: returns an error (boot failure — caller should os.Exit(1))
func NewValidator(log *slog.Logger) (Validator, error) {
	secret := os.Getenv("GATEWAY_JWT_SECRET")
	jwksURL := os.Getenv("GATEWAY_JWT_JWKS_URL")

	switch {
	case jwksURL != "" && secret != "":
		log.Warn("both GATEWAY_JWT_SECRET and GATEWAY_JWT_JWKS_URL are set; JWKS (RS256) takes precedence",
			slog.String("jwks_url", jwksURL))
		return NewJWKSValidator(jwksURL), nil
	case jwksURL != "":
		return NewJWKSValidator(jwksURL), nil
	case secret != "":
		return NewHS256Validator(secret), nil
	default:
		return nil, errors.New("auth: JWT validator not configured; set GATEWAY_JWT_SECRET or GATEWAY_JWT_JWKS_URL")
	}
}
