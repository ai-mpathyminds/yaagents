package config_test

import (
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure env vars are unset.
	t.Setenv("GATEWAY_PORT", "")
	t.Setenv("GATEWAY_ROUTES_FILE", "")
	t.Setenv("GATEWAY_AUDIT_LOG", "")

	cfg := config.Load()

	if cfg.Port != "8120" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "8120")
	}
	if cfg.RoutesFile != "routes.yaml" {
		t.Errorf("RoutesFile: got %q, want %q", cfg.RoutesFile, "routes.yaml")
	}
	if cfg.AuditLog != "stdout" {
		t.Errorf("AuditLog: got %q, want %q", cfg.AuditLog, "stdout")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("GATEWAY_PORT", "9000")
	t.Setenv("GATEWAY_ROUTES_FILE", "/etc/gw/routes.yaml")
	t.Setenv("GATEWAY_AUDIT_LOG", "/var/log/gw-audit.ndjson")

	cfg := config.Load()

	if cfg.Port != "9000" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "9000")
	}
	if cfg.RoutesFile != "/etc/gw/routes.yaml" {
		t.Errorf("RoutesFile: got %q, want %q", cfg.RoutesFile, "/etc/gw/routes.yaml")
	}
	if cfg.AuditLog != "/var/log/gw-audit.ndjson" {
		t.Errorf("AuditLog: got %q, want %q", cfg.AuditLog, "/var/log/gw-audit.ndjson")
	}
}

