// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
)

// Content-Type header value constants for YAAgents agentic vendor media types.
// Usage as Content-Type header values is permitted per spec §8 rule 7.
const (
	ctOperation        = "application/vnd.yaagents.operation+json"
	ctClarification    = "application/vnd.yaagents.clarification+json"
	ctValidationError  = "application/vnd.yaagents.validation-error+json"
	ctApprovalRequired = "application/vnd.yaagents.approval-required+json"
	ctError            = "application/vnd.yaagents.error+json"
	ctConflict         = "application/vnd.yaagents.conflict+json"
	ctJSON             = "application/json"
)

// AgenticWritable is the interface that all agentic response factory values
// implement. Pass the value to Write(); do not inspect it directly.
type AgenticWritable interface {
	Status() int
	ContentType() string
	Body() ([]byte, error)
}

// agenticResp is the internal implementation of AgenticWritable shared by
// all ten response factories.
type agenticResp struct {
	status      int
	contentType string
	body        any
}

func (r agenticResp) Status() int           { return r.status }
func (r agenticResp) ContentType() string   { return r.contentType }
func (r agenticResp) Body() ([]byte, error) { return json.Marshal(r.body) }

// traceFrom builds a Trace from an AgenticContext.
func traceFrom(ctx AgenticContext) Trace {
	return Trace{
		CorrelationID: ctx.CorrelationID,
		RequestID:     ctx.RequestID,
	}
}

// newUUID generates a random UUID v4 using crypto/rand (per PRD §5.10 allowed imports).
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// AgenticResponse is the zero-value factory for all ten profile response types.
// Usage:
//
//	var ar sdkgo.AgenticResponse
//	sdkgo.Write(w, ar.Created(ctx, payload))
type AgenticResponse struct{}

// Accepted returns a 202 application/vnd.yaagents.operation+json response.
// operationID is a stable identifier the client uses to poll for status.
func (AgenticResponse) Accepted(ctx AgenticContext, operationID string) AgenticWritable {
	return agenticResp{
		status:      http.StatusAccepted,
		contentType: ctOperation,
		body: OperationAccepted{
			Type:        "operation_accepted",
			Code:        "OPERATION_ACCEPTED",
			Message:     "Operation accepted for asynchronous processing.",
			OperationID: operationID,
			StatusURL:   "",
			Trace:       traceFrom(ctx),
		},
	}
}

// Done returns a 200 application/json response.
// body is the domain-specific success payload and is marshaled directly.
func (AgenticResponse) Done(_ AgenticContext, body any) AgenticWritable {
	return agenticResp{
		status:      http.StatusOK,
		contentType: ctJSON,
		body:        body,
	}
}

// Created returns a 201 application/json response.
// body is the domain-specific created-resource payload and is marshaled directly.
func (AgenticResponse) Created(_ AgenticContext, body any) AgenticWritable {
	return agenticResp{
		status:      http.StatusCreated,
		contentType: ctJSON,
		body:        body,
	}
}

// Failed returns a 500 application/vnd.yaagents.error+json response.
func (AgenticResponse) Failed(ctx AgenticContext, message string) AgenticWritable {
	return agenticResp{
		status:      http.StatusInternalServerError,
		contentType: ctError,
		body: AgenticErrorBody{
			Type:    "error",
			Code:    "INTERNAL_ERROR",
			Message: message,
			Trace:   traceFrom(ctx),
		},
	}
}

// ClarificationRequired returns a 400 application/vnd.yaagents.clarification+json response.
// inputs describes the fields the caller must supply before the operation can proceed.
func (AgenticResponse) ClarificationRequired(ctx AgenticContext, inputs []RequiredInput) AgenticWritable {
	return agenticResp{
		status:      http.StatusBadRequest,
		contentType: ctClarification,
		body: ClarificationRequiredBody{
			Type:           "clarification_required",
			Code:           "CLARIFICATION_REQUIRED",
			Message:        "Additional information is required.",
			RequiredInputs: inputs,
			Trace:          traceFrom(ctx),
		},
	}
}

// ValidationFailed returns a 422 application/vnd.yaagents.validation-error+json response.
func (AgenticResponse) ValidationFailed(ctx AgenticContext, errors []ValidationError) AgenticWritable {
	return agenticResp{
		status:      http.StatusUnprocessableEntity,
		contentType: ctValidationError,
		body: ValidationFailedBody{
			Type:    "validation_failed",
			Code:    "VALIDATION_FAILED",
			Message: "Request validation failed.",
			Errors:  errors,
			Trace:   traceFrom(ctx),
		},
	}
}

// ApprovalRequired returns a 412 application/vnd.yaagents.approval-required+json response.
// An opaque approvalToken (UUID v4) is generated and included in the body.
// approvers is informational; reason becomes the human-readable message.
func (AgenticResponse) ApprovalRequired(ctx AgenticContext, _ []string, reason string) AgenticWritable {
	return agenticResp{
		status:      http.StatusPreconditionFailed,
		contentType: ctApprovalRequired,
		body: ApprovalRequiredBody{
			Type:          "approval_required",
			Code:          "APPROVAL_REQUIRED",
			Message:       reason,
			ApprovalToken: newUUID(),
			Trace:         traceFrom(ctx),
		},
	}
}

// Forbidden returns a 403 application/vnd.yaagents.error+json response.
func (AgenticResponse) Forbidden(ctx AgenticContext, message string) AgenticWritable {
	return agenticResp{
		status:      http.StatusForbidden,
		contentType: ctError,
		body: AgenticErrorBody{
			Type:    "forbidden",
			Code:    "FORBIDDEN",
			Message: message,
			Trace:   traceFrom(ctx),
		},
	}
}

// Conflict returns a 409 application/vnd.yaagents.conflict+json response.
func (AgenticResponse) Conflict(ctx AgenticContext, message string) AgenticWritable {
	return agenticResp{
		status:      http.StatusConflict,
		contentType: ctConflict,
		body: ConflictBody{
			Type:    "conflict",
			Code:    "CONFLICT",
			Message: message,
			Trace:   traceFrom(ctx),
		},
	}
}

// FailedDependency returns a 424 application/vnd.yaagents.error+json response.
// dependency names the upstream service; message describes the failure.
func (AgenticResponse) FailedDependency(ctx AgenticContext, dependency, message string) AgenticWritable {
	msg := message
	if dependency != "" {
		msg = fmt.Sprintf("[%s] %s", dependency, message)
	}
	return agenticResp{
		status:      http.StatusFailedDependency,
		contentType: ctError,
		body: AgenticErrorBody{
			Type:    "failed_dependency",
			Code:    "FAILED_DEPENDENCY",
			Message: msg,
			Trace:   traceFrom(ctx),
		},
	}
}
