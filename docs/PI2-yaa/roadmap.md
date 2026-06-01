# PI2-yaa — YAAgents v0.2 Apache 2.0 + Plugin Middleware + Go Client + LLM Gateway Convergence — Roadmap (master)

Status: [READY] (A-4 platform-engineer NFR pass complete 2026-06-01; all component WIs [READY])
PI: PI2-yaa · Lane: yaa (parallel to Lane A oppor + Lane B plt-aip)
Author: yaagents-architect · Date: 2026-05-30
PRD: `yaagents/system-refs/yaagents-v0.2_detailed.md` (§ refs throughout)
Seed: `yaagents/system-refs/yaagents-v0.2.seed.md` [READY]
Phase B gate: PI1-yaa PC-6 close (PI1-yaa Phase B still in flight as of A-3).

## Scope

Four tightly coupled capabilities ship together as v0.2.0:

1. **Apache 2.0 license flip** — `LICENSE` → Apache 2.0 verbatim; `COMMERCIAL.md`
   retired; SPDX header sweep across all source components. v0.1.x packages
   stay Community License (**non-retroactive** — ADR PI2-yaa-0003).
2. **Plugin middleware system** — typed `Plugin` interface + Init/Handler/
   Shutdown lifecycle + YAML config; five first-party plugins (token-validator
   always-on; tenant-injector + license-check default-on; prompt-sanitize +
   otel-audit stubs off-by-default); community-extensible via Go module import.
3. **Go client SDK** `client-go/` — Py/TS parity; resource-oriented; typed Go
   errors + `AgenticResult` result-style; stdlib-only; published via Go module
   proxy at tag `client-go/v0.2.0`.
4. **LLM gateway convergence + ai-platform/ai-gateway decommission** —
   `ai-platform/ai-gateway` LLM-specialisation (SSE proxy, per-tenant SSE
   limiter, execution timeout, CORS) **MOVED** into yaagents under
   Apache 2.0 (ADR PI2-yaa-0002 Option A layer-atop, amended
   2026-05-30T19:30Z user-direct MOVE not COPY; ai-platform-side
   `services/ai-gateway/` deleted in same PI via chief-architect
   cross-lane direction). **No production consumers exist on
   ai-platform/ai-gateway**, so breaking changes are welcome — no
   backward-compat shims for old route paths, env-var names, config
   field names, or Keycloak-realm-aware JWKS construction. Reference
   example at `examples/llm-gateway/`.

K8s/Helm/Cosign/SBOM-attestation hardening + reference impls for
prompt-sanitize/otel-audit = **PI3-yaa** (seed authored at PI2-yaa PC-6).
Consumer migration off ai-platform/ai-gateway = trivially N/A (no
consumers). ai-platform-side cleanup (delete `services/ai-gateway/`,
compose entry, deploy workflow, portfolio TF) routes to chief-architect
at A-3 handoff — those lanes are outside yaagents-architect writable
paths (ai-platform-architect + platform-engineer-portfolio).

## ADR slate (all [Accepted], `docs/adr/`)

| ADR | Decision | Resolves |
|-----|----------|----------|
| PI2-yaa-0001 | Plugin interface: no `Version()` method; versioning by Go module tag (`gateway/plugin/v0.2`); declaration-order = execution-order; reverse-order Shutdown; static import-side-effect registration (no `plugin.Open`/`dlopen`) | OQ-6 |
| PI2-yaa-0002 | ai-gateway absorption = **Option A (layer-atop)**; LLM-specialisation lives in `yaagents/gateway/internal/llm/` + plugin set; code lineage = **copy under Apache 2.0 with attribution** (NOT go-modules-consume from ai-platform) | OQ-1, OQ-2 |
| PI2-yaa-0003 | **Apache 2.0 license flip**; supersedes PI1-yaa-0004; v0.1.x non-retroactive boundary clause; legal-review-pending banner travels with WI until counsel sign-off (PC-6 non-engineering checklist item) | OQ-5 (legal routing) |
| PI2-yaa-0005 | JWKS sourcing (OQ-3): `portfolio/packages/go/auth-jwks/` does **NOT** exist as of A-3 (verified — `portfolio/packages/go/` contains only `README.md`; LIBRARIES.md row re-targeted PI14-oppor; no PI14-oppor docs landed yet). token-validator plugin **re-implements minimally inline**. Stub behaviour (OQ-7): prompt-sanitize when `enabled: true` logs warn-once + passes through (does NOT exit 1) | OQ-3, OQ-7 |

> ADR PI2-yaa-0004 number is intentionally skipped to honour PRD §11 citation
> table verbatim (0001/0002/0003/0005). No engineering decision is hidden in
> the gap; ADRs are referenced by ID, not contiguity.

