// Internal package tests for parseResponse + AgenticResult (GOC-3).
package yaagentsclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// responseSrv builds a test server that responds with status, Content-Type ct,
// and body. It returns a *Client wired to that server.
func responseSrv(t *testing.T, status int, ct, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(status)
		if body != "" {
			fmt.Fprint(w, body)
		}
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL)
}

// callDo issues a bare GET /x through the client's do() and returns the result.
func callDo(t *testing.T, c *Client) (*AgenticResult, error) {
	t.Helper()
	return c.do(context.Background(), http.MethodGet, "/x", nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// 10-row normative table (spec §4)
// ─────────────────────────────────────────────────────────────────────────────

func TestParse_Success_200(t *testing.T) {
	c := responseSrv(t, 200, "application/json", `{"id":"c1"}`)
	r, err := callDo(t, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "success" {
		t.Errorf("Type = %q; want success", r.Type)
	}
	if r.Status != 200 {
		t.Errorf("Status = %d; want 200", r.Status)
	}
	if r.Err() != nil {
		t.Error("Err() should be nil for success")
	}
	if len(r.Resource) == 0 {
		t.Error("Resource should be populated")
	}
}

func TestParse_Created_201(t *testing.T) {
	c := responseSrv(t, 201, "application/json", `{"id":"new-c1"}`)
	r, err := callDo(t, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "created" {
		t.Errorf("Type = %q; want created", r.Type)
	}
	if r.Status != 201 {
		t.Errorf("Status = %d; want 201", r.Status)
	}
	if r.Err() != nil {
		t.Error("Err() should be nil for created")
	}
}

func TestParse_Accepted_202(t *testing.T) {
	body := `{
		"type":"operation_accepted","code":"OPERATION_ACCEPTED",
		"message":"Processing.","operationId":"op-001",
		"trace":{"correlationId":"corr-501","requestId":"req-501"}}`
	c := responseSrv(t, 202, ctOperation, body)
	r, err := callDo(t, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != "accepted" {
		t.Errorf("Type = %q; want accepted", r.Type)
	}
	if r.Status != 202 {
		t.Errorf("Status = %d; want 202", r.Status)
	}
	if r.OperationID != "op-001" {
		t.Errorf("OperationID = %q; want op-001", r.OperationID)
	}
	if r.Err() != nil {
		t.Error("Err() should be nil for accepted (202)")
	}
}

func TestParse_ClarificationRequired_400(t *testing.T) {
	body := `{
		"type":"clarification_required","code":"CLARIFICATION_REQUIRED",
		"message":"Need more info.",
		"requiredInputs":[{"name":"successMetric","type":"string","question":"Which metric?"}],
		"trace":{"correlationId":"corr-001","requestId":"req-001"}}`
	c := responseSrv(t, 400, ctClarification, body)
	r, err := callDo(t, c)

	if r.Type != "clarification_required" {
		t.Errorf("Type = %q; want clarification_required", r.Type)
	}
	if r.Status != 400 {
		t.Errorf("Status = %d; want 400", r.Status)
	}
	if len(r.RequiredInputs) != 1 || r.RequiredInputs[0].Name != "successMetric" {
		t.Errorf("RequiredInputs = %v; want [{Name:successMetric}]", r.RequiredInputs)
	}

	var clar *ClarificationRequired
	if !errors.As(err, &clar) {
		t.Fatalf("errors.As(*ClarificationRequired) failed; err = %v", err)
	}
	if len(clar.RequiredInputs) != 1 {
		t.Errorf("typed error RequiredInputs len = %d; want 1", len(clar.RequiredInputs))
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 400")
	}
}

func TestParse_ValidationFailed_422(t *testing.T) {
	body := `{
		"type":"validation_failed","code":"VALIDATION_FAILED",
		"message":"Invalid inputs.",
		"errors":[{"field":"budget","message":"Must be > 0"}],
		"trace":{"correlationId":"corr-101","requestId":"req-101"}}`
	c := responseSrv(t, 422, ctValidationError, body)
	r, err := callDo(t, c)

	if r.Type != "validation_failed" {
		t.Errorf("Type = %q; want validation_failed", r.Type)
	}
	if r.Status != 422 {
		t.Errorf("Status = %d; want 422", r.Status)
	}

	var vf *ValidationFailed
	if !errors.As(err, &vf) {
		t.Fatalf("errors.As(*ValidationFailed) failed; err = %v", err)
	}
	if len(vf.Errors) != 1 || vf.Errors[0].Field != "budget" {
		t.Errorf("vf.Errors = %v; want [{Field:budget}]", vf.Errors)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 422")
	}
}

func TestParse_ApprovalRequired_412(t *testing.T) {
	body := `{
		"type":"approval_required","code":"APPROVAL_REQUIRED",
		"message":"Human approval needed.","approvalToken":"tok-abc",
		"trace":{"correlationId":"corr-201","requestId":"req-201"}}`
	c := responseSrv(t, 412, ctApprovalRequired, body)
	r, err := callDo(t, c)

	if r.Type != "approval_required" {
		t.Errorf("Type = %q; want approval_required", r.Type)
	}
	if r.Status != 412 {
		t.Errorf("Status = %d; want 412", r.Status)
	}

	var ar *ApprovalRequired
	if !errors.As(err, &ar) {
		t.Fatalf("errors.As(*ApprovalRequired) failed; err = %v", err)
	}
	if ar.ApprovalToken != "tok-abc" {
		t.Errorf("ApprovalToken = %q; want tok-abc", ar.ApprovalToken)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 412")
	}
}

func TestParse_Forbidden_403(t *testing.T) {
	body := `{
		"type":"forbidden","code":"PERMISSION_DENIED",
		"message":"No permission.",
		"trace":{"correlationId":"corr-401","requestId":"req-401"}}`
	c := responseSrv(t, 403, ctError, body)
	r, err := callDo(t, c)

	if r.Type != "forbidden" {
		t.Errorf("Type = %q; want forbidden", r.Type)
	}
	if r.Status != 403 {
		t.Errorf("Status = %d; want 403", r.Status)
	}

	var af *AgenticForbidden
	if !errors.As(err, &af) {
		t.Fatalf("errors.As(*AgenticForbidden) failed; err = %v", err)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 403")
	}
}

func TestParse_Conflict_409(t *testing.T) {
	body := `{
		"type":"conflict","code":"CAMPAIGN_LOCKED",
		"message":"Campaign is locked.","conflictingResourceId":"camp-001",
		"trace":{"correlationId":"corr-301","requestId":"req-301"}}`
	c := responseSrv(t, 409, ctConflict, body)
	r, err := callDo(t, c)

	if r.Type != "conflict" {
		t.Errorf("Type = %q; want conflict", r.Type)
	}
	if r.Status != 409 {
		t.Errorf("Status = %d; want 409", r.Status)
	}

	var co *Conflict
	if !errors.As(err, &co) {
		t.Fatalf("errors.As(*Conflict) failed; err = %v", err)
	}
	if co.ConflictingResourceID != "camp-001" {
		t.Errorf("ConflictingResourceID = %q; want camp-001", co.ConflictingResourceID)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 409")
	}
}

func TestParse_FailedDependency_424(t *testing.T) {
	body := `{
		"type":"failed_dependency","code":"UPSTREAM_UNAVAILABLE",
		"message":"AI model unavailable.",
		"trace":{"correlationId":"corr-402","requestId":"req-402"}}`
	c := responseSrv(t, 424, ctError, body)
	r, err := callDo(t, c)

	if r.Type != "failed_dependency" {
		t.Errorf("Type = %q; want failed_dependency", r.Type)
	}
	if r.Status != 424 {
		t.Errorf("Status = %d; want 424", r.Status)
	}

	var fd *FailedDependency
	if !errors.As(err, &fd) {
		t.Fatalf("errors.As(*FailedDependency) failed; err = %v", err)
	}
	if fd.Dependency != "UPSTREAM_UNAVAILABLE" {
		t.Errorf("Dependency = %q; want UPSTREAM_UNAVAILABLE", fd.Dependency)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 424")
	}
}

func TestParse_Error_500(t *testing.T) {
	body := `{
		"type":"error","code":"INTERNAL_ERROR",
		"message":"Unexpected error.",
		"trace":{"correlationId":"corr-403","requestId":"req-403"}}`
	c := responseSrv(t, 500, ctError, body)
	r, err := callDo(t, c)

	if r.Type != "error" {
		t.Errorf("Type = %q; want error", r.Type)
	}
	if r.Status != 500 {
		t.Errorf("Status = %d; want 500", r.Status)
	}

	var ae *AgenticError
	if !errors.As(err, &ae) {
		t.Fatalf("errors.As(*AgenticError) failed; err = %v", err)
	}
	if ae.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q; want INTERNAL_ERROR", ae.Code)
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 500")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DESERIALIZE_ERROR — truncated JSON bodies
// ─────────────────────────────────────────────────────────────────────────────

func TestDeserializeError_TruncatedBody_Operation(t *testing.T) {
	c := responseSrv(t, 202, ctOperation, `{"truncated`)
	r, err := callDo(t, c)

	if r == nil {
		t.Fatal("result must never be nil")
	}

	var ae *AgenticError
	if !errors.As(err, &ae) {
		t.Fatalf("errors.As(*AgenticError) failed; err = %v", err)
	}
	if ae.Code != "DESERIALIZE_ERROR" {
		t.Errorf("Code = %q; want DESERIALIZE_ERROR", ae.Code)
	}
}

func TestDeserializeError_TruncatedBody_Clarification(t *testing.T) {
	c := responseSrv(t, 400, ctClarification, `{bad json`)
	r, err := callDo(t, c)

	if r == nil {
		t.Fatal("result must never be nil")
	}
	var ae *AgenticError
	if !errors.As(err, &ae) || ae.Code != "DESERIALIZE_ERROR" {
		t.Errorf("want DESERIALIZE_ERROR; got err = %v", err)
	}
}

func TestDeserializeError_TruncatedBody_Error(t *testing.T) {
	c := responseSrv(t, 500, ctError, `{{{{`)
	r, err := callDo(t, c)

	if r == nil {
		t.Fatal("result must never be nil")
	}
	var ae *AgenticError
	if !errors.As(err, &ae) || ae.Code != "DESERIALIZE_ERROR" {
		t.Errorf("want DESERIALIZE_ERROR; got err = %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Trace propagation
// ─────────────────────────────────────────────────────────────────────────────

func TestParse_TracePopulated(t *testing.T) {
	body := `{
		"type":"forbidden","code":"X","message":"m",
		"trace":{"correlationId":"cid-99","requestId":"rid-77"}}`
	c := responseSrv(t, 403, ctError, body)
	r, _ := callDo(t, c)

	if r.Trace.CorrelationID != "cid-99" {
		t.Errorf("Trace.CorrelationID = %q; want cid-99", r.Trace.CorrelationID)
	}
	if r.Trace.RequestID != "rid-77" {
		t.Errorf("Trace.RequestID = %q; want rid-77", r.Trace.RequestID)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Error() method coverage — all typed error types must produce a non-empty
// message (acceptance criterion: "useful Error() message").
// ─────────────────────────────────────────────────────────────────────────────

func TestErrorMethods_NonEmpty(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"ClarificationRequired", &ClarificationRequired{
			RequiredInputs: []RequiredInput{{Name: "field1"}},
		}},
		{"ClarificationRequired/empty", &ClarificationRequired{}},
		{"ValidationFailed", &ValidationFailed{
			Errors: []ValidationError{{Field: "budget", Message: "bad"}},
		}},
		{"ValidationFailed/empty", &ValidationFailed{}},
		{"AgenticForbidden", &AgenticForbidden{Message: "no access"}},
		{"FailedDependency", &FailedDependency{Dependency: "upstream", Message: "down"}},
		{"Conflict/with-resource", &Conflict{
			Code: "LOCKED", Message: "locked", ConflictingResourceID: "res-1",
		}},
		{"Conflict/no-resource", &Conflict{Code: "LOCKED", Message: "locked"}},
		{"ApprovalRequired", &ApprovalRequired{Code: "APPROVAL_REQUIRED", Message: "need approval"}},
		{"AgenticError", &AgenticError{Code: "INTERNAL_ERROR", Message: "oops"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.err.Error()
			if msg == "" {
				t.Errorf("%T.Error() returned empty string", tc.err)
			}
		})
	}
}

// TestErrNilReceiver confirms Err() on a nil *AgenticResult is safe.
func TestErrNilReceiver(t *testing.T) {
	var r *AgenticResult
	if r.Err() != nil {
		t.Error("Err() on nil receiver should return nil")
	}
}

// TestParse_UnknownCT_4xx covers the default-branch non-2xx path.
func TestParse_UnknownCT_4xx(t *testing.T) {
	c := responseSrv(t, 400, "text/plain", "bad request")
	r, err := callDo(t, c)
	if r == nil {
		t.Fatal("result must not be nil")
	}
	if r.Type != "error" {
		t.Errorf("Type = %q; want error", r.Type)
	}
	var ae *AgenticError
	if !errors.As(err, &ae) || ae.Code != "UNKNOWN_CONTENT_TYPE" {
		t.Errorf("want UNKNOWN_CONTENT_TYPE AgenticError; got %v", err)
	}
}

// TestParse_AppJSON_NonSuccess covers application/json with a non-2xx status.
func TestParse_AppJSON_NonSuccess(t *testing.T) {
	c := responseSrv(t, 400, "application/json", `{"error":"bad"}`)
	r, err := callDo(t, c)
	if r == nil {
		t.Fatal("result must not be nil")
	}
	if r.Type != "error" {
		t.Errorf("Type = %q; want error", r.Type)
	}
	var ae *AgenticError
	if !errors.As(err, &ae) || ae.Code != "UNEXPECTED_STATUS" {
		t.Errorf("want UNEXPECTED_STATUS AgenticError; got %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// mediaType helper
// ─────────────────────────────────────────────────────────────────────────────

func TestMediaType_StripParams(t *testing.T) {
	cases := []struct{ in, want string }{
		{"application/json", "application/json"},
		{"application/json; charset=utf-8", "application/json"},
		{"application/vnd.yaagents.error+json; charset=utf-8", "application/vnd.yaagents.error+json"},
		{"", ""},
		{"  text/plain  ", "text/plain"},
	}
	for _, tc := range cases {
		if got := mediaType(tc.in); got != tc.want {
			t.Errorf("mediaType(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
