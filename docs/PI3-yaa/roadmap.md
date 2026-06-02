# PI3-yaa — YAAgents v0.3 Public Launch + sdk-go + Repo Restructure + Pages — Roadmap (master)

Status: [READY] (A-4 NFR pass complete 2026-06-02; all WIs [READY]; 5 NFR WIs + bin/yaagents-public-mirror-verify.sh authored)
PI: PI3-yaa · Lane: yaa (parallel to Lane A oppor + Lane B plt-aip)
Author: yaagents-architect · Date: 2026-06-02
PRD: `yaagents/system-refs/yaagents-v0.3_detailed.md` (§ refs throughout)
Onepager: `yaagents/system-refs/yaagents-v0.3_onepager.md`
Seed: `yaagents/system-refs/yaagents-v0.3.seed.md` [READY]
Intake: `portfolio/INTAKE/PI3-yaa-intake.md` (7 LOCKED architecture decisions A-1..A-7)
Planning runbook: `portfolio/RUNBOOKS/pi3-yaa-planning.yml`
Phase B gate: PI2-yaa PC-6 closed 2026-06-02.

## Scope

PI3-yaa is the **public launch PI**. Five tightly coupled tracks ship together as v0.3.0:

1. **TRACK SDK-GO** — Go server SDK (`sdk-go/`, NEW). Idiomatic analog to `sdk-fastapi`; vendor types
   generated from `schemas/v0.3/*.json` (one canonical source); router-agnostic (`net/http` core +
   chi/gin/echo adapters); zero non-stdlib runtime deps; `ProfileVersion = "v0.3"`; reference example
   `examples/campaign-api-go/`; published as Go module `github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0`.
   Cross-lane stretch (`A-3b`): ai-platform/agent-api canary — 1 endpoint end-to-end (PRD §11 OQ-6).
2. **TRACK REPO** — Split the monorepo into meta-repo + 7 public submodule repos at
   `github.com/ai-mpathyminds/yaagents-*` per intake §A-1; `spec/schemas/openapi/examples/docs` stay
   at meta-repo root per §A-2; Go module path migration per §A-3.
3. **TRACK SCRUB** — Move internal planning artifacts (`docs/PI*-yaa/`, internal ADRs,
   `system-refs/*.seed.md`, `*_detailed.md`, `*_onepager.md`, `.claude/`, `CLAUDE.md`) into
   `portfolio/yaagents-internal/` per intake §A-4; orphan-baseline squashed-history commit on each
   submodule repo.
4. **TRACK PAGES** — Astro Starlight site at `docs/` (meta-repo root); hosted on **GitHub Pages**
   at `https://ai-mpathyminds.github.io/yaagents/` per intake §A-5 (revised 2026-06-02 from
   Cloudflare Pages; custom domain deferred post-traction). 10 content sections.
5. **TRACK LAUNCH** — Prod publish: PyPI ×3 (`yaagents-fastapi`, `yaagents-client`, `yaagents-cli`)
   + npm ×1 (`@aimpathyminds/yaagents-client`) + GHCR (`yaagents-gateway:0.3.0` multi-arch + Cosign
   + Syft SBOM) + Go modules ×2 (`yaagents-sdk-go@v0.3.0`, `yaagents-client-go@v0.3.0`); flip 8
   repos PRIVATE→PUBLIC; push Pages site live.

**Cross-track dependency order (locked, PRD §2):** SCRUB before any repo goes PUBLIC; REPO before
LAUNCH; SDK-GO tag `v0.3.0` before LAUNCH publishes; PAGES deploys before LAUNCH flips.

**Out of scope (PRD §13):** GitHub org migration to `yaagents/` standalone (PI4-yaa+ traction-gated);
custom domain `yaagents.dev` (1-WI post-traction); v0.4 plugin scope (license-check multi-backend,
cookie JWT transport, JTI revocation, OAuth 2.0 introspection); K8s/Helm; LangGraph/SK adapters;
prompt-sanitize/otel-audit full implementations; ai-platform/agent-api **full** migration (canary
only).

