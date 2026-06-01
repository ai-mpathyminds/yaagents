// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Inline JWKS fetch + RS256 validation (ADR PI2-yaa-0005 Decision 1).
// portfolio/packages/go/auth-jwks/ does not exist yet; this implementation is
// the minimum viable inline. The Init signature is stable so the import path
// can switch when the extraction lands.
//
// Features (v1 — preserved):
//   - In-memory key cache keyed on "kid" with configurable TTL.
//   - Stale-while-revalidate: on JWKS refresh failure, previously-good keys are
//     served and a warn log is emitted; requests are not hard-failed.
//   - fetchCount atomic counter exposed via hitCount() for test assertions.
//
// Features (v2 additions — PLG-3b / ADR PI2-yaa-0007):
//   - errJWKSUnavailable sentinel: cold-start fetch failure returns a typed
//     error so the handler can respond 503 vs 401.
//   - validateWith: RS256 validation without library-level audience check and
//     with configurable clock-skew leeway (jwt.WithLeeway).
//   - issuerEntry + jwksPool: per-issuer JWKS validator pool for multi-IdP
//     deployments.

package tokenvalidator

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ── errJWKSUnavailable ────────────────────────────────────────────────────────

// errJWKSUnavailable is returned by getKey when the JWKS endpoint could not be
// reached AND no cached keys exist (cold-start). The v2 handler maps this to
// the configured on_failure.jwks_unavailable code (default 503).
type errJWKSUnavailable struct{ cause error }

func (e *errJWKSUnavailable) Error() string {
	return "jwks unavailable (cold-start): " + e.cause.Error()
}

func (e *errJWKSUnavailable) Unwrap() error { return e.cause }

// isJWKSUnavailable reports whether err is or wraps errJWKSUnavailable.
func isJWKSUnavailable(err error) bool {
	var e *errJWKSUnavailable
	return errors.As(err, &e)
}

// ── jwk / jwkSet types ────────────────────────────────────────────────────────

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

// ── jwksValidator ─────────────────────────────────────────────────────────────

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
// (v1 method — preserved for backward compat.)
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

// validateWith validates a JWT using the provided algorithm allowlist and clock
// skew. Audience validation is intentionally absent — the v2 handler performs
// multi-audience matching itself (PLG-3b amendment 3).
//
// Returns errJWKSUnavailable when the JWKS endpoint is unreachable and no
// cached keys exist; the handler maps this to 503.
func (v *jwksValidator) validateWith(tokenStr string, methods []string, skew time.Duration) (jwt.MapClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods(methods),
		jwt.WithExpirationRequired(),
	}
	if skew > 0 {
		opts = append(opts, jwt.WithLeeway(skew))
	}
	token, err := jwt.Parse(tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			// Only RSA methods are supported by getKey; EC support is a v0.3+ addition.
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method %q", t.Header["alg"])
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
//
// On refresh failure:
//   - If stale keys exist for kid: serves them (stale-while-revalidate).
//   - If no keys have ever been cached (cold-start): returns errJWKSUnavailable.
//   - Otherwise: returns the original refresh error.
func (v *jwksValidator) getKey(kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	expired := v.fetchAt.IsZero() || time.Now().After(v.fetchAt.Add(v.ttl))
	v.mu.RUnlock()

	if ok && !expired {
		return key, nil // fast path: cache hit within TTL
	}

	if err := v.refresh(); err != nil {
		v.mu.RLock()
		staleKey, staleOK := v.keys[kid]
		hasAnyStaleKeys := len(v.keys) > 0
		v.mu.RUnlock()

		if staleOK {
			slog.Warn("jwks refresh failed; serving stale key",
				slog.String("kid", kid),
				slog.String("error", err.Error()))
			return staleKey, nil
		}
		if !hasAnyStaleKeys {
			// Cold-start: never fetched successfully — signal 503.
			return nil, &errJWKSUnavailable{cause: err}
		}
		return nil, fmt.Errorf("jwks: refresh failed and kid %q not in stale set: %w", kid, err)
	}

	v.mu.RLock()
	key, ok = v.keys[kid]
	v.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("jwks: no key found for kid %q", kid)
	}
	return key, nil
}

// refresh fetches the JWKS from v.url, updates the key cache.
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

// ── jwksPool — per-issuer JWKS validator pool (PLG-3b amendment 1) ────────────

// issuerEntry pairs an issuer string with its dedicated JWKS validator.
type issuerEntry struct {
	issuer    string // expected "iss" claim value; "" matches any issuer (v1 shim)
	validator *jwksValidator
}

// jwksPool holds per-issuer JWKS validators for multi-IdP deployments.
// Each entry corresponds to one element of the "issuers:" config list.
type jwksPool struct {
	entries []issuerEntry
}

// findValidator returns the JWKS validator whose issuer field matches iss.
// An entry with an empty issuer acts as a wildcard fallback (v1 compat shim).
// Returns nil when no matching entry exists.
func (p *jwksPool) findValidator(iss string) *jwksValidator {
	var wildcard *jwksValidator
	for i := range p.entries {
		if p.entries[i].issuer == iss {
			return p.entries[i].validator
		}
		if p.entries[i].issuer == "" {
			wildcard = p.entries[i].validator
		}
	}
	return wildcard
}
