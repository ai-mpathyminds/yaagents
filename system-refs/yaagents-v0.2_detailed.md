# YAAgents v0.2 — Apache 2.0 OSS + Plugin Middleware + Go Client + LLM Gateway Convergence — Detailed PRD

Status: [READY]
Owner: product-manager (yaagents)
PI: PI2-yaa
Date: 2026-05-30
Profile version: v0.2

> Seeded by chief-architect. Expanded by product-manager 2026-05-30.

---

## 1. Problem & Context

YAAgents v0.1.x (published under the "YAAgents Community License") solves the agentic-API
interface problem. But two structural defects limit its trajectory:

**Defect 1 — Hardcoded middleware chain.**
Authentication, tenant injection, RBAC, and audit are compiled as one non-extensible block
inside the gateway binary. Teams that need a different auth provider, their own tenant
model, an OPA hook, or a custom observability pipeline must fork the gateway. There is no
published plugin contract; there is no community flywheel.

**Defect 2 — Non-OSI license gates community contribution.**
The Community License paywall (≥10 employees OR ≥USD 1M revenue → commercial license
required) is a structural contribution barrier. External contributors cannot submit plugins,
adapters, or SDK improvements under a license that may retroactively require payment. The
license structure is a mismatch for a project whose value proposition depends on ecosystem.

**Compounding consequence — `ai-platform/ai-gateway` duplicates the gateway.**
The internal LLM-routing Go service (`ai-platform/ai-gateway`) exists because yaagents/gateway
had no plugin model. It carries diverging copies of JWT auth (Keycloak realm-aware JWKS),
tenant context, SSE streaming for long-running LLM responses, per-tenant SSE concurrency
limits, execution timeouts, and CORS configuration. The PI1-yaa forward-dependency — "re-base
internal gateway on yaagents/gateway" — cannot land until the plugin model exists. PI2-yaa
collapses this dep by absorbing the LLM-specialisation into yaagents.

---

## 2. Solution Overview

PI2-yaa ships four tightly coupled capabilities as a single v0.2.0 release:

| Capability | What it delivers |
|-----------|-----------------|
| **Apache 2.0 license flip** | `LICENSE` → Apache 2.0 text verbatim; `COMMERCIAL.md` retired; copyright-header SPDX sweep; v0.1.x packages stay Community (non-retroactive) |
| **Plugin middleware system** | Typed `Plugin` interface + `Init → Handler chain → Shutdown` lifecycle; per-plugin YAML config; community-extensible registry; five first-party plugins (a–e) |
| **Go client SDK** (`client-go/`) | Py/TS parity; resource-oriented; typed Go error types for all agentic responses; Go modules tag-driven publish |
| **LLM gateway convergence** | LLM-specialisation concerns from `ai-platform/ai-gateway` absorbed; shape ratified by ADR PI2-yaa-0002; reference example at `examples/llm-gateway/` |

The governing architecture principle does not change:

> Keep your domain resources. Make selected operations agentic.
> Build the agent however you want. Expose it like a governed API.

---

## 3. Target Users

| Tier | Who | Need |
|------|-----|------|
| External primary | Platform engineers, API architects, Go/Python/TS backend engineers, AI-platform teams | Govern agentic capabilities behind standard REST contracts under a true OSS license |
| External secondary | OSS contributors authoring plugins; FastAPI/Go developers; OpenAPI-first teams | Plugin contract to extend gateway; standard interface without license friction |
| Internal forward dep | AimpathyMinds `ai-platform/ai-gateway` consumers | Future migration substrate (plt-aip-lane PI; NOT PI2-yaa scope) |

---

## 4. Agentic REST Response Profile (NORMATIVE)

The following table is the normative Agentic REST Response Profile. **All components MUST
implement and honour it.** This table is copied verbatim from `YAAgents_PRD_README.md` and
carries forward unchanged from v0.1.

| Response Type | HTTP Status | Content-Type |
|---|---:|---|
| `success` | `200` | `application/json` |
| `created` | `201` | `application/json` |
| `accepted` | `202` | `application/vnd.yaagents.operation+json` |
| `clarification_required` | `400` | `application/vnd.yaagents.clarification+json` |
| `validation_failed` | `422` | `application/vnd.yaagents.validation-error+json` |
| `approval_required` | `412` | `application/vnd.yaagents.approval-required+json` |
| `forbidden` | `403` | `application/vnd.yaagents.error+json` |
| `conflict` | `409` | `application/vnd.yaagents.conflict+json` |
| `failed_dependency` | `424` | `application/vnd.yaagents.error+json` |
| `error` | `500` | `application/vnd.yaagents.error+json` |

**Profile versioning:** Every published package MUST declare the profile version it supports.
For v0.2.0: `Supports-YAAgents-Profile: v0.2` in package metadata and `X-YAAgents-Profile: v0.2`
response header.

### 4.1 Clarification Required — canonical body shape (unchanged from v0.1)

```json
{
  "type": "clarification_required",
  "code": "CLARIFICATION_REQUIRED",
  "message": "Additional information is required.",
  "requiredInputs": [
    {
      "name": "successMetric",
      "location": "body",
      "type": "string",
      "required": true,
      "question": "Which success metric should be optimized?",
      "allowedValues": ["ctr", "cpl", "conversion_rate", "lead_quality"]
    }
  ],
  "trace": {
    "correlationId": "corr-123",
    "requestId": "req-456"
  }
}
```

`correlationId` and `requestId` are mandatory in every agentic response body (`trace` block).

---

## 5. Component Contracts

### 5.1 Component 1 — Agentic REST Response Profile (`spec/`) — unchanged

**Directory:** `spec/`

**Responsibilities:** Authoritative prose definition of the Response Profile; profile-version
declaration; normative status × media-type table (§4 above); clarification body contract;
correlation-ID propagation contract; acceptance criteria for conformance.

**v0.2.0 change:** profile version bumped to `v0.2`; no schema changes to existing response
types; SSE streaming contract added (see §7 LLM Gateway).

---

### 5.2 Component 2 — JSON Schemas (`schemas/`) — unchanged

**Directory:** `schemas/`

**Schema list (6 — unchanged from v0.1):**

