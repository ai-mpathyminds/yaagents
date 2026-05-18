# PI1-yaa — Component: Python client (`client-python/`)

Owner lane: python-developer. Sprint 4. Published as PyPI `yaagents-client`.
ADR: PI1-yaa-0002, PI1-yaa-0003 (Hatch).

> Duplication note (library-gates Gate 4): PYC and TSC are deliberate
> parallel-language clients of the same `spec/examples/v0.1` contract — they
> differ by language ecosystem, not by `service:` only, so Gate 4's
> same-roadmap-mirror heuristic does not fire. `duplication_override:
> intentional dual-language public client SDKs; shared contract is the
> golden corpus, not extractable code (rule-of-three N/A — 2 languages).`

---

### WI-1yaa.PYC-1: `YaAgentsClient` + resource fluent accessors [DONE]
service: yaagents/client-python
brief: `YaAgentsClient(base_url, token, tenant_id)`; `.campaigns(id)` →
`CampaignResource`; `CampaignResource.optimizations.create(body)` →
`POST /campaigns/{id}/optimizations`; `.assets.generate(body)` →
`POST /campaigns/{id}/assets:generate`. Default headers: `Authorization:
Bearer`, `X-Tenant-ID`, auto `X-Correlation-ID` (overridable). `httpx` or
stdlib — minimal deps.
acceptance:
- Headers injected on every request; correlation-id auto + overridable
- Resource accessors build correct method+path+body
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-1]

### WI-1yaa.PYC-2: Typed exception mapping [READY]
service: yaagents/client-python
brief: Map response status+vendor content-type → typed exceptions:
`ClarificationRequired` (`.required_inputs`), `ValidationFailed` (`.errors`),
`FailedDependency` (`.dependency`), `AgenticForbidden`, `AgenticError` (base).
`success`/`created` return the deserialized payload. Unknown vendor type →
`AgenticError` with raw body.
acceptance:
- Each mandatory error media type raises its specific exception with parsed attributes
- `created` returns payload object; base `AgenticError` catches all
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.PYC-1, WI-1yaa.SPEC-2]

### WI-1yaa.PYC-3: Client conformance tests [READY]
service: yaagents/client-python
brief: pytest suite driving the client against a fixture HTTP server that
replays the `spec/examples/v0.1` golden corpus; assert correct
exception/payload per fixture. Coverage ≥80%; `ruff`+`mypy`+`pip-audit` clean.
acceptance:
- One test per corpus fixture asserting the mapped outcome
- Coverage ≥80%; lint/type/audit clean
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.PYC-2, WI-1yaa.SPEC-5]

---

## NFR Addendum — A-4 platform-engineer pass (2026-05-17)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] dependency audit (`pip-audit`) | feature WI | PYC-3 acceptance criteria |
| [SRE] health/readiness/logs | N/A | library, not a running service |
| [SUPPLY-CHAIN] OIDC publish — no long-lived token | feature WI | REL-3 (PyPI Trusted Publisher) |
| [SUPPLY-CHAIN] reproducible builds (Hatch) | feature WI | REL-3 + ADR PI1-yaa-0003 |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI1-yaa |

No new NFR WIs required — all applicable dimensions covered by feature WIs.
