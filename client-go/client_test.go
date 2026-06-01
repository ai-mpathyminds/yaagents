// Internal package tests — access to unexported do() and newUUID().
package yaagentsclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

// uuidV4Re validates RFC 4122 §4.4 UUID v4 shape: version nibble = 4,
// variant nibble ∈ {8,9,a,b}.
var uuidV4Re = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

// captureHandler records request headers and responds 200 OK.
func captureHandler(out *http.Header) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*out = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})
}

// TestHeaders_AllFourPresent asserts that Authorization, X-Tenant-ID,
// X-Correlation-ID (auto UUID v4), and Content-Type are all injected when
// token + tenantID are set and a non-nil body is supplied.
func TestHeaders_AllFourPresent(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(captureHandler(&got))
	defer srv.Close()

	c := New(srv.URL,
		WithToken("t"),
		WithTenantID("tn"),
	)
	if _, err := c.do(context.Background(), http.MethodPost, "/test", strings.NewReader(`{}`)); err != nil {
		t.Fatalf("do: %v", err)
	}

	if v := got.Get("Authorization"); v != "Bearer t" {
		t.Errorf("Authorization = %q; want \"Bearer t\"", v)
	}
	if v := got.Get("X-Tenant-ID"); v != "tn" {
		t.Errorf("X-Tenant-ID = %q; want \"tn\"", v)
	}
	corrID := got.Get("X-Correlation-ID")
	if corrID == "" {
		t.Fatal("X-Correlation-ID not set")
	}
	if !uuidV4Re.MatchString(corrID) {
		t.Errorf("X-Correlation-ID = %q; not a UUID v4", corrID)
	}
	if v := got.Get("Content-Type"); v != "application/json" {
		t.Errorf("Content-Type = %q; want \"application/json\"", v)
	}
}

// TestWithCorrelationID_Override verifies that a static WithCorrelationID
// replaces the auto UUID and is forwarded verbatim.
func TestWithCorrelationID_Override(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(captureHandler(&got))
	defer srv.Close()

	c := New(srv.URL, WithCorrelationID("custom-corr"))
	if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get("X-Correlation-ID"); v != "custom-corr" {
		t.Errorf("X-Correlation-ID = %q; want \"custom-corr\"", v)
	}
}

// recTransport wraps a RoundTripper and records that it was called.
type recTransport struct {
	next   http.RoundTripper
	called bool
}

func (rt *recTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.called = true
	return rt.next.RoundTrip(r)
}

// TestWithHTTPClient_ReplacesTransport verifies that WithHTTPClient's
// transport is the one actually used for every request.
func TestWithHTTPClient_ReplacesTransport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rt := &recTransport{next: http.DefaultTransport}
	custom := &http.Client{Timeout: 5 * time.Second, Transport: rt}

	c := New(srv.URL, WithHTTPClient(custom))
	if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if !rt.called {
		t.Error("custom transport was not invoked")
	}
}

// TestNoAuthorizationHeader_WhenTokenEmpty confirms that Authorization is
// absent when no WithToken option is supplied.
func TestNoAuthorizationHeader_WhenTokenEmpty(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(captureHandler(&got))
	defer srv.Close()

	c := New(srv.URL)
	if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get("Authorization"); v != "" {
		t.Errorf("Authorization = %q; want empty", v)
	}
}

// TestNoXTenantID_WhenTenantEmpty confirms that X-Tenant-ID is absent when no
// WithTenantID option is supplied.
func TestNoXTenantID_WhenTenantEmpty(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(captureHandler(&got))
	defer srv.Close()

	c := New(srv.URL)
	if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get("X-Tenant-ID"); v != "" {
		t.Errorf("X-Tenant-ID = %q; want empty", v)
	}
}

// TestNoContentType_WhenBodyNil confirms Content-Type is absent for nil body.
func TestNoContentType_WhenBodyNil(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(captureHandler(&got))
	defer srv.Close()

	c := New(srv.URL)
	if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
		t.Fatalf("do: %v", err)
	}
	if v := got.Get("Content-Type"); v != "" {
		t.Errorf("Content-Type = %q; want empty for nil body", v)
	}
}

// TestNewUUID_IsValidV4 runs 200 iterations to confirm every output matches
// the UUID v4 shape and that they are not all identical (randomness smoke).
func TestNewUUID_IsValidV4(t *testing.T) {
	seen := make(map[string]struct{}, 200)
	for i := 0; i < 200; i++ {
		u := newUUID()
		if !uuidV4Re.MatchString(u) {
			t.Errorf("newUUID() = %q; not a UUID v4", u)
		}
		seen[u] = struct{}{}
	}
	if len(seen) < 190 {
		t.Errorf("only %d distinct UUIDs in 200 iterations; expected near-200", len(seen))
	}
}

// TestAutoCorrelationID_UniquePerRequest confirms that two consecutive requests
// without WithCorrelationID carry distinct X-Correlation-ID values.
func TestAutoCorrelationID_UniquePerRequest(t *testing.T) {
	ids := make([]string, 0, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids = append(ids, r.Header.Get("X-Correlation-ID"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(srv.URL)
	for i := 0; i < 2; i++ {
		if _, err := c.do(context.Background(), http.MethodGet, "/test", nil); err != nil {
			t.Fatalf("do[%d]: %v", i, err)
		}
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 captured IDs; got %d", len(ids))
	}
	if ids[0] == ids[1] {
		t.Errorf("successive requests share correlation ID %q; expected unique", ids[0])
	}
}
