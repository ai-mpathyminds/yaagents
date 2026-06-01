// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// stubConfig is a minimal PluginConfig implementation for tests.
type stubConfig struct{ m map[string]any }

func newStub(m map[string]any) plugin.PluginConfig     { return &stubConfig{m} }
func (s *stubConfig) GetString(k string) string         { v, _ := s.m[k].(string); return v }
func (s *stubConfig) GetBool(k string) bool             { v, _ := s.m[k].(bool); return v }
func (s *stubConfig) GetInt(k string) int               { v, _ := s.m[k].(int); return v }
func (s *stubConfig) GetStringSlice(k string) []string {
	v, _ := s.m[k].([]string)
	return v
}
func (s *stubConfig) Raw() map[string]any { return s.m }

// makePlugin builds and inits a CORSPlugin with the given allowed origins.
func makePlugin(t *testing.T, origins []string) *CORSPlugin {
	t.Helper()
	p := &CORSPlugin{}
	cfg := newStub(map[string]any{"allowed_origins": origins})
	if err := p.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return p
}

// nextRecorder captures whether next was called and its request.
type nextRecorder struct {
	called bool
	req    *http.Request
}

func (n *nextRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.called = true
	n.req = r
	w.WriteHeader(http.StatusOK)
}

// --- Tests ---

func TestCORSPlugin_Name(t *testing.T) {
	p := &CORSPlugin{}
	if p.Name() != "cors" {
		t.Errorf("Name: want cors, got %q", p.Name())
	}
}

func TestCORSPlugin_Shutdown(t *testing.T) {
	p := &CORSPlugin{}
	if err := p.Shutdown(nil); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}

// TestCORSPlugin_EmptyOrigins_PassThrough verifies that allowed_origins: []
// disables the plugin — all requests (including OPTIONS) pass to next.
func TestCORSPlugin_EmptyOrigins_PassThrough(t *testing.T) {
	p := makePlugin(t, []string{})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	if !next.called {
		t.Error("next should be called when allowed_origins is empty (plugin disabled)")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	// No CORS headers injected.
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO header should be absent, got %q", got)
	}
}

// TestCORSPlugin_MatchedOrigin_OPTIONS_ReturnsCORSHeaders is the primary
// LLM-3 AC: matched origin + OPTIONS preflight → 200 with ACAO header.
func TestCORSPlugin_MatchedOrigin_OPTIONS_ReturnsCORSHeaders(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com"})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	// Must short-circuit: next is NOT called for preflight.
	if next.called {
		t.Error("next should NOT be called for OPTIONS preflight")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO: want https://app.example.com, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods should be set")
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers should be set")
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got == "" {
		t.Error("Access-Control-Max-Age should be set")
	}
}

// TestCORSPlugin_MismatchedOrigin_OPTIONS_NoCORSHeaders is the LLM-3 AC:
// OPTIONS from a mismatched origin → 200 but NO ACAO header.
func TestCORSPlugin_MismatchedOrigin_OPTIONS_NoCORSHeaders(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com"})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	req.Header.Set("Origin", "https://evil.attacker.com")
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	// OPTIONS is still short-circuited (not passed to next) even on mismatch.
	if next.called {
		t.Error("next should NOT be called for OPTIONS preflight")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO header must be absent for mismatched origin, got %q", got)
	}
}

// TestCORSPlugin_MatchedOrigin_NonOPTIONS_AddsACOAHeader verifies that a
// non-OPTIONS request from a matching origin gets the ACAO header + next called.
func TestCORSPlugin_MatchedOrigin_NonOPTIONS_AddsACOAHeader(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com"})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	if !next.called {
		t.Error("next should be called for non-OPTIONS requests")
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO: want https://app.example.com, got %q", got)
	}
}

// TestCORSPlugin_MismatchedOrigin_NonOPTIONS_PassThrough verifies that a
// non-OPTIONS request from a mismatched origin passes through without ACAO.
func TestCORSPlugin_MismatchedOrigin_NonOPTIONS_PassThrough(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com"})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Origin", "https://other.com")
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	if !next.called {
		t.Error("next should be called for non-OPTIONS even with mismatched origin")
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO should be absent for mismatched origin, got %q", got)
	}
}

// TestCORSPlugin_NoOriginHeader_PassThrough verifies that requests without an
// Origin header (not a cross-origin request) are forwarded unchanged.
func TestCORSPlugin_NoOriginHeader_PassThrough(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com"})
	next := &nextRecorder{}

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	// No Origin header
	w := httptest.NewRecorder()

	p.Handler(next).ServeHTTP(w, req)

	if !next.called {
		t.Error("next should be called for requests without Origin header")
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO should be absent without Origin header, got %q", got)
	}
}

// TestCORSPlugin_MultipleOrigins verifies that multiple entries in
// allowed_origins work independently.
func TestCORSPlugin_MultipleOrigins(t *testing.T) {
	p := makePlugin(t, []string{"https://app.example.com", "https://admin.example.com"})

	for _, origin := range []string{"https://app.example.com", "https://admin.example.com"} {
		next := &nextRecorder{}
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		p.Handler(next).ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Errorf("origin %s: ACAO want %q, got %q", origin, origin, got)
		}
	}
}

// TestCORSPlugin_Init_DefaultConfig verifies that Init with minimal config
// uses defaults for allow_methods, allow_headers, and max_age.
func TestCORSPlugin_Init_DefaultConfig(t *testing.T) {
	p := &CORSPlugin{}
	cfg := newStub(map[string]any{
		"allowed_origins": []string{"https://app.example.com"},
	})
	if err := p.Init(cfg); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if p.allowMethods != defaultAllowMethods {
		t.Errorf("default allowMethods: want %q, got %q", defaultAllowMethods, p.allowMethods)
	}
	if p.allowHeaders != defaultAllowHeaders {
		t.Errorf("default allowHeaders: want %q, got %q", defaultAllowHeaders, p.allowHeaders)
	}
	if p.maxAge != "86400" {
		t.Errorf("default maxAge: want 86400, got %q", p.maxAge)
	}
}

// TestCORSPlugin_Registration verifies the CORS plugin is auto-registered
// via init() (present in the global registry).
func TestCORSPlugin_Registration(t *testing.T) {
	registered := plugin.Registered()
	for _, p := range registered {
		if p.Name() == "cors" {
			return // found
		}
	}
	t.Error("cors plugin not found in global registry after init()")
}
