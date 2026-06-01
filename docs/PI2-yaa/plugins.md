# PI2-yaa — Component: Plugin middleware system (`gateway/plugin/` + `gateway/internal/plugins/*`)

Owner lane: **go-developer**. Sprints 1–2. ADR: PI2-yaa-0001 (plugin
interface contract + versioning + registration semantics), PI2-yaa-0005
(token-validator JWKS sourcing + prompt-sanitize stub behaviour).

> **Library gate (mandatory citation):** all WIs in this file carry
> `library_justify: novel plugin-middleware abstraction; no portfolio shared
> library applies (the plugin interface IS the abstraction being authored).`
> Exception: PLG-3 token-validator carries an additional explicit
> `library_justify` clause about `portfolio/packages/go/auth-jwks/` (see WI).

---

### WI-2yaa.PLG-1: Plugin interface + registry + PluginConfig accessor [READY] — Sprint 1
service: yaagents/gateway/plugin
parent_feature: F-PLUGIN
brief: Create Go package `github.com/ai-mpathyminds/yaagents/gateway/plugin`.
Define the `Plugin` interface verbatim per PRD §6.2 (`Name()`, `Init(cfg
PluginConfig) error`, `Handler(next http.Handler) http.Handler`,
`Shutdown(ctx context.Context) error`). Define `PluginConfig` interface
verbatim (`GetString`, `GetBool`, `GetStringSlice`, `GetInt`, `Raw()
map[string]any`). Define global registry: `Register(p Plugin)` (idempotent
on `Name()` — second Register with same name → error log + skip; documented
"first wins"); `Registered() []Plugin` returning the slice in registration
order (NOT declaration order — declaration order is the YAML's job; plugins
register at process-init time and a registered plugin not in the YAML is
silently inert). The `PluginConfig` impl is a thin wrapper around a
`map[string]any` parsed from the plugins-YAML block. Per ADR PI2-yaa-0001,
the `Plugin` interface does **NOT** carry a `Version() string` method;
versioning is by Go module tag (`gateway/plugin/v0.2`). Per ADR
PI2-yaa-0001 §3, registration is import-side-effect via `init()`; the
gateway will NEVER call `plugin.Open()` or `dlopen` (security floor — see
PRD §10 `[SEC]`).
acceptance:
- Package compiles; `golangci-lint` clean; godoc-style comments on every exported symbol
- Unit test: two `Register()` calls with same `Name()` → second is skipped + warn-logged; both plugins-list slices remain stable
- Unit test: `PluginConfig.GetString("missing")` returns `""`; `GetBool("missing")` returns `false`; type coercion errors return zero-value + warn log
- ≥85% line coverage on the `plugin` package (registry + config impl)
library_justify: novel plugin-middleware abstraction; no portfolio shared library applies (the plugin interface IS the abstraction being authored).
depends_on: [WI-2yaa.LIC-1]

### WI-2yaa.PLG-3: Plugin (a) — `token-validator` (always-on; JWT RS256/JWKS + HS256-test) [READY] — Sprint 2
service: yaagents/gateway/internal/plugins/token-validator
parent_feature: F-PLUGIN
brief: Implement `token-validator` plugin (PRD §6.5 plugin a). JWT RS256
validation via JWKS URL (`jwks_url` config; `cache_ttl_seconds` default 600,
in-memory cache with stale-while-revalidate fallback). Dev/test mode: HS256
via `jwt_secret` (only when `test_mode: true`). Optional `audience`
validation (skip if empty). On validation failure: return
`403 application/vnd.yaagents.error+json` from the `Handler` immediately —
do NOT call `next`. Per ADR PI2-yaa-0005 (verified at A-3),
`portfolio/packages/go/auth-jwks/` does NOT exist; **re-implement JWKS
fetch + RS256 validation minimally inline** under
`gateway/internal/plugins/token-validator/jwks.go`. Plugin `Init` signature
is stable so the import path can switch in a future PI when the extraction
lands. The plugin MUST `init()` → `plugin.Register(&TokenValidator{})`.
**Always-on enforcement**: this plugin's `Init` returns error if `cfg.GetBool("enabled")` is `false` — the gateway core also asserts this at boot (PLG-6), but the plugin itself rejects too (defence-in-depth).
acceptance:
- HS256 happy-path + JWKS happy-path unit-tested; tampered/expired token → 403 vendor-error body with `trace.correlationId` populated
- JWKS cache: first request triggers HTTP GET to `jwks_url`; subsequent requests within `cache_ttl_seconds` reuse cached keys; verified via test HTTP server hit-count assertion
- `Init` returns non-nil error when `cfg.GetBool("enabled") == false`
- `Init` returns non-nil error when `test_mode: true` but `jwt_secret` is empty
- `Handler` does NOT call `next` on token validation failure (verified by next-handler call-counter)
- ≥85% line coverage on `token-validator` package
library_justify: portfolio/packages/go/auth-jwks/ extraction NOT yet landed (verified at A-3 — portfolio/packages/go/ contains only README.md; LIBRARIES.md row re-targeted PI14-oppor; oppor/docs/PI14-oppor/ absent). Per ADR PI2-yaa-0005 re-implement minimally inline. Plugin Init signature stable so import-path can switch when extraction lands. Open extraction row tracked in LIBRARIES.md.
depends_on: [WI-2yaa.PLG-1, WI-2yaa.PLG-2]

