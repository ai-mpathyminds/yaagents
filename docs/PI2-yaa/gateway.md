# PI2-yaa — Component: Gateway core (`gateway/`) — plugin chain + LLM specialisation

Owner lane: **go-developer**. Sprints 1, 2, 4, 5. Published as
`ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` (publish WI in
`release-and-publish.md`).

ADRs: PI1-yaa-0001 (net-new base; carries forward — gateway remains the
future BASE for `platform-services/gateway` + `ai-platform/ai-gateway`),
PI2-yaa-0001 (plugin interface contract + versioning), PI2-yaa-0002
(ai-gateway absorption shape = Option A layer-atop + code lineage = copy
under Apache 2.0 with attribution), PI2-yaa-0005 (token-validator JWKS
sourcing; prompt-sanitize stub behaviour).

> **Library gate (mandatory citation):** every gateway-core WI carries
> `library_ref: ADR PI1-yaa-0001 + ADR PI2-yaa-0001 + ADR PI2-yaa-0002`
> (per WI as applicable — see each entry). The yaagents gateway remains the
> net-new OSS base; PI2-yaa adds the plugin interface (PI2-yaa-0001) and
> absorbs LLM specialisation as a layer-atop concern (PI2-yaa-0002).
> Internal-gateway cross-product import in either direction is still
> forbidden in PI2-yaa.

---

