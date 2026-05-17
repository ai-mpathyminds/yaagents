# PI1-yaa — YAAgents v0.1 Agentic REST Profile MVP — Roadmap (master)

Status: [READY] (A-4 platform-engineer NFR pass complete 2026-05-17; all 35 WIs [READY])
PI: PI1-yaa · Lane: yaa (parallel to Lane A oppor + Lane B plt-aip)
Author: yaagents-architect · Date: 2026-05-17
PRD: `yaagents/system-refs/yaagents-mvp_detailed.md` (§ refs throughout)
Seed: `yaagents/system-refs/yaagents-mvp.seed.md` [READY]

## Scope

MVP → **published** artifacts: normative Agentic REST Response Profile + 6 JSON
schemas + OpenAPI components, Go gateway, FastAPI SDK, Python + TS clients, CLI
validator, Campaign reference example, Compose demo, and public publishing
(PyPI×3 + npm + GHCR, all OIDC). K8s/Helm/Cosign/SBOM-attestation hardening =
**PI2-yaa** (NOT pulled in — PRD §11). GTM content = founder-owned.

## ADR slate (all [Accepted], `docs/adr/`)

| ADR | Decision | Resolves |
|-----|----------|----------|
| PI1-yaa-0001 | Gateway = net-new standalone OSS **base** (not consumer/fork of internal gateways); thin `net/http` proxy; JWT HS256-dev + JWKS-prod | OQ-2; internal-gateway lineage (mandatory cite on all GW WIs) |
| PI1-yaa-0002 | `spec/` sole authoritative table source; JSON Schema Draft-07; `schemas/v0.1/` path; 202 schema-only (polling = v0.2); shared golden corpus | OQ-3 |
| PI1-yaa-0003 | Hatch for 3 PyPI pkgs; dual ESM+CJS zero-dep TS client; per-package flat config | OQ-4, OQ-6 |
| PI1-yaa-0004 | Dual-license source-available; combined-AND threshold placeholder; legal review = launch gate | OQ-1 (routes lock to chief-architect/user @ PC-6) |
| PI1-yaa-0005 | OIDC trusted publishing everywhere (PyPI TP / npm provenance / GHCR token); SBOM ships, Cosign=PI2-yaa | OQ-5 (SBOM format → platform-engineer A-4) |

OQ-5 (SBOM format) deliberately left to platform-engineer A-4. OQ-1 threshold
number is a chief-architect/user lock before PC-6 — not an engineering blocker.

## Sprint plan (5 sprints; contract first, S5 reserved for e2e+conformance)

| Sprint | Theme | WIs | Component files |
|--------|-------|-----|-----------------|
| **S1** | Contract foundation + scaffolding | SPEC-1..5, REL-1, REL-2 | `spec-schemas-openapi.md`, `release-and-publish.md` |
| **S2** | Go gateway core | GW-1..5 | `gateway.md` |
| **S3** | FastAPI SDK + CLI validators | SDK-1..4, CLI-1..3 | `sdk-fastapi.md`, `cli.md` |
| **S4** | Clients + reference server | PYC-1..3, TSC-1..3, EX-1, EX-2 | `client-python.md`, `client-ts.md`, `examples-campaign-api.md` |
| **S5** | Compose e2e + conformance gate + publish | EX-3, CLI-4, EX-4, REL-3..6 | `examples-campaign-api.md`, `cli.md`, `release-and-publish.md` |

S5 honours runbook rule 3 (Compose end-to-end + CLI conformance gate reserved
for the last sprint: EX-3 demo → CLI-4 conformance-test → EX-4 PI gate).

## WI index (30 WIs)

