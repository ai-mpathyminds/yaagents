package tenant_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/auth"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/tenant"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func noopLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// runCtxMiddle wraps request through ContextMiddleware and returns the
// recorder + the captured request context from the downstream handler.
func runCtxMiddle(r *http.Request) (*httptest.ResponseRecorder, context.Context) {
	var captured context.Context
	inner := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		captured = req.Context()
		w.WriteHeader(http.StatusOK)
	})
	h := tenant.ContextMiddleware(noopLog())(inner)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, r)
	return rr, captured
}

// chainMiddles wires: ContextMiddleware → EnforceTenant(required) → next.
func chainMiddles(required bool, next http.Handler) http.Handler {
	return tenant.ContextMiddleware(noopLog())(
		tenant.EnforceTenant(required)(next),
	)
}

// ── ContextMiddleware — correlation-id ────────────────────────────────────────

func TestContextMiddleware_CorrelationID_Generated(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// No X-Correlation-ID header.
	_, ctx := runCtxMiddle(req)
	if corrID := reqctx.CorrelationID(ctx); corrID == "" {
		t.Error("correlation ID should be generated when header is absent")
	}
}

func TestContextMiddleware_CorrelationID_Preserved(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "my-corr-id")
	_, ctx := runCtxMiddle(req)
	if got := reqctx.CorrelationID(ctx); got != "my-corr-id" {
		t.Errorf("CorrelationID: got %q, want my-corr-id", got)
	}
}

func TestContextMiddleware_CorrelationID_UniquePerRequest(t *testing.T) {
	_, ctx1 := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	_, ctx2 := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	id1, id2 := reqctx.CorrelationID(ctx1), reqctx.CorrelationID(ctx2)
	if id1 == id2 {
		t.Errorf("generated correlation IDs should be unique: both %q", id1)
	}
}

// ── ContextMiddleware — request-id ────────────────────────────────────────────

func TestContextMiddleware_RequestID_AlwaysGenerated(t *testing.T) {
	_, ctx := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	if reqID := reqctx.RequestID(ctx); reqID == "" {
		t.Error("request ID should always be generated")
	}
}

func TestContextMiddleware_RequestID_UniquePerRequest(t *testing.T) {
	_, ctx1 := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	_, ctx2 := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	id1, id2 := reqctx.RequestID(ctx1), reqctx.RequestID(ctx2)
	if id1 == id2 {
		t.Errorf("request IDs should be unique per request: both %q", id1)
	}
}

// ── ContextMiddleware — tenant-id ─────────────────────────────────────────────

func TestContextMiddleware_TenantID_Extracted(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", "tenant-abc")
	_, ctx := runCtxMiddle(req)
	if got := reqctx.TenantID(ctx); got != "tenant-abc" {
		t.Errorf("TenantID: got %q, want tenant-abc", got)
	}
}

func TestContextMiddleware_TenantID_EmptyWhenAbsent(t *testing.T) {
	_, ctx := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	if got := reqctx.TenantID(ctx); got != "" {
		t.Errorf("TenantID: got %q, want empty", got)
	}
}

// ── ContextMiddleware — actor from auth.Claims ─────────────────────────────────

func TestContextMiddleware_Actor_ExtractedFromClaims(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// Simulate what auth.Middleware stores in context.
	ctx0 := context.WithValue(req.Context(), auth.ClaimsKey, auth.Claims{
		Subject: "user-xyz",
		Roles:   []string{"admin", "editor"},
	})
	_, ctx := runCtxMiddle(req.WithContext(ctx0))

	if got := reqctx.ActorSubject(ctx); got != "user-xyz" {
		t.Errorf("ActorSubject: got %q, want user-xyz", got)
	}
	roles := reqctx.ActorRoles(ctx)
	if len(roles) != 2 || roles[0] != "admin" || roles[1] != "editor" {
		t.Errorf("ActorRoles: got %v, want [admin editor]", roles)
	}
}

func TestContextMiddleware_Actor_EmptyWithoutClaims(t *testing.T) {
	_, ctx := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	if got := reqctx.ActorSubject(ctx); got != "" {
		t.Errorf("ActorSubject: got %q, want empty", got)
	}
	if roles := reqctx.ActorRoles(ctx); len(roles) != 0 {
		t.Errorf("ActorRoles: got %v, want empty", roles)
	}
}

// ── ContextMiddleware — response headers ──────────────────────────────────────

func TestContextMiddleware_ResponseHeaders_Set(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-resp")
	rr, _ := runCtxMiddle(req)

	if got := rr.Header().Get("X-Correlation-ID"); got != "corr-resp" {
		t.Errorf("response X-Correlation-ID: got %q, want corr-resp", got)
	}
	if got := rr.Header().Get("X-Request-ID"); got == "" {
		t.Error("response X-Request-ID should be set")
	}
}

func TestContextMiddleware_ResponseHeaders_GeneratedCorrID(t *testing.T) {
	rr, _ := runCtxMiddle(httptest.NewRequest("GET", "/", nil))
	// Should still echo the generated ID on the response.
	if got := rr.Header().Get("X-Correlation-ID"); got == "" {
		t.Error("response X-Correlation-ID should be echoed even when generated")
	}
}

