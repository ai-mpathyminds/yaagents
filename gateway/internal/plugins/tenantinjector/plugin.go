// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package tenantinjector implements the tenant-injector plugin v2 (WI-2yaa.PLG-4b).
//
// # Design (ADR PI2-yaa-0006 Decision 1)
//
// Tenant identity is derived from a validated JWT claim plus an HTTP lookup
// against a tenant-directory service.  The client cannot influence the injected
// tenant ID (anti-smuggling: inbound inject.tenant_header is stripped before
// injection).
//
// Per-request flow (executes after token-validator PLG-3):
//  1. Strip any inbound inject.tenant_header from r.Header (anti-smuggling).
//  2. Read principal from validated JWT claims (reqctx.JWTClaims, populated by PLG-3).
//  3. Check per-principal LRU+TTL cache.
//  4. On cache miss: call lookup.url via HTTP (singleflight-coalesced).
//  5. Cache positive/negative result.
//  6. Apply optional post-derivation allowlist.
//  7. Inject derived tenant ID via inject.tenant_header (+ optional principal_header).
//
// Failure codes are fully configurable via on_failure.* (defaults 503/503/403/401).
// Boot fail-open: gateway starts even when lookup.url is unreachable; per-request
// returns on_failure.lookup_network_error (503) from cold-start onward.
//
// Registration: init() → plugin.Register(&TenantInjector{}) per ADR PI2-yaa-0001 §3.
package tenantinjector

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&TenantInjector{})
}

// Sentinel errors for the lookup result type.
var (
	errPrincipalNotFound = errors.New("principal not found (404)")
	errLookupTimeout     = errors.New("lookup timeout")
	errLookupNetwork     = errors.New("lookup network error")
)

// failureCodes holds the configured HTTP status codes for each failure mode.
type failureCodes struct {
	lookupNetworkError int
	lookupTimeout      int
	principalNotFound  int
	claimMissing       int
}

// cacheEntry is one slot in the per-principal LRU cache.
type cacheEntry struct {
	key       string
	tenantID  string // "" for negative entries
	negative  bool
	expiresAt time.Time
}

// TenantInjector is the tenant-injector plugin v2.
// Zero value is invalid; always call Init before Handler.
type TenantInjector struct {
	// Validated config fields
	principalClaim      string
	lookupURL           string // contains {principal} placeholder
	lookupMethod        string
	lookupHeaders       map[string]string
	bearerToken         string // empty when auth.mode != bearer
	tenantIDField       string
	injectTenantHeader  string
	injectPrincipalHdr  string // "" disables principal injection
	allowlist           map[string]struct{}
	cacheTTL            time.Duration
	negativeCacheTTL    time.Duration
	failures            failureCodes
	httpClient          *http.Client

	// Cache state
	mu      sync.Mutex
	lruList *list.List
	lruMap  map[string]*list.Element
	maxSize int

	// Singleflight deduplicates concurrent first-fetches per principal.
	sfGroup singleflight.Group
}

// Name returns the canonical plugin identifier.
func (ti *TenantInjector) Name() string { return "tenant-injector" }

