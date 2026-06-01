// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tokenvalidator provides the PI1-yaa auth + tenant-context bridge,
// expressed as a Plugin so the gateway's plugin chain replaces the former
// hardcoded authMiddle(ctxMiddle(dispatcher)) pattern.
//
// PLG-3 (Sprint 2) will supersede this bridge with the full JWT RS256 / JWKS
// implementation. At that point this package will be removed and main.go will
// import the PLG-3 package instead.
//
// Registration: init() calls plugin.Register so the gateway binary wires this
// plugin by import side-effect (ADR PI2-yaa-0001 §3).
package tokenvalidator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/auth"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/logger"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/tenant"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&TokenValidator{})
}

// TokenValidator is the PI1-yaa JWT validation + tenant-context bridge plugin.
// It composes auth.Middleware and tenant.ContextMiddleware in exactly the same
// order as the PI1-yaa hardcoded chain.
type TokenValidator struct {
	chain func(http.Handler) http.Handler
}

// Name returns the canonical plugin name used for registry lookup and YAML matching.
func (tv *TokenValidator) Name() string { return "token-validator" }

// Init validates the configuration and constructs the auth + context middleware chain.
// Returns an error (boot failure) when:
//   - enabled: false is set (token-validator is always-on per ADR PI2-yaa-0001 §5)
//   - neither jwt_secret nor jwks_url is configured
//
// The loader merges GATEWAY_JWT_SECRET / GATEWAY_JWT_JWKS_URL into cfg before
// calling Init when the YAML block omits them (PRD §5.4.1).
func (tv *TokenValidator) Init(cfg plugin.PluginConfig) error {
	if !cfg.GetBool("enabled") {
		return fmt.Errorf("token-validator cannot be disabled")
	}

	jwksURL := cfg.GetString("jwks_url")
	jwtSecret := cfg.GetString("jwt_secret")

	var v auth.Validator
	switch {
	case jwksURL != "":
		// JWKS takes precedence (ADR PI1-yaa-0001 §3).
		v = auth.NewJWKSValidator(jwksURL)
	case jwtSecret != "":
		v = auth.NewHS256Validator(jwtSecret)
	default:
		return fmt.Errorf("token-validator: neither jwks_url nor jwt_secret is configured; " +
			"set one in the plugin config or via GATEWAY_JWT_JWKS_URL / GATEWAY_JWT_SECRET")
	}

	log := logger.New()
	authMiddle := auth.Middleware(v, log)
	ctxMiddle := tenant.ContextMiddleware(log)
	tv.chain = func(next http.Handler) http.Handler {
		return authMiddle(ctxMiddle(next))
	}
	return nil
}

// Handler returns the composed auth + tenant-context middleware for this request.
func (tv *TokenValidator) Handler(next http.Handler) http.Handler {
	return tv.chain(next)
}

// Shutdown is a no-op for the bridge plugin (no background resources to release).
func (tv *TokenValidator) Shutdown(_ context.Context) error { return nil }
