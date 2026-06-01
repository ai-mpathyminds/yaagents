// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
)

// sseTransport is a clone of http.DefaultTransport with DisableCompression=true.
// Prevents the upstream from gzip-encoding SSE bytes so the flush loop can
// forward raw event-stream data to the client without re-decompressing.
var sseTransport = func() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableCompression = true
	return t
}()

// NewSSEProxy returns an http.HandlerFunc that proxies requests to upstream.
//
// When the upstream responds with Content-Type: text/event-stream the handler
// uses an explicit Read→Write→Flush loop so each SSE chunk is forwarded to the
// client immediately (pipe-and-flush semantics). For non-SSE upstream responses
// the handler falls through to a standard io.Copy.
//
// Request-id and correlation-id are propagated from the request context via
// reqctx (yaagents-canonical; no ai-platform imports).
// Accept-Encoding is stripped before forwarding to prevent gzip-encoded SSE.
func NewSSEProxy(upstream *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build the outbound URL: preserve path + query from the inbound request.
		target := *upstream
		target.Path = r.URL.Path
		target.RawQuery = r.URL.RawQuery

		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
		if err != nil {
			http.Error(w, "internal_error", http.StatusInternalServerError)
			return
		}

		// Clone inbound headers (includes X-Tenant-ID set by the tenant-injector
		// plugin that runs before this handler in the plugin chain).
		outReq.Header = r.Header.Clone()

		// Propagate correlation IDs from context (populated by ContextMiddleware
		// and the plugin chain before this handler runs).
		if rid := reqctx.RequestID(r.Context()); rid != "" {
			outReq.Header.Set("X-Request-Id", rid)
		}
		if cid := reqctx.CorrelationID(r.Context()); cid != "" {
			outReq.Header.Set("X-Correlation-Id", cid)
		}

		// Strip Accept-Encoding so the upstream never gzip-encodes SSE bytes
		// (gzip-encoded SSE cannot be flushed chunk-by-chunk to the client).
		outReq.Header.Del("Accept-Encoding")

		resp, err := sseTransport.RoundTrip(outReq)
		if err != nil {
			switch r.Context().Err() {
			case context.DeadlineExceeded:
				// Execution timeout fired before upstream responded (LLM-3).
				response.WriteError(w, http.StatusInternalServerError, response.ErrorBody{
					Type:    "error",
					Code:    "EXECUTION_TIMEOUT",
					Message: "execution timeout exceeded",
					Trace: response.Trace{
						CorrelationID: reqctx.CorrelationID(r.Context()),
						RequestID:     reqctx.RequestID(r.Context()),
					},
				})
			case context.Canceled:
				// Client disconnected; no response to write.
			default:
				http.Error(w, "upstream_error", http.StatusBadGateway)
			}
			return
		}
		defer resp.Body.Close()

		// Forward all upstream response headers to the client.
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		isSSE := strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream")
		if isSSE {
			// Disable any downstream proxy buffering or compression so the client
			// receives each event chunk without batching.
			w.Header().Set("Content-Encoding", "identity")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("X-Accel-Buffering", "no")
		}

		w.WriteHeader(resp.StatusCode)

		flusher, canFlush := w.(http.Flusher)
		if !isSSE || !canFlush {
			// Non-SSE upstream response or non-flushable writer: standard copy.
			_, _ = io.Copy(w, resp.Body)
			return
		}

		// SSE pipe-and-flush loop: flush after every read so the client receives
		// each event as soon as it leaves the upstream.
		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				_, _ = w.Write(buf[:n])
				flusher.Flush()
			}
			if readErr != nil {
				return
			}
		}
	}
}

// SSEContextTimeout returns middleware that wraps the request context with a
// deadline. When the deadline fires the upstream HTTP request (which inherits
// the context) is cancelled, terminating the SSE read loop and closing the
// client connection cleanly.
//
// For SSE routes the caller should pass d = executionTimeoutSeconds + 30s
// (PRD §7.1 SSEReadTimeout = ExecutionTimeoutSeconds + 30 s). For non-SSE
// routes use d = executionTimeoutSeconds exactly (LLM-3).
func SSEContextTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
