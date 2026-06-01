// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package plugin defines the Plugin interface, the PluginConfig accessor, and
// the global plugin registry for the yaagents gateway.
//
// Plugins register themselves at process-init time via [Register] inside an
// init() function. The gateway core reads the ordered list via [Registered]
// and composes the middleware chain in declaration order.
//
// Security note: the gateway NEVER calls plugin.Open() or dlopen. Community
// plugins are compiled into the operator binary via Go module imports at build
// time (ADR PI2-yaa-0001 §3; PRD §10 [SEC]).
package plugin

import (
	"context"
	"net/http"
)

// Plugin is the contract every gateway plugin must satisfy.
//
// Versioning is by Go module tag — no Version() method on the interface per
// ADR PI2-yaa-0001 §1. If the interface ever needs a non-additive change, the
// module path bumps to gateway/plugin/v1.
type Plugin interface {
	// Name returns the unique, stable identifier for this plugin. It is used
	// for deduplication in the registry, for YAML-config block matching, and
	// as a label in structured log lines.
	Name() string

	// Init is called once at gateway startup with the plugin's YAML config
	// block. A non-nil return causes the gateway to exit 1.
	Init(cfg PluginConfig) error

	// Handler wraps next to implement this plugin's middleware behaviour.
	// Declaration order equals execution order (ADR PI2-yaa-0001 §2); the
	// gateway core composes handlers in that order.
	Handler(next http.Handler) http.Handler

	// Shutdown is called in reverse declaration order on SIGTERM
	// (ADR PI2-yaa-0001 §4). ctx carries the graceful-shutdown deadline.
	Shutdown(ctx context.Context) error
}

// PluginConfig is the read accessor passed to [Plugin.Init].
//
// Values are sourced from the plugin's YAML configuration block. Missing keys
// return the zero value for the target type. Type-coercion failures also
// return the zero value and emit a warn-level structured log line.
type PluginConfig interface {
	// GetString returns the value at key as a string, or "" if absent or
	// if the stored value cannot be coerced to string.
	GetString(key string) string

	// GetBool returns the value at key as a bool, or false if absent or
	// if the stored value cannot be coerced to bool.
	GetBool(key string) bool

	// GetStringSlice returns the value at key as []string, or nil if
	// absent or if the stored value cannot be coerced to []string.
	GetStringSlice(key string) []string

	// GetInt returns the value at key as int, or 0 if absent or if the
	// stored value cannot be coerced to int.
	GetInt(key string) int

	// Raw returns the underlying map for advanced or type-specific access.
	// Callers must not mutate the returned map.
	Raw() map[string]any
}
