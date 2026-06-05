# yaagents v0.4 — One-Pager
Status: [DRAFT]

> Seeded by chief-architect. Expanded by product-manager 2026-06-05.

## What

yaagents v0.4 closes three production-readiness gaps left after the PI3-yaa public launch
(PyPI ×3 + npm + GHCR + Pages + Go modules + ai-platform/agent-api canary, smoke 10/10 PASS).
The runtime and publish substrate are real; PI4-yaa makes the surface production-grade and intelligible.

## Goals

### Goal 1 — sdk-go parity with sdk-fastapi (Profile v0.3 audit-emission surface)

`sdk-go` lacks the Profile v0.3 audit-API surface that `sdk-fastapi` already exposes, blocking
AmpathyMinds' own Go-native agent migration. PI4-yaa closes this specific gap.
Scope: **strict Profile v0.3 surface only** — Python-side conveniences not replicated unless
profile-mandatory. Conformance bar is the spec, not `sdk-fastapi`.

**Success signal (seed §1):** sdk-go audit-API e2e test exercising Profile v0.3 audit-emission
contract passes against a real gateway + sdk-go service.

### Goal 2 — Gateway plugin production-grade implementation + documentation

All 5 first-party gateway plugins (token-validator, tenant-injector, license-check,
prompt-sanitize, otel-audit) flip **Preview→Stable** with full implementation + per-plugin
Pages docs + e2e test exercising through the gateway. "Coming in v0.4" markers removed
from `production-checklist`, `plugin-pipeline`, `reference-architecture`, and per-plugin pages.

Per-plugin config highlights (user-direct 2026-06-05):
- **token-validator**: multi-IDP / pluggable JWT issuer. Architect designs plug-point at A-3.
- **tenant-injector**: URL-based webhook resolution. Architect investigates PI2-yaa PLG-4b v2 FIRST.
- **license-check, prompt-sanitize, otel-audit**: architect discovers existing contracts at A-3.

Stable bar (default-confirmed): feature-complete + tested + documented.
Benchmarking + load-testing deferred to PI5-yaa.

**Success signal (seed §2):** all 5 plugins ship Stable badge + per-plugin Pages docs + e2e green.

### Goal 3 — README + Pages narrative pivot to ecom-recommendation use case

Drop A2A vs agentic comparison framing. Replace with: (a) target-audience callout above the fold;
(b) e-commerce recommendations real-life use case end-to-end (builds on `examples/store/` Python +
`examples/store-go/` Go); (c) "who uses yaagents" framing for both agents-familiar and agents-curious
audiences.

README ↔ `start-here/overview.mdx`: **independent surfaces** (README = GitHub-reader one-screen
decision funnel; overview.mdx = Pages-reader deeper read; two separate authoring WIs).
README ↔ case-study page: **same content with cross-link** (`case-studies/ecommerce-product-recommendations.mdx`
embeds or cross-links the README ecom-narrative — single source of truth).

**Success signal (seed §3):** reviewer test with 2+ external developers reports clear understanding
of "what yaagents is for + what it lets them build" reading README → ecom case study → overview.mdx.

## Tracks (PI3-yaa carry-forwards)

| Track | Source | Key Deliverable |
|-------|--------|-----------------|
| **INFRA** | Process carry-forward 2/3 | `bin/pi-topology.sh` self-discovery refactor; PRIORITIES mechanical-close trigger |
| **GH** | INTAKE-1 | ROADMAP/GOOD_FIRST_ISSUES/ADOPTERS rows filed as actionable GH issues |
| **DOCS** | INTAKE-4, INTAKE-5 | `how-to/host-in-production.mdx` (single-VM/docker-compose/TLS); `architecture/audit-and-observability.mdx` (couples with otel-audit Goal 2) |
| **EXAMPLES** | INTAKE-7 | 2 domain-breadth skeleton services: `examples/customer-support-triage/` (Python) + `examples/financial-risk-screening/` (Python) |
| **PROCESS** | INTAKE-8 + Process 1/4 | Cross-lane co-sign formalization; yaagents-architect ratification confirmed; `pi2-yaa-postmortem.yml` committed at A-1 |

## Launch Readiness Gate

PI4-yaa closes via a launch-readiness smoke report at
`portfolio/REPORTS/smoke/PI4-yaa-launch-<sha>.md` (mirror of PI3-yaa LA-PI-GATE pattern).
All three Goal success signals MUST be met before PI close.

Full PRD: `yaagents/system-refs/yaagents-v0.4_detailed.md`
