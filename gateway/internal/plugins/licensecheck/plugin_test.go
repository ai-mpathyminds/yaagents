// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Internal package test so we can inject a custom httpClient and inspect
// unexported cache state without exporting test-only symbols.
package licensecheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// newPlugin builds a fully initialised LicenseCheck pointed at licenseURL.
// If client is non-nil it overrides the HTTP client created by Init (allows
// injecting a very-short-timeout client without touching config integers).
func newPlugin(t *testing.T, licenseURL string, client *http.Client) *LicenseCheck {
	t.Helper()
	lc := &LicenseCheck{}
	cfg := plugin.NewMapConfig(map[string]any{
		"license_url":    licenseURL,
		"header":         "X-License-Token",
		"cache_ttl_seconds": 300,
		"max_cache_size": 4, // small so LRU eviction is testable
	})
	if err := lc.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if client != nil {
		lc.httpClient = client
	}
	return lc
}

// hitCounter starts a test HTTP server that returns statusCode and counts calls.
func hitCounter(t *testing.T, statusCode int) (serverURL string, calls *atomic.Int64, close func()) {
	t.Helper()
	var n atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv.URL, &n, srv.Close
}

// hangServer starts a test HTTP server that never responds (simulates timeout).
func hangServer(t *testing.T) (serverURL string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the client disconnects (timeout fires).
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// decodeError decodes the response body as response.ErrorBody.
func decodeError(t *testing.T, rr *httptest.ResponseRecorder) response.ErrorBody {
	t.Helper()
	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return body
}

// ── Init validation ───────────────────────────────────────────────────────────

func TestInit_MissingLicenseURL(t *testing.T) {
	lc := &LicenseCheck{}
	err := lc.Init(plugin.NewMapConfig(map[string]any{}))
	if err == nil {
		t.Fatal("expected error for missing license_url, got nil")
	}
}

func TestInit_InvalidLicenseURL_NoScheme(t *testing.T) {
	lc := &LicenseCheck{}
	err := lc.Init(plugin.NewMapConfig(map[string]any{
		"license_url": "not-a-url",
	}))
	if err == nil {
		t.Fatal("expected error for URL with no scheme")
	}
}

func TestInit_InvalidLicenseURL_NoHost(t *testing.T) {
	lc := &LicenseCheck{}
	err := lc.Init(plugin.NewMapConfig(map[string]any{
		"license_url": "http://",
	}))
	if err == nil {
		t.Fatal("expected error for URL with empty host")
	}
}

func TestInit_ValidLicenseURL(t *testing.T) {
	lc := &LicenseCheck{}
	err := lc.Init(plugin.NewMapConfig(map[string]any{
		"license_url": "https://license.example.com/verify",
	}))
	if err != nil {
		t.Fatalf("Init with valid URL: unexpected error: %v", err)
	}
}

func TestInit_Defaults(t *testing.T) {
	lc := &LicenseCheck{}
	_ = lc.Init(plugin.NewMapConfig(map[string]any{
		"license_url": "https://lic.example.com",
	}))
	if lc.header != "X-License-Token" {
		t.Errorf("header default: got %q, want X-License-Token", lc.header)
	}
	if lc.cacheTTL != 300*time.Second {
		t.Errorf("cacheTTL default: got %v, want 300s", lc.cacheTTL)
	}
	if lc.maxSize != 1024 {
		t.Errorf("maxSize default: got %d, want 1024", lc.maxSize)
	}
}

func TestInit_CustomHeader(t *testing.T) {
	lc := &LicenseCheck{}
	_ = lc.Init(plugin.NewMapConfig(map[string]any{
		"license_url": "https://lic.example.com",
		"header":      "X-My-License",
	}))
	if lc.header != "X-My-License" {
		t.Errorf("header: got %q, want X-My-License", lc.header)
	}
}

// ── Name ──────────────────────────────────────────────────────────────────────

func TestName(t *testing.T) {
	lc := &LicenseCheck{}
	if got := lc.Name(); got != "license-check" {
		t.Errorf("Name(): got %q, want %q", got, "license-check")
	}
}

// ── Handler — valid token → 2xx → pass through ───────────────────────────────

func TestHandler_ValidToken_PassThrough(t *testing.T) {
	srvURL, calls, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil)

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-License-Token", "good-token")
	rr := httptest.NewRecorder()

	lc.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if !upstreamCalled {
		t.Error("next must be called for a valid token")
	}
	if calls.Load() != 1 {
		t.Errorf("license server calls: got %d, want 1", calls.Load())
	}
}

