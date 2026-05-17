# PI1-yaa — Component: spec / schemas / openapi (the contract)

Owner lane: python-developer (yaml/json contract authoring; CLI is the natural
consumer/validator). Sprint 1. **Gates every other component** — sequenced
first per the runbook A-3 ordering rule. ADR: PI1-yaa-0002.

> Architect note: `spec/`, `schemas/`, `openapi/` are unassigned in the
> portfolio lane table (not a `{lang}` service dir). Assigned to
> `python-developer` because the Python CLI validator is the primary
> first-party consumer and the artifacts are language-neutral json/yaml. This
> is an A-3 architect decision; recorded for scrum-master visibility.

---

### WI-1yaa.SPEC-1: Authoritative Response Profile spec [DONE]
service: yaagents/spec
brief: Author `spec/agentic-rest-profile.md` — authoritative prose definition
of the Agentic REST Response Profile. The PRD §4 normative status × media-type
table is copied **verbatim** (no paraphrase). Include: profile version
declaration, the `clarification_required` canonical body shape (PRD §4.1),
the mandatory `trace` block contract (`correlationId`+`requestId` in every
agentic body), and conformance acceptance criteria. Add `spec/VERSION` = `0.1`.
acceptance:
- `spec/agentic-rest-profile.md` contains the 10-row table byte-identical to PRD §4
- `spec/VERSION` = `0.1`; profile header literal documented as `X-YAAgents-Profile: v0.1`
- Markdown lints clean; no other component redefines the table (grep check noted in EX-4 gate)
library_justify: novel; standalone OSS surface (normative contract — no catalog entry)
depends_on: []

### WI-1yaa.SPEC-2: Six JSON schemas (Draft-07) [DONE]
service: yaagents/schemas
brief: Author the 6 schemas under `schemas/v0.1/`:
`clarification-required.schema.json`, `validation-failed.schema.json`,
`approval-required.schema.json`, `conflict.schema.json`,
`agentic-error.schema.json` (forbidden/failed_dependency/error),
`operation-accepted.schema.json`. Each carries `$schema` (Draft-07), `$id`,
`title`, and a required `trace` object (`correlationId`, `requestId`). Mirror
the §4.1 clarification body shape exactly.
acceptance:
- 6 files under `schemas/v0.1/`; each validates as Draft-07 metaschema-clean
- Every schema requires `type`, `code`, `message`, `trace.correlationId`, `trace.requestId`
- `clarification-required` models `requiredInputs[]` (name/location/type/required/question/allowedValues)
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-1]

### WI-1yaa.SPEC-3: OpenAPI reusable components [DONE]
service: yaagents/openapi
brief: Author `openapi/yaagents-components.yaml` — standard headers
(`X-Correlation-ID`, `X-Request-ID`, `X-Tenant-ID`, `X-YAAgents-Profile`),
standard response schemas (`$ref` to the 6 schemas), standard vendor media
types, and standard reusable error responses keyed by the normative table.
acceptance:
- `components.schemas` covers all 6 vendor bodies; `components.headers` + `components.responses` reusable
- `openapi-spec-validator` (or `swagger-cli validate`) passes on the file
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-2]

### WI-1yaa.SPEC-4: `x-yaagents` OpenAPI extension + example surface [WIP]
service: yaagents/openapi
brief: Author `openapi/yaagents-response-profile.yaml` — the `x-yaagents`
operation-metadata extension (`resource`, `operationKind`
[recommendation·generation·mutation·analysis], `deterministic`, `mutating`)
plus a worked operation surface for `POST /campaigns/{campaignId}/optimizations`
mapping each response type to its vendor content-type + schema `$ref` (PRD §5.3).
acceptance:
- Extension documented with all 4 fields + allowed `operationKind` enum
- Example operation wires 201/400/422/424 to correct content-types + schema refs per §5.3
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-3]

### WI-1yaa.SPEC-5: Shared golden conformance corpus [READY]
service: yaagents/spec
brief: Author `spec/examples/v0.1/` — for each of the 6 media types, one
`*.valid.json` and ≥1 `*.invalid.json` fixture (the invalid set covers a
missing `trace`, a wrong `type`, and a schema-violating field). This is the
**single cross-component conformance fixture** (ADR PI1-yaa-0002 §5): SDK,
both clients, and CLI all test against it — prevents per-component drift.
acceptance:
- ≥12 valid + ≥18 invalid fixtures; every valid fixture passes its schema, every invalid fails
- A `spec/examples/INDEX.md` maps fixture → media type → expected validator verdict
library_justify: novel; standalone OSS surface (de-facto contract test corpus)
depends_on: [WI-1yaa.SPEC-2]

---

## NFR Addendum — A-4 platform-engineer pass (2026-05-17)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] secrets / access controls | N/A | static JSON/YAML/Markdown contract files; no secrets surface |
| [SRE] health/readiness/logs/resource limits | N/A | static artifacts; no running service |
| [SUPPLY-CHAIN] versioned archive published | feature WI | REL-6 (schemas + openapi archive on GitHub Release) |
| [SUPPLY-CHAIN] OIDC | N/A | static artifacts; no package registry publish |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI1-yaa |

No new NFR WIs required — static contract artifacts; all applicable dimensions covered.
