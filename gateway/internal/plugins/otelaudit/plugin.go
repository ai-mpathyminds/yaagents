// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package otelaudit implements the otel-audit stub plugin (PRD §6.5 plugin e).
//
// # Stub status
//
// Full OTel span export is deferred to PI3-yaa or community
// (ADR PI2-yaa-0005 Decision 2). This implementation:
//   - Validates that endpoint, if configured, is a parseable absolute URL.
//     A malformed endpoint is an operator config bug → Init returns non-nil
//     error (gateway exit 1).
//   - Emits a single structured warn log on the first request when
//     enabled: true AND endpoint is empty (exporter not configured).
//   - Records a no-op span placeholder via [noopTracer] (hand-rolled;
//     go.opentelemetry.io/otel/trace/noop not yet in go.mod per
//     ADR PI2-yaa-0005 Decision 1 — "else hand-rolled noop tracer").
//   - Passes every request through to next.
//
// Registration: init() calls plugin.Register so the gateway wires this plugin
// by import side-effect (ADR PI2-yaa-0001 §3; no plugin.Open / dlopen).
package otelaudit

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&OtelAudit{})
}

// noopTracer is a zero-allocation placeholder for the OTel tracer.
// Replaced by go.opentelemetry.io/otel/trace/noop in a future PI when the
// exporter is wired.
//
// startSpan returns ctx unchanged and an end function that is a no-op.
type noopTracer struct{}

func (noopTracer) startSpan(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}

// tracer is the package-level noop tracer instance.
var tracer = noopTracer{}

// OtelAudit is the otel-audit stub plugin.
// Zero value is invalid; always call Init before Handler.
type OtelAudit struct {
	enabled  bool
	endpoint string
	warnOnce sync.Once
}

// Name returns the canonical plugin identifier.
func (a *OtelAudit) Name() string { return "otel-audit" }

// Init validates the endpoint URL (when provided) and stores configuration.
//
// Returns a non-nil error (gateway exit 1) when endpoint is non-empty but is
// not a valid absolute URL (scheme + host required). An empty endpoint is
// accepted — the plugin emits a warn-once at request time instead.
func (a *OtelAudit) Init(cfg plugin.PluginConfig) error {
	endpoint := cfg.GetString("endpoint")
	if endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("otel-audit: endpoint %q must be a valid absolute URL (e.g. http://otel-collector:4317)", endpoint)
		}
	}
	a.enabled = cfg.GetBool("enabled")
	a.endpoint = endpoint
	return nil
}

// Handler returns a middleware that records a no-op span and passes the
// request through to next.
//
// When enabled and endpoint is empty, the first request emits a single
// structured warn log noting the exporter is not configured (sync.Once guard).
func (a *OtelAudit) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.enabled && a.endpoint == "" {
			a.warnOnce.Do(func() {
				slog.Warn("otel-audit exporter not configured; spans will be dropped (set endpoint in plugin config)")
			})
		}

		// Emit a no-op span placeholder; full export wired in PI3-yaa or community.
		ctx, endSpan := tracer.startSpan(r.Context(), "yaagents.gateway.request")
		defer endSpan()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Shutdown is a no-op; the plugin holds no background goroutines.
func (a *OtelAudit) Shutdown(_ context.Context) error { return nil }
