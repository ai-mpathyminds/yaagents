// Package proxy implements the yaagents gateway route dispatcher:
// per-route RBAC enforcement, typed-response passthrough via
// httputil.ReverseProxy, X-YAAgents-Profile header injection, and optional
// per-route audit logging + Prometheus metrics observation (WI-1yaa.GW-5).
//
// ADR: PI1-yaa-0001 (net/http only; no framework imports; no cross-product deps).
package proxy

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/audit"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/metrics"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/tenant"
)

// ProfileHeader is added to every proxied response (Agentic REST Profile §3).
const ProfileHeader = "X-YAAgents-Profile"

// ProfileVersion is the current Agentic REST Profile version advertised on responses.
const ProfileVersion = "v0.1"

// routeEntry pairs a validated Route with its pre-built HTTP handler.
// The handler chain is: observe → EnforceTenant → RBAC → reverse-proxy.
type routeEntry struct {
	route   routes.Route
	handler http.Handler
}

// RouteDispatcher matches inbound requests to the route table and forwards
// them to the upstream target with RBAC enforcement and typed-response passthrough.
// auditLog and reg are optional (nil disables the feature).
type RouteDispatcher struct {
	entries  []routeEntry
	log      *slog.Logger
	auditLog *audit.Logger
	reg      *metrics.Registry
}

// New builds a RouteDispatcher from a validated route list.
// auditLog and reg may be nil to disable audit/metrics respectively.
// Returns an error if any route's target URL cannot be parsed.
func New(routeList []routes.Route, log *slog.Logger, auditLog *audit.Logger, reg *metrics.Registry) (*RouteDispatcher, error) {
	d := &RouteDispatcher{
		entries:  make([]routeEntry, 0, len(routeList)),
		log:      log,
		auditLog: auditLog,
		reg:      reg,
	}
	for _, r := range routeList {
		handler, err := makeRouteHandler(r, log)
		if err != nil {
			return nil, fmt.Errorf("building handler for route %q: %w", r.ID, err)
		}
		// Wrap with per-request observation (metrics + audit).
		handler = d.observeHandler(handler, r)
		d.entries = append(d.entries, routeEntry{route: r, handler: handler})
	}
	return d, nil
}

// ServeHTTP implements http.Handler. It matches the request, applies the per-route
// handler chain, and writes a 404 vendor-error when no route matches.
func (d *RouteDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	entry, ok := d.matchEntry(r)
	if !ok {
		response.WriteError(w, http.StatusNotFound, response.ErrorBody{
			Type:    "error",
			Code:    "ROUTE_NOT_FOUND",
			Message: fmt.Sprintf("no route matched %s %s", r.Method, r.URL.Path),
			Trace:   traceFromReq(r),
		})
		return
	}
	entry.handler.ServeHTTP(w, r)
}

// matchEntry returns the first routeEntry whose Method and Path match the request.
func (d *RouteDispatcher) matchEntry(r *http.Request) (routeEntry, bool) {
	for _, e := range d.entries {
		if strings.EqualFold(e.route.Method, r.Method) && matchPath(e.route.Path, r.URL.Path) {
			return e, true
		}
	}
	return routeEntry{}, false
}

// matchPath returns true when the route pattern (which may contain {param}
// placeholders) matches requestPath.
// Rules:
//   - segment count must match
//   - literal segments must match exactly (case-sensitive)
//   - {param} segments match any non-empty path segment
func matchPath(pattern, requestPath string) bool {
	pSegs := splitPath(pattern)
	rSegs := splitPath(requestPath)
	if len(pSegs) != len(rSegs) {
		return false
	}
	for i, ps := range pSegs {
		if strings.HasPrefix(ps, "{") && strings.HasSuffix(ps, "}") {
			if rSegs[i] == "" {
				return false
			}
			continue
		}
		if ps != rSegs[i] {
			return false
		}
	}
	return true
}

// splitPath splits a URL path by "/" and filters empty segments so that
// "/foo/bar/" and "/foo/bar" produce the same result.
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// responseRecorder wraps http.ResponseWriter to capture the first status code
// written by downstream handlers (including httputil.ReverseProxy).
type responseRecorder struct {
	http.ResponseWriter
	code        int
	wroteHeader bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.wroteHeader {
		rr.code = code
		rr.wroteHeader = true
	}
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.wroteHeader {
		rr.code = http.StatusOK
		rr.wroteHeader = true
	}
	return rr.ResponseWriter.Write(b)
}

