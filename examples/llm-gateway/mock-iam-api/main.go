// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// mock-iam-api is a tiny tenant-directory stub for the yaagents llm-gateway
// compose demo. It implements the tenant-injector lookup-service contract
// (WI-2yaa.PLG-4b / ADR PI2-yaa-0006 Decision 2) so the demo runs green
// without an external IAM dependency.
//
// Endpoint:
//
//	GET /api/v1/principals/{encoded-principal}/tenant
//	  200 → {"principal":"<value>","tenant_id":"<value>"}
//	  404 → principal not in tenants file
//
// Configuration:
//
//	TENANTS_FILE  path to YAML file (default: mock-tenants.yaml)
//	PORT          listen port (default: 8122)
package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// tenantsFile is the shape of mock-tenants.yaml.
type tenantsFile struct {
	Principals map[string]string `yaml:"principals"`
}

var principals map[string]string

func loadTenants(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var tf tenantsFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	principals = tf.Principals
	if principals == nil {
		principals = map[string]string{}
	}
	slog.Info("mock-iam-api: loaded tenants", slog.Int("count", len(principals)))
	return nil
}

func handleTenantLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Path: /api/v1/principals/{encoded}/tenant
	const prefix = "/api/v1/principals/"
	const suffix = "/tenant"
	p := r.URL.Path
	if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, suffix) {
		http.NotFound(w, r)
		return
	}
	encoded := p[len(prefix) : len(p)-len(suffix)]
	principal, err := url.PathUnescape(encoded)
	if err != nil || principal == "" {
		http.Error(w, "invalid principal encoding", http.StatusBadRequest)
		return
	}

	tenantID, ok := principals[principal]
	if !ok {
		slog.Info("mock-iam-api: principal not found", slog.String("principal", principal))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"principal": principal,
		"tenant_id": tenantID,
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	tenantsPath := os.Getenv("TENANTS_FILE")
	if tenantsPath == "" {
		tenantsPath = "mock-tenants.yaml"
	}
	if err := loadTenants(tenantsPath); err != nil {
		slog.Error("mock-iam-api: failed to load tenants", slog.String("error", err.Error()))
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8122"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/principals/", handleTenantLookup)
	mux.HandleFunc("/healthz", healthz)

	addr := ":" + port
	slog.Info("mock-iam-api: listening", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("mock-iam-api: server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
