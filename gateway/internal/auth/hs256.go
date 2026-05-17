package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// HS256Validator validates JWT tokens signed with HMAC-SHA256.
// Used in dev/demo mode (GATEWAY_JWT_SECRET env var).
type HS256Validator struct {
	secret []byte
}

// NewHS256Validator creates a Validator using the given symmetric secret.
func NewHS256Validator(secret string) *HS256Validator {
	return &HS256Validator{secret: []byte(secret)}
}

// Validate parses and verifies a HS256-signed JWT.
// Returns an error if the signature is invalid, the token is expired,
// or the signing method is not HS256.
func (v *HS256Validator) Validate(tokenStr string) (Claims, error) {
	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %q; expected HS256",
					t.Header["alg"])
			}
			return v.secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("hs256: %w", err)
	}
	return extractClaims(token)
}
