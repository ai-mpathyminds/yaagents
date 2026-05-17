// Command gateway is the yaagents API gateway — a lightweight reverse proxy
// that adds authn, tenant/actor context, RBAC, and typed-response passthrough
// for the Agentic REST Profile (ADR PI1-yaa-0001).
//
// Configuration is via environment variables:
//
//	GATEWAY_PORT          TCP port to listen on (default: 8120)
//	GATEWAY_ROUTES_FILE   Path to routes.yaml (default: routes.yaml)
//	GATEWAY_AUDIT_LOG     Audit sink: "stdout" or file path (default: stdout)
//	GATEWAY_JWT_SECRET    HS256 secret for dev/demo JWT validation (WI-1yaa.GW-2)
//	GATEWAY_JWT_JWKS_URL  JWKS URL for RS256 production JWT validation (WI-1yaa.GW-2)
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

	"github.com/ai-mpathyminds/yaagents/gateway/internal/config"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/logger"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", makeReadyzHandler(len(routeList) > 0))

	// Routing middleware (auth, RBAC, proxy) wired in WI-1yaa.GW-2 through GW-4.

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGTERM / SIGINT.
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

// makeReadyzHandler returns a readiness handler that returns 200 when routes are
// loaded and 503 otherwise. Full readiness logic (JWKS reachability, etc.) is
// added in WI-1yaa.GW-5.
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