### WI-2yaa.PLG-4: Plugin (b) — `tenant-injector` [READY] — Sprint 2
service: yaagents/gateway/internal/plugins/tenant-injector
parent_feature: F-PLUGIN
brief: Implement `tenant-injector` plugin (PRD §6.5 plugin b). Read tenant
ID from the configured `header` (default `X-Tenant-ID`); if `allowlist` is
non-empty and the tenant ID is not in the list → return
`403 application/vnd.yaagents.error+json`. Inject the configured
`inject_header` (default `X-Actor-Tenant`) into the **upstream** request
(modify `r.Header` before calling `next.ServeHTTP`). Default-on; operator
may disable per-gateway or per-route (per-route override handled by PLG-6).
init() → plugin.Register.
acceptance:
- Allowlist empty → all tenants accepted; allowlist non-empty + tenant in list → accepted; tenant NOT in list → 403 vendor-error body
- `inject_header` appears in upstream request (verified by test upstream that echoes `r.Header`)
- `Init` validates: `inject_header` non-empty; `header` non-empty
- ≥85% line coverage on `tenant-injector` package
library_justify: novel plugin-middleware abstraction; no portfolio shared library applies.
depends_on: [WI-2yaa.PLG-1]

### WI-2yaa.PLG-5: Plugin (c) — `license-check` [READY] — Sprint 2
service: yaagents/gateway/internal/plugins/license-check
parent_feature: F-PLUGIN
brief: Implement `license-check` plugin (PRD §6.5 plugin c). Read a
product-license token from the configured `header` (default
`X-License-Token`); call `license_url` to verify (with `cache_ttl_seconds`
default 300 response caching keyed on the token string — bounded cache size,
LRU eviction). On verification failure (HTTP non-2xx from `license_url`, or
network error after retry budget) → return `403
application/vnd.yaagents.error+json`. Default-on; operator may disable
per-gateway or per-route. **HTTP client**: use a `http.Client` with a hard
timeout (default 5 s); plugin Init validates `license_url` is a parseable
URL. init() → plugin.Register.
acceptance:
- Valid token + `license_url` returns 2xx → request passes through to `next`
- Invalid token + `license_url` returns 4xx → 403 vendor-error body; `next` NOT called
- `license_url` timeout → 403 with `dependency: "license-server"` field in trace
- Cache: same token requested twice within TTL → only one outbound HTTP call (verified by hit-counter)
- ≥80% line coverage on `license-check` package
library_justify: novel plugin-middleware abstraction; no portfolio shared library applies.
depends_on: [WI-2yaa.PLG-1]

