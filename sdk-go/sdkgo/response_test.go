// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// testCtx is a populated AgenticContext for all tests.
var testCtx = AgenticContext{
	CorrelationID: "corr-test-123",
	RequestID:     "req-test-456",
	ActorTenant:   "tenant-001",
	Principal:     "user@example.com",
}

// profileTable maps each factory call to its expected (status, contentType).
// Covers all 10 PRD §4 normative table rows.
func TestAgenticResponse_StatusAndContentType(t *testing.T) {
	ar := AgenticResponse{}

	cases := []struct {
		name        string
		resp        AgenticWritable
		wantStatus  int
		wantCT      string
	}{
		{
			name:       "Accepted/202/operation",
			resp:       ar.Accepted(testCtx, "op-001"),
			wantStatus: http.StatusAccepted,
			wantCT:     "application/vnd.yaagents.operation+json",
		},
		{
			name:       "Done/200/json",
			resp:       ar.Done(testCtx, map[string]string{"ok": "true"}),
			wantStatus: http.StatusOK,
			wantCT:     "application/json",
		},
		{
			name:       "Created/201/json",
			resp:       ar.Created(testCtx, map[string]string{"id": "abc"}),
			wantStatus: http.StatusCreated,
			wantCT:     "application/json",
		},
		{
			name:       "Failed/500/error",
			resp:       ar.Failed(testCtx, "boom"),
			wantStatus: http.StatusInternalServerError,
			wantCT:     "application/vnd.yaagents.error+json",
		},
		{
			name: "ClarificationRequired/400/clarification",
			resp: ar.ClarificationRequired(testCtx, []RequiredInput{
				{Name: "goal", Location: "body", Type: "string", Required: true, Question: "What is your goal?"},
			}),
			wantStatus: http.StatusBadRequest,
			wantCT:     "application/vnd.yaagents.clarification+json",
		},
		{
			name: "ValidationFailed/422/validation-error",
			resp: ar.ValidationFailed(testCtx, []ValidationError{
				{Field: "name", Message: "required"},
			}),
			wantStatus: http.StatusUnprocessableEntity,
			wantCT:     "application/vnd.yaagents.validation-error+json",
		},
		{
			name:       "ApprovalRequired/412/approval-required",
			resp:       ar.ApprovalRequired(testCtx, []string{"mgr@example.com"}, "high-value operation"),
			wantStatus: http.StatusPreconditionFailed,
			wantCT:     "application/vnd.yaagents.approval-required+json",
		},
		{
			name:       "Forbidden/403/error",
			resp:       ar.Forbidden(testCtx, "access denied"),
			wantStatus: http.StatusForbidden,
			wantCT:     "application/vnd.yaagents.error+json",
		},
		{
			name:       "Conflict/409/conflict",
			resp:       ar.Conflict(testCtx, "already exists"),
			wantStatus: http.StatusConflict,
			wantCT:     "application/vnd.yaagents.conflict+json",
		},
		{
			name:       "FailedDependency/424/error",
			resp:       ar.FailedDependency(testCtx, "crm-api", "timeout"),
			wantStatus: http.StatusFailedDependency,
			wantCT:     "application/vnd.yaagents.error+json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.resp == nil {
				t.Fatal("factory returned nil AgenticWritable")
			}
			if got := tc.resp.Status(); got != tc.wantStatus {
				t.Errorf("Status() = %d, want %d", got, tc.wantStatus)
			}
			if got := tc.resp.ContentType(); got != tc.wantCT {
				t.Errorf("ContentType() = %q, want %q", got, tc.wantCT)
			}
			b, err := tc.resp.Body()
			if err != nil {
				t.Fatalf("Body() error: %v", err)
			}
			if len(b) == 0 {
				t.Error("Body() returned empty bytes")
			}
		})
	}
}

// TestAgenticResponse_TracePopulated verifies that every vendor-typed response
// embeds the Trace from the AgenticContext.
func TestAgenticResponse_TracePopulated(t *testing.T) {
	ar := AgenticResponse{}

	vendorResps := []struct {
		name string
		resp AgenticWritable
	}{
		{"Accepted", ar.Accepted(testCtx, "op-002")},
		{"Failed", ar.Failed(testCtx, "err")},
		{"ClarificationRequired", ar.ClarificationRequired(testCtx, []RequiredInput{
			{Name: "x", Location: "body", Type: "string", Required: true, Question: "q?"},
		})},
		{"ValidationFailed", ar.ValidationFailed(testCtx, []ValidationError{{Field: "f", Message: "m"}})},
		{"ApprovalRequired", ar.ApprovalRequired(testCtx, nil, "reason")},
		{"Forbidden", ar.Forbidden(testCtx, "no")},
		{"Conflict", ar.Conflict(testCtx, "dup")},
		{"FailedDependency", ar.FailedDependency(testCtx, "svc", "down")},
	}

	type traceEnvelope struct {
		Trace Trace `json:"trace"`
	}

	for _, tc := range vendorResps {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.resp.Body()
			if err != nil {
				t.Fatalf("Body() error: %v", err)
			}
			var env traceEnvelope
			if err := json.Unmarshal(b, &env); err != nil {
				t.Fatalf("json.Unmarshal: %v (body: %s)", err, b)
			}
			if env.Trace.CorrelationID != testCtx.CorrelationID {
				t.Errorf("trace.correlationId = %q, want %q", env.Trace.CorrelationID, testCtx.CorrelationID)
			}
			if env.Trace.RequestID != testCtx.RequestID {
				t.Errorf("trace.requestId = %q, want %q", env.Trace.RequestID, testCtx.RequestID)
			}
		})
	}
}

