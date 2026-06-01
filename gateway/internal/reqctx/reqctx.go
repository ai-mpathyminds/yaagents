// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package reqctx provides typed context keys and accessors for per-request
// values propagated through the yaagents gateway middleware chain:
// correlation ID, request ID, tenant ID, actor subject, and actor roles.
//
// Usage:
//
//	ctx = reqctx.WithCorrelationID(ctx, corrID)
//	id  := reqctx.CorrelationID(ctx)
package reqctx

import (
	"context"
	"crypto/rand"
	"fmt"
)

// key is an unexported type to prevent collisions with other packages.
type key int

const (
	correlationIDKey key = iota
	requestIDKey
	tenantIDKey
	actorSubjectKey
	actorRolesKey
)

// NewUUID generates a random UUID v4 string.
func NewUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// WithCorrelationID stores the correlation ID in ctx.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationID retrieves the correlation ID from ctx (empty string if absent).
func CorrelationID(ctx context.Context) string {
	s, _ := ctx.Value(correlationIDKey).(string)
	return s
}

// WithRequestID stores the request ID in ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID retrieves the request ID from ctx (empty string if absent).
func RequestID(ctx context.Context) string {
	s, _ := ctx.Value(requestIDKey).(string)
	return s
}

// WithTenantID stores the tenant ID in ctx.
func WithTenantID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, tenantIDKey, id)
}

// TenantID retrieves the tenant ID from ctx (empty string if absent).
func TenantID(ctx context.Context) string {
	s, _ := ctx.Value(tenantIDKey).(string)
	return s
}

// WithActorSubject stores the actor subject claim in ctx.
func WithActorSubject(ctx context.Context, sub string) context.Context {
	return context.WithValue(ctx, actorSubjectKey, sub)
}

// ActorSubject retrieves the actor subject from ctx (empty string if absent).
func ActorSubject(ctx context.Context) string {
	s, _ := ctx.Value(actorSubjectKey).(string)
	return s
}

// WithActorRoles stores actor role claims in ctx.
func WithActorRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, actorRolesKey, roles)
}

// ActorRoles retrieves actor roles from ctx (nil if absent).
func ActorRoles(ctx context.Context) []string {
	r, _ := ctx.Value(actorRolesKey).([]string)
	return r
}
