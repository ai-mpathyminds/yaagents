// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Inline JWKS fetch + RS256 validation (ADR PI2-yaa-0005 Decision 1).
// portfolio/packages/go/auth-jwks/ does not exist yet; this ~150-line
// implementation is the minimum viable inline. The Init signature is stable
// so the import path can switch when the extraction lands.
//
// Features:
//   - In-memory key cache keyed on "kid" with configurable TTL.
//   - Stale-while-revalidate: on JWKS refresh failure, previously-good keys are
//     served and a warn log is emitted; requests are not hard-failed.
//   - fetchCount atomic counter exposed via hitCount() for test assertions.

package tokenvalidator

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwk is a single JSON Web Key (RSA subset only).
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

// jwksValidator validates RS256 JWTs using public keys fetched from a JWKS endpoint.
type jwksValidator struct {
	url    string
	client *http.Client
	ttl    time.Duration

	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey // current keys (possibly stale after TTL expiry)
	fetchAt time.Time                 // time of last successful fetch

	fetchCount int64 // atomic; counts JWKS HTTP fetches; used in tests
}

// newJWKSValidator creates a jwksValidator for the given JWKS URL and cache TTL.
func newJWKSValidator(url string, ttl time.Duration) *jwksValidator {
	return &jwksValidator{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
		ttl:    ttl,
		keys:   map[string]*rsa.PublicKey{},
	}
}

// hitCount returns the number of JWKS HTTP fetches performed.
// Used in tests to verify cache behaviour (first request fetches; subsequent
// requests within TTL reuse the in-memory cache).
func (v *jwksValidator) hitCount() int64 {
	return atomic.LoadInt64(&v.fetchCount)
}

// validate parses and verifies a RS256 JWT. audience is validated when non-empty.
func (v *jwksValidator) validate(tokenStr, audience string) (jwt.MapClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
	}
	if audience != "" {
		opts = append(opts, jwt.WithAudience(audience))
	}
	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("expected RS256, got %q", t.Header["alg"])
			}
			kid, _ := t.Header["kid"].(string)
			return v.getKey(kid)
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

// getKey retrieves the RSA public key for kid.
// Uses the cache when live; triggers a refresh on TTL expiry.
// On refresh failure: serves stale keys if available (stale-while-revalidate);
// otherwise returns the error.
func (v *jwksValidator) getKey(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	expired := v.fetchAt.IsZero() || time.Now().After(v.fetchAt.Add(v.ttl))
	v.mu.RUnlock()

	if ok && !expired {
		return key, nil // fast path: cache hit within TTL
	}

	if err := v.refresh(); err != nil {
		// Stale-while-revalidate: on refresh failure serve the previously-fetched
		// key (still in v.keys; refresh only replaces v.keys on success).
		v.mu.RLock()
		staleKey, staleOK := v.keys[kid]
		v.mu.RUnlock()
		if staleOK {
			slog.Warn("jwks refresh failed; serving stale key",
				slog.String("kid", kid),
				slog.String("error", err.Error()))
			return staleKey, nil
		}
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

// refresh fetches the JWKS from v.url, updates the key cache, and saves the
// previous key set as staleKeys for stale-while-revalidate.
func (v *jwksValidator) refresh() error {
	atomic.AddInt64(&v.fetchCount, 1)

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

// rsaFromJWK builds an *rsa.PublicKey from base64url-encoded n and e JWK fields.
func rsaFromJWK(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode E: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil, fmt.Errorf("RSA exponent too large")
	}
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}