// TestAgenticResponse_Accepted_OperationID checks operationId propagation.
func TestAgenticResponse_Accepted_OperationID(t *testing.T) {
	ar := AgenticResponse{}
	resp := ar.Accepted(testCtx, "op-xyz")

	b, err := resp.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}
	var body OperationAccepted
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.OperationID != "op-xyz" {
		t.Errorf("operationId = %q, want %q", body.OperationID, "op-xyz")
	}
	if body.Type != "operation_accepted" {
		t.Errorf("type = %q, want operation_accepted", body.Type)
	}
}

// TestAgenticResponse_ClarificationRequired_Body checks the canonical §4.1 body shape.
func TestAgenticResponse_ClarificationRequired_Body(t *testing.T) {
	ar := AgenticResponse{}
	inputs := []RequiredInput{
		{
			Name:          "successMetric",
			Location:      "body",
			Type:          "string",
			Required:      true,
			Question:      "Which success metric should be optimized?",
			AllowedValues: []string{"ctr", "cpl", "conversion_rate", "lead_quality"},
		},
	}
	resp := ar.ClarificationRequired(testCtx, inputs)

	b, err := resp.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}
	var body ClarificationRequiredBody
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Type != "clarification_required" {
		t.Errorf("type = %q, want clarification_required", body.Type)
	}
	if body.Code != "CLARIFICATION_REQUIRED" {
		t.Errorf("code = %q, want CLARIFICATION_REQUIRED", body.Code)
	}
	if len(body.RequiredInputs) != 1 {
		t.Fatalf("requiredInputs len = %d, want 1", len(body.RequiredInputs))
	}
	ri := body.RequiredInputs[0]
	if ri.Name != "successMetric" {
		t.Errorf("requiredInputs[0].name = %q, want successMetric", ri.Name)
	}
	if len(ri.AllowedValues) != 4 {
		t.Errorf("allowedValues len = %d, want 4", len(ri.AllowedValues))
	}
}

// TestAgenticResponse_ApprovalRequired_Token verifies the token is non-empty
// and looks like a UUID v4 (format: 8-4-4-4-12 hex chars).
func TestAgenticResponse_ApprovalRequired_Token(t *testing.T) {
	ar := AgenticResponse{}
	resp := ar.ApprovalRequired(testCtx, []string{"approver@example.com"}, "high-value op")

	b, err := resp.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}
	var body ApprovalRequiredBody
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.ApprovalToken == "" {
		t.Error("approvalToken is empty")
	}
	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx (36 chars with dashes)
	parts := strings.Split(body.ApprovalToken, "-")
	if len(parts) != 5 {
		t.Errorf("approvalToken %q does not have 5 UUID segments", body.ApprovalToken)
	}
}

// TestAgenticResponse_FailedDependency_WithDependency checks the dependency prefix.
func TestAgenticResponse_FailedDependency_WithDependency(t *testing.T) {
	ar := AgenticResponse{}
	resp := ar.FailedDependency(testCtx, "crm-api", "connection refused")

	b, err := resp.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}
	var body AgenticErrorBody
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !strings.Contains(body.Message, "crm-api") {
		t.Errorf("message %q does not contain dependency name", body.Message)
	}
	if body.Type != "failed_dependency" {
		t.Errorf("type = %q, want failed_dependency", body.Type)
	}
}

// TestAgenticResponse_FailedDependency_NoDependency checks the no-prefix path.
func TestAgenticResponse_FailedDependency_NoDependency(t *testing.T) {
	ar := AgenticResponse{}
	resp := ar.FailedDependency(testCtx, "", "unknown upstream failure")

	b, err := resp.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}
	var body AgenticErrorBody
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Message != "unknown upstream failure" {
		t.Errorf("message = %q, want %q", body.Message, "unknown upstream failure")
	}
}
