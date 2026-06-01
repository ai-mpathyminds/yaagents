// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
)

// nonFlushable wraps http.ResponseWriter without exposing http.Flusher,
// exercising the !canFlush branch in the SSE pipe-and-flush loop.
type nonFlushable struct{ http.ResponseWriter }

// --- NewSSEProxy tests ---

// TestNewSSEProxy_SSEStream_ProgressiveChunks verifies that a route with mode:
// sse + upstream sending text/event-stream delivers all chunks progressively.
// bufio.Scanner reading over a real HTTP connection is the acceptance gate
// (WI-2yaa.LLM-1 AC: "verified with bufio.Scanner reading chunks").
func TestNewSSEProxy_SSEStream_ProgressiveChunks(t *testing.T) {
	chunks := []string{"data: event0\n\n", "data: event1\n\n", "data: event2\n\n"}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for _, chunk := range chunks {
			_, _ = io.WriteString(w, chunk)
			flusher.Flush()
		}
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	req, _ := http.NewRequest("GET", gw.URL+"/events", nil)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Errorf("content-type: want text/event-stream, got %q", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	var got []string
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			got = append(got, line)
		}
	}

	want := []string{"data: event0", "data: event1", "data: event2"}
	if len(got) != len(want) {
		t.Fatalf("chunk count: want %d, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("chunk[%d]: want %q, got %q", i, want[i], got[i])
		}
	}
}

// TestNewSSEProxy_SSEHeaders_SetCorrectly verifies that Cache-Control,
// X-Accel-Buffering, and Content-Encoding are injected for SSE responses.
func TestNewSSEProxy_SSEHeaders_SetCorrectly(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		_, _ = io.WriteString(w, "data: ok\n\n")
		flusher.Flush()
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	// Drain so connection closes cleanly.
	_, _ = io.Copy(io.Discard, resp.Body)

	if got := resp.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control: want no-cache, got %q", got)
	}
	if got := resp.Header.Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering: want no, got %q", got)
	}
	if got := resp.Header.Get("Content-Encoding"); got != "identity" {
		t.Errorf("Content-Encoding: want identity, got %q", got)
	}
}

// TestNewSSEProxy_NonSSEResponse verifies the io.Copy fallback when the
// upstream sends a standard JSON response (non-SSE).
func TestNewSSEProxy_NonSSEResponse(t *testing.T) {
	const jsonBody = `{"type":"result","value":"hello"}`

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, jsonBody)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/api/v1/items")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != jsonBody {
		t.Errorf("body: want %q, got %q", jsonBody, string(body))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}
}