// Init validates configuration and prepares the plugin for use.
//
// Validation rules (returns non-nil error → gateway exit 1):
//  1. principal.claim non-empty.
//  2. lookup.url parseable AND contains exactly one {principal}.
//  3. lookup.method ∈ {GET, POST}.
//  4. lookup.timeout_ms > 0 AND ≤ 30000.
//  5. lookup.auth.mode ∈ {none, bearer, mtls}; bearer: env non-empty; mtls: files readable.
//  6. lookup.response.mode == "single"; tenant_id_field non-empty.
//  7. lookup.cache.ttl_seconds > 0; max_entries > 0.
//  8. inject.tenant_header non-empty.
//  9. enabled: false → error (defence-in-depth).
func (ti *TenantInjector) Init(cfg plugin.PluginConfig) error {
	if !cfg.GetBool("enabled") {
		return fmt.Errorf("tenant-injector: plugin cannot be disabled " +
			"(defence-in-depth per ADR PI2-yaa-0006)")
	}

	raw := cfg.Raw()

	// Rule 1 — principal.claim
	principalClaim := mapStr(submap(raw, "principal"), "claim")
	if principalClaim == "" {
		return fmt.Errorf("tenant-injector: principal.claim must be non-empty")
	}

	// Rule 2 — lookup.url
	lookupMap := submap(raw, "lookup")
	lookupURL := mapStr(lookupMap, "url")
	if _, err := url.Parse(lookupURL); err != nil || lookupURL == "" {
		return fmt.Errorf("tenant-injector: lookup.url %q is not a valid URL", lookupURL)
	}
	if strings.Count(lookupURL, "{principal}") != 1 {
		return fmt.Errorf("tenant-injector: lookup.url must contain exactly one {principal} placeholder")
	}

	// Rule 3 — lookup.method
	method := mapStr(lookupMap, "method")
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "POST" {
		return fmt.Errorf("tenant-injector: lookup.method must be GET or POST, got %q", method)
	}

	// Rule 4 — lookup.timeout_ms
	timeoutMs := mapInt(lookupMap, "timeout_ms")
	if timeoutMs <= 0 {
		timeoutMs = 500
	}
	if timeoutMs > 30000 {
		return fmt.Errorf("tenant-injector: lookup.timeout_ms %d exceeds maximum 30000", timeoutMs)
	}

	// Rule 5 — lookup.auth
	authMap := submap(lookupMap, "auth")
	authMode := mapStr(authMap, "mode")
	if authMode == "" {
		authMode = "none"
	}
	var bearerToken string
	switch authMode {
	case "none":
		// no-op
	case "bearer":
		envName := mapStr(authMap, "bearer_token_env")
		if envName == "" {
			return fmt.Errorf("tenant-injector: lookup.auth.bearer_token_env must be set when mode is bearer")
		}
		bearerToken = os.Getenv(envName)
		if bearerToken == "" {
			return fmt.Errorf("tenant-injector: env var %s is unset or empty (required for auth.mode: bearer)", envName)
		}
	case "mtls":
		certPath := mapStr(authMap, "client_cert_path")
		keyPath := mapStr(authMap, "client_key_path")
		if certPath == "" || keyPath == "" {
			return fmt.Errorf("tenant-injector: client_cert_path and client_key_path required for auth.mode: mtls")
		}
		if _, err := os.Stat(certPath); err != nil {
			return fmt.Errorf("tenant-injector: client_cert_path %q not readable: %v", certPath, err)
		}
		if _, err := os.Stat(keyPath); err != nil {
			return fmt.Errorf("tenant-injector: client_key_path %q not readable: %v", keyPath, err)
		}
	default:
		return fmt.Errorf("tenant-injector: lookup.auth.mode must be none|bearer|mtls, got %q", authMode)
	}

	// Rule 6 — lookup.response
	responseMap := submap(lookupMap, "response")
	responseMode := mapStr(responseMap, "mode")
	if responseMode == "" {
		responseMode = "single"
	}
	if responseMode != "single" {
		return fmt.Errorf("tenant-injector: lookup.response.mode must be \"single\" in PI2-yaa (got %q); multi-tenant is v0.3+", responseMode)
	}
	tenantIDField := mapStr(responseMap, "tenant_id_field")
	if tenantIDField == "" {
		tenantIDField = "tenant_id"
	}

	// Rule 7 — lookup.cache
	// mapIntPresent distinguishes "key absent → use default" from "key == 0 → error".
	cacheMap := submap(lookupMap, "cache")
	ttlSec, ttlPresent := mapIntPresent(cacheMap, "ttl_seconds")
	if ttlPresent && ttlSec <= 0 {
		return fmt.Errorf("tenant-injector: lookup.cache.ttl_seconds must be > 0")
	}
	if !ttlPresent {
		ttlSec = 300
	}
	negTTLSec := mapIntDefault(cacheMap, "negative_ttl_seconds", 30)
	maxEntries, maxPresent := mapIntPresent(cacheMap, "max_entries")
	if maxPresent && maxEntries <= 0 {
		return fmt.Errorf("tenant-injector: lookup.cache.max_entries must be > 0")
	}
	if !maxPresent {
		maxEntries = 10000
	}

	// Rule 8 — inject.tenant_header
	injectMap := submap(raw, "inject")
	tenantHeader := mapStr(injectMap, "tenant_header")
	if tenantHeader == "" {
		return fmt.Errorf("tenant-injector: inject.tenant_header must be non-empty")
	}
	principalHeader := mapStr(injectMap, "principal_header") // "" disables principal injection

	// on_failure codes (defaults: 503/503/403/401)
	failMap := submap(raw, "on_failure")
	fc := failureCodes{
		lookupNetworkError: mapIntDefault(failMap, "lookup_network_error", 503),
		lookupTimeout:      mapIntDefault(failMap, "lookup_timeout", 503),
		principalNotFound:  mapIntDefault(failMap, "principal_not_found", 403),
		claimMissing:       mapIntDefault(failMap, "claim_missing", 401),
	}

	// allowlist (post-derivation admission gate)
	var allowlist map[string]struct{}
	if sl := cfg.GetStringSlice("allowlist"); len(sl) > 0 {
		allowlist = make(map[string]struct{}, len(sl))
		for _, id := range sl {
			allowlist[id] = struct{}{}
		}
	}

	// extra lookup request headers
	lookupHeaders := mapStrMap(lookupMap, "headers")

	ti.principalClaim = principalClaim
	ti.lookupURL = lookupURL
	ti.lookupMethod = method
	ti.lookupHeaders = lookupHeaders
	ti.bearerToken = bearerToken
	ti.tenantIDField = tenantIDField
	ti.injectTenantHeader = tenantHeader
	ti.injectPrincipalHdr = principalHeader
	ti.allowlist = allowlist
	ti.cacheTTL = time.Duration(ttlSec) * time.Second
	ti.negativeCacheTTL = time.Duration(negTTLSec) * time.Second
	ti.failures = fc
	ti.httpClient = &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	ti.maxSize = maxEntries
	ti.lruList = list.New()
	ti.lruMap = make(map[string]*list.Element, maxEntries)

	return nil
}