// ── EnforceTenant ─────────────────────────────────────────────────────────────

func TestEnforceTenant_Required_Missing_Rejects(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/resource", nil)
	// No X-Tenant-ID header.
	rr := httptest.NewRecorder()
	chainMiddles(true, okHandler()).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/vnd.yaagents.error+json" {
		t.Errorf("Content-Type: got %q, want application/vnd.yaagents.error+json", ct)
	}
}

func TestEnforceTenant_Required_Missing_TracePopulated(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Correlation-ID", "corr-tenant")
	rr := httptest.NewRecorder()
	chainMiddles(true, okHandler()).ServeHTTP(rr, req)

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	tr, ok := body["trace"].(map[string]interface{})
	if !ok {
		t.Fatal("trace missing from vendor-error body")
	}
	if tr["correlationId"] != "corr-tenant" {
		t.Errorf("trace.correlationId: got %v, want corr-tenant", tr["correlationId"])
	}
	if reqID, _ := tr["requestId"].(string); reqID == "" {
		t.Error("trace.requestId should be non-empty")
	}
}

func TestEnforceTenant_Required_Present_Passes(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Tenant-ID", "tenant-ok")
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	rr := httptest.NewRecorder()
	chainMiddles(true, inner).ServeHTTP(rr, req)

	if !called {
		t.Error("next handler should be called when tenant is present")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestEnforceTenant_NotRequired_Missing_Passes(t *testing.T) {
	// No X-Tenant-ID, but not required — must not reject.
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	rr := httptest.NewRecorder()
	chainMiddles(false, inner).ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if !called {
		t.Error("next handler should be called when tenantRequired is false")
	}
}

// ── InjectUpstreamHeaders ─────────────────────────────────────────────────────

func TestInjectUpstreamHeaders_AllHeadersSet(t *testing.T) {
	upstream := httptest.NewRequest("GET", "/upstream", nil)
	ctx := upstream.Context()
	ctx = reqctx.WithCorrelationID(ctx, "corr-up")
	ctx = reqctx.WithRequestID(ctx, "req-up")
	ctx = reqctx.WithTenantID(ctx, "tenant-up")
	ctx = reqctx.WithActorSubject(ctx, "user-up")
	ctx = reqctx.WithActorRoles(ctx, []string{"role1", "role2"})
	upstream = upstream.Clone(ctx)

	tenant.InjectUpstreamHeaders(upstream)

	check := func(header, want string) {
		t.Helper()
		if got := upstream.Header.Get(header); got != want {
			t.Errorf("%s: got %q, want %q", header, got, want)
		}
	}
	check("X-Correlation-ID", "corr-up")
	check("X-Request-ID", "req-up")
	check("X-Tenant-ID", "tenant-up")
	check("X-Actor-Subject", "user-up")
	actorRoles := upstream.Header.Get("X-Actor-Roles")
	if !strings.Contains(actorRoles, "role1") || !strings.Contains(actorRoles, "role2") {
		t.Errorf("X-Actor-Roles: got %q, want role1 and role2", actorRoles)
	}
}

func TestInjectUpstreamHeaders_EmptyTenantOmitted(t *testing.T) {
	upstream := httptest.NewRequest("GET", "/", nil)
	ctx := upstream.Context()
	ctx = reqctx.WithCorrelationID(ctx, "c")
	ctx = reqctx.WithRequestID(ctx, "r")
	// No tenant, no actor.
	upstream = upstream.Clone(ctx)

	tenant.InjectUpstreamHeaders(upstream)

	if got := upstream.Header.Get("X-Tenant-ID"); got != "" {
		t.Errorf("X-Tenant-ID should be absent when empty, got %q", got)
	}
	if got := upstream.Header.Get("X-Actor-Subject"); got != "" {
		t.Errorf("X-Actor-Subject should be absent when empty, got %q", got)
	}
	if got := upstream.Header.Get("X-Actor-Roles"); got != "" {
		t.Errorf("X-Actor-Roles should be absent when empty, got %q", got)
	}
}

func TestInjectUpstreamHeaders_CorrelationAndRequestAlwaysSet(t *testing.T) {
	upstream := httptest.NewRequest("GET", "/", nil)
	ctx := upstream.Context()
	ctx = reqctx.WithCorrelationID(ctx, "corr-always")
	ctx = reqctx.WithRequestID(ctx, "req-always")
	upstream = upstream.Clone(ctx)

	tenant.InjectUpstreamHeaders(upstream)

	if got := upstream.Header.Get("X-Correlation-ID"); got != "corr-always" {
		t.Errorf("X-Correlation-ID: got %q, want corr-always", got)
	}
	if got := upstream.Header.Get("X-Request-ID"); got != "req-always" {
		t.Errorf("X-Request-ID: got %q, want req-always", got)
	}
}

// ── reqctx UUID uniqueness ─────────────────────────────────────────────────────

func TestReqctxNewUUID_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		id := reqctx.NewUUID()
		if id == "" {
			t.Fatal("NewUUID returned empty string")
		}
		if seen[id] {
			t.Fatalf("NewUUID collision at iteration %d: %q", i, id)
		}
		seen[id] = true
	}
}
