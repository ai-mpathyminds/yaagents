# yaagents v0.4 — Detailed PRD
Status: [DRAFT]

> Seeded by chief-architect. Expanded by product-manager 2026-06-05.
> Prior full PRD: `yaagents/system-refs/yaagents-v0.3_detailed.md` [READY]

---

## 1. Product Context

YAAgents is an Agentic REST Profile for building deterministic, resource-oriented agentic APIs.
The core thesis: expose domain resources, not agent invocations.

```
POST /campaigns/{campaignId}/optimizations   ← yaagents
POST /agents/campaign-agent/invoke           ← NOT yaagents
```

v0.4 is the production-readiness increment following v0.3's public launch milestone.
Profile version remains **v0.3**. Package versions bump to **v0.4.x**.

Version lineage:
- v0.1 — Agentic REST Profile + 8 MVP components + Community License
- v0.2 — Plugin middleware system + sdk-go bring-up + Apache 2.0 transition
- v0.3 — Profile v0.3 bump + sdk-go scaffolding + public launch (PyPI/npm/GHCR/Go modules)
- **v0.4** — sdk-go audit-API parity + 5-plugin Stable + README/Pages narrative pivot

---

## 2. Problem

PI3-yaa shipped the runtime and publish substrate. Three production-readiness gaps remain:

1. **sdk-go audit-API gap**: `sdk-go` lacks the Profile v0.3 audit-emission surface that
   `sdk-fastapi` already exposes. AmpathyMinds' Go-native agent migration is blocked until
   sdk-go is a full first-class server SDK.

2. **Gateway plugins at stub/partial maturity**: The 5 first-party plugins (token-validator,
   tenant-injector, license-check, prompt-sanitize, otel-audit) exist as runtime registrations
   but several carry stub bodies — notably prompt-sanitize + otel-audit per PI3-yaa postmortem
   `defer_to_next_pi_seed` INTAKE-2. "Coming in v0.4" markers appear across reference pages.

3. **README narrative opaque to reviewers**: The meta-repo root README relies on A2A vs agentic
   comparison framing. Reviewer feedback (2 external developers per PI3-yaa user retro) reports
   this framing is opaque — readers cannot translate it into "what can I build."

These three are interdependent v0.4 launch-readiness contracts. Partial delivery ships a degraded v0.4.

---

## 3. Why Now

- AmpathyMinds' own Go agent migration is the immediate downstream consumer of sdk-go parity.
- Community adoption is the immediate downstream consumer of plugin-Stable + clear README.
- PI3-yaa shipped the runtime; PI4-yaa makes it production-grade and intelligible.
- v0.5 (Helm chart + bench/load + domain-breadth case studies) depends on v0.4 being solid.

---

## 4. Scope

### In Scope

- **Goal 1**: sdk-go audit-API addition + Profile v0.3 strict-surface conformance gap-close
- **Goal 2**: All 5 plugins Preview→Stable with full implementation + per-plugin Pages docs + e2e tests
- **Goal 3**: README narrative pivot + ecom case study + overview.mdx independent rewrite
- **Track INFRA**: `pi-topology.sh` self-discovery refactor; PRIORITIES mechanical-close trigger; `gh-pi-merge.sh` B-44a regression test
- **Track GH**: GitHub Issues backlog mirror (INTAKE-1)
- **Track DOCS**: `how-to/host-in-production.mdx` (INTAKE-4); `architecture/audit-and-observability.mdx` (INTAKE-5)
- **Track EXAMPLES**: 2 domain-breadth example skeletons (INTAKE-7)
- **Track PROCESS**: Cross-lane co-sign formalization (INTAKE-8); 4 process carry-forwards (resolved at A-1)

### Out of Scope

- AWS substrate or `portfolio/infrastructure/` touch (no cloud-state-grounder dispatch needed)
- Helm chart deployment manifests (PI5-yaa)
- Plugin benchmarking + load-testing (PI5-yaa)
- sdk-go feature-for-feature with sdk-fastapi beyond Profile v0.3 (Python-side conveniences not replicated)
- Headline case study beyond ecom (other domains are skeleton-only INTAKE-7 scope)
- K8s/Kubernetes manifests authoring (PI2-yaa scope; existing manifests not modified)
- GTM content (articles, demo videos, social posts)
- v0.2+ adapter work (Spring Boot, ASP.NET Core, NestJS)

