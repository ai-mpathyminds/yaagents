// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo_test

import (
	"net/http"
	"testing"

	"github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

func TestFromRequest_AllHeaders(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	r.Header.Set("X-Correlation-ID", "corr-abc-123")
	r.Header.Set("X-Request-ID", "req-def-456")
	r.Header.Set("X-Tenant-ID", "tenant-acme")
	r.Header.Set("X-Actor-Principal", "user@acme.example")

	ctx := sdkgo.FromRequest(r)

	if ctx.CorrelationID != "corr-abc-123" {
		t.Errorf("CorrelationID: got %q, want %q", ctx.CorrelationID, "corr-abc-123")
	}
	if ctx.RequestID != "req-def-456" {
		t.Errorf("RequestID: got %q, want %q", ctx.RequestID, "req-def-456")
	}
	if ctx.ActorTenant != "tenant-acme" {
		t.Errorf("ActorTenant: got %q, want %q", ctx.ActorTenant, "tenant-acme")
	}
	if ctx.Principal != "user@acme.example" {
		t.Errorf("Principal: got %q, want %q", ctx.Principal, "user@acme.example")
	}
}

func TestFromRequest_MissingHeaders(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}

	ctx := sdkgo.FromRequest(r)

	if ctx.CorrelationID != "" {
		t.Errorf("CorrelationID: want empty, got %q", ctx.CorrelationID)
	}
	if ctx.RequestID != "" {
		t.Errorf("RequestID: want empty, got %q", ctx.RequestID)
	}
	if ctx.ActorTenant != "" {
		t.Errorf("ActorTenant: want empty, got %q", ctx.ActorTenant)
	}
	if ctx.Principal != "" {
		t.Errorf("Principal: want empty, got %q", ctx.Principal)
	}
}

func TestFromRequest_PartialHeaders(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatalf("http.NewRequest: %v", err)
	}
	r.Header.Set("X-Correlation-ID", "trace-xyz")
	// X-Request-ID, X-Tenant-ID, X-Actor-Principal intentionally absent

	ctx := sdkgo.FromRequest(r)

	if ctx.CorrelationID != "trace-xyz" {
		t.Errorf("CorrelationID: got %q, want %q", ctx.CorrelationID, "trace-xyz")
	}
	if ctx.RequestID != "" {
		t.Errorf("RequestID: want empty, got %q", ctx.RequestID)
	}
	if ctx.ActorTenant != "" {
		t.Errorf("ActorTenant: want empty, got %q", ctx.ActorTenant)
	}
	if ctx.Principal != "" {
		t.Errorf("Principal: want empty, got %q", ctx.Principal)
	}
}
