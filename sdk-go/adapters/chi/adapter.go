// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package chi is the go-chi/chi/v5 adapter for the YAAgents Go SDK.
// It re-exports the core sdkgo API and adds a URLParam helper backed by
// chi's URL-parameter extraction, so callers need only one import.
//
// Usage:
//
//	import chi "github.com/ai-mpathyminds/yaagents-sdk-go/adapters/chi"
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    ctx := chi.FromRequest(r)
//	    id  := chi.URLParam(r, "id")
//	    ar  := chi.AgenticResponse{}
//	    chi.Write(w, ar.Created(ctx, payload))
//	}
package chi

import (
	"net/http"

	gochi "github.com/go-chi/chi/v5"

	sdkgo "github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

// URLParam returns the URL routing parameter named key from the chi router
// context stored in r. The chi router must be the one that handled r.
func URLParam(r *http.Request, key string) string {
	return gochi.URLParam(r, key)
}

// FromRequest extracts the AgenticContext from gateway-injected request headers.
// Delegates to sdkgo.FromRequest.
func FromRequest(r *http.Request) sdkgo.AgenticContext { return sdkgo.FromRequest(r) }

// Write serializes resp to w, setting HTTP status, Content-Type, and
// X-YAAgents-Profile: v0.3. Delegates to sdkgo.Write.
func Write(w http.ResponseWriter, resp sdkgo.AgenticWritable) error {
	return sdkgo.Write(w, resp)
}

// Type aliases — importing this package gives access to the full SDK surface.
type (
	AgenticContext  = sdkgo.AgenticContext
	AgenticResponse = sdkgo.AgenticResponse
	AgenticWritable = sdkgo.AgenticWritable
	RequiredInput   = sdkgo.RequiredInput
	ValidationError = sdkgo.ValidationError
)
