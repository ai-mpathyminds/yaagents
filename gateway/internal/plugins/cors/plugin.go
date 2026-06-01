// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package cors implements the CORS plugin for the yaagents gateway (WI-2yaa.LLM-3).
//
// The plugin performs exact-match origin gating on cross-origin requests:
//   - OPTIONS preflight + matching origin → 200 with full CORS headers (short-circuit).
//   - OPTIONS preflight + mismatched origin → 200 without CORS headers (browser blocks the real request).
//   - Non-OPTIONS + matching origin → Access-Control-Allow-Origin injected, next called.
//   - Non-OPTIONS + mismatched or absent origin → next called unchanged.
//   - allowed_origins: [] (empty) → plugin is disabled, all requests pass through to next.
//
// Configuration keys (all optional):
//
//	allowed_origins:    []string  exact-match origin list; empty = disabled.
//	allow_methods:      string    override the default allowed methods list.
//	allow_headers:      string    override the default allowed headers list.
//	max_age:            int       override the default max-age (seconds, default 86400).
//
// Registration: init() → plugin.Register(&CORSPlugin{}) per ADR PI2-yaa-0001 §3.
package cors

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

const (
	defaultAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	defaultAllowHeaders = "Content-Type, Authorization, X-Request-Id, X-Correlation-Id"
	defaultMaxAge       = 86400
)

func init() {
	plugin.Register(&CORSPlugin{})
}

// CORSPlugin handles CORS preflight and header injection.
// Zero value is valid (disabled); Init configures it from YAML.
type CORSPlugin struct {
	allowedOrigins map[string]struct{}
	allowMethods   string
	allowHeaders   string
	maxAge         string // pre-formatted as string for header
}

// Name returns the canonical plugin identifier.
func (c *CORSPlugin) Name() string { return "cors" }

// Init reads configuration and prepares the plugin.
// A non-nil error causes the gateway to exit 1.
// Empty allowed_origins is valid and disables the plugin.
func (c *CORSPlugin) Init(cfg plugin.PluginConfig) error {
	origins := cfg.GetStringSlice("allowed_origins")
	c.allowedOrigins = make(map[string]struct{}, len(origins))
	for _, o := range origins {
		c.allowedOrigins[o] = struct{}{}
	}

	c.allowMethods = cfg.GetString("allow_methods")
	if c.allowMethods == "" {
		c.allowMethods = defaultAllowMethods
	}

	c.allowHeaders = cfg.GetString("allow_headers")
	if c.allowHeaders == "" {
		c.allowHeaders = defaultAllowHeaders
	}

	maxAge := cfg.GetInt("max_age")
	if maxAge <= 0 {
		maxAge = defaultMaxAge
	}
	c.maxAge = strconv.Itoa(maxAge)

	return nil
}

// Handler returns the CORS middleware.
//
// Decision tree per request:
//  1. allowed_origins empty → pass through (disabled).
//  2. No Origin header → pass through (not a cross-origin request).
//  3. Origin present + OPTIONS → return 200 with/without CORS headers (short-circuit).
//  4. Origin present + non-OPTIONS → inject ACAO if matched, then call next.
func (c *CORSPlugin) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Disabled: pass through unchanged.
		if len(c.allowedOrigins) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		// 2. Not a cross-origin request: pass through.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		_, matched := c.allowedOrigins[origin]

		// 3. OPTIONS preflight: respond directly (do not call next).
		if r.Method == http.MethodOptions {
			if matched {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", c.allowMethods)
				w.Header().Set("Access-Control-Allow-Headers", c.allowHeaders)
				w.Header().Set("Access-Control-Max-Age", c.maxAge)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// 4. Non-OPTIONS: inject ACAO header when origin matches, then delegate.
		if matched {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		next.ServeHTTP(w, r)
	})
}

// Shutdown is a no-op; the plugin holds no background goroutines.
func (c *CORSPlugin) Shutdown(_ context.Context) error { return nil }
