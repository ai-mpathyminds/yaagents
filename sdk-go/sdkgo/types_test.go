// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package sdkgo_test

import (
	"encoding/json"
	"testing"

	"github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

// canonical is the PRD §4.1 canonical body shape for ClarificationRequiredBody.
// Used as the ground-truth for round-trip verification (whitespace-normalised).
var canonical = `{"type":"clarification_required","code":"CLARIFICATION_REQUIRED","message":"Additional information is required.","requiredInputs":[{"name":"successMetric","location":"body","type":"string","required":true,"question":"Which success metric should be optimized?","allowedValues":["ctr","cpl","conversion_rate","lead_quality"]}],"trace":{"correlationId":"corr-123","requestId":"req-456"}}`

// TestClarificationRequiredBody_RoundTrip marshals a ClarificationRequiredBody
// populated with the PRD §4.1 values, verifies all fields survive unmarshal,
// and checks the marshalled JSON matches the canonical body shape (ignoring
// whitespace) per WI-3yaa.SG-2 acceptance criteria.
func TestClarificationRequiredBody_RoundTrip(t *testing.T) {
	original := sdkgo.ClarificationRequiredBody{
		Type:    "clarification_required",
		Code:    "CLARIFICATION_REQUIRED",
		Message: "Additional information is required.",
		RequiredInputs: []sdkgo.RequiredInput{
			{
				Name:          "successMetric",
				Location:      "body",
				Type:          "string",
				Required:      true,
				Question:      "Which success metric should be optimized?",
				AllowedValues: []string{"ctr", "cpl", "conversion_rate", "lead_quality"},
			},
		},
		Trace: sdkgo.Trace{
			CorrelationID: "corr-123",
			RequestID:     "req-456",
		},
	}

	// Marshal.
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify output matches PRD §4.1 canonical body shape (compact JSON comparison).
	if string(data) != canonical {
		t.Errorf("marshalled JSON does not match PRD §4.1 canonical shape\ngot:  %s\nwant: %s", data, canonical)
	}

	// Unmarshal back and verify structural equality.
	var got sdkgo.ClarificationRequiredBody
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Type != original.Type {
		t.Errorf("Type: got %q, want %q", got.Type, original.Type)
	}
	if got.Code != original.Code {
		t.Errorf("Code: got %q, want %q", got.Code, original.Code)
	}
	if got.Message != original.Message {
		t.Errorf("Message: got %q, want %q", got.Message, original.Message)
	}
	if len(got.RequiredInputs) != 1 {
		t.Fatalf("RequiredInputs length: got %d, want 1", len(got.RequiredInputs))
	}
	ri := got.RequiredInputs[0]
	if ri.Name != "successMetric" {
		t.Errorf("RequiredInputs[0].Name: got %q, want %q", ri.Name, "successMetric")
	}
	if ri.Location != "body" {
		t.Errorf("RequiredInputs[0].Location: got %q, want %q", ri.Location, "body")
	}
	if ri.Type != "string" {
		t.Errorf("RequiredInputs[0].Type: got %q, want %q", ri.Type, "string")
	}
	if !ri.Required {
		t.Error("RequiredInputs[0].Required: want true, got false")
	}
	if len(ri.AllowedValues) != 4 {
		t.Errorf("RequiredInputs[0].AllowedValues length: got %d, want 4", len(ri.AllowedValues))
	}
	if got.Trace.CorrelationID != "corr-123" {
		t.Errorf("Trace.CorrelationID: got %q, want %q", got.Trace.CorrelationID, "corr-123")
	}
	if got.Trace.RequestID != "req-456" {
		t.Errorf("Trace.RequestID: got %q, want %q", got.Trace.RequestID, "req-456")
	}
}

// TestAllStructsPresent verifies all 9 exported types are present and usable
// (zero-value construction + marshal/unmarshal; structural smoke test).
func TestAllStructsPresent(t *testing.T) {
	types := []interface{}{
		&sdkgo.Trace{},
		&sdkgo.RequiredInput{},
		&sdkgo.ValidationError{},
		&sdkgo.OperationAccepted{},
		&sdkgo.ClarificationRequiredBody{},
		&sdkgo.ValidationFailedBody{},
		&sdkgo.ApprovalRequiredBody{},
		&sdkgo.ConflictBody{},
		&sdkgo.AgenticErrorBody{},
	}
	for _, v := range types {
		data, err := json.Marshal(v)
		if err != nil {
			t.Errorf("Marshal %T: %v", v, err)
			continue
		}
		if err := json.Unmarshal(data, v); err != nil {
			t.Errorf("Unmarshal %T: %v", v, err)
		}
	}
}