// Handler returns the tenant-injection middleware.
func (ti *TenantInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Step 1: anti-smuggling — strip inbound inject.tenant_header unconditionally.
		r.Header.Del(ti.injectTenantHeader)

		// Step 2: read principal from validated JWT claims (populated by PLG-3).
		claims := reqctx.JWTClaims(r.Context())
		if claims == nil {
			ti.writeErr(w, r, ti.failures.claimMissing, "claim_missing",
				"JWT claims not found in context (token-validator must run before tenant-injector)", "")
			return
		}
		principal, _ := claims[ti.principalClaim].(string)
		if principal == "" {
			ti.writeErr(w, r, ti.failures.claimMissing, "claim_missing",
				fmt.Sprintf("JWT claim %q is absent or empty", ti.principalClaim), "")
			return
		}

		// Step 3 & 4: cache lookup → singleflight HTTP lookup on miss.
		tenantID, err := ti.resolveTenant(principal)
		if err != nil {
			switch {
			case errors.Is(err, errPrincipalNotFound):
				slog.Info("tenant-injector: principal not found in directory",
					slog.String("principal", principal))
				ti.writeErr(w, r, ti.failures.principalNotFound, "principal_not_found",
					"principal has no associated tenant", "")
			case errors.Is(err, errLookupTimeout):
				slog.Warn("tenant-injector: lookup timed out",
					slog.String("principal", principal))
				ti.writeErr(w, r, ti.failures.lookupTimeout, "lookup_timeout",
					"tenant directory lookup timed out", "iam-lookup")
			default:
				slog.Warn("tenant-injector: lookup network error",
					slog.String("principal", principal),
					slog.String("error", err.Error()))
				ti.writeErr(w, r, ti.failures.lookupNetworkError, "lookup_network_error",
					"tenant directory lookup failed", "iam-lookup")
			}
			return
		}

		// Step 5: post-derivation allowlist admission gate.
		if len(ti.allowlist) > 0 {
			if _, ok := ti.allowlist[tenantID]; !ok {
				slog.Info("tenant-injector: derived tenant not in allowlist",
					slog.String("tenant_id", tenantID))
				ti.writeErr(w, r, ti.failures.principalNotFound, "tenant_not_in_allowlist",
					"derived tenant is not in the configured allowlist", "")
				return
			}
		}

		// Step 6: inject derived tenant (and optionally principal) headers.
		r.Header.Set(ti.injectTenantHeader, tenantID)
		if ti.injectPrincipalHdr != "" {
			r.Header.Set(ti.injectPrincipalHdr, principal)
		}

		next.ServeHTTP(w, r)
	})
}

// Shutdown is a no-op; the plugin holds no background goroutines.
func (ti *TenantInjector) Shutdown(_ context.Context) error { return nil }

