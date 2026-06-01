// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/audit"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/metrics"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"
)

// nullLog returns a discard logger for tests.
func nullLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// ctxRequest builds a request pre-populated with reqctx values so that
// ContextMiddleware is not required in unit tests.
func ctxRequest(method, path string, body io.Reader, roles []string, tenantID string) *http.Request {
	r := httptest.NewRequest(method, path, body)
	ctx := r.Context()
	ctx = reqctx.WithCorrelationID(ctx, "corr-001")
	ctx = reqctx.WithRequestID(ctx, "req-001")
	ctx = reqctx.WithTenantID(ctx, tenantID)
	ctx = reqctx.WithActorRoles(ctx, roles)
	return r.WithContext(ctx)
}

// makeDispatcher builds a RouteDispatcher with a single upstream test server.
// auditLog and reg may be nil to disable audit/metrics in a given test.
func makeDispatcher(t *testing.T, upstream *httptest.Server, routeDef routes.Route) *RouteDispatcher {
	t.Helper()
	return makeDispatcherWithObs(t, upstream, routeDef, nil, nil)
}

func makeDispatcherWithObs(t *testing.T, upstream *httptest.Server, routeDef routes.Route, auditLog *audit.Logger, reg *metrics.Registry) *RouteDispatcher {
	t.Helper()
	if upstream != nil {
		routeDef.Target = upstream.URL
	}
	d, err := New([]routes.Route{routeDef}, nullLog(), auditLog, reg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return d
}

// --- Tests ---

// TestDispatcher_MissingRole_403 verifies that an actor missing a required role
// gets a 403 with the vendor error content type.
func TestDispatcher_MissingRole_403(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r1", Method: "GET", Path: "/things", Target: upstream.URL,
		Roles: []string{"admin"},
	}
	d := makeDispatcher(t, nil, route)
	d.entries[0].route.Target = upstream.URL // already set by makeDispatcher

	req := ctxRequest("GET", "/things", nil, []string{"viewer"}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/vnd.yaagents.error+json" {
		t.Fatalf("expected vendor error CT, got %q", ct)
	}
}

