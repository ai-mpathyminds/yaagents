// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Internal tests — package plugin (not plugin_test) so tests can reset global
// registry state between subtests.
package plugin

import (
	"context"
	"net/http"
	"testing"
)

// ── registry helpers ───────────────────────────────────────────────────────────

// stub is a minimal Plugin used to exercise the registry.
type stub struct{ name string }

func (s *stub) Name() string                           { return s.name }
func (s *stub) Init(_ PluginConfig) error              { return nil }
func (s *stub) Handler(next http.Handler) http.Handler { return next }
func (s *stub) Shutdown(_ context.Context) error       { return nil }

// resetRegistry clears the global registry for the lifetime of t.
// It also registers a t.Cleanup to restore the empty state after the test.
func resetRegistry(t *testing.T) {
	t.Helper()
	mu.Lock()
	registry = nil
	byName = map[string]struct{}{}
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		registry = nil
		byName = map[string]struct{}{}
		mu.Unlock()
	})
}

// ── registry tests ─────────────────────────────────────────────────────────────

// TestRegister_DuplicateName verifies that a second Register call with the
// same Name() is a no-op: the duplicate is skipped (warn-logged) and the
// registered slice stays stable with exactly one entry.
func TestRegister_DuplicateName(t *testing.T) {
	resetRegistry(t)

	a1 := &stub{name: "alpha"}
	a2 := &stub{name: "alpha"} // duplicate name

	Register(a1)
	snapAfterFirst := Registered()

	Register(a2) // must be a no-op
	snapAfterSecond := Registered()

	if len(snapAfterFirst) != 1 {
		t.Fatalf("after first Register: got %d plugin(s), want 1", len(snapAfterFirst))
	}
	if len(snapAfterSecond) != 1 {
		t.Fatalf("after duplicate Register: got %d plugin(s), want 1 (duplicate must be skipped)",
			len(snapAfterSecond))
	}
	if snapAfterSecond[0] != a1 {
		t.Error("registered plugin should be the first-registered instance")
	}
}

// TestRegister_Order verifies that Registered() preserves registration order.
func TestRegister_Order(t *testing.T) {
	resetRegistry(t)

	names := []string{"first", "second", "third"}
	for _, n := range names {
		Register(&stub{name: n})
	}

	got := Registered()
	if len(got) != len(names) {
		t.Fatalf("got %d plugins, want %d", len(got), len(names))
	}
	for i, p := range got {
		if p.Name() != names[i] {
			t.Errorf("position %d: got %q, want %q", i, p.Name(), names[i])
		}
	}
}

// TestRegistered_EmptyRegistry verifies that Registered() returns an empty
// (non-nil) slice when nothing has been registered.
func TestRegistered_EmptyRegistry(t *testing.T) {
	resetRegistry(t)

	got := Registered()
	if got == nil {
		t.Fatal("Registered() must return non-nil slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty registry, got %d entries", len(got))
	}
}

// TestRegistered_ReturnsCopy verifies that mutating the returned slice does
// not affect the canonical registry.
func TestRegistered_ReturnsCopy(t *testing.T) {
	resetRegistry(t)

	Register(&stub{name: "x"})
	snap := Registered()
	snap[0] = &stub{name: "mutated"}

	canonical := Registered()
	if canonical[0].Name() != "x" {
		t.Error("mutating returned slice must not affect the registry")
	}
}

// ── MapConfig / PluginConfig tests ─────────────────────────────────────────────

// TestMapConfig_GetString_Hit exercises the happy path.
func TestMapConfig_GetString_Hit(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"url": "https://example.com"})
	if got := cfg.GetString("url"); got != "https://example.com" {
		t.Errorf("GetString: got %q, want %q", got, "https://example.com")
	}
}

// TestMapConfig_GetString_Missing verifies zero-value on absent key.
func TestMapConfig_GetString_Missing(t *testing.T) {
	cfg := NewMapConfig(map[string]any{})
	if got := cfg.GetString("missing"); got != "" {
		t.Errorf("GetString(missing): got %q, want empty string", got)
	}
}

// TestMapConfig_GetString_WrongType verifies coercion failure returns "" + does not panic.
func TestMapConfig_GetString_WrongType(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"count": 42})
	if got := cfg.GetString("count"); got != "" {
		t.Errorf("GetString(wrong-type): got %q, want empty string", got)
	}
}

// TestMapConfig_GetBool_Hit exercises the happy path.
func TestMapConfig_GetBool_Hit(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"enabled": true})
	if got := cfg.GetBool("enabled"); !got {
		t.Error("GetBool(enabled): got false, want true")
	}
}

// TestMapConfig_GetBool_Missing verifies zero-value on absent key.
func TestMapConfig_GetBool_Missing(t *testing.T) {
	cfg := NewMapConfig(map[string]any{})
	if got := cfg.GetBool("missing"); got {
		t.Error("GetBool(missing): got true, want false")
	}
}

