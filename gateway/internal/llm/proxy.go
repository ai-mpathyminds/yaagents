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
)

// ModeSSE is the routes.yaml mode value that activates SSE pipe-and-flush proxying.
// Routes declaring mode: sse use NewProxy; all other routes use the standard
// httputil.ReverseProxy path via the dispatcher.
const ModeSSE = "sse"

// NewProxy returns an http.Handler for an LLM-mode route. It uses SSE
// pipe-and-flush when the upstream responds with Content-Type: text/event-stream,
// falling back to io.Copy for standard (non-SSE) upstream responses.
//
// The error return is reserved for LLM-2/LLM-3 initialisation that may fail;
// the current LLM-1 implementation always succeeds.
func NewProxy(upstream *url.URL) (http.Handler, error) {
	return NewSSEProxy(upstream), nil
}
