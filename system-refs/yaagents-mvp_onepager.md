# YAAgents v0.1 — Agentic REST Profile MVP — One-Pager

Status: [READY]
Owner: product-manager (yaagents)
PI: PI1-yaa

> Seeded by chief-architect. Expanded by product-manager 2026-05-17.

---

## Problem

Production systems are resource-oriented; chat endpoints (`/agents/invoke`) are the wrong
integration surface for business workflows. Exposing agentic capabilities this way produces
loosely typed I/O, inconsistent access control, ad-hoc clarification, and bypasses the API
gateway and microservice practices teams already rely on. Agent frameworks leak into the
external integration model.

---

## Solution

YAAgents provides a **normative Agentic REST Response Profile** — a fixed table of HTTP
status codes × vendor media types — plus the full implementation stack that enforces it:
a Go gateway, Python and TypeScript SDKs, JSON schemas, OpenAPI components, a CLI
validator, and a Campaign reference example. The governing principle:

> Keep your domain resources. Make selected operations agentic.
> Build the agent however you want. Expose it like a governed API.

---

## Scope (PI1-yaa)

**In:**

| # | Component | Artifact path |
|---|-----------|--------------|
| 1 | Agentic REST Response Profile (spec) | `spec/` |
| 2 | JSON schemas (6) | `schemas/` |
| 3 | OpenAPI reusable components | `openapi/` |
| 4 | Go Gateway | `gateway/` |
| 5 | Python FastAPI SDK | `sdk-fastapi/` |
| 6 | Python Client | `client-python/` |
| 7 | TypeScript Client | `client-ts/` |
| 8 | CLI Validator | `cli/` |
| — | Campaign reference example | `examples/campaign-api/` |
| — | Docker Compose demo | `examples/campaign-api/docker-compose.yml` |
| — | Published packages (OIDC) | PyPI ×3 · npm ×1 · GHCR multi-arch image |
| — | Repo scaffolding | README · CONTRIBUTING · SECURITY · CoC · issue templates |
| — | Dual-license | Community + Commercial (source-available) |

**Out of PI1-yaa → PI2-yaa:** Kubernetes manifests, Helm chart (GHCR OCI), Cosign image
signing, SBOM attestation hardening.

**Out of both PIs:** GTM content (demo videos, launch blog, social), v0.2+ adapters
(Spring/ASP.NET, OTel, OPA, LangGraph/SK plugins, async-operation profile, approval-flow
runtime).

---

## Success Signal

A developer runs `docker compose up` in `examples/campaign-api/`, exercises `created`,
`clarification_required`, `validation_failed`, and `failed_dependency` through the gateway
with route-RBAC enforced and correlation-id propagated; `yaagents-cli conformance-test`
returns PASS; and `pip install yaagents-fastapi` / `npm install @yaagents/client` /
`docker pull ghcr.io/yaagents/gateway:0.1.0` succeed from the public registries.

---

## Target Users

| Tier | Who |
|------|-----|
| External (primary) | Platform engineers, API architects, backend engineers, AI-platform teams building governed agentic APIs |
| External (secondary) | FastAPI developers, OpenAPI-first teams, LangGraph/Semantic Kernel users, enterprise architects |
| Internal (forward dep) | AimpathyMinds — future re-base of `platform-services/gateway` + `ai-gateway` on this gateway |

---

## Why Now

User-directed parallel track (2026-05-17). The yaagents Go gateway is prioritized because
`platform-services/gateway` + `ai-gateway` will be **re-based on top of it** — getting
the standalone OSS contract right first unblocks that convergence and the public
source-available launch. Runs on lane `yaa` parallel to PI15-plt-aip (CIAM); does not
disturb the in-flight Lane B work.

---

## Handoff

Next: `yaagents-architect` — expand into PI1-yaa roadmap at `yaagents/docs/PI1-yaa/`
(roadmap.md + per-component WI files + ADRs under `yaagents/docs/adr/`).

Artifact: `yaagents/system-refs/yaagents-mvp_detailed.md` [READY]
