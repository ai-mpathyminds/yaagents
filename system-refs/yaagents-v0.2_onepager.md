# YAAgents v0.2 — Apache 2.0 OSS + Plugin Middleware + Go Client + LLM Gateway Convergence — One-Pager

Status: [READY]
Owner: product-manager (yaagents)
PI: PI2-yaa

> Seeded by chief-architect. Expanded by product-manager 2026-05-30.

---

## Problem

YAAgents v0.1.x ships under the "YAAgents Community License" (source-available; paywall for
organisations with ≥10 employees or ≥USD 1M annual revenue) and exposes a **hardcoded
middleware chain** — authn, tenant, RBAC, and audit baked as one non-extensible block inside
the gateway binary. Two compounding consequences:

1. **Community contribution is structurally gated.** The non-OSI license discourages external
   contributors; even willing contributors have no plugin contract to extend. The gateway
   cannot accept middleware from the community without a fork.
2. **`ai-platform/ai-gateway` drifts in parallel.** The LLM-routing Go service was built as a
   separate codebase precisely because yaagents/gateway has no plugin model. The PI1-yaa
   forward-dependency ("re-base internal gateway on yaagents/gateway") cannot land until that
   plugin model exists. Today both gateways carry diverging copies of JWT auth, tenant
   injection, and audit.

---

## Solution

Four user-direct asks bundled into one PI:

1. **Apache 2.0 license flip** — v0.2.0 ships under Apache 2.0; COMMERCIAL.md retired. v0.1.x
   packages stay under the Community License (non-retroactive). Legal-review-pending disclaimer
   travels with the license-flip WI until a lawyer signs off; lawyer sign-off gates the public
   re-announce, not PI2-yaa close.
2. **Go client SDK** (`client-go/`) — closes the SDK trifecta (Python / TypeScript / Go);
   resource-oriented; typed `ClarificationRequired` / `ValidationFailed` / `FailedDependency`
   errors; published via Go modules tag-driven at `github.com/ai-mpathyminds/yaagents/client-go@v0.2.0`.
3. **Plugin middleware system** — typed `Plugin` interface + `Init → Handler chain → Shutdown`
   lifecycle; per-plugin YAML config block under routes; community-extensible registry;
   five first-party plugins (a–e).
4. **LLM gateway convergence** — LLM-specialisation concerns from `ai-platform/ai-gateway`
   (SSE streaming, per-tenant SSE concurrency limits, execution timeout, CORS) absorbed into
   yaagents; shape (layer-atop vs sibling) ratified by ADR PI2-yaa-0002 at A-3; reference
   example at `examples/llm-gateway/` proves the abstraction works.

---

## Scope (PI2-yaa)

**In:**

| Component | Path | Notes |
|-----------|------|-------|
| Gateway plugin system refactor | `gateway/` | Plugin interface + lifecycle + registry + YAML config schema |
| Plugin: token-validator | `plugins/token-validator/` | JWT RS256/JWKS; **always-on; cannot be disabled** |
| Plugin: tenant-injector | `plugins/tenant-injector/` | Header parse + actor-ctx inject; default-on |
| Plugin: license-check | `plugins/license-check/` | Product-license token verify; default-on |
| Plugin: prompt-sanitize | `plugins/prompt-sanitize/` | Interface + stub; off-by-default |
| Plugin: otel-audit | `plugins/otel-audit/` | OTEL interface + stub; off-by-default |
| Go client SDK | `client-go/` | Py/TS parity; resource-oriented |
| LLM gateway reference example | `examples/llm-gateway/` | Absorbed LLM-specialisation; Compose demo |
| Apache 2.0 license sweep | repo root + all components | LICENSE → Apache 2.0; COMMERCIAL.md retired; SPDX header sweep |

**Out of PI2-yaa → PI3-yaa:** Kubernetes manifests, Helm chart (GHCR OCI), Cosign image
signing, SBOM attestation hardening, prompt-sanitize reference impl, OTEL plugin reference
impl.

**Out of scope entirely:** ai-platform consumer migration off `ai-platform/ai-gateway`
(future plt-aip-lane PI); GTM content; plugin marketplace UI / discovery service.

---

## Success Signal

A community developer:
- `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` succeeds.
- `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` succeeds.
- Configures only `plugins.token-validator.enabled: true` (b/c/d/e off), fires a request with a valid
  JWT, receives `201 application/json` (Agentic REST `created`).
- With token + tenant + license + a community-contributed plugin all enabled, the request flows
  through all four middleware handlers in chain order.
- `pip install yaagents-fastapi==0.2.0` — `pip show` reports `License: Apache-2.0`.
- `cd examples/llm-gateway && docker compose up` exercises an LLM provider call end-to-end
  through the absorbed specialisation.
- `yaagents/LICENSE` file contains the Apache 2.0 text; no `COMMERCIAL.md` in the repository.

---

## Target Users

| Tier | Who |
|------|-----|
| External (primary) | Go/Python/TS engineers building governed agentic APIs; OSS contributors authoring plugins |
| External (secondary) | Teams self-hosting yaagents-gateway with custom auth/tenant/license/observability stacks |
| Internal (forward dep) | AimpathyMinds — future migration of `ai-platform/ai-gateway` consumers (separate plt-aip-lane PI, not PI2-yaa scope) |

---

## Why Now

User-direct strategic flip (2026-05-30): the original "source-available, not OSI" stance is
reversed. The cost of not flipping — gated community contribution and the ai-gateway
forward-dependency stuck in purgatory — outweighs the lost commercial-paywall optionality.
Apache 2.0 enables the community-contributable plugin model; the plugin model is the
architectural prerequisite for ai-gateway convergence; the Go client closes the SDK trifecta
for the v0.2.0 OSS launch moment. All three are interdependent; staggering across three PIs
is strictly worse.

---

## Handoff

Next: `yaagents-architect` — PI2-yaa WI breakdown at `yaagents/docs/PI2-yaa/` + ADRs
PI2-yaa-0001..0005 at `yaagents/docs/adr/`.

Artifact: `yaagents/system-refs/yaagents-v0.2_detailed.md` [READY]
