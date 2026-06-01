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

// ═══════════════════════════════════════════════════════════════════════════════
// PLG-3b / ADR PI2-yaa-0007 — v2 tests (9 amendments)
// ═══════════════════════════════════════════════════════════════════════════════

// ── v2 test helpers ───────────────────────────────────────────────────────────

// makeIssuers builds the []any form expected by plugin.NewMapConfig for an
// issuers: list.
func makeIssuers(entries ...map[string]any) []any {
	out := make([]any, len(entries))
	for i, e := range entries {
		out[i] = e
	}
	return out
}

// newV2JWKSPlugin builds a TokenValidator in v2 mode using the given issuers
// list plus optional config overrides. The caller supplies a pre-built
// makeIssuers(…) slice.
func newV2JWKSPlugin(t *testing.T, issuers []any, overrides map[string]any) *TokenValidator {
	t.Helper()
	cfg := map[string]any{
		"enabled":              true,
		"issuers":              issuers,
		"algorithms":           []string{"RS256"},
		"clock_skew_seconds":   0,
		"required_claims":      []string{"sub"},
		"propagate_claims":     map[string]any{"mode": "all"},
		"token":                map[string]any{"header": "Authorization", "scheme": "Bearer"},
		"max_token_bytes":      8192,
	}
	for k, v := range overrides {
		cfg[k] = v
	}
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(cfg)); err != nil {
		t.Fatalf("newV2JWKSPlugin: Init: %v", err)
	}
	return tv
}

// assertStatus verifies status code and that next was or was not called.
func assertStatus(t *testing.T, rr *httptest.ResponseRecorder, want int, nextCalled bool, expectNext bool) {
	t.Helper()
	if rr.Code != want {
		t.Errorf("status: got %d, want %d", rr.Code, want)
	}
	if nextCalled != expectNext {
		t.Errorf("nextCalled: got %v, want %v", nextCalled, expectNext)
	}
}

// ── Amendment 1: Multi-issuer ─────────────────────────────────────────────────

func TestInit_V2_Issuers_Valid(t *testing.T) {
	priv := testRSAKey(t)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{"k": &priv.PublicKey})
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled": true,
		"issuers": makeIssuers(map[string]any{
			"issuer":  "https://iam.example.com",
			"jwks_url": srvURL,
		}),
	}))
	if err != nil {
		t.Fatalf("Init with valid issuers: unexpected error: %v", err)
	}
	if tv.issuersPool == nil || len(tv.issuersPool.entries) != 1 {
		t.Fatal("issuersPool should have 1 entry")
	}
}

func TestInit_V2_Issuers_Empty_TestModeFalse_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled": true,
		"issuers": []any{}, // explicitly set but empty
	}))
	if err == nil {
		t.Fatal("expected non-nil error for empty issuers + test_mode:false")
	}
}

func TestInit_V2_Issuers_MissingJWKSURL_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled": true,
		"issuers": makeIssuers(map[string]any{
			"issuer": "https://iam.example.com",
			// jwks_url absent
		}),
	}))
	if err == nil {
		t.Fatal("expected non-nil error when jwks_url is missing from issuer entry")
	}
}