// resolveTenant returns the tenant ID for principal, consulting the cache first
// and falling back to a singleflight-coalesced HTTP lookup.
func (ti *TenantInjector) resolveTenant(principal string) (string, error) {
	// Cache hit path.
	if tid, neg, ok := ti.cacheGet(principal); ok {
		if neg {
			return "", errPrincipalNotFound
		}
		return tid, nil
	}

	// Cache miss — singleflight ensures only one outbound HTTP call per principal
	// even under 50 concurrent first-requests for the same principal.
	type result struct {
		tenantID string
		err      error
	}
	v, callErr, _ := ti.sfGroup.Do(principal, func() (any, error) {
		tid, err := ti.doHTTPLookup(principal)
		// Cache the result immediately so future requests skip singleflight.
		if err == nil {
			ti.cacheSet(principal, tid, false, ti.cacheTTL)
		} else if errors.Is(err, errPrincipalNotFound) {
			ti.cacheSet(principal, "", true, ti.negativeCacheTTL)
		}
		// Network/timeout errors are not cached (next request retries).
		return result{tenantID: tid, err: err}, nil // singleflight error is always nil
	})
	if callErr != nil {
		// Should not happen since the fn above never returns a non-nil error to singleflight.
		return "", errLookupNetwork
	}
	res := v.(result)
	return res.tenantID, res.err
}

// doHTTPLookup calls the tenant-directory endpoint for principal and returns
// the tenant ID (or a sentinel error).
func (ti *TenantInjector) doHTTPLookup(principal string) (string, error) {
	finalURL := strings.ReplaceAll(ti.lookupURL, "{principal}", url.PathEscape(principal))

	ctx, cancel := context.WithTimeout(context.Background(), ti.httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, ti.lookupMethod, finalURL, nil)
	if err != nil {
		return "", errLookupNetwork
	}

	for k, v := range ti.lookupHeaders {
		req.Header.Set(k, v)
	}
	if ti.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+ti.bearerToken)
	}

	resp, err := ti.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", errLookupTimeout
		}
		// Check for timeout via url.Error
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return "", errLookupTimeout
		}
		return "", errLookupNetwork
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errPrincipalNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errLookupNetwork
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", errLookupNetwork
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", errLookupNetwork
	}

	tenantID, _ := payload[ti.tenantIDField].(string)
	if tenantID == "" {
		return "", errPrincipalNotFound
	}
	return tenantID, nil
}

// writeErr writes a vendor-typed error response.
func (ti *TenantInjector) writeErr(w http.ResponseWriter, r *http.Request, status int, code, msg, dep string) {
	corrID := reqctx.CorrelationID(r.Context())
	reqID := reqctx.RequestID(r.Context())
	response.WriteError(w, status, response.ErrorBody{
		Type:    "error",
		Code:    code,
		Message: msg,
		Trace: response.Trace{
			CorrelationID: corrID,
			RequestID:     reqID,
			Dependency:    dep,
		},
	})
}

// ── LRU+TTL cache ─────────────────────────────────────────────────────────────

func (ti *TenantInjector) cacheGet(key string) (tenantID string, negative bool, hit bool) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	el, ok := ti.lruMap[key]
	if !ok {
		return "", false, false
	}
	e := el.Value.(*cacheEntry)
	if time.Now().After(e.expiresAt) {
		ti.lruList.Remove(el)
		delete(ti.lruMap, key)
		return "", false, false
	}
	ti.lruList.MoveToFront(el)
	return e.tenantID, e.negative, true
}

func (ti *TenantInjector) cacheSet(key, tenantID string, negative bool, ttl time.Duration) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	if el, ok := ti.lruMap[key]; ok {
		e := el.Value.(*cacheEntry)
		e.tenantID = tenantID
		e.negative = negative
		e.expiresAt = time.Now().Add(ttl)
		ti.lruList.MoveToFront(el)
		return
	}
	if ti.lruList.Len() >= ti.maxSize {
		if tail := ti.lruList.Back(); tail != nil {
			ti.lruList.Remove(tail)
			delete(ti.lruMap, tail.Value.(*cacheEntry).key)
		}
	}
	entry := &cacheEntry{key: key, tenantID: tenantID, negative: negative,
		expiresAt: time.Now().Add(ttl)}
	el := ti.lruList.PushFront(entry)
	ti.lruMap[key] = el
}

// ── Config helpers ─────────────────────────────────────────────────────────────

func submap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, _ := m[key].(map[string]any)
	return v
}

func mapStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func mapInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

func mapIntDefault(m map[string]any, key string, def int) int {
	v := mapInt(m, key)
	if v == 0 {
		return def
	}
	return v
}

// mapIntPresent returns (value, true) when the key is present in m, else (0, false).
// Unlike mapIntDefault it distinguishes "absent" from "explicitly zero".
func mapIntPresent(m map[string]any, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch tv := v.(type) {
	case int:
		return tv, true
	case int64:
		return int(tv), true
	case float64:
		return int(tv), true
	default:
		return 0, true
	}
}

func mapStrMap(m map[string]any, key string) map[string]string {
	if m == nil {
		return nil
	}
	raw, ok := m[key].(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
