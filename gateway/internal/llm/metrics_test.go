// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
)

// --- SSEMetrics unit tests ---

// TestSSEMetrics_Inc_ActiveCount verifies Inc increments the gauge and
// ActiveCount reflects it.
func TestSSEMetrics_Inc_ActiveCount(t *testing.T) {
	m := NewSSEMetrics()
	m.Inc("tenant-001", "completions")
	m.Inc("tenant-001", "completions")
	m.Inc("tenant-002", "completions")

	if got := m.ActiveCount("tenant-001", "completions"); got != 2 {
		t.Errorf("after 2 Inc: want 2, got %d", got)
	}
	if got := m.ActiveCount("tenant-002", "completions"); got != 1 {
		t.Errorf("tenant-002 after 1 Inc: want 1, got %d", got)
	}
}

// TestSSEMetrics_Dec verifies Dec decrements the gauge.
func TestSSEMetrics_Dec(t *testing.T) {
	m := NewSSEMetrics()
	m.Inc("t", "r")
	m.Inc("t", "r")
	m.Dec("t", "r")
	if got := m.ActiveCount("t", "r"); got != 1 {
		t.Errorf("after 1 Dec: want 1, got %d", got)
	}
	m.Dec("t", "r")
	if got := m.ActiveCount("t", "r"); got != 0 {
		t.Errorf("after 2 Decs: want 0, got %d", got)
	}
}

// TestSSEMetrics_Dec_BelowZero_Prunes verifies Dec never drives the counter
// below 0 — zero entries are pruned from the internal map.
func TestSSEMetrics_Dec_BelowZero_Prunes(t *testing.T) {
	m := NewSSEMetrics()
	m.Inc("t", "r")
	m.Dec("t", "r") // to 0 → pruned
	m.Dec("t", "r") // over-Dec; must not go negative
	if got := m.ActiveCount("t", "r"); got != 0 {
		t.Errorf("after over-Dec: want 0, got %d", got)
	}
}

// TestSSEMetrics_Error_ErrorCount verifies Error increments and ErrorCount reads.
func TestSSEMetrics_Error_ErrorCount(t *testing.T) {
	m := NewSSEMetrics()
	m.Error("tenant-001", "completions", "limit_exceeded")
	m.Error("tenant-001", "completions", "limit_exceeded")
	m.Error("tenant-001", "completions", "timeout")
	m.Error("tenant-002", "completions", "client_disconnect")

	if got := m.ErrorCount("tenant-001", "completions", "limit_exceeded"); got != 2 {
		t.Errorf("limit_exceeded: want 2, got %d", got)
	}
	if got := m.ErrorCount("tenant-001", "completions", "timeout"); got != 1 {
		t.Errorf("timeout: want 1, got %d", got)
	}
	if got := m.ErrorCount("tenant-002", "completions", "client_disconnect"); got != 1 {
		t.Errorf("client_disconnect: want 1, got %d", got)
	}
	if got := m.ErrorCount("tenant-001", "completions", "upstream_error"); got != 0 {
		t.Errorf("upstream_error (never recorded): want 0, got %d", got)
	}
}

// TestSSEMetrics_WritePrometheus_MetricNames verifies both HELP+TYPE lines
// are always emitted (LLM-4 AC: "/metrics includes both new SSE metrics").
func TestSSEMetrics_WritePrometheus_MetricNames(t *testing.T) {
	m := NewSSEMetrics()
	var sb strings.Builder
	m.WritePrometheus(&sb)
	out := sb.String()

	for _, want := range []string{
		"# HELP yaagents_gateway_sse_connections_active",
		"# TYPE yaagents_gateway_sse_connections_active gauge",
		"# HELP yaagents_gateway_sse_errors_total",
		"# TYPE yaagents_gateway_sse_errors_total counter",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in WritePrometheus output:\n%s", want, out)
		}
	}
}

// TestSSEMetrics_WritePrometheus_ActiveLine verifies a labelled gauge line
// appears for a known active-connections entry.
func TestSSEMetrics_WritePrometheus_ActiveLine(t *testing.T) {
	m := NewSSEMetrics()
	m.Inc("tenant-001", "completions")
	m.Inc("tenant-001", "completions")

	var sb strings.Builder
	m.WritePrometheus(&sb)
	out := sb.String()

	want := `yaagents_gateway_sse_connections_active{tenant_id="tenant-001",route_id="completions"} 2`
	if !strings.Contains(out, want) {
		t.Errorf("expected gauge line %q in output:\n%s", want, out)
	}
}

