// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// --- Limiter unit tests ---

// TestLimiter_TryAcquire_WithinLimit verifies that acquisitions within the
// configured limit all succeed and the counter tracks correctly.
func TestLimiter_TryAcquire_WithinLimit(t *testing.T) {
	lim := NewLimiter(3)
	for i := 1; i <= 3; i++ {
		if !lim.TryAcquire("tenant") {
			t.Fatalf("acquire %d should succeed (limit 3)", i)
		}
		if got := lim.Count("tenant"); got != int64(i) {
			t.Errorf("count after %d acquires: want %d, got %d", i, i, got)
		}
	}
}

// TestLimiter_TryAcquire_11th_Rejected is the primary LLM-2 AC: the 11th
// concurrent SSE request from a tenant (limit 10) → acquire fails; counter
// stays at 10 (the failed acquire must not leave a residual increment).
func TestLimiter_TryAcquire_11th_Rejected(t *testing.T) {
	lim := NewLimiter(10)
	for i := 0; i < 10; i++ {
		if !lim.TryAcquire("tenant-001") {
			t.Fatalf("acquire %d should succeed", i+1)
		}
	}
	if got := lim.Count("tenant-001"); got != 10 {
		t.Errorf("count before 11th: want 10, got %d", got)
	}
	// 11th must fail.
	if lim.TryAcquire("tenant-001") {
		t.Error("11th acquire should fail")
	}
	// Counter must remain at 10 — the failed acquire must undo its increment.
	if got := lim.Count("tenant-001"); got != 10 {
		t.Errorf("count after failed 11th: want 10, got %d", got)
	}
}

// TestLimiter_Release_Decrement verifies that Release decrements the counter
// and that a subsequent TryAcquire succeeds after Release.
func TestLimiter_Release_Decrement(t *testing.T) {
	lim := NewLimiter(1)
	if !lim.TryAcquire("t") {
		t.Fatal("first acquire should succeed")
	}
	// At limit: second acquire must fail.
	if lim.TryAcquire("t") {
		t.Error("second acquire should fail when at limit")
	}
	// Release one slot.
	lim.Release("t")
	if got := lim.Count("t"); got != 0 {
		t.Errorf("count after release: want 0, got %d", got)
	}
	// Now a new acquire must succeed.
	if !lim.TryAcquire("t") {
		t.Error("acquire after release should succeed")
	}
	lim.Release("t")
}

// TestLimiter_NoDoubleDecrement_SyncOnce verifies that the sync.Once pattern
// used in the SSE handler (proxy.go) prevents the counter from going below
// zero when Release is called twice for the same acquisition (LLM-2 AC).
func TestLimiter_NoDoubleDecrement_SyncOnce(t *testing.T) {
	lim := NewLimiter(5)
	if !lim.TryAcquire("tenant") {
		t.Fatal("acquire failed")
	}

	// Simulate the double-release scenario that could occur when both
	// server-side stream end and client disconnect fire close together.
	var once sync.Once
	releaseOnce := func() { once.Do(func() { lim.Release("tenant") }) }

	releaseOnce() // first call: decrements
	releaseOnce() // second call: no-op due to sync.Once

	if got := lim.Count("tenant"); got != 0 {
		t.Errorf("count after double release: want 0, got %d (double-decrement occurred)", got)
	}
}

// TestLimiter_IsolatedPerTenant verifies that limits are tracked independently
// per tenant.
func TestLimiter_IsolatedPerTenant(t *testing.T) {
	lim := NewLimiter(1)
	if !lim.TryAcquire("tenant-A") {
		t.Fatal("tenant-A: first acquire should succeed")
	}
	if !lim.TryAcquire("tenant-B") {
		t.Fatal("tenant-B: should succeed independently of tenant-A")
	}
	if lim.TryAcquire("tenant-A") {
		t.Error("tenant-A: second acquire should fail")
	}
	lim.Release("tenant-A")
	lim.Release("tenant-B")
}

// TestLimiter_DefaultLimit verifies that NewLimiter(0) clamps to default 10.
func TestLimiter_DefaultLimit(t *testing.T) {
	lim := NewLimiter(0)
	for i := 0; i < 10; i++ {
		if !lim.TryAcquire("t") {
			t.Fatalf("acquire %d should succeed with default limit 10", i+1)
		}
	}
	if lim.TryAcquire("t") {
		t.Error("11th acquire should fail at default limit 10")
	}
}

// --- HTTP-level tests via NewProxy ---

// sseBlockingUpstream returns an httptest.Server that sends one SSE chunk
// then blocks until release is closed or the request context is cancelled.
// The caller must close release before calling upstream.Close().
func sseBlockingUpstream(release <-chan struct{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "data: active\n\n")
		flusher.Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
}

// readFirstChunk reads from body until it has received some data, confirming
// the SSE stream handler is active (past TryAcquire). Returns when data
// arrives or t.Fatal fires on timeout.
func readFirstChunk(t *testing.T, body io.Reader, label string) {
	t.Helper()
	ch := make(chan struct{}, 1)
	go func() {
		buf := make([]byte, 32)
		_, _ = body.Read(buf)
		ch <- struct{}{}
	}()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s: timeout waiting for initial SSE chunk", label)
	}
}

