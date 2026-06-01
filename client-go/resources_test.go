// Internal package tests for resource accessor chain (GOC-2).
package yaagentsclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// captureRequest records method, path, and body; responds with statusCode.
type capturedReq struct {
	Method string
	Path   string
	Body   []byte
}

func captureSrv(t *testing.T, statusCode int) (*httptest.Server, *capturedReq) {
	t.Helper()
	var cap capturedReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.Method = r.Method
		cap.Path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		cap.Body = b
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv, &cap
}

// ── Optimizations.Create ───────────────────────────────────────────────────

func TestOptimizationsCreate_POST_URL_Body(t *testing.T) {
	srv, cap := captureSrv(t, http.StatusCreated)

	c := New(srv.URL)
	payload := map[string]any{"goal": "reduce_cost_per_lead"}
	result, err := c.Campaigns().ByID("c1").Optimizations().Create(context.Background(), payload)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if cap.Method != http.MethodPost {
		t.Errorf("method = %q; want POST", cap.Method)
	}
	if cap.Path != "/campaigns/c1/optimizations" {
		t.Errorf("path = %q; want /campaigns/c1/optimizations", cap.Path)
	}

	// Body must be valid JSON containing the key we sent.
	var parsed map[string]any
	if err := json.Unmarshal(cap.Body, &parsed); err != nil {
		t.Fatalf("body not valid JSON: %v", err)
	}
	if parsed["goal"] != "reduce_cost_per_lead" {
		t.Errorf("body[goal] = %v; want reduce_cost_per_lead", parsed["goal"])
	}

	if result.Status != http.StatusCreated {
		t.Errorf("result.Status = %d; want 201", result.Status)
	}
}

// ── Optimizations.Get ─────────────────────────────────────────────────────

func TestOptimizationsGet_GET_URL_NoBody(t *testing.T) {
	srv, cap := captureSrv(t, http.StatusOK)

	c := New(srv.URL)
	result, err := c.Campaigns().ByID("c1").Optimizations().Get(context.Background(), "o1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if cap.Method != http.MethodGet {
		t.Errorf("method = %q; want GET", cap.Method)
	}
	if cap.Path != "/campaigns/c1/optimizations/o1" {
		t.Errorf("path = %q; want /campaigns/c1/optimizations/o1", cap.Path)
	}
	if len(cap.Body) != 0 {
		t.Errorf("expected empty body for GET; got %q", cap.Body)
	}
	if result.Status != http.StatusOK {
		t.Errorf("result.Status = %d; want 200", result.Status)
	}
}

// ── Assets.Generate ───────────────────────────────────────────────────────

func TestAssetsGenerate_POST_URL_Body(t *testing.T) {
	srv, cap := captureSrv(t, http.StatusCreated)

	c := New(srv.URL)
	payload := map[string]any{"type": "banner", "size": "1080x1080"}
	result, err := c.Campaigns().ByID("c1").Assets().Generate(context.Background(), payload)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if cap.Method != http.MethodPost {
		t.Errorf("method = %q; want POST", cap.Method)
	}
	if cap.Path != "/campaigns/c1/assets:generate" {
		t.Errorf("path = %q; want /campaigns/c1/assets:generate", cap.Path)
	}
	var parsed map[string]any
	if err := json.Unmarshal(cap.Body, &parsed); err != nil {
		t.Fatalf("body not valid JSON: %v", err)
	}
	if result.Status != http.StatusCreated {
		t.Errorf("result.Status = %d; want 201", result.Status)
	}
}

// ── Campaign ID interpolation ─────────────────────────────────────────────

// TestByID_InterpolatesID checks that a different campaign ID produces the
// correct path for both sub-resources.
func TestByID_InterpolatesID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantPath string
		run      func(c *Client, id string) error
	}{
		{
			name:     "Optimizations.Get",
			id:       "camp-42",
			wantPath: "/campaigns/camp-42/optimizations/op-99",
			run: func(c *Client, id string) error {
				_, err := c.Campaigns().ByID(id).Optimizations().Get(context.Background(), "op-99")
				return err
			},
		},
		{
			name:     "Assets.Generate",
			id:       "camp-42",
			wantPath: "/campaigns/camp-42/assets:generate",
			run: func(c *Client, id string) error {
				_, err := c.Campaigns().ByID(id).Assets().Generate(context.Background(), struct{}{})
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, cap := captureSrv(t, http.StatusOK)
			c := New(srv.URL)
			if err := tc.run(c, tc.id); err != nil {
				t.Fatalf("run: %v", err)
			}
			if cap.Path != tc.wantPath {
				t.Errorf("path = %q; want %q", cap.Path, tc.wantPath)
			}
		})
	}
}

// ── Context cancellation ─────────────────────────────────────────────────

// TestContextCancel_PropagatesDisconnect cancels the context while the server
// is blocked waiting; verifies the client unblocks and returns a non-nil error.
func TestContextCancel_PropagatesDisconnect(t *testing.T) {
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
			// Client disconnected — normal path for this test.
		case <-time.After(5 * time.Second):
			t.Error("server timeout: client did not disconnect")
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := New(srv.URL)

	done := make(chan error, 1)
	go func() {
		_, err := c.Campaigns().ByID("c1").Optimizations().Get(ctx, "o1")
		done <- err
	}()

	// Wait for server to receive the request, then cancel.
	<-started
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected non-nil error after context cancel; got nil")
		}
		if !errors.Is(err, context.Canceled) {
			// Wrapped errors are acceptable; just log what we got.
			t.Logf("context cancel error (expected): %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("request did not return within 3 s after context cancel")
	}
}

// ── marshalBody error path ────────────────────────────────────────────────

// TestCreate_MarshalError confirms that an unmarshalable body returns an
// error without panicking or attempting an HTTP call.
func TestCreate_MarshalError(t *testing.T) {
	c := New("http://127.0.0.1:0") // port 0 — won't accept connections
	// channels are not JSON-serialisable.
	_, err := c.Campaigns().ByID("c1").Optimizations().Create(
		context.Background(), make(chan int),
	)
	if err == nil {
		t.Error("expected error for unmarshalable body; got nil")
	}
}

// TestGenerate_MarshalError mirrors TestCreate_MarshalError for Assets.Generate.
func TestGenerate_MarshalError(t *testing.T) {
	c := New("http://127.0.0.1:0")
	_, err := c.Campaigns().ByID("c1").Assets().Generate(
		context.Background(), make(chan int),
	)
	if err == nil {
		t.Error("expected error for unmarshalable body; got nil")
	}
}
