// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Command gateway is the yaagents API gateway — a lightweight reverse proxy
// that adds authn, tenant/actor context, RBAC, typed-response passthrough,
// audit logging, and Prometheus metrics for the Agentic REST Profile
// (ADR PI1-yaa-0001; plugin chain per ADR PI2-yaa-0001).
//
// Configuration is via environment variables:
//
//	GATEWAY_PORT              TCP port to listen on (default: 8120)
//	GATEWAY_ROUTES_FILE       Path to routes.yaml (default: routes.yaml)
//	GATEWAY_AUDIT_LOG         Audit sink: "stdout" or file path (default: stdout)
//	GATEWAY_PLUGINS_FILE      Optional path to standalone plugins.yaml
//	GATEWAY_JWT_SECRET        HS256 secret — injected into token-validator config
//	GATEWAY_JWT_JWKS_URL      JWKS URL  — injected into token-validator config
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/audit"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/config"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/loader"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/logger"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/metrics"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/proxy"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"

	// Import token-validator bridge plugin for side-effect registration
	// (ADR PI2-yaa-0001 §3). PLG-3 will replace this import.
	_ "github.com/ai-mpathyminds/yaagents/gateway/internal/plugins/tokenvalidator"
)

func main() {
	cfg := config.Load()
	log := logger.New()

	routeList, err := routes.Load(cfg.RoutesFile)
	if err != nil {
		log.Error("invalid route configuration — cannot start", "error", err.Error())
		os.Exit(1)
	}

	log.Info("gateway starting",
		slog.String("port", cfg.Port),
		slog.String("routes_file", cfg.RoutesFile),
		slog.Int("routes_loaded", len(routeList)),
	)

	// Audit log sink (WI-1yaa.GW-5).
	auditSink, closeAudit, auditErr := audit.OpenSink(cfg.AuditLog)
	if auditErr != nil {
		log.Error("cannot open audit log — cannot start", "error", auditErr.Error())
		os.Exit(1)
	}
	defer closeAudit()
	auditLog := audit.New(auditSink)

	// Prometheus metrics registry (WI-1yaa.GW-5).
	reg := metrics.New()

	// Plugin loader — reads plugins: block, validates always-on assertion,
	// merges env vars into token-validator config, calls Init in declaration order.
	// Any error is a fatal boot failure (ADR PI2-yaa-0001 §5; PRD §6.4).
	ldr, ldrErr := loader.Load(log, cfg.PluginsFile, cfg.RoutesFile, cfg.JWTSecret, cfg.JWTJWKSURL)
	if ldrErr != nil {
		log.Error("plugin loader failed — cannot start", "error", ldrErr.Error())
		os.Exit(1)
	}

	// Route dispatcher: RBAC + typed-response passthrough + audit + metrics (GW-4/GW-5).
	dispatcher, dispErr := proxy.New(routeList, log, auditLog, reg)
	if dispErr != nil {
		log.Error("failed to build route dispatcher — cannot start", "error", dispErr.Error())
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", makeReadyzHandler(len(routeList) > 0))
	mux.HandleFunc("GET /metrics", reg.Handler())
	// Catch-all: plugin chain (token-validator → … ) → route dispatcher.
	mux.Handle("/", ldr.Chain(dispatcher))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGTERM / SIGINT.
	// Shutdown order: HTTP server drain → plugin Shutdown in reverse order.
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-stopCh
		log.Info("shutdown signal received — draining", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if shutErr := srv.Shutdown(ctx); shutErr != nil {
			log.Error("graceful shutdown error", "error", shutErr.Error())
		}
		ldr.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server listen error", "error", err.Error())
		os.Exit(1)
	}
	log.Info("gateway stopped")
}

// handleHealthz responds 200 OK for liveness checks (always up while running).
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `{"status":"ok"}`)
}

// makeReadyzHandler returns a readiness handler: 200 when at least one route is
// loaded, 503 otherwise. Routes are validated at boot so a non-zero count means
// configuration is valid (WI-1yaa.GW-5 AC).
func makeReadyzHandler(ready bool) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if ready {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"status":"ready"}`)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprint(w, `{"status":"not ready","reason":"no routes loaded"}`)
		}
	}
}