OQ-4 (LLM example port — 8121 vs 8122) is a platform-engineer A-4 decision
via compose-linter (advisory in `examples-llm-gateway.md`). OQ-5 legal
sign-off routes to chief-architect/user at PC-6 (not an engineering gate).

## Sprint plan (5 sprints; contract first, S5 reserved for e2e+conformance+re-publish)

| Sprint | Theme | WIs | Component files |
|--------|-------|-----|-----------------|
| **S1** | License flip + plugin contract scaffolding | LIC-1, LIC-2, PLG-1, PLG-2, REL-S | `license-and-headers.md`, `plugins.md`, `gateway.md`, `release-and-publish.md` |
| **S2** | Plugin chain + 5 first-party plugins | PLG-3, PLG-4, PLG-5, PLG-6, PLG-7 | `plugins.md`, `gateway.md` |
| **S3** | Go client SDK | GOC-1, GOC-2, GOC-3, GOC-4 | `client-go.md` |
| **S4** | LLM gateway specialisation + reference example | LLM-1, LLM-2, LLM-3, LLM-4, EX-LLM-1, EX-LLM-2 | `gateway.md`, `examples-llm-gateway.md` |
| **S5** | Version bumps + Compose e2e + re-publish | BUMP-1, BUMP-2, BUMP-3, EX-LLM-3, CLI-CONF, REL-1, REL-2, REL-3, REL-4, PI-GATE | `sdk-fastapi.md`, `client-python.md`, `client-ts.md`, `cli.md`, `examples-llm-gateway.md`, `release-and-publish.md` |

S5 honours runbook rule 3 (Compose end-to-end + CLI conformance gate +
publish reserved for last sprint).

## WI index (~30 WIs across 11 component files)

- **License & headers** (`license-and-headers.md`): LIC-1 `LICENSE`→Apache 2.0 + `COMMERCIAL.md` retirement, LIC-2 SPDX header sweep (single commit across all source dirs)
- **Plugins** (`plugins.md`): PLG-1 plugin interface + registry, PLG-3 token-validator (always-on), PLG-4 tenant-injector, PLG-5 license-check, PLG-7 stubs (prompt-sanitize + otel-audit)
- **Gateway core** (`gateway.md`): PLG-2 gateway plugin-loader refactor, PLG-6 plugin chain handler + per-route overrides + reverse-Shutdown, LLM-1 SSE proxy, LLM-2 per-tenant SSE concurrency limiter, LLM-3 execution-timeout + CORS, LLM-4 SSE Prometheus metrics, BUMP-3 profile v0.2 response header + spec/schema bumps
- **Go client** (`client-go.md`): GOC-1 client+options, GOC-2 resource accessors, GOC-3 AgenticResult + typed errors, GOC-4 golden corpus tests
- **FastAPI SDK** (`sdk-fastapi.md`): BUMP-1a SPDX + license metadata + profile-v0.2 (slice of BUMP-1)
- **Python client** (`client-python.md`): BUMP-1b SPDX + license metadata + profile-v0.2 (slice of BUMP-1)
- **TS client** (`client-ts.md`): BUMP-2 `"license": "Apache-2.0"` + SPDX + profile-v0.2
- **CLI** (`cli.md`): BUMP-1c SPDX + license metadata, CLI-CONF conformance-test v0.2 (plugin chain + `X-YAAgents-Profile: v0.2` header + multi-plugin-config matrix)
- **LLM example** (`examples-llm-gateway.md`): EX-LLM-1 mock LLM backend + routes.yaml + plugins.yaml, EX-LLM-2 docker-compose.yml, EX-LLM-3 Compose e2e (all 5 §13.2 flows)
- **Release/publish** (`release-and-publish.md`): REL-S CONTRIBUTING.md + README license-badge swap, REL-1 PyPI×3 re-publish 0.2.0, REL-2 npm re-publish 0.2.0, REL-3 GHCR gateway 0.2.0, REL-4 Go modules tag `client-go/v0.2.0` + proxy.golang.org verification, PI-GATE end-of-PI gate (both Compose demos green + 5 installable artefacts + 0 Community SPDX in v0.2.0)

## Critical path & dependency spine

```
LIC-1/2 ──► PLG-1 ──► PLG-2 ──► PLG-3 ─┐
                                  PLG-4 ┼─► PLG-6 ─► (gateway core ready)
                                  PLG-5 ┘                  │
                                  PLG-7 (stubs)            │
                                                            ├─► LLM-1..4 (S4) ─► EX-LLM-1 ─► EX-LLM-2 ─► EX-LLM-3 ─► PI-GATE
                                                            │
                                                  GOC-1 ─► GOC-2 ─► GOC-3 ─► GOC-4 ┤
                                                                                   │
                                            BUMP-1a/b/c, BUMP-2, BUMP-3 ───────────┤
                                                            │                      │
                                                  CLI-CONF ─┘                      │
                                                                                   ▼
                                                                        REL-1/2/3/4 (publish, S5)
```