func TestHandler_V2_MultiIssuer_MatchFirst(t *testing.T) {
	priv1 := testRSAKey(t)
	priv2 := testRSAKey(t)
	const (
		iss1 = "https://iam1.example.com"
		iss2 = "https://iam2.example.com"
		kid1 = "k1"
		kid2 = "k2"
	)
	url1, _ := jwksServer(t, map[string]*rsa.PublicKey{kid1: &priv1.PublicKey})
	url2, _ := jwksServer(t, map[string]*rsa.PublicKey{kid2: &priv2.PublicKey})

	tv := newV2JWKSPlugin(t, makeIssuers(
		map[string]any{"issuer": iss1, "jwks_url": url1},
		map[string]any{"issuer": iss2, "jwks_url": url2},
	), nil)

	claims := jwt.MapClaims{"sub": "alice", "exp": time.Now().Add(time.Hour).Unix(), "iss": iss1}
	tok := signRS256(t, priv1, kid1, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

func TestHandler_V2_MultiIssuer_MatchSecond(t *testing.T) {
	priv1 := testRSAKey(t)
	priv2 := testRSAKey(t)
	const (
		iss1 = "https://iam1.example.com"
		iss2 = "https://iam2.example.com"
		kid1 = "k1"
		kid2 = "k2"
	)
	url1, _ := jwksServer(t, map[string]*rsa.PublicKey{kid1: &priv1.PublicKey})
	url2, _ := jwksServer(t, map[string]*rsa.PublicKey{kid2: &priv2.PublicKey})

	tv := newV2JWKSPlugin(t, makeIssuers(
		map[string]any{"issuer": iss1, "jwks_url": url1},
		map[string]any{"issuer": iss2, "jwks_url": url2},
	), nil)

	claims := jwt.MapClaims{"sub": "bob", "exp": time.Now().Add(time.Hour).Unix(), "iss": iss2}
	tok := signRS256(t, priv2, kid2, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

func TestHandler_V2_MultiIssuer_UnknownIssuer_Returns401(t *testing.T) {
	priv := testRSAKey(t)
	const (
		knownIss = "https://iam.example.com"
		kid      = "k"
	)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})
	tv := newV2JWKSPlugin(t, makeIssuers(
		map[string]any{"issuer": knownIss, "jwks_url": srvURL},
	), nil)

	claims := jwt.MapClaims{"sub": "charlie", "exp": time.Now().Add(time.Hour).Unix(), "iss": "https://unknown.example.com"}
	tok := signRS256(t, priv, kid, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

// ── Amendment 2: Algorithm allowlist ─────────────────────────────────────────

func TestInit_V2_AlgorithmsNone_Error(t *testing.T) {
	priv := testRSAKey(t)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{"k": &priv.PublicKey})
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled": true,
		"issuers": makeIssuers(map[string]any{
			"issuer":  "https://iam.example.com",
			"jwks_url": srvURL,
		}),
		"algorithms": []string{"none", "RS256"},
	}))
	if err == nil {
		t.Fatal("expected non-nil error when algorithms contains 'none'")
	}
}

