// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package otelaudit

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// ── slog capture helpers ──────────────────────────────────────────────────────

type warnCounter struct {
	target string
	count  atomic.Int64
}

func (h *warnCounter) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= slog.LevelWarn
}
func (h *warnCounter) Handle(_ context.Context, r slog.Record) error {
	if strings.Contains(r.Message, h.target) {
		h.count.Add(1)
	}
	return nil
}
func (h *warnCounter) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *warnCounter) WithGroup(_ string) slog.Handler      { return h }

func setLogger(t *testing.T, h slog.Handler) {
	t.Helper()
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newPlugin(t *testing.T, cfg map[string]any) *OtelAudit {
	t.Helper()
	a := &OtelAudit{}
	if err := a.Init(plugin.NewMapConfig(cfg)); err != nil {
		t.Fatalf("Init: unexpected error: %v", err)
	}
	return a
}

func passThrough() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// ── Name ─────────────────────────────────────────────────────────────────────

func TestName(t *testing.T) {
	a := &OtelAudit{}
	if got := a.Name(); got != "otel-audit" {
		t.Errorf("Name(): got %q, want %q", got, "otel-audit")
	}
}

// ── Init — validation ─────────────────────────────────────────────────────────

func TestInit_MalformedEndpoint_NoScheme(t *testing.T) {
	a := &OtelAudit{}
	err := a.Init(plugin.NewMapConfig(map[string]any{
		"enabled":  true,
		"endpoint": "not-a-url",
	}))
	if err == nil {
		t.Fatal("Init with no-scheme endpoint: expected non-nil error, got nil")
	}
}

func TestInit_MalformedEndpoint_NoHost(t *testing.T) {
	a := &OtelAudit{}
	err := a.Init(plugin.NewMapConfig(map[string]any{
		"enabled":  true,
		"endpoint": "grpc://",
	}))
	if err == nil {
		t.Fatal("Init with empty-host endpoint: expected non-nil error, got nil")
	}
}

func TestInit_EmptyEndpoint_OK(t *testing.T) {
	a := &OtelAudit{}
	err := a.Init(plugin.NewMapConfig(map[string]any{"enabled": true}))
	if err != nil {
		t.Fatalf("Init with empty endpoint: unexpected error: %v", err)
	}
}

func TestInit_ValidEndpoint_OK(t *testing.T) {
	a := &OtelAudit{}
	err := a.Init(plugin.NewMapConfig(map[string]any{
		"enabled":  true,
		"endpoint": "http://otel-collector:4317",
	}))
	if err != nil {
		t.Fatalf("Init with valid endpoint: unexpected error: %v", err)
	}
	if a.endpoint != "http://otel-collector:4317" {
		t.Errorf("endpoint stored: got %q, want %q", a.endpoint, "http://otel-collector:4317")
	}
}

func TestInit_Disabled_EmptyEndpoint(t *testing.T) {
	a := &OtelAudit{}
	err := a.Init(plugin.NewMapConfig(map[string]any{"enabled": false}))
	if err != nil {
		t.Fatalf("Init(disabled, no endpoint): unexpected error: %v", err)
	}
}

// ── Handler — enabled + empty endpoint → warn-once, pass-through ──────────────

func TestHandler_EnabledNoEndpoint_WarnOnce(t *testing.T) {
	h := &warnCounter{target: "otel-audit exporter not configured"}
	setLogger(t, h)

	a := newPlugin(t, map[string]any{"enabled": true})

	handler := a.Handler(passThrough())
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: status %d, want 200", i, rr.Code)
		}
	}
	if n := h.count.Load(); n != 1 {
		t.Errorf("warn count: got %d, want 1 (sync.Once)", n)
	}
}

func TestHandler_EnabledNoEndpoint_WarnOnce_Concurrent(t *testing.T) {
	h := &warnCounter{target: "otel-audit exporter not configured"}
	setLogger(t, h)

	a := newPlugin(t, map[string]any{"enabled": true})
	var called atomic.Int64
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	handler := a.Handler(upstream)

	const goroutines = 40
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			handler.ServeHTTP(httptest.NewRecorder(),
				httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()

	if c := called.Load(); c != goroutines {
		t.Errorf("upstream calls: got %d, want %d", c, goroutines)
	}
	if n := h.count.Load(); n != 1 {
		t.Errorf("warn count: got %d, want 1 (concurrent sync.Once)", n)
	}
}

// ── Handler — enabled + valid endpoint → no warn, pass-through ───────────────

func TestHandler_EnabledWithEndpoint_NoWarn(t *testing.T) {
	h := &warnCounter{target: "otel-audit exporter not configured"}
	setLogger(t, h)

	a := newPlugin(t, map[string]any{
		"enabled":  true,
		"endpoint": "grpc://otel-collector:4317",
	})
	handler := a.Handler(passThrough())

	for i := 0; i < 5; i++ {
		handler.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if n := h.count.Load(); n != 0 {
		t.Errorf("warn count: got %d, want 0 (endpoint configured)", n)
	}
}

// ── Handler — disabled → no warn, pass-through ───────────────────────────────

func TestHandler_Disabled_NoWarn_PassThrough(t *testing.T) {
	h := &warnCounter{target: "otel-audit exporter not configured"}
	setLogger(t, h)

	a := newPlugin(t, map[string]any{"enabled": false})
	var called atomic.Int64
	upstream := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	handler := a.Handler(upstream)

	for i := 0; i < 5; i++ {
		handler.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if n := h.count.Load(); n != 0 {
		t.Errorf("warn count: got %d, want 0 for disabled plugin", n)
	}
	if c := called.Load(); c != 5 {
		t.Errorf("upstream calls: got %d, want 5", c)
	}
}

// ── Handler — noop tracer propagates context ──────────────────────────────────

func TestHandler_ContextPropagated(t *testing.T) {
	a := newPlugin(t, map[string]any{"enabled": false})

	type ctxKey struct{}
	var gotCtx context.Context
	upstream := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotCtx = r.Context()
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "sentinel"))
	a.Handler(upstream).ServeHTTP(httptest.NewRecorder(), req)

	if gotCtx == nil {
		t.Fatal("context not propagated to upstream")
	}
	if gotCtx.Value(ctxKey{}) != "sentinel" {
		t.Error("original context values must survive noop span injection")
	}
}

// ── noopTracer ────────────────────────────────────────────────────────────────

func TestNoopTracer_StartSpan(t *testing.T) {
	ctx := context.Background()
	tr := noopTracer{}
	gotCtx, end := tr.startSpan(ctx, "test.span")
	if gotCtx == nil {
		t.Error("startSpan must return non-nil context")
	}
	end() // must not panic
}

// ── Shutdown ──────────────────────────────────────────────────────────────────

func TestShutdown_NoError(t *testing.T) {
	a := newPlugin(t, map[string]any{"enabled": true})
	if err := a.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}