// TestSSEMetrics_WritePrometheus_ErrorLine verifies a labelled counter line
// appears for a known error entry.
func TestSSEMetrics_WritePrometheus_ErrorLine(t *testing.T) {
	m := NewSSEMetrics()
	m.Error("tenant-001", "completions", "limit_exceeded")

	var sb strings.Builder
	m.WritePrometheus(&sb)
	out := sb.String()

	want := `yaagents_gateway_sse_errors_total{tenant_id="tenant-001",route_id="completions",error_kind="limit_exceeded"} 1`
	if !strings.Contains(out, want) {
		t.Errorf("expected counter line %q in output:\n%s", want, out)
	}
}

// TestSSEMetrics_WritePrometheus_SortedOutput verifies deterministic sort order.
func TestSSEMetrics_WritePrometheus_SortedOutput(t *testing.T) {
	m := NewSSEMetrics()
	m.Inc("z-tenant", "route-b")
	m.Inc("a-tenant", "route-a")

	var sb strings.Builder
	m.WritePrometheus(&sb)
	lines := strings.Split(sb.String(), "\n")

	firstDataIdx := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "yaagents_gateway_sse_connections_active{") {
			firstDataIdx = i
			break
		}
	}
	if firstDataIdx < 0 {
		t.Fatal("no active gauge data lines found")
	}
	if !strings.Contains(lines[firstDataIdx], "a-tenant") {
		t.Errorf("first sorted line should be a-tenant, got: %s", lines[firstDataIdx])
	}
}

// --- Integration tests via NewProxy (HTTP-level) ---

// tenantGateway wraps handler to inject a tenant ID into the request context,
// simulating the tenant-injector plugin running upstream.
func tenantGateway(tenantID string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := reqctx.WithTenantID(r.Context(), tenantID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TestNewProxy_ActiveGauge_RisesAndFalls is the primary LLM-4 integration AC:
// the active gauge rises when an SSE stream is open and falls when it closes.
func TestNewProxy_ActiveGauge_RisesAndFalls(t *testing.T) {
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	// Register cleanups in LIFO order: close clients+release first, then servers.
	t.Cleanup(upstream.Close)

	met := NewSSEMetrics()
	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "completions", nil, met)

	gw := httptest.NewServer(tenantGateway("tenant-001", h))
	t.Cleanup(gw.Close)
	t.Cleanup(func() { close(release) })

	// Open SSE connection; read first chunk to confirm handler passed TryAcquire.
	resp, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		resp.Body.Close()
		t.Fatal("expected initial SSE chunk")
	}

	// Gauge must be 1 while stream is open.
	if got := met.ActiveCount("tenant-001", "completions"); got != 1 {
		t.Errorf("active while open: want 1, got %d", got)
	}

	// Close client → handler context cancelled → gauge decrements.
	resp.Body.Close()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if met.ActiveCount("tenant-001", "completions") == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := met.ActiveCount("tenant-001", "completions"); got != 0 {
		t.Errorf("active after close: want 0, got %d", got)
	}
}

// TestNewProxy_LimitExceeded_RecordsMetric is the LLM-4 AC:
// `yaagents_gateway_sse_errors_total{error_kind="limit_exceeded"}` increments
// when LLM-2 rejects the 11th connection.
func TestNewProxy_LimitExceeded_RecordsMetric(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	lim := NewLimiter(10) // limit=10
	// Exhaust all 10 slots directly via the limiter API.
	for i := 0; i < 10; i++ {
		lim.TryAcquire("tenant-001")
	}

	met := NewSSEMetrics()
	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "completions", lim, met)

	gw := httptest.NewServer(tenantGateway("tenant-001", h))
	t.Cleanup(gw.Close)

	// 11th request → 429 + limit_exceeded metric.
	resp, err := http.Get(gw.URL + "/stream")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status: want 429, got %d", resp.StatusCode)
	}
	if got := met.ErrorCount("tenant-001", "completions", "limit_exceeded"); got != 1 {
		t.Errorf("limit_exceeded counter: want 1, got %d", got)
	}
}

// TestNewProxy_MetricsNil_NoOp verifies that passing nil met disables metrics
// without panicking.
func TestNewProxy_MetricsNil_NoOp(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	u, _ := url.Parse(upstream.URL)
	h, _ := NewProxy(u, "completions", nil, nil) // nil met

	gw := httptest.NewServer(h)
	t.Cleanup(gw.Close)

	resp, err := http.Get(gw.URL + "/")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()
	// No panic = pass.
}
