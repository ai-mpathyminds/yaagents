# YAAgents v0.1 — Agentic REST Profile MVP — Detailed PRD

Status: [READY]
Owner: product-manager (yaagents)
PI: PI1-yaa
Date: 2026-05-17
Profile version: v0.1

> Seeded by chief-architect. Expanded by product-manager 2026-05-17.

---

## 1. Problem & Context

Production systems are resource-oriented; chat endpoints (`/agents/invoke`) are the wrong
integration surface for business workflows. Exposing agentic capabilities this way produces:

- Loosely typed inputs and outputs
- Inconsistent access control
- Client applications requiring custom parsing logic
- Unstructured clarification flows
- APIs that are difficult to document and test
- Bypassed API gateway and microservice practices
- Agent framework internals leaking into the external integration model

YAAgents solves this by making agentic behavior look and behave like normal REST APIs, with
the agent remaining an implementation detail behind a controlled resource operation.

The governing contrast:

```
POST /campaigns/{campaignId}/optimizations   ← YAAgents pattern
POST /agents/campaign-agent/invoke           ← NOT this
```

---

## 2. Solution Overview

YAAgents provides the **interface, gateway, response-contract, and client-consumption layer**
for agentic capabilities built using any internal framework. It does not introduce a new
agent runtime. The MVP ships:

- A normative **Agentic REST Response Profile** (status × media-type contract)
- 6 **JSON schemas** validating each vendor media type
- Reusable **OpenAPI components** with `x-yaagents` metadata
- A **Go gateway** handling authn, tenant/actor context, RBAC, audit, typed-response passthrough
- A **Python FastAPI SDK** for server-side response building and OpenAPI generation
- A **Python client** with typed exception handling
- A **TypeScript client** with result-style response handling
- A **CLI validator** for conformance testing
- A **Campaign reference example** with Docker Compose demo
- Published and installable packages (PyPI ×3, npm ×1, GHCR image) via OIDC

The operating thesis: **Bring your own agent. YAAgents makes it a product-grade API.**

---

## 3. Target Users

| Tier | Who | Need |
|------|-----|------|
| External primary | Platform engineers, API architects, backend engineers, AI-platform teams | Govern agentic capabilities behind standard REST contracts |
| External secondary | FastAPI developers, OpenAPI-first teams, LangGraph/SK users, enterprise architects | Framework-neutral interface layer for existing agent implementations |
| Internal forward dep | AimpathyMinds (`platform-services/gateway` + `ai-gateway`) | Future re-base on the yaagents gateway; convergence unblocked by PI1-yaa |

---

## 4. Agentic REST Response Profile (NORMATIVE)

The following table is the normative Agentic REST Response Profile. All components MUST
implement and honour it. Paraphrasing or diverging from these media types is prohibited.

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

**Profile versioning:** Every published package MUST declare the profile version it supports,
e.g., `Supports YAAgents Profile v0.1`.

### 4.1 Clarification Required — canonical body shape

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

### 5.1 Component 1 — Agentic REST Response Profile (`spec/`)

**Directory:** `spec/`

**Responsibilities:**
- Authoritative prose definition of the Response Profile
- Profile version declaration (`v0.1`)
- Normative status × media-type table (§4 above)
- Clarification body contract
- Correlation-ID propagation contract
- Acceptance criteria for conformance

**Contract:** The `spec/` directory MUST be the single authoritative source. All other
components reference it; none redefine it.

---

### 5.2 Component 2 — JSON Schemas (`schemas/`)

**Directory:** `schemas/`

**Schema list (6):**

| File | Validates |
|------|-----------|
| `clarification-required.schema.json` | `application/vnd.yaagents.clarification+json` |
| `validation-failed.schema.json` | `application/vnd.yaagents.validation-error+json` |
| `approval-required.schema.json` | `application/vnd.yaagents.approval-required+json` |
| `conflict.schema.json` | `application/vnd.yaagents.conflict+json` |
| `agentic-error.schema.json` | `application/vnd.yaagents.error+json` (forbidden / failed_dependency / error) |
| `operation-accepted.schema.json` | `application/vnd.yaagents.operation+json` |

**Contract:**
- Every schema MUST carry a `$schema` (JSON Schema Draft-07 minimum), `$id`, and `title`.
- The CLI, server SDK, client SDKs, and conformance tests all reference these schemas.
- Schemas are versioned with the profile; a `v0.1/` path prefix under `schemas/` is
  recommended for future compatibility.