// observeHandler wraps h to measure latency, record metrics, and emit an audit
// event (only when route.Audit is true). Metrics and audit logging are no-ops
// when the respective field on RouteDispatcher is nil.
func (d *RouteDispatcher) observeHandler(h http.Handler, route routes.Route) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &responseRecorder{ResponseWriter: w, code: http.StatusOK}
		start := time.Now()

		h.ServeHTTP(rr, r)

		latencyMS := float64(time.Since(start).Milliseconds())

		if d.reg != nil {
			d.reg.Record(route.ID, rr.code, latencyMS)
		}
		if route.Audit && d.auditLog != nil {
			ctx := r.Context()
			d.auditLog.Log(audit.Event{
				Timestamp:     audit.Timestamp(),
				RouteID:       route.ID,
				Method:        r.Method,
				Path:          r.URL.Path,
				TenantID:      reqctx.TenantID(ctx),
				ActorSubject:  reqctx.ActorSubject(ctx),
				StatusCode:    rr.code,
				LatencyMS:     latencyMS,
				CorrelationID: reqctx.CorrelationID(ctx),
				RequestID:     reqctx.RequestID(ctx),
			})
		}
	})
}

// makeRouteHandler constructs the per-route handler chain:
//
//	EnforceTenant(route.TenantRequired) → RBAC check → reverse-proxy
func makeRouteHandler(route routes.Route, log *slog.Logger) (http.Handler, error) {
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		return nil, fmt.Errorf("parsing target %q: %w", route.Target, err)
	}

	rp := buildProxy(targetURL, route, log)

	// Inner handler: RBAC then proxy.
	rbacAndProxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := enforceRoles(r, route.Roles); err != nil {
			response.WriteError(w, http.StatusForbidden, response.ErrorBody{
				Type:    "forbidden",
				Code:    "INSUFFICIENT_ROLES",
				Message: err.Error(),
				Trace:   traceFromReq(r),
			})
			return
		}
		rp.ServeHTTP(w, r)
	})

	// Wrap with per-route tenant enforcement.
	return tenant.EnforceTenant(route.TenantRequired)(rbacAndProxy), nil
}

// buildProxy constructs an httputil.ReverseProxy for the given target URL.
//
//   - Director: rewrites the upstream URL (scheme + host) and injects
//     tenant/actor/correlation headers. Does NOT re-encode the body or
//     alter the method — typed-response passthrough is byte-level.
//   - ModifyResponse: adds X-YAAgents-Profile to every upstream response
//     without touching status, Content-Type, or body.
//   - ErrorHandler: emits a 502 vendor-error when the upstream is unreachable.
func buildProxy(targetURL *url.URL, route routes.Route, log *slog.Logger) *httputil.ReverseProxy {
	p := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host

			tenant.InjectUpstreamHeaders(req)

			log.Debug("proxying request",
				slog.String("route_id", route.ID),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("target", targetURL.String()),
			)
		},
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Set(ProfileHeader, ProfileVersion)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("upstream proxy error",
				slog.String("route_id", route.ID),
				slog.String("target", targetURL.String()),
				slog.String("error", err.Error()),
			)
			response.WriteError(w, http.StatusBadGateway, response.ErrorBody{
				Type:    "failed_dependency",
				Code:    "UPSTREAM_UNAVAILABLE",
				Message: "upstream service did not respond",
				Trace:   traceFromReq(r),
			})
		},
	}
	return p
}

// enforceRoles verifies that every required role is present in the actor's
// role claims. Returns a descriptive error listing the first missing role.
func enforceRoles(r *http.Request, required []string) error {
	if len(required) == 0 {
		return nil
	}
	actorRoles := reqctx.ActorRoles(r.Context())
	roleSet := make(map[string]bool, len(actorRoles))
	for _, role := range actorRoles {
		roleSet[role] = true
	}
	for _, req := range required {
		if !roleSet[req] {
			return errors.New("actor is missing required role: " + req)
		}
	}
	return nil
}

// traceFromReq builds a response.Trace from the request context.
func traceFromReq(r *http.Request) response.Trace {
	ctx := r.Context()
	return response.Trace{
		CorrelationID: reqctx.CorrelationID(ctx),
		RequestID:     reqctx.RequestID(ctx),
	}
}
