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

### WI-2yaa.PLG-1: Plugin interface + registry + PluginConfig accessor [DONE] — Sprint 1
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

### WI-2yaa.PLG-3: Plugin (a) — `token-validator` (always-on; JWT RS256/JWKS + HS256-test) [DONE] — Sprint 2
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
note: **EXTENDED (not superseded) by PLG-3b per ADR PI2-yaa-0007.** v1 is a valid single-issuer/single-audience foundation; PLG-3b is additive — multi-issuer, multi-audience, algorithm allowlist, RFC-correct 401 codes, propagate-claims contract, configurable token header, clock skew, required claims, token size cap. v1 code at `gateway/internal/plugins/token-validator/` and tests in `plugin_test.go` are preserved; PLG-3b extends the same package.

### WI-2yaa.PLG-3b: Plugin (a) — `token-validator` hardening for generic OSS deployments [WIP] — Sprint 2
service: yaagents/gateway/internal/plugins/token-validator
parent_feature: F-PLUGIN
brief: Extend `token-validator` v1 (PLG-3, commit 442cece) per ADR
PI2-yaa-0007 with 9 production-grade configuration surfaces. PLG-3b is
**additive** — v1 code at `gateway/internal/plugins/token-validator/` is
preserved and existing tests must continue to pass. The amendments make
the plugin generic for multi-IdP OSS deployments and close
algorithm-confusion attack vectors at the plugin level.

The 9 amendments:

1. **Multi-issuer + multi-JWKS** — replace singular `jwks_url` with an
   `issuers: [{issuer, jwks_url, jwks_cache_ttl_seconds}]` list. Validator
   reads `iss` from the token (unverified pre-check), matches against the
   list, validates with the matched entry's JWKS. Empty list with
   `test_mode: false` is an Init error.
2. **Algorithm allowlist** — `algorithms: [RS256, ES256]` default. `none`
   forbidden; Init refuses to start if `none` appears. HS256 only when
   `test_mode: true`.
3. **Multi-audience** — `audiences: ["a", "b"]` list (was single string).
   Token's `aud` must match at least one. Empty list = audience validation
   skipped (v1 behaviour preserved).
4. **Clock skew tolerance** — `clock_skew_seconds: 60` default; applied
   to `exp` / `nbf` / `iat` validation.
5. **Required claims** — `required_claims: ["sub"]` default. Listed
   claims MUST be present and non-empty.
6. **Propagate-claims contract** — `propagate_claims.mode: all | allowlist`.
   Explicit contract for which validated claims land in request context
   for downstream plugins (PLG-4b depends on this).
7. **Configurable token header** — `token.header: "Authorization"` +
   `token.scheme: "Bearer"`. Default unchanged; configurable for
   `X-API-Token` / `X-Auth-Token` / gateway-to-gateway flows.
8. **RFC-correct status codes** — `on_failure.*` map; defaults flip from
   `403` to `401` for every credential-validation failure (per RFC 7235).
   Operators can override to preserve v1's `403` if downstream depends on it.
9. **Token size cap** — `max_token_bytes: 8192` default. Refuse oversized
   tokens with `400` BEFORE parse, preventing memory-amplification.

Extended plugin YAML schema:

```yaml
token-validator:
  enabled: true                                # REQUIRED true (v1 invariant)

  # (1) multi-issuer
  issuers:
    - issuer: "https://iam.aimpathyminds.com"
      jwks_url: "https://iam.aimpathyminds.com/.well-known/jwks.json"
      jwks_cache_ttl_seconds: 600
    - issuer: "https://login.microsoftonline.com/tenant-id/v2.0"
      jwks_url: "https://login.microsoftonline.com/tenant-id/discovery/v2.0/keys"
      jwks_cache_ttl_seconds: 3600

  # (2) algorithm allowlist — "none" FORBIDDEN; HS256 only in test_mode
  algorithms: [RS256, ES256]

  # (3) multi-audience
  audiences: ["yaagents-gateway"]              # empty = skip audience check

  # (4) clock skew
  clock_skew_seconds: 60

  # (5) required claims
  required_claims: ["sub"]

  # (6) propagate-claims contract
  propagate_claims:
    mode: all                                  # all | allowlist
    claims: []                                 # required when mode: allowlist

  # (7) configurable token header
  token:
    header: "Authorization"
    scheme: "Bearer"                           # "" = no prefix to strip

  # (8) RFC-correct status codes
  on_failure:
    missing_token:           401
    invalid_signature:       401
    expired:                 401
    not_yet_valid:           401
    unknown_issuer:          401
    audience_mismatch:       401
    required_claim_missing:  401
    disallowed_algorithm:    401
    oversized_token:         400
    jwks_unavailable:        503               # cold-start: all JWKS unreachable

  # (9) token size cap
  max_token_bytes: 8192

  # test-mode HS256 (v1 — preserved)
  test_mode: false
  jwt_secret: ""                               # required when test_mode: true
```

