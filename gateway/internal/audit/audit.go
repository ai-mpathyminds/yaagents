// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package audit writes per-request structured JSON audit events to a configured
// sink (stdout or file). One JSON line per event; thread-safe.
//
// Only requests on routes with audit:true produce events (Agentic REST Profile
// ADR PI1-yaa-0001; see WI-1yaa.GW-5).
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Event is one audit record. All fields are populated by the observation
// middleware in the proxy package after the upstream responds.
type Event struct {
	Timestamp     string  `json:"timestamp"`               // RFC3339Nano, UTC
	RouteID       string  `json:"route_id"`
	Method        string  `json:"method"`
	Path          string  `json:"path"`
	TenantID      string  `json:"tenant_id,omitempty"`
	ActorSubject  string  `json:"actor_subject,omitempty"`
	StatusCode    int     `json:"status_code"`
	LatencyMS     float64 `json:"latency_ms"`
	CorrelationID string  `json:"correlation_id"`
	RequestID     string  `json:"request_id"`
}

// Logger writes audit events as newline-delimited JSON to its sink.
// Concurrent calls to Log are serialised by an internal mutex.
type Logger struct {
	mu  sync.Mutex
	enc *json.Encoder
}

// New wraps sink in a thread-safe JSON audit logger.
func New(sink io.Writer) *Logger {
	enc := json.NewEncoder(sink)
	enc.SetEscapeHTML(false) // keep URLs readable
	return &Logger{enc: enc}
}

// Log encodes e as a single JSON line. Thread-safe.
func (l *Logger) Log(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = l.enc.Encode(e)
}

// OpenSink resolves the GATEWAY_AUDIT_LOG value to an io.Writer.
//
//   - "" or "stdout" → os.Stdout (no-op closer)
//   - any other value → opened as an append-only file (mode 0640)
//
// The second return value is a close function; callers MUST invoke it on
// shutdown. For stdout the closer is a no-op.
func OpenSink(path string) (io.Writer, func(), error) {
	if path == "" || path == "stdout" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, nil, fmt.Errorf("opening audit log %q: %w", path, err)
	}
	return f, func() { _ = f.Close() }, nil
}

// nowRFC3339Nano returns the current UTC time in RFC3339Nano format.
// Extracted as a variable so tests can override it.
var nowRFC3339Nano = func() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// Timestamp returns the current formatted time. Called by the observe middleware.
func Timestamp() string { return nowRFC3339Nano() }
