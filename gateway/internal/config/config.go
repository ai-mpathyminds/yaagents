// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package config loads gateway runtime configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Config holds all gateway runtime settings.
type Config struct {
	// Port is the TCP port the gateway listens on (env: GATEWAY_PORT, default: 8120).
	Port string
	// RoutesFile is the path to the YAML route configuration file
	// (env: GATEWAY_ROUTES_FILE, default: routes.yaml).
	RoutesFile string
	// AuditLog is the sink for structured audit events: "stdout" (default) or a
	// file path (env: GATEWAY_AUDIT_LOG). Used by WI-1yaa.GW-5.
	AuditLog string
	// PluginsFile is the optional path to a standalone plugins YAML file
	// (env: GATEWAY_PLUGINS_FILE). When empty, the loader reads the plugins: block
	// from RoutesFile. Used by WI-2yaa.PLG-2.
	PluginsFile string
	// JWTSecret is the HS256 secret for dev/demo JWT validation
	// (env: GATEWAY_JWT_SECRET). Injected into token-validator plugin config
	// when the plugins YAML block omits it (PRD §5.4.1).
	JWTSecret string
	// JWTJWKSURL is the JWKS URL for RS256 production JWT validation
	// (env: GATEWAY_JWT_JWKS_URL). Injected into token-validator plugin config
	// when the plugins YAML block omits it (PRD §5.4.1).
	JWTJWKSURL string
	// ShutdownTimeoutS is the graceful-shutdown deadline in seconds
	// (env: GATEWAY_SHUTDOWN_TIMEOUT_S, default: 30).
	// The gateway drains in-flight requests and runs plugin.Shutdown within
	// this window before exiting (PLG-6 / PRD §6.4).
	ShutdownTimeoutS int
	// LLMMaxSSEPerTenant is the maximum concurrent SSE connections allowed per
	// tenant (env: GATEWAY_LLM_MAX_SSE_PER_TENANT, default: 10; LLM-2).
	// Excess connections receive 429 application/vnd.yaagents.error+json with
	// retryAfter: 60. Values ≤ 0 are clamped to the default.
	LLMMaxSSEPerTenant int
}

// Load reads gateway configuration from environment variables, applying defaults.
func Load() Config {
	return Config{
		Port:             envOr("GATEWAY_PORT", "8120"),
		RoutesFile:       envOr("GATEWAY_ROUTES_FILE", "routes.yaml"),
		AuditLog:         envOr("GATEWAY_AUDIT_LOG", "stdout"),
		PluginsFile:      os.Getenv("GATEWAY_PLUGINS_FILE"),
		JWTSecret:        os.Getenv("GATEWAY_JWT_SECRET"),
		JWTJWKSURL:       os.Getenv("GATEWAY_JWT_JWKS_URL"),
		ShutdownTimeoutS:   envInt("GATEWAY_SHUTDOWN_TIMEOUT_S", 30),
		LLMMaxSSEPerTenant: envInt("GATEWAY_LLM_MAX_SSE_PER_TENANT", 10),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