| File | Validates |
|------|-----------|
| `clarification-required.schema.json` | `application/vnd.yaagents.clarification+json` |
| `validation-failed.schema.json` | `application/vnd.yaagents.validation-error+json` |
| `approval-required.schema.json` | `application/vnd.yaagents.approval-required+json` |
| `conflict.schema.json` | `application/vnd.yaagents.conflict+json` |
| `agentic-error.schema.json` | `application/vnd.yaagents.error+json` (forbidden / failed_dependency / error) |
| `operation-accepted.schema.json` | `application/vnd.yaagents.operation+json` |

**Contract:** Every schema carries `$schema` (JSON Schema Draft-07 minimum), `$id`, and `title`.
The CLI, server SDK, client SDKs, Go client, and conformance tests all reference these schemas.
Schemas are versioned with the profile; `v0.2/` path prefix under `schemas/` for forward compat.

---

### 5.3 Component 3 — OpenAPI Components (`openapi/`) — unchanged

**Directory:** `openapi/`

**Files:**

| File | Contents |
|------|----------|
| `yaagents-components.yaml` | Standard headers, standard response schemas, standard media types, standard error responses |
| `yaagents-response-profile.yaml` | `x-yaagents` operation-metadata extension |

**`x-yaagents` extension fields (unchanged):**

| Field | Type | Description |
|-------|------|-------------|
| `resource` | string | Domain resource name (e.g. `Campaign`) |
| `operationKind` | string | `recommendation` · `generation` · `mutation` · `analysis` |
| `deterministic` | boolean | Whether the operation is deterministic |
| `mutating` | boolean | Whether the operation mutates state |

**OpenAPI response component surface (example — unchanged):**

```yaml
paths:
  /campaigns/{campaignId}/optimizations:
    post:
      operationId: createCampaignOptimization
      x-yaagents:
        resource: Campaign
        operationKind: recommendation
        deterministic: true
        mutating: false
      responses:
        "201":
          content:
            application/json:
              schema: { $ref: "#/components/schemas/CampaignOptimization" }
        "400":
          content:
            application/vnd.yaagents.clarification+json:
              schema: { $ref: "#/components/schemas/ClarificationRequired" }
        "422":
          content:
            application/vnd.yaagents.validation-error+json:
              schema: { $ref: "#/components/schemas/ValidationFailed" }
        "424":
          content:
            application/vnd.yaagents.error+json:
              schema: { $ref: "#/components/schemas/AgenticError" }
```

---

### 5.4 Component 4 — Go Gateway (`gateway/`) — UPDATED for v0.2.0

**Directory:** `gateway/`
**Published image:** `ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` (GHCR, multi-arch: `linux/amd64` + `linux/arm64`)
**Go module:** `github.com/ai-mpathyminds/yaagents/gateway`

**Responsibilities (unchanged + plugin system added):**

| Responsibility | Detail |
|----------------|--------|
| Authentication | Delegated to `token-validator` plugin (JWT RS256/JWKS; always-on) |
| Tenant/actor context | Delegated to `tenant-injector` plugin (default-on) |
| License verification | Delegated to `license-check` plugin (default-on) |
| Plugin chain execution | Execute enabled plugins in declaration order per request |
| Route-level RBAC | Enforce `roles:` from route config; return `403 application/vnd.yaagents.error+json` on failure |
| Context header injection | Forward `X-Correlation-ID` (generate if absent) and `X-Request-ID` to upstream |
| Request routing | Proxy to `target:` URL per route config; preserve HTTP method and body |
| Typed-response passthrough | Preserve upstream `Content-Type` and status; do not re-encode agentic vendor types |
| Response header normalisation | Add `X-YAAgents-Profile: v0.2` to all proxied responses |
| Audit log | Emit structured JSON audit event per request: route, tenant, actor, status, latency, correlation-id |
| Health checks | `GET /healthz` (liveness) · `GET /readyz` (readiness); preserved through plugin chain |
| Metrics | Prometheus text format on `GET /metrics` |
| Graceful shutdown | SIGTERM triggers reverse-order `Shutdown(ctx)` on all enabled plugins; drain deadline configurable |

#### 5.4.1 Gateway configuration (env / config file) — v0.2.0

| Key | Description |
|-----|-------------|
| `GATEWAY_PORT` | Listen port (default `8080`; portfolio allocation `8120`) |
| `GATEWAY_ROUTES_FILE` | Path to route-config YAML |
| `GATEWAY_PLUGINS_FILE` | Path to plugin-config YAML (or inline in routes file) |
| `GATEWAY_SHUTDOWN_TIMEOUT_S` | Plugin drain deadline in seconds (default `30`) |
| `GATEWAY_JWT_SECRET` | HS256 secret (dev/test only — `JWT_TEST_MODE`) |
| `GATEWAY_JWT_JWKS_URL` | JWKS URL for RS256 (production; consumed by token-validator plugin) |
| `GATEWAY_AUDIT_LOG` | `stdout` (default) or future sink config |

**Note:** `GATEWAY_JWT_SECRET` / `GATEWAY_JWT_JWKS_URL` move from gateway-core env vars into the
`token-validator` plugin config block (`jwks_url`, `test_mode`) in v0.2.0. The env vars remain
as convenience overrides for the plugin's Init phase.

#### 5.4.2 Gateway route-config schema (YAML) — v0.2.0

```yaml
# gateway/routes.yaml
routes:
  - id: string                 # Required. Unique route identifier.
    method: string             # Required. HTTP method (GET/POST/PUT/PATCH/DELETE).
    path: string               # Required. Path pattern; {param} placeholders supported.
    target: string             # Required. Upstream URL (e.g. http://campaign-api:8080).
    roles:                     # Optional. Array of required role strings. All must match.
      - string
    tenantRequired: boolean    # Optional (default: false). Reject if X-Tenant-ID absent.
    audit: boolean             # Optional (default: false). Emit audit log entry.
    plugins:                   # Optional. Per-route plugin config overrides (v0.2.0 NEW).
      token-validator:
        enabled: true          # CANNOT be false; gateway rejects config with token-validator.enabled: false.
      tenant-injector:
        enabled: false         # Per-route disable allowed for tenant-injector.
      license-check:
        enabled: false         # Per-route disable allowed for license-check.
```

---

### 5.5 Component 5 — Python FastAPI SDK (`sdk-fastapi/`) — version bump only

**Directory:** `sdk-fastapi/`
**PyPI package:** `yaagents-fastapi`
**v0.2.0 installs via:** `pip install yaagents-fastapi==0.2.0`

