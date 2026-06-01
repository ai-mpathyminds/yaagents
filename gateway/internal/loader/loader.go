// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package loader implements the yaagents gateway plugin YAML loader.
//
// Boot sequence per PRD §6.4 and ADR PI2-yaa-0001:
//  1. Plugin init() functions run as import side-effects (registration into the global registry).
//  2. Load reads the plugins: YAML sequence in declaration order.
//  3. The always-on assertion verifies token-validator is not disabled.
//  4. GATEWAY_JWT_SECRET / GATEWAY_JWT_JWKS_URL are merged into token-validator config
//     when the YAML block omits them (PRD §5.4.1 convenience overrides).
//  5. Each listed plugin's Init(cfg) is called; a non-nil return is a fatal boot error.
//  6. Loader.Chain(next) composes the per-request middleware chain in declaration order.
package loader

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/routes"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

const tokenValidatorName = "token-validator"

// entry holds one plugin's resolved name and its configuration map (name key removed).
type entry struct {
	name string
	cfg  map[string]any
}

// pluginsEnvelope is the YAML envelope used for both standalone plugin files and
// the inline plugins: block inside routes.yaml. The sequence preserves declaration order.
type pluginsEnvelope struct {
	Plugins []map[string]any `yaml:"plugins"`
}

// Loader holds the ordered, initialized plugin list and is ready for request-time
// chain composition and graceful-shutdown orchestration.
type Loader struct {
	log     *slog.Logger
	ordered []plugin.Plugin
}

// Load reads plugin definitions from pluginsPath (standalone file; env: GATEWAY_PLUGINS_FILE)
// when non-empty, or falls back to the inline plugins: block inside routesPath.
// It returns a ready Loader or a non-nil error that the caller should treat as a
// fatal boot failure (os.Exit(1)).
//
// envSecret and envJWKSURL are the raw values of GATEWAY_JWT_SECRET and
// GATEWAY_JWT_JWKS_URL, injected into token-validator config when the YAML omits them.
func Load(log *slog.Logger, pluginsPath, routesPath, envSecret, envJWKSURL string) (*Loader, error) {
	entries, err := readEntries(pluginsPath, routesPath)
	if err != nil {
		return nil, err
	}

	if err := assertTokenValidatorAlwaysOn(entries); err != nil {
		return nil, err
	}
	mergeTokenValidatorEnv(entries, envSecret, envJWKSURL)

	reg := plugin.Registered()
	regByName := make(map[string]plugin.Plugin, len(reg))
	for _, p := range reg {
		regByName[p.Name()] = p
	}

	var ordered []plugin.Plugin
	for _, e := range entries {
		p, ok := regByName[e.name]
		if !ok {
			log.Warn("plugin listed in config but not registered; skipping",
				slog.String("plugin", e.name))
			continue
		}
		cfg := plugin.NewMapConfig(e.cfg)
		if initErr := p.Init(cfg); initErr != nil {
			return nil, fmt.Errorf("plugin %q Init failed: %w", e.name, initErr)
		}
		ordered = append(ordered, p)
		log.Info("plugin initialized", slog.String("plugin", e.name))
	}
	return &Loader{log: log, ordered: ordered}, nil
}

// Chain composes the initialized plugins into a middleware chain in declaration
// order and returns the resulting handler. next is the innermost handler (the
// route dispatcher); plugins wrap it outermost-first in YAML declaration order.
func (l *Loader) Chain(next http.Handler) http.Handler {
	h := next
	for i := len(l.ordered) - 1; i >= 0; i-- {
		h = l.ordered[i].Handler(h)
	}
	return h
}

// ChainFor builds a per-route middleware chain in declaration order, skipping
// any plugin whose entry in perRoutePlugins carries enabled: false (PLG-6
// per-route override, PRD §5.4.2). next is the innermost handler.
//
// token-validator may not be disabled per-route; call ValidateRouteOverrides
// at boot to enforce this invariant before building per-route chains.
func (l *Loader) ChainFor(perRoutePlugins map[string]map[string]any, next http.Handler) http.Handler {
	h := next
	for i := len(l.ordered) - 1; i >= 0; i-- {
		p := l.ordered[i]
		if override, ok := perRoutePlugins[p.Name()]; ok {
			if en, ok := override["enabled"].(bool); ok && !en {
				continue // this plugin is disabled for this route
			}
		}
		h = p.Handler(h)
	}
	return h
}

