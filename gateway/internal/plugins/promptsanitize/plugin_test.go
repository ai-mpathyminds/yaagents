// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package promptsanitize

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

// warnCounter is a minimal slog.Handler that counts Warn-level messages
// containing a target substring.
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

// setLogger installs h as the default slog logger for the duration of t and
// restores the previous default in t.Cleanup.
func setLogger(t *testing.T, h slog.Handler) {
	t.Helper()
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newPlugin(t *testing.T, enabled bool) *PromptSanitize {
	t.Helper()
	p := &PromptSanitize{}
	cfg := plugin.NewMapConfig(map[string]any{"enabled": enabled})
	if err := p.Init(cfg); err != nil {
		t.Fatalf("Init: unexpected error: %v", err)
	}
	return p
}

// passCounter counts how many times the upstream handler is called.
func passCounter() (http.Handler, *atomic.Int64) {
	var n atomic.Int64
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n.Add(1)
		w.WriteHeader(http.StatusOK)
	}), &n
}

// ── Name ─────────────────────────────────────────────────────────────────────

func TestName(t *testing.T) {
	p := &PromptSanitize{}
	if got := p.Name(); got != "prompt-sanitize" {
		t.Errorf("Name(): got %q, want %q", got, "prompt-sanitize")
	}
}

// ── Init ─────────────────────────────────────────────────────────────────────

func TestInit_Disabled(t *testing.T) {
	p := &PromptSanitize{}
	err := p.Init(plugin.NewMapConfig(map[string]any{"enabled": false}))
	if err != nil {
		t.Fatalf("Init(disabled): unexpected error: %v", err)
	}
	if p.enabled {
		t.Error("enabled should be false")
	}
}

func TestInit_Enabled(t *testing.T) {
	p := &PromptSanitize{}
	err := p.Init(plugin.NewMapConfig(map[string]any{"enabled": true}))
	if err != nil {
		t.Fatalf("Init(enabled): unexpected error: %v", err)
	}
	if !p.enabled {
		t.Error("enabled should be true")
	}
}

func TestInit_EmptyConfig(t *testing.T) {
	p := &PromptSanitize{}
	if err := p.Init(plugin.NewMapConfig(nil)); err != nil {
		t.Fatalf("Init(empty): unexpected error: %v", err)
	}
}

// ── Handler — disabled: pass-through, no warn ────────────────────────────────

func TestHandler_Disabled_PassThrough(t *testing.T) {
	p := newPlugin(t, false)
	upstream, calls := passCounter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/optimizations", nil)
	rr := httptest.NewRecorder()
	p.Handler(upstream).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if calls.Load() != 1 {
		t.Errorf("upstream calls: got %d, want 1", calls.Load())
	}
}

// ── Handler — enabled: pass-through on every request ─────────────────────────

func TestHandler_Enabled_PassThrough(t *testing.T) {
	p := newPlugin(t, true)
	upstream, calls := passCounter()

	const n = 5
	for i := 0; i < n; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/optimizations", nil)
		rr := httptest.NewRecorder()
		p.Handler(upstream).ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: status %d, want 200", i, rr.Code)
		}
	}
	if c := calls.Load(); c != n {
		t.Errorf("upstream calls: got %d, want %d", c, n)
	}
}

// ── Handler — enabled: warn emitted exactly once (sync.Once) ─────────────────

func TestHandler_Enabled_WarnOnce_Sequential(t *testing.T) {
	h := &warnCounter{target: "prompt-sanitize is a stub"}
	setLogger(t, h)

	p := newPlugin(t, true)
	upstream, _ := passCounter()

	handler := p.Handler(upstream)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	if n := h.count.Load(); n != 1 {
		t.Errorf("warn count: got %d, want 1 (sync.Once must gate repeated emission)", n)
	}
}

func TestHandler_Enabled_WarnOnce_Concurrent(t *testing.T) {
	h := &warnCounter{target: "prompt-sanitize is a stub"}
	setLogger(t, h)

	p := newPlugin(t, true)
	upstream, calls := passCounter()
	handler := p.Handler(upstream)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}()
	}
	wg.Wait()

	// All requests must pass through.
	if c := calls.Load(); c != goroutines {
		t.Errorf("upstream calls: got %d, want %d", c, goroutines)
	}
	// Warn must fire exactly once despite concurrent requests.
	if n := h.count.Load(); n != 1 {
		t.Errorf("warn count: got %d, want 1 (sync.Once must prevent concurrent double-emit)", n)
	}
}

// Disabled plugin emits no warn even after many requests.
func TestHandler_Disabled_NoWarn(t *testing.T) {
	h := &warnCounter{target: "prompt-sanitize is a stub"}
	setLogger(t, h)

	p := newPlugin(t, false)
	upstream, _ := passCounter()
	handler := p.Handler(upstream)

	for i := 0; i < 5; i++ {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if n := h.count.Load(); n != 0 {
		t.Errorf("warn count: got %d, want 0 for disabled plugin", n)
	}
}

// ── Shutdown ──────────────────────────────────────────────────────────────────

func TestShutdown_NoError(t *testing.T) {
	p := newPlugin(t, true)
	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown: unexpected error: %v", err)
	}
}
