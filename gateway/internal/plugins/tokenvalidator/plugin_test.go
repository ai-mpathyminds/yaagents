// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Internal package tests give access to jwksVal.hitCount() and unexported helpers.
package tokenvalidator

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// ── RSA test helpers ──────────────────────────────────────────────────────────

// testRSAKey generates a 1024-bit RSA key pair for tests (fast; not for production).
func testRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return priv
}

// publicKeyToJWK encodes an RSA public key as a minimal JWK map.
func publicKeyToJWK(pub *rsa.PublicKey, kid string) map[string]string {
	nBytes := pub.N.Bytes()
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	return map[string]string{
		"kty": "RSA",
		"kid": kid,
		"alg": "RS256",
		"use": "sig",
		"n":   base64.RawURLEncoding.EncodeToString(nBytes),
		"e":   base64.RawURLEncoding.EncodeToString(eBytes),
	}
}

// jwksServer starts a test HTTP server that serves the given keys as a JWKS document.
// hitCalls is incremented on each request so tests can verify cache behaviour.
func jwksServer(t *testing.T, keys map[string]*rsa.PublicKey) (url string, hitCalls *atomic.Int64) {
	t.Helper()
	var n atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		var jwkList []map[string]string
		for kid, pub := range keys {
			jwkList = append(jwkList, publicKeyToJWK(pub, kid))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"keys": jwkList})
	}))
	t.Cleanup(srv.Close)
	return srv.URL, &n
}

// signRS256 signs a JWT with the given RSA private key, kid, and claims.
func signRS256(t *testing.T, priv *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(priv)
	if err != nil {
		t.Fatalf("sign RS256 token: %v", err)
	}
	return s
}

// signHS256 signs a JWT with the given HS256 secret.
func signHS256(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign HS256 token: %v", err)
	}
	return s
}

// validClaims returns a MapClaims with a 1-hour expiry.
func validClaims(sub string) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
}

// expiredClaims returns a MapClaims with a past expiry.
func expiredClaims(sub string) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(-time.Hour).Unix(),
	}
}

// ── Init validation ───────────────────────────────────────────────────────────

func TestInit_EnabledFalse_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{"enabled": false}))
	if err == nil {
		t.Fatal("Init(enabled:false): expected non-nil error, got nil")
	}
}

func TestInit_TestModeEmptySecret_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":   true,
		"test_mode": true,
		// jwt_secret absent → empty string
	}))
	if err == nil {
		t.Fatal("Init(test_mode:true, jwt_secret:\"\"): expected non-nil error, got nil")
	}
}

func TestInit_NeitherConfigured_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{"enabled": true}))
	if err == nil {
		t.Fatal("Init(no jwks_url, no test_mode): expected non-nil error")
	}
}

func TestInit_TestMode_Valid(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": "super-secret",
	}))
	if err != nil {
		t.Fatalf("Init(test_mode valid): unexpected error: %v", err)
	}
	if !tv.testMode {
		t.Error("testMode should be true")
	}
}

func TestInit_JWKSMode_Valid(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":  true,
		"jwks_url": "https://auth.example.com/.well-known/jwks.json",
	}))
	if err != nil {
		t.Fatalf("Init(jwks_url valid): unexpected error: %v", err)
	}
	if tv.jwksVal == nil {
		t.Error("jwksVal should be initialised")
	}
}

func TestInit_CacheTTLDefault(t *testing.T) {
	tv := &TokenValidator{}
	_ = tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":  true,
		"jwks_url": "https://auth.example.com/.well-known/jwks.json",
	}))
	if tv.jwksVal.ttl != 600*time.Second {
		t.Errorf("default TTL: got %v, want 600s", tv.jwksVal.ttl)
	}
}

// ── Name ─────────────────────────────────────────────────────────────────────

func TestName(t *testing.T) {
	tv := &TokenValidator{}
	if got := tv.Name(); got != "token-validator" {
		t.Errorf("Name(): got %q, want token-validator", got)
	}
}

// ── HS256 happy path ──────────────────────────────────────────────────────────

