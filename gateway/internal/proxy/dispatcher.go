// Package proxy implements the yaagents gateway route dispatcher:
// per-route RBAC enforcement, typed-response passthrough via
// httputil.ReverseProxy, and X-YAAgents-Profile header injection.
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
// The handler chain is: EnforceTenant → RBAC → reverse-proxy.
type routeEntry struct {
	route   routes.Route
	handler http.Handler
}

// RouteDispatcher matches inbound requests to the route table and forwards
// them to the upstream target with RBAC enforcement and typed-response passthrough.
type RouteDispatcher struct {
	entries []routeEntry
	log     *slog.Logger
}

// New builds a RouteDispatcher from a validated route list.
// Returns an error if any route's target URL cannot be parsed (belt-and-suspenders;
// routes.Load already validates URLs at boot).
func New(routeList []routes.Route, log *slog.Logger) (*RouteDispatcher, error) {
	entries := make([]routeEntry, 0, len(routeList))
	for _, r := range routeList {
		handler, err := makeRouteHandler(r, log)
		if err != nil {
			return nil, fmt.Errorf("building handler for route %q: %w", r.ID, err)
		}
		entries = append(entries, routeEntry{route: r, handler: handler})
	}
	return &RouteDispatcher{entries: entries, log: log}, nil
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
			// Placeholder — match any non-empty segment.
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

// makeRouteHandler constructs the per-route handler chain:
//
//	EnforceTenant(route.TenantRequired) → RBAC check → reverse-proxy
func makeRouteHandler(route routes.Route, log *slog.Logger) (http.Handler, error) {
	targetURL, err := url.Parse(route.Target)
	if err != nil {
		return nil, fmt.Errorf("parsing target %q: %w", route.Target, err)
	}

	proxy := buildProxy(targetURL, route, log)

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
		proxy.ServeHTTP(w, r)
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
			// Preserve the original request path (including query string and
			// {param} segments) — do not rewrite to targetURL.Path.
			req.Host = targetURL.Host

			// Inject tenant/actor context headers for the upstream service.
			tenant.InjectUpstreamHeaders(req)

			log.Debug("proxying request",
				slog.String("route_id", route.ID),
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("target", targetURL.String()),
			)
		},
		ModifyResponse: func(resp *http.Response) error {
			// Add profile header without touching status, Content-Type, or body
			// (typed-response passthrough per AC).
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
		return nil // Route is open to all authenticated callers.
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