**Changes from v0.1.0:** Apache 2.0 license metadata; SPDX headers; `Supports-YAAgents-Profile: v0.2`
package metadata; no API surface changes.

**API surface (unchanged — full table in PI1-yaa detailed PRD §5.5):**

| Symbol | Kind | Description |
|--------|------|-------------|
| `@agentic_operation(...)` | decorator | Declare an agentic endpoint; injects `AgenticContext`, generates OpenAPI response stanzas |
| `AgenticResponse` | class | Factory for all 10 response types |
| `AgenticContext` | class | Injected context: `tenant_id`, `actor_id`, `correlation_id`, `request_id` |
| `RequiredInput` | dataclass | Clarification input descriptor |
| `AgenticResponses` | class | OpenAPI response-collection helpers for `@agentic_operation` |

---

### 5.6 Component 6 — Python Client (`client-python/`) — version bump only

**Directory:** `client-python/`
**PyPI package:** `yaagents-client`
**v0.2.0 installs via:** `pip install yaagents-client==0.2.0`

**Changes from v0.1.0:** Apache 2.0 license metadata; SPDX headers; `Supports-YAAgents-Profile: v0.2`.
No API surface changes.

**API surface (unchanged — full table in PI1-yaa detailed PRD §5.6):**

| Symbol | Kind | Description |
|--------|------|-------------|
| `YaAgentsClient(base_url, token, tenant_id)` | class | Root client |
| `ClarificationRequired` | exception | `400 application/vnd.yaagents.clarification+json`; `.required_inputs` |
| `ValidationFailed` | exception | `422 application/vnd.yaagents.validation-error+json`; `.errors` |
| `FailedDependency` | exception | `424 application/vnd.yaagents.error+json`; `.dependency` |
| `AgenticForbidden` | exception | `403 application/vnd.yaagents.error+json` |

---

### 5.7 Component 7 — TypeScript Client (`client-ts/`) — version bump only

**Directory:** `client-ts/`
**npm package:** `@aimpathyminds/yaagents-client`
**v0.2.0 installs via:** `npm install @aimpathyminds/yaagents-client@0.2.0`

**Changes from v0.1.0:** `"license": "Apache-2.0"` in `package.json`; SPDX headers; profile
version bump to `v0.2`. No API surface changes.

**API surface (unchanged — full table in PI1-yaa detailed PRD §5.7):**

| Symbol | Kind | Description |
|--------|------|-------------|
| `YaAgentsClient({ baseUrl, token, tenantId })` | class | Root client |
| `AgenticResult<T>` | type | Discriminated union of all response types |
| `result.type === 'clarification_required'` | discriminant | Access `result.requiredInputs` |

---

### 5.8 Component 8 — CLI Validator (`cli/`) — version bump only

**Directory:** `cli/`
**PyPI package:** `yaagents-cli`
**v0.2.0 installs via:** `pip install yaagents-cli==0.2.0`

**Changes from v0.1.0:** Apache 2.0 license metadata; SPDX headers; `Supports-YAAgents-Profile: v0.2`.

**CLI command surface (unchanged):**

| Command | Description |
|---------|-------------|
| `yaagents validate-openapi <file.yaml>` | Validates `x-yaagents` metadata, correct content-type per response type, schema refs |
| `yaagents validate-response <file.json>` | Validates body against the relevant JSON schema |
| `yaagents conformance-test <base-url>` | Exercises all mandatory response types; checks headers, correlation-id, plugin headers |
| `yaagents init fastapi` | Generates a FastAPI starter with `@agentic_operation` and correct response wiring |

---

### 5.9 Component 9 — Go Client SDK (`client-go/`) — NEW in v0.2.0

**Directory:** `client-go/`
**Go module:** `github.com/ai-mpathyminds/yaagents/client-go`
**Published via:** Go module proxy (`proxy.golang.org`); tag-driven at `v0.2.0`
**Install via:** `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0`

The Go client is the idiomatic Go analog to `client-python` and `client-ts`. It is a
library, not a CLI tool. Design constraints: zero non-stdlib runtime dependencies (stdlib
`net/http` only); context-propagation throughout; error-style (typed errors on non-success
agentic responses); result-style variant via `AgenticResult` for callers that prefer union
dispatch.

**API surface:**

| Symbol | Kind | Description |
|--------|------|-------------|
| `New(baseURL string, opts ...Option) *Client` | constructor | Root client; one instance per service |
| `WithToken(token string) Option` | option | Bearer token |
| `WithTenantID(id string) Option` | option | `X-Tenant-ID` header |
| `WithCorrelationID(id string) Option` | option | Overrides auto-generated UUID |
| `WithHTTPClient(c *http.Client) Option` | option | Custom HTTP client for transport / TLS |
| `client.Campaigns() CampaignsResource` | method | Root resource accessor |
| `CampaignsResource.ByID(id string) CampaignResource` | method | Sub-resource for a specific campaign |
| `CampaignResource.Optimizations() OptimizationsResource` | method | Optimizations sub-resource |
| `OptimizationsResource.Create(ctx, body) (*AgenticResult, error)` | method | `POST /campaigns/{id}/optimizations` |
| `OptimizationsResource.Get(ctx, id) (*AgenticResult, error)` | method | `GET /campaigns/{id}/optimizations/{id}` |
| `CampaignResource.Assets() AssetsResource` | method | Assets sub-resource |
| `AssetsResource.Generate(ctx, body) (*AgenticResult, error)` | method | `POST /campaigns/{id}/assets:generate` |

**AgenticResult (result-style):**

```go
// AgenticResult is the discriminated result returned by all client methods.
// Callers may switch on Type or cast to a typed error via Err().
type AgenticResult struct {
    Type     string          // "created" | "success" | "accepted" | "clarification_required" | etc.
    Status   int             // HTTP status code
    Resource json.RawMessage // Populated for created / success
    // clarification_required
    RequiredInputs []RequiredInput
    // accepted
    OperationID string
    // error fields
    Message string
    Trace   Trace
}

// Err returns a typed error for non-success results, nil for success/created/accepted.
func (r *AgenticResult) Err() error
```

**Typed errors (error-style):**