---

### 5.3 Component 3 — OpenAPI Components (`openapi/`)

**Directory:** `openapi/`

**Files:**

| File | Contents |
|------|----------|
| `yaagents-components.yaml` | Standard headers, standard response schemas, standard media types, standard error responses |
| `yaagents-response-profile.yaml` | `x-yaagents` operation metadata extension |

**`x-yaagents` extension fields:**

| Field | Type | Description |
|-------|------|-------------|
| `resource` | string | Domain resource name (e.g. `Campaign`) |
| `operationKind` | string | `recommendation` · `generation` · `mutation` · `analysis` |
| `deterministic` | boolean | Whether the operation is deterministic |
| `mutating` | boolean | Whether the operation mutates state |

**OpenAPI response component surface (per response type):**

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

**Contract:** Every agentic endpoint in the reference example MUST use these components.
The CLI `validate-openapi` command verifies `x-yaagents` metadata presence and correct
content-type/schema wiring.

---

### 5.4 Component 4 — Go Gateway (`gateway/`)

**Directory:** `gateway/`
**Published image:** `ghcr.io/ai-mpathyminds/yaagents-gateway:0.1.0` (GHCR, multi-arch: `linux/amd64` + `linux/arm64`)

**Responsibilities:**

| Responsibility | Detail |
|----------------|--------|
| Authentication | Validate bearer token (JWT RS256 in prod; configurable); reject unauthenticated requests |
| Tenant/actor context | Extract `X-Tenant-ID` and actor claims; inject as request headers to upstream |
| Route-level RBAC | Enforce `roles:` from route config; return `403 application/vnd.yaagents.error+json` on failure |
| Context header injection | Forward `X-Correlation-ID` (generate if absent) and `X-Request-ID` to upstream |
| Request routing | Proxy to `target:` URL per route config; preserve HTTP method and body |
| Typed-response passthrough | Preserve upstream `Content-Type` and status; do not re-encode agentic vendor types |
| Response header normalization | Add `X-YAAgents-Profile: v0.1` to all proxied responses |
| Audit log | Emit structured JSON audit event per request: route, tenant, actor, status, latency, correlation-id |
| Health checks | `GET /healthz` (liveness) · `GET /readyz` (readiness) |
| Metrics | Prometheus text format on `GET /metrics` |

