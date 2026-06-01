// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package llm provides LLM-specific proxy components for the yaagents gateway:
// SSE pipe-and-flush proxy (LLM-1), per-tenant SSE concurrency limiter (LLM-2),
// execution-timeout context wrapper (LLM-3), and Prometheus SSE metrics (LLM-4).
//
// Activation is config-driven: a route declares mode: sse in routes.yaml to
// opt into SSE pipe-and-flush; standard JSON routes pay no cost (the SSE code
// path is dormant when no route activates it — ADR PI2-yaa-0002 §1).
package llm

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
)

// ModeSSE is the routes.yaml mode value that activates SSE pipe-and-flush proxying.
// Routes declaring mode: sse use NewProxy; all other routes use the standard
// httputil.ReverseProxy path via the dispatcher.
const ModeSSE = "sse"

// NewProxy returns an http.Handler for an LLM-mode route. It uses SSE
// pipe-and-flush when the upstream responds with Content-Type: text/event-stream,
// falling back to io.Copy for standard (non-SSE) upstream responses.
//
// lim controls per-tenant SSE concurrency (LLM-2). When lim is nil no
// limiting is applied. When the limit is exceeded the handler returns
// 429 application/vnd.yaagents.error+json with retryAfter: 60.
//
// The decrement (Release) runs in a sync.Once-guarded defer so it fires
// exactly once regardless of whether the stream ended server-side or
// via client disconnect — preventing double-decrement (LLM-2 AC).
func NewProxy(upstream *url.URL, lim *Limiter) (http.Handler, error) {
	rawSSE := NewSSEProxy(upstream)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lim != nil {
			tenantID := reqctx.TenantID(r.Context())
			if !lim.TryAcquire(tenantID) {
				response.WriteError(w, http.StatusTooManyRequests, response.ErrorBody{
					Type:       "error",
					Code:       "SSE_CONCURRENCY_LIMIT_EXCEEDED",
					Message:    "too many concurrent SSE connections for this tenant",
					RetryAfter: 60,
					Trace: response.Trace{
						CorrelationID: reqctx.CorrelationID(r.Context()),
						RequestID:     reqctx.RequestID(r.Context()),
					},
				})
				return
			}
			// sync.Once ensures Release is called exactly once even if
			// both server-side stream end and client disconnect race.
			var once sync.Once
			defer func() { once.Do(func() { lim.Release(tenantID) }) }()
		}
		rawSSE(w, r)
	}), nil
}
