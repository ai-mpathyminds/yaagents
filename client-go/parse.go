package yaagentsclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Media-type constants from spec/agentic-rest-profile.md §4 (normative table).
const (
	ctJSON             = "application/json"
	ctOperation        = "application/vnd.yaagents.operation+json"
	ctClarification    = "application/vnd.yaagents.clarification+json"
	ctValidationError  = "application/vnd.yaagents.validation-error+json"
	ctError            = "application/vnd.yaagents.error+json"
	ctApprovalRequired = "application/vnd.yaagents.approval-required+json"
	ctConflict         = "application/vnd.yaagents.conflict+json"
)

// mediaType strips Content-Type parameters and returns the bare media type.
//
//	"application/json; charset=utf-8" → "application/json"
//	"" → ""
func mediaType(ct string) string {
	mt, _, _ := strings.Cut(ct, ";")
	return strings.TrimSpace(mt)
}

// parseResponse maps an HTTP response to an *AgenticResult using the 10-row
// Agentic REST Response Profile table (spec/agentic-rest-profile.md §4).
//
// The caller (do) is responsible for closing resp.Body after parseResponse
// returns; the function reads but does not close the body.
//
// Contract:
//   - success (200) / created (201) → (*AgenticResult, nil); result.Err() == nil
//   - accepted (202) → (*AgenticResult, nil); result.Err() == nil
//   - all 4xx/5xx vendor types → (*AgenticResult, typedError); result.Err() == typedError
//   - any JSON parse failure → (*AgenticResult, &AgenticError{Code:"DESERIALIZE_ERROR"})
func parseResponse(resp *http.Response) (*AgenticResult, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ae := &AgenticError{Code: "DESERIALIZE_ERROR", Message: err.Error()}
		return &AgenticResult{Status: resp.StatusCode, Type: "error", Message: ae.Message, err: ae}, ae
	}

	ct := mediaType(resp.Header.Get("Content-Type"))
	status := resp.StatusCode

	// deErr wraps a JSON unmarshal error as DESERIALIZE_ERROR.
	deErr := func(e error) (*AgenticResult, error) {
		ae := &AgenticError{Code: "DESERIALIZE_ERROR", Message: e.Error()}
		return &AgenticResult{Status: status, Type: "error", Message: ae.Message, err: ae}, ae
	}

	switch ct {

	// ── 200 success / 201 created ────────────────────────────────────────
	case ctJSON:
		if status >= 200 && status < 300 {
			t := "success"
			if status == 201 {
				t = "created"
			}
			return &AgenticResult{Type: t, Status: status, Resource: json.RawMessage(body)}, nil
		}
		// application/json with a non-2xx status is outside the spec.
		ae := &AgenticError{
			Code:    "UNEXPECTED_STATUS",
			Message: fmt.Sprintf("HTTP %d with %s", status, ctJSON),
		}
		return &AgenticResult{Status: status, Type: "error", Message: ae.Message, err: ae}, ae

	// ── 202 accepted ─────────────────────────────────────────────────────
	case ctOperation:
		var wire struct {
			OperationID string `json:"operationId"`
			Message     string `json:"message"`
			Trace       Trace  `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		return &AgenticResult{
			Type: "accepted", Status: status,
			OperationID: wire.OperationID, Message: wire.Message, Trace: wire.Trace,
		}, nil

	// ── 400 clarification_required ───────────────────────────────────────
	case ctClarification:
		var wire struct {
			RequiredInputs []RequiredInput `json:"requiredInputs"`
			Message        string          `json:"message"`
			Trace          Trace           `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		typed := &ClarificationRequired{RequiredInputs: wire.RequiredInputs, Trace: wire.Trace}
		r := &AgenticResult{
			Type: "clarification_required", Status: status,
			RequiredInputs: wire.RequiredInputs, Message: wire.Message, Trace: wire.Trace,
			err: typed,
		}
		return r, typed

	// ── 422 validation_failed ────────────────────────────────────────────
	case ctValidationError:
		var wire struct {
			Errors  []ValidationError `json:"errors"`
			Message string            `json:"message"`
			Trace   Trace             `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		typed := &ValidationFailed{Errors: wire.Errors, Trace: wire.Trace}
		r := &AgenticResult{
			Type: "validation_failed", Status: status,
			Message: wire.Message, Trace: wire.Trace,
			err: typed,
		}
		return r, typed

	// ── 403 forbidden / 424 failed_dependency / 500 error ────────────────
	case ctError:
		var wire struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Trace   Trace  `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		switch status {
		case http.StatusForbidden: // 403
			typed := &AgenticForbidden{Message: wire.Message, Trace: wire.Trace}
			r := &AgenticResult{Type: "forbidden", Status: status, Message: wire.Message, Trace: wire.Trace, err: typed}
			return r, typed
		case http.StatusFailedDependency: // 424
			typed := &FailedDependency{Dependency: wire.Code, Message: wire.Message, Trace: wire.Trace}
			r := &AgenticResult{Type: "failed_dependency", Status: status, Message: wire.Message, Trace: wire.Trace, err: typed}
			return r, typed
		default: // 500 + anything else carrying the error media type
			typed := &AgenticError{Code: wire.Code, Message: wire.Message, Trace: wire.Trace}
			r := &AgenticResult{Type: "error", Status: status, Message: wire.Message, Trace: wire.Trace, err: typed}
			return r, typed
		}

	// ── 412 approval_required ────────────────────────────────────────────
	case ctApprovalRequired:
		var wire struct {
			Code          string `json:"code"`
			Message       string `json:"message"`
			ApprovalToken string `json:"approvalToken"`
			Trace         Trace  `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		typed := &ApprovalRequired{Code: wire.Code, Message: wire.Message, ApprovalToken: wire.ApprovalToken, Trace: wire.Trace}
		r := &AgenticResult{Type: "approval_required", Status: status, Message: wire.Message, Trace: wire.Trace, err: typed}
		return r, typed

	// ── 409 conflict ─────────────────────────────────────────────────────
	case ctConflict:
		var wire struct {
			Code                  string `json:"code"`
			Message               string `json:"message"`
			ConflictingResourceID string `json:"conflictingResourceId"`
			Trace                 Trace  `json:"trace"`
		}
		if err := json.Unmarshal(body, &wire); err != nil {
			return deErr(err)
		}
		typed := &Conflict{
			Code: wire.Code, Message: wire.Message,
			ConflictingResourceID: wire.ConflictingResourceID, Trace: wire.Trace,
		}
		r := &AgenticResult{Type: "conflict", Status: status, Message: wire.Message, Trace: wire.Trace, err: typed}
		return r, typed

	// ── unknown / missing Content-Type ───────────────────────────────────
	default:
		// For 2xx with unknown/absent CT (e.g. bare test servers), treat as
		// success so resource-accessor tests and simple backends remain green.
		if status >= 200 && status < 300 {
			t := "success"
			if status == 201 {
				t = "created"
			}
			return &AgenticResult{Type: t, Status: status, Resource: json.RawMessage(body)}, nil
		}
		ae := &AgenticError{
			Code:    "UNKNOWN_CONTENT_TYPE",
			Message: fmt.Sprintf("unexpected Content-Type %q (HTTP %d)", ct, status),
		}
		return &AgenticResult{Status: status, Type: "error", Message: ae.Message, err: ae}, ae
	}
}
