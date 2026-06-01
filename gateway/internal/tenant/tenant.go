// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tenant provides middleware for tenant/actor context propagation
// and per-route tenant enforcement per the Agentic REST Profile (ADR PI1-yaa-0001).
//
// Middleware chain order (after auth.Middleware):
//
//	ContextMiddleware → EnforceTenant(route.TenantRequired) → handler
//
// Upstream header injection is called from the reverse proxy in WI-1yaa.GW-4.
package tenant

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/auth"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
)

// ContextMiddleware is a global middleware that runs after auth.Middleware.
// It:
//   - generates X-Correlation-ID (UUID v4) if the header is absent; passes it
//     through unchanged if present
//   - always generates a fresh X-Request-ID for this request
//   - extracts X-Tenant-ID from the inbound header
//   - extracts actor subject + roles from auth.Claims stored in context by GW-2
//   - stores all four values in the request context via reqctx
//   - echoes X-Correlation-ID and X-Request-ID on the response
func ContextMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			corrID := r.Header.Get("X-Correlation-ID")
			if corrID == "" {
				corrID = reqctx.NewUUID()
			}
			reqID := reqctx.NewUUID()
			tenantID := r.Header.Get("X-Tenant-ID")

			actorSubject := ""
			var actorRoles []string
			if claims, ok := r.Context().Value(auth.ClaimsKey).(auth.Claims); ok {
				actorSubject = claims.Subject
				actorRoles = claims.Roles
			}

			ctx := reqctx.WithCorrelationID(r.Context(), corrID)
			ctx = reqctx.WithRequestID(ctx, reqID)
			ctx = reqctx.WithTenantID(ctx, tenantID)
			ctx = reqctx.WithActorSubject(ctx, actorSubject)
			ctx = reqctx.WithActorRoles(ctx, actorRoles)

			// Echo IDs on every response so callers can correlate downstream.
			w.Header().Set("X-Correlation-ID", corrID)
			w.Header().Set("X-Request-ID", reqID)

			log.Debug("request context populated",
				slog.String("correlation_id", corrID),
				slog.String("request_id", reqID),
				slog.String("tenant_id", tenantID),
				slog.String("actor_subject", actorSubject),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// EnforceTenant returns a middleware that rejects requests on routes where
// tenantRequired is true but X-Tenant-ID was not supplied by the caller.
// Rejection uses 403 application/vnd.yaagents.error+json (profile §7.7).
// Called per matched route in WI-1yaa.GW-4 after ContextMiddleware has run.
func EnforceTenant(tenantRequired bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tenantRequired && reqctx.TenantID(r.Context()) == "" {
				response.WriteError(w, http.StatusForbidden, response.ErrorBody{
					Type:    "forbidden",
					Code:    "TENANT_REQUIRED",
					Message: "route requires X-Tenant-ID header",
					Trace: response.Trace{
						CorrelationID: reqctx.CorrelationID(r.Context()),
						RequestID:     reqctx.RequestID(r.Context()),
					},
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// InjectUpstreamHeaders sets tenant, actor, and correlation headers on an
// outgoing upstream request before it is forwarded by the reverse proxy (GW-4).
// Headers are only added when the corresponding context value is non-empty.
func InjectUpstreamHeaders(r *http.Request) {
	ctx := r.Context()
	if tid := reqctx.TenantID(ctx); tid != "" {
		r.Header.Set("X-Tenant-ID", tid)
	}
	if sub := reqctx.ActorSubject(ctx); sub != "" {
		r.Header.Set("X-Actor-Subject", sub)
	}
	if roles := reqctx.ActorRoles(ctx); len(roles) > 0 {
		r.Header.Set("X-Actor-Roles", strings.Join(roles, ","))
	}
	r.Header.Set("X-Correlation-ID", reqctx.CorrelationID(ctx))
	r.Header.Set("X-Request-ID", reqctx.RequestID(ctx))
}
