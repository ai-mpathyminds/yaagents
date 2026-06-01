// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package promptsanitize implements the prompt-sanitize stub plugin
// (PRD §6.5 plugin d).
//
// # Stub status
//
// Full prompt-injection defence is deferred to PI3-yaa or community
// (ADR PI2-yaa-0005 Decision 2 / OQ-7). This implementation:
//   - Stores configuration in Init (no-op validation beyond that).
//   - Passes every request through to next without modification.
//   - Emits a single structured warn log on the first request when
//     enabled: true, using a [sync.Once] guard so the message is not
//     spammed across concurrent requests.
//
// The plugin does NOT exit 1 at boot when enabled — log-and-pass-through
// is the chosen behaviour per OQ-7 resolution.
//
// Registration: init() calls plugin.Register so the gateway wires this plugin
// by import side-effect (ADR PI2-yaa-0001 §3; no plugin.Open / dlopen).
package promptsanitize

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&PromptSanitize{})
}

// PromptSanitize is the prompt-sanitize stub plugin.
// Zero value is valid (enabled defaults to false); Init stores config.
type PromptSanitize struct {
	enabled  bool
	warnOnce sync.Once
}

// Name returns the canonical plugin identifier.
func (p *PromptSanitize) Name() string { return "prompt-sanitize" }

// Init stores the enabled flag. No-op validation — the stub has no
// configurable parameters beyond enabled that require checking.
func (p *PromptSanitize) Init(cfg plugin.PluginConfig) error {
	p.enabled = cfg.GetBool("enabled")
	return nil
}

// Handler returns a pass-through middleware. When enabled, it emits a single
// structured warn log on the first request (sync.Once) noting that the full
// prompt-injection defence is deferred.
func (p *PromptSanitize) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.enabled {
			p.warnOnce.Do(func() {
				slog.Warn("prompt-sanitize is a stub; full prompt-injection defence deferred to PI3-yaa or community")
			})
		}
		next.ServeHTTP(w, r)
	})
}

// Shutdown is a no-op; the plugin holds no background resources.
func (p *PromptSanitize) Shutdown(_ context.Context) error { return nil }
