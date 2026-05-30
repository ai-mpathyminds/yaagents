# YAAgents v0.2 — Apache 2.0 + Plugin Middleware + Go Client + LLM Gateway Convergence (PI2-yaa) — PRD Seed
Status: [READY]

> Seeded by chief-architect. Expanded by product-manager 2026-05-30.
Target product: yaagents
Target services: yaagents/gateway (plugin system refactor + LLM-specialization MOVE from ai-platform), yaagents/client-go (NEW), yaagents/plugins/{token-validator,tenant-injector,license-check} (NEW; required), yaagents/plugins/{prompt-sanitize,otel-audit} (NEW; stubs/off-by-default), yaagents/examples/llm-gateway (NEW), yaagents top-level LICENSE + COMMERCIAL.md retire + copyright headers · **cross-lane stretch (added 2026-05-30T19:30Z user-direct)**: ai-platform/services/ai-gateway (DELETE) + ai-platform/docker-compose.yml ai-gateway entry (DELETE) + ai-platform/.github/workflows/ai-gateway-deploy.yml (DELETE) + portfolio/infrastructure/roots/{compute,ci,observability}/ ai-gateway TF resources (TF-teardown: ECS module call + ALB target group + listener rule + ECR repo + SSM params + IAM role policy scopes)
Owner: product-manager (yaagents)

## Problem

v0.1.x ships under the source-available "YAAgents Community License" (paywall for 10+ employees / >$1M revenue) and exposes a **hardcoded middleware chain** (authn / RBAC / audit baked into the gateway binary). Two compounding consequences:

1. Community contribution is structurally gated — the non-OSI license discourages external authors, and even if someone wanted to add a middleware (e.g. their own tenant model, an OPA hook, a custom OTEL pipeline), there is no plugin contract to extend.
2. `ai-platform/ai-gateway` (the LLM-routing Go service) was built as a separate codebase precisely because yaagents/gateway has no plugin model. The PI1-yaa forward-dep "re-base internal gateway on yaagents/gateway" cannot land until that plugin model exists. Today both gateways drift on auth/tenant/audit — the duplication is the symptom.

## Why now

User-direct decision flip (2026-05-30): the original source-available stance is reversed. PI1-yaa is publishing v0.1.x packages today (REL-3/4/5/6 landed under Community License); v0.2.0 is the clean SemVer minor cut to Apache 2.0. Coupling the license flip with the plugin refactor + ai-gateway absorption + Go client in one PI is intentional: (a) Apache 2.0 enables the community plugin flywheel, (b) the plugin model is the prerequisite for ai-gateway convergence, (c) v0.2.0 is the OSS launch moment — Go client closes the SDK trifecta (Py/TS/Go). Staggering across three PIs is strictly worse.

## Rough scope

- **In**:
  - **License flip**: `LICENSE` → Apache 2.0 (text verbatim); `COMMERCIAL.md` retired with README pointer; copyright-header sweep across `gateway/`, `sdk-fastapi/`, `client-python/`, `client-ts/`, `cli/`, `examples/`; v0.2.0 release notes flag the discontinuity (v0.1.x stays Community).
  - **Plugin middleware system**: typed plugin interface (`Plugin{Name(), Init(cfg) error, Handler(next http.Handler) http.Handler, Shutdown(ctx) error}`); lifecycle (`Init` at boot, `Handler` chained per request, `Shutdown` on SIGTERM); per-plugin YAML config block under route definitions; plugin registry for community-contributed plugins; hot-reload deferred to PI3-yaa.
  - **Three required plugins (a/b/c)**: **token-validator** (always-on; cannot be disabled via config; JWT RS256 via JWKS; reuses `portfolio/packages/go/auth-jwks/` IFF PI14-oppor extracted it — architect call at A-3); **tenant-validator + injector** (default-on, disable-able; parse tenant header, validate vs allowlist, inject `actor_ctx`); **license-check** (default-on, disable-able; product-license token verification for commercial yaagents consumers).
  - **Two stub plugins (d/e) — off-by-default**: **prompt-sanitize** (interface + stub; full prompt-injection defense deferred to PI3-yaa or community); **otel-audit** (interface + stub; OpenTelemetry trace/log emission to configurable collector; reference collector wiring deferred).
  - **Go client SDK** `yaagents/client-go/`: resource-oriented client analog to client-python/client-ts; idiomatic `clarification_required` handling via typed Go errors; published via Go modules tag-driven (`go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` → proxy.golang.org).
  - **ai-gateway absorption**: pull LLM-specific concerns (provider routing, streaming, model selection, prompt envelope) from `ai-platform/ai-gateway` into yaagents — architect-call at A-3 between (i) a `yaagents/gateway` LLM-specialization layer activated by config, or (ii) a sibling `yaagents/llm-gateway` component depending on `yaagents/gateway`. Reference example at `yaagents/examples/llm-gateway/` proving the abstraction has a consumer.

