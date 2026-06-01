// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Internal tests — package loader — so we can reach unexported helpers
// (readEntries, assertTokenValidatorAlwaysOn, mergeTokenValidatorEnv, entry).
package loader

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

// ── recorder: reusable test Plugin ────────────────────────────────────────────

// recorder captures Init calls and can record Handler + Shutdown invocation order.
type recorder struct {
	name      string
	initErr   error
	gotCfg    plugin.PluginConfig
	handOrder *[]string // set per-test; nil = don't record
	shutOrder *[]string // set per-test; nil = don't record
}

func (r *recorder) Name() string { return r.name }
func (r *recorder) Init(cfg plugin.PluginConfig) error {
	r.gotCfg = cfg
	return r.initErr
}
func (r *recorder) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.handOrder != nil {
			*r.handOrder = append(*r.handOrder, r.name)
		}
		next.ServeHTTP(w, req)
	})
}
func (r *recorder) Shutdown(_ context.Context) error {
	if r.shutOrder != nil {
		*r.shutOrder = append(*r.shutOrder, r.name)
	}
	return nil
}

// Package-level recorder singletons (registered once in TestMain).
var (
	recA  = &recorder{name: "test-plg-a"}
	recB  = &recorder{name: "test-plg-b"}
	recC  = &recorder{name: "test-plg-c"}
	recTV = &recorder{name: tokenValidatorName}
)

// TestMain registers test plugins and runs the suite.
func TestMain(m *testing.M) {
	plugin.Register(recA)
	plugin.Register(recB)
	plugin.Register(recC)
	plugin.Register(recTV)
	os.Exit(m.Run())
}

// reset clears per-test state on the shared recorder singletons.
func reset() {
	for _, r := range []*recorder{recA, recB, recC, recTV} {
		r.gotCfg = nil
		r.initErr = nil
		r.handOrder = nil
		r.shutOrder = nil
	}
}

func noopLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// writeYAML writes content to a temp file and returns its path.
func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

// ── readEntries ────────────────────────────────────────────────────────────────

func TestReadEntries_HappyPath(t *testing.T) {
	path := writeYAML(t, `
plugins:
  - name: test-plg-a
    enabled: true
  - name: test-plg-b
    timeout: 30
`)
	entries, err := readEntries("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].name != "test-plg-a" || entries[1].name != "test-plg-b" {
		t.Errorf("names: %v", []string{entries[0].name, entries[1].name})
	}
	// 'name' must not appear in the config map.
	if _, ok := entries[0].cfg["name"]; ok {
		t.Error("'name' key must not appear in the config map")
	}
	if entries[0].cfg["enabled"] != true {
		t.Errorf("enabled not parsed for test-plg-a")
	}
}

func TestReadEntries_StandalonePluginsFile(t *testing.T) {
	pluginsPath := writeYAML(t, `
plugins:
  - name: test-plg-a
`)
	routesPath := writeYAML(t, `
routes:
  - id: r1
    method: GET
    path: /foo
    target: http://localhost:8080
`)
	entries, err := readEntries(pluginsPath, routesPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 || entries[0].name != "test-plg-a" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestReadEntries_NoPluginsBlock(t *testing.T) {
	path := writeYAML(t, `
routes:
  - id: r1
    method: GET
    path: /foo
    target: http://localhost:8080
`)
	entries, err := readEntries("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestReadEntries_MissingFile(t *testing.T) {
	_, err := readEntries("", t.TempDir()+"/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadEntries_InvalidYAML(t *testing.T) {
	path := writeYAML(t, `{: not valid yaml]`)
	_, err := readEntries("", path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestReadEntries_MissingNameField(t *testing.T) {
	path := writeYAML(t, `
plugins:
  - enabled: true
`)
	_, err := readEntries("", path)
	if err == nil {
		t.Fatal("expected error for entry without 'name'")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name': %v", err)
	}
}

func TestReadEntries_EmptyPluginsList(t *testing.T) {
	path := writeYAML(t, `plugins: []`)
	entries, err := readEntries("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

// ── assertTokenValidatorAlwaysOn ──────────────────────────────────────────────

func TestAssertTokenValidatorAlwaysOn_EnabledFalse_Error(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{"enabled": false}}}
	err := assertTokenValidatorAlwaysOn(entries)
	if err == nil {
		t.Fatal("expected error when token-validator.enabled is false")
	}
	if !strings.Contains(err.Error(), "token-validator cannot be disabled") {
		t.Errorf("error must contain 'token-validator cannot be disabled': %v", err)
	}
}

func TestAssertTokenValidatorAlwaysOn_EnabledTrue_NoError(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{"enabled": true}}}
	if err := assertTokenValidatorAlwaysOn(entries); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAssertTokenValidatorAlwaysOn_NoEnabledKey_NoError(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{}}}
	if err := assertTokenValidatorAlwaysOn(entries); err != nil {
		t.Errorf("unexpected error for absent 'enabled' key: %v", err)
	}
}

func TestAssertTokenValidatorAlwaysOn_Absent_NoError(t *testing.T) {
	entries := []entry{{name: "test-plg-a", cfg: map[string]any{}}}
	if err := assertTokenValidatorAlwaysOn(entries); err != nil {
		t.Errorf("unexpected error when token-validator absent: %v", err)
	}
}

func TestAssertTokenValidatorAlwaysOn_EmptyEntries_NoError(t *testing.T) {
	if err := assertTokenValidatorAlwaysOn(nil); err != nil {
		t.Errorf("unexpected error for empty entries: %v", err)
	}
}

// ── mergeTokenValidatorEnv ────────────────────────────────────────────────────

func TestMergeTokenValidatorEnv_MergesSecret(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{}}}
	mergeTokenValidatorEnv(entries, "env-secret", "")
	if entries[0].cfg["jwt_secret"] != "env-secret" {
		t.Errorf("jwt_secret not merged: %v", entries[0].cfg)
	}
}

func TestMergeTokenValidatorEnv_MergesJWKSURL(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{}}}
	mergeTokenValidatorEnv(entries, "", "https://example.com/jwks")
	if entries[0].cfg["jwks_url"] != "https://example.com/jwks" {
		t.Errorf("jwks_url not merged: %v", entries[0].cfg)
	}
}