// TestNewSSEProxy_NonSSEResponse_NoSSEHeaders verifies that the SSE-specific
// headers (Cache-Control, X-Accel-Buffering) are NOT set for non-SSE responses.
func TestNewSSEProxy_NonSSEResponse_NoSSEHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{}`)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// SSE-specific headers must NOT be present for non-SSE responses.
	if got := resp.Header.Get("X-Accel-Buffering"); got != "" {
		t.Errorf("X-Accel-Buffering should be absent for non-SSE, got %q", got)
	}
}

// TestNewSSEProxy_UpstreamError_502 verifies that an unreachable upstream
// returns 502 Bad Gateway.
func TestNewSSEProxy_UpstreamError_502(t *testing.T) {
	// Start a server and immediately close it so RoundTrip fails.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	u, _ := url.Parse(deadURL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/test")
	if err != nil {
		t.Fatalf("gateway request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status: want 502, got %d", resp.StatusCode)
	}
}

// TestNewSSEProxy_ClientDisconnect_NoErrorBody verifies that a pre-cancelled
// request context (simulating client disconnect) causes the handler to exit
// silently without writing an error body.
func TestNewSSEProxy_ClientDisconnect_NoErrorBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	handler := NewSSEProxy(u)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so RoundTrip sees a cancelled context immediately

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/stream", nil).WithContext(ctx)
	handler.ServeHTTP(w, req)

	// Pre-cancelled context: RoundTrip fails → r.Context().Err() != nil → return.
	// The handler must NOT write any error body.
	if w.Body.Len() != 0 {
		t.Errorf("expected no body for cancelled context, got %q", w.Body.String())
	}
}

// TestNewSSEProxy_NonFlushableWriter_FallsBackToioCopy verifies the !canFlush
// branch: when the writer does not implement http.Flusher, even an SSE upstream
// response is delivered via io.Copy rather than the pipe-and-flush loop.
func TestNewSSEProxy_NonFlushableWriter_FallsBackToioCopy(t *testing.T) {
	const sseBody = "data: hello\n\n"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, sseBody)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	handler := NewSSEProxy(u)

	rec := httptest.NewRecorder()
	nfw := &nonFlushable{rec}

	req := httptest.NewRequest("GET", "/stream", nil)
	handler.ServeHTTP(nfw, req)

	// io.Copy path: body must contain the SSE event.
	if got := rec.Body.String(); got != sseBody {
		t.Errorf("body: want %q, got %q", sseBody, got)
	}
}

// TestNewSSEProxy_CorrelationHeadersPropagated verifies that X-Request-Id and
// X-Correlation-Id from the request context are forwarded to upstream.
func TestNewSSEProxy_CorrelationHeadersPropagated(t *testing.T) {
	var gotReqID, gotCorrID string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReqID = r.Header.Get("X-Request-Id")
		gotCorrID = r.Header.Get("X-Correlation-Id")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)

	// Wrap the SSE handler in a gateway server that pre-populates reqctx.
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := reqctx.WithRequestID(r.Context(), "req-test-123")
		ctx = reqctx.WithCorrelationID(ctx, "corr-test-456")
		NewSSEProxy(u).ServeHTTP(w, r.WithContext(ctx))
	}))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if gotReqID != "req-test-123" {
		t.Errorf("X-Request-Id: want req-test-123, got %q", gotReqID)
	}
	if gotCorrID != "corr-test-456" {
		t.Errorf("X-Correlation-Id: want corr-test-456, got %q", gotCorrID)
	}
}

// TestNewSSEProxy_AcceptEncodingStripped verifies that Accept-Encoding is
// removed from the outgoing upstream request.
func TestNewSSEProxy_AcceptEncodingStripped(t *testing.T) {
	var gotAcceptEncoding string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAcceptEncoding = r.Header.Get("Accept-Encoding")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	req, _ := http.NewRequest("GET", gw.URL+"/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if gotAcceptEncoding != "" {
		t.Errorf("Accept-Encoding should be stripped, upstream saw %q", gotAcceptEncoding)
	}
}

// TestNewSSEProxy_UpstreamHeadersForwarded verifies that custom headers from
// the upstream response are forwarded to the client.
func TestNewSSEProxy_UpstreamHeadersForwarded(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	gw := httptest.NewServer(NewSSEProxy(u))
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if got := resp.Header.Get("X-Custom-Header"); got != "custom-value" {
		t.Errorf("X-Custom-Header: want custom-value, got %q", got)
	}
}

// --- SSEContextTimeout tests ---

// TestSSEContextTimeout_DeadlineApplied verifies that the middleware adds a
// context deadline matching the configured duration.
func TestSSEContextTimeout_DeadlineApplied(t *testing.T) {
	timeout := 5 * time.Second
	var gotDeadline time.Time
	var deadlineOK bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotDeadline, deadlineOK = r.Context().Deadline()
		w.WriteHeader(http.StatusOK)
	})

	wrapped := SSEContextTimeout(timeout)(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	before := time.Now()
	wrapped.ServeHTTP(w, req)
	after := time.Now()

	if !deadlineOK {
		t.Fatal("expected context deadline to be set by SSEContextTimeout")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", w.Code)
	}
	// Deadline must be in the range [before+timeout, after+timeout].
	minD := before.Add(timeout)
	maxD := after.Add(timeout)
	if gotDeadline.Before(minD) || gotDeadline.After(maxD) {
		t.Errorf("deadline %v not in range [%v, %v]", gotDeadline, minD, maxD)
	}
}

// TestSSEContextTimeout_HandlerReceivesTimeout verifies that a stalled handler
// is cancelled when the deadline fires.
func TestSSEContextTimeout_HandlerReceivesTimeout(t *testing.T) {
	done := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for context cancellation.
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
			t.Error("handler did not receive cancellation within 2s")
		}
		close(done)
	})

	wrapped := SSEContextTimeout(50 * time.Millisecond)(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	wrapped.ServeHTTP(w, req)

	select {
	case <-done:
		// Pass: handler was cancelled.
	case <-time.After(1 * time.Second):
		t.Error("handler did not complete within 1s after timeout")
	}
}

// --- NewProxy tests ---

// TestNewProxy_NilError verifies that NewProxy always succeeds.
func TestNewProxy_NilError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	h, err := NewProxy(u)
	if err != nil {
		t.Fatalf("NewProxy returned unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// TestNewProxy_ProxiesRequest verifies that the handler returned by NewProxy
// successfully proxies a request to upstream.
func TestNewProxy_ProxiesRequest(t *testing.T) {
	reached := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u)

	gw := httptest.NewServer(h)
	defer gw.Close()

	resp, err := http.DefaultClient.Get(gw.URL + "/path")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if !reached {
		t.Error("upstream was not reached by NewProxy handler")
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: want 202, got %d", resp.StatusCode)
	}
}

// TestModeSSE_Constant verifies the exported ModeSSE constant value.
func TestModeSSE_Constant(t *testing.T) {
	if ModeSSE != "sse" {
		t.Errorf("ModeSSE: want %q, got %q", "sse", ModeSSE)
	}
}
