# yaagents v0.4 — PRD Seed
Status: [READY]

> Seeded by chief-architect. Expanded by product-manager 2026-06-05.
Target product: yaagents
Target services: yaagents/{sdk-go, gateway/internal/plugins/{token-validator,tenant-injector,license-check,prompt-sanitize,otel-audit}}, yaagents/docs (Astro Starlight Pages), yaagents (meta-repo root README)
Owner: product-manager (yaagents) → yaagents-architect at A-3

## Problem

PI3-yaa shipped public launch (PyPI ×3 + npm + GHCR + Pages + Go modules + agent-api canary, smoke 10/10 PASS) but the v0.3 surface has three production-readiness gaps: (1) sdk-go lacks the Profile v0.3 audit-emission surface that sdk-fastapi already exposes, blocking AmpathyMinds' own Go-native agent migration; (2) all 5 first-party gateway plugins (token-validator, tenant-injector, license-check, prompt-sanitize, otel-audit) ship with stub bodies or partial implementations behind Preview maturity badges; (3) the meta-repo root README leans on A2A vs agentic comparison framing that 2 external reviewers report is opaque — readers can't translate it into "what can I build."

## Why now

These three are interdependent v0.4 launch-readiness contracts. Partial delivery (sdk-go without plugins; plugins without narrative; narrative without parity) ships a degraded v0.4. PI3-yaa shipped the runtime; PI4-yaa makes it production-grade and intelligible. AmpathyMinds' own Go agent migration is the immediate downstream consumer of sdk-go parity; community adoption is the immediate downstream consumer of plugin-Stable + clear README.

## Rough scope

- **In**: (1) sdk-go audit API addition + Profile v0.3 strict-surface conformance gap-close (audit-first WI then close WIs); (2) all 5 plugins Preview→Stable with full implementation + per-plugin Pages docs + e2e tests + maturity-badge flips + "Coming in v0.4" marker removal; (3) README narrative pivot — drop A2A/agentic comparison, add target-audience callout, e-commerce-recommendations real-life use case end-to-end (uses existing `examples/store/` + `examples/store-go/`), "who uses yaagents" framing; case-study page `case-studies/ecommerce-product-recommendations.mdx` shares content with README via cross-link; `start-here/overview.mdx` independently rewritten Pages-reader-oriented; (4) **8 PI3-yaa carry-forwards**: GH Issues backlog mirror, production-hosting guide, audit-and-observability consolidation page, 2 domain-breadth example skeletons (customer-support-triage + financial-risk-screening), cross-lane co-sign formalization; (5) **4 process structural decisions**: yaagents-architect ratification (file already exists), pi-topology.sh self-discovery refactor (platform-engineer lane), PRIORITIES mechanical-close trigger, pi2-yaa-postmortem.yml commit; (6) **7 adoption-signal verifications** from PI3-yaa PC-5 [PROPOSED] PROCESS deltas (PC-5-02/03/06/08/10/11/12 close-time validations).
- **Out**: AWS substrate or `portfolio/infrastructure/` touch (no cloud-state-grounder dispatch at A-1 needed). Helm chart (PI5-yaa). Plugin benchmarking + load-testing (PI5-yaa). sdk-go feature-for-feature with sdk-fastapi beyond Profile v0.3 (Python-side conveniences not replicated unless profile-mandatory). Headline case study beyond ecom (other domains are skeleton-only INTAKE-7 scope).

## Dependencies

- PI3-yaa PC-6 close logged (commit 380efa1 on origin/main 2026-06-05; gate cleared)
- yaagents-architect agent file at `portfolio/yaagents-internal/.claude/agents/yaagents-architect.md` (already present per PI3-yaa A-3; Process-1 ratifies)
- `gh-pi-merge.sh` B-44a unrelated-histories tolerance (PI3-yaa PC-5-13 commit 28085a6; already merged; PI4-yaa close benefits)
- No external deps: no AWS, no third-party services, no cross-lane gating beyond optional ai-platform/agent-api canary re-run

## Success signal

(1) sdk-go audit-API e2e test exercising Profile v0.3 audit-emission contract passes against a real gateway+sdk-go service; (2) all 5 plugins ship with Stable maturity badge + per-plugin Pages docs + e2e green; (3) reviewer test with 2+ external developers reports clear understanding of "what yaagents is for + what it lets them build" reading README → ecom case study → overview.mdx flow. Captured at PI4-yaa close via a launch-readiness smoke report at `portfolio/REPORTS/smoke/PI4-yaa-launch-<sha>.md` (mirror of PI3-yaa LA-PI-GATE pattern).
