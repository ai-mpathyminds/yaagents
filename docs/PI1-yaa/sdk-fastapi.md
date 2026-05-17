# PI1-yaa — Component: Python FastAPI SDK (`sdk-fastapi/`)

Owner lane: python-developer. Sprint 3. Published as PyPI `yaagents-fastapi`.
ADR: PI1-yaa-0002 (normative table source, 202 schema-only), PI1-yaa-0003 (Hatch).

---

### WI-1yaa.SDK-1: `AgenticResponse` factory — all 10 response types [DONE]
service: yaagents/sdk-fastapi
brief: `AgenticResponse` class with the 10 factory methods (PRD §5.5):
`success/created/accepted/clarification_required/validation_failed/
approval_required/forbidden/conflict/failed_dependency/error`. Each returns a
FastAPI `Response` with status code + `Content-Type` taken from the **spec/
normative table** (do not hardcode-redefine; reference `spec/VERSION`). Inject
`trace.correlationId`/`trace.requestId`. `accepted(operation_id,...)` ships the
202 body shape (no polling runtime — ADR 0002 §4).
acceptance:
- 10 methods, each emitting the exact status+content-type from the §4 table
- Every body carries a populated `trace` block; `ruff`+`mypy` clean
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-1, WI-1yaa.SPEC-2]

### WI-1yaa.SDK-2: `AgenticContext` + `RequiredInput` [DONE]
service: yaagents/sdk-fastapi
brief: `AgenticContext` FastAPI dependency extracting `tenant_id`, `actor_id`,
`correlation_id`, `request_id` from request headers (the gateway-injected set);
generate correlation_id if absent (standalone-without-gateway use). `RequiredInput`
dataclass (name/location/type/required/question/allowed_values) feeding
`clarification_required`.
acceptance:
- `AgenticContext` injectable via `Depends`; all 4 fields populated from headers
- `RequiredInput` round-trips into a §4.1-shaped `requiredInputs[]` entry
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SDK-1]

### WI-1yaa.SDK-3: `@agentic_operation` decorator + OpenAPI generation [DONE]
service: yaagents/sdk-fastapi
brief: `@agentic_operation(resource, operation_kind, mutating, roles,
responses)` decorator: injects `AgenticContext`, and generates the FastAPI
OpenAPI `responses:` block with the correct vendor `Content-Type` + schema
`$ref` per declared response type **and** the `x-yaagents` operation metadata
(per `openapi/yaagents-response-profile.yaml`). `AgenticResponses` helper class
for declaring the response set.
acceptance:
- Decorated endpoint's generated OpenAPI has `x-yaagents` + per-response vendor content-types
- `yaagents validate-openapi` (CLI-2) passes on the SDK-generated `/openapi.json` of the example
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SDK-2, WI-1yaa.SPEC-4]

### WI-1yaa.SDK-4: SDK schema-conformance tests [DONE]
service: yaagents/sdk-fastapi
brief: pytest suite asserting every `AgenticResponse.*` body validates against
the matching `schemas/v0.1` schema using the shared `spec/examples/v0.1`
golden corpus (SPEC-5) as the oracle. Coverage ≥80% on the SDK package.
acceptance:
- One parametrised test per media type validating SDK output against schema + corpus
- `pytest` green; coverage ≥80%; `pip-audit` clean
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SDK-3, WI-1yaa.SPEC-5]

---

## NFR Addendum — A-4 platform-engineer pass (2026-05-17)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] dependency audit (`pip-audit`) | feature WI | SDK-4 acceptance criteria |
| [SRE] health/readiness endpoints | N/A | library, not a running service |
| [SRE] resource limits | N/A | library, not a running service |
| [SUPPLY-CHAIN] OIDC publish — no long-lived token | feature WI | REL-3 (PyPI Trusted Publisher) |
| [SUPPLY-CHAIN] reproducible builds (Hatch) | feature WI | REL-3 + ADR PI1-yaa-0003 |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI1-yaa |

No new NFR WIs required — all applicable dimensions covered by feature WIs.
