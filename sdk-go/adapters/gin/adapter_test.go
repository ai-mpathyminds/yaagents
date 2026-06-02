// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package gin_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gogin "github.com/gin-gonic/gin"

	ginadapter "github.com/ai-mpathyminds/yaagents-sdk-go/adapters/gin"
)

// TestGinAdapter_Created201RoundTrip mounts an httptest.NewServer backed by a
// gin router with InjectContext() middleware, fires a POST, and asserts the
// 201 + Content-Type + X-YAAgents-Profile: v0.3 round-trip.
func TestGinAdapter_Created201RoundTrip(t *testing.T) {
	gogin.SetMode(gogin.TestMode)
	r := gogin.New()
	r.Use(ginadapter.InjectContext())
	r.POST("/campaigns/:id/optimizations", func(c *gogin.Context) {
		ctx := ginadapter.FromRequest(c.Request)
		id := ginadapter.URLParam(c.Request, "id")
		ar := ginadapter.AgenticResponse{}
		if err := ginadapter.Write(c.Writer, ar.Created(ctx, map[string]string{"campaignId": id})); err != nil {
			t.Errorf("Write: %v", err)
		}
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost,
		srv.URL+"/campaigns/cmp-123/optimizations", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Correlation-ID", "corr-gin-1")
	req.Header.Set("X-Request-ID", "req-gin-1")

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
