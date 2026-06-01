// Package yaagentsclient is the idiomatic Go client for the YAAgents Agentic
// REST Profile. It is a thin net/http wrapper with zero non-stdlib runtime
// dependencies (design constraint: PRD §5.9).
package yaagentsclient

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Option configures a Client during construction.
type Option func(*clientConfig)

// clientConfig holds all optional settings that the caller may override.
type clientConfig struct {
	token         string
	tenantID      string
	correlationID string // static override; empty → auto UUID v4 per request
	httpClient    *http.Client
}

// Client is the root yaagents client. One instance per target service.
// Obtain via New; do not copy after first use.
type Client struct {
	baseURL string
	cfg     clientConfig
}

// New constructs a Client targeting baseURL. Options are applied in order.
//
// Default behaviour:
//   - HTTP client: 30-second timeout, stdlib DefaultTransport.
//   - X-Correlation-ID: auto-generated UUID v4 per request (crypto/rand).
func New(baseURL string, opts ...Option) *Client {
	cfg := clientConfig{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(&cfg)
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		cfg:     cfg,
	}
}

// WithToken sets the Bearer token injected as "Authorization: Bearer {token}".
func WithToken(token string) Option {
	return func(c *clientConfig) { c.token = token }
}

// WithTenantID sets the value injected as "X-Tenant-ID: {id}".
func WithTenantID(id string) Option {
	return func(c *clientConfig) { c.tenantID = id }
}

// WithCorrelationID overrides the per-request auto UUID v4 with a fixed value.
// Useful for test determinism; callers in production prefer the default.
func WithCorrelationID(id string) Option {
	return func(c *clientConfig) { c.correlationID = id }
}

// WithHTTPClient replaces the default HTTP client entirely.
// Use this to inject a custom TLS root pool, proxy, or pinned transport.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) { c.httpClient = hc }
}

// newUUID returns a random UUID v4 string using crypto/rand.
// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx (RFC 4122 §4.4).
// Panics only if crypto/rand is broken (OS-level failure; should never occur).
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("yaagentsclient: crypto/rand unavailable: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version bits = 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits = 10xx
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:]),
	)
}

// do executes an HTTP request against path (relative to baseURL) with the
// standard header set applied. body may be nil for GET/DELETE; non-nil
// triggers "Content-Type: application/json".
//
// Callers (GOC-2 resource accessors) provide a pre-serialised JSON body as
// an io.Reader. Response parsing into AgenticResult is added in GOC-3.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("yaagentsclient: build request: %w", err)
	}

	// Authorization
	if c.cfg.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.token)
	}
	// Tenant routing
	if c.cfg.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.cfg.tenantID)
	}
	// Correlation ID — static override wins; otherwise fresh UUID v4
	corrID := c.cfg.correlationID
	if corrID == "" {
		corrID = newUUID()
	}
	req.Header.Set("X-Correlation-ID", corrID)
	// Content-Type only for requests that carry a body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.cfg.httpClient.Do(req)
}