---

## 5. Goal 1 — sdk-go parity with sdk-fastapi

### 5.1 User-Authored Block (verbatim, 2026-06-05)

> **Specific sdk-fastapi feature sdk-go lacks**: **audit API**. sdk-fastapi emits audit events
> via the Profile v0.3 audit surface; sdk-go does not. PI4-yaa Goal 1 closes this specific gap.
> Architect at A-3 verifies via sdk-go ↔ sdk-fastapi diff against the Profile audit-emission contract.
>
> **Parity scope (OQ-1)**: **Strict Profile v0.3 surface only** (user-direct 2026-06-05).
> Python-side conveniences (Pydantic-validation ergonomics, FastAPI decorator sugar) NOT
> replicated unless they're profile-mandatory. Conformance bar is the spec, not sdk-fastapi.

### 5.2 Work Breakdown

| Work Item | Description |
|-----------|-------------|
| SG-01: Audit-API interface design | Define `AuditEmitter` interface in `sdk-go/` matching Profile v0.3 audit-emission contract; ADR PI4-yaa-0001 |
| SG-02: Audit-API implementation | Wire emitter to request lifecycle hooks in `sdk-go/` middleware chain |
| SG-03: Conformance test parity | sdk-go passes same conformance test set as sdk-fastapi against Profile v0.3 audit-emission contract |
| SG-04: ai-platform canary re-run | PI3-yaa AIP-1/AIP-2 canary re-runnable against sdk-go in new shape (optional; decide at A-3) |
| SG-05: Feature matrix doc | `sdk-go/docs/parity-matrix.md` — 1:1 parity with sdk-fastapi OR explicitly-justified omissions |

### 5.3 Acceptance Criteria

- [ ] `sdk-go/` exports an `AuditEmitter` interface with methods matching the Profile v0.3 audit-emission contract
- [ ] Conformance test suite exercises audit-emission e2e against a real gateway + sdk-go service
- [ ] `parity-matrix.md` documents each sdk-fastapi surface item and whether sdk-go has it, with rationale for omissions
- [ ] ai-platform/agent-api canary AIP-1/AIP-2 (if decided IN at A-3): re-runs against sdk-go in new shape and passes
- [ ] **Seed success signal §1**: sdk-go audit-API e2e test exercising Profile v0.3 audit-emission contract passes against a real gateway + sdk-go service

---

## 6. Goal 2 — Gateway plugin production-grade implementation

### 6.1 User-Authored Block (verbatim, 2026-06-05)

