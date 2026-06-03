// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Command store-go is the Go reference example for the YAAgents Agentic REST Profile v0.3.
// It mirrors examples/store/ (Python) using sdk-go + net/http (no router framework).
//
// Endpoint:
//
//	POST /products/{id}/recommendations — returns same-category mock recommendations.
//
// To extend with a real LLM, see examples/store/skills/ for AI-tool starter prompts
// (the skill files apply equally to the Python and Go implementations).
//
// Port: 8121 (portfolio band 8120-8129; portfolio-conventions.md §Port Allocation).
package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	sdkgo "github.com/ai-mpathyminds/yaagents-sdk-go/sdkgo"
)

const (
	defaultPort    = "8121"
	profileVersion = "v0.3"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	products, err := loadProducts()
	if err != nil {
		log.Error("failed to load products.json", "error", err)
		os.Exit(1)
	}
	customers, err := loadCustomers()
	if err != nil {
		log.Error("failed to load customers.json", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz)
	mux.HandleFunc("POST /products/{id}/recommendations",
		handleRecommendations(log, products, customers))

	addr := ":" + port
	log.Info("store-go starting",
		slog.String("port", port),
		slog.String("profile", profileVersion),
	)

	srv := &http.Server{Addr: addr, Handler: mux}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"status":"ok"}`)
}

func handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"status":"ok","profile":"v0.3"}`)
}

// recommendRequest is the JSON body for POST /products/{id}/recommendations.
type recommendRequest struct {
	Limit            int  `json:"limit"`
	ExcludePurchased bool `json:"exclude_purchased"`
}

// handleRecommendations returns the net/http handler for POST /products/{id}/recommendations.
func handleRecommendations(
	log *slog.Logger,
	products []Product,
	customers []Customer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := sdkgo.FromRequest(r)
		var ar sdkgo.AgenticResponse
		productID := r.PathValue("id")

		log.Info("recommend request",
			slog.String("productId", productID),
			slog.String("correlationId", ctx.CorrelationID),
		)

		// Locate seed product.
		seed := findProduct(products, productID)
		if seed == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"detail":"Product not found."}`)
			return
		}

		// Parse optional request body.
		var req recommendRequest
		req.Limit = 3
		req.ExcludePurchased = true
		if r.ContentLength > 0 {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			r.Body.Close()
			if err != nil {
				_ = sdkgo.Write(w, ar.Failed(ctx, "could not read request body"))
				return
			}
			if len(body) > 0 {
				if err := json.Unmarshal(body, &req); err != nil {
					_ = sdkgo.Write(w, ar.Failed(ctx, "invalid JSON body"))
					return
				}
			}
		}

		// Get optional customer.
		customerID := r.Header.Get("X-Customer-Id")
		var customer *Customer
		if customerID != "" {
			customer = findCustomer(customers, customerID)
		}

		// Build recommendations.
		recommendations, reason := mockRecommend(products, seed, customer, req.Limit, req.ExcludePurchased)

		_ = sdkgo.Write(w, ar.Done(ctx, map[string]any{
			"seed_product_id": productID,
			"recommendations": recommendations,
			"reasoning":       reason,
		}))
	}
}
