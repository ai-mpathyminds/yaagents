// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

// activeKey labels one active-connections time series (LLM-4).
type activeKey struct{ tenantID, routeID string }

// errorKey labels one error-count time series (LLM-4).
type errorKey struct{ tenantID, routeID, errorKind string }

// SSEMetrics accumulates per-tenant+route SSE connection gauges and error
// counters, exposing them in Prometheus text format via WritePrometheus.
//
// No external prometheus library is used — text format is written directly,
// consistent with the gateway's internal/metrics package convention
// (ADR PI1-yaa-0001 §2: net/http only, no heavy framework deps).
//
// Two metric families (yaagents-canonical names, breaking from ai-platform):
//
//	yaagents_gateway_sse_connections_active  gauge   tenant_id × route_id
//	yaagents_gateway_sse_errors_total        counter tenant_id × route_id × error_kind
//
// error_kind values: client_disconnect | upstream_error | timeout | limit_exceeded
type SSEMetrics struct {
	mu     sync.Mutex
	active map[activeKey]int64
	errors map[errorKey]int64
}

// NewSSEMetrics returns an empty SSEMetrics ready for use.
func NewSSEMetrics() *SSEMetrics {
	return &SSEMetrics{
		active: make(map[activeKey]int64),
		errors: make(map[errorKey]int64),
	}
}

// Inc increments the active-connections gauge for (tenantID, routeID).
// Call on SSE stream entry after the concurrency limit is satisfied (LLM-1).
func (m *SSEMetrics) Inc(tenantID, routeID string) {
	m.mu.Lock()
	m.active[activeKey{tenantID, routeID}]++
	m.mu.Unlock()
}

// Dec decrements the active-connections gauge for (tenantID, routeID).
// Call via defer in the SSE handler so it fires on stream end or disconnect.
func (m *SSEMetrics) Dec(tenantID, routeID string) {
	m.mu.Lock()
	k := activeKey{tenantID, routeID}
	if v := m.active[k] - 1; v <= 0 {
		delete(m.active, k) // prune zero / negative entries
	} else {
		m.active[k] = v
	}
	m.mu.Unlock()
}

// Error increments the error counter for (tenantID, routeID, errorKind).
//
// Accepted errorKind values (LLM-4 spec):
//   - "client_disconnect" — context.Canceled in SSE read loop
//   - "upstream_error"    — non-context transport error
//   - "timeout"           — context.DeadlineExceeded (execution timeout, LLM-3)
//   - "limit_exceeded"    — concurrency limit rejection (LLM-2)
func (m *SSEMetrics) Error(tenantID, routeID, errorKind string) {
	m.mu.Lock()
	m.errors[errorKey{tenantID, routeID, errorKind}]++
	m.mu.Unlock()
}

// ActiveCount returns the current active-connections gauge value for
// (tenantID, routeID). Returns 0 when no active streams exist.
func (m *SSEMetrics) ActiveCount(tenantID, routeID string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active[activeKey{tenantID, routeID}]
}

// ErrorCount returns the accumulated error count for (tenantID, routeID, kind).
func (m *SSEMetrics) ErrorCount(tenantID, routeID, kind string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errors[errorKey{tenantID, routeID, kind}]
}

// WritePrometheus emits both SSE metric families in Prometheus text format.
// Output is sorted for deterministic test comparisons.
//
// Emitted even when all counters are zero so the /metrics endpoint always
// advertises the metric names (helps operators discover available metrics).
func (m *SSEMetrics) WritePrometheus(w io.Writer) {
	m.mu.Lock()
	// Snapshot under lock; write outside lock.
	active := make(map[activeKey]int64, len(m.active))
	for k, v := range m.active {
		active[k] = v
	}
	errors := make(map[errorKey]int64, len(m.errors))
	for k, v := range m.errors {
		errors[k] = v
	}
	m.mu.Unlock()

	// ── yaagents_gateway_sse_connections_active (gauge) ──────────────────────
	_, _ = fmt.Fprintln(w, "# HELP yaagents_gateway_sse_connections_active Active concurrent SSE connections by tenant and route")
	_, _ = fmt.Fprintln(w, "# TYPE yaagents_gateway_sse_connections_active gauge")

	aKeys := make([]activeKey, 0, len(active))
	for k := range active {
		aKeys = append(aKeys, k)
	}
	sort.Slice(aKeys, func(i, j int) bool {
		if aKeys[i].tenantID != aKeys[j].tenantID {
			return aKeys[i].tenantID < aKeys[j].tenantID
		}
		return aKeys[i].routeID < aKeys[j].routeID
	})
	for _, k := range aKeys {
		_, _ = fmt.Fprintf(w, "yaagents_gateway_sse_connections_active{tenant_id=%q,route_id=%q} %d\n",
			k.tenantID, k.routeID, active[k])
	}

	// ── yaagents_gateway_sse_errors_total (counter) ───────────────────────────
	_, _ = fmt.Fprintln(w, "# HELP yaagents_gateway_sse_errors_total SSE error events by tenant, route, and error kind")
	_, _ = fmt.Fprintln(w, "# TYPE yaagents_gateway_sse_errors_total counter")

	eKeys := make([]errorKey, 0, len(errors))
	for k := range errors {
		eKeys = append(eKeys, k)
	}
	sort.Slice(eKeys, func(i, j int) bool {
		if eKeys[i].tenantID != eKeys[j].tenantID {
			return eKeys[i].tenantID < eKeys[j].tenantID
		}
		if eKeys[i].routeID != eKeys[j].routeID {
			return eKeys[i].routeID < eKeys[j].routeID
		}
		return eKeys[i].errorKind < eKeys[j].errorKind
	})
	for _, k := range eKeys {
		_, _ = fmt.Fprintf(w, "yaagents_gateway_sse_errors_total{tenant_id=%q,route_id=%q,error_kind=%q} %d\n",
			k.tenantID, k.routeID, k.errorKind, errors[k])
	}
}