// TestMapConfig_GetBool_WrongType verifies coercion failure returns false + does not panic.
func TestMapConfig_GetBool_WrongType(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"enabled": "yes"})
	if got := cfg.GetBool("enabled"); got {
		t.Error("GetBool(wrong-type): got true, want false")
	}
}

// TestMapConfig_GetStringSlice_StringSlice verifies direct []string value.
func TestMapConfig_GetStringSlice_StringSlice(t *testing.T) {
	want := []string{"a", "b", "c"}
	cfg := NewMapConfig(map[string]any{"roles": want})
	got := cfg.GetStringSlice("roles")
	if len(got) != len(want) {
		t.Fatalf("GetStringSlice: len %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetStringSlice[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

// TestMapConfig_GetStringSlice_AnySlice verifies []any→[]string coercion.
func TestMapConfig_GetStringSlice_AnySlice(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"tags": []any{"x", "y"}})
	got := cfg.GetStringSlice("tags")
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Errorf("GetStringSlice([]any): got %v, want [x y]", got)
	}
}

// TestMapConfig_GetStringSlice_AnySliceWrongElem verifies []any with a
// non-string element returns nil.
func TestMapConfig_GetStringSlice_AnySliceWrongElem(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"items": []any{"ok", 99}})
	if got := cfg.GetStringSlice("items"); got != nil {
		t.Errorf("GetStringSlice(bad elem): got %v, want nil", got)
	}
}

// TestMapConfig_GetStringSlice_WrongType verifies non-slice type returns nil.
func TestMapConfig_GetStringSlice_WrongType(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"items": "not-a-slice"})
	if got := cfg.GetStringSlice("items"); got != nil {
		t.Errorf("GetStringSlice(wrong type): got %v, want nil", got)
	}
}

// TestMapConfig_GetStringSlice_Missing verifies nil on absent key.
func TestMapConfig_GetStringSlice_Missing(t *testing.T) {
	cfg := NewMapConfig(map[string]any{})
	if got := cfg.GetStringSlice("missing"); got != nil {
		t.Errorf("GetStringSlice(missing): got %v, want nil", got)
	}
}

// TestMapConfig_GetInt_Int exercises the int happy path.
func TestMapConfig_GetInt_Int(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"ttl": 600})
	if got := cfg.GetInt("ttl"); got != 600 {
		t.Errorf("GetInt(int): got %d, want 600", got)
	}
}

// TestMapConfig_GetInt_Int64 exercises int64→int coercion.
func TestMapConfig_GetInt_Int64(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"ttl": int64(300)})
	if got := cfg.GetInt("ttl"); got != 300 {
		t.Errorf("GetInt(int64): got %d, want 300", got)
	}
}

// TestMapConfig_GetInt_Float64 exercises float64→int coercion (JSON/YAML numbers).
func TestMapConfig_GetInt_Float64(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"ttl": float64(120)})
	if got := cfg.GetInt("ttl"); got != 120 {
		t.Errorf("GetInt(float64): got %d, want 120", got)
	}
}

// TestMapConfig_GetInt_Missing verifies zero-value on absent key.
func TestMapConfig_GetInt_Missing(t *testing.T) {
	cfg := NewMapConfig(map[string]any{})
	if got := cfg.GetInt("missing"); got != 0 {
		t.Errorf("GetInt(missing): got %d, want 0", got)
	}
}

// TestMapConfig_GetInt_WrongType verifies coercion failure returns 0.
func TestMapConfig_GetInt_WrongType(t *testing.T) {
	cfg := NewMapConfig(map[string]any{"ttl": "fast"})
	if got := cfg.GetInt("ttl"); got != 0 {
		t.Errorf("GetInt(wrong type): got %d, want 0", got)
	}
}

// TestMapConfig_Raw verifies Raw returns a map with all keys and a new map
// instance (mutations do not propagate back).
func TestMapConfig_Raw(t *testing.T) {
	orig := map[string]any{"a": "1", "b": true}
	cfg := NewMapConfig(orig)
	raw := cfg.Raw()

	if len(raw) != 2 {
		t.Fatalf("Raw len: got %d, want 2", len(raw))
	}
	if raw["a"] != "1" || raw["b"] != true {
		t.Errorf("Raw contents mismatch: %v", raw)
	}
	// Mutation of raw must not affect cfg.
	raw["a"] = "mutated"
	if cfg.GetString("a") != "1" {
		t.Error("Raw mutation propagated back to MapConfig")
	}
}

// TestNewMapConfig_Nil verifies that a nil map is treated as empty config.
func TestNewMapConfig_Nil(t *testing.T) {
	cfg := NewMapConfig(nil)
	if got := cfg.GetString("anything"); got != "" {
		t.Errorf("nil-map config GetString: got %q, want empty", got)
	}
	if raw := cfg.Raw(); len(raw) != 0 {
		t.Errorf("nil-map config Raw: got %v, want empty map", raw)
	}
}
