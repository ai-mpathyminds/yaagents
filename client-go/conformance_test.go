// Conformance tests — wire the shared golden corpus (tests/golden/) against
// the Go client and assert that each canonical fixture parses to the correct
// AgenticResult.Type + Status + typed-error result.
//
// One test per row of the 10-row Agentic REST Response Profile
// (spec/agentic-rest-profile.md §4 normative table).
package yaagentsclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	yaagentsclient "github.com/ai-mpathyminds/yaagents/client-go"
)

// goldenFixture is the JSON envelope stored in tests/golden/*.json.
type goldenFixture struct {
	Status       int             `json:"status"`
	ContentType  string          `json:"contentType"`
	ExpectedType string          `json:"expectedType"`
	Body         json.RawMessage `json:"body"`
}

// loadGolden reads and parses a named fixture from tests/golden/.
// Tests run with cwd = the package directory (client-go/), so the corpus is
// one level up at ../tests/golden/.
func loadGolden(t *testing.T, name string) goldenFixture {
	t.Helper()
	path := filepath.Join("..", "tests", "golden", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loadGolden(%q): %v", name, err)
	}
	var f goldenFixture
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("loadGolden(%q) parse: %v", name, err)
	}
	return f
}

// fixtureServer returns a test server that replays a golden fixture and a
// client wired to it.
func fixtureServer(t *testing.T, f goldenFixture) *yaagentsclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.ContentType != "" {
			w.Header().Set("Content-Type", f.ContentType)
		}
		w.WriteHeader(f.Status)
		fmt.Fprint(w, string(f.Body))
	}))
	t.Cleanup(srv.Close)
	return yaagentsclient.New(srv.URL)
}

// makeRequest uses the exported resource-accessor chain to exercise the full
// do() + parseResponse() path. The server returns whatever fixture is
// configured, independent of the request path.
func makeRequest(t *testing.T, c *yaagentsclient.Client) (*yaagentsclient.AgenticResult, error) {
	t.Helper()
	return c.Campaigns().ByID("c1").Optimizations().Get(context.Background(), "o1")
}

// ─────────────────────────────────────────────────────────────────────────────
// 10-row golden corpus tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGolden_Success(t *testing.T) {
	f := loadGolden(t, "success")
	r, err := makeRequest(t, fixtureServer(t, f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	if len(r.Resource) == 0 {
		t.Error("Resource should be populated for success")
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v; want nil for success", r.Err())
	}
}

func TestGolden_Created(t *testing.T) {
	f := loadGolden(t, "created")
	r, err := makeRequest(t, fixtureServer(t, f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v; want nil for created", r.Err())
	}
}

func TestGolden_Accepted(t *testing.T) {
	f := loadGolden(t, "accepted")
	r, err := makeRequest(t, fixtureServer(t, f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	if r.OperationID == "" {
		t.Error("OperationID should be populated for accepted")
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v; want nil for accepted (202)", r.Err())
	}
}

func TestGolden_ClarificationRequired(t *testing.T) {
	f := loadGolden(t, "clarification-required")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.ClarificationRequired
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*ClarificationRequired) failed; err = %v", err)
	}
	if len(typed.RequiredInputs) == 0 {
		t.Error("RequiredInputs must not be empty")
	}
	if r.Err() == nil {
		t.Error("Err() must be non-nil for 400")
	}
}

func TestGolden_ValidationFailed(t *testing.T) {
	f := loadGolden(t, "validation-failed")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.ValidationFailed
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*ValidationFailed) failed; err = %v", err)
	}
	if len(typed.Errors) == 0 {
		t.Error("ValidationFailed.Errors must not be empty")
	}
}

func TestGolden_ApprovalRequired(t *testing.T) {
	f := loadGolden(t, "approval-required")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.ApprovalRequired
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*ApprovalRequired) failed; err = %v", err)
	}
	if typed.ApprovalToken == "" {
		t.Error("ApprovalToken must not be empty")
	}
}

func TestGolden_Forbidden(t *testing.T) {
	f := loadGolden(t, "forbidden")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.AgenticForbidden
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*AgenticForbidden) failed; err = %v", err)
	}
}

func TestGolden_Conflict(t *testing.T) {
	f := loadGolden(t, "conflict")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.Conflict
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*Conflict) failed; err = %v", err)
	}
	if typed.ConflictingResourceID == "" {
		t.Error("ConflictingResourceID must not be empty")
	}
}

func TestGolden_FailedDependency(t *testing.T) {
	f := loadGolden(t, "failed-dependency")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.FailedDependency
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*FailedDependency) failed; err = %v", err)
	}
}

func TestGolden_Error(t *testing.T) {
	f := loadGolden(t, "error")
	r, err := makeRequest(t, fixtureServer(t, f))
	if r.Type != f.ExpectedType {
		t.Errorf("Type = %q; want %q", r.Type, f.ExpectedType)
	}
	if r.Status != f.Status {
		t.Errorf("Status = %d; want %d", r.Status, f.Status)
	}
	var typed *yaagentsclient.AgenticError
	if !errors.As(err, &typed) {
		t.Fatalf("errors.As(*AgenticError) failed; err = %v", err)
	}
	if typed.Code != "INTERNAL_ERROR" {
		t.Errorf("Code = %q; want INTERNAL_ERROR", typed.Code)
	}
}