Init validation (additions on top of v1 — return non-nil error → gateway exit 1):

1. `issuers` non-empty when `test_mode: false`.
2. Every `issuers[].issuer` non-empty AND `issuers[].jwks_url` parseable URL.
3. `algorithms` non-empty AND does NOT contain `none`.
4. `clock_skew_seconds` >= 0 AND <= 600.
5. `propagate_claims.mode` ∈ {all, allowlist}; when `allowlist`, `claims` non-empty.
6. `token.header` non-empty.
7. `max_token_bytes` > 0 AND <= 65536.
8. v1 invariants preserved: `enabled: true` required; `test_mode: true` requires non-empty `jwt_secret`.

**Backwards compatibility**: a v1 config (`jwks_url` + `audience` singular) MUST continue to load — PLG-3b interprets singular `jwks_url` as `issuers: [{issuer: "", jwks_url: <val>, jwks_cache_ttl_seconds: cache_ttl_seconds}]` and singular `audience` as `audiences: [audience]` when non-empty. This shim emits `WARN: token-validator config uses v1 single-issuer form; please migrate to issuers: list per ADR PI2-yaa-0007` at boot.

acceptance:
- Multi-issuer: 3 issuers configured; token with `iss` matching #1 → validated against #1's JWKS; token with `iss` matching #2 → validated against #2's JWKS; token with `iss` matching none → 401
- Algorithm allowlist: token with `alg: none` → 401; token with `alg: HS256` outside test_mode → 401; Init returns error when `algorithms: [none, RS256]` configured
- Multi-audience: token `aud: ["a","b"]` matches configured `audiences: ["b","c"]` → pass; token `aud: ["x"]` against `audiences: ["a","b"]` → 401
- Clock skew: token with `exp` 30s in the past + `clock_skew_seconds: 60` → pass; same token with `clock_skew_seconds: 10` → 401 (expired)
- Required claims: `required_claims: ["sub", "tenant_id"]`; token missing `tenant_id` → 401
- Propagate claims: `mode: allowlist`, `claims: ["sub", "email"]`; downstream context contains only `sub` + `email`; other claims (if present in JWT) NOT propagated
- Configurable header: `token.header: "X-Auth-Token"`, `scheme: ""`; request with `X-Auth-Token: <raw-jwt>` (no Bearer prefix) → validated
- Status codes: every failure mode returns the configured code; default suite returns 401 for all credential failures; oversized token returns 400; cold-start JWKS-unavailable returns 503
- Token size cap: 9000-byte token with `max_token_bytes: 8192` → 400 BEFORE JWT parse (verified: no parse-time log entry; no JWT lib invocation in trace)
- v1 backwards compat: existing config with `jwks_url: "..."` + `audience: "..."` loads cleanly + WARN line emitted; behaviour matches v1
- Cold start: ALL `issuers[].jwks_url` unreachable at first request → 503; subsequent successful JWKS fetch → next request 200
- All v1 tests in `plugin_test.go` continue to pass unchanged
- New tests added for each of the 9 amendments; ≥85% line coverage on extended package
library_justify: extending v1 inline (matches PLG-3 v1's library_justify clause). Adds NO new external dependencies (uses stdlib `time` + the existing JWKS HTTP client + the existing JWT lib already in v1).
depends_on: [WI-2yaa.PLG-3]

### WI-2yaa.PLG-4: Plugin (b) — `tenant-injector` v1 (SUPERSEDED by PLG-4b) [DONE] — Sprint 2
service: yaagents/gateway/internal/plugins/tenantinjector
parent_feature: F-PLUGIN
brief: **SUPERSEDED 2026-06-01 by WI-2yaa.PLG-4b per ADR PI2-yaa-0006.**
The v1 design read tenant ID from a client-supplied `X-Tenant-ID` header
gated by an optional allowlist. This conflates tenant admission with
tenant identity proof — a caller with a valid JWT for `tenant-A` can
assert `X-Tenant-ID: tenant-B` and the gateway accepts it. The v1
implementation landed at 2026-06-01 14:03
(`gateway/internal/plugins/tenantinjector/plugin.go`); the code path is
**deleted and replaced** by PLG-4b before B-12 (chain wiring) integrates.
Status `[DONE]` preserved for audit trail; the v1 design is NOT shipped
in v0.2.0.
acceptance: (superseded — see PLG-4b)
library_justify: superseded — see PLG-4b
depends_on: [WI-2yaa.PLG-1]

### WI-2yaa.PLG-4b: Plugin (b) — `tenant-injector` v2 (JWT-derived + HTTP lookup) [DONE] — Sprint 2
service: yaagents/gateway/internal/plugins/tenantinjector
parent_feature: F-PLUGIN
brief: Re-implement `tenant-injector` per ADR PI2-yaa-0006. **Replace**
the v1 client-header trust model with a JWT-claim-derived principal plus
HTTP lookup against a tenant-directory service. Per-request flow: (1)
read configured claim from validated JWT claims (populated by PLG-3
token-validator into request context); (2) call `lookup.url` (HTTP, with
`{principal}` URL-encoded placeholder) to resolve principal → tenant;
(3) parse response per `response.tenant_id_field`; (4) inject derived
tenant into upstream via `inject.tenant_header`. **Strip inbound
`X-Actor-Tenant`** before injection (anti-smuggling — the client must
not be able to smuggle a header the gateway also writes). **Caching**:
per-principal LRU with `golang.org/x/sync/singleflight` to collapse
concurrent first-fetches into a single outbound call; positive +
negative TTL configurable. **Failure modes**: configurable HTTP codes
via `on_failure.*` — defaults 503 (lookup network error / timeout), 403
(principal not found / allowlist miss), 401 (principal claim missing).
**Boot behaviour**: fail-open — gateway boots even if lookup service is
unreachable; per-request failures return the configured status from
cold-start onward. **PI2-yaa scope**: `response.mode: single` only
(one tenant per principal); multi-tenant principals are a v0.3+
enhancement per ADR PI2-yaa-0006. Companion artifact: mock-iam-api stub
binary under `examples/llm-gateway/mock-iam-api/` (~80 LOC Go binary,
single GET endpoint, in-memory map driven by `mock-tenants.yaml`) — wires
the compose demo green without an external IAM dependency.

Default plugin YAML schema:

```yaml
tenant-injector:
  enabled: true
  principal:
    claim: sub                                 # JWT claim naming the principal
  lookup:
    url: "https://iam.internal/api/v1/principals/{principal}/tenant"
    method: GET                                # GET | POST
    timeout_ms: 500
    auth:
      mode: none                               # none | bearer | mtls
      bearer_token_env: TENANT_LOOKUP_TOKEN    # env var, never in YAML; required when mode: bearer
      client_cert_path: ""                     # required when mode: mtls
      client_key_path: ""                      # required when mode: mtls
    headers:                                   # extra headers sent on the lookup call
      X-Gateway-Identity: "yaagents-gateway"
    response:
      mode: single                             # PI2-yaa: single only; multi is v0.3+
      tenant_id_field: "tenant_id"             # JSON field path to the tenant id in the reply
    cache:
      ttl_seconds: 300
      negative_ttl_seconds: 30                 # cache "principal has no tenant" briefly
      max_entries: 10000
      singleflight: true                       # coalesce concurrent first-fetches
  inject:
    tenant_header: "X-Actor-Tenant"
    principal_header: "X-Actor-Principal"      # "" disables principal injection
  allowlist: []                                # post-derivation admission gate; empty = allow all
  on_failure:
    lookup_network_error: 503
    lookup_timeout: 503
    principal_not_found: 403
    claim_missing: 401
```

Lookup-service contract (canonical; documented for OSS adopters):

```http
GET {lookup.url with {principal} URL-encoded and substituted}
Authorization: <per lookup.auth.mode>           # absent when mode: none
X-Gateway-Identity: yaagents-gateway            # plus any other lookup.headers entries

200 OK
Content-Type: application/json
{"principal":"<value>","tenant_id":"tenant-001"}

404 Not Found     → principal_not_found (HTTP 404 from IAM, or 200 with empty tenant_id)
5xx / network err → lookup_network_error
>timeout_ms       → lookup_timeout
```

The v1 code at `gateway/internal/plugins/tenantinjector/` is **deleted
entirely**; re-implemented from scratch. Package name + import path
preserved so PLG-6 (B-12) chain composition sees no API drift. `init()`
→ `plugin.Register(&TenantInjector{})`. Init validation — returns
non-nil error on any violation (gateway exit 1):

1. `principal.claim` non-empty string.
2. `lookup.url` parseable URL AND contains exactly one `{principal}` placeholder.
3. `lookup.method` ∈ {GET, POST}.
4. `lookup.timeout_ms` > 0 and ≤ 30000.
5. `lookup.auth.mode` ∈ {none, bearer, mtls}; when `bearer`, env var resolves to non-empty; when `mtls`, both cert + key paths readable.
6. `lookup.response.mode` == `"single"` (PI2-yaa scope); `tenant_id_field` non-empty.
7. `lookup.cache.ttl_seconds` > 0; `max_entries` > 0.
8. `inject.tenant_header` non-empty.
9. `enabled: false` → Init returns error (defence-in-depth alongside PLG-6 boot assertion; matches PLG-3 token-validator semantics).

acceptance:
- v1 code at `gateway/internal/plugins/tenantinjector/` fully replaced; `grep -rn "X-Tenant-ID" gateway/internal/plugins/tenantinjector/` returns 0 hits (except possibly in inbound-header-stripping logic if the v1 header name is configured-strippable; v2 strips only `inject.tenant_header`)
- JWT with `principal.claim` populated + IAM 2xx → upstream request carries `inject.tenant_header` with value matching `response.tenant_id_field` from lookup (verified via test upstream that echoes `r.Header`)
- JWT with `principal.claim` populated + IAM 404 → 403 vendor-error body; `next` NOT called
- JWT with `principal.claim` populated + IAM timeout → 503 vendor-error body; trace carries `dependency: "iam-lookup"`
- JWT missing the configured principal claim → 401 vendor-error body
- Inbound `X-Actor-Tenant` header is stripped from `r.Header` before plugin handler runs (verified: malicious client sets `X-Actor-Tenant: tenant-evil`; upstream sees only the plugin-derived value, never `tenant-evil`)
- Cache hit: same principal twice within `cache.ttl_seconds` → exactly one outbound HTTP call (hit-counter)
- Singleflight: 50 concurrent first-requests for the same principal → exactly one outbound HTTP call (verified via test IAM hit-counter)
- Negative cache: 404 from IAM cached for `negative_ttl_seconds`; subsequent same-principal request within window returns 403 without re-calling IAM
- Allowlist non-empty + derived tenant NOT in list → 403 vendor-error (post-derivation admission gate)
- Init returns error on each of the 9 validation rules above (one test per rule)
- Boot behaviour: gateway with `lookup.url` pointing at an unreachable host starts cleanly (`/readyz` returns 200); per-request returns 503 from the first call
- `gateway/internal/plugins/tenantinjector/` ≥85% line coverage
- `examples/llm-gateway/mock-iam-api/` compiles; Docker build green; compose demo `up` shows end-to-end principal→tenant→upstream flow green
library_justify: novel tenant-identity-derivation logic; no portfolio shared library applies. Adds `golang.org/x/sync/singleflight` — Go-ecosystem standard, MIT-licensed, single-function utility, widely adopted; `platform-librarian` vet logged at B-11a dispatch time per Library Gate 2.
depends_on: [WI-2yaa.PLG-1, WI-2yaa.PLG-3]

### WI-2yaa.PLG-5: Plugin (c) — `license-check` [DONE] — Sprint 2
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

### WI-2yaa.PLG-7: Stub plugins (d) `prompt-sanitize` + (e) `otel-audit` [DONE] — Sprint 2
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

### WI-2yaa.NFR-PLG-1: govulncheck + no-dynamic-load CI gate [DONE]
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