**Route-config schema (YAML):**

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
```

**Gateway configuration (env / config file):**

| Key | Description |
|-----|-------------|
| `GATEWAY_PORT` | Listen port (default `8080`; portfolio allocation `8120`) |
| `GATEWAY_ROUTES_FILE` | Path to route-config YAML |
| `GATEWAY_JWT_SECRET` | HS256 secret (dev only) |
| `GATEWAY_JWT_JWKS_URL` | JWKS URL for RS256 (production) |
| `GATEWAY_AUDIT_LOG` | `stdout` (default) or future sink config |

**Go module:** `github.com/ai-mpathyminds/yaagents/gateway`

**Build:** Multi-stage Alpine Dockerfile, non-root user, CGO_ENABLED=0. SBOM generated
at publish time (PI1-yaa); Cosign signing deferred to PI2-yaa.

---

### 5.5 Component 5 — Python FastAPI SDK (`sdk-fastapi/`)

**Directory:** `sdk-fastapi/`
**PyPI package:** `yaagents-fastapi`
**Installs via:** `pip install yaagents-fastapi`

**API surface:**

| Symbol | Kind | Description |
|--------|------|-------------|
| `@agentic_operation(...)` | decorator | Declare an agentic endpoint; injects `AgenticContext`, generates OpenAPI response stanzas |
| `AgenticResponse` | class | Factory for all 10 response types |
| `AgenticResponse.success(payload)` | method | `200 application/json` |
| `AgenticResponse.created(payload)` | method | `201 application/json` |
| `AgenticResponse.accepted(operation_id, ...)` | method | `202 application/vnd.yaagents.operation+json` |
| `AgenticResponse.clarification_required(message, required_inputs)` | method | `400 application/vnd.yaagents.clarification+json` |
| `AgenticResponse.validation_failed(errors)` | method | `422 application/vnd.yaagents.validation-error+json` |
| `AgenticResponse.approval_required(...)` | method | `412 application/vnd.yaagents.approval-required+json` |
| `AgenticResponse.forbidden(message)` | method | `403 application/vnd.yaagents.error+json` |
| `AgenticResponse.conflict(message)` | method | `409 application/vnd.yaagents.conflict+json` |
| `AgenticResponse.failed_dependency(message, dependency)` | method | `424 application/vnd.yaagents.error+json` |
| `AgenticResponse.error(message)` | method | `500 application/vnd.yaagents.error+json` |
| `AgenticContext` | class | Injected request context: `tenant_id`, `actor_id`, `correlation_id`, `request_id` |
| `RequiredInput` | dataclass | Clarification input descriptor |
| `AgenticResponses` | class | OpenAPI response-collection helpers for `@agentic_operation` |

**`@agentic_operation` parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `resource` | str | Domain resource name |
| `operation_kind` | str | `recommendation` · `generation` · `mutation` · `analysis` |
| `mutating` | bool | Whether the operation mutates state |
| `roles` | list[str] | Informational roles for OpenAPI documentation |
| `responses` | list | `AgenticResponses.*` entries for OpenAPI generation |

**SDK handles automatically:**
- HTTP status code mapping from `AgenticResponse` type
- `Content-Type` header from normative table (§4)
- `trace.correlationId` / `trace.requestId` injection from `AgenticContext`
- OpenAPI `responses:` block generation with correct content-type per response type

---

### 5.6 Component 6 — Python Client (`client-python/`)

**Directory:** `client-python/`
**PyPI package:** `yaagents-client`
**Installs via:** `pip install yaagents-client`

**API surface:**

| Symbol | Kind | Description |
|--------|------|-------------|
| `YaAgentsClient(base_url, token, tenant_id)` | class | Root client; one instance per service base URL |
| `client.campaigns(campaign_id)` | method | Returns a `CampaignResource` sub-client |
| `CampaignResource.optimizations.create(body)` | method | `POST /campaigns/{id}/optimizations` |
| `CampaignResource.assets.generate(body)` | method | `POST /campaigns/{id}/assets:generate` |
| `ClarificationRequired` | exception | Raised when `400 application/vnd.yaagents.clarification+json`; carries `.required_inputs` |
| `ValidationFailed` | exception | Raised on `422 application/vnd.yaagents.validation-error+json`; carries `.errors` |
| `FailedDependency` | exception | Raised on `424 application/vnd.yaagents.error+json`; carries `.dependency` |
| `AgenticForbidden` | exception | Raised on `403 application/vnd.yaagents.error+json` |
| `AgenticError` | exception | Base class for all agentic error exceptions |

**Headers injected by default:**
- `Authorization: Bearer {token}`
- `X-Tenant-ID: {tenant_id}`
- `X-Correlation-ID: {auto-generated UUID}` (overridable)

---

### 5.7 Component 7 — TypeScript Client (`client-ts/`)

**Directory:** `client-ts/`
**npm package:** `@aimpathyminds/yaagents-client`
**Installs via:** `npm install @aimpathyminds/yaagents-client`

**Note:** This is a **library**, not a UI application. Targets ESM + CommonJS; supports
Node.js and browser environments. TypeScript types are first-class.

**API surface:**

| Symbol | Kind | Description |
|--------|------|-------------|
| `YaAgentsClient({ baseUrl, token, tenantId })` | class | Root client |
| `client.campaigns.byId(id)` | method | Returns a `CampaignResource` fluent accessor |
| `CampaignResource.optimizations().create(body)` | method | `POST /campaigns/{id}/optimizations` |
| `CampaignResource.assets().generate(body)` | method | `POST /campaigns/{id}/assets:generate` |
| `AgenticResult<T>` | type | Discriminated union of all response types |
| `result.type === 'created'` | discriminant | Access `result.resource: T` |
| `result.type === 'clarification_required'` | discriminant | Access `result.requiredInputs` |
| `result.type === 'validation_failed'` | discriminant | Access `result.errors` |
| `result.type === 'failed_dependency'` | discriminant | Access `result.dependency` |
| `result.type === 'accepted'` | discriminant | Access `result.operationId` |

**Result-style vs exception-style:**
- Default: `AgenticResult<T>` discriminated union (no throws; caller switches on `result.type`)
- Alternative: `client.strict()` wrapper that throws typed exceptions on non-success responses

---

### 5.8 Component 8 — CLI Validator (`cli/`)

**Directory:** `cli/`
**PyPI package:** `yaagents-cli`
**Installs via:** `pip install yaagents-cli`

**Command surface:**

| Command | Input | Output | Description |
|---------|-------|--------|-------------|
| `yaagents validate-openapi <file.yaml>` | OpenAPI YAML | pass/fail + findings | Validates `x-yaagents` metadata, correct content-type per response type, schema refs |
| `yaagents validate-response <file.json>` | JSON response body | pass/fail + schema errors | Validates body against the relevant JSON schema (inferred from `type` field) |
| `yaagents conformance-test <base-url>` | Live service URL | PASS / FAIL report | Exercises all mandatory response types, checks headers, verifies correlation-id propagation |
| `yaagents init fastapi` | — | Scaffold files | Generates a FastAPI starter with `@agentic_operation` and correct response wiring |

**Conformance report output format:**

```text
YAAgents Conformance Report