- **Contract** (`spec-schemas-openapi.md`): SPEC-1 spec+table, SPEC-2 6 schemas, SPEC-3 OpenAPI components, SPEC-4 `x-yaagents` ext, SPEC-5 golden corpus
- **Gateway** (`gateway.md`): GW-1 skeleton+routes, GW-2 authn, GW-3 context/corr-id, GW-4 RBAC+passthrough, GW-5 audit+health+metrics
- **FastAPI SDK** (`sdk-fastapi.md`): SDK-1 AgenticResponse, SDK-2 Context+RequiredInput, SDK-3 decorator+OpenAPI-gen, SDK-4 conformance tests
- **CLI** (`cli.md`): CLI-1 validate-response, CLI-2 validate-openapi, CLI-3 init, CLI-4 conformance-test (S5)
- **Python client** (`client-python.md`): PYC-1 client+accessors, PYC-2 typed exceptions, PYC-3 corpus tests
- **TS client** (`client-ts.md`): TSC-1 client+accessors, TSC-2 result-union+strict, TSC-3 dual-build+tests
- **Reference example** (`examples-campaign-api.md`): EX-1 FastAPI server, EX-2 routes.yaml, EX-3 Compose demo (S5), EX-4 e2e conformance gate (S5)
- **Release/publish** (`release-and-publish.md`): REL-1 scaffolding, REL-2 license, REL-3 PyPI×3, REL-4 npm, REL-5 GHCR+SBOM, REL-6 CI matrix+spec archive

## Critical path & dependency spine

```
SPEC-1/2 ──► GW-* (S2) ─────────────────┐
        ├──► SDK-1..4 (S3) ──► EX-1 ─────┤
        ├──► CLI-1..3 (S3)               ├─► EX-2 ─► EX-3 ─► CLI-4 ─► EX-4 (PI gate)
        ├──► PYC-1..3 (S4) ──────────────┤                              │
        └──► TSC-1..3 (S4) ──────────────┘                              ▼
                                                      REL-3/4/5/6 (publish, S5)
SPEC-5 (golden corpus) feeds SDK-4 / PYC-3 / TSC-3 / CLI-1 conformance.
REL-1/2 (scaffolding+license) run S1 in parallel — no contract dependency.
```

Everything `depends_on` SPEC (contract gates all) → spec/schema WIs are S1,
sequenced first per runbook rule 2.

## Library gates (library-gates.md Gate 3)

yaagents is net-new public OSS — **no `portfolio/LIBRARIES.md` entry applies**.
Every feature WI carries `library_justify: novel; standalone OSS surface`
**except** all `GW-*` + EX-2 which carry `library_ref: ADR PI1-yaa-0001`
(mandatory internal-gateway-lineage citation: the yaagents gateway is the
future BASE for `platform-services/gateway`+`ai-gateway`, a reverse dependency,
NOT a consumer — no cross-product import in PI1-yaa). REL-2 cites ADR
PI1-yaa-0004; REL-3/4/5 cite ADR PI1-yaa-0005. Gate 4: PYC vs TSC are
intentional dual-language clients of one contract (override noted in both
component files — heuristic does not fire on cross-language pairs).

## Out of scope (PRD §11 — explicitly NOT in PI1-yaa)

K8s manifests · Helm · K8s guide · Cosign · SBOM attestation hardening →
**PI2-yaa** (seed authored at PI1-yaa PC-6). Spring/ASP.NET adapters · OTel ·
OPA · LangGraph/SK plugins · async-operation polling runtime ·
approval-flow runtime → v0.2+. Internal gateway re-base → future seed.
GTM content → founder-owned.

## Success criteria (PRD §12)

§12.1–9 verified by **EX-4** (the e2e conformance gate WI). §12.10–12
(public-registry installability) verified by **REL-3/4/5**. PI1-yaa is DONE
when EX-4 reports `Overall: PASS` and all 12 criteria are checked green.

## Handoff

```
next:        platform-engineer
artifact:    yaagents/docs/PI1-yaa/roadmap.md + 8 component files + docs/adr/PI1-yaa-0001..0005.md
intent:      A-4 NFR / supply-chain pass — append [SEC]/[SRE]/[SUPPLY-CHAIN] WI
             bodies, run compose-linter on EX-3, pick SBOM format (OQ-5),
             state [FIN] N/A explicitly, flip all PI1-yaa WIs [DRAFT]→[READY].
cwd:         yaagents/
```