```go
// ClarificationRequired is returned when the server responds 400 clarification.
type ClarificationRequired struct {
    RequiredInputs []RequiredInput
    Trace          Trace
}
func (e *ClarificationRequired) Error() string

// ValidationFailed is returned on 422.
type ValidationFailed struct {
    Errors []ValidationError
    Trace  Trace
}
func (e *ValidationFailed) Error() string

// FailedDependency is returned on 424.
type FailedDependency struct {
    Dependency string
    Message    string
    Trace      Trace
}
func (e *FailedDependency) Error() string

// AgenticForbidden is returned on 403.
type AgenticForbidden struct {
    Message string
    Trace   Trace
}
func (e *AgenticForbidden) Error() string
```

**Idiomatic usage example:**

```go
import "github.com/ai-mpathyminds/yaagents/client-go"

client := yaagentsclient.New(
    "http://localhost:8120",
    yaagentsclient.WithToken("my-jwt"),
    yaagentsclient.WithTenantID("tenant-001"),
)

result, err := client.Campaigns().ByID("cmp-123").Optimizations().Create(ctx, map[string]any{
    "goal": "reduce_cost_per_lead",
})
if err != nil {
    var clarify *yaagentsclient.ClarificationRequired
    if errors.As(err, &clarify) {
        for _, input := range clarify.RequiredInputs {
            fmt.Printf("Required: %s — %s\n", input.Name, input.Question)
        }
        return
    }
    return err
}
fmt.Printf("Created: %s\n", result.Resource)
```

**Headers injected by default:**
- `Authorization: Bearer {token}`
- `X-Tenant-ID: {tenantID}` (if set)
- `X-Correlation-ID: {auto-generated UUID v4}` (overridable via `WithCorrelationID`)
- `Content-Type: application/json`

**Go module path:** `github.com/ai-mpathyminds/yaagents/client-go`
**Minimum Go version:** `go 1.22`
**Runtime dependencies:** stdlib only (`net/http`, `encoding/json`, `context`, `errors`)

---

## 6. Plugin Middleware System (PI2-yaa — NEW)

### 6.1 Design Principles

1. **Contract-first.** The `Plugin` interface is the stable contract. All five first-party
   plugins (a–e) are consumers of the same interface; no special-casing inside the gateway
   core for any first-party plugin.
2. **Community-extensible.** A community developer authors a Go module, implements `Plugin`,
   calls `Register()` from an `init()` function, and builds a gateway binary that imports
   their plugin. The plugin contract is versioned and semver-stable.
3. **Config-driven.** Each plugin declares its config schema. Operators configure plugins via
   YAML (see §6.3). The gateway validates the YAML against each plugin's requirements at boot.
4. **Always-on floor.** `token-validator` (plugin a) is the security floor. Its `enabled` flag
   CANNOT be set to `false` in any YAML config. The gateway MUST reject startup if the
   routes or plugins YAML contains `token-validator.enabled: false`.
5. **Graceful shutdown.** SIGTERM triggers `Shutdown(ctx)` on all active plugins in reverse
   declaration order. `ctx` carries a configurable drain deadline (`GATEWAY_SHUTDOWN_TIMEOUT_S`,
   default 30 s).

### 6.2 Plugin Interface (Go)

```go
// Package plugin defines the yaagents gateway plugin contract.
// Community plugins implement Plugin and call Register() from an init().
package plugin

import (
    "context"
    "net/http"
)

// Plugin is the extension point for yaagents gateway middleware.
type Plugin interface {
    // Name returns the canonical plugin name. MUST match the YAML config key
    // (e.g. "token-validator", "tenant-injector").
    Name() string

    // Init is called once at gateway startup with the plugin-specific config.
    // Returns a non-nil error to abort gateway startup; the gateway logs the
    // error and exits non-zero.
    Init(cfg PluginConfig) error

    // Handler wraps next. Called once per request in plugin-chain order
    // (declaration order in YAML, innermost plugin closest to upstream).
    // Must call next.ServeHTTP(w, r) or return an agentic error response.
    Handler(next http.Handler) http.Handler

    // Shutdown is called on SIGTERM. ctx carries the drain deadline
    // (GATEWAY_SHUTDOWN_TIMEOUT_S). Called in reverse declaration order.
    Shutdown(ctx context.Context) error
}

// PluginConfig is the typed accessor for the deserialized YAML plugin block.
type PluginConfig interface {
    GetString(key string) string
    GetBool(key string) bool
    GetStringSlice(key string) []string
    GetInt(key string) int
    Raw() map[string]any
}

// Register adds p to the gateway's global plugin registry.
// Call from an init() function in your plugin package so the plugin is
// available before gateway.Run() is invoked.
func Register(p Plugin)

// Registered returns the ordered list of registered plugins.
// Used by gateway internals; not intended for plugin authors.
func Registered() []Plugin
```

### 6.3 Plugin YAML Config Schema

The top-level `plugins:` block lives in the gateway config file (or inline in `routes.yaml`).
Each key MUST match the return value of `Plugin.Name()`.

```yaml
plugins:
  # (a) token-validator — JWT RS256/JWKS; ALWAYS-ON; enabled: false is FORBIDDEN.
  token-validator:
    enabled: true                      # REQUIRED true; gateway rejects false.
    jwks_url: "https://auth.example.com/.well-known/jwks.json"
    cache_ttl_seconds: 600             # JWKS cache TTL; default 600.
    audience: ""                       # Optional; validated if non-empty.
    test_mode: false                   # true = HS256 secret (dev only).
    jwt_secret: ""                     # Required when test_mode: true.

  # (b) tenant-injector — parse tenant header, validate, inject actor_ctx.
  tenant-injector:
    enabled: true                      # Default true; operator CAN set false.
    header: "X-Tenant-ID"             # Header name to read; default X-Tenant-ID.
    allowlist: []                      # Optional; if non-empty, only these tenant IDs accepted.
    inject_header: "X-Actor-Tenant"   # Header injected into upstream request.

  # (c) license-check — product-license token verification for commercial consumers.
  license-check:
    enabled: true                      # Default true; operator CAN set false.
    license_url: ""                    # URL to verify product-license token; required if enabled.
    cache_ttl_seconds: 300            # License-check response cache TTL; default 300.
    header: "X-License-Token"         # Header name carrying the license token.

  # (d) prompt-sanitize — prompt injection defence; PI2-yaa ships interface + stub.
  prompt-sanitize:
    enabled: false                     # Default false; off-by-default in PI2-yaa.
    # Full config schema deferred to PI3-yaa or community reference impl.

  # (e) otel-audit — OpenTelemetry trace + log emission; PI2-yaa ships interface + stub.
  otel-audit:
    enabled: false                     # Default false; off-by-default in PI2-yaa.
    endpoint: ""                       # OTEL collector endpoint; required if enabled.
    service_name: "yaagents-gateway"  # OTEL service.name attribute.
    # Full config schema deferred to PI3-yaa or community reference impl.
```

