package metrics

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestRegistry_Record_IncrementsCounts(t *testing.T) {
	reg := New()
	reg.Record("r1", 200, 10.0)
	reg.Record("r1", 200, 5.0)
	reg.Record("r1", 422, 3.0)

	reg.mu.Lock()
	cnt200 := reg.counts[labelKey{"r1", "200"}]
	cnt422 := reg.counts[labelKey{"r1", "422"}]
	latency := reg.latencySum["r1"]
	reg.mu.Unlock()

	if cnt200 != 2 {
		t.Errorf("counts[r1,200]: want 2, got %d", cnt200)
	}
	if cnt422 != 1 {
		t.Errorf("counts[r1,422]: want 1, got %d", cnt422)
	}
	if latency != 18.0 {
		t.Errorf("latencySum[r1]: want 18.0, got %f", latency)
	}
}

func TestRegistry_WritePrometheus_ContainsMetricNames(t *testing.T) {
	reg := New()
	reg.Record("orders", 200, 25.0)
	reg.Record("orders", 500, 5.0)

	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	out := buf.String()

	for _, want := range []string{
		"yaagents_gateway_requests_total",
		"yaagents_gateway_request_duration_ms_total",
		`route="orders"`,
		`status="200"`,
		`status="500"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestRegistry_WritePrometheus_HelpAndTypeLines(t *testing.T) {
	reg := New()
	reg.Record("r1", 201, 1.0)

	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	out := buf.String()

	for _, want := range []string{
		"# HELP yaagents_gateway_requests_total",
		"# TYPE yaagents_gateway_requests_total counter",
		"# HELP yaagents_gateway_request_duration_ms_total",
		"# TYPE yaagents_gateway_request_duration_ms_total counter",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing line %q", want)
		}
	}
}

func TestRegistry_WritePrometheus_DeterministicOrder(t *testing.T) {
	reg := New()
	// Insert in reverse order.
	reg.Record("z-route", 200, 1.0)
	reg.Record("a-route", 200, 1.0)

	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	out := buf.String()

	iA := strings.Index(out, `route="a-route"`)
	iZ := strings.Index(out, `route="z-route"`)
	if iA < 0 || iZ < 0 {
		t.Fatalf("both routes should appear in output")
	}
	if iA > iZ {
		t.Error("a-route should appear before z-route (sorted output)")
	}
}

func TestRegistry_WritePrometheus_Empty(t *testing.T) {
	reg := New()
	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	// Should at least have the HELP/TYPE headers.
	out := buf.String()
	if !strings.Contains(out, "# HELP") {
		t.Error("empty registry should still emit HELP lines")
	}
}

func TestRegistry_MultipleRoutes(t *testing.T) {
	reg := New()
	reg.Record("alpha", 200, 5.0)
	reg.Record("beta", 201, 7.5)
	reg.Record("alpha", 400, 1.0)

	var buf bytes.Buffer
	reg.WritePrometheus(&buf)
	out := buf.String()

	for _, want := range []string{`route="alpha"`, `route="beta"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output", want)
		}
	}
}

func TestHandler_ContentType(t *testing.T) {
	reg := New()
	h := reg.Handler()

	// Minimal fake ResponseWriter.
	rec := &headerCapture{header: make(map[string][]string)}
	h.ServeHTTP(rec, nil)

	ct := rec.header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain Content-Type, got %q", ct)
	}
}

// headerCapture is a minimal http.ResponseWriter for testing Handler.
type headerCapture struct {
	header    headerMap
	code      int
	bodyBytes bytes.Buffer
}
type headerMap map[string][]string

func (h headerMap) Get(k string) string {
	if v := h[k]; len(v) > 0 {
		return v[0]
	}
	return ""
}

func (hc *headerCapture) Header() http.Header       { return http.Header(hc.header) }
func (hc *headerCapture) WriteHeader(code int)      { hc.code = code }
func (hc *headerCapture) Write(b []byte) (int, error) { return hc.bodyBytes.Write(b) }