## ADR slate (all [Accepted] at A-3 close; `docs/adr/`)

| ADR | Decision | Resolves |
|-----|----------|----------|
| PI3-yaa-0001 | Meta-repo + 7 public submodule shape; `spec/schemas/openapi/examples/docs` stay at meta-repo root (NOT separate submodules). Profile changes propagate via meta-repo coordination PRs. | Intake §A-1 + §A-2 |
| PI3-yaa-0002 | Go module path migration: `github.com/ai-mpathyminds/yaagents/{client-go,gateway}` → `github.com/ai-mpathyminds/yaagents-{client-go,gateway}`; new `github.com/ai-mpathyminds/yaagents-sdk-go`; v0.1.0/v0.2.0 monorepo-path tags retired (never published publicly). | Intake §A-3 |
| PI3-yaa-0003 | Portfolio scrub mechanism: per-submodule `git checkout --orphan public-main` squashed-history baseline commit; internal artifacts move out to `portfolio/yaagents-internal/`; OSS users get code+license+contributing, not internal PI narrative; provenance preserved via baseline commit message `<sha>` reference. | Intake §A-4 |
| PI3-yaa-0004 | Pages site on **GitHub Pages** at `ai-mpathyminds.github.io/yaagents/` (project-pages mode; Astro `site: 'https://ai-mpathyminds.github.io'`, `base: '/yaagents'`); GitHub-managed `*.github.io` TLS; zero analytics; deploy via `actions/deploy-pages@v4`. Revised 2026-06-02 from Cloudflare Pages — "use what GitHub gives us until we need more". | Intake §A-5 |
| PI3-yaa-0005 | GitHub org branding: stay at `ai-mpathyminds/yaagents-*` for v0.3 launch. Standalone `yaagents/` GitHub org defers to v0.4+ traction gate; mid-launch org migration creates path-rewrite cascade (Go module paths, workflow files, Trusted Publisher configs, doc links). | Intake §A-6 |

> Intake §A-7 (PI2-yaa REL-* deferral to PI3-yaa) and §A-8 (B-01 user-direct publish-prep first
> Phase B entry) are not ADRs — they are scope amendments + runbook directives, captured in the
> planning runbook `context.first_b_entry_intent` field and the PRD §10.6 acceptance criteria.

## Sprint plan (7 sprints; contract+scaffold first, S7 reserved for public-flip + Pages-live + PI-GATE)

| Sprint | Theme | WIs | Component files |
|--------|-------|-----|-----------------|
| **S1** | Profile v0.3 bump + sdk-go scaffold + scrub move | SP-1, SG-1, SG-2, SC-1, SC-2 | `contracts.md`, `sdk-go.md`, `scrub.md` |
| **S2** | sdk-go core: response factory + adapters + tests | SG-3, SG-4, SG-5 | `sdk-go.md` |
| **S3** | sdk-go example + meta-repo skeleton | SG-6, RP-META, RP-SUBMOD | `sdk-go.md`, `repo.md` |
| **S4** | 7 submodule inits + cross-reference sweep | RP-GATEWAY-INIT, RP-SDKFASTAPI-INIT, RP-SDKGO-INIT, RP-CLIPY-INIT, RP-CLITS-INIT, RP-CLIGO-INIT, RP-CLI-INIT, RP-XREF | `repo.md` |
| **S5** | Pages site content + build workflow | PG-1, PG-2, PG-3, PG-4, PG-5, PG-6, PG-7, PG-8 | `pages.md` |
| **S6** | Publish wave (PyPI×3 + npm + GHCR + Go×2) | SG-7, LA-PYPI-FASTAPI, LA-PYPI-CLIPY, LA-PYPI-CLI, LA-NPM, LA-GHCR, LA-GO-CLIENT | `sdk-go.md`, `launch.md` |
| **S7** | Public-flip + Pages-live + PI-GATE | LA-PUBLIC-FLIP, LA-PAGES-DEPLOY, LA-PI-GATE | `launch.md` |

