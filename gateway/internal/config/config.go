// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package config loads gateway runtime configuration from environment variables.
package config

import "os"

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
}

// Load reads gateway configuration from environment variables, applying defaults.
func Load() Config {
	return Config{
		Port:        envOr("GATEWAY_PORT", "8120"),
		RoutesFile:  envOr("GATEWAY_ROUTES_FILE", "routes.yaml"),
		AuditLog:    envOr("GATEWAY_AUDIT_LOG", "stdout"),
		PluginsFile: os.Getenv("GATEWAY_PLUGINS_FILE"),
		JWTSecret:   os.Getenv("GATEWAY_JWT_SECRET"),
		JWTJWKSURL:  os.Getenv("GATEWAY_JWT_JWKS_URL"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