### WI-2yaa.PLG-7: Stub plugins (d) `prompt-sanitize` + (e) `otel-audit` [READY] — Sprint 2
service: yaagents/gateway/internal/plugins/{prompt-sanitize,otel-audit}
parent_feature: F-PLUGIN
brief: Ship the `Plugin` interface implementations for both **stub** plugins
per PRD §6.5 plugins d & e. **prompt-sanitize**: `Init` validates (no-op
beyond storing config); `Handler` is pass-through; emits a `warn:
prompt-sanitize is a stub; full prompt-injection defence deferred to
PI3-yaa or community` log line **once** on first request when `enabled:
true` (sync.Once guard). Per ADR PI2-yaa-0005 OQ-7 resolution: the stub
does **NOT** exit 1 on `enabled: true` — log-and-pass-through is the chosen
behaviour. **otel-audit**: `Init` validates `endpoint` is a parseable URL
if non-empty (returns error otherwise); `Handler` emits a no-op span
placeholder (uses `go.opentelemetry.io/otel/trace/noop` — already in
standard `otel` go-imports if available, else hand-rolled noop tracer);
emits a `warn: otel-audit exporter not configured` log line **once** on
first request when `enabled: true` AND `endpoint` is empty. Both plugins
init() → plugin.Register.
acceptance:
- `prompt-sanitize` with `enabled: true` passes every request through to `next`; warn log emitted exactly once across N concurrent requests (sync.Once verified)
- `otel-audit` with `enabled: true` + `endpoint: ""` → warn-once, request passes through
- `otel-audit` with `enabled: true` + malformed `endpoint` → `Init` returns non-nil error (gateway exit 1)
- Both plugins compile clean; ≥70% line coverage (stubs are intentionally thin)
library_justify: novel plugin-middleware abstraction; stubs deferred-impl per PRD §6.5 + §12 (PI3-yaa or community).
depends_on: [WI-2yaa.PLG-1]

---

## Plugin chain handler (PLG-2 + PLG-6 belong in gateway.md)

The plugin **runtime** (Init lifecycle, chain composition, per-route override
application, reverse-Shutdown on SIGTERM, always-on assertion that
`token-validator.enabled == true`) lives in the gateway-core component and
is enumerated in `gateway.md` as WIs PLG-2 (loader refactor) and PLG-6
(chain handler + per-route override + Shutdown).

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] plugin sandboxing (no dynamic loading) | feature WI | PLG-1 (import-side-effect registration; `plugin.Open`/`dlopen` explicitly forbidden per ADR PI2-yaa-0001 §3) |
| [SEC] govulncheck on plugin packages | **NFR WI below** | WI-2yaa.NFR-PLG-1 |
| [SEC] plugin.Open/dlopen CI grep gate | **NFR WI below** | WI-2yaa.NFR-PLG-1 |
| [SRE] per-plugin structured logs + correlation_id | feature WI | PLG-6 (chain integration test asserts corr_id propagation through each plugin) |
| [SRE] per-plugin latency histogram | feature WI | gateway.md LLM-4 (advisory; `yaagents_gateway_plugin_latency_seconds`) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; plugins are compiled-in Go code with zero cloud cost implication |

### WI-2yaa.NFR-PLG-1: govulncheck + no-dynamic-load CI gate [READY]
service: yaagents/gateway/internal/plugins
parent_feature: F-PLUGIN
brief: [SEC] Two CI checks wired in `.github/workflows/ci.yml`:
(1) **`govulncheck`**: run `govulncheck ./gateway/internal/plugins/...` on
every PR + main push; target 0 HIGH/CRITICAL findings. This covers the
token-validator JWKS fetch path (`gateway/internal/plugins/token-validator/`),
which is the only plugin with an outbound HTTP dependency.
(2) **`plugin.Open`/`dlopen` grep gate**: `grep -rn "plugin\.Open\|dlopen"
gateway/internal/plugins/ gateway/plugin/` returns 0 hits; CI FAILs on any
match. Community plugins arrive as Go module imports (compiled-in); no
dynamic loading ever permitted (ADR PI2-yaa-0001 §3 + PRD §10 `[SEC]`).
acceptance:
- CI step `govulncheck-plugins` added; runs `govulncheck ./gateway/internal/plugins/...`; exits 1 on any HIGH/CRITICAL (verified: introduce a dep with a known vuln on a branch → CI FAIL → revert)
- CI step `no-dynamic-load-scan` grep exits 0 on the clean codebase; exits 1 when `plugin.Open` is injected into a test file (verified in CI dry-run)
- `govulncheck` passes on the v0.2.0 tagged commit
library_ref: ADR PI2-yaa-0001, ADR PI2-yaa-0005
depends_on: [WI-2yaa.PLG-1, WI-2yaa.PLG-3, WI-2yaa.PLG-4, WI-2yaa.PLG-5, WI-2yaa.PLG-7]