S1 honours runbook rule 2 (contract-first: profile v0.3 bump on `spec/schemas/openapi` is the
non-negotiable first WI; every component cites the bumped contract).
S7 honours runbook rule 3 (Compose end-to-end + CLI conformance gate + publish reserved for last
sprint; PI-GATE is the prod-install regression check).

## WI index (~33 WIs across 6 component files)

- **Contracts** (`contracts.md`): SP-1 Profile v0.3 spec + schemas + OpenAPI bump
- **sdk-go** (`sdk-go.md`): SG-1 module scaffold + AgenticContext + FromRequest, SG-2 vendor codegen from schemas, SG-3 AgenticResponse factory + Write + AgenticWritable, SG-4 chi/gin/echo adapters, SG-5 unit tests + ≥80% coverage, SG-6 `examples/campaign-api-go/` reference, SG-7 Go module tag `v0.3.0` + proxy.golang.org verification (LAUNCH coupling)
- **Repo restructure** (`repo.md`): RP-META meta-repo public skeleton + community health files, RP-SUBMOD `.gitmodules` pointers, RP-GATEWAY-INIT, RP-SDKFASTAPI-INIT, RP-SDKGO-INIT, RP-CLIPY-INIT, RP-CLITS-INIT, RP-CLIGO-INIT, RP-CLI-INIT (×7 orphan-baseline + module-path/package-metadata migration), RP-XREF cross-reference sweep
- **Scrub** (`scrub.md`): SC-1 internal-artifacts move into `portfolio/yaagents-internal/`, SC-2 `.claude/` + `CLAUDE.md` removal from working tree
- **Pages site** (`pages.md`): PG-1 Astro Starlight scaffold + Astro config, PG-2 Hero + Why yaagents copy, PG-3 Quick Start (10-min walkthrough), PG-4 Profile Spec MDX render, PG-5 6-target SDK Quickstarts, PG-6 Examples walkthroughs, PG-7 Plugin authoring + Public Roadmap + Contributing + Community, PG-8 `.github/workflows/pages.yml` build+deploy workflow stub
- **Launch** (`launch.md`): LA-PYPI-FASTAPI, LA-PYPI-CLIPY, LA-PYPI-CLI (×3 PyPI Trusted Publisher tag-driven), LA-NPM (`@aimpathyminds/yaagents-client` provenance), LA-GHCR (`yaagents-gateway:0.3.0` multi-arch + Cosign + SBOM via Syft), LA-GO-CLIENT (`yaagents-client-go@v0.3.0` tag + proxy index), LA-PUBLIC-FLIP (8 repos PRIVATE→PUBLIC), LA-PAGES-DEPLOY (initial Pages site publish), LA-PI-GATE (acceptance gate: 5 install validations + Cosign verify + npm provenance + scrub verify zero hits)

> SG-7 (sdk-go module tag) lives in `sdk-go.md` rather than `launch.md` because the developer lane
> owns the tag-push action (the codegen + tests must be green at tag time). LA-* WIs in
> `launch.md` are operator-/CI-driven mechanical entries dispatched only after B-01 publish-prep
> precheck returns PASS (PRD §10.6, intake §A-8). B-01 itself is NOT a roadmap WI — it is a
> mechanical first-Phase-B-entry authored by execution-runbook-generator at A-6 per the planning
> runbook `context.first_b_entry_intent` field.

## Critical path & dependency spine

