package proxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
func makeDispatcher(t *testing.T, upstream *httptest.Server, routeDef routes.Route) *RouteDispatcher {
	t.Helper()
	if upstream != nil {
		routeDef.Target = upstream.URL
	}
	d, err := New([]routes.Route{routeDef}, nullLog())
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

	// Capture what Director sets by calling it directly.
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
