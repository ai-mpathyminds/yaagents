// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package plugin

import (
	"fmt"
	"log/slog"
)

// MapConfig is the concrete [PluginConfig] implementation backed by a
// map[string]any parsed from the plugin's YAML configuration block.
//
// Construct via [NewMapConfig]; the gateway core builds a MapConfig for each
// plugin listed in the YAML before calling [Plugin.Init].
type MapConfig struct {
	data map[string]any
}

// NewMapConfig constructs a MapConfig from m. A nil map is treated as an
// empty configuration (all accessor calls return zero values).
func NewMapConfig(m map[string]any) *MapConfig {
	if m == nil {
		m = map[string]any{}
	}
	return &MapConfig{data: m}
}

// GetString returns the value at key as a string.
// Missing key → "". Stored value not a string → "" and a warn log.
func (c *MapConfig) GetString(key string) string {
	v, ok := c.data[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		slog.Warn("plugin config: type coercion failed",
			slog.String("key", key),
			slog.String("want", "string"),
			slog.String("got", fmt.Sprintf("%T", v)))
		return ""
	}
	return s
}

// GetBool returns the value at key as a bool.
// Missing key → false. Stored value not a bool → false and a warn log.
func (c *MapConfig) GetBool(key string) bool {
	v, ok := c.data[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		slog.Warn("plugin config: type coercion failed",
			slog.String("key", key),
			slog.String("want", "bool"),
			slog.String("got", fmt.Sprintf("%T", v)))
		return false
	}
	return b
}

// GetStringSlice returns the value at key as []string.
// Missing key → nil. Accepted stored types: []string or []any (each element
// must be a string). Any other type, or a []any with a non-string element,
// returns nil and emits a warn log.
func (c *MapConfig) GetStringSlice(key string) []string {
	v, ok := c.data[key]
	if !ok {
		return nil
	}
	switch typed := v.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for i, elem := range typed {
			s, ok := elem.(string)
			if !ok {
				slog.Warn("plugin config: type coercion failed",
					slog.String("key", key),
					slog.Int("index", i),
					slog.String("want", "string"),
					slog.String("got", fmt.Sprintf("%T", elem)))
				return nil
			}
			out = append(out, s)
		}
		return out
	default:
		slog.Warn("plugin config: type coercion failed",
			slog.String("key", key),
			slog.String("want", "[]string"),
			slog.String("got", fmt.Sprintf("%T", v)))
		return nil
	}
}

// GetInt returns the value at key as int.
// Missing key → 0. Accepted stored types: int, int64, float64 (YAML/JSON
// decoders may produce any of these for integer literals). Any other type
// returns 0 and emits a warn log.
func (c *MapConfig) GetInt(key string) int {
	v, ok := c.data[key]
	if !ok {
		return 0
	}
	switch typed := v.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		slog.Warn("plugin config: type coercion failed",
			slog.String("key", key),
			slog.String("want", "int"),
			slog.String("got", fmt.Sprintf("%T", v)))
		return 0
	}
}

// Raw returns a shallow copy of the underlying configuration map for advanced
// or type-specific access. Callers must not mutate map values.
func (c *MapConfig) Raw() map[string]any {
	out := make(map[string]any, len(c.data))
	for k, v := range c.data {
		out[k] = v
	}
	return out
}