```
SP-1 (profile v0.3) ──► SG-1 ──► SG-2 ──► SG-3 ──► SG-4 ──► SG-5 ──► SG-6 ─┐
                                                                            │
SC-1, SC-2 (scrub) ──┐                                                       │
                      ├─► RP-META ──► RP-SUBMOD ──► RP-{GATEWAY,SDKFASTAPI,SDKGO,CLIPY,CLITS,CLIGO,CLI}-INIT
                      │                                                  │
                      │                                                  ▼
                      │                                                RP-XREF (cross-ref sweep)
                      │                                                  │
                      └─► [SCRUB+REPO must be clean before any LAUNCH] ──┤
                                                                          │
PG-1 ──► PG-2..PG-7 (content) ──► PG-8 (workflow) ──────────────────────┤
                                                                          │
                                                                          ▼
                                                      [B-01 PRECHECK PASS gate — operator-driven]
                                                                          │
                                                                          ▼
                                                       SG-7 ─► LA-PYPI×3, LA-NPM, LA-GHCR, LA-GO-CLIENT
                                                                          │
                                                                          ▼
                                                            LA-PUBLIC-FLIP ─► LA-PAGES-DEPLOY ─► LA-PI-GATE
```

SP-1 (profile v0.3 spec/schema/openapi bump) is the **non-negotiable first WI** — `sdk-go`
vendor codegen depends on `schemas/v0.3/*.json` existing; every component's `Supports-YAAgents-Profile`
metadata cites it.

## Library gates (`.claude/rules/library-gates.md`)

- **Gate 1 (A-1 platform-librarian dispatch)**: SKIPPED at A-1 with rationale per planning runbook
  `context.library_gates.gate_1_a1_librarian_dispatch` — scope is dominated by repo restructure
  (mechanical), Pages site (static docs framework), sdk-go (new abstraction with no portfolio
  analog). Architect re-confirms at A-3: no scope expansion encountered that would consume a
  portfolio-shared library.
- **Gate 2 (`LIBRARIES.md` SLA)**: trivially satisfied — 0 rows in
  `portfolio/LIBRARIES.md §Extraction proposals (open)` target PI3-yaa (verified at A-1 via grep
  per planning runbook).
- **Gate 3 (per-WI `library_ref` / `library_justify`)**: every WI carries one of the two fields.
  - **sdk-go WIs (SG-*)**: `library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10 design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).`
  - **Repo restructure WIs (RP-*)**: `library_ref: ADR PI3-yaa-0001` (meta-repo + 7 submodule shape) + `library_ref: ADR PI3-yaa-0002` (Go module path migration) + `library_ref: ADR PI3-yaa-0003` (orphan-baseline squashed history) for INIT WIs.
  - **Scrub WIs (SC-*)**: `library_ref: ADR PI3-yaa-0003` (squashed-history orphan-baseline trade-off).
  - **Pages WIs (PG-*)**: `library_ref: ADR PI3-yaa-0004` (GitHub Pages + Astro Starlight choice) + `library_justify: Astro Starlight is the canonical Astro docs framework; default theme = no custom component design; no portfolio shared library applies (Pages site is yaagents-public-only).`
  - **Launch WIs (LA-*)**: `library_ref: ADR PI1-yaa-0005` (OIDC trusted publishing carries forward unchanged) + `library_ref: ADR PI2-yaa-0003` (Apache 2.0 metadata) + `library_ref: ADR PI3-yaa-0002` (new Go module paths).
  - **SP-1**: `library_justify: profile-version bump only; spec/schemas/openapi are the contract canonical source — they DEFINE the abstraction other libraries depend on.`
- **Gate 4 (duplication detection)**:
  - **TRACK REPO 7-way INIT WIs (RP-GATEWAY-INIT..RP-CLI-INIT)**: Gate 4 **FIRES** — 7 WIs differ only in their per-submodule component target. **ARCHITECT OVERRIDE per architect judgment + intake §A-1 + PRD §6.2**: each submodule carries distinct per-language package metadata (Go `go.mod` module path; Python `pyproject.toml` package name + classifiers + Hatch build target; TypeScript `package.json` name + scope + ESM/CJS build config; Docker image metadata) and distinct per-component cross-reference updates. Extracting a shared "orphan-baseline-and-migrate" abstraction would couple 4 different language toolchains' package-metadata writers into one shared module — net complexity ↑, not ↓. Rule-of-three threshold IS exceeded (7 consumers > 3), BUT the per-consumer differentiation (4 different language toolchains + per-component metadata) eliminates the extraction value. `duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.` See each `RP-*-INIT` WI body for the same override (carried verbatim per WI per `.claude/rules/library-gates.md §Gate 4`).
  - **PG-* content WIs**: no duplication; each section is distinct content.
  - **LA-PYPI×3**: 3 publish WIs differ only in package name. Architect override: each WI is a 3-line GitHub Actions tag-push that delegates to the PyPI Trusted Publisher OIDC flow — already abstracted by the registry-side trust config. No code shared across WIs to extract. `duplication_override: each LA-PYPI-* WI is a tag-push trigger only; OIDC Trusted Publisher OWNS the publish action; no shared portfolio code to extract.`

