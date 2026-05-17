# YAAgents v0.1 — Agentic REST Profile MVP (PI1-yaa) — PRD Seed
Status: [DRAFT]
Target product: yaagents
Target services: spec/schemas/openapi, gateway (Go), sdk-fastapi (Py), client-python (Py), client-ts (TS), cli (Py), examples/campaign-api, docker-compose demo
Owner: product-manager (yaagents)

## Problem

Agentic capabilities are exposed as generic chat (`/agents/invoke`) or framework-specific endpoints: loosely typed I/O, inconsistent access control, ad-hoc clarification, hard to document/test, and existing API-gateway/microservice practice is bypassed. Production systems are resource-oriented; chat is the wrong integration surface for business workflows.

## Why now

User-directed parallel track. The yaagents Go gateway is taken on priority because the internal generic gateway (`platform-services/gateway` + `ai-gateway`) will later be **re-based on top of it** — getting the standalone OSS gateway + Agentic REST contract right first unblocks that convergence and the source-available public launch. Runs on the new `yaa` lane parallel to PI15-plt-aip so the in-flight CIAM work is undisturbed.

## Rough scope

- In: Agentic REST Response Profile (normative status × media-type table) · 6 JSON schemas · reusable OpenAPI components (`x-yaagents` metadata) · Go gateway (authn, tenant/actor context, route-level RBAC, audit log, typed-response passthrough, YAML route config) · `yaagents-fastapi` server SDK · `yaagents-client` (Python) · `@yaagents/client` (TS) · `yaagents-cli` validator (`validate-openapi`/`validate-response`/`conformance-test`/`init`) · Campaign reference example · Docker Compose demo · **publish: PyPI ×3 + npm ×1 + GHCR multi-arch gateway image + SBOM, all via OIDC trusted publishing** · dual-license (Community + Commercial, source-available; legal-review-pending disclaimer carried verbatim) + repo scaffolding (README, CONTRIBUTING, SECURITY, CoC, issue templates).
- Out: Kubernetes manifests / Helm chart / K8s guide / Cosign signing / SBOM attestation hardening → **PI2-yaa**. GTM content (videos, launch blog, HN/PH/social) → founder-owned, not an engineering PI. v0.2+ (Spring/ASP.NET adapters, OTel, OPA, LangGraph/SK plugins, async-operation profile, approval-flow runtime) → later PIs.

## Dependencies

- New `yaa` lane + product onboarding done at A-0 2026-05-17 (repo init, agent roster, registration, PROCESS [ADOPTED]).
- External: PyPI / npm / GHCR projects + OIDC trusted-publisher config (a Phase-B WI; no long-lived tokens).
- Forward (NOT this PI): future seed to re-base `platform-services/gateway` + `ai-gateway` on this gateway — recorded in PRIORITIES so it is not lost.
- Root `docker-compose.yml` `include:` of the yaagents demo compose is deferred until that file exists (Phase B); chief-architect adds the include then.

## Success signal

A developer runs `docker compose up` in `examples/campaign-api/` and exercises `created`, `clarification_required`, `validation_failed`, and `failed_dependency` through the gateway with route-RBAC enforced and correlation-id propagated; `yaagents-cli conformance-test` returns PASS; and `pip install yaagents-fastapi` / `npm install @yaagents/client` / `docker pull ghcr.io/yaagents/gateway:0.1.0` succeed from the public registries.