- **Out**:
  - **Consumer migration to alternate LLM gateway** — N/A: user-verified 2026-05-30T19:30Z that **no production consumers exist on `ai-platform/services/ai-gateway`**; decommission proceeds without a migration WI. (Original out-of-scope row "ai-platform/ai-gateway decommission + consumer migration" SUPERSEDED 2026-05-30T19:30Z — ai-platform/ai-gateway decommission is now **IN scope**, see Target services line; only the consumer migration sub-item remains out, as trivially N/A.)
  - **Reference impls for sanitize + OTEL plugins** — interface ships; impls deferred to PI3-yaa or community.
  - **K8s/Helm/Cosign/SBOM-attestation hardening** — displaced from old-PI2-yaa scope to PI3-yaa.
  - **Plugin marketplace UI / discovery service** — registry contract only; no UI, no central registry server.
  - **Frontend / docs site** — no UI surface (consistent with PI1-yaa zero-UI evidence).
  - **Retroactive re-licensing of v0.1.x** — v0.1.x already-published packages stay Community License.

## Dependencies

- **PI1-yaa Phase B sufficiently complete** before PI2-yaa Phase B dispatches: yaagents/gateway at v0.1.x publishable; spec/schemas/openapi stable. Verified 2026-05-30: REL-3/4/5/6 landed; gateway publishable. PI2-yaa Phase A (planning) can run in parallel; Phase B execution gates on PI1-yaa PC-6 close.
- **portfolio/packages/go/auth-jwks/** (if PI14-oppor extracted it): token-validator plugin reuses it; if absent, plugin re-implements minimally and the extraction proposal stays open in `portfolio/LIBRARIES.md`.
- **Apache 2.0 commercial-model impact**: COMMERCIAL.md commercial-paywall retires entirely. Legal-review-pending disclaimer travels with the license-flip WI until a lawyer signs off; lawyer sign-off gates public re-announce, not PI2-yaa close.
- **`ai-platform/ai-gateway/` source access**: yaagents repo doesn't own ai-platform code today; either (i) copy under Apache 2.0 with attribution + retire ai-platform copy in a future PI, or (ii) make ai-platform/ai-gateway a yaagents consumer that downloads via Go modules — architect-call at A-3 (ADR PI2-yaa-0002).

## Success signal

A community developer `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` + `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0`, drops a YAML config enabling **only the token-validator plugin** (`plugins: { token-validator: { enabled: true, jwks_url: ... } }` with b/c/d/e off), starts the gateway, calls a protected route with a valid JWT, gets the standard Agentic REST `created` response. Separately, with the same image and a different YAML enabling token + tenant + license + a community-authored example plugin loaded via the registry contract, the request flows through all four. Apache 2.0 `LICENSE` in repo root; `pip install yaagents-fastapi==0.2.0` reports Apache 2.0 metadata; `yaagents/examples/llm-gateway/docker compose up` runs an LLM provider call end-to-end through the absorbed LLM-specialization.
