// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// mock-llm-api is a canned LLM backend stub for the yaagents llm-gateway demo
// (WI-2yaa.EX-LLM-1 / ADR PI2-yaa-0002).
//
// Endpoints:
//
//	POST /completions  — JSON or SSE completions
//	GET  /healthz      — liveness probe (always 200)
//	GET  /readyz       — readiness probe (always 200)
//
// POST /completions request body (JSON):
//
//	{
//	  "model":            "mock-gpt",        // optional; any string
//	  "prompt":           "hello",           // required
//	  "stream":           false,             // true → SSE response
//	  "simulate_timeout": false              // true → sleep 60 s (triggers gateway timeout)
//	}
//
// If the request carries Accept: text/event-stream the response is always SSE
// regardless of the "stream" field value.
//
// Non-streaming response: 201 application/json (PRD §4).
// Streaming response: 200 text/event-stream; each chunk carries one token, then
// a terminal "data: [DONE]" event.
//
// Configuration (environment variables):
//
//	PORT               listen port (default: 8123)
//	STREAM_CHUNKS      number of SSE token chunks to emit (default: 5)
//	STREAM_DELAY_MS    delay in ms between SSE chunks (default: 100)
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// completionRequest is the JSON body accepted by POST /completions.
type completionRequest struct {
	Model           string `json:"model"`
	Prompt          string `json:"prompt"`
	Stream          bool   `json:"stream"`
	SimulateTimeout bool   `json:"simulate_timeout"`
	// HoldOpen, when true, uses a 2-second inter-chunk delay to keep SSE
	// connections alive for concurrency tests (EX-LLM-3 Flow 3).
	HoldOpen bool `json:"hold_open"`
}

// choice is one completion choice in a response.
type choice struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// completionResponse is the non-streaming 201 response body.
type completionResponse struct {
	Type    string   `json:"type"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

// deltaEvent is one SSE chunk body.
type deltaEvent struct {
	Type    string   `json:"type"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

// config holds resolved runtime parameters.
type config struct {
	port        string
	streamChunks int
	streamDelay  time.Duration
}

func loadConfig() config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8123"
	}

	chunks := 5
	if v := os.Getenv("STREAM_CHUNKS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			chunks = n
		}
	}

	delayMs := 100
	if v := os.Getenv("STREAM_DELAY_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			delayMs = n
		}
	}

	return config{
		port:        port,
		streamChunks: chunks,
		streamDelay:  time.Duration(delayMs) * time.Millisecond,
	}
}

var cfg config

func handleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req completionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, `{"error":"prompt is required"}`, http.StatusBadRequest)
		return
	}

	model := req.Model
	if model == "" {
		model = "mock-gpt"
	}

	// Simulate a long-running upstream that triggers an execution timeout.
	// Use this to demonstrate the gateway's executionTimeoutSeconds=30 feature.
	if req.SimulateTimeout {
		slog.Info("mock-llm-api: simulating timeout (sleeping 60s)")
		select {
		case <-time.After(60 * time.Second):
		case <-r.Context().Done():
			slog.Info("mock-llm-api: context cancelled during timeout simulation")
			return
		}
	}

	// SSE mode: request has stream:true OR client accepts text/event-stream.
	wantsSSE := req.Stream || strings.Contains(r.Header.Get("Accept"), "text/event-stream")
	if wantsSSE {
		delay := cfg.streamDelay
		if req.HoldOpen {
			// Slow-stream mode for concurrency testing (EX-LLM-3 Flow 3).
			// 2-second inter-chunk delay keeps SSE slots occupied long enough
			// to trigger the gateway's per-tenant SSE concurrency limit.
			delay = 2 * time.Second
		}
		handleSSE(w, r, model, delay)
		return
	}

	// Standard non-streaming JSON response (201 per PRD §4).
	text := fmt.Sprintf("This is a mocked completion for: %s", req.Prompt)
	resp := completionResponse{
		Type:  "completion",
		Model: model,
		Choices: []choice{
			{Index: 0, Text: text},
		},
	}
	// Echo X-Correlation-Id so conformance-test's correlation-ID check passes
	// (the SSE proxy forwards all upstream response headers to the client).
	if cid := r.Header.Get("X-Correlation-Id"); cid != "" {
		w.Header().Set("X-Correlation-Id", cid)
	}
	// Also handle the canonical Go net/http form "X-Correlation-ID"
	if cid := r.Header.Get("X-Correlation-ID"); cid != "" && r.Header.Get("X-Correlation-Id") == "" {
		w.Header().Set("X-Correlation-Id", cid)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func handleSSE(w http.ResponseWriter, r *http.Request, model string, delay time.Duration) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Emit a series of token chunks.
	tokens := buildTokens(cfg.streamChunks)
	for i, token := range tokens {
		select {
		case <-r.Context().Done():
			slog.Info("mock-llm-api: SSE stream cancelled by client")
			return
		default:
		}

		evt := deltaEvent{
			Type:  "completion.delta",
			Model: model,
			Choices: []choice{
				{Index: 0, Text: token},
			},
		}
		data, _ := json.Marshal(evt)
		_, err := fmt.Fprintf(w, "data: %s\n\n", data)
		if err != nil {
			slog.Info("mock-llm-api: write error during SSE", slog.String("error", err.Error()))
			return
		}
		flusher.Flush()

		if i < len(tokens)-1 && delay > 0 {
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		}
	}

	// Terminal [DONE] event per OpenAI-style SSE convention.
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
	slog.Info("mock-llm-api: SSE stream completed", slog.Int("chunks", len(tokens)))
}

// buildTokens returns a slice of n short token strings for the SSE demo.
func buildTokens(n int) []string {
	phrases := []string{
		"This ", "is ", "a ", "mocked ", "LLM ", "response ",
		"from ", "the ", "yaagents ", "demo ",
	}
	tokens := make([]string, n)
	for i := range tokens {
		tokens[i] = phrases[i%len(phrases)]
	}
	return tokens
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ok"}`)
}

func handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ready"}`)
}

func main() {
	cfg = loadConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/completions", handleCompletions)
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", handleReadyz)

	addr := ":" + cfg.port
	slog.Info("mock-llm-api: starting",
		slog.String("addr", addr),
		slog.Int("stream_chunks", cfg.streamChunks),
		slog.Duration("stream_delay", cfg.streamDelay),
	)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("mock-llm-api: server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
