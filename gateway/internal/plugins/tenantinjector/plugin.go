// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tenantinjector implements the tenant-injector plugin (PRD §6.5 plugin b).
//
// The plugin reads the tenant ID from a configured request header, validates it
// against an optional allowlist, and injects it into the upstream request via a
// second configured header. Requests whose tenant ID is absent from a non-empty
// allowlist receive a 403 application/vnd.yaagents.error+json response; next is
// not called.
//
// Configuration keys:
//
//	header:        request header from which the tenant ID is read (typical: X-Tenant-ID)
//	inject_header: header written into the upstream request (typical: X-Actor-Tenant)
//	allowlist:     []string of permitted tenant IDs; empty means all tenants accepted
//
// Both header and inject_header must be non-empty; Init returns an error if
// either is absent or blank so the gateway exits on misconfiguration.
//
// Registration: init() calls plugin.Register so the gateway wires this plugin by
// import side-effect (ADR PI2-yaa-0001 §3; no plugin.Open / dlopen).
package tenantinjector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&TenantInjector{})
}

// TenantInjector is the tenant-injector plugin.
// Zero value is invalid; always initialise via Init.
type TenantInjector struct {
	header       string
	injectHeader string
	allowlist    map[string]struct{} // nil or empty → all tenants accepted
}

// Name returns the canonical plugin identifier used for registry lookup and
// YAML-config block matching.
func (ti *TenantInjector) Name() string { return "tenant-injector" }

// Init validates the plugin configuration and stores operational values.
//
// Returns a non-nil error (gateway exit 1) when:
//   - header is empty or absent (typical value: X-Tenant-ID)
//   - inject_header is empty or absent (typical value: X-Actor-Tenant)
func (ti *TenantInjector) Init(cfg plugin.PluginConfig) error {
	hdr := cfg.GetString("header")
	if hdr == "" {
		return fmt.Errorf("tenant-injector: header must be non-empty " +
			"(typical value: X-Tenant-ID)")
	}

	injectHdr := cfg.GetString("inject_header")
	if injectHdr == "" {
		return fmt.Errorf("tenant-injector: inject_header must be non-empty " +
			"(typical value: X-Actor-Tenant)")
	}

	ti.header = hdr
	ti.injectHeader = injectHdr

	allowSlice := cfg.GetStringSlice("allowlist")
	ti.allowlist = make(map[string]struct{}, len(allowSlice))
	for _, id := range allowSlice {
		ti.allowlist[id] = struct{}{}
	}

	return nil
}

// Handler returns an http.Handler that enforces tenant allowlist policy and
// injects the tenant ID into the upstream request header.
//
// Execution order:
//  1. Read tenant ID from r.Header.Get(ti.header).
//  2. If allowlist is non-empty and the tenant ID is not present → write 403
//     application/vnd.yaagents.error+json and return without calling next.
//  3. Set r.Header[ti.injectHeader] = tenantID on the upstream request.
//  4. Call next.ServeHTTP(w, r).
func (ti *TenantInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get(ti.header)

		if len(ti.allowlist) > 0 {
			if _, ok := ti.allowlist[tenantID]; !ok {
				corrID := reqctx.CorrelationID(r.Context())
				reqID := reqctx.RequestID(r.Context())
				slog.Warn("tenant-injector: tenant not in allowlist",
					slog.String("tenant_id", tenantID),
					slog.String("correlation_id", corrID))
				response.WriteError(w, http.StatusForbidden, response.ErrorBody{
					Type:    "forbidden",
					Code:    "tenant_not_allowed",
					Message: "tenant ID is not in the allowlist",
					Trace: response.Trace{
						CorrelationID: corrID,
						RequestID:     reqID,
					},
				})
				return
			}
		}

		// Inject tenant ID into the upstream request before calling next.
		r.Header.Set(ti.injectHeader, tenantID)
		next.ServeHTTP(w, r)
	})
}

// Shutdown is a no-op; the plugin holds no background resources.
func (ti *TenantInjector) Shutdown(_ context.Context) error { return nil }