func newHS256Plugin(t *testing.T, secret string) *TokenValidator {
	t.Helper()
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
	})); err != nil {
		t.Fatalf("Init HS256: %v", err)
	}
	return tv
}

func TestHandler_HS256_ValidToken_CallsNext(t *testing.T) {
	const secret = "testsecret"
	tv := newHS256Plugin(t, secret)

	tok := signHS256(t, secret, validClaims("alice"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/things", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	tv.Handler(upstream).ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("next must be called for valid HS256 token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestHandler_HS256_ClaimsPropagated(t *testing.T) {
	const secret = "testsecret"
	tv := newHS256Plugin(t, secret)

	claims := jwt.MapClaims{
		"sub":       "alice",
		"exp":       time.Now().Add(time.Hour).Unix(),
		"roles":     []interface{}{"admin", "editor"},
		"tenant_id": "acme",
	}
	tok := signHS256(t, secret, claims)

	var gotSub, gotTenant string
	var gotRoles []string
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotSub = reqctx.ActorSubject(r.Context())
		gotRoles = reqctx.ActorRoles(r.Context())
		gotTenant = reqctx.TenantID(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	tv.Handler(upstream).ServeHTTP(httptest.NewRecorder(), req)

	if gotSub != "alice" {
		t.Errorf("ActorSubject: got %q, want alice", gotSub)
	}
	if len(gotRoles) != 2 || gotRoles[0] != "admin" {
		t.Errorf("ActorRoles: got %v, want [admin editor]", gotRoles)
	}
	if gotTenant != "acme" {
		t.Errorf("TenantID: got %q, want acme", gotTenant)
	}
}

// ── Handler — validation failures: 403, next NOT called ──────────────────────

func assertForbidden(t *testing.T, rr *httptest.ResponseRecorder, nextCalled bool) {
	t.Helper()
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	if nextCalled {
		t.Error("next must NOT be called on validation failure")
	}
	ct := rr.Header().Get("Content-Type")
	if ct != response.ContentTypeError {
		t.Errorf("Content-Type: got %q, want %q", ct, response.ContentTypeError)
	}
}

func TestHandler_MissingToken_Returns403(t *testing.T) {
	tv := newHS256Plugin(t, "s")

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil) // no Authorization header
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

func TestHandler_HS256_TamperedToken_Returns403(t *testing.T) {
	const secret = "testsecret"
	tv := newHS256Plugin(t, secret)

	tok := signHS256(t, secret, validClaims("bob"))
	tampered := tok[:len(tok)-4] + "XXXX"

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

func TestHandler_HS256_ExpiredToken_Returns403(t *testing.T) {
	const secret = "testsecret"
	tv := newHS256Plugin(t, secret)

	tok := signHS256(t, secret, expiredClaims("carol"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

// ── 403 trace.correlationId populated ────────────────────────────────────────

func TestHandler_403_TraceCorrelationID(t *testing.T) {
	const secret = "testsecret"
	tv := newHS256Plugin(t, secret)

	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header → 403.
	// Set correlation ID via reqctx on the incoming request context.
	ctx := reqctx.WithCorrelationID(req.Context(), "corr-abc-123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Trace.CorrelationID != "corr-abc-123" {
		t.Errorf("trace.correlationId: got %q, want %q",
			body.Trace.CorrelationID, "corr-abc-123")
	}
}

func TestHandler_403_TraceCorrelationID_FallbackHeader(t *testing.T) {
	tv := newHS256Plugin(t, "s")
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "from-header-456")
	// No reqctx correlation ID set; falls back to the header.
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Trace.CorrelationID != "from-header-456" {
		t.Errorf("trace.correlationId fallback: got %q, want from-header-456",
			body.Trace.CorrelationID)
	}
}

// ── JWKS happy path ───────────────────────────────────────────────────────────

func newJWKSPlugin(t *testing.T, jwksURL string) *TokenValidator {
	t.Helper()
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":           true,
		"jwks_url":          jwksURL,
		"cache_ttl_seconds": 600,
	})); err != nil {
		t.Fatalf("Init JWKS: %v", err)
	}
	return tv
}

func TestHandler_JWKS_ValidToken_CallsNext(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "k1"
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})
	tv := newJWKSPlugin(t, srvURL)

	tok := signRS256(t, priv, kid, validClaims("dave"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	tv.Handler(upstream).ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("next must be called for valid RS256 token")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

// ── JWKS cache hit-counter ────────────────────────────────────────────────────

func TestHandler_JWKS_Cache_SingleFetch(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "k2"
	srvURL, serverCalls := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})
	tv := newJWKSPlugin(t, srvURL)

	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 5; i++ {
		tok := signRS256(t, priv, kid, validClaims("eve"))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		tv.Handler(upstream).ServeHTTP(httptest.NewRecorder(), req)
	}

	// Only one JWKS fetch should have occurred — subsequent requests serve from cache.
	if n := tv.jwksVal.hitCount(); n != 1 {
		t.Errorf("JWKS fetch count: got %d, want 1 (cache must serve subsequent requests)", n)
	}
	if n := serverCalls.Load(); n != 1 {
		t.Errorf("server hit count: got %d, want 1", n)
	}
}

// ── JWKS validation failures ──────────────────────────────────────────────────

func TestHandler_JWKS_TamperedToken_Returns403(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "k3"
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})
	tv := newJWKSPlugin(t, srvURL)

	tok := signRS256(t, priv, kid, validClaims("frank"))
	tampered := tok[:len(tok)-4] + "ZZZZ"

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

func TestHandler_JWKS_ExpiredToken_Returns403(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "k4"
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})
	tv := newJWKSPlugin(t, srvURL)

	tok := signRS256(t, priv, kid, expiredClaims("grace"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

func TestHandler_JWKS_UnknownKID_Returns403(t *testing.T) {
	priv := testRSAKey(t)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{"known-kid": &priv.PublicKey})
	tv := newJWKSPlugin(t, srvURL)

	// Sign with a kid that is NOT in the JWKS.
	tok := signRS256(t, priv, "unknown-kid", validClaims("han"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

// ── Audience validation ───────────────────────────────────────────────────────

func TestHandler_HS256_AudienceMismatch_Returns403(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	_ = tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"audience":   "expected-audience",
	}))

	claims := jwt.MapClaims{
		"sub": "ivan",
		"exp": time.Now().Add(time.Hour).Unix(),
		"aud": "wrong-audience",
	}
	tok := signHS256(t, secret, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertForbidden(t, rr, nextCalled)
}

// ── JWKS stale-while-revalidate ───────────────────────────────────────────────

func TestJWKS_StaleWhileRevalidate(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "sk1"

	// Server that can be stopped mid-test.
	var serverDown atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if serverDown.Load() {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]string{publicKeyToJWK(&priv.PublicKey, kid)},
		})
	}))
	t.Cleanup(srv.Close)

	v := newJWKSValidator(srv.URL, 600*time.Second)

	// First call: fetches JWKS and caches it.
	tok := signRS256(t, priv, kid, validClaims("judy"))
	if _, err := v.validate(tok, ""); err != nil {
		t.Fatalf("first validate: %v", err)
	}
	if v.hitCount() != 1 {
		t.Fatalf("expected 1 fetch after first validate, got %d", v.hitCount())
	}

	// Force TTL expiry so the next call triggers a refresh.
	v.mu.Lock()
	v.fetchAt = time.Now().Add(-700 * time.Second)
	v.mu.Unlock()

	// Server goes down — refresh will fail.
	serverDown.Store(true)

	// Second call: refresh fails; stale-while-revalidate should serve the cached key.
	tok2 := signRS256(t, priv, kid, validClaims("judy"))
	_, err := v.validate(tok2, "")
	if err != nil {
		t.Errorf("stale-while-revalidate: expected success using stale keys, got error: %v", err)
	}
}

// ── Shutdown ──────────────────────────────────────────────────────────────────

func TestShutdown_NoError(t *testing.T) {
	tv := newHS256Plugin(t, "s")
	if err := tv.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}
