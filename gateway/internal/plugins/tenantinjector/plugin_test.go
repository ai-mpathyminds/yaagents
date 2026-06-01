// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package tenantinjector_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/plugins/tenantinjector"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// newPlugin returns an initialised TenantInjector via the Plugin interface.
func newPlugin(t *testing.T, cfg map[string]any) plugin.Plugin {
	t.Helper()
	p := &tenantinjector.TenantInjector{}
	if err := p.Init(plugin.NewMapConfig(cfg)); err != nil {
		t.Fatalf("Init: unexpected error: %v", err)
	}
	return p
}

// stdCfg is a valid minimal configuration.
func stdCfg() map[string]any {
	return map[string]any{
		"header":        "X-Tenant-ID",
		"inject_header": "X-Actor-Tenant",
	}
}

// ── Init validation ───────────────────────────────────────────────────────────

func TestInit_MissingHeader(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	err := p.Init(plugin.NewMapConfig(map[string]any{
		"inject_header": "X-Actor-Tenant",
	}))
	if err == nil {
		t.Fatal("Init with missing header: expected non-nil error, got nil")
	}
}

func TestInit_EmptyHeader(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	err := p.Init(plugin.NewMapConfig(map[string]any{
		"header":        "",
		"inject_header": "X-Actor-Tenant",
	}))
	if err == nil {
		t.Fatal("Init with empty header: expected non-nil error, got nil")
	}
}

func TestInit_MissingInjectHeader(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	err := p.Init(plugin.NewMapConfig(map[string]any{
		"header": "X-Tenant-ID",
	}))
	if err == nil {
		t.Fatal("Init with missing inject_header: expected non-nil error, got nil")
	}
}

func TestInit_EmptyInjectHeader(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	err := p.Init(plugin.NewMapConfig(map[string]any{
		"header":        "X-Tenant-ID",
		"inject_header": "",
	}))
	if err == nil {
		t.Fatal("Init with empty inject_header: expected non-nil error, got nil")
	}
}

func TestInit_Valid(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	err := p.Init(plugin.NewMapConfig(stdCfg()))
	if err != nil {
		t.Fatalf("Init with valid config: unexpected error: %v", err)
	}
}

func TestInit_ValidWithAllowlist(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	cfg := stdCfg()
	cfg["allowlist"] = []string{"acme", "globex"}
	err := p.Init(plugin.NewMapConfig(cfg))
	if err != nil {
		t.Fatalf("Init with allowlist: unexpected error: %v", err)
	}
}

// ── Name ──────────────────────────────────────────────────────────────────────

func TestName(t *testing.T) {
	p := &tenantinjector.TenantInjector{}
	if got := p.Name(); got != "tenant-injector" {
		t.Errorf("Name(): got %q, want %q", got, "tenant-injector")
	}
}

// ── Handler — allowlist empty (all tenants accepted) ─────────────────────────

func TestHandler_AllowlistEmpty_Accepted(t *testing.T) {
	p := newPlugin(t, stdCfg())

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "any-tenant")
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if !upstreamCalled {
		t.Error("next handler must be called when allowlist is empty")
	}
}

// Empty tenant ID with empty allowlist still passes through.
func TestHandler_AllowlistEmpty_EmptyTenantID(t *testing.T) {
	p := newPlugin(t, stdCfg())

	var called bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	p.Handler(upstream).ServeHTTP(rr, req)

	if !called {
		t.Error("upstream must be called when allowlist is empty, even with no tenant header")
	}
}

// ── Handler — allowlist non-empty, tenant in list ─────────────────────────────

func TestHandler_AllowlistHit_Accepted(t *testing.T) {
	cfg := stdCfg()
	cfg["allowlist"] = []string{"acme", "globex"}
	p := newPlugin(t, cfg)

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "acme")
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if !upstreamCalled {
		t.Error("next must be called for a listed tenant")
	}
}

// ── Handler — allowlist non-empty, tenant NOT in list → 403 ──────────────────

func TestHandler_AllowlistMiss_Returns403(t *testing.T) {
	cfg := stdCfg()
	cfg["allowlist"] = []string{"acme", "globex"}
	p := newPlugin(t, cfg)

	var upstreamCalled bool
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		upstreamCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "intruder")
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	if upstreamCalled {
		t.Error("next must NOT be called on allowlist miss")
	}

	// Verify vendor error content-type.
	ct := rr.Header().Get("Content-Type")
	if ct != response.ContentTypeError {
		t.Errorf("Content-Type: got %q, want %q", ct, response.ContentTypeError)
	}

	// Verify the body decodes as ErrorBody with type "forbidden".
	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Type != "forbidden" {
		t.Errorf("ErrorBody.Type: got %q, want %q", body.Type, "forbidden")
	}
	if body.Code != "tenant_not_allowed" {
		t.Errorf("ErrorBody.Code: got %q, want %q", body.Code, "tenant_not_allowed")
	}
}

// Correlation-id propagates into the 403 trace.
func TestHandler_AllowlistMiss_TracePopulated(t *testing.T) {
	cfg := stdCfg()
	cfg["allowlist"] = []string{"acme"}
	p := newPlugin(t, cfg)

	upstream := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "unknown")
	ctx := reqctx.WithCorrelationID(req.Context(), "test-corr-id-123")
	ctx = reqctx.WithRequestID(ctx, "req-id-456")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Trace.CorrelationID != "test-corr-id-123" {
		t.Errorf("Trace.CorrelationID: got %q, want %q", body.Trace.CorrelationID, "test-corr-id-123")
	}
	if body.Trace.RequestID != "req-id-456" {
		t.Errorf("Trace.RequestID: got %q, want %q", body.Trace.RequestID, "req-id-456")
	}
}

// ── Handler — inject_header present in upstream request ───────────────────────

func TestHandler_InjectHeader_PresentInUpstream(t *testing.T) {
	p := newPlugin(t, stdCfg())

	var gotHeader string
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Actor-Tenant")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "globex")
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if gotHeader != "globex" {
		t.Errorf("inject_header in upstream: got %q, want %q", gotHeader, "globex")
	}
}

// inject_header is set even when the source tenant header is absent (empty string).
func TestHandler_InjectHeader_EmptyTenant(t *testing.T) {
	p := newPlugin(t, stdCfg())

	var gotHeader string
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Actor-Tenant")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No X-Tenant-ID header set → tenant ID is ""
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if gotHeader != "" {
		t.Errorf("inject_header with no source header: got %q, want empty", gotHeader)
	}
}

// Custom header and inject_header names are honoured.
func TestHandler_CustomHeaders(t *testing.T) {
	p := newPlugin(t, map[string]any{
		"header":        "X-My-Tenant",
		"inject_header": "X-Forwarded-Tenant",
	})

	var gotHeader string
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Forwarded-Tenant")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-My-Tenant", "custom-corp")
	rr := httptest.NewRecorder()

	p.Handler(upstream).ServeHTTP(rr, req)

	if gotHeader != "custom-corp" {
		t.Errorf("custom inject_header: got %q, want %q", gotHeader, "custom-corp")
	}
}

// ── Shutdown ───────────────────────────────────────────────────────────────────

func TestShutdown_NoError(t *testing.T) {
	p := newPlugin(t, stdCfg())
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}