✓ OpenAPI includes x-yaagents metadata
✓ Clarification response uses correct content type
✓ 400 response matches clarification schema
✓ Correlation ID propagated
✓ Gateway route requires tenant context

Overall: PASS
```

---

## 6. Reference Example: Campaign API (`examples/campaign-api/`)

**Directory:** `examples/campaign-api/`
**Compose file:** `examples/campaign-api/docker-compose.yml`
**Ports:** Gateway `8120`, Campaign API `8121` (portfolio port table)

### 6.1 Endpoints

```http
POST   /campaigns
GET    /campaigns/{campaignId}
POST   /campaigns/{campaignId}/optimizations
GET    /campaigns/{campaignId}/optimizations/{optimizationId}
POST   /campaigns/{campaignId}/assets:generate
```

### 6.2 Demonstrated flows

| Flow | Trigger | Expected response |
|------|---------|-------------------|
| Successful optimization | Request with all required fields | `201 application/json` |
| Clarification required | Request missing `successMetric` | `400 application/vnd.yaagents.clarification+json` |
| Validation failed | Request with invalid field types | `422 application/vnd.yaagents.validation-error+json` |
| Failed dependency | Downstream LLM unavailable | `424 application/vnd.yaagents.error+json` |
| Gateway RBAC failure | Token missing required role | `403 application/vnd.yaagents.error+json` |
| Client typed response | Python/TS client SDK | `ClarificationRequired` exception / `result.type === 'clarification_required'` |
| OpenAPI generation | `GET /openapi.json` | Includes `x-yaagents` metadata and vendor content-types |

### 6.3 Compose topology

```
yaagents-gateway (port 8120)   ←→   campaign-api (port 8121, internal)
```

The gateway container reads `routes.yaml` from a bind-mount; the campaign-api
container runs the FastAPI SDK-based server.

### 6.4 Quick start

```bash
cd examples/campaign-api
docker compose up

# Clarification required
curl -X POST http://localhost:8120/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal": "reduce_cost_per_lead"}'

