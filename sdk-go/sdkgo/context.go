// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo

import "net/http"

// AgenticContext holds the gateway-injected request metadata propagated via
// HTTP headers from the YAAgents gateway (token-validator + tenant-injector
// middleware chain).
type AgenticContext struct {
	// CorrelationID is the end-to-end trace identifier injected by the gateway
	// from the X-Correlation-ID header.
	CorrelationID string

	// RequestID is the per-request unique identifier injected by the gateway
	// from the X-Request-ID header.
	RequestID string

	// ActorTenant is the resolved tenant identifier injected by the gateway
	// from the X-Tenant-ID header.
	ActorTenant string

	// Principal is the authenticated actor identity injected by the gateway
	// from the X-Actor-Principal header.
	Principal string
}

// FromRequest extracts gateway-injected context from request headers per PRD §5.10.1.
// Header names match the gateway (token-validator + tenant-injector) injection contract:
//
//	X-Correlation-ID  → AgenticContext.CorrelationID
//	X-Request-ID      → AgenticContext.RequestID
//	X-Tenant-ID       → AgenticContext.ActorTenant
//	X-Actor-Principal → AgenticContext.Principal
//
// Missing headers produce empty-string fields; the function never panics.
func FromRequest(r *http.Request) AgenticContext {
	return AgenticContext{
		CorrelationID: r.Header.Get("X-Correlation-ID"),
		RequestID:     r.Header.Get("X-Request-ID"),
		ActorTenant:   r.Header.Get("X-Tenant-ID"),
		Principal:     r.Header.Get("X-Actor-Principal"),
	}
}