// ── Handler — invalid token → 4xx → 403 vendor-error ────────────────────────

func TestHandler_InvalidToken_Returns403(t *testing.T) {
	srvURL, _, _ := hitCounter(t, http.StatusForbidden)
	lc := newPlugin(t, srvURL, nil)

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		upstreamCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-License-Token", "bad-token")
	rr := httptest.NewRecorder()

	lc.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	if upstreamCalled {
		t.Error("next must NOT be called for invalid token")
	}
	ct := rr.Header().Get("Content-Type")
	if ct != response.ContentTypeError {
		t.Errorf("Content-Type: got %q, want %q", ct, response.ContentTypeError)
	}
	body := decodeError(t, rr)
	if body.Type != "forbidden" {
		t.Errorf("body.Type: got %q, want forbidden", body.Type)
	}
	if body.Code != "license_invalid" {
		t.Errorf("body.Code: got %q, want license_invalid", body.Code)
	}
	// dependency must be absent on HTTP non-2xx (server answered; no network failure)
	if body.Trace.Dependency != "" {
		t.Errorf("dependency: got %q, want empty for HTTP rejection", body.Trace.Dependency)
	}
}

// ── Handler — timeout → 403 with dependency: "license-server" ────────────────

func TestHandler_Timeout_Returns403WithDependency(t *testing.T) {
	srvURL := hangServer(t)
	// Override httpClient with 1ms timeout so the test finishes quickly.
	shortClient := &http.Client{Timeout: 1 * time.Millisecond}
	lc := newPlugin(t, srvURL, shortClient)

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		upstreamCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-License-Token", "tok")
	rr := httptest.NewRecorder()

	lc.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	if upstreamCalled {
		t.Error("next must NOT be called on timeout")
	}
	body := decodeError(t, rr)
	if body.Trace.Dependency != "license-server" {
		t.Errorf("dependency: got %q, want %q", body.Trace.Dependency, "license-server")
	}
}

// ── Handler — cache: same token within TTL → single outbound call ─────────────

func TestHandler_Cache_HitCounterSingleOutbound(t *testing.T) {
	srvURL, calls, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil)

	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	doReq := func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-License-Token", "cached-tok")
		rr := httptest.NewRecorder()
		lc.Handler(upstream).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("status: got %d, want 200", rr.Code)
		}
	}

	doReq() // first — cache miss → 1 outbound call
	doReq() // second — cache hit → no new outbound call
	doReq() // third — cache hit → no new outbound call

	if n := calls.Load(); n != 1 {
		t.Errorf("license server calls: got %d, want 1 (cache must serve subsequent requests)", n)
	}
}

// ── Handler — cached invalid result is also served from cache ─────────────────

func TestHandler_Cache_InvalidTokenCached(t *testing.T) {
	srvURL, calls, _ := hitCounter(t, http.StatusUnauthorized)
	lc := newPlugin(t, srvURL, nil)

	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	doReq := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-License-Token", "invalid-tok")
		rr := httptest.NewRecorder()
		lc.Handler(upstream).ServeHTTP(rr, req)
		return rr.Code
	}

	code1 := doReq() // cache miss → outbound → 401 from server → cache entry
	code2 := doReq() // cache hit → served from cache

	if code1 != http.StatusForbidden || code2 != http.StatusForbidden {
		t.Errorf("status: first=%d second=%d, both want 403", code1, code2)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("license server calls: got %d, want 1 (invalid token must be cached)", n)
	}
}

// ── Handler — correlation/request IDs propagate into error trace ──────────────

