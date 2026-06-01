package yaagentsclient

import "encoding/json"

// AgenticResult is the discriminated union returned by every client method.
// Callers may switch on Type (result-style) or call Err() (error-style).
//
// GOC-3: full parsing wired; err field populated by parseResponse.
type AgenticResult struct {
	// Type is the logical result kind per spec §4:
	// "success", "created", "accepted", "clarification_required",
	// "validation_failed", "approval_required", "forbidden", "conflict",
	// "failed_dependency", "error".
	Type string
	// Status is the raw HTTP status code.
	Status int
	// Resource holds the JSON payload for success/created responses.
	Resource json.RawMessage
	// RequiredInputs is populated for clarification_required responses.
	RequiredInputs []RequiredInput
	// OperationID is populated for accepted (202) async responses.
	OperationID string
	// Message is a human-readable description (error responses).
	Message string
	// Trace carries correlation/request IDs from the server.
	Trace Trace

	// err is the typed error stored by parseResponse; nil for
	// success/created/accepted. Accessed only via Err().
	err error
}

// Err returns the typed error for non-success results (4xx/5xx), nil for
// success (200), created (201), and accepted (202).
// Supports result-style callers: switch on Type, then call Err() to obtain
// the typed error for errors.As / errors.Is.
func (r *AgenticResult) Err() error {
	if r == nil {
		return nil
	}
	return r.err
}

// RequiredInput describes a single clarification field the server requires.
// Wire format matches spec/agentic-rest-profile.md §4.1.
type RequiredInput struct {
	Name          string   `json:"name"`
	Location      string   `json:"location,omitempty"`
	Type          string   `json:"type"`
	Required      bool     `json:"required,omitempty"`
	Question      string   `json:"question,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
}

// Trace carries the server-side correlation and request IDs for observability.
type Trace struct {
	CorrelationID string `json:"correlationId,omitempty"`
	RequestID     string `json:"requestId,omitempty"`
}
