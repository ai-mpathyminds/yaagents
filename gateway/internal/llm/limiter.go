// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package llm

import (
	"sync"
	"sync/atomic"
)

// Limiter is a per-tenant SSE concurrency counter (WI-2yaa.LLM-2).
//
// It tracks the number of active SSE connections per tenant using a
// sync.Map[string]*atomic.Int64. The limit is configured at construction time
// via gateway.llm.max_sse_connections_per_tenant (env GATEWAY_LLM_MAX_SSE_PER_TENANT,
// default 10 per WI-2yaa.LLM-2).
//
// TryAcquire/Release are safe for concurrent use. The acquire-before-check
// pattern (increment then compare) avoids races between counter reads and
// increments under concurrent load.
type Limiter struct {
	maxPerTenant int64
	counts       sync.Map // map[string]*atomic.Int64
}

// NewLimiter returns a Limiter that allows at most maxPerTenant concurrent SSE
// connections per tenant. Values ≤ 0 are clamped to the default of 10.
func NewLimiter(maxPerTenant int) *Limiter {
	if maxPerTenant <= 0 {
		maxPerTenant = 10
	}
	return &Limiter{maxPerTenant: int64(maxPerTenant)}
}

// TryAcquire attempts to acquire one SSE slot for tenantID.
// It increments the counter first (atomic); if the new value exceeds the
// limit the increment is immediately reversed and false is returned.
// Callers that receive true MUST call Release(tenantID) exactly once —
// typically via defer inside a sync.Once to prevent double-decrement.
func (l *Limiter) TryAcquire(tenantID string) bool {
	c := l.counter(tenantID)
	if c.Add(1) > l.maxPerTenant {
		c.Add(-1) // undo; counter stays at the limit
		return false
	}
	return true
}

// Release decrements the SSE slot counter for tenantID. It is idempotent
// when called via sync.Once (the recommended pattern in the SSE handler).
// Calling Release more times than TryAcquire returned true is a logic error
// and will drive the counter negative.
func (l *Limiter) Release(tenantID string) {
	l.counter(tenantID).Add(-1)
}

// Count returns the current active SSE count for tenantID. Zero is returned
// for tenants that have never called TryAcquire.
func (l *Limiter) Count(tenantID string) int64 {
	v, ok := l.counts.Load(tenantID)
	if !ok {
		return 0
	}
	return v.(*atomic.Int64).Load()
}

// counter returns (or creates) the atomic counter for tenantID.
func (l *Limiter) counter(tenantID string) *atomic.Int64 {
	v, _ := l.counts.LoadOrStore(tenantID, new(atomic.Int64))
	return v.(*atomic.Int64)
}
