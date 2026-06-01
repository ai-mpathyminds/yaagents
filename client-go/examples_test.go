// Verifiable idiomatic-usage examples for the yaagents Go client.
// This file proves the PRD §5.9 usage pattern compiles and runs correctly.
package yaagentsclient_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	yaagentsclient "github.com/ai-mpathyminds/yaagents/client-go"
)

// TestProfileVersion verifies the client declares the expected profile version.
func TestProfileVersion(t *testing.T) {
	if yaagentsclient.ProfileVersion != "v0.2" {
		t.Errorf("ProfileVersion = %q; want v0.2", yaagentsclient.ProfileVersion)
	}
}

// TestExample_IdiomaticUsage exercises the exact usage pattern from PRD §5.9.
// It verifies that the documented API surface compiles and produces the correct
// behaviour against a live test server — serving as the "doc tests" gate.
func TestExample_IdiomaticUsage(t *testing.T) {
	// Spin up a server that returns a created (201) response — the happy-path
	// result for Optimizations.Create shown in PRD §5.9.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"opt-001","campaignId":"cmp-123","status":"queued"}`)
	}))
	defer srv.Close()

	ctx := context.Background()

	// ── PRD §5.9 idiomatic usage block (verbatim pattern) ────────────────
	client := yaagentsclient.New(
		srv.URL,
		yaagentsclient.WithToken("my-jwt"),
		yaagentsclient.WithTenantID("tenant-001"),
	)

	result, err := client.Campaigns().ByID("cmp-123").Optimizations().Create(ctx, map[string]any{
		"goal": "reduce_cost_per_lead",
	})
	if err != nil {
		var clarify *yaagentsclient.ClarificationRequired
		if errors.As(err, &clarify) {
			for _, input := range clarify.RequiredInputs {
				t.Logf("Required: %s — %s", input.Name, input.Question)
			}
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the response was correctly parsed from the test server.
	if result.Type != "created" {
		t.Errorf("result.Type = %q; want created", result.Type)
	}
	if len(result.Resource) == 0 {
		t.Error("result.Resource should be populated")
	}
	t.Logf("Created: %s", result.Resource)
}

// TestExample_ErrorStyle exercises the error-style caller pattern.
func TestExample_ErrorStyle(t *testing.T) {
	// Server returns clarification_required — the common agentic 400.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.yaagents.clarification+json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{
			"type":"clarification_required","code":"CLARIFICATION_REQUIRED",
			"message":"Need metric.",
			"requiredInputs":[{"name":"successMetric","type":"string","question":"Which metric?"}],
			"trace":{"correlationId":"c1","requestId":"r1"}}`)
	}))
	defer srv.Close()

	client := yaagentsclient.New(srv.URL, yaagentsclient.WithToken("tok"))
	_, err := client.Campaigns().ByID("c1").Optimizations().Create(context.Background(), map[string]any{})

	var clarify *yaagentsclient.ClarificationRequired
	if !errors.As(err, &clarify) {
		t.Fatalf("expected *ClarificationRequired; got %T: %v", err, err)
	}
	if len(clarify.RequiredInputs) == 0 {
		t.Error("RequiredInputs must not be empty")
	}
	if clarify.RequiredInputs[0].Name != "successMetric" {
		t.Errorf("RequiredInputs[0].Name = %q; want successMetric", clarify.RequiredInputs[0].Name)
	}
}