### 6.4 Plugin Lifecycle

```
Gateway boot:
  1. All plugin init() functions run → plugin.Register() called for each.
  2. Gateway reads plugins YAML block.
  3. For each plugin listed (in declaration order): plugin.Init(cfg) called.
     → If any Init returns error: gateway logs + exits 1.
  4. Gateway asserts token-validator.enabled == true; exit 1 if violated.

Per request:
  5. Plugin handlers are applied in declaration order as middleware.
     (declaration order = the order keys appear in the YAML plugins: block)
     Innermost handler is closest to upstream.

SIGTERM:
  6. plugin.Shutdown(ctx) called in REVERSE declaration order.
     ctx deadline = time.Now() + GATEWAY_SHUTDOWN_TIMEOUT_S.
  7. Gateway drains in-flight requests; then exits 0.
```

### 6.5 First-Party Plugin Set

| ID | Plugin name | Default | Can disable? | PI2-yaa status |
|----|-------------|---------|--------------|----------------|
| a | `token-validator` | always-on | **NO** — security floor | Full implementation |
| b | `tenant-injector` | on | Yes | Full implementation |
| c | `license-check` | on | Yes | Full implementation |
| d | `prompt-sanitize` | off | Yes | Interface + stub; impl deferred to PI3-yaa |
| e | `otel-audit` | off | Yes | Interface + stub; impl deferred to PI3-yaa |

**Plugin (a) — token-validator:**
- JWT RS256 validation via JWKS URL (configurable `cache_ttl_seconds`).
- Dev/test mode: HS256 via `jwt_secret` (only when `test_mode: true`).
- Reuses `portfolio/packages/go/auth-jwks/` if PI14-oppor extraction has landed; otherwise
  re-implements minimally (inline). Architect confirms at A-3 ADR PI2-yaa-0005.
- On validation failure: returns `403 application/vnd.yaagents.error+json` immediately; does
  not call `next`.
- **Cannot be disabled.** Gateway exit 1 if YAML carries `token-validator.enabled: false`.

**Plugin (b) — tenant-injector:**
- Reads the tenant ID from the configured `header` (default `X-Tenant-ID`).
- If `allowlist` is non-empty and the tenant ID is not in the list: returns `403`.
- Injects `X-Actor-Tenant` header into the upstream request.
- Default-on; operator may disable per-gateway or per-route.

**Plugin (c) — license-check:**
- Reads a product-license token from the configured `header`.
- Calls `license_url` to verify the token (with `cache_ttl_seconds` caching).
- On verification failure: returns `403 application/vnd.yaagents.error+json`.
- Default-on; operator may disable per-gateway or per-route (e.g. public health endpoints).

**Plugin (d) — prompt-sanitize (stub):**
- PI2-yaa ships the `Plugin` interface implementation with `Init` and `Handler` no-ops.
- `Handler` passes through all requests unchanged; emits a `warn: prompt-sanitize is a stub`
  log line on first request when `enabled: true`.
- Full prompt-injection defence: deferred to PI3-yaa or community reference impl.

**Plugin (e) — otel-audit (stub):**
- PI2-yaa ships the `Plugin` interface implementation; `Init` validates `endpoint` if set.
- `Handler` emits a span placeholder (noop tracer); real OTEL exporter wiring deferred.
- Full OpenTelemetry export: deferred to PI3-yaa or community reference impl.

### 6.6 Community Plugin Authoring Contract

A community developer creates a new Go module (e.g. `github.com/my-org/yaagents-my-plugin`):

```go
package myplugin

import (
    "net/http"
    "context"
    "github.com/ai-mpathyminds/yaagents/gateway/plugin"
)

type MyPlugin struct { ... }
func (p *MyPlugin) Name() string { return "my-plugin" }
func (p *MyPlugin) Init(cfg plugin.PluginConfig) error { ... }
func (p *MyPlugin) Handler(next http.Handler) http.Handler { ... }
func (p *MyPlugin) Shutdown(ctx context.Context) error { ... }

func init() { plugin.Register(&MyPlugin{}) }
```

The gateway operator builds a custom binary that imports both yaagents/gateway and the community
plugin module. The plugin is wired by import side-effect; no gateway fork required.

---

## 7. LLM Gateway Convergence (PI2-yaa — NEW)

### 7.1 ai-platform/ai-gateway LLM Specialisation — What Gets Absorbed

The `ai-platform/ai-gateway` Go service (path: `ai-platform/services/ai-gateway/`) carries the
following LLM-specific concerns not present in yaagents/gateway v0.1.x:

| Capability | Source (ai-platform/ai-gateway) | Absorption notes |
|-----------|--------------------------------|-----------------|
| SSE streaming proxy | `internal/proxy/sse.go` + `internal/proxy/proxy.go` | Pipe-and-flush semantics for LLM streaming responses |
| Per-tenant SSE concurrency limit | `MaxSSEConnectionsPerTenant` config | Rate-limit concurrent SSE streams per tenant; `429` on excess |
| Long-running execution timeout | `ExecutionTimeoutSeconds` + `SSEReadTimeout()` | Context deadline = execution_timeout + 30 s for SSE routes |
| Keycloak realm-aware JWKS | `KeycloakRealm` in config.go | JWKS URL constructed from realm name |
| CORS configuration | `AllowedOrigins` / `CORS_ALLOWED_ORIGINS` | Browser callers; exact-match allow-list |
| SSE-specific route handling | `/api/v1/agents/drafts/{draftId}:startTestSession`, `/api/v1/agents/test-sessions/{id}/messages` | SSE routes registered with per-route timeout middleware |
| Prometheus SSE metrics | `internal/metrics/sse_proxy.go` | Active SSE connections gauge + error counter |

**Absorption shape:** architect-call at A-3 via ADR PI2-yaa-0002. Two options:

- **Option A (layer-atop):** LLM-specialisation becomes a set of yaagents-compatible plugins
  and route-config extensions activated inside `yaagents/gateway`. SSE proxy, SSE limiter,
  CORS, and execution-timeout become plugins or gateway-core extensions.