## Cross-lane stretch (A-3b — `ai-platform-architect`)

PRD §5.10 + intake §A-3 + planning runbook A-3b: ai-platform-architect adds 1–3 WIs in
`ai-platform/docs/PI3-yaa/agent-api-canary.md` for ai-platform/agent-api adoption of
`yaagents-sdk-go` on **one** resource endpoint (PRD §11 OQ-6 resolution).

**Cross-lane edge**: sdk-go acceptance criteria SG-5 (≥80% coverage on `sdkgo/` core; all 10
response-type Status()+ContentType() unit-tested) MUST be green before agent-api canary WIs
dispatch. Sequencing: yaagents-side SG-1..SG-5 land in S2 → ai-platform-side canary WIs land in
S3 or later (slack present). ai-platform-architect's WI bodies cross-reference the SDK API
surface (PRD §5.10.1) and the sdk-go module tag (`v0.3.0`); they MUST tag `paired_yaa_wi:` so
the execution runbook sequences correctly.

## Out of scope (PRD §13)

- **GitHub org migration** to standalone `yaagents/` GitHub org → PI4-yaa+ (traction-gated; ADR PI3-yaa-0005)
- **Custom domain** `yaagents.dev` → 1-WI post-traction (intake §A-5 §Migration)
- **K8s manifests / Helm chart** → re-evaluated for PI4-yaa (was PI2-yaa deferred; community demand will inform priority)
- **v0.4 plugin scope** — license-check multi-backend, cookie JWT transport, JTI revocation, OAuth 2.0 introspection → PI4-yaa per ADR PI2-yaa-0007 deferred-items table
- **prompt-sanitize / otel-audit full implementations** → PI4-yaa or community
- **LangGraph / Semantic Kernel / Spring Boot adapters** → v0.4 or community
- **Async-operation profile + approval-flow runtime** → PI4-yaa candidate
- **Discord / Slack community channel setup** → separate community initiative; Pages site links to whatever exists at LAUNCH time (default: GitHub Discussions on meta-repo)
- **ai-platform/agent-api full migration** off custom AI gateway → plt-aip lane; NOT yaa
- **Retroactive re-licensing of v0.1.x packages** → not in scope (v0.1.x stays Community License; v0.2.x ships Apache 2.0)
- **Production analytics on yaagents.dev** → privacy-first; deferred indefinitely

## Success criteria (PRD §1 Goals + PRD §10.6 + verified by LA-PI-GATE)

PI3-yaa is DONE when ALL of the following hold:

