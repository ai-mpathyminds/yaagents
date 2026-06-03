// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package communityplugin is an example community plugin demonstrating the
// yaagents gateway plugin authoring contract (PRD §6.6, ADR PI2-yaa-0001 §3).
//
// # How to use this plugin in a custom gateway binary
//
//  1. In your gateway binary's main.go, add a blank import:
//
//     import _ "github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin"
//
//  2. Add the module to go.mod:
//
//     require github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin v0.0.0
//     replace github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin => <local-path>
//
//  3. Build the binary — the gateway loader picks up the plugin via
//     plugin.Registered() (ADR PI2-yaa-0001 §3).
//
// This plugin adds the response header X-Community-Plugin: active to every
// proxied response, proving the community plugin authoring contract.
//
// The gateway NEVER uses dlopen/plugin.Open — community plugins are compiled
// into the operator binary at build time (ADR PI2-yaa-0001 §3; PRD §10 [SEC]).
package communityplugin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ai-mpathyminds/yaagents-gateway/plugin"
)

// pluginVersion is the community plugin's own version string.
// Decoupled from the gateway profile version — operators version their plugins
// independently (ADR PI2-yaa-0001 §1).
const pluginVersion = "0.1.0"

func init() {
	plugin.Register(&CommunityPlugin{})
}

// CommunityPlugin is a minimal example plugin that injects a custom response
// header on every proxied request.  Production plugins would implement more
// sophisticated logic here (rate-limiting, header enrichment, etc.).
type CommunityPlugin struct{}

// Name returns the stable plugin identifier.  The gateway YAML must use this
// exact string in the plugins: list to activate the plugin.
func (p *CommunityPlugin) Name() string { return "community-example" }

// Init is called once at gateway startup with the plugin's config block.
// A non-nil return causes the gateway to exit 1 (boot-time fatal error).
func (p *CommunityPlugin) Init(_ plugin.PluginConfig) error {
	slog.Info("community-plugin: registered",
		slog.String("plugin", p.Name()),
		slog.String("version", pluginVersion),
	)
	return nil
}

// Handler returns the middleware function.  The gateway composes plugins in
// YAML declaration order; this plugin runs after token-validator and
// tenant-injector (as declared in plugins.yaml).
func (p *CommunityPlugin) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inject a custom response header so callers can verify the plugin is active.
		w.Header().Set("X-Community-Plugin", "active")
		next.ServeHTTP(w, r)
	})
}

// Shutdown is a no-op; this plugin holds no background goroutines or resources.
func (p *CommunityPlugin) Shutdown(_ context.Context) error { return nil }