// ValidateRouteOverrides returns a fatal boot error if any route's plugins:
// block disables token-validator. token-validator is the always-on security
// floor (ADR PI2-yaa-0001 §5) and must run on every proxied request.
func ValidateRouteOverrides(routeList []routes.Route) error {
	for _, r := range routeList {
		if override, ok := r.Plugins[tokenValidatorName]; ok {
			if en, ok := override["enabled"].(bool); ok && !en {
				return fmt.Errorf("route %q: token-validator cannot be disabled per-route "+
					"(ADR PI2-yaa-0001 §5 always-on invariant)", r.ID)
			}
		}
	}
	return nil
}

// Shutdown calls each plugin's Shutdown in reverse declaration order
// (ADR PI2-yaa-0001 §4). Errors are logged but do not abort remaining shutdowns.
func (l *Loader) Shutdown(ctx context.Context) {
	for i := len(l.ordered) - 1; i >= 0; i-- {
		p := l.ordered[i]
		if err := p.Shutdown(ctx); err != nil {
			l.log.Error("plugin shutdown error",
				slog.String("plugin", p.Name()),
				slog.String("error", err.Error()))
		}
	}
}

// readEntries parses the plugins: YAML sequence from pluginsPath (when non-empty)
// or from the inline block in routesPath. Returns entries in declaration order.
func readEntries(pluginsPath, routesPath string) ([]entry, error) {
	path := pluginsPath
	if path == "" {
		path = routesPath
	}
	data, err := os.ReadFile(path) // #nosec G304 — operator-supplied config file path
	if err != nil {
		return nil, fmt.Errorf("loading plugins config from %q: %w", path, err)
	}
	var env pluginsEnvelope
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing plugins YAML in %q: %w", path, err)
	}
	entries := make([]entry, 0, len(env.Plugins))
	for i, raw := range env.Plugins {
		name, _ := raw["name"].(string)
		if name == "" {
			return nil, fmt.Errorf("plugins[%d]: missing required 'name' field", i)
		}
		cfg := make(map[string]any, len(raw))
		for k, v := range raw {
			if k != "name" {
				cfg[k] = v
			}
		}
		entries = append(entries, entry{name: name, cfg: cfg})
	}
	return entries, nil
}

// assertTokenValidatorAlwaysOn returns an error when the token-validator entry
// is present in the plugins sequence but has enabled: false.
// Per ADR PI2-yaa-0001 §5, token-validator is the security floor and cannot be disabled.
func assertTokenValidatorAlwaysOn(entries []entry) error {
	for _, e := range entries {
		if e.name != tokenValidatorName {
			continue
		}
		if en, ok := e.cfg["enabled"].(bool); ok && !en {
			return fmt.Errorf("token-validator cannot be disabled: " +
				"set enabled: true or remove the 'enabled: false' key from the plugins config")
		}
		return nil
	}
	return nil
}

// mergeTokenValidatorEnv injects GATEWAY_JWT_SECRET / GATEWAY_JWT_JWKS_URL into
// the token-validator config entry when the YAML block does not supply them.
// The YAML value takes precedence if already present (PRD §5.4.1).
func mergeTokenValidatorEnv(entries []entry, envSecret, envJWKSURL string) {
	for i := range entries {
		if entries[i].name != tokenValidatorName {
			continue
		}
		if _, ok := entries[i].cfg["jwt_secret"]; !ok && envSecret != "" {
			entries[i].cfg["jwt_secret"] = envSecret
		}
		if _, ok := entries[i].cfg["jwks_url"]; !ok && envJWKSURL != "" {
			entries[i].cfg["jwks_url"] = envJWKSURL
		}
		return
	}
}
