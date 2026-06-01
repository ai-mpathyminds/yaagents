// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package plugin

import (
	"log/slog"
	"sync"
)

var (
	mu       sync.Mutex
	registry []Plugin
	byName   = map[string]struct{}{}
)

// Register adds p to the global plugin registry.
//
// Registration is idempotent on [Plugin.Name]: if a plugin with the same name
// has already been registered, the second call is a no-op and emits a
// warn-level log line ("first wins" semantics — ADR PI2-yaa-0001).
//
// Plugins call Register from an init() function; the gateway binary wires
// plugins by import side-effect (no plugin.Open / dlopen — ADR PI2-yaa-0001 §3).
func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	name := p.Name()
	if _, dup := byName[name]; dup {
		slog.Warn("plugin already registered; skipping duplicate",
			slog.String("plugin", name))
		return
	}
	byName[name] = struct{}{}
	registry = append(registry, p)
}

// Registered returns all registered plugins in registration order.
//
// The returned slice is a snapshot copy; callers must not modify it.
// Plugins that registered but are absent from the gateway YAML are silently
// inert — the gateway core calls Init only for plugins listed in the YAML.
func Registered() []Plugin {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Plugin, len(registry))
	copy(out, registry)
	return out
}
