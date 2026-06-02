// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package echo_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	goecho "github.com/labstack/echo/v4"

	echoadapter "github.com/ai-mpathyminds/yaagents-sdk-go/adapters/echo"
)

// TestEchoAdapter_Created201RoundTrip mounts an httptest.NewServer backed by
// an echo router with InjectContext() middleware, fires a POST, and asserts
// the 201 + Content-Type + X-YAAgents-Profile: v0.3 round-trip.
func TestEchoAdapter_Created201RoundTrip(t *testing.T) {
	e := goecho.New()
	e.Use(echoadapter.InjectContext())
	e.POST("/campaigns/:id/optimizations", func(c goecho.Context) error {
		ctx := echoadapter.FromRequest(c.Request())
		id := echoadapter.URLParam(c.Request(), "id")
		ar := echoadapter.AgenticResponse{}
		return echoadapter.Write(c.Response(), ar.Created(ctx, map[string]string{"campaignId": id}))
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost,
		srv.URL+"/campaigns/cmp-123/optimizations", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Correlation-ID", "corr-echo-1")
	req.Header.Set("X-Request-ID", "req-echo-1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if prof := resp.Header.Get("X-YAAgents-Profile"); prof != "v0.3" {
		t.Errorf("X-YAAgents-Profile = %q, want v0.3", prof)
	}

	raw, _ := io.ReadAll(resp.Body)
	var body map[string]string
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("json.Unmarshal: %v (raw: %s)", err, raw)
	}
	if body["campaignId"] != "cmp-123" {
		t.Errorf("URLParam round-trip: campaignId = %q, want cmp-123", body["campaignId"])
	}
}