func TestMergeTokenValidatorEnv_DoesNotOverrideYAMLSecret(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{"jwt_secret": "yaml-secret"}}}
	mergeTokenValidatorEnv(entries, "env-secret", "")
	if entries[0].cfg["jwt_secret"] != "yaml-secret" {
		t.Error("YAML jwt_secret must not be overridden by env var")
	}
}

func TestMergeTokenValidatorEnv_DoesNotOverrideYAMLJWKS(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{"jwks_url": "https://yaml.example.com/jwks"}}}
	mergeTokenValidatorEnv(entries, "", "https://env.example.com/jwks")
	if entries[0].cfg["jwks_url"] != "https://yaml.example.com/jwks" {
		t.Error("YAML jwks_url must not be overridden by env var")
	}
}

func TestMergeTokenValidatorEnv_NoopWhenTVAbsent(t *testing.T) {
	entries := []entry{{name: "test-plg-a", cfg: map[string]any{}}}
	// Must not panic; entry is unmodified.
	mergeTokenValidatorEnv(entries, "env-secret", "https://example.com/jwks")
	if len(entries[0].cfg) != 0 {
		t.Error("non-token-validator entry must not be modified")
	}
}

func TestMergeTokenValidatorEnv_EmptyEnvVars_NoMerge(t *testing.T) {
	entries := []entry{{name: tokenValidatorName, cfg: map[string]any{}}}
	mergeTokenValidatorEnv(entries, "", "")
	if _, ok := entries[0].cfg["jwt_secret"]; ok {
		t.Error("empty env var must not be merged as jwt_secret")
	}
	if _, ok := entries[0].cfg["jwks_url"]; ok {
		t.Error("empty env var must not be merged as jwks_url")
	}
}

// ── Load ──────────────────────────────────────────────────────────────────────

