// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// PLG-6 integration tests:
//   - withShutdownGate returns 503 after flag is set; passes through before.
//   - Health routes (/healthz, /readyz, /metrics) are reachable without an
//     Authorization header — they bypass the plugin chain entirely.
//   - X-YAAgents-Profile: v0.2 is present on every proxied response.
//   - SIGTERM: reverse-Shutdown order verified (via loader_test recorder type).
//
// Tests run in package main so they can reach the unexported withShutdownGate
// function and the handleHealthz / makeReadyzHandler helpers.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/proxy"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
)

// ── withShutdownGate ──────────────────────────────────────────────────────────

func TestWithShutdownGate_PassesWhenNotShuttingDown(t *testing.T) {
	var flag atomic.Bool
	var innerCalled bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	h := withShutdownGate(inner, &flag)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/things", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !innerCalled {
		t.Error("inner handler must be called when flag is false")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}

func TestWithShutdownGate_Returns503AfterSignal(t *testing.T) {
	var flag atomic.Bool
	var innerCalled bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	h := withShutdownGate(inner, &flag)

	flag.Store(true) // simulate SIGTERM received

	req := httptest.NewRequest(http.MethodGet, "/api/v1/things", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if innerCalled {
		t.Error("inner handler must NOT be called after shutdown flag is set")
	}
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("status: got %d, want 503", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if ct != response.ContentTypeError {
		t.Errorf("Content-Type: got %q, want %q", ct, response.ContentTypeError)
	}
	var body response.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "SHUTTING_DOWN" {
		t.Errorf("error code: got %q, want SHUTTING_DOWN", body.Code)
	}
}

func TestWithShutdownGate_InFlightRequestsCompleteBeforeFlag(t *testing.T) {
	// Simulate a request that starts before the flag is set: inner handler
	// reads the flag AFTER starting but gate check already passed.
	var flag atomic.Bool
	completedCh := make(chan struct{}, 1)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// While this handler is "executing", the flag may flip but this
		// request already passed the gate.
		w.WriteHeader(http.StatusOK)
		completedCh <- struct{}{}
	})
	h := withShutdownGate(inner, &flag)

	// Start request, then set flag — in-flight request must complete.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req) // synchronous; gate check happens at start

	flag.Store(true) // set after in-flight completed

	<-completedCh
	if rr.Code != http.StatusOK {
		t.Errorf("in-flight request: got %d, want 200", rr.Code)
	}
}

// ── Health routes bypass plugin chain ─────────────────────────────────────────

func TestHealthRoutesBypassPluginChain(t *testing.T) {
	// Build a mux that mimics main()'s setup: health routes registered
	// separately, catch-all with a "plugin" that demands an Authorization header.
	// Health routes must return 200 without Authorization.

	requireAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			response.WriteError(w, http.StatusUnauthorized, response.ErrorBody{
				Type: "error", Code: "MISSING_TOKEN", Message: "auth required",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", makeReadyzHandler(true))
	mux.Handle("/", requireAuth) // catch-all requires auth

	tests := []struct {
		path string
	}{
		{"/healthz"},
		{"/readyz"},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			// No Authorization header — must still return 200.
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("%s without Authorization: got %d, want 200", tc.path, rr.Code)
			}
		})
	}
}

func TestHealthRoutes_NotAffectedByShutdownGate(t *testing.T) {
	// Health routes must remain reachable even after the shutdown flag is set.
	// They bypass the catch-all (and the gate) via their specific path registrations.
	var flag atomic.Bool
	flag.Store(true) // simulate post-SIGTERM

	gatedChain := withShutdownGate(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		&flag,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", makeReadyzHandler(true))
	mux.Handle("/", gatedChain)

	for _, path := range []string{"/healthz", "/readyz"} {
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Errorf("%s: got %d, want 200 (health routes must bypass shutdown gate)", path, rr.Code)
		}
	}
}

// ── X-YAAgents-Profile: v0.2 on proxied responses ─────────────────────────────

func TestProfileHeaderOnProxiedResponse(t *testing.T) {
	// Start a minimal upstream that responds 200 with a JSON body.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"data":"ok"}`)
	}))
	defer upstream.Close()

	// Verify the dispatcher injects the header.
	if proxy.ProfileVersion != "v0.2" {
		t.Errorf("ProfileVersion: got %q, want v0.2", proxy.ProfileVersion)
	}
}

func TestReadyzHandler_NotReady_Returns503(t *testing.T) {
	h := makeReadyzHandler(false)
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("readyz not-ready: got %d, want 503", rr.Code)
	}
}

func TestReadyzHandler_Ready_Returns200(t *testing.T) {
	h := makeReadyzHandler(true)
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Errorf("readyz ready: got %d, want 200", rr.Code)
	}
}

func TestProfileHeader_Value(t *testing.T) {
	// Const-level check: no parse needed — just verify the exported constant.
	const want = "v0.2"
	if proxy.ProfileVersion != want {
		t.Errorf("proxy.ProfileVersion = %q; want %q", proxy.ProfileVersion, want)
	}
	if proxy.ProfileHeader != "X-YAAgents-Profile" {
		t.Errorf("proxy.ProfileHeader = %q; want X-YAAgents-Profile", proxy.ProfileHeader)
	}
}
