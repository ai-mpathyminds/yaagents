// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package response provides shared vendor media-type helpers for the yaagents gateway.
//
// Every agentic vendor-typed response body MUST include a trace object per
// Agentic REST Profile §5 (yaagents/spec/agentic-rest-profile.md).
// Schema: yaagents/schemas/v0.1/agentic-error.schema.json
package response

import (
	"encoding/json"
	"net/http"
)

// ContentTypeError is the vendor media type for agentic error responses.
const ContentTypeError = "application/vnd.yaagents.error+json"

// Trace holds cross-service correlation identifiers propagated from the
// inbound request (Agentic REST Profile §5).
type Trace struct {
	CorrelationID string `json:"correlationId"`
	RequestID     string `json:"requestId"`
}

// ErrorBody is the canonical shape for application/vnd.yaagents.error+json.
// type must be one of: "forbidden", "failed_dependency", "error".
type ErrorBody struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Trace   Trace  `json:"trace"`
}

// WriteError writes body as application/vnd.yaagents.error+json with the given
// HTTP status code. It sets Content-Type and calls WriteHeader before encoding.
func WriteError(w http.ResponseWriter, status int, body ErrorBody) {
	w.Header().Set("Content-Type", ContentTypeError)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
