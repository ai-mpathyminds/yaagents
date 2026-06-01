package yaagentsclient

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Typed errors — one per non-success row in the Agentic REST Response Profile
// (spec/agentic-rest-profile.md §4 normative table).
//
// Each type:
//   • implements error (Error() string)
//   • carries a Trace block for observability
//   • is compatible with errors.As / errors.Is
// ─────────────────────────────────────────────────────────────────────────────

// ClarificationRequired is returned on HTTP 400
// Content-Type: application/vnd.yaagents.clarification+json.
type ClarificationRequired struct {
	RequiredInputs []RequiredInput
	Trace          Trace
}

func (e *ClarificationRequired) Error() string {
	names := make([]string, len(e.RequiredInputs))
	for i, ri := range e.RequiredInputs {
		names[i] = ri.Name
	}
	return fmt.Sprintf("clarification required: fields [%s]", strings.Join(names, ", "))
}

// ValidationFailed is returned on HTTP 422
// Content-Type: application/vnd.yaagents.validation-error+json.
type ValidationFailed struct {
	Errors []ValidationError
	Trace  Trace
}

func (e *ValidationFailed) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s: %s", e.Errors[0].Field, e.Errors[0].Message)
}

// ValidationError is a single field-level validation failure within ValidationFailed.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AgenticForbidden is returned on HTTP 403
// Content-Type: application/vnd.yaagents.error+json.
type AgenticForbidden struct {
	Message string
	Trace   Trace
}

func (e *AgenticForbidden) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Message)
}

// FailedDependency is returned on HTTP 424
// Content-Type: application/vnd.yaagents.error+json.
// Dependency holds the JSON "code" field — it identifies the failing upstream.
type FailedDependency struct {
	Dependency string // maps to JSON "code"
	Message    string
	Trace      Trace
}

func (e *FailedDependency) Error() string {
	return fmt.Sprintf("failed dependency (%s): %s", e.Dependency, e.Message)
}

// Conflict is returned on HTTP 409
// Content-Type: application/vnd.yaagents.conflict+json.
type Conflict struct {
	Code                  string
	Message               string
	ConflictingResourceID string
	Trace                 Trace
}

func (e *Conflict) Error() string {
	if e.ConflictingResourceID != "" {
		return fmt.Sprintf("conflict (%s) on %s: %s", e.Code, e.ConflictingResourceID, e.Message)
	}
	return fmt.Sprintf("conflict (%s): %s", e.Code, e.Message)
}

// ApprovalRequired is returned on HTTP 412
// Content-Type: application/vnd.yaagents.approval-required+json.
type ApprovalRequired struct {
	Code          string
	Message       string
	ApprovalToken string
	Trace         Trace
}

func (e *ApprovalRequired) Error() string {
	return fmt.Sprintf("approval required (%s): %s", e.Code, e.Message)
}

// AgenticError is returned on HTTP 500 and for internal client errors such as
// DESERIALIZE_ERROR (malformed or truncated response body).
type AgenticError struct {
	Code    string
	Message string
	Trace   Trace
}

func (e *AgenticError) Error() string {
	return fmt.Sprintf("agentic error (%s): %s", e.Code, e.Message)
}
