package yaagentsclient

import "encoding/json"

// AgenticResult is the discriminated union returned by every client method.
// Callers may switch on Type (result-style) or call Err() (error-style).
//
// Full Content-Type → field parsing is implemented in GOC-3; this file
// defines the data shapes so GOC-2 resource accessors can compile.
type AgenticResult struct {
	// Type is the logical result kind. Values: "created", "success",
	// "accepted", "clarification_required", "validation_failed",
	// "forbidden", "failed_dependency", "approval_required", "conflict", "error".
	Type string
	// Status is the raw HTTP status code.
	Status int
	// Resource holds the JSON payload for created/success responses.
	Resource json.RawMessage
	// RequiredInputs is populated for clarification_required responses.
	RequiredInputs []RequiredInput
	// OperationID is populated for accepted (202) async responses.
	OperationID string
	// Message is a human-readable description for error responses.
	Message string
	// Trace carries correlation/request IDs from the server.
	Trace Trace
}

// Err returns a typed error for non-success results, nil for success/created/accepted.
// Full implementation is provided in GOC-3; stub always returns nil.
func (r *AgenticResult) Err() error { return nil }

// RequiredInput describes a single clarification field the server requires.
type RequiredInput struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options,omitempty"`
}

// Trace carries the server-side correlation and request IDs for observability.
type Trace struct {
	CorrelationID string `json:"correlationId,omitempty"`
	RequestID     string `json:"requestId,omitempty"`
}
