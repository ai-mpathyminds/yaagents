// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

// Package licensecheck implements the license-check plugin (PRD §6.5 plugin c).
//
// The plugin reads a product-license token from a configured request header,
// verifies it against a remote license server URL, and caches results with a
// bounded LRU cache.  Requests with an invalid or unverifiable token receive a
// 403 application/vnd.yaagents.error+json response; next is not called.
//
// On network errors (including HTTP client timeout) the 403 trace includes
// dependency: "license-server" so the caller can distinguish a service failure
// from a policy rejection.
//
// Configuration keys:
//
//	license_url:         required — absolute http/https URL of the license server
//	header:              request header carrying the token (default: X-License-Token)
//	cache_ttl_seconds:   how long to cache a successful HTTP response (default: 300)
//	max_cache_size:      maximum number of cached entries — LRU eviction (default: 1024)
//	timeout_seconds:     HTTP client hard timeout per attempt (default: 5)
//
// Registration: init() calls plugin.Register so the gateway wires this plugin
// by import side-effect (ADR PI2-yaa-0001 §3; no plugin.Open / dlopen).
package licensecheck

import (
	"container/list"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ai-mpathyminds/yaagents/gateway/internal/reqctx"
	"github.com/ai-mpathyminds/yaagents/gateway/internal/response"
	"github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

func init() {
	plugin.Register(&LicenseCheck{})
}

// cacheEntry is one entry in the LRU cache.
type cacheEntry struct {
	token    string
	valid    bool
	expireAt time.Time
}

// LicenseCheck is the license-check plugin.
// Zero value is invalid; always call Init before Handler.
type LicenseCheck struct {
	header     string
	licenseURL string
	cacheTTL   time.Duration
	maxSize    int
	httpClient *http.Client

	mu      sync.Mutex
	lruList *list.List
	lruMap  map[string]*list.Element // token → *cacheEntry element
}

// Name returns the canonical plugin identifier.
func (lc *LicenseCheck) Name() string { return "license-check" }

// Init validates configuration and initialises the plugin.
//
// Returns a non-nil error (gateway exit 1) when:
//   - license_url is absent or not a valid absolute http/https URL
func (lc *LicenseCheck) Init(cfg plugin.PluginConfig) error {
	licURL := cfg.GetString("license_url")
	if licURL == "" {
		return fmt.Errorf("license-check: license_url must be configured")
	}
	u, err := url.Parse(licURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("license-check: license_url %q must be a valid absolute URL (http or https)", licURL)
	}

	hdr := cfg.GetString("header")
	if hdr == "" {
		hdr = "X-License-Token"
	}

	ttlSecs := cfg.GetInt("cache_ttl_seconds")
	if ttlSecs <= 0 {
		ttlSecs = 300
	}

	maxSize := cfg.GetInt("max_cache_size")
	if maxSize <= 0 {
		maxSize = 1024
	}

	timeoutSecs := cfg.GetInt("timeout_seconds")
	if timeoutSecs <= 0 {
		timeoutSecs = 5
	}

	lc.header = hdr
	lc.licenseURL = licURL
	lc.cacheTTL = time.Duration(ttlSecs) * time.Second
	lc.maxSize = maxSize
	lc.httpClient = &http.Client{Timeout: time.Duration(timeoutSecs) * time.Second}
	lc.lruList = list.New()
	lc.lruMap = make(map[string]*list.Element)

	return nil
}

// Handler returns an http.Handler that enforces license token validity.
//
// Execution order:
//  1. Read token from r.Header.Get(lc.header).
//  2. Consult LRU cache; on hit serve from cache.
//  3. On cache miss, call lc.verify against the license server.
//  4. Cache the result when the server returned an HTTP response (valid or not).
//     Network errors are not cached so the next request retries.
//  5. On invalid / network-error → 403 vendor-error body (next not called).
//  6. On valid → call next.
func (lc *LicenseCheck) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(lc.header)
		corrID := reqctx.CorrelationID(r.Context())
		reqID := reqctx.RequestID(r.Context())

		// Cache lookup.
		if valid, hit := lc.cacheGet(token); hit {
			if valid {
				next.ServeHTTP(w, r)
				return
			}
			response.WriteError(w, http.StatusForbidden, response.ErrorBody{
				Type:    "forbidden",
				Code:    "license_invalid",
				Message: "license token is not valid",
				Trace:   response.Trace{CorrelationID: corrID, RequestID: reqID},
			})
			return
		}

		// Verify against the license server.
		valid, dep := lc.verify(r.Context(), token)
		if valid {
			lc.cacheSet(token, true)
			next.ServeHTTP(w, r)
			return
		}

		if dep == "" {
			// HTTP non-2xx response — cache the rejected result.
			lc.cacheSet(token, false)
		}
		// Network/timeout error (dep == "license-server") — do not cache.

		response.WriteError(w, http.StatusForbidden, response.ErrorBody{
			Type:    "forbidden",
			Code:    "license_invalid",
			Message: "license verification failed",
			Trace: response.Trace{
				CorrelationID: corrID,
				RequestID:     reqID,
				Dependency:    dep,
			},
		})
	})
}