- **Option B (sibling):** A new `yaagents/llm-gateway` component extends `yaagents/gateway`
  with LLM-specific concerns as a thin sibling binary. Gateway core remains unchanged.

Both options produce a reference example at `examples/llm-gateway/`. Code lineage (copy vs
go-modules-consume from ai-platform) is also an ADR PI2-yaa-0002 decision.

### 7.2 Reference Example: LLM Gateway (`examples/llm-gateway/`)

**Directory:** `examples/llm-gateway/`
**Compose file:** `examples/llm-gateway/docker-compose.yml`
**Ports:** Gateway `8120` (or `8122` if campaign-api example retains `8121`; platform-engineer
confirms at A-4 compose-linter pass).

**Demonstrated flows:**

| Flow | Trigger | Expected response |
|------|---------|-------------------|
| Standard LLM call | Request with valid JWT + tenant; non-streaming | `201 application/json` with LLM result |
| SSE streaming response | Request with `Accept: text/event-stream` | Progressive SSE stream through gateway |
| SSE concurrency limit exceeded | Concurrent SSE requests beyond `max_sse_connections` | `429 application/vnd.yaagents.error+json` |
| Execution timeout | LLM call exceeds `execution_timeout_seconds` | `500 application/vnd.yaagents.error+json` |
| Plugin chain with CORS | Browser preflight request | `200` with CORS headers; no upstream call |

**Compose topology:**

```
yaagents-gateway (port 8120, LLM config)  ←→  llm-api (internal; mock LLM backend)
```

The reference example uses a mock LLM backend that can simulate streaming delays and timeouts
without requiring an actual LLM provider API key.

**Quick start:**

```bash
cd examples/llm-gateway
docker compose up

# Standard call
curl -X POST http://localhost:8120/completions \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Generate a campaign headline", "stream": false}'

# SSE streaming
curl -X POST http://localhost:8120/completions \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Accept: text/event-stream" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Generate a campaign headline", "stream": true}'
```

---

## 8. Apache 2.0 License Model

### 8.1 License Flip — v0.2.0

**v0.2.0 ships under the Apache License, Version 2.0.**

This decision (user-direct, 2026-05-30) reverses the PI1-yaa-0004 source-available model.
ADR PI2-yaa-0003 supersedes ADR PI1-yaa-0004. The YAAgents Community License and
`COMMERCIAL.md` are retired with v0.2.0.

**Non-retroactive boundary:** v0.1.x packages already published to PyPI/npm/GHCR stay under the
YAAgents Community License. v0.2.0 is the clean SemVer minor cut; users who need Apache 2.0
must upgrade to v0.2.0.

**Legal-review-pending disclaimer (verbatim from GTM README §Appendix):**

> This GTM README includes a draft licensing strategy for product planning. It is not legal
> advice. Before publishing the license publicly or accepting external contributions, consult
> a qualified software licensing lawyer.

Legal review gates the public re-announce and the removal of the `legal-review-pending` banner
from `CONTRIBUTING.md`. It does NOT gate PI2-yaa close; the license file, headers, and package
metadata ship with the Apache 2.0 text before legal sign-off.

### 8.2 Apache License, Version 2.0 (verbatim)

```
                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship made available under the
      License, as indicated by a copyright notice that is included in or
      attached to the work (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other
      transformations represent, as a whole, an original work of authorship.
      For the purposes of this License, Derivative Works shall not include
      works that remain separable from, or merely link (or bind by name) to
      the interfaces of, the Work and Derivative Works thereof.

      "Contribution" shall mean, as submitted to the Licensor for inclusion
      in the Work by the copyright owner or by an individual or Legal Entity
      authorized to submit on behalf of the copyright owner. For the purposes
      of this definition, "submit" means any form of electronic, verbal, or
      written communication sent to the Licensor or its representatives,
      including but not limited to communication on electronic mailing lists,
      source code control systems, and issue tracking systems that are managed
      by, or on behalf of, the Licensor for the purpose of discussing
      improvements to the Work, but excluding communication that is
      conspicuously marked or otherwise designated in writing by the copyright
      owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any Legal Entity on behalf of
      whom a Contribution has been received by the Licensor and included
      within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable by
      such Contributor that are necessarily infringed by their Contribution(s)
      alone or by the combined Contribution(s) with the Work to which such
      Contribution(s) was submitted. If You institute patent litigation against
      any entity (including a cross-claim or counterclaim in a lawsuit) alleging
      that the Work or any patent claim embodied in the Work constitutes direct
      or contributory patent infringement, then any patent licenses granted to
      You under this License for that Work shall terminate as of the date such
      litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the Work or
      Derivative Works thereof in any medium, with or without modifications,
      and in Source or Object form, provided that You meet the following
      conditions:

      (a) You must give any other recipients of the Work or Derivative Works
          a copy of this License; and

      (b) You must cause any modified files to carry prominent notices stating
          that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works that You
          distribute, all copyright, patent, trademark, and attribution notices
          from the Source form of the Work, excluding those notices that do not
          pertain to any part of the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its distribution,
          You must include a readable copy of the attribution notices contained
          within such NOTICE file, in at least one of the following places:
          within a NOTICE text file distributed as part of the Derivative Works;
          within the Source form or documentation, if provided along with the
          Derivative Works; or, within a display generated by the Derivative
          Works, if and wherever such third-party notices normally appear. The
          contents of the NOTICE file are for informational purposes only and do
          not modify the License. You may add Your own attribution notices within
          Derivative Works that You distribute, alongside or as an addition to
          the NOTICE text from the Work, provided that such additional attribution
          notices cannot be construed as modifying the License.

      You may add Your own license statement for Your modifications and may
      provide additional grant of rights to use, copy, modify, merge, publish,
      distribute, sublicense, and/or sell copies of the Modifications, and to
      permit persons to whom the Software is furnished to do so.

   5. Submission of Contributions. Unless You explicitly state otherwise, any
      Contribution intentionally submitted for inclusion in the Work by You to
      the Licensor shall be under the terms and conditions of this License,
      without any additional terms or conditions. Notwithstanding the above,
      nothing herein shall supersede or modify the terms of any separate license
      agreement you may have executed with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade names,
      trademarks, service marks, or product names of the Licensor, except as
      required for reasonable and customary use in describing the origin of the
      Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or agreed to in
      writing, Licensor provides the Work (and each Contributor provides its
      Contributions) on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
      ANY KIND, either express or implied, including, without limitation, any
      warranties or conditions of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or
      FITNESS FOR A PARTICULAR PURPOSE. You are solely responsible for
      determining the appropriateness of using or reproducing the Work and assume
      any risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory, whether in
      tort (including negligence), contract, or otherwise, unless required by
      applicable law (such as deliberate and grossly negligent acts) or agreed to
      in writing, shall any Contributor be liable to You for damages, including
      any direct, indirect, special, incidental, or exemplary damages of any
      character arising as a result of this License or out of the use or
      inability to use the Work (including but not limited to damages for loss of
      goodwill, work stoppage, computer failure or malfunction, or all other
      commercial damages or losses), even if such Contributor has been advised of
      the possibility of such damages.

   9. Accepting Warranty or Indemnifying a Third Party. While redistributing the
      Work or Derivative Works thereof, You may choose to offer, and charge a fee
      for, acceptance of support, warranty, indemnity, or other liability
      obligations and rights consistent with this License. However, in accepting
      such obligations, You may act only on Your own behalf and on Your sole
      responsibility, not on behalf of any other Contributor, and only if You
      agree to indemnify, defend, and hold each Contributor harmless for any
      liability incurred by, or claims asserted against, such Contributor by
      reason of your accepting any such warranty or additional liability.

   END OF TERMS AND CONDITIONS

   APPENDIX: How to apply the Apache License to your work.

      To apply the Apache License to your work, attach the following boilerplate
      notice, with the fields enclosed by brackets "[]" replaced with your own
      identifying information. (Don't include the brackets!) The text should be
      enclosed in the appropriate comment syntax for the programming language
      concerned.

   Copyright [yyyy] [name of copyright owner]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
   implied. See the License for the specific language governing
   permissions and limitations under the License.
```