### WI-2yaa.PLG-2: Gateway plugin-loader refactor (PI1-yaa middleware → plugin chain) [DONE] — Sprint 1
service: yaagents/gateway
parent_feature: F-PLUGIN
brief: Refactor `gateway/cmd/` + `gateway/internal/` so the PI1-yaa
hardcoded middleware chain (auth → tenant → RBAC → audit) is **replaced**
by the plugin loader. Boot sequence per PRD §6.4: (1) all plugin `init()`
functions have already run (import-side-effect); (2) read plugins YAML
block from `GATEWAY_PLUGINS_FILE` (or inline in `routes.yaml`); (3) for
each plugin key in declaration order — call `plugin.Init(cfg)`; on any
non-nil error → log + `os.Exit(1)`. **Always-on assertion**: gateway core
verifies `plugins["token-validator"]["enabled"] == true`; on `false` →
log + `os.Exit(1)` (PRD §6.1 #4 + ADR PI2-yaa-0001 §2). Route-level RBAC
stays in the gateway core (NOT a plugin — PI1-yaa GW-4 logic preserved).
The PI1-yaa env vars `GATEWAY_JWT_SECRET` and `GATEWAY_JWT_JWKS_URL` remain
as convenience overrides for the `token-validator` plugin's Init phase
(PRD §5.4.1 note) — gateway core reads them and writes them into the
plugin's PluginConfig if the YAML block does not explicitly set them.
acceptance:
- `gateway.Run()` builds the plugin chain from YAML; PI1-yaa-style hardcoded chain code paths removed (`grep -rn "authMiddleware\|tenantMiddleware\|auditMiddleware" gateway/internal/` returns 0 hits after refactor)
- `plugins["token-validator"]["enabled"] == false` → `os.Exit(1)` with explicit error message containing `"token-validator cannot be disabled"`
- Any plugin `Init` returning non-nil error → `os.Exit(1)` with the plugin name in the error message
- `GATEWAY_JWT_SECRET` / `GATEWAY_JWT_JWKS_URL` env vars propagate into token-validator PluginConfig when YAML block omits them
- ≥80% coverage on the refactored loader; existing PI1-yaa integration tests under `gateway/integration/` pass unchanged (regression gate)
library_ref: ADR PI1-yaa-0001, ADR PI2-yaa-0001
depends_on: [WI-2yaa.PLG-1]

### WI-2yaa.PLG-6: Plugin chain handler + per-route overrides + reverse-Shutdown [DONE] — Sprint 2
service: yaagents/gateway
parent_feature: F-PLUGIN
brief: Implement the request-time plugin chain composition (PRD §6.4 per
request): walk registered + YAML-enabled plugins in **declaration order**;
compose `Handler(next)` with the innermost handler closest to upstream.
Per-route plugin overrides (PRD §5.4.2): a route's `plugins:` block may set
`enabled: false` on any plugin **except `token-validator`** (gateway boots
exit 1 if a route disables `token-validator`); the override is applied
when building the per-route chain. **SIGTERM Shutdown**: on signal, call
`plugin.Shutdown(ctx)` in **reverse declaration order**; `ctx` deadline =
`time.Now() + GATEWAY_SHUTDOWN_TIMEOUT_S` (default 30 s); drain in-flight
requests; then `os.Exit(0)`. Add `X-YAAgents-Profile: v0.2` response
header to all proxied responses (PRD §5.4 + §4 profile bump). Health
routes (`/healthz`, `/readyz`, `/metrics`) MUST be **pre-auth** — they
bypass the plugin chain entirely (PRD §10 `[SEC]` Gateway).
acceptance:
- 4-plugin chain (token-validator + tenant-injector + license-check + a test community plugin) executes in declaration order — verified via test plugin that records its invocation index in each `Handler`; chain assertion checks the order matches YAML
- Per-route override: a route with `plugins: {tenant-injector: {enabled: false}}` bypasses tenant-injector for that route only; other routes still run it
- Per-route override that sets `token-validator: {enabled: false}` → gateway exit 1 at boot
- SIGTERM: 3-plugin chain Shutdown sequence verified in reverse order (test plugins record their Shutdown invocation timestamps; assertion checks reverse-monotonic)
- SIGTERM drain: requests in-flight at signal time complete within the deadline; new requests rejected with 503 after signal
- `X-YAAgents-Profile: v0.2` present on every proxied response
- `/healthz`, `/readyz`, `/metrics` reachable with NO Authorization header (verified in integration test) — they MUST NOT pass through the plugin chain
- ≥80% coverage on chain handler
library_ref: ADR PI1-yaa-0001, ADR PI2-yaa-0001
depends_on: [WI-2yaa.PLG-2, WI-2yaa.PLG-3, WI-2yaa.PLG-4, WI-2yaa.PLG-5, WI-2yaa.PLG-7]

---

## LLM specialisation (PRD §7 — Option A layer-atop per ADR PI2-yaa-0002)

Per ADR PI2-yaa-0002 (amended 2026-05-30T19:30Z — user-direct MOVE not
COPY) the LLM-specific concerns from `ai-platform/services/ai-gateway/`
are **MOVED** into `yaagents/gateway/internal/llm/` (SSE proxy +
concurrency limiter + execution-timeout) **and** as a small CORS plugin
under `gateway/internal/plugins/cors/` (community-promotable later). The
ai-platform-side `services/ai-gateway/` is **deleted** in the same PI
(cross-lane direction at chief-architect — see ADR PI2-yaa-0002
Addendum). No production consumers exist; **breaking changes are
welcome**, so yaagents-canonical naming is authoritative (no
backward-compat shims for old route paths, config field names, env-var
names, or Keycloak-realm-aware JWKS construction).

Activation is **config-driven**: a route declares `mode: sse` in
`routes.yaml` to opt into SSE pipe-and-flush; standard JSON routes pay no
cost (the SSE code path is dormant when no route activates it — PRD §7.1
Option A trade-off accepted).

Activation is **config-driven**: a route declares `mode: sse` in
`routes.yaml` to opt into SSE pipe-and-flush; standard JSON routes pay no
cost (the SSE code path is dormant when no route activates it — PRD §7.1
Option A trade-off accepted).

### WI-2yaa.LLM-1: SSE proxy (pipe-and-flush) [DONE] — Sprint 4
service: yaagents/gateway/internal/llm
parent_feature: F-LLM
brief: **MOVE** (not copy) `ai-platform/services/ai-gateway/internal/proxy/sse.go`
and `internal/proxy/proxy.go` into `yaagents/gateway/internal/llm/sse.go`
and `yaagents/gateway/internal/llm/proxy.go` (rename package to `llm`).
Each moved file carries the standard SPDX header per LIC-2 only:
```
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds
```
(Per ADR PI2-yaa-0002 amendment 2026-05-30T19:30Z the attribution comment
is dropped — git history of the ai-platform-side `git rm` commit
preserves the lineage pointer.) Add a `routes.yaml` field `mode: sse`
(string; default `""` = standard JSON proxy). When `mode: sse` +
`Accept: text/event-stream` present on the request → use the SSE proxy
path (pipe-and-flush semantics; flush per upstream chunk; preserve
`Content-Type: text/event-stream`). **Breaking changes welcome**:
rename internal symbols / config knobs to yaagents-canonical form
freely; no backward-compat shims for the old ai-platform names.
acceptance:
- Two attribution-headered files exist under `gateway/internal/llm/` with the originating commit hash recorded
- A `routes.yaml` route with `mode: sse` + an upstream that emits an SSE stream → progressive `text/event-stream` response observed at the client (verified with `bufio.Scanner` reading chunks)
- Mode default behaviour unchanged: standard JSON routes path through `httputil.ReverseProxy` exactly as PI1-yaa GW-4
- ≥75% coverage on the SSE proxy path
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.PLG-6]