// Shutdown is a no-op; the plugin holds no background goroutines.
func (lc *LicenseCheck) Shutdown(_ context.Context) error { return nil }

// verify calls the license server with up to 2 attempts for network errors.
//
// Returns (true, "") on 2xx.
// Returns (false, "") on HTTP non-2xx (no retry; server answered).
// Returns (false, "license-server") when all attempts fail with a network error.
func (lc *LicenseCheck) verify(ctx context.Context, token string) (valid bool, dependency string) {
	const maxAttempts = 2

	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, lc.licenseURL, nil)
		if err != nil {
			// Malformed URL after Init (should never happen); treat as dependency failure.
			return false, "license-server"
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := lc.httpClient.Do(req)
		if err != nil {
			slog.Warn("license-check: network error contacting license server",
				slog.Int("attempt", attempt+1),
				slog.String("error", err.Error()))
			continue // retry if budget remains
		}
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return true, ""
		}
		// HTTP error response — don't retry.
		slog.Info("license-check: license server rejected token",
			slog.Int("status", resp.StatusCode))
		return false, ""
	}

	// Exhausted retry budget.
	return false, "license-server"
}

// cacheGet returns (valid, true) on a live cache hit, (false, false) on miss or expiry.
func (lc *LicenseCheck) cacheGet(token string) (valid bool, hit bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	elem, ok := lc.lruMap[token]
	if !ok {
		return false, false
	}

	entry := elem.Value.(*cacheEntry)
	if time.Now().After(entry.expireAt) {
		lc.lruList.Remove(elem)
		delete(lc.lruMap, token)
		return false, false
	}

	lc.lruList.MoveToFront(elem)
	return entry.valid, true
}

// cacheSet adds or updates a cache entry and evicts the LRU entry when full.
func (lc *LicenseCheck) cacheSet(token string, valid bool) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if elem, ok := lc.lruMap[token]; ok {
		e := elem.Value.(*cacheEntry)
		e.valid = valid
		e.expireAt = time.Now().Add(lc.cacheTTL)
		lc.lruList.MoveToFront(elem)
		return
	}

	// Evict LRU tail when at capacity.
	if lc.lruList.Len() >= lc.maxSize {
		tail := lc.lruList.Back()
		if tail != nil {
			lc.lruList.Remove(tail)
			delete(lc.lruMap, tail.Value.(*cacheEntry).token)
		}
	}

	entry := &cacheEntry{
		token:    token,
		valid:    valid,
		expireAt: time.Now().Add(lc.cacheTTL),
	}
	elem := lc.lruList.PushFront(entry)
	lc.lruMap[token] = elem
}