// TestDispatcher_AllRolesPresent_Proxied verifies that when all required roles
// are present the request is proxied and reaches the upstream.
func TestDispatcher_AllRolesPresent_Proxied(t *testing.T) {
	reached := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r2", Method: "GET", Path: "/things", Target: upstream.URL,
		Roles: []string{"admin", "editor"},
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("GET", "/things", nil, []string{"admin", "editor", "viewer"}, "t1")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if !reached {
		t.Fatal("upstream was not reached")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestDispatcher_NoRolesRequired_Proxied verifies that a route with no roles
// is open to any authenticated actor.
func TestDispatcher_NoRolesRequired_Proxied(t *testing.T) {
	reached := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r3", Method: "GET", Path: "/public", Target: upstream.URL,
		Roles: nil,
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("GET", "/public", nil, []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if !reached {
		t.Fatal("upstream was not reached for no-roles route")
	}
}

// TestDispatcher_TypedResponsePassthrough verifies that the upstream's
// 422 application/vnd.yaagents.validation-error+json response passes through
// byte-identical (status, Content-Type, body) with X-YAAgents-Profile added.
func TestDispatcher_TypedResponsePassthrough(t *testing.T) {
	const vendorCT = "application/vnd.yaagents.validation-error+json"
	const body = `{"type":"validation_error","code":"INVALID_FIELD","message":"bad"}`

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", vendorCT)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, body)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r4", Method: "POST", Path: "/items", Target: upstream.URL,
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("POST", "/items", strings.NewReader(`{}`), []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != vendorCT {
		t.Fatalf("Content-Type mangled: want %q, got %q", vendorCT, got)
	}
	if got := w.Body.String(); got != body {
		t.Fatalf("body mangled: want %q, got %q", body, got)
	}
	if got := w.Header().Get(ProfileHeader); got != ProfileVersion {
		t.Fatalf("profile header: want %q, got %q", ProfileVersion, got)
	}
}

// TestDispatcher_MethodBodyQueryPreserved verifies that method, body, and query
// string are forwarded to upstream unchanged.
func TestDispatcher_MethodBodyQueryPreserved(t *testing.T) {
	var gotMethod, gotBody, gotQuery string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r5", Method: "POST", Path: "/cmd", Target: upstream.URL,
	}
	d := makeDispatcher(t, upstream, route)

	const payload = `{"action":"run"}`
	req := ctxRequest("POST", "/cmd?dry=true&limit=5", strings.NewReader(payload), []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if gotMethod != "POST" {
		t.Errorf("method: want POST, got %q", gotMethod)
	}
	if gotBody != payload {
		t.Errorf("body: want %q, got %q", payload, gotBody)
	}
	if gotQuery != "dry=true&limit=5" {
		t.Errorf("query: want %q, got %q", "dry=true&limit=5", gotQuery)
	}
}

// TestDispatcher_PathWithParam_Matched verifies that a route with a {param}
// placeholder matches a concrete request path segment.
func TestDispatcher_PathWithParam_Matched(t *testing.T) {
	reached := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID:     "r6",
		Method: "POST",
		Path:   "/campaigns/{campaignId}/optimizations",
		Target: upstream.URL,
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("POST", "/campaigns/cmp-123/optimizations", nil, []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if !reached {
		t.Fatal("upstream not reached for parameterised path")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// TestDispatcher_RouteNotFound_404 verifies that an unmatched request gets a
// 404 vendor-error body (not a generic 404).
func TestDispatcher_RouteNotFound_404(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r7", Method: "GET", Path: "/things", Target: upstream.URL,
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("GET", "/no-such-path", nil, []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/vnd.yaagents.error+json" {
		t.Fatalf("expected vendor error CT on 404, got %q", ct)
	}
}

// TestDispatcher_TenantRequired_Missing_403 verifies that a route with
// tenantRequired=true rejects requests missing X-Tenant-ID with a 403.
func TestDispatcher_TenantRequired_Missing_403(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r8", Method: "GET", Path: "/tenant-only", Target: upstream.URL,
		TenantRequired: true,
	}
	d := makeDispatcher(t, upstream, route)

	// No tenant ID in context.
	req := ctxRequest("GET", "/tenant-only", nil, []string{}, "" /* empty tenantID */)
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/vnd.yaagents.error+json" {
		t.Fatalf("expected vendor error CT, got %q", ct)
	}
}

// TestDispatcher_ProfileHeader_OnEveryResponse verifies that X-YAAgents-Profile
// is present on every successfully proxied response.
func TestDispatcher_ProfileHeader_OnEveryResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer upstream.Close()

	route := routes.Route{
		ID: "r9", Method: "POST", Path: "/res", Target: upstream.URL,
	}
	d := makeDispatcher(t, upstream, route)

	req := ctxRequest("POST", "/res", nil, []string{}, "")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if got := w.Header().Get(ProfileHeader); got != ProfileVersion {
		t.Fatalf("X-YAAgents-Profile: want %q, got %q", ProfileVersion, got)
	}
}

// --- Unit tests for helpers ---

func TestMatchPath_Literal(t *testing.T) {
	if !matchPath("/campaigns/list", "/campaigns/list") {
		t.Error("identical literal paths should match")
	}
	if matchPath("/campaigns/list", "/campaigns/list/extra") {
		t.Error("different segment count should not match")
	}
}

func TestMatchPath_Param(t *testing.T) {
	if !matchPath("/campaigns/{id}", "/campaigns/cmp-001") {
		t.Error("{id} should match concrete segment")
	}
	if matchPath("/campaigns/{id}", "/campaigns/") {
		t.Error("{id} should not match empty segment")
	}
}

func TestMatchPath_MultiParam(t *testing.T) {
	if !matchPath("/a/{b}/c/{d}", "/a/1/c/2") {
		t.Error("multiple params should match")
	}
	if matchPath("/a/{b}/c/{d}", "/a/1/X/2") {
		t.Error("literal mismatch should fail")
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"/foo/bar", []string{"foo", "bar"}},
		{"/foo/bar/", []string{"foo", "bar"}},
		{"foo/bar", []string{"foo", "bar"}},
		{"/", []string{}},
	}
	for _, c := range cases {
		got := splitPath(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitPath(%q): len %d, want %d", c.in, len(got), len(c.want))
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("splitPath(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestEnforceRoles_AllPresent(t *testing.T) {
	req := ctxRequest("GET", "/", nil, []string{"admin", "editor"}, "")
	if err := enforceRoles(req, []string{"admin"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEnforceRoles_Missing(t *testing.T) {
	req := ctxRequest("GET", "/", nil, []string{"viewer"}, "")
	if err := enforceRoles(req, []string{"admin"}); err == nil {
		t.Error("expected error for missing role")
	}
}

func TestEnforceRoles_Empty(t *testing.T) {
	req := ctxRequest("GET", "/", nil, nil, "")
	if err := enforceRoles(req, nil); err != nil {
		t.Errorf("empty required roles should always pass: %v", err)
	}
}

func TestBuildProxy_DirectorSetsSchemeAndHost(t *testing.T) {
	targetURL, _ := url.Parse("http://backend.internal:9090")
	var gotScheme, gotHost string

	p := buildProxy(targetURL, routes.Route{ID: "x", Method: "GET", Path: "/p", Target: targetURL.String()}, nullLog())

	req := ctxRequest("GET", "/p", nil, nil, "")
	p.Director(req)
	gotScheme = req.URL.Scheme
	gotHost = req.URL.Host

	if gotScheme != "http" {
		t.Errorf("scheme: want http, got %q", gotScheme)
	}
	if gotHost != "backend.internal:9090" {
		t.Errorf("host: want backend.internal:9090, got %q", gotHost)
	}
}

// --- Audit + Metrics observation tests (WI-1yaa.GW-5) ---

// TestDispatcher_AuditTrue_EmitsEvent verifies that a route with audit:true
// causes one JSON audit event to be written after the request completes.
func TestDispatcher_AuditTrue_EmitsEvent(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	var buf bytes.Buffer
	auditLog := audit.New(&buf)

	route := routes.Route{
		ID: "audited", Method: "GET", Path: "/data", Target: upstream.URL,
		Audit: true,
	}
	d := makeDispatcherWithObs(t, upstream, route, auditLog, nil)

	req := ctxRequest("GET", "/data", nil, []string{}, "tenant-x")
	w := httptest.NewRecorder()
	d.ServeHTTP(w, req)

	if buf.Len() == 0 {
		t.Fatal("expected audit event to be written; buffer is empty")
	}
	var evt audit.Event
	if err := json.Unmarshal(buf.Bytes(), &evt); err != nil {
		t.Fatalf("audit event is not valid JSON: %v\nraw: %s", err, buf.String())
	}
	if evt.RouteID != "audited" {
		t.Errorf("route_id: want %q, got %q", "audited", evt.RouteID)
	}
	if evt.StatusCode != http.StatusOK {
		t.Errorf("status_code: want 200, got %d", evt.StatusCode)
	}
	if evt.TenantID != "tenant-x" {
		t.Errorf("tenant_id: want tenant-x, got %q", evt.TenantID)
	}
	if evt.Timestamp == "" {
		t.Error("timestamp should be non-empty")
	}
}

// TestDispatcher_AuditFalse_NoEvent verifies that a route with audit:false
// (the default) does NOT write any audit event.
func TestDispatcher_AuditFalse_NoEvent(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	var buf bytes.Buffer
	auditLog := audit.New(&buf)

	route := routes.Route{
		ID: "silent", Method: "GET", Path: "/data", Target: upstream.URL,
		Audit: false,
	}
	d := makeDispatcherWithObs(t, upstream, route, auditLog, nil)

	req := ctxRequest("GET", "/data", nil, []string{}, "")
	d.ServeHTTP(httptest.NewRecorder(), req)

	if buf.Len() != 0 {
		t.Errorf("expected no audit event for audit:false route; got: %s", buf.String())
	}
}

// TestDispatcher_MetricsRecorded verifies that the metrics registry records one
// observation per request with the correct route ID.
func TestDispatcher_MetricsRecorded(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer upstream.Close()

	reg := metrics.New()
	route := routes.Route{
		ID: "m1", Method: "POST", Path: "/items", Target: upstream.URL,
	}
	d := makeDispatcherWithObs(t, upstream, route, nil, reg)

	req := ctxRequest("POST", "/items", nil, []string{}, "")
	d.ServeHTTP(httptest.NewRecorder(), req)

	// Check Prometheus output contains the route + status.
	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	out := buf.String()

	if !strings.Contains(out, `route="m1"`) {
		t.Errorf("expected route m1 in metrics output; got:\n%s", out)
	}
	if !strings.Contains(out, `status="201"`) {
		t.Errorf("expected status 201 in metrics output; got:\n%s", out)
	}
}

// TestResponseRecorder_CapturesStatus verifies the responseRecorder wrapper.
func TestResponseRecorder_CapturesStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rr := &responseRecorder{ResponseWriter: w, code: http.StatusOK}

	rr.WriteHeader(http.StatusAccepted)
	if rr.code != http.StatusAccepted {
		t.Errorf("code: want 202, got %d", rr.code)
	}
	// Second WriteHeader should not overwrite the first captured code.
	rr.WriteHeader(http.StatusInternalServerError)
	if rr.code != http.StatusAccepted {
		t.Errorf("code should not change after first WriteHeader, got %d", rr.code)
	}
}

// TestResponseRecorder_DefaultsTo200OnWrite verifies that a Write without a
// prior WriteHeader is recorded as 200.
func TestResponseRecorder_DefaultsTo200OnWrite(t *testing.T) {
	w := httptest.NewRecorder()
	rr := &responseRecorder{ResponseWriter: w, code: http.StatusOK}
	_, _ = rr.Write([]byte("hello"))
	if rr.code != http.StatusOK {
		t.Errorf("expected 200 default, got %d", rr.code)
	}
	if !rr.wroteHeader {
		t.Error("wroteHeader should be true after Write")
	}
}