func TestHandler_TracePopulated(t *testing.T) {
	srvURL, _, _ := hitCounter(t, http.StatusUnauthorized)
	lc := newPlugin(t, srvURL, nil)

	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-License-Token", "tok")
	ctx := reqctx.WithCorrelationID(req.Context(), "corr-999")
	ctx = reqctx.WithRequestID(ctx, "req-888")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	lc.Handler(upstream).ServeHTTP(rr, req)

	body := decodeError(t, rr)
	if body.Trace.CorrelationID != "corr-999" {
		t.Errorf("CorrelationID: got %q, want %q", body.Trace.CorrelationID, "corr-999")
	}
	if body.Trace.RequestID != "req-888" {
		t.Errorf("RequestID: got %q, want %q", body.Trace.RequestID, "req-888")
	}
}

// ── Handler — different tokens use separate cache entries ─────────────────────

func TestHandler_Cache_DifferentTokensSeparateEntries(t *testing.T) {
	// Server returns 200 for all tokens.
	srvURL, calls, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil)

	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tok := range []string{"alpha", "beta", "gamma"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-License-Token", tok)
		rr := httptest.NewRecorder()
		lc.Handler(upstream).ServeHTTP(rr, req)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("calls: got %d, want 3 (one per distinct token)", n)
	}
}

// ── LRU eviction — maxSize=4, 5th distinct token evicts the LRU entry ─────────

func TestCache_LRUEviction(t *testing.T) {
	// Plugin initialised with max_cache_size: 4 (set in newPlugin helper).
	srvURL, calls, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil) // maxSize=4

	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	sendReq := func(tok string) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-License-Token", tok)
		rr := httptest.NewRecorder()
		lc.Handler(upstream).ServeHTTP(rr, req)
	}

	// Fill cache with 4 distinct tokens (t1=LRU, t4=MRU after this sequence).
	sendReq("t1") // calls=1
	sendReq("t2") // calls=2
	sendReq("t3") // calls=3
	sendReq("t4") // calls=4
	if n := calls.Load(); n != 4 {
		t.Fatalf("after filling cache: calls=%d, want 4", n)
	}

	// 5th token evicts LRU (t1). Cache is now {t5,t4,t3,t2}.
	sendReq("t5") // calls=5
	if n := calls.Load(); n != 5 {
		t.Fatalf("after 5th token: calls=%d, want 5", n)
	}

	// Verify t2..t5 are still cached — must produce zero new outbound calls.
	// (Do this BEFORE re-adding t1 to avoid cascaded evictions.)
	before := calls.Load()
	sendReq("t2")
	sendReq("t3")
	sendReq("t4")
	sendReq("t5")
	if n := calls.Load(); n != before {
		t.Errorf("t2..t5 should be cached after t5 add: %d extra calls, want 0", n-before)
	}

	// t1 was evicted — re-requesting it causes a new outbound call.
	sendReq("t1")
	if n := calls.Load(); n != before+1 {
		t.Errorf("evicted t1 re-request: calls=%d, want %d", calls.Load(), before+1)
	}
}

// ── Shutdown ───────────────────────────────────────────────────────────────────

func TestShutdown_NoError(t *testing.T) {
	srvURL, _, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil)
	if err := lc.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}

// ── cacheSet update path — updating an existing entry ────────────────────────

func TestCacheSet_UpdateExisting(t *testing.T) {
	srvURL, calls, _ := hitCounter(t, http.StatusOK)
	lc := newPlugin(t, srvURL, nil)

	// First request — cache miss → valid → cached.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-License-Token", "update-tok")
	rr1 := httptest.NewRecorder()
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	lc.Handler(upstream).ServeHTTP(rr1, req1)

	// Manually update cached entry to invalid to test the update path.
	lc.cacheSet("update-tok", false)

	// Second request — cache hit (now invalid) → 403.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-License-Token", "update-tok")
	rr2 := httptest.NewRecorder()
	lc.Handler(upstream).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusForbidden {
		t.Errorf("after cache update to invalid: got %d, want 403", rr2.Code)
	}
	// Still only 1 outbound call.
	if n := calls.Load(); n != 1 {
		t.Errorf("calls: got %d, want 1", n)
	}
}
