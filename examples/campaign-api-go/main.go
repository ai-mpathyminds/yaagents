// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Command campaign-api-go is the Go reference example for the YAAgents Agentic REST Profile v0.3.
// It mirrors examples/campaign-api/ (Python) using sdk-go + net/http (no router framework).
//
// PRD §8.2: demonstrates the pure sdk-go sequence:
//
//	sdkgo.FromRequest(r) → sdkgo.AgenticResponse{} → sdkgo.Write(w, ...)
//
// Port: 8121 (portfolio band 8120-8129; portfolio-conventions.md §Port Allocation).
//
// Five handler flows (PRD §13.2 / §8.1):
//
//  1. Clarification     — missing goal field     → 400 application/vnd.yaagents.clarification+json
//  2. Created           — valid body             → 201 application/json
//  3. Accepted (async)  — Prefer: respond-async  → 202 application/vnd.yaagents.operation+json
//  4. Validation failed — invalid field type     → 422 application/vnd.yaagents.validation-error+json
//  5. Auth failure      — no JWT (via gateway)   → 401 (gateway-side; no upstream involvement)
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	sdkgo "github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

const (
	defaultPort    = "8121"
	profileVersion = "v0.3"
	maxBodyBytes   = 1 << 20 // 1 MiB
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz)
	mux.HandleFunc("POST /campaigns/{id}/optimizations", handleOptimizations(log))

	addr := ":" + port
	log.Info("campaign-api-go starting",
		slog.String("port", port),
		slog.String("profile", profileVersion),
	)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

// handleHealthz returns a liveness response (GET /healthz).
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"status":"ok"}`)
}

// handleReadyz returns a readiness response with profile version (GET /readyz).
func handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"ok","profile":%q}`, profileVersion)
}

// handleOptimizations returns the net/http handler for POST /campaigns/{id}/optimizations.
//
// It implements all five PRD §13.2 / §8.1 demo flows using the pure sdk-go
// sequence: sdkgo.FromRequest(r) → sdkgo.AgenticResponse{} → sdkgo.Write(w, resp).
//
// Flow resolution order:
//  1. Body parse failure           → ValidationFailed (422)
//  2. Missing goal field           → ClarificationRequired (400)
//  3. goal field present but wrong type → ValidationFailed (422)
//  4. Prefer: respond-async header → Accepted (202)
//  5. All valid                    → Created (201)
func handleOptimizations(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: extract gateway-injected context headers.
		ctx := sdkgo.FromRequest(r)
		var ar sdkgo.AgenticResponse
		campaignID := r.PathValue("id")

		log.Info("optimize request",
			slog.String("campaignId", campaignID),
			slog.String("correlationId", ctx.CorrelationID),
			slog.String("tenant", ctx.ActorTenant),
		)

		// Step 2: read body.
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
		r.Body.Close()
		if err != nil {
			_ = sdkgo.Write(w, ar.Failed(ctx, "could not read request body"))
			return
		}

		// Step 3: parse into a raw map so we can distinguish missing vs wrong-type.
		var rawMap map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawMap); err != nil {
			// Flow 4: body is not a valid JSON object.
			_ = sdkgo.Write(w, ar.ValidationFailed(ctx, []sdkgo.ValidationError{
				{Field: "body", Message: "request body must be a valid JSON object"},
			}))
			return
		}

		// Flow 1: Clarification — goal key absent or explicitly null.
		goalRaw, hasGoal := rawMap["goal"]
		if !hasGoal || string(goalRaw) == "null" {
			_ = sdkgo.Write(w, ar.ClarificationRequired(ctx, []sdkgo.RequiredInput{
				{
					Name:          "goal",
					Location:      "body",
					Type:          "string",
					Required:      true,
					Question:      "Which optimization goal should this campaign target?",
					AllowedValues: []string{"ctr", "cpl", "conversion_rate", "roas"},
				},
			}))
			return
		}

		// Flow 4: Validation failed — goal is present but not a string or is empty.
		var goal string
		if err := json.Unmarshal(goalRaw, &goal); err != nil {
			_ = sdkgo.Write(w, ar.ValidationFailed(ctx, []sdkgo.ValidationError{
				{
					Field:   "goal",
					Message: `must be a string (e.g. "ctr", "cpl", "conversion_rate", "roas")`,
				},
			}))
			return
		}
		if goal == "" {
			// Empty string is treated as clarification (value unknown, not invalid type).
			_ = sdkgo.Write(w, ar.ClarificationRequired(ctx, []sdkgo.RequiredInput{
				{
					Name:          "goal",
					Location:      "body",
					Type:          "string",
					Required:      true,
					Question:      "Which optimization goal should this campaign target?",
					AllowedValues: []string{"ctr", "cpl", "conversion_rate", "roas"},
				},
			}))
			return
		}

		// Flow 3: Accepted (async) — Prefer: respond-async header present.
		if r.Header.Get("Prefer") == "respond-async" {
			opID := fmt.Sprintf("op-%s-%s", campaignID, goal)
			_ = sdkgo.Write(w, ar.Accepted(ctx, opID))
			return
		}

		// Flow 2: Created — valid body, synchronous response.
		opt := map[string]any{
			"id":         fmt.Sprintf("opt-%s-%s-1", campaignID, goal),
			"campaignId": campaignID,
			"goal":       goal,
			"suggestions": []map[string]any{
				{
					"id":              "sug-1",
					"objective":       goal,
					"suggestion":      fmt.Sprintf("Increase bid for %q by 10%% during peak hours.", goal),
					"estimatedImpact": "+12% improvement",
				},
			},
		}
		_ = sdkgo.Write(w, ar.Created(ctx, map[string]any{"optimization": opt}))
	}
}