func TestHandler_V2_DisallowedAlgorithm_HS256_Returns401(t *testing.T) {
	// HS256 token submitted to a v2 non-test-mode plugin (algorithms: [RS256])
	priv := testRSAKey(t)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{"k": &priv.PublicKey})
	tv := newV2JWKSPlugin(t, makeIssuers(
		map[string]any{"issuer": "https://iam.example.com", "jwks_url": srvURL},
	), map[string]any{"algorithms": []string{"RS256"}})

	const secret = "s"
	tok := signHS256(t, secret, validClaims("dave"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

// ── Amendment 3: Multi-audience ───────────────────────────────────────────────

func TestHandler_V2_MultiAudience_Match(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"audiences":  []string{"b", "c"},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "eve",
		"exp": time.Now().Add(time.Hour).Unix(),
		"aud": []interface{}{"a", "b"}, // "b" is in configured audiences
	}
	tok := signHS256(t, secret, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

func TestHandler_V2_MultiAudience_Mismatch_Returns401(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"audiences":  []string{"a", "b"},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	claims := jwt.MapClaims{
		"sub": "frank",
		"exp": time.Now().Add(time.Hour).Unix(),
		"aud": "x", // no match
	}
	tok := signHS256(t, secret, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

// ── Amendment 4: Clock skew ───────────────────────────────────────────────────

func TestHandler_V2_ClockSkew_WithinTolerance(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":            true,
		"test_mode":          true,
		"jwt_secret":         secret,
		"clock_skew_seconds": 60, // 60 s leeway
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Token expired 30 s ago — within the 60 s clock skew window.
	claims := jwt.MapClaims{
		"sub": "grace",
		"exp": time.Now().Add(-30 * time.Second).Unix(),
	}
	tok := signHS256(t, secret, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

func TestHandler_V2_ClockSkew_Exceeded_Returns401(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":            true,
		"test_mode":          true,
		"jwt_secret":         secret,
		"clock_skew_seconds": 10, // only 10 s leeway
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Token expired 30 s ago — exceeds the 10 s window.
	claims := jwt.MapClaims{
		"sub": "hank",
		"exp": time.Now().Add(-30 * time.Second).Unix(),
	}
	tok := signHS256(t, secret, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

// ── Amendment 5: Required claims ──────────────────────────────────────────────

func TestHandler_V2_RequiredClaims_Missing_Returns401(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":         true,
		"test_mode":       true,
		"jwt_secret":      secret,
		"required_claims": []string{"sub", "tenant_id"},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// token has sub but NOT tenant_id
	tok := signHS256(t, secret, validClaims("iris"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

// ── Amendment 6: Propagate-claims ─────────────────────────────────────────────

func TestHandler_V2_PropagateClaims_All(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":          true,
		"test_mode":        true,
		"jwt_secret":       secret,
		"propagate_claims": map[string]any{"mode": "all"},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	claims := jwt.MapClaims{
		"sub":   "judy",
		"email": "judy@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	tok := signHS256(t, secret, claims)

	var gotClaims map[string]any
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotClaims = reqctx.JWTClaims(r.Context())
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	tv.Handler(upstream).ServeHTTP(httptest.NewRecorder(), req)

	if _, ok := gotClaims["email"]; !ok {
		t.Error("mode:all should propagate email claim")
	}
	if _, ok := gotClaims["sub"]; !ok {
		t.Error("mode:all should propagate sub claim")
	}
}

func TestHandler_V2_PropagateClaims_Allowlist(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"propagate_claims": map[string]any{
			"mode":   "allowlist",
			"claims": []any{"sub", "email"},
		},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	claims := jwt.MapClaims{
		"sub":       "karen",
		"email":     "karen@example.com",
		"tenant_id": "acme", // should NOT be propagated
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
	tok := signHS256(t, secret, claims)

	var gotClaims map[string]any
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotClaims = reqctx.JWTClaims(r.Context())
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	tv.Handler(upstream).ServeHTTP(httptest.NewRecorder(), req)

	if _, ok := gotClaims["sub"]; !ok {
		t.Error("allowlist should include sub")
	}
	if _, ok := gotClaims["email"]; !ok {
		t.Error("allowlist should include email")
	}
	if _, ok := gotClaims["tenant_id"]; ok {
		t.Error("allowlist must NOT include tenant_id")
	}
}

func TestInit_V2_PropagateAllowlist_EmptyClaims_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": "s",
		"propagate_claims": map[string]any{
			"mode":   "allowlist",
			"claims": []any{}, // empty → error
		},
	}))
	if err == nil {
		t.Fatal("expected non-nil error for mode:allowlist with empty claims")
	}
}

// ── Amendment 7: Configurable token header ────────────────────────────────────

func TestHandler_V2_CustomHeader_NoScheme(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"token": map[string]any{
			"header": "X-Auth-Token",
			"scheme": "", // no prefix strip
		},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	tok := signHS256(t, secret, validClaims("lena"))

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Auth-Token", tok) // raw token, no "Bearer" prefix
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

// ── Amendment 8: RFC-correct status codes ─────────────────────────────────────

func TestHandler_V2_Default_MissingToken_Returns401(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"audiences":  []string{}, // v2 config trigger
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil) // no token
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusUnauthorized, nextCalled, false)
}

func TestHandler_V2_OnFailure_Override_Returns403(t *testing.T) {
	// Operator overrides default 401 → 403 to preserve v1 semantics.
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"test_mode":  true,
		"jwt_secret": secret,
		"audiences":  []string{},
		"on_failure": map[string]any{
			"missing_token": 403,
			"expired":       403,
		},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusForbidden, nextCalled, false)
}

// ── Amendment 9: Token size cap ───────────────────────────────────────────────

func TestHandler_V2_MaxTokenBytes_Exceeded_Returns400(t *testing.T) {
	const secret = "sec"
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":         true,
		"test_mode":       true,
		"jwt_secret":      secret,
		"audiences":       []string{},
		"max_token_bytes": 64, // tiny cap to force failure
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Build a token that exceeds 64 bytes (any normal HS256 token is longer).
	tok := signHS256(t, secret, validClaims("mike"))
	if len(tok) <= 64 {
		t.Skip("token too short for this test — adjust cap")
	}

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusBadRequest, nextCalled, false)
	ct := rr.Header().Get("Content-Type")
	if ct != response.ContentTypeError {
		t.Errorf("Content-Type: got %q, want %q", ct, response.ContentTypeError)
	}
}

// ── V1 backwards compat shim ──────────────────────────────────────────────────

func TestInit_V2_V1Shim_JWKSURLLoadsCleanly(t *testing.T) {
	// v1 config (jwks_url + audience) must load even when on_failure (a v2 key)
	// is also present — the shim fires, WARN is emitted, behaviour is v1-compat.
	priv := testRSAKey(t)
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{"k": &priv.PublicKey})
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":    true,
		"jwks_url":   srvURL,
		"audience":   "myapp",
		"on_failure": map[string]any{"missing_token": 401}, // trigger v2 mode
	}))
	if err != nil {
		t.Fatalf("v1 compat shim: unexpected Init error: %v", err)
	}
	// jwksVal must be set so existing tests that access it continue to work.
	if tv.jwksVal == nil {
		t.Error("v1 shim must set jwksVal")
	}
}

func TestHandler_V2_V1Shim_ValidToken_Passes(t *testing.T) {
	priv := testRSAKey(t)
	const kid = "k"
	srvURL, _ := jwksServer(t, map[string]*rsa.PublicKey{kid: &priv.PublicKey})

	// Use the v1-compat shim path: jwks_url + audiences (v2 key trigger).
	tv2 := &TokenValidator{}
	if err := tv2.Init(plugin.NewMapConfig(map[string]any{
		"enabled":   true,
		"jwks_url":  srvURL,
		"audiences": []string{}, // v2 key trigger
	})); err != nil {
		t.Fatalf("Init v1 shim: %v", err)
	}

	claims := jwt.MapClaims{"sub": "nina", "exp": time.Now().Add(time.Hour).Unix()}
	tok := signRS256(t, priv, kid, claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv2.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusOK, nextCalled, true)
}

// ── Cold-start: all JWKS unreachable → 503 ───────────────────────────────────

func TestHandler_V2_JWKSUnavailable_Returns503(t *testing.T) {
	// Point to a URL that is guaranteed unreachable (no server started).
	tv := &TokenValidator{}
	if err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled": true,
		"issuers": makeIssuers(map[string]any{
			"issuer":  "https://iam.example.com",
			"jwks_url": "http://127.0.0.1:19999/not-reachable",
		}),
		"algorithms": []string{"RS256"},
	})); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Build a minimal plausible RS256-header token (won't pass signature check,
	// but the JWKS fetch happens first and returns an unavailable error).
	priv := testRSAKey(t)
	claims := jwt.MapClaims{"sub": "oscar", "exp": time.Now().Add(time.Hour).Unix(), "iss": "https://iam.example.com"}
	tok := signRS256(t, priv, "kid", claims)

	var nextCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { nextCalled = true })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	tv.Handler(upstream).ServeHTTP(rr, req)

	assertStatus(t, rr, http.StatusServiceUnavailable, nextCalled, false)
}

// ── Init validation: clock_skew and max_token_bytes bounds ───────────────────

func TestInit_V2_ClockSkewOutOfBounds_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":            true,
		"test_mode":          true,
		"jwt_secret":         "s",
		"clock_skew_seconds": 700, // > 600
	}))
	if err == nil {
		t.Fatal("expected error for clock_skew_seconds > 600")
	}
}

func TestInit_V2_MaxTokenBytesOutOfBounds_Error(t *testing.T) {
	tv := &TokenValidator{}
	err := tv.Init(plugin.NewMapConfig(map[string]any{
		"enabled":         true,
		"test_mode":       true,
		"jwt_secret":      "s",
		"max_token_bytes": 70000, // > 65536
	}))
	if err == nil {
		t.Fatal("expected error for max_token_bytes > 65536")
	}
}