1. **Source repo public**: `github.com/ai-mpathyminds/yaagents` meta-repo + 7 submodule repos are PUBLIC on GitHub (`gh repo view --json visibility` returns `PUBLIC` on all 8).
2. **PyPI ×3 installable from prod**: `pip install yaagents-fastapi==0.3.0 yaagents-client==0.3.0 yaagents-cli==0.3.0` from public PyPI succeeds in a fresh venv; each package metadata reports `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3` (LA-PYPI-*).
3. **npm installable from prod**: `npm install @aimpathyminds/yaagents-client@0.3.0` from public npm succeeds with provenance attestation (LA-NPM).
4. **GHCR installable from prod**: `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` succeeds on both `linux/amd64` + `linux/arm64`; OCI label `org.opencontainers.image.licenses=Apache-2.0` present; Cosign signature verifies (`cosign verify ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0`); Syft SBOM attached (LA-GHCR).
5. **Go modules installable from prod**: `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0` and `go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` succeed via `proxy.golang.org` within 30 min of tag push (SG-7 + LA-GO-CLIENT).
6. **Pages site live**: `https://ai-mpathyminds.github.io/yaagents/` serves the MVP site (10 sections per PRD §5.11) with Lighthouse ≥90 Perf + A11y (platform-engineer A-4 NFR gate enforces).
7. **ai-platform/agent-api canary green**: ≥1 ai-platform/agent-api resource endpoint runs through `sdk-go` with `AgenticResult` response shape verified end-to-end (cross-lane A-3b WI close).
8. **Portfolio scrub clean**: `bin/yaagents-public-mirror-verify.sh` returns 0 hits on all 8 public repos (`.claude/`, `portfolio/`, `PI[0-9]*-yaa`, `system-refs/`, `CLAUDE.md` markers absent; CHANGELOG references to past versions excluded as legitimate).
9. **Apache 2.0 cross-component clean**: `grep -rn "Community License" --include='*.go' --include='*.py' --include='*.ts'` returns 0 hits across all 8 public repos at v0.3.0; SPDX `Apache-2.0` headers present on every source file.
10. **examples/campaign-api-go Compose green**: `cd examples/campaign-api-go && docker compose up` runs all 5 PRD §13.2 flows green; smoke curl returns `201 application/json` for happy path; `400 application/vnd.yaagents.clarification+json` for clarification flow (SG-6 + CI gate).

## Handoff

```
next:        platform-engineer (yaagents lane — workspace-root /.claude/agents/platform-engineer.md)
artifact:    yaagents/docs/PI3-yaa/roadmap.md + 6 component files (contracts.md, sdk-go.md, repo.md, scrub.md, pages.md, launch.md) + 5 ADRs (docs/adr/PI3-yaa-0001..0005.md)
intent:      A-4 NFR / supply-chain pass on yaagents/docs/PI3-yaa/ per planning runbook A-4 intent. Author yaagents/docs/PI3-yaa/platform-engineer.md (NFR section per PRD §12 seeds: [SEC] govulncheck on sdk-go + meta-repo workflows + zero non-stdlib runtime deps verification on sdkgo/; [SUPPLY] per-submodule publish workflow scaffolds, Cosign sign on GHCR, Syft SBOM, OIDC-only across all 4 registries, license-clean scan 0 non-Apache-2.0 SPDX, GHCR multi-arch confirm at v0.3.0, Go proxy index verification within 30 min; [SRE] Pages Lighthouse ≥90 Perf+A11y CI gate, campaign-api-go Compose smoke in CI, sdk-go ≥80% coverage gate; [FIN] N/A explicit — no AWS substrate touched). Author bin/yaagents-public-mirror-verify.sh per planning runbook A-4 artifact (TRACK SCRUB acceptance script; grep audit for .claude/+portfolio/+PI*-yaa+system-refs+CLAUDE.md references). Flip all PI3-yaa WIs [DRAFT]→[READY]. Then handoff to scrum-master for A-5 mechanical pi-open (creates pi3-yaa branch + tag on yaagents and ai-platform repos).
cwd:         yaagents/
```

```
> note (cross-lane handoff to ai-platform-architect at A-3b — parallel to A-4):
  ai-platform-architect: read this roadmap §Cross-lane stretch + PRD §5.10 sdk-go API surface
  + PRD §11 OQ-6 (one endpoint end-to-end). Author ai-platform/docs/PI3-yaa/agent-api-canary.md
  (1-3 WIs) + ai-platform/docs/adr/PI3-yaa-0001-ai-platform-side.md (cross-lane mirror referencing
  yaagents-side PI3-yaa-0001 as primary binding). Each canary WI tagged with paired_yaa_wi:
  WI-3yaa.SG-{N} so execution runbook sequences correctly. cwd: ai-platform/
```
