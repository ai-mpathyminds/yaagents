package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwk is a single JSON Web Key (RSA subset).
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// jwkSet is the top-level JWKS document.
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// JWKSValidator validates RS256 JWT tokens using public keys fetched from a
// JWKS endpoint. Keys are cached in memory with a 5-minute TTL.
// Used in production (GATEWAY_JWT_JWKS_URL env var).
type JWKSValidator struct {
	url    string
	client *http.Client

	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetchAt time.Time
	ttl     time.Duration
}

// NewJWKSValidator creates a Validator that fetches RS256 public keys from url.
func NewJWKSValidator(url string) *JWKSValidator {
	return &JWKSValidator{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
		ttl:    5 * time.Minute,
		keys:   make(map[string]*rsa.PublicKey),
	}
}

// Validate parses and verifies an RS256-signed JWT using the JWKS public key
// identified by the token's "kid" header field.
func (v *JWKSValidator) Validate(tokenStr string) (Claims, error) {
	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method %q; expected RS256",
					t.Header["alg"])
			}
			kid, _ := t.Header["kid"].(string)
			return v.getKey(kid)
		},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("jwks: %w", err)
	}
	return extractClaims(token)
}

// getKey returns the RSA public key for kid, refreshing the cache when stale.
func (v *JWKSValidator) getKey(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	expired := v.fetchAt.IsZero() || time.Now().After(v.fetchAt.Add(v.ttl))
	v.mu.RUnlock()

	if ok && !expired {
		return key, nil
	}

	// Refresh outside the read lock; HTTP fetch must not hold the lock.
	if err := v.refresh(); err != nil {
		return nil, err
	}

	v.mu.RLock()
	key, ok = v.keys[kid]
	v.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("jwks: no key found for kid %q", kid)
	}
	return key, nil
}

// refresh fetches the JWKS from v.url and updates the key cache.
func (v *JWKSValidator) refresh() error {
	resp, err := v.client.Get(v.url)
	if err != nil {
		return fmt.Errorf("jwks fetch: %w", err)
	}
	defer resp.Body.Close()

	var ks jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&ks); err != nil {
		return fmt.Errorf("jwks decode: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(ks.Keys))
	for _, k := range ks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := rsaFromJWK(k)
		if err != nil {
			return fmt.Errorf("jwks key %q: %w", k.Kid, err)
		}
		keys[k.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchAt = time.Now()
	v.mu.Unlock()

	return nil
}

// rsaFromJWK builds an *rsa.PublicKey from base64url-encoded n and e fields.
func rsaFromJWK(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil, fmt.Errorf("RSA exponent too large")
	}
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