func TestLoad_HappyPath(t *testing.T) {
	reset()
	path := writeYAML(t, `
plugins:
  - name: test-plg-a
    key: value
`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ldr == nil {
		t.Fatal("got nil Loader")
	}
	if recA.gotCfg == nil {
		t.Error("test-plg-a Init must be called")
	}
	if recA.gotCfg.GetString("key") != "value" {
		t.Errorf("config not passed: got %q", recA.gotCfg.GetString("key"))
	}
}

func TestLoad_TokenValidatorDisabled_Error(t *testing.T) {
	reset()
	path := writeYAML(t, `
plugins:
  - name: token-validator
    enabled: false
`)
	_, err := Load(noopLog(), "", path, "", "")
	if err == nil {
		t.Fatal("expected error when token-validator.enabled is false")
	}
	if !strings.Contains(err.Error(), "token-validator cannot be disabled") {
		t.Errorf("error must contain 'token-validator cannot be disabled': %v", err)
	}
}

func TestLoad_PluginInitError_ContainsPluginName(t *testing.T) {
	reset()
	recA.initErr = errors.New("bad config")
	defer func() { recA.initErr = nil }()

	path := writeYAML(t, `
plugins:
  - name: test-plg-a
`)
	_, err := Load(noopLog(), "", path, "", "")
	if err == nil {
		t.Fatal("expected error when plugin Init fails")
	}
	if !strings.Contains(err.Error(), "test-plg-a") {
		t.Errorf("error must contain plugin name: %v", err)
	}
}

func TestLoad_UnregisteredPlugin_Skipped(t *testing.T) {
	reset()
	path := writeYAML(t, `
plugins:
  - name: no-such-plugin
  - name: test-plg-a
`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recA.gotCfg == nil {
		t.Error("test-plg-a must be initialized even when a prior plugin is unregistered")
	}
	_ = ldr
}

func TestLoad_EnvVarSecret_MergedIntoTokenValidator(t *testing.T) {
	reset()
	path := writeYAML(t, `
plugins:
  - name: token-validator
    enabled: true
`)
	_, err := Load(noopLog(), "", path, "injected-secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recTV.gotCfg == nil {
		t.Fatal("token-validator Init not called")
	}
	if recTV.gotCfg.GetString("jwt_secret") != "injected-secret" {
		t.Errorf("jwt_secret not injected: got %q", recTV.gotCfg.GetString("jwt_secret"))
	}
}

func TestLoad_EnvVarJWKS_MergedIntoTokenValidator(t *testing.T) {
	reset()
	path := writeYAML(t, `
plugins:
  - name: token-validator
    enabled: true
`)
	_, err := Load(noopLog(), "", path, "", "https://idp.example.com/jwks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recTV.gotCfg.GetString("jwks_url") != "https://idp.example.com/jwks" {
		t.Errorf("jwks_url not injected: got %q", recTV.gotCfg.GetString("jwks_url"))
	}
}

func TestLoad_NoPluginsBlock_EmptyLoader(t *testing.T) {
	reset()
	path := writeYAML(t, `routes: []`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ldr == nil {
		t.Fatal("got nil Loader for empty plugins")
	}
}

func TestLoad_StandalonePluginsFile(t *testing.T) {
	reset()
	pluginsPath := writeYAML(t, `
plugins:
  - name: test-plg-b
`)
	routesPath := writeYAML(t, `routes: []`)
	_, err := Load(noopLog(), pluginsPath, routesPath, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recB.gotCfg == nil {
		t.Error("test-plg-b must be initialized when using standalone plugins file")
	}
}

// ── Chain ─────────────────────────────────────────────────────────────────────

func TestChain_DeclarationOrder(t *testing.T) {
	reset()
	var order []string
	recA.handOrder = &order
	recB.handOrder = &order
	defer func() {
		recA.handOrder = nil
		recB.handOrder = nil
	}()

	path := writeYAML(t, `
plugins:
  - name: test-plg-a
  - name: test-plg-b
`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	innerCalled := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	})
	h := ldr.Chain(inner)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if !innerCalled {
		t.Error("inner handler not reached")
	}
	if len(order) != 2 || order[0] != "test-plg-a" || order[1] != "test-plg-b" {
		t.Errorf("handler order: got %v, want [test-plg-a test-plg-b]", order)
	}
}

func TestChain_SinglePlugin(t *testing.T) {
	reset()
	var order []string
	recC.handOrder = &order
	defer func() { recC.handOrder = nil }()

	path := writeYAML(t, `
plugins:
  - name: test-plg-c
`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	ldr.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if len(order) != 1 || order[0] != "test-plg-c" {
		t.Errorf("handler order: got %v, want [test-plg-c]", order)
	}
}

func TestChain_EmptyLoader_PassesThrough(t *testing.T) {
	path := writeYAML(t, `routes: []`)
	ldr, err := Load(noopLog(), "", path, "", "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	ldr.Chain(inner).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !called {
		t.Error("inner handler not called for empty chain")
	}
}

// ── Shutdown ──────────────────────────────────────────────────────────────────

func TestShutdown_ReverseOrder(t *testing.T) {
	var shutOrder []string
	s1 := &recorder{name: "shut-1", shutOrder: &shutOrder}
	s2 := &recorder{name: "shut-2", shutOrder: &shutOrder}
	s3 := &recorder{name: "shut-3", shutOrder: &shutOrder}

	ldr := &Loader{
		log:     noopLog(),
		ordered: []plugin.Plugin{s1, s2, s3},
	}
	ldr.Shutdown(context.Background())

	want := []string{"shut-3", "shut-2", "shut-1"}
	if len(shutOrder) != len(want) {
		t.Fatalf("shutdown called %d times, want %d", len(shutOrder), len(want))
	}
	for i, n := range want {
		if shutOrder[i] != n {
			t.Errorf("shutdown[%d]: got %q, want %q", i, shutOrder[i], n)
		}
	}
}

func TestShutdown_ContinuesPastErrors(t *testing.T) {
	// Both plugins must be shut down even if they return errors.
	var shutOrder []string
	ep := &recorder{name: "err-plugin", shutOrder: &shutOrder}
	good := &recorder{name: "good-plugin", shutOrder: &shutOrder}

	ldr := &Loader{
		log:     noopLog(),
		ordered: []plugin.Plugin{ep, good},
	}
	ldr.Shutdown(context.Background())
	if len(shutOrder) != 2 {
		t.Errorf("both plugins must be shut down; got %d calls", len(shutOrder))
	}
}

func TestShutdown_Empty_NoPanic(t *testing.T) {
	ldr := &Loader{log: noopLog(), ordered: nil}
	ldr.Shutdown(context.Background()) // must not panic
}