LIC-1 (Apache 2.0 `LICENSE` file) is the **non-negotiable first WI** — every
other component's SPDX header + package metadata cites it. LIC-2 (SPDX sweep)
is the single commit that touches all source dirs — explicitly scheduled S1
to avoid partial-sweep states in git history (PRD §8.4).

## Library gates (library-gates.md Gate 3)

- **Plugin WIs** (PLG-1, PLG-3..7) — `library_justify: novel plugin-middleware abstraction; no portfolio shared library applies (the plugin interface IS the abstraction being authored here).`
- **token-validator (PLG-3)** — `library_justify: portfolio/packages/go/auth-jwks/ extraction NOT yet landed (verified at A-3: portfolio/packages/go/ contains only README.md; LIBRARIES.md row re-targeted PI14-oppor; oppor/docs/PI14-oppor/ absent). Per ADR PI2-yaa-0005, re-implement minimally inline. Plugin Init signature stable so import-path can switch in a future PI when extraction lands.`
- **Gateway core / LLM-* / EX-LLM-*** — `library_ref: ADR PI2-yaa-0002` (ai-gateway absorption shape + code lineage) plus carry-forward `library_ref: ADR PI1-yaa-0001` (net-new base) where applicable
- **License + BUMP-* + REL-S** — `library_ref: ADR PI2-yaa-0003` (Apache 2.0 supersession)
- **REL-1..4** — `library_ref: ADR PI1-yaa-0005` (OIDC trusted publishing carries forward unchanged) + `library_ref: ADR PI2-yaa-0003` (license metadata field updates)
- **GOC-1..4** — `library_justify: novel Go client; idiomatic analog to client-python/client-ts; stdlib net/http only (zero non-stdlib runtime deps per PRD §5.9 design constraints).`
- **Gate 4 (duplication)**: GOC-* vs PYC-* vs TSC-* are intentional dual-language clients of one contract; override noted in each component file (the heuristic does not fire on cross-language pairs).
- **Gate 1 (A-1 librarian consultation)**: per planning-runbook, librarian was skipped with rationale (one open extraction row — oppor JWKS triplet — referenced via ADR PI2-yaa-0005). Architect confirms at A-3: re-implement inline.

## Out of scope (revised 2026-05-30T19:30Z — user-direct MOVE + decommission shift)

**Still out of scope (carry-forward from PRD §12):**

- Kubernetes manifests, Helm chart, GHCR OCI Helm publish → **PI3-yaa**
- Cosign image signing + SBOM attestation hardening → **PI3-yaa**
- prompt-sanitize + otel-audit reference implementations → **PI3-yaa or community**
- Plugin marketplace UI / discovery service → not planned (registry is API-only)
- Retroactive re-licensing of v0.1.x packages → out of scope (v0.1.x stays Community License)
- Frontend / UI of any kind → no UI surface in PI2-yaa (consistent with PI1-yaa zero-UI evidence)
- v0.3+ adapters (Spring Boot, ASP.NET, LangGraph, SK) → future PIs or community
- Async-operation profile + approval-flow runtime → PI3-yaa or v0.3+
- GTM content (demo videos, launch blog, social posts) → founder-owned

**Newly out of scope (added 2026-05-30T19:30Z):**

- Consumer migration off `ai-platform/ai-gateway` → **trivially N/A** (no production consumers exist per user direct); no migration plan, no compat shim, no parallel-run window.
- Backward-compat preservation of `ai-platform/ai-gateway`'s old public surface (route paths, config field names, env-var names, Keycloak-realm-aware JWKS URL construction, metric prefixes) → **NOT carried forward**. yaagents-canonical names are authoritative.
- Migration of any **other** ai-platform service (knowledge-api, agent-api, tooling-api, evaluation-service, agent-runtime, knowledge-worker) into yaagents → **NOT in PI2-yaa scope**. Only the `ai-gateway` service decommissions.
- `ai-platform` repo full-product-wide re-org → not in scope (only the one service is touched).

**Newly IN scope (was OUT previously; flipped 2026-05-30T19:30Z — ADR PI2-yaa-0002 amendment):**

- `ai-platform/services/ai-gateway/` deletion → IN scope.
- `ai-platform/docker-compose.yml` ai-gateway entry removal → IN scope.
- `ai-platform/.github/workflows/ai-gateway-deploy.yml` deletion → IN scope.
- `portfolio/infrastructure/` ai-gateway ECS module + ALB rule + ECR repo + SSM parameters + IAM roles → IN scope.

