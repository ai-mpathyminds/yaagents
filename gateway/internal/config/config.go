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
}

// Load reads gateway configuration from environment variables, applying defaults.
func Load() Config {
	return Config{
		Port:       envOr("GATEWAY_PORT", "8120"),
		RoutesFile: envOr("GATEWAY_ROUTES_FILE", "routes.yaml"),
		AuditLog:   envOr("GATEWAY_AUDIT_LOG", "stdout"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