### WI-2yaa.LLM-2: Per-tenant SSE concurrency limiter [DONE] — Sprint 4
service: yaagents/gateway/internal/llm
parent_feature: F-LLM
brief: Per-tenant counter (sync.Map[string]*atomic.Int64 keyed by tenantID
from `X-Tenant-ID` post-tenant-injector). Config: top-level
`gateway.llm.max_sse_connections_per_tenant` (default 10). On request entry
to an SSE route: increment counter; if new value > limit → decrement +
return `429 application/vnd.yaagents.error+json` body with `retryAfter:
60` field. On response stream completion (or client disconnect): decrement
counter (defer in the SSE handler). The limiter runs **inside** the LLM
proxy path (NOT a plugin — it is intrinsic to SSE accounting).
acceptance:
- 11th concurrent SSE request from `tenant-001` (limit 10) → 429 vendor-error body; counter remains at 10
- Counter decrements on client disconnect (test: open SSE, close client, fire one more — accepted)
- Counter does NOT decrement twice (e.g. server-side stream end + client disconnect should not double-decrement)
- Non-SSE requests do NOT touch the counter (verified with metrics)
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.LLM-1]

### WI-2yaa.LLM-3: Execution timeout + CORS plugin [DONE] — Sprint 4
service: yaagents/gateway/{internal/llm,internal/plugins/cors}
parent_feature: F-LLM
brief: **Execution timeout**: per-route field `executionTimeoutSeconds`
(int; default 0 = no timeout). When set, the gateway wraps the request
context with `context.WithDeadline(ctx, time.Now() + N*time.Second + 30*time.Second)`
for SSE routes (PRD §7.1: `SSEReadTimeout = ExecutionTimeoutSeconds + 30
s`); for non-SSE routes the deadline is exactly `N*time.Second`. On
deadline exceeded → return `500 application/vnd.yaagents.error+json` with
`code: "EXECUTION_TIMEOUT"`. **CORS plugin** under
`gateway/internal/plugins/cors/`: implements `Plugin` interface (PRD §6.2);
config: `allowed_origins: []string` (exact-match; default empty disables
CORS handling); responds to OPTIONS preflight with appropriate CORS
headers; passes non-OPTIONS requests to `next` after adding
`Access-Control-Allow-Origin` if origin matches. init() →
plugin.Register(&CORSPlugin{}).
acceptance:
- SSE route with `executionTimeoutSeconds: 5` + upstream that stalls 10 s → 500 vendor-error body with `code: EXECUTION_TIMEOUT` at ~35 s (5 + 30 SSE read budget)
- Non-SSE route with `executionTimeoutSeconds: 5` + upstream stall 10 s → 500 at ~5 s
- CORS plugin with `allowed_origins: [https://app.example.com]` + OPTIONS preflight from that origin → 200 with `Access-Control-Allow-Origin: https://app.example.com`
- CORS plugin with mismatched origin → 200 OPTIONS but NO `Access-Control-Allow-Origin` header
- CORS plugin disabled (`allowed_origins: []`) → OPTIONS passes through to `next` unchanged
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.LLM-1, WI-2yaa.PLG-1]