The four "Newly IN scope" items above are NOT enumerated as yaagents WIs
(out of writable lane — ai-platform-architect + platform-engineer-portfolio
own those paths). They are surfaced to chief-architect at the A-3 handoff
below for cross-lane scheduling. yaagents-side absorption (LLM-1..4 +
EX-LLM-1..3) is fully self-contained and platform-engineer A-4 NFR pass
proceeds in parallel.

## Success criteria (PRD §2, §6.6, §7.2, §8 — verified by PI-GATE)

PI2-yaa is DONE when:

1. `LICENSE` is Apache 2.0 verbatim; `COMMERCIAL.md` retired; `grep -rn "Community License" --include='*.go' --include='*.py' --include='*.ts'` returns zero hits in v0.2.0 artefacts (LIC-1, LIC-2).
2. Plugin chain executes in declaration order; reverse-order Shutdown on SIGTERM; gateway exit 1 if `token-validator.enabled: false` (PLG-6).
3. Gateway runs with a single-plugin YAML (token-validator only); request flows end-to-end (PRD §13.2 Flow 1).
4. Gateway runs with token + tenant + license + a community-authored example plugin loaded via the registry; request flows end-to-end (PRD §13.2 Flow 5 / PRD §6.6).
5. `pip install yaagents-fastapi==0.2.0|yaagents-client==0.2.0|yaagents-cli==0.2.0` from public PyPI succeeds; package metadata reports `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.2` (REL-1).
6. `npm install @aimpathyminds/yaagents-client@0.2.0` from public npm succeeds with provenance (REL-2).
7. `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` succeeds, both architectures; image label `org.opencontainers.image.licenses=Apache-2.0`; SBOM attached (REL-3).
8. `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` succeeds via `proxy.golang.org` within 30 min of tag push (REL-4).
9. `examples/llm-gateway/docker compose up` runs all 5 PRD §13.2 flows green (EX-LLM-3).
10. `examples/campaign-api/docker compose up` still runs PI1-yaa flows green at v0.2.0 package versions (CLI-CONF / PI-GATE regression check).

## Handoff

```
next:        chief-architect
artifact:    yaagents/docs/PI2-yaa/roadmap.md + 10 component files + docs/adr/PI2-yaa-0001/0002 (AMENDED)/0003/0005.md
intent:      Cross-lane direction needed for ai-platform/ai-gateway decommission (user-direct scope shift 2026-05-30T19:30Z — ADR PI2-yaa-0002 amended MOVE not COPY). The yaagents-side absorption (LLM-1..4 + EX-LLM-1..3) is fully self-contained in this roadmap. The cross-lane cleanup spans three writable paths outside yaagents-architect's lane: (a) ai-platform-architect → delete ai-platform/services/ai-gateway/ + remove compose entry + delete .github/workflows/ai-gateway-deploy.yml; (b) platform-engineer-portfolio → portfolio/infrastructure/ TF deletion (ai-gateway ECS module + ALB target group + listener rule + ECR repo + SSM parameters + IAM role scopes); (c) product-manager (yaagents) → PRD §12 row "ai-platform/ai-gateway ECS service changes → service runs unchanged this PI" amendment to reflect the new posture. Suggested cross-lane ordering: yaagents-side absorption (LLM-1..4) lands first under platform-engineer A-4 → A-5 dispatch; ai-platform deletion + TF cleanup can land any time after but before PI2-yaa PC-6 close gate. Parallel: platform-engineer (yaagents lane) A-4 NFR / supply-chain pass on this roadmap proceeds independently of the cross-lane work.
> blocker:   ai-platform/services/ai-gateway/ + portfolio/infrastructure/ ai-gateway TF + PRD §12 row are outside yaagents-architect writable lane. Cannot land MOVE semantics from this lane alone — yaagents only absorbs the code; the ai-platform-side `git rm` and the TF deletion require a chief-architect cross-lane direction. Verified no production consumers per user-direct 2026-05-30T19:30Z (no consumer-migration WI required).
cwd:         yaagents/

(After chief-architect dispatches the cross-lane direction, the parallel yaagents-side handoff is:
 next:     platform-engineer (yaagents lane)
 intent:   A-4 NFR / supply-chain pass on yaagents/docs/PI2-yaa/ — append [SEC]/[SRE]/[SUPPLY] WI bodies per PRD §10 seeds; run compose-linter on EX-LLM-2 + EX-LLM-3; lock OQ-4 port = 8122; confirm Syft SPDX 2.3 JSON SBOM carry-forward from PI1-yaa REL-5; explicitly state [FIN] N/A; flip all yaagents-side PI2-yaa WIs [DRAFT]→[READY].
 cwd:      yaagents/)
```
