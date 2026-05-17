// Package metrics accumulates per-route request counts and latency sums,
// exposing them in Prometheus text format on GET /metrics.
//
// No external prometheus library is used — text format is written directly
// per ADR PI1-yaa-0001 §2 (net/http only, no heavy framework deps).
//
// Prometheus text format spec: https://prometheus.io/docs/instrumenting/exposition_formats/
package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
)

// labelKey identifies one time series by route ID and HTTP status code string.
type labelKey struct {
	route  string
	status string
}

// Registry accumulates request observations keyed by route+status.
// All methods are thread-safe.
type Registry struct {
	mu         sync.Mutex
	counts     map[labelKey]int64
	latencySum map[string]float64 // route ID → cumulative ms
}

// New returns an empty Registry ready for use.
func New() *Registry {
	return &Registry{
		counts:     make(map[labelKey]int64),
		latencySum: make(map[string]float64),
	}
}

// Record adds one observation: increments the counter for routeID+statusCode
// and adds latencyMS to the route's cumulative latency sum.
func (r *Registry) Record(routeID string, statusCode int, latencyMS float64) {
	k := labelKey{route: routeID, status: fmt.Sprintf("%d", statusCode)}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counts[k]++
	r.latencySum[routeID] += latencyMS
}

// WritePrometheus emits a Prometheus text-format snapshot to w.
//
// Two metric families:
//
//	yaagents_gateway_requests_total{route="…",status="…"} <count>
//	yaagents_gateway_request_duration_ms_total{route="…"} <sum_ms>
//
// Output is sorted for deterministic test comparison.
func (r *Registry) WritePrometheus(w io.Writer) {
	// Copy under lock to minimise hold time.
	r.mu.Lock()
	counts := make(map[labelKey]int64, len(r.counts))
	for k, v := range r.counts {
		counts[k] = v
	}
	latency := make(map[string]float64, len(r.latencySum))
	for k, v := range r.latencySum {
		latency[k] = v
	}
	r.mu.Unlock()

	// Sorted keys for deterministic output.
	keys := make([]labelKey, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		return keys[i].status < keys[j].status
	})

	_, _ = fmt.Fprintln(w, "# HELP yaagents_gateway_requests_total Total proxied requests by route and HTTP status")
	_, _ = fmt.Fprintln(w, "# TYPE yaagents_gateway_requests_total counter")
	for _, k := range keys {
		_, _ = fmt.Fprintf(w, "yaagents_gateway_requests_total{route=%q,status=%q} %d\n",
			k.route, k.status, counts[k])
	}

	routeIDs := make([]string, 0, len(latency))
	for id := range latency {
		routeIDs = append(routeIDs, id)
	}
	sort.Strings(routeIDs)

	_, _ = fmt.Fprintln(w, "# HELP yaagents_gateway_request_duration_ms_total Total request latency in milliseconds by route")
	_, _ = fmt.Fprintln(w, "# TYPE yaagents_gateway_request_duration_ms_total counter")
	for _, id := range routeIDs {
		_, _ = fmt.Fprintf(w, "yaagents_gateway_request_duration_ms_total{route=%q} %.3f\n",
			id, latency[id])
	}
}

// Handler returns an http.HandlerFunc that writes a Prometheus text snapshot.
func (r *Registry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		r.WritePrometheus(w)
	}
}
