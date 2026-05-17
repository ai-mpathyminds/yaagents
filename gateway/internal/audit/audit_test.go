package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogger_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.Log(Event{
		Timestamp:     "2026-05-17T00:00:00Z",
		RouteID:       "r1",
		Method:        "POST",
		Path:          "/campaigns/cmp-1/optimizations",
		TenantID:      "tenant-a",
		ActorSubject:  "user-1",
		StatusCode:    200,
		LatencyMS:     12.5,
		CorrelationID: "corr-001",
		RequestID:     "req-001",
	})

	var got Event
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("event is not valid JSON: %v\nraw: %s", err, buf.String())
	}
	if got.RouteID != "r1" {
		t.Errorf("route_id: want r1, got %q", got.RouteID)
	}
	if got.StatusCode != 200 {
		t.Errorf("status_code: want 200, got %d", got.StatusCode)
	}
	if got.LatencyMS != 12.5 {
		t.Errorf("latency_ms: want 12.5, got %f", got.LatencyMS)
	}
}

func TestLogger_OneLinePerEvent(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.Log(Event{RouteID: "r1", StatusCode: 200})
	l.Log(Event{RouteID: "r2", StatusCode: 422})

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), buf.String())
	}
	var e1, e2 Event
	if err := json.Unmarshal([]byte(lines[0]), &e1); err != nil {
		t.Fatalf("line 1 not JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &e2); err != nil {
		t.Fatalf("line 2 not JSON: %v", err)
	}
	if e1.RouteID != "r1" || e2.RouteID != "r2" {
		t.Errorf("route IDs: want r1/r2, got %q/%q", e1.RouteID, e2.RouteID)
	}
}

func TestLogger_OmitsEmptyOptionalFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.Log(Event{RouteID: "r3", StatusCode: 200}) // TenantID and ActorSubject empty

	raw := buf.String()
	if strings.Contains(raw, "tenant_id") {
		t.Errorf("tenant_id should be omitted when empty; got: %s", raw)
	}
	if strings.Contains(raw, "actor_subject") {
		t.Errorf("actor_subject should be omitted when empty; got: %s", raw)
	}
}

func TestOpenSink_Stdout(t *testing.T) {
	w, close, err := OpenSink("stdout")
	if err != nil {
		t.Fatalf("OpenSink stdout: %v", err)
	}
	defer close()
	if w == nil {
		t.Error("expected non-nil writer for stdout")
	}
}

func TestOpenSink_Empty_IsStdout(t *testing.T) {
	w, close, err := OpenSink("")
	if err != nil {
		t.Fatalf("OpenSink empty: %v", err)
	}
	defer close()
	if w == nil {
		t.Error("expected non-nil writer for empty path")
	}
}

func TestOpenSink_BadPath(t *testing.T) {
	_, _, err := OpenSink("/no/such/directory/audit.log")
	if err == nil {
		t.Error("expected error for unwritable path")
	}
}

func TestTimestamp_NonEmpty(t *testing.T) {
	if Timestamp() == "" {
		t.Error("Timestamp() should return non-empty string")
	}
}