# Conformance test
yaagents conformance-test http://localhost:8120
```

---

## 7. Publishing Model

### 7.1 Gateway image

```
ghcr.io/ai-mpathyminds/yaagents-gateway:0.1.0
ghcr.io/ai-mpathyminds/yaagents-gateway:latest   (convenience alias)
```

- Registry: GitHub Container Registry (GHCR)
- Architectures: `linux/amd64`, `linux/arm64`
- Build: multi-stage Alpine; non-root user; CGO_ENABLED=0
- SBOM: generated at publish (Syft or equivalent)
- CI: GitHub Actions; **OIDC trusted publishing via `docker/login-action` with GHCR OIDC** — no long-lived registry tokens
- Signing: Cosign deferred to PI2-yaa

### 7.2 Python packages (PyPI)

| Package | Version | Description |
|---------|---------|-------------|
| `yaagents-fastapi` | `0.1.0` | FastAPI server SDK |
| `yaagents-client` | `0.1.0` | Python client |
| `yaagents-cli` | `0.1.0` | CLI validator |

- Build tool: Hatch (preferred) or Poetry
- CI: GitHub Actions; **OIDC trusted publishing via PyPI Trusted Publisher** — no API tokens stored in CI
- Metadata: `Supports-YAAgents-Profile: v0.1` in package metadata
- TestPyPI run gates production publish

### 7.3 TypeScript package (npm)

| Package | Version | Description |
|---------|---------|-------------|
| `@aimpathyminds/yaagents-client` | `0.1.0` | TypeScript/JavaScript client |

- Format: ESM primary + CommonJS compat bundle
- Types: TypeScript declarations (`.d.ts`) bundled
- CI: GitHub Actions; **OIDC trusted publishing via npm Provenance** — no `NPM_TOKEN` stored in CI
- Peer deps: none (zero runtime dependencies target)

### 7.4 Schemas and OpenAPI components

Distributed in-repo (`schemas/`, `openapi/`) and attached to GitHub Releases as versioned
archives. Future: stable URLs (e.g., `spec.yaagents.dev/v0.1/`) — not PI1-yaa scope.

---

## 8. License Model

> **Disclaimer (verbatim from GTM README §Appendix):** This GTM README includes a draft
> licensing strategy for product planning. It is not legal advice. Before publishing the
> license publicly or accepting external contributions, consult a qualified software
> licensing lawyer.

YAAgents uses a **dual-license model** and is **source-available / fair-code**. It is
**NOT** OSI-approved open source and MUST NOT be marketed as such.

Correct wording: **source-available** · **fair-code** · _free for academics, individuals,
and small developers_ · _commercial license required for larger organizations_.

### 8.1 Community License (free)

Permitted for:
- Individual developers
- Academic research, teaching, student projects
- Non-commercial use
- Evaluation, testing, proof-of-concept
- Organizations with **< 10 employees AND < USD 1,000,000 annual gross revenue**

### 8.2 Commercial License (required)

Required for:
- Organizations with ≥ 10 employees OR ≥ USD 1,000,000 annual gross revenue
- Production commercial SaaS / hosted / managed service offerings
- Embedding or redistributing YAAgents in a commercial product
- Consulting or platform companies delivering paid services using YAAgents
- Any use that competes with a paid YAAgents offering

### 8.3 PI1-yaa deliverables

| Artifact | Action |
|----------|--------|
| `LICENSE` | YAAgents Community License v0.1 (draft from GTM README §14) |
| `COMMERCIAL.md` | Commercial license terms and contact |
| `CONTRIBUTING.md` | Contributor guide with CLA note |
| `SECURITY.md` | Vulnerability disclosure policy |
| `CODE_OF_CONDUCT.md` | Code of conduct |

Legal review of the license text MUST happen before accepting external contributions and
before public launch. This is a legal gate, not a PI1-yaa engineering gate.

---

## 9. Repo Scaffolding

| File / directory | Owner | Description |
|------------------|-------|-------------|
| `README.md` | platform-engineer | Product overview, quick-start, architecture diagram, badge set |
| `LICENSE` | platform-engineer | YAAgents Community License v0.1 draft |
| `COMMERCIAL.md` | platform-engineer | Commercial licensing terms + contact |
| `CONTRIBUTING.md` | platform-engineer | Contributor guide; CLA placeholder |
| `SECURITY.md` | platform-engineer | Vulnerability disclosure |
| `CODE_OF_CONDUCT.md` | platform-engineer | Code of conduct (Contributor Covenant) |
| `.github/ISSUE_TEMPLATE/` | platform-engineer | Bug report, feature request, adapter request templates |
| `.github/workflows/` | platform-engineer | CI (test, lint), publish (gateway/PyPI/npm) via OIDC |

---

## 10. NFR Seeds

> Platform-engineer expands these into WI bodies at A-4.

| NFR area | Seed requirement |
|----------|-----------------|
| Supply chain | OIDC trusted publishing for all 4 package targets (GHCR, PyPI ×3, npm); no long-lived tokens in CI |
| Container security | Multi-stage Alpine build; non-root user; `trivy` image scan in CI; SBOM attached to GitHub Release |
| Go quality | `golangci-lint` clean; `govulncheck` clean; ≥80% coverage on gateway core logic |
| Python quality | `ruff` + `mypy` clean; `pip-audit` in CI; ≥80% coverage on SDK + client + CLI |
| TypeScript quality | `eslint` + `tsc --noEmit` clean; `npm audit` in CI; vitest coverage ≥80% on client |
| Dev-host ceiling | Compose demo MUST run within 16 GB RAM, no CGO; heavy/live tests deferred to CI |
| Correlation-id | `X-Correlation-ID` generated at gateway if absent; propagated to upstream; echoed in every response `trace` block |
| Health / readiness | `/healthz` + `/readyz` on gateway (liveness/readiness); `/metrics` Prometheus text format |
| Logging | Structured JSON logs with `request_id` / `correlation_id` on every request log line |
| Versioning | `X-YAAgents-Profile: v0.1` response header on all proxied responses |

---

## 11. Out-of-Scope

| Capability | Deferred to |
|------------|------------|
| Kubernetes manifests | PI2-yaa |
| Helm chart (GHCR OCI) | PI2-yaa |
| K8s deployment guide | PI2-yaa |
| Cosign image signing | PI2-yaa |
| SBOM attestation hardening | PI2-yaa |
| Spring Boot / ASP.NET adapters | v0.2+ (later PIs) |
| OpenTelemetry support | v0.3+ (later PIs) |
| OPA policy integration | v0.3+ (later PIs) |
| LangGraph / Semantic Kernel plugins | v0.3+ (later PIs) |
| Async-operation profile | v0.2+ (later PIs) |
| Approval-flow runtime | v0.2+ (later PIs) |
| GTM content (demo videos, launch blog, social) | founder-owned; not an engineering PI |
| Internal gateway re-base (`platform-services/gateway` + `ai-gateway`) | Future seed; NOT yaagents scope |

---

## 12. Success Criteria

1. A developer can expose an agentic REST endpoint using the FastAPI SDK.
2. The endpoint can return `created`, `clarification_required`, `validation_failed`, and `failed_dependency`.
3. The SDK maps each response to the correct status code and content-type per the normative table (§4).
4. OpenAPI includes response-specific content types (`x-yaagents` metadata + vendor `Content-Type` per response).
5. The gateway enforces route-level RBAC; requests missing required roles return `403 application/vnd.yaagents.error+json`.
6. The gateway propagates tenant, actor, request-id, and correlation-id context.
7. Python and TypeScript clients handle `clarification_required` natively (typed exception / discriminated union).
8. The Campaign reference example runs with `docker compose up` in `examples/campaign-api/`.
9. `yaagents-cli conformance-test http://localhost:8120` returns PASS.
10. `pip install yaagents-fastapi`, `pip install yaagents-client`, `pip install yaagents-cli` succeed from public PyPI.
11. `npm install @aimpathyminds/yaagents-client` succeeds from public npm.
12. `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.1.0` succeeds from GHCR.

