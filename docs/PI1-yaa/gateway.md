# PI1-yaa — Component: Go gateway (`gateway/`)

Owner lane: go-developer. Sprint 2. Published as `ghcr.io/yaagents/gateway:0.1.0`
(publish WIs in `release-and-publish.md`). ADR: PI1-yaa-0001 (net-new base,
NOT a consumer/fork of internal gateways; HS256+JWKS dual JWT).

> **Library gate (mandatory citation):** every GW WI carries
> `library_ref: ADR PI1-yaa-0001` — the yaagents gateway is a net-new
> standalone OSS codebase and the future BASE for `platform-services/gateway`
> + `ai-gateway` (reverse dependency, not a consumer). No internal-gateway
> import in either direction during PI1-yaa.

---

### WI-1yaa.GW-1: Gateway skeleton + config + route-YAML loader [DRAFT]
service: yaagents/gateway
brief: Go module `github.com/yaagents/gateway`; `net/http` server (no heavy
framework per ADR 0001 §2). Config from env (`GATEWAY_PORT` default 8080 /
portfolio 8120, `GATEWAY_ROUTES_FILE`, `GATEWAY_AUDIT_LOG`). Parse + validate
`routes.yaml` against the PRD §5.4 schema (id/method/path/target/roles/
tenantRequired/audit). Structured JSON logger with `request_id`/`correlation_id`
fields. Fail-fast on invalid route config at boot.
acceptance:
- `go build` clean, `golangci-lint` clean; invalid routes.yaml → non-zero exit + clear error
- `{param}` path placeholders parsed; duplicate route `id` rejected
library_ref: ADR PI1-yaa-0001 (net-new standalone OSS gateway; future internal base)
depends_on: [WI-1yaa.SPEC-1]

### WI-1yaa.GW-2: Authentication middleware (HS256 dev / JWKS prod) [DRAFT]
service: yaagents/gateway
brief: JWT bearer validation. `GATEWAY_JWT_SECRET` → HS256 (dev/demo default);
`GATEWAY_JWT_JWKS_URL` → RS256 via cached JWKS (prod); JWKS precedence if both
set (warn-log) per ADR 0001 §3. Missing/invalid token → `401` with
`application/vnd.yaagents.error+json` body (trace block populated).
acceptance:
- HS256 happy path + JWKS happy path unit-tested; tampered/expired token → 401 vendor-error body
- No mock JWKS server built (demo uses HS256); ≥80% coverage on auth pkg
library_ref: ADR PI1-yaa-0001
depends_on: [WI-1yaa.GW-1, WI-1yaa.SPEC-2]

### WI-1yaa.GW-3: Tenant/actor context + correlation/request-id injection [DRAFT]
service: yaagents/gateway
brief: Extract `X-Tenant-ID` + actor claims from the validated token; enforce
`tenantRequired` per route (reject `400`/`403` per profile when absent).
Generate `X-Correlation-ID` (UUID) if absent; pass through if present; always
set `X-Request-ID`. Inject tenant/actor/correlation/request headers to upstream.
acceptance:
- Correlation-ID generated-if-absent, preserved-if-present (both unit-tested)
- `tenantRequired:true` route w/o `X-Tenant-ID` rejected with vendor-error body
- actor + tenant forwarded as upstream request headers
library_ref: ADR PI1-yaa-0001
depends_on: [WI-1yaa.GW-2]

### WI-1yaa.GW-4: Route RBAC + typed-response passthrough proxy [DRAFT]
service: yaagents/gateway
brief: Enforce route `roles:` (ALL must be present in actor claims) — failure →
`403 application/vnd.yaagents.error+json`. Reverse-proxy
(`httputil.ReverseProxy`) to route `target:` preserving method + body. **Do not
re-encode** upstream agentic vendor content-types/status (typed-response
passthrough). Add `X-YAAgents-Profile: v0.1` to every proxied response.
acceptance:
- Missing-role request → 403 vendor-error body (not generic 403)
- Upstream `422 application/vnd.yaagents.validation-error+json` passes through byte-identical with profile header added
- Method/body/query preserved through proxy
library_ref: ADR PI1-yaa-0001
depends_on: [WI-1yaa.GW-3]

### WI-1yaa.GW-5: Audit log + health/readiness + metrics [DRAFT]
service: yaagents/gateway
brief: Per-request structured JSON audit event when route `audit:true`
(route id, tenant, actor, status, latency_ms, correlation_id) to
`GATEWAY_AUDIT_LOG` sink (stdout default). `GET /healthz` (liveness),
`GET /readyz` (readiness — config loaded + routes valid), `GET /metrics`
(Prometheus text: request count/latency/status by route). Graceful shutdown.
acceptance:
- Audit event emitted only for `audit:true` routes; one line per request, valid JSON
- `/healthz` 200 always-up; `/readyz` 503 until routes loaded; `/metrics` Prometheus-parseable
- SIGTERM drains in-flight then exits
library_ref: ADR PI1-yaa-0001
depends_on: [WI-1yaa.GW-4]