### 8.3 COMMERCIAL.md Retirement

`COMMERCIAL.md` is deleted from the repository root as part of the license-flip WI. The
`README.md` carries a pointer:

> **License:** Apache 2.0 — see `LICENSE`. v0.1.x packages shipped under the YAAgents
> Community License remain under that license (non-retroactive). For questions about
> historical v0.1.x usage, contact bhaskar@aimpathyminds.com.

### 8.4 Copyright Header Sweep

All source files in `gateway/`, `sdk-fastapi/`, `client-python/`, `client-ts/`, `cli/`,
`plugins/`, `client-go/`, and `examples/` receive the SPDX header:

```
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds
```

Python files use `#` comment syntax. YAML/JSON files omit SPDX headers (no standard comment
syntax). The go-developer and python-developer lanes own the sweep for their respective
components; the license-flip WI is the single commit that contains all header changes to avoid
partial-sweep states in git history.

---

## 9. Publishing Model (v0.2.0)

### 9.1 Gateway image (GHCR)

```
ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0
ghcr.io/ai-mpathyminds/yaagents-gateway:latest   (convenience alias)
```

- Registry: GitHub Container Registry (GHCR)
- Architectures: `linux/amd64`, `linux/arm64` (multi-arch)
- Build: multi-stage Alpine; non-root user; `CGO_ENABLED=0`
- SBOM: generated via Syft at publish time
- CI: GitHub Actions; **OIDC trusted publishing via `docker/login-action` with GHCR OIDC** — no long-lived registry tokens
- Cosign signing: deferred to PI3-yaa
- Image metadata: `org.opencontainers.image.licenses=Apache-2.0`

### 9.2 Python packages (PyPI)

| Package | Version | Description |
|---------|---------|-------------|
| `yaagents-fastapi` | `0.2.0` | FastAPI server SDK |
| `yaagents-client` | `0.2.0` | Python client |
| `yaagents-cli` | `0.2.0` | CLI validator |

- Build tool: Hatch
- CI: GitHub Actions; **OIDC trusted publishing via PyPI Trusted Publisher** — no API tokens in CI
- Package metadata: `License: Apache-2.0`; `Supports-YAAgents-Profile: v0.2`
- TestPyPI run gates production publish

### 9.3 TypeScript package (npm)

| Package | Version | Description |
|---------|---------|-------------|
| `@aimpathyminds/yaagents-client` | `0.2.0` | TypeScript/JavaScript client |

- Format: ESM primary + CommonJS compat bundle
- CI: GitHub Actions; **OIDC trusted publishing via npm Provenance** — no `NPM_TOKEN` in CI
- `package.json`: `"license": "Apache-2.0"`; `"description"` carries profile version

### 9.4 Go client (Go module proxy)

| Module | Version | Path |
|--------|---------|------|
| `github.com/ai-mpathyminds/yaagents/client-go` | `v0.2.0` | Tag-driven; `proxy.golang.org` |

- Publish via: `git tag client-go/v0.2.0` + `git push --tags` on the yaagents repository;
  `proxy.golang.org` indexes the tag automatically within ~30 min.
- Module path: `github.com/ai-mpathyminds/yaagents/client-go` (module within monorepo)
- Module file: `client-go/go.mod` with `module github.com/ai-mpathyminds/yaagents/client-go`
- No CI-triggered publish needed; tag push is sufficient for Go modules.

### 9.5 Schemas and OpenAPI components

Distributed in-repo (`schemas/v0.2/`, `openapi/`) and attached to GitHub Releases as versioned
archives. No standalone registry publish.

---

## 10. NFR Seeds (Platform-Engineer Expands at A-4)

The following NFR concerns are seeded here for platform-engineer to size and attach to the
appropriate WI files at A-4. This list is advisory; platform-engineer may reorder, combine, or
add.