---

## 13. Open Questions

| # | Question | Owner | Resolution target |
|---|----------|-------|-------------------|
| OQ-1 | License threshold enforcement: revenue vs employee vs funding — which single model for v0.1 launch? GTM README §13 recommends combined (< 10 employees AND < $1M). Does chief-architect / user want to lock this before legal review, or keep placeholder? | chief-architect / user | Before PI1-yaa PC-6 (pre-publish) |
| OQ-2 | JWT validation strategy for gateway dev mode: HS256 symmetric (simpler local dev) or mock JWKS endpoint? Impacts demo ergonomics. | yaagents-architect | A-3 |
| OQ-3 | `accepted` (202) response flow in PI1-yaa: should the gateway support async operation polling (`GET /campaigns/{id}/optimizations/{opId}/status`)? The schema ships (PI1-yaa), but the runtime polling loop may be v0.2 scope. | yaagents-architect | A-3 |
| OQ-4 | Python packaging tool: Hatch vs Poetry — team preference affects the three PyPI packages' `pyproject.toml` structure and CI commands. | python-developer | A-3 |
| OQ-5 | SBOM format: Syft (SPDX JSON) or CycloneDX? GHCR supports both. Affects the platform-engineer publishing WI. | platform-engineer | A-4 |
| OQ-6 | npm package target: pure ESM or dual ESM+CJS bundle? CJS adds 10–15% bundle overhead but improves Jest compatibility for consumers. | frontend-developer | A-3 |

---

## 14. Handoff

```
next:        yaagents-architect
artifact:    yaagents/system-refs/yaagents-mvp_detailed.md
intent:      Expand into PI1-yaa roadmap (docs/PI1-yaa/roadmap.md + per-component WI
             files + ADRs under docs/adr/). OQs 2-6 are A-3 architect decision points.
cwd:         yaagents/
```