### WI-2yaa.LLM-4: SSE Prometheus metrics [READY] — Sprint 4
service: yaagents/gateway/internal/llm
parent_feature: F-LLM
brief: **MOVE** `ai-platform/services/ai-gateway/internal/metrics/sse_proxy.go`
into `yaagents/gateway/internal/llm/metrics.go` (rename package). Standard
SPDX header per LIC-2; no attribution comment (per ADR PI2-yaa-0002
amendment). Metric names are yaagents-canonical (breaking change from
ai-platform — no `ai_gateway_*` prefix carry-forward). Expose two
metrics on `/metrics`: `yaagents_gateway_sse_connections_active` (gauge, labels:
`tenant_id`, `route_id`) and `yaagents_gateway_sse_errors_total` (counter,
labels: `tenant_id`, `route_id`, `error_kind` ∈ `client_disconnect` /
`upstream_error` / `timeout` / `limit_exceeded`). Wire counter increments
into LLM-1 + LLM-2 + LLM-3 paths. Existing `/metrics` (PI1-yaa GW-5
request/latency/status histograms) preserved unchanged.
acceptance:
- `/metrics` includes both new SSE metrics in Prometheus text format
- `yaagents_gateway_sse_connections_active{tenant_id="tenant-001", route_id="completions"}` rises and falls with concurrent SSE streams (verified in integration test)
- `yaagents_gateway_sse_errors_total{error_kind="limit_exceeded"}` increments on LLM-2 rejection (verified)
- ≥75% coverage on the metrics module
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.LLM-2, WI-2yaa.LLM-3]

---

## Profile version bump (cross-component touchpoint that lives here)

### WI-2yaa.BUMP-3: Profile v0.2 header + spec/schema path bump [READY] — Sprint 5
service: yaagents/{gateway,spec,schemas,openapi}
parent_feature: F-LICENSE
brief: Update the gateway's response-header injector (PLG-6) to write
`X-YAAgents-Profile: v0.2` (PI1-yaa wrote `v0.1`). Move JSON schemas from
`schemas/v0.1/` to `schemas/v0.2/` (PRD §5.2 — forward-compat path) — keep
`schemas/v0.1/` in-place (frozen artefacts for v0.1.x consumers; do not
delete). Add SSE-streaming contract addendum to `spec/` (PRD §5.1: profile
v0.2 normative addition — SSE streaming over plugin chain; reference §7
LLM Gateway). Update `openapi/yaagents-components.yaml` profile version to
v0.2. **Single commit** so the bump is atomic across the contract layer.
acceptance:
- `grep -rn "v0\\.1" spec/ openapi/ --include='*.md' --include='*.yaml'` returns only `schemas/v0.1/`-related references (the frozen path is intentional)
- `schemas/v0.2/` contains all 6 schemas from PRD §5.2 with `$id` updated to the `v0.2/` path
- Gateway response sends `X-YAAgents-Profile: v0.2` (verified by integration test)
- `openapi/yaagents-components.yaml` declares profile v0.2
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.PLG-6]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] authn | feature WI | PLG-3 (token-validator: HS256/JWKS; 403 on failure; always-on enforcement) |
| [SEC] RBAC enforcement | feature WI | PLG-6 (per-route override; token-validator exit-1 if disabled per route) |
| [SEC] no secret in image/config | **NFR WI below** | WI-2yaa.NFR-GW-1 |
| [SEC] govulncheck + trivy on Go image | **NFR WI below** | WI-2yaa.NFR-GW-2 |
| [SEC] /healthz + /readyz pre-auth | feature WI | PLG-6 (explicitly: health routes bypass plugin chain; integration test asserts 200 with no Authorization) |
| [SRE] /healthz + /readyz | feature WI | PLG-6 (carries forward PI1-yaa GW-5; always reachable) |
| [SRE] structured JSON logs + correlation_id | feature WI | PLG-2 (plugin loader propagates request context); PLG-6 (chain integration test inspects per-plugin log lines) |
| [SRE] graceful shutdown | feature WI | PLG-6 (SIGTERM → reverse-Shutdown; 30 s deadline default; deadline configurable via GATEWAY_SHUTDOWN_TIMEOUT_S) |
| [SRE] per-plugin latency metric (advisory) | feature WI | LLM-4 (`yaagents_gateway_plugin_latency_seconds`, label `plugin_name`) |
| [SUPPLY] multi-arch image (amd64 + arm64) | feature WI | REL-3 (`docker/build-push-action` multi-arch; GHCR `:0.2.0` + `:latest`) |
| [SUPPLY] SBOM (Syft SPDX 2.3 JSON) | feature WI | REL-3 (carry-forward from PI1-yaa OQ-5; attached to GitHub Release) |
| [SUPPLY] OIDC trusted publishing (no long-lived token) | feature WI | REL-3 (`GITHUB_TOKEN` OIDC; no PAT) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI2-yaa; GHCR is free for public images |