// TestNewProxy_11thSSE_Returns429 exercises the full LLM-2 acceptance criterion
// at the HTTP handler level: 10 concurrent SSE connections are held open; the
// 11th returns 429 application/vnd.yaagents.error+json with retryAfter: 60
// and the counter remains at 10.
//
// Cleanup uses t.Cleanup in LIFO order so the release channel is closed before
// servers are shut down, avoiding httptest.Server.Close deadlocks.
func TestNewProxy_11thSSE_Returns429(t *testing.T) {
	release := make(chan struct{})
	upstream := sseBlockingUpstream(release)
	// Register server cleanups first (run last in LIFO order).
	t.Cleanup(upstream.Close)

	lim := NewLimiter(10)
	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "", lim, nil)
	gw := httptest.NewServer(h)
	t.Cleanup(gw.Close)

	// Open 10 SSE connections and confirm each is active (past TryAcquire).
	clients := make([]*http.Response, 10)
	for i := 0; i < 10; i++ {
		resp, err := http.Get(gw.URL + "/stream")
		if err != nil {
			t.Fatalf("conn %d: %v", i+1, err)
		}
		clients[i] = resp
		readFirstChunk(t, resp.Body, "conn "+string(rune('0'+i+1)))
	}
	// Register client + release cleanup last (run first in LIFO order).
	t.Cleanup(func() {
		for _, c := range clients {
			c.Body.Close()
		}
		close(release)
	})

	// 11th request must return 429.
	resp11, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("11th request: %v", err)
	}
	defer resp11.Body.Close()
	body11, _ := io.ReadAll(resp11.Body)

	if resp11.StatusCode != http.StatusTooManyRequests {
		t.Errorf("11th status: want 429, got %d", resp11.StatusCode)
	}
	if ct := resp11.Header.Get("Content-Type"); ct != "application/vnd.yaagents.error+json" {
		t.Errorf("content-type: want vendor error, got %q", ct)
	}

	var errBody struct {
		Code       string `json:"code"`
		RetryAfter int    `json:"retryAfter"`
	}
	if err := json.Unmarshal(body11, &errBody); err != nil {
		t.Fatalf("unmarshal 429 body: %v\nraw: %s", err, body11)
	}
	if errBody.Code != "SSE_CONCURRENCY_LIMIT_EXCEEDED" {
		t.Errorf("code: want SSE_CONCURRENCY_LIMIT_EXCEEDED, got %q", errBody.Code)
	}
	if errBody.RetryAfter != 60 {
		t.Errorf("retryAfter: want 60, got %d", errBody.RetryAfter)
	}

	// Counter must remain at 10 (failed acquire must not leave a residual).
	if got := lim.Count(""); got != 10 {
		t.Errorf("counter after 11th: want 10, got %d", got)
	}
}

// TestNewProxy_ClientDisconnect_DecrementsCounter verifies LLM-2 AC: the
// counter decrements when the client disconnects (body closed), freeing the
// slot for the next request from the same tenant.
func TestNewProxy_ClientDisconnect_DecrementsCounter(t *testing.T) {
	release := make(chan struct{})
	upstream := sseBlockingUpstream(release)
	t.Cleanup(upstream.Close)

	lim := NewLimiter(1) // limit=1 for easy single-slot testing
	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "", lim, nil)
	gw := httptest.NewServer(h)
	t.Cleanup(gw.Close)
	t.Cleanup(func() { close(release) })

	// Open the first SSE connection and confirm it is active.
	resp1, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	readFirstChunk(t, resp1.Body, "conn-1")

	if got := lim.Count(""); got != 1 {
		t.Errorf("count with one open connection: want 1, got %d", got)
	}

	// Second request at limit=1 must be rejected.
	resp2, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Errorf("second request: want 429, got %d", resp2.StatusCode)
	}

	// Simulate client disconnect: close the first response body.
	resp1.Body.Close()

	// Poll until the counter drops to 0 (gateway handler returns after
	// detecting the client disconnect via context cancellation).
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if lim.Count("") == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := lim.Count(""); got != 0 {
		t.Errorf("count after disconnect: want 0, got %d", got)
	}

	// A third request from the same tenant must now succeed (slot freed).
	resp3, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("third request: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("third request: want 200, got %d", resp3.StatusCode)
	}
}

// TestNewProxy_NilLimiter_NoLimit verifies that a nil Limiter means no
// concurrency limiting — all requests succeed regardless of count.
func TestNewProxy_NilLimiter_NoLimit(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "", nil, nil)
	gw := httptest.NewServer(h)
	t.Cleanup(gw.Close)

	for i := 0; i < 20; i++ {
		resp, err := http.Get(gw.URL + "/stream")
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Errorf("request %d: unexpected 429 with nil limiter", i+1)
		}
	}
}