| Category | Seed concern |
|----------|-------------|
| `[SEC]` Gateway | `token-validator` plugin cannot be disabled; gateway exit 1 verification test |
| `[SEC]` Gateway | Plugin sandboxing posture: community plugins compiled-in (not dynamic loading); no `plugin.Open()` or `dlopen` |
| `[SEC]` Gateway | `govulncheck` + `trivy` pass on gateway binary + image; target 0 HIGH/CRITICAL |
| `[SEC]` Gateway | `/healthz` and `/readyz` routes are pre-auth (no token-validator wrapping); verify in integration tests |
| `[SEC]` client-go | TLS: `http.DefaultTransport` used by default (system CAs); `WithHTTPClient` option allows TLS pinning |
| `[SEC]` client-go | `govulncheck` pass; target 0 deps |
| `[SEC]` License | No remaining Community License SPDX identifiers in v0.2.0 artifacts; automated scan in CI |
| `[SRE]` Gateway | JSON structured logging with `request_id` / `correlation_id` propagation through plugin chain |
| `[SRE]` Gateway | Graceful shutdown: all plugins drain before process exit; max `GATEWAY_SHUTDOWN_TIMEOUT_S` s |
| `[SRE]` Gateway | Prometheus metrics preserved through plugin chain; per-plugin latency histogram (advisory) |
| `[SRE]` Gateway | `healthz` / `readyz` pass with all five plugins loaded (including stubs d+e when disabled) |
| `[SUPPLY]` All | Multi-arch gateway image `v0.2.0` (amd64 + arm64); SBOM attached to GHCR image |
| `[SUPPLY]` All | OIDC trusted publishing for PyPI (`yaagents-fastapi`, `yaagents-client`, `yaagents-cli`) v0.2.0 |
| `[SUPPLY]` All | OIDC trusted publishing for npm (`@aimpathyminds/yaagents-client`) v0.2.0 |
| `[SUPPLY]` All | Go modules tag (`client-go/v0.2.0`) pushed; verify `proxy.golang.org` index within 30 min |
| `[SUPPLY]` License | Apache 2.0 `LICENSE` present in repo root; SPDX header sweep complete before release tag |
| `[FIN]` | N/A — no AWS run-rate change in PI2-yaa (no TF edits; GHCR/PyPI/npm are public registries) |

Cosign signing + SBOM attestation hardening: displaced to PI3-yaa; note only in A-4 NFR pass.

---

## 11. Open Questions

| ID | Question | Owner | Gate |
|----|----------|-------|------|
| OQ-1 | **ai-gateway absorption shape**: layer-atop `yaagents/gateway` (Option A) vs sibling `yaagents/llm-gateway` (Option B)? | yaagents-architect | ADR PI2-yaa-0002 at A-3 |
| OQ-2 | **ai-gateway code lineage**: (i) copy under Apache 2.0 with attribution + ai-platform/ai-gateway becomes legacy, or (ii) publish yaagents plugin interface + ai-platform/ai-gateway becomes a yaagents consumer via go modules? | yaagents-architect | ADR PI2-yaa-0002 at A-3 |
| OQ-3 | **token-validator JWKS reuse**: `portfolio/packages/go/auth-jwks/` — has PI14-oppor extracted it? If yes, token-validator imports it; if no, plugin re-implements minimally inline. | yaagents-architect | ADR PI2-yaa-0005 at A-3; verify `portfolio/LIBRARIES.md` + PI14-oppor roadmap status |
| OQ-4 | **LLM example port**: if `examples/campaign-api/` keeps port `8121`, LLM example uses `8122`; if campaign-api retired/merged, `8121` is available. Compose-linter validates at A-4. | platform-engineer | A-4 compose-linter pass |
| OQ-5 | **Legal review timeline**: lawyer sign-off on Apache 2.0 adoption + CONTRIBUTING.md CLA-lift gates public re-announce and removal of `legal-review-pending` banner. PI2-yaa close does NOT gate on this. | chief-architect/user | PC-6 non-engineering checklist item |
| OQ-6 | **Plugin registry versioning**: does the `Plugin` interface carry a `Version() string` method, or is versioning by the Go module tag? | yaagents-architect | ADR PI2-yaa-0001 at A-3 |
| OQ-7 | **prompt-sanitize stub behaviour**: when `enabled: true` is set for the stub, does the gateway (a) log-and-pass-through, or (b) exit 1 at boot with a "stub not suitable for production" error? | yaagents-architect | ADR PI2-yaa-0005 at A-3 |

---

## 12. Out of Scope

| Item | Where it lands |
|------|---------------|
| Kubernetes manifests, Helm chart (GHCR OCI) | PI3-yaa |
| Cosign image signing + SBOM attestation hardening | PI3-yaa |
| prompt-sanitize reference implementation | PI3-yaa or community contribution |
| otel-audit reference implementation | PI3-yaa or community contribution |
| Plugin marketplace UI / discovery service | Not planned; registry is API-only |
| ai-platform consumer migration off `ai-platform/ai-gateway` | Future plt-aip-lane PI (separate seed) |
| Retroactive re-licensing of v0.1.x packages | Not in scope; v0.1.x stays Community License |
| GTM content (demo videos, launch blog, social posts) | Not engineering scope |
| v0.3+ adapters (Spring Boot, ASP.NET, LangGraph, SK) | Future PIs or community |
| Async operation profile + approval-flow runtime | PI3-yaa or v0.3+ |
| Frontend / UI of any kind | No UI surface in PI2-yaa |
| `ai-platform/ai-gateway` ECS service changes | Service runs unchanged this PI |

---

## 13. Reference Example Flows (Summary)

### 13.1 Campaign API (PI1-yaa — unchanged, updated to v0.2.0 package versions)

```http
POST /campaigns/{campaignId}/optimizations
```

All flows from PI1-yaa §6.2 apply unchanged. Package versions bump to `0.2.0`;
`X-YAAgents-Profile: v0.2` in response headers.

### 13.2 LLM Gateway (PI2-yaa — new)

**Flow 1 — Plugin chain with token-validator only:**

```
Client → Gateway (token-validator plugin validates JWT) → llm-api upstream → 201 application/json
```

**Flow 2 — Full plugin chain (a + b + c):**

```
Client [JWT + X-Tenant-ID + X-License-Token] →
  token-validator (validates JWT) →
  tenant-injector (validates + injects X-Actor-Tenant) →
  license-check (validates license token) →
  Gateway routes to llm-api →
  201 application/json
```

**Flow 3 — SSE streaming:**

```
Client [Accept: text/event-stream] →
  Plugin chain →
  Gateway SSE proxy (pipe-and-flush) →
  text/event-stream progressive response
```

**Flow 4 — SSE concurrency limit:**

```
11th concurrent SSE request from tenant-001 →
  per-tenant SSE limiter →
  429 application/vnd.yaagents.error+json
```

**Flow 5 — Community plugin integrated:**

```
Client →
  token-validator →
  tenant-injector →
  license-check →
  my-community-plugin (imported as init() side-effect) →
  upstream →
  201 application/json
```