### WI-2yaa.NFR-GW-1: No secrets in gateway image/config [DONE]
service: yaagents/gateway
parent_feature: F-PLUGIN
brief: [SEC] Enforce secret hygiene in the gateway Dockerfile and default
config. `docker/gateway/Dockerfile` MUST NOT contain any `ENV` instruction
that sets a secret value (JWT secret, API key, password).
`GATEWAY_JWT_SECRET` is a **runtime env-var only** — never a Dockerfile
default. The image's embedded config (if any) MUST NOT supply a functional
`GATEWAY_JWT_SECRET`; `GATEWAY_JWT_JWKS_URL` is the production auth path
(ADR PI1-yaa-0001 §3; carries forward to v0.2.0). No `.env` file
committed to the repo. CI step greps for `ENV *SECRET*` / `ENV *TOKEN*`
literals in any Dockerfile under `docker/gateway/` and FAILs on any hit.
acceptance:
- `docker inspect ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` Env list contains no `GATEWAY_JWT_SECRET=*` default
- `trivy config docker/gateway/Dockerfile` exits 0 (no secret in Dockerfile layer)
- CI grep `grep -nE "ENV.*(SECRET|TOKEN|PASSWORD).*=" docker/gateway/Dockerfile` returns 0 hits (step fails on any match)
- No `.env` file in repo (`git ls-files "*.env" ".env*"` returns empty)
library_ref: ADR PI1-yaa-0001 (secret hygiene; carries forward)
depends_on: [WI-2yaa.PLG-2]

### WI-2yaa.NFR-GW-2: govulncheck + trivy CI gate on gateway binary + image [READY]
service: yaagents/gateway
parent_feature: F-PLUGIN
brief: [SEC + SUPPLY] Two CI checks on the v0.2.0 gateway artefact:
(1) **`govulncheck`**: run `govulncheck ./gateway/...` on every PR + main
push; target 0 HIGH/CRITICAL findings. The gateway module includes the
plugin chain, the LLM specialisation (LLM-1..4), and the token-validator
JWKS client — all in scope.
(2) **`trivy image`**: run `trivy image --severity HIGH,CRITICAL
ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` in the publish workflow
(REL-3 pipeline step); FAIL publish on any HIGH/CRITICAL hit. Scan is
against the pushed multi-arch manifest; amd64 slice is the reference
scan. The Alpine base layer is expected clean; any base-layer HIGH finding
triggers a base-image bump WI in the next sprint before publish proceeds.
acceptance:
- CI step `govulncheck-gateway` added; runs on every PR + push to main; exits 1 on HIGH/CRITICAL
- REL-3 publish workflow carries a `trivy-scan` step before the final tag push; workflow exits 1 on any HIGH/CRITICAL finding
- `govulncheck ./gateway/...` passes on the v0.2.0 tagged commit (0 findings)
- `trivy image ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` passes at REL-3 (0 HIGH/CRITICAL)
library_ref: ADR PI1-yaa-0001 (net-new base carries forward), ADR PI2-yaa-0002
depends_on: [WI-2yaa.LLM-4, WI-2yaa.REL-3]