> **Phasing (Goal-2 multi-choice)**: **ALL 5 plugins reach Stable in PI4-yaa** (user-direct 2026-06-05).
> token-validator + tenant-injector + license-check + prompt-sanitize + otel-audit all flip
> Preview→Stable with full implementation + per-plugin Pages docs + e2e test exercising through
> the gateway. "Coming in v0.4" markers removed across `production-checklist`, `plugin-pipeline`,
> `reference-architecture`, and per-plugin reference pages.
>
> **OQ-2 Stable bar (default-confirmed)**: per-plugin Stable = (a) feature-complete + tested +
> documented. Benchmarking + load-testing deferred to PI5-yaa.
>
> **Open config schemas**:
> - (a) **token-validator** — **multi-IDP / pluggable JWT issuer**. Single-IDP-per-deployment
>   NOT sufficient. Architect designs the plug-point at A-3.
> - (b) **tenant-injector** — **URL-based webhook resolution**. Gateway calls a configured
>   webhook URL with principal info; webhook returns the tenant ID. User notes "this might have
>   even implemented" — **A-3 architect investigates existing implementation FIRST** (verify what's
>   in `yaagents/gateway/internal/plugins/tenant-injector/` from PI2-yaa PLG-4b v2 work) before
>   authoring close WIs. If implemented but undocumented, the work is doc-only + Stable badge.
>   If partial, close the gaps.
> - (c) **license-check, prompt-sanitize, otel-audit** — user no contract concerns surfaced.
>   Architect at A-3 discovers existing contracts (yaagents/gateway/internal/plugins/*)
>   + authors close WIs against discovered surfaces.

### 6.2 Plugin Contracts (architect discovers at A-3; these are design targets)

#### 6.2.1 token-validator (ADR PI4-yaa-0002)

**Purpose**: Validate incoming JWTs from any configured OIDC/OAuth2 issuer.

Config shape (design target; architect finalises at A-3):
```yaml
token-validator:
  issuers:
    - url: https://auth.example.com/realms/my-realm
      audience: my-service
      jwks_cache_ttl: 300s
    - url: https://accounts.google.com
      audience: my-google-client-id
      jwks_cache_ttl: 600s
  claim_mappings:
    subject: sub
    roles: realm_access.roles
```

Plug-point: pluggable issuer-registry with JWKS-cache per issuer; per-issuer claim mappers.

**Stable criteria**: Multi-IDP config parses; JWKS endpoint fetched per issuer; invalid token returns `403 application/vnd.yaagents.error+json`; e2e test with 2 issuers passes.

#### 6.2.2 tenant-injector (ADR PI4-yaa-0003)

**Purpose**: Resolve tenant ID from principal info via webhook; inject `X-Tenant-ID` header.

Config shape (design target):
```yaml
tenant-injector:
  webhook_url: https://tenant-svc.internal/resolve
  request_envelope:
    subject_claim: sub
    token_header: Authorization
  response_field: tenant_id
  cache_ttl: 60s
  on_failure: reject   # or: pass-through
```

A-3 action: investigate `yaagents/gateway/internal/plugins/tenant-injector/` from PI2-yaa PLG-4b v2 FIRST.

**Stable criteria**: Webhook called with principal; `X-Tenant-ID` injected; failure mode (reject/pass-through) respected; e2e test green.

#### 6.2.3 license-check

**Purpose**: Gate requests against license entitlements for the resolved tenant.

Config shape (design target — architect discovers existing shape at A-3):
```yaml
license-check:
  backend: header   # or: webhook
  header_name: X-License-Tier
  allowed_tiers: [community, professional, enterprise]
  on_unlicensed: 403
```

**Stable criteria**: License header/webhook queried; unlicensed tenant receives `403 application/vnd.yaagents.error+json`; e2e test green.

#### 6.2.4 prompt-sanitize

**Purpose**: Detect and reject or redact potentially harmful content in request payloads before forwarding to backend agentic services.

Config shape (design target):
```yaml
prompt-sanitize:
  strategy: reject   # or: redact
  patterns:
    - name: pii_email
      regex: '[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}'
      action: redact
    - name: injection_attempt
      keywords: ["ignore previous instructions", "jailbreak"]
      action: reject
  on_match_reject_status: 400
```

**Stable criteria**: Configured patterns applied; `reject` returns `400 application/vnd.yaagents.clarification+json` (or `422`); `redact` removes matched content; e2e test green.

#### 6.2.5 otel-audit

**Purpose**: Emit OpenTelemetry-compatible audit spans/events for every gateway-proxied agentic request.

Config shape (design target):
```yaml
otel-audit:
  exporter: stdout   # or: otlp
  otlp_endpoint: http://otel-collector:4317
  attributes:
    - tenant_id: from_header X-Tenant-ID
    - actor: from_claim sub
    - operation: from_route_id
  include_request_body: false   # default: false (PII risk)
```

**Stable criteria**: OTLP/stdout exporter wired; span emitted per request with tenant + actor + operation attributes; e2e test exercising through gateway shows emitted spans; PI4-yaa `architecture/audit-and-observability.mdx` doc covers this plugin.

### 6.3 Acceptance Criteria (all 5 plugins)

- [ ] Each plugin: real implementation (no stub body), full Pages doc page, e2e test exercising through gateway
- [ ] token-validator: multi-IDP config parses; JWKS fetched per issuer
- [ ] tenant-injector: webhook resolution implemented OR existing PI2-yaa PLG-4b v2 impl confirmed + documented
- [ ] All maturity badges flipped Preview→Stable on Pages
- [ ] "Coming in v0.4" markers removed from `production-checklist`, `plugin-pipeline`, `reference-architecture`, per-plugin reference pages
- [ ] **Seed success signal §2**: all 5 plugins ship Stable badge + per-plugin Pages docs + e2e green

---

## 7. Goal 3 — README + Pages narrative pivot

### 7.1 User-Authored Block (verbatim, 2026-06-05)

> **README ↔ overview.mdx coupling (FF-3)**: **Independent for now**. README is
> GitHub-reader-oriented (one-screen, decision-funnel — what is yaagents, what does it let
> you build, why ecom recommendations is a clear example). `start-here/overview.mdx` is
> Pages-reader-oriented (deeper, with sidebar context — same target audience but expects to
> read more). Two separate authoring WIs; each tuned for its surface; no forced same-content
> lockstep.
>
> **OQ-4 case-study consolidation**: **Same content with cross-link**. README ecom-narrative
> IS the canonical case study; the dedicated `case-studies/ecommerce-product-recommendations.mdx`
> page either embeds the same narrative or cross-links to the README. Single source of truth
> between README + case-study page. (This is distinct from the README ↔ overview.mdx split
> above — overview.mdx is its own thing for the Pages reader funnel; the case-study and README
> share content.)

### 7.2 Deliverable Breakdown

| Deliverable | Surface | Content Shape |
|-------------|---------|---------------|
| DOC-01: README rewrite | meta-repo root `README.md` | One-screen GitHub reader. Drop A2A/agentic comparison. Above-the-fold: one-line target-audience callout. Below fold: ecom recommendations use case end-to-end, "who uses yaagents" framing. Links to Pages for deeper reading. |
| DOC-02: overview.mdx rewrite | `start-here/overview.mdx` | Pages-reader-oriented. Sidebar context. Deeper architecture + positioning. Tuned for developer who has landed on the Pages site. Independent of README. |
| DOC-03: ecom case study page | `case-studies/ecommerce-product-recommendations.mdx` | Embeds or cross-links README ecom-narrative. Single source of truth with README §ecom section. Uses `examples/store/` (Python) + `examples/store-go/` (Go). |

### 7.3 Ecom Use Case Content Outline

The e-commerce recommendations flow demonstrates:
1. A product catalog service receives `POST /recommendations/{customerId}` requests
2. The yaagents gateway authenticates the request (token-validator plugin)
3. The gateway injects tenant context (tenant-injector plugin)
4. The backend Python service (using sdk-fastapi) recommends products
5. If the recommendation engine needs clarification (e.g., no purchase history), it returns `clarification_required`
6. The Go client (using sdk-go) handles `clarification_required` natively
7. The audit log (otel-audit plugin) records the operation

Both `examples/store/` (Python) and `examples/store-go/` (Go) provide the runnable reference.

### 7.4 Acceptance Criteria

- [ ] `README.md`: A2A/agentic comparison framing removed; target-audience callout above the fold; ecom recommendations flow present; "who uses yaagents" section
- [ ] `start-here/overview.mdx`: independently rewritten for Pages-reader; no forced lockstep with README
- [ ] `case-studies/ecommerce-product-recommendations.mdx`: exists, shares content with README ecom section (embed or cross-link), uses `examples/store/` + `examples/store-go/`
- [ ] **Seed success signal §3**: reviewer test with 2+ external developers reports clear understanding reading README → ecom case study → overview.mdx flow

---

## 8. Tracks

### 8.1 Track INFRA

Source: Process carry-forward 2 (pi-topology.sh), Process carry-forward 3 (PRIORITIES mechanical-close trigger).

| Work Item | Description |
|-----------|-------------|
| INFRA-01: pi-topology.sh self-discovery | Refactor `bin/pi-topology.sh` to discover PI/branch topology via git branch query rather than hardcoded map. ADR PI4-yaa-0004. Adoption signal: PI4-yaa close runs `bash bin/dispatch-entry.sh PI4-yaa PC-5-NN --invoke` with NO manual topology patching. |
| INFRA-02: gh-pi-merge.sh regression test | Add regression test against the PI3-yaa B-44a unrelated-histories tolerance class (commit 28085a6). Prevents regression of the merge fix. |
| INFRA-03: PRIORITIES mechanical-close trigger | PC-5-NN terminal mechanical-close entry gains a trigger asserting PRIORITIES retire happened (greps for Active rows matching this PI; fails if any present). Update `TEMPLATE-postmortem.yml`. |

### 8.2 Track GH — GitHub Issues backlog mirror

Source: PI3-yaa INTAKE-1.

| Work Item | Description |
|-----------|-------------|
| GH-01: Backlog mirror | File ROADMAP/GOOD_FIRST_ISSUES/ADOPTERS rows as actionable issues on `ai-mpathyminds/yaagents` + submodule repos. Labels per GTM README §10 (`good first issue`, `help wanted`, `adapter`, etc.). |

### 8.3 Track DOCS

Source: PI3-yaa INTAKE-4 (prod-hosting guide) + INTAKE-5 (audit-observability page).

| Work Item | Description |
|-----------|-------------|
| DOCS-01: Production hosting guide | `how-to/host-in-production.mdx` — covers single-VM (docker compose) + reverse-proxy + TLS. Forward link to "Coming in v0.5" Helm chart. No AWS-specific content. |
| DOCS-02: Audit and observability page | `architecture/audit-and-observability.mdx` — covers v0.3 stdout log-line audit + correlation-id + Prometheus metrics + v0.4 otel-audit OTLP export roadmap. Couples with otel-audit plugin (Goal 2). |

### 8.4 Track EXAMPLES

Source: PI3-yaa INTAKE-7.

| Work Item | Description |
|-----------|-------------|
| EX-01: customer-support-triage skeleton | `examples/customer-support-triage/` — Python service skeleton demonstrating `POST /tickets/{id}:triage` agentic operation with sdk-fastapi. `clarification_required` flow for missing context. `README.md` inside. |
| EX-02: financial-risk-screening skeleton | `examples/financial-risk-screening/` — Python service skeleton demonstrating `POST /claims/{id}/risk-screens` agentic operation. `approval_required` flow for high-risk decisions. `README.md` inside. |
| EX-03: Update examples overview | `examples/overview.mdx` domain map updated to include both new skeletons. |

### 8.5 Track PROCESS

Source: PI3-yaa INTAKE-8 (cross-lane co-sign) + 4 process carry-forwards (all resolved at A-1).

| Work Item | Description |
|-----------|-------------|
| PROC-01: Cross-lane co-sign formalization | ai-platform-architect explicit acknowledgment of cross-lane canary closure (PI3-yaa smoke cfff8a6 §E7 advisory). A-1 deliverable. |
| PROC-02: yaagents-architect ratification | PC-5-02 [PROPOSED]→[ADOPTED] flipped at A-1 (agent file confirmed at `portfolio/yaagents-internal/.claude/agents/yaagents-architect.md`). Complete. |
| PROC-03: pi2-yaa-postmortem.yml committed | Process-4 default: committed at A-1 as user-authored fast-close ledger. Complete. |

---

## 9. Component Contracts (all 8 MVP components)

These contracts are unchanged from v0.3. PI4-yaa fills implementation gaps without introducing new components.

### Component 1: Agentic REST Response Profile

The normative status × content-type table (copy from PRD README — normative):

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

Every agentic response MUST emit `X-YAAgents-Profile: v0.3` header.

### Component 2: Go Gateway

Responsibilities (unchanged from v0.3 PRD):
- Authenticate incoming requests
- Extract tenant and actor context
- Enforce route-level RBAC
- Inject standard context headers (`X-Tenant-ID`, `X-Actor-ID`, `X-Request-ID`, `X-Correlation-ID`)
- Route requests to backend services
- Normalize response headers
- Emit audit logs
- Preserve typed agentic responses (passthrough, no re-wrapping)

Route-config schema (normative):
```yaml
routes:
  - id: <string>                  # unique route identifier
    method: <GET|POST|PUT|PATCH|DELETE>
    path: <resource-path>         # e.g. /campaigns/{campaignId}/optimizations
    target: <backend-url>
    roles:                        # required roles (empty = unauthenticated allowed)
      - <role-name>
    tenantRequired: <bool>        # if true, X-Tenant-ID must resolve
    audit: <bool>                 # if true, otel-audit plugin emits span
```

Plugin pipeline (v0.4 Stable):
```
token-validator → tenant-injector → license-check → prompt-sanitize → [route] → otel-audit
```

Published: `ghcr.io/yaagents/gateway:<version>`

### Component 3: Python FastAPI SDK (`sdk-fastapi`)

Published: `pip install yaagents-fastapi`

API surface (v0.3 — unchanged):
- `@agentic_operation(resource, operation_kind, mutating, roles, responses)` decorator
- `AgenticResponse.created(payload)` → `201 application/json`
- `AgenticResponse.clarification_required(message, required_inputs)` → `400 application/vnd.yaagents.clarification+json`
- `AgenticResponse.validation_failed(errors)` → `422 application/vnd.yaagents.validation-error+json`
- `AgenticResponse.approval_required(reason, approvers)` → `412 application/vnd.yaagents.approval-required+json`
- `AgenticResponse.failed_dependency(dependency, message)` → `424 application/vnd.yaagents.error+json`
- `AgenticContext` injected via FastAPI dependency — contains `tenant_id`, `actor_id`, `request_id`, `correlation_id`
- **Audit-emission surface** (Profile v0.3): `AgenticContext.audit_emit(event_type, payload)` — SDK forwards to configured sink
- `AgenticResponses.{response_type}()` — OpenAPI response schema helpers
- HTTP status mapping, Content-Type mapping, trace metadata injection

### Component 4: Go Server SDK (`sdk-go`)

Published: `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.4.0`
`ProfileVersion = "v0.3"`; router-agnostic (`net/http` core + chi/gin/echo adapters); zero non-stdlib runtime deps.

**v0.4 additions (Goal 1):**
- `AuditEmitter` interface with `EmitAuditEvent(ctx, event)` method
- `AuditEvent` struct: `EventType`, `TenantID`, `ActorID`, `RequestID`, `CorrelationID`, `Operation`, `ResourceID`, `Timestamp`
- `AgenticContext.AuditEmit(event_type, payload)` — idiomatic Go equivalent of sdk-fastapi audit surface
- Default no-op emitter; OTLP emitter available via `sdk-go/audit/otlp`
- Conformance test: `sdk-go/conformance/` — exercises same Profile v0.3 audit-emission contract as sdk-fastapi

Existing API surface (v0.3 carry-forward):
- `AgenticResponse{Type, Status, ContentType, Body, TraceMetadata}`
- `WriteAgenticResponse(w, response)` — writes correct status + Content-Type + body
- `NewAgenticContext(r)` → extracts standard headers into `AgenticContext`
- `RequireRoles(roles...)` — middleware asserting JWT roles claim
- `RequireTenant()` — middleware asserting `X-Tenant-ID` present

### Component 5: Python Client SDK (`client-python`)

Published: `pip install yaagents-client`

API surface (v0.3 — unchanged):
- `YaAgentsClient(base_url, token, tenant_id)`
- `client.resource_path(id).sub_resource().create(payload)`
- Native exceptions: `ClarificationRequired(required_inputs)`, `ValidationFailed(errors)`, `ApprovalRequired(reason)`, `FailedDependency(dependency)`, `YAAgentsError`
- Sync + async variants

### Component 6: TypeScript Client SDK (`@yaagents/client`)

Published: `npm install @yaagents/client`

API surface (v0.3 — unchanged):
- `YaAgentsClient({baseUrl, token, tenantId})`
- `client.resources.byId(id).sub().create(payload)` → typed union result
- Result type: `{type: "created" | "clarification_required" | "validation_failed" | "approval_required" | "failed_dependency" | "error"; ...}`
- ESM + CJS; browser + Node.js

### Component 7: JSON Schemas

Location: `schemas/v0.3/` (frozen; `$id` path = `/v0.3/`; body unchanged from v0.2)

Schema files:
- `clarification-required.schema.json`
- `validation-failed.schema.json`
- `approval-required.schema.json`
- `conflict.schema.json`
- `agentic-error.schema.json`
- `operation-accepted.schema.json`

Consumed by: server SDKs, client SDKs, gateway, CLI, conformance tests, OpenAPI components.

### Component 8: CLI (`yaagents-cli`)

Published: `pip install yaagents-cli`

Command surface (v0.3 — unchanged):
```bash
yaagents validate-openapi <openapi.yaml>
yaagents validate-response <response.json>
yaagents conformance-test <http://service:port>
yaagents init fastapi
```

Output: `YAAgents Conformance Report` with per-check `✓`/`✗` lines + `Overall: PASS/FAIL`.

---

## 10. OpenAPI Component Surface

Location:
```
openapi/
  yaagents-components.yaml
  yaagents-response-profile.yaml
```

Components (v0.3 — unchanged):
- Standard response headers: `X-YAAgents-Profile`, `X-Request-ID`, `X-Correlation-ID`
- Standard response schemas (refs to `schemas/v0.3/*.schema.json`)
- Standard media types (normative table in §9 above)
- `x-yaagents` operation metadata extension:
  ```yaml
  x-yaagents:
    resource: <ResourceName>
    operationKind: <recommendation|generation|assessment|review|triage>
    deterministic: <bool>
    mutating: <bool>
  ```

---

## 11. Reference Example Flows

### Existing: Campaign Optimization (`examples/campaign-api/`)

Endpoints:
```
POST /campaigns
GET  /campaigns/{campaignId}
POST /campaigns/{campaignId}/optimizations
GET  /campaigns/{campaignId}/optimizations/{optimizationId}
POST /campaigns/{campaignId}/assets:generate
```

Demonstrated flows: success creation, clarification_required, validation_failed, failed_dependency, gateway RBAC failure, client typed response handling, OpenAPI generation.

### New (v0.4): E-Commerce Recommendations (`examples/store/` + `examples/store-go/`)

Endpoints (design target):
```
POST /recommendations/{customerId}             → clarification_required if no purchase history
GET  /recommendations/{customerId}/{id}
POST /catalog/{productId}/descriptions:enhance → created (content generation)
```

Flow used in README + case-study page (§7.3 above).
Python client (`examples/store/`) + Go client (`examples/store-go/`) both demonstrate sdk-go audit-API.

### New (v0.4): Customer Support Triage skeleton (`examples/customer-support-triage/`)

Endpoint:
```
POST /tickets/{ticketId}:triage   → clarification_required if missing severity/category
```

### New (v0.4): Financial Risk Screening skeleton (`examples/financial-risk-screening/`)

Endpoint:
```
POST /claims/{claimId}/risk-screens   → approval_required for high-risk decisions
```

---

## 12. Publishing Model

### Gateway — GHCR (unchanged)
```
ghcr.io/yaagents/gateway:0.4.0
```
Multi-arch (linux/amd64, linux/arm64). GHA OIDC push from `ai-mpathyminds/yaagents-gateway` repo.

### Python packages — PyPI
```
yaagents-fastapi==0.4.0
yaagents-client==0.4.0
yaagents-cli==0.4.0
```
Published via GHA `trusted-publishing` OIDC (no API key stored in secrets).
Signed provenance attestations included.

### TypeScript package — npm
```
@yaagents/client@0.4.0
```
Published via GHA `NPM_TOKEN` secret (npm OIDC publish not yet GA for scoped packages; track at PI5-yaa).

### Go modules — Go module proxy
```
github.com/ai-mpathyminds/yaagents-sdk-go@v0.4.0
github.com/ai-mpathyminds/yaagents-client-go@v0.4.0
```
Via `git tag v0.4.0` on submodule repos + push to public GitHub; no registry push required.

### OpenAPI components + JSON schemas
Attached to GitHub Release `v0.4.0` as release assets; versioned directory in-repo (`schemas/v0.3/` frozen; `openapi/` updated in-place).

---

## 13. License Model

**v0.4 ships under Apache 2.0** (transition effective from v0.2.0 per GTM README §12 Amendment 2026-05-30).
All yaagents components are Apache 2.0 — a single unified edition. No Community/Commercial split.

- `LICENSE` file in repository root: Apache License, Version 2.0 verbatim text
- `CONTRIBUTING.md`: `legal-review-pending` banner retained until legal sign-off (does not gate PI4-yaa close; per v0.2.0 precedent)
- Copyright headers: `Copyright (c) AimpathyMinds` in all source files

**For historical reference (v0.1.x only):** v0.1.x packages on PyPI/npm/GHCR remain under the YAAgents Community License (source-available, dual-license). Non-retroactive. Users on v0.1.x who need Apache 2.0 must upgrade.

> **Legal disclaimer (verbatim from GTM README §Appendix):**
> This GTM README includes a draft licensing strategy for product planning. It is not legal
> advice. Before publishing the license publicly or accepting external contributions, consult
> a qualified software licensing lawyer.

---

## 14. NFR Seeds (platform-engineer expands at A-4)

These are seed-level NFR statements. Platform-engineer appends acceptance criteria to each WI at A-4.

| NFR Area | Seed Statement | Target WIs |
|----------|---------------|------------|
| Integration test discipline | Every plugin Stable-flip WI includes an e2e test exercising through the gateway. `t.Skip` / `pytest.skip` on infra-unreachable is forbidden per `.claude/rules/integration-test-discipline.md`. | PLG-NN (all 5 plugin WIs) |
| sdk-go test coverage | sdk-go audit-API implementation must have ≥80% line coverage on new code. CI gate enforced. | SG-01, SG-02, SG-03 |
| Pages link-audit | Every Pages-touching WI (DOC-NN, plugin doc pages) gets `bin/yaagents-pages-link-audit.sh` pre-deploy assertion — zero broken links at PR merge. | DOC-NN, PLG-doc-NN |
| pi-topology regression test | INFRA-01 pi-topology.sh refactor gets regression test against PI3-yaa B-44a unrelated-histories class + missing-topology class. | INFRA-01 |
| PRIORITIES mechanical-close | INFRA-03 PRIORITIES trigger tested against a synthetic stale Active-row PI before merging. | INFRA-03 |
| Publish-gate depends_on | All PI4-yaa publish-wave entries carry explicit `depends_on: [B-01]` per PC-5-03 adoption signal. | All publish-wave entries |
| PRECHECK_MANUAL_OK | Every `PRECHECK_MANUAL_OK` in B-01 (or equivalent) records explicit per-WARN rationale. Per PC-5-12 adoption signal. | B-01 |

---

## 15. Open Questions (for architect at A-3)

| ID | Question | Owner | Resolution Path |
|----|----------|-------|-----------------|
| OQ-A3-1 | What is the exact audit-emission contract in Profile v0.3 `spec/agentic-rest-profile.md`? Does it mandate specific event types, field names, delivery guarantees? | yaagents-architect | Read `spec/agentic-rest-profile.md` §audit; ADR PI4-yaa-0001 documents the shape |
| OQ-A3-2 | Does `yaagents/gateway/internal/plugins/tenant-injector/` from PI2-yaa PLG-4b v2 already implement webhook resolution? If so, is it tested? | yaagents-architect | Inspect directory at A-3 FIRST; work scope = doc-only + badge flip OR gap-close |
| OQ-A3-3 | Is the ai-platform/agent-api canary re-run (sdk-go shape change) IN or OUT for PI4-yaa? | yaagents-architect (OQ at A-3) | Decide after sdk-go diff assessment; cross-lane only if needed |
| OQ-A3-4 | What existing contracts do license-check, prompt-sanitize, otel-audit plugins expose? Any in-flight implementations? | yaagents-architect | Inspect `yaagents/gateway/internal/plugins/*/` before authoring close WIs |
| OQ-A3-5 | Does the plugin pipeline order need to be configurable per-route, or is the fixed order (token-validator → tenant-injector → license-check → prompt-sanitize → [route] → otel-audit) sufficient for v0.4 Stable? | yaagents-architect | ADR PI4-yaa-0002 may address; fixed order acceptable for PI4-yaa |

---

## 16. Out of Scope

- AWS substrate or `portfolio/infrastructure/` touch
- Helm chart / Kubernetes manifests authoring (PI5-yaa)
- Plugin benchmarking + load-testing (PI5-yaa)
- sdk-go feature-for-feature with sdk-fastapi beyond Profile v0.3 strict surface
- Headline case study beyond ecom (customer-support-triage + financial-risk-screening are skeletons)
- K8s/Helm (PI5-yaa per seed Out + intake OQ-5 resolved)
- GTM content (articles, demo videos, Product Hunt campaign)
- v0.2+ adapter work (Spring Boot, ASP.NET Core, NestJS — roadmap items, not v0.4)
- Any feature ask beyond the 3 user-direct goals + 8 PI3-yaa carry-forwards + 4 process carry-forwards

---

## 17. Dependencies

- PI3-yaa PC-6 close logged (commit 380efa1 on origin/main 2026-06-05; gate cleared)
- yaagents-architect agent file confirmed at `portfolio/yaagents-internal/.claude/agents/yaagents-architect.md` (PI3-yaa A-3)
- `gh-pi-merge.sh` B-44a unrelated-histories tolerance (PI3-yaa PC-5-13 commit 28085a6; merged)
- No external deps: no AWS, no third-party services, no cross-lane gating beyond optional ai-platform/agent-api canary

---

*Handoff: yaagents-architect reads this PRD at A-3 and authors the WI breakdown at `portfolio/yaagents-internal/docs/PI4-yaa/`.*
