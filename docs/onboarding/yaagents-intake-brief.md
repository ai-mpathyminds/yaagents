# Intake brief — onboarding `yaagents` (meta repo + 7 submodule repos)

**Artifact type:** `intake-brief` (`application/vnd.wakalix.intake-brief+json; v=1`, rendered here as a document)
**Produced by:** `onboarding-agent` / `confirm-grounding`
**Packet:** `01a0703b-c6ec-7501-ba5e-32677bd53b92`
**Skill run:** `01a0703b-91c0-76e3-a2dc-68ad903f63a4`
**Product:** `yaagents`
**Repository:** `github.com/ai-mpathyminds/yaagents`
**Branch:** `wakalix/run-01a0703b91c0` (base `main`), review mode `pr-human`
**Work item:** none — the packet carried `tail.work_item_id: null`
**Date:** 2026-09-05

This is an intake, not a plan and not a change. It records what somebody asked for, what the
workspace already answers, and what a person still has to settle before a grounding cycle can close.
No file outside this document was created, edited or deleted.

---

## 1. The ask, verbatim

From the packet, `tail.input.intent`:

> The root repo is a meta repo and includes multiple sub module repos , you have to onboard all as part of yaagent

`tail.acceptance_criteria` was empty. `tail.intake` was `null`. No prior conversation was supplied.

### How I read it

The ask has two clauses, and the second one settles the shape of the first:

1. **`yaagents` is not a single repository.** It is a meta repository that carries seven git
   submodules. Confirmed by inspection — see §3.
2. **"onboard all *as part of yaagent*"** — the seven submodules are to be onboarded as
   *components of the one product `yaagents`*, not as seven separate products.

Clause 2 is my reading of the operator's words, not a fact the workspace states. It is the reading
the phrase "as part of yaagent" most plainly supports, and it is consistent with the packet, which
names exactly one `delivery.repository`. It is recorded as **asked (operator intent), medium
confidence** in §4 and it is restated as open question **Q1** in §6, because the alternative reading
changes every downstream dimension.

---

## 2. What I read

Per `cite-what-you-read`, everything this brief rests on, named.

### wakalix record

| Source | Call | What it gave |
|---|---|---|
| Agent catalog | `catalog_read(role="onboarding-agent")` | Role, description, `produces: [intake-brief, grounding-manifest]`, skills `confirm-grounding` and `scan-workspace`. |
| Artifact registry | `registry_read()` | The 12 registered artifact types and what each is for. |
| Cycle artifacts | `artifact_read()` (no argument) | **`{"artifacts": []}`** — this organisation has produced nothing yet. |
| Packet | `.wakalix/inbox/01a0703b-…packet.json` | Rules, contract, delivery, `tail`. |

`prior.artifacts` and `prior.attempts` in the packet were both empty, and the cycle listing agrees:
**this is the first dispatch on this ask, and there was no earlier artifact to read instead of
re-deriving.**

### Workspace (paths, all read at `wakalix/run-01a0703b91c0`)

| Path | What I took from it |
|---|---|
| `.gitmodules` | The seven submodule paths and their upstream URLs. |
| `README.md` | Product statement, repository layout, response-profile table, published-artifact versions, install surfaces, licence summary. |
| `CONTRIBUTING.md` | Dev setup per component, branch naming, commit convention, DCO + agent trailers, code-style gates, release process and release-wave order, plugin contribution path. |
| `docs/README.md` | Docs site is Astro Starlight, `pnpm dev/build/preview`, published to GitHub Pages at `/yaagents/`. |
| `.github/workflows/ci.yml` | The full CI gate set, per-job working directories, `submodules: recursive` checkout, and the 2026-06-09 advisory degradations. |
| `.markdownlint.json` | `default: true` with `MD013` and `MD060` off. |
| `.gitignore` | Ignore set; nothing that would swallow this document. |
| `spec/VERSION` | `0.3`. |
| Directory listings | Repo root; `docs/` (depth 2); `examples/`; `spec/`, `schemas/`, `openapi/`, `tools/`, `tests/`, `bin/`, `docker/` (recursive); `.github/`. |
| `git status --short` | Empty — clean worktree. |

---

## 3. The product's shape, as inspected

`yaagents` is a gateway plus SDKs that put an agent implementation behind a governed REST API
(`README.md`). The normative contract is the **Agentic REST Response Profile v0.3**
(`spec/agentic-rest-profile.md`, `spec/VERSION` = `0.3`).

### Components — one meta repo and seven submodules

Paths and URLs are from `.gitmodules`; roles are from `README.md` §Repository layout and
`CONTRIBUTING.md` §Dev setup.

| # | Path | Upstream repository | Role | Language / toolchain |
|---:|---|---|---|---|
| 0 | *(root)* | `ai-mpathyminds/yaagents` | Meta repo: spec, schemas, OpenAPI, docs site, examples, CI | Markdown / JSON / YAML; Go tool in `tools/gen` |
| 1 | `gateway/` | `ai-mpathyminds/yaagents-gateway` | Go gateway + plugin chain | Go 1.22+ |
| 2 | `sdk-go/` | `ai-mpathyminds/yaagents-sdk-go` | Go server SDK | Go 1.22+ |
| 3 | `sdk-fastapi/` | `ai-mpathyminds/yaagents-sdk-fastapi` | Python FastAPI server SDK (`yaagents-fastapi`) | Python 3.11+, Hatch |
| 4 | `client-python/` | `ai-mpathyminds/yaagents-client-python` | Python client (`yaagents-client`) | Python 3.11+, Hatch |
| 5 | `client-ts/` | `ai-mpathyminds/yaagents-client-ts` | TypeScript client (`@aimpathyminds/yaagents-client`) | Node 20+/22, pnpm 9 |
| 6 | `client-go/` | `ai-mpathyminds/yaagents-client-go` | Go client SDK | Go 1.22+, zero non-stdlib runtime deps |
| 7 | `cli/` | `ai-mpathyminds/yaagents-cli` | CLI validator (`yaagents-cli`) | Python 3.11+, Hatch |

`CONTRIBUTING.md` §Release process independently corroborates the count: a major version bump
"must tag all **eight** components at the same version in the same release cycle" — meta repo plus
the seven above.

### Meta-repo-only content (not in any submodule)

- `spec/` — the normative profile, `VERSION`, and a validity corpus under `spec/examples/v0.1/`
  (valid and deliberately-invalid fixtures per response type), indexed by `spec/examples/INDEX.md`.
- `schemas/` — JSON Schema for six response types, versioned `v0.1`, `v0.2`, `v0.3`.
- `openapi/` — `yaagents-components.yaml`, `yaagents-response-profile.yaml`.
- `tests/golden/` — ten golden response bodies, one per profile response type.
- `docs/` — Astro Starlight site source (`src/content/` is the published collection).
- `examples/` — `agent-graph-ecom`, `campaign-api`, `campaign-api-go`, `customer-support-triage`,
  `financial-risk-screening`, `llm-gateway`, `store`.
- `tools/gen/`, `bin/yaagents-public-mirror-verify.sh`, `docker/gateway/Dockerfile`.
- `.github/` — 10 workflows and 4 issue templates.

### Submodules were **not** inspectable in this run

The seven submodule directories exist and are **empty**, and `git status --short` is clean —
i.e. the gitlinks are recorded but the submodules are not initialised in this worktree. Everything
this brief says about submodule *contents* therefore comes from the meta repo's own descriptions
(`README.md`, `CONTRIBUTING.md`, `.github/workflows/ci.yml`), **never** from reading the submodules.
I did not read a single file inside any submodule. I also could not read the recorded gitlink
commit pins: `git submodule status` and `git ls-tree HEAD` were both refused by this session's
command policy. See §6 **Q3**.

---

## 4. Dimensions — value, source, confidence

Per `a-source-per-dimension` and `settle-what-you-can-and-name-what-you-cannot`. Source is one of
**inspected** (a path in §2), **asked** (the operator's intent in §1), or **unresolved**.

> The dimension names below are my own working set for this intake. `registry_read` describes the
> `grounding-manifest` type but does not publish its dimension vocabulary or schema, so I could not
> align these names to a canonical list. If wakalix has one, this table should be re-keyed to it
> during `scan-workspace`. Flagged rather than guessed.

| Dimension | Settled value | Source | Confidence |
|---|---|---|---|
| Product identity | `yaagents` — gateway + SDKs exposing agents as governed REST APIs | inspected — `README.md` | High |
| Product home | Meta repo `github.com/ai-mpathyminds/yaagents` | inspected — packet `delivery`, `README.md` | High |
| Repository topology | 1 meta repo + 7 git submodules, 8 components total | inspected — `.gitmodules`, `README.md`, `CONTRIBUTING.md` | High |
| Onboarding scope | All 8 onboarded as components of the one product `yaagents` | asked — operator intent §1 | **Medium — see Q1** |
| Normative contract | Agentic REST Response Profile v0.3; 10 response types with fixed status + media type | inspected — `README.md`, `spec/VERSION` | High |
| Schema surface | `schemas/v0.1`, `v0.2`, `v0.3`; 6 types each | inspected — directory listing | High |
| Languages | Go 1.22+, Python 3.11+ (CI on 3.12), TypeScript on Node 22 | inspected — `CONTRIBUTING.md`, `ci.yml` | High |
| Build / test commands | Per-component, listed in `CONTRIBUTING.md` §Dev setup | inspected — `CONTRIBUTING.md` | High |
| Lint / type gates | Go `golangci-lint`; Python `ruff` + `mypy --strict`; TS `eslint` + `tsc --noEmit`; docs `pnpm build` | inspected — `CONTRIBUTING.md` §Code style, `ci.yml` | High |
| Coverage floor | ≥80% for Go, Python (`--cov-fail-under=80`) and TS (vitest thresholds) | inspected — `ci.yml` | High |
| Security gates | `govulncheck`, `pip-audit`, `pnpm audit --audit-level=high`, no-`math/rand`, no-dynamic-plugin-load, no-secret-in-Dockerfile, no-`.env`-in-repo, license-clean-scan | inspected — `ci.yml` | High |
| CI checkout mode | Every job checks out `submodules: recursive`; jobs run with `working-directory` set to a submodule | inspected — `ci.yml` | High |
| Branch naming (human) | `feature/…`, `fix/…`, `docs/…` | inspected — `CONTRIBUTING.md` | High |
| Commit convention | Conventional Commits with component scope; DCO `Signed-off-by:` mandatory | inspected — `CONTRIBUTING.md` | High |
| Agent trailers | Maintainer commits carry `Agent: <role>` and `WI: <id>`; internal routing metadata, not required of external contributors | inspected — `CONTRIBUTING.md` | High |
| Review model | `pr-human`; branch unit `wi` | inspected — packet `delivery` | High |
| Release model | Tag-driven; cross-component wave `gateway` → `sdk-*` → `client-*` → `cli` → meta-repo, each publish green before the next | inspected — `CONTRIBUTING.md` | High |
| Published versions | Artifacts at v0.3.0; v0.4.0 in progress for "PI5-yaa" | inspected — `README.md` | High |
| Licence | Apache 2.0; v0.1.x packages remain under the prior source-available licence, non-retroactive | inspected — `LICENSE` reference in `README.md` | High — but see F3 |
| Security contact | `security@aimpathyminds.com`; no public issues for vulnerabilities | inspected — `CONTRIBUTING.md`, `SECURITY.md` | High |
| Docs surface | Astro Starlight in `docs/`, published to `https://ai-mpathyminds.github.io/yaagents/` via `.github/workflows/pages.yml` | inspected — `docs/README.md` | High |
| Submodule contents | — | **unresolved** — not initialised here | **None — see Q3** |
| Submodule gitlink pins | — | **unresolved** — read commands refused | **None — see Q3** |
| Delivery across submodules | — | **unresolved** | **None — see Q2** |
| Governing ADRs / PRD / GTM | — | **unresolved** — cited but absent | **None — see Q4** |

---

## 5. Findings from inspection

Discrepancies the workspace itself shows. These are observations for the grounding cycle to carry
forward. Per `grounding-reads-and-does-not-change`, **none of them was fixed**.

**F1 — `.github/workflows/release.yml` is referenced but absent.**
`CONTRIBUTING.md` §Release process says the release workflow lives at
`.github/workflows/release.yml` and publishes to PyPI, npm, GHCR and the Go proxy. The workflows
actually present are `bench`, `ci`, `gateway-image`, `npm-publish`, `pages`, `pypi-publish`,
`sdk-go-ci`, `sdk-go-smoke`, `supply-chain-audit`, `verify-go-module`. The publishing work looks
split across `npm-publish.yml`, `pypi-publish.yml` and `gateway-image.yml` rather than unified in
one file — so the doc names a file that does not exist. Grounding should record which of the two
is the intended truth.

**F2 — Five CI gates are advisory, deliberately and with stated re-promotion conditions.**
`ci.yml` carries dated `continue-on-error: true` degradations (2026-06-09 post-close): `golangci-lint`
(golangci-lint v2 config-schema churn), `govulncheck-gateway`, `govulncheck-plugins` and
`govulncheck-client-go` (unpatchable Go 1.25.0 stdlib CVEs GO-2026-5037 / GO-2026-5039, awaiting a
Go 1.25.x patch), and `go test` plus the ≥80% coverage gate (17 pre-existing gateway test failures:
16 YAML fixture-indent bugs in `loader_test.go` / `routes_test.go`, and `TestBUMP3_SpecVersion`
expecting `spec/VERSION` = `0.2` when it is `0.3`). The git log at the head of this branch shows the
same pattern. **This is documented, intentional debt with named exit conditions, not drift** — but it
means the Go side of CI is currently non-blocking, which any plan that relies on CI as a gate must
account for.

**F3 — `docs/README.md` calls the project "source-available".**
`docs/README.md` line 4 describes YAAgents as "the source-available Agentic REST Profile", while
`README.md` and `LICENSE` state Apache 2.0 with the source-available terms applying only to
historical v0.1.x packages. The docs README reads as stale relative to the licence change. Low
severity; naming it so grounding does not have to rediscover it.

**F4 — External governing documents are cited from inside this repo but do not live in it.**
ADRs `PI1-yaa-0001`, `PI1-yaa-0005`, `PI2-yaa-0001`, `PI2-yaa-0003`, `PI3-yaa-0004`,
`PI4-yaa-0003`; a PRD (`PRD §5.9`, `§10`); a "GTM README §Appendix"; work-item ids in the
`WI-1yaa.REL-6` / `WI-2yaa.NFR-*` / `WI-4yaa.PLG-*` / `LA-*` series; and PI labels up to `PI5-yaa`.
None of these is present in this repository. They are load-bearing for `yaagents` requirements
(the coverage floor, the crypto/rand floor, the no-dynamic-loading rule, the licensing position all
cite them). Grounding cannot be complete while their home is unknown. See **Q4**.

---

## 6. What a person still has to settle

Named, per `settle-what-you-can-and-name-what-you-cannot`. Q1 and Q2 gate the grounding cycle;
Q3 and Q4 gate its completeness.

**Q1 — One product or eight?**
I read "onboard all as part of yaagent" as: one product `yaagents`, seven submodules as components
under it. The alternative — each submodule onboarded as its own wakalix product with its own cycles
and its own delivery — is a different engine configuration and a different plan. The packet naming a
single `delivery.repository` leans toward my reading, but the packet is the dispatch envelope, not a
statement of product topology. **Confirm the reading before grounding closes.**

**Q2 — Where does managed work land for a submodule change?**
The packet gives one `delivery.repository` and one branch, and `delivery.target_remote` is `null`.
A change to, say, the gateway is a change in `ai-mpathyminds/yaagents-gateway`, which this envelope
does not address. The candidates are: (a) wakalix opens a branch and PR in each submodule repo, then
a follow-on pointer-bump PR in the meta repo; (b) all managed work lands in the meta repo only and
submodule changes are out of scope; (c) submodules are vendored or de-submoduled. This affects
`branch_unit`, review routing and what an evidence bundle can even contain. **I did not assume an
answer.**

**Q3 — Should wakalix worktrees initialise submodules?**
The submodules are empty here, so I could not inspect seven of the eight components, and the command
policy also refused `git submodule status` and `git ls-tree HEAD`, so I could not read their pinned
commits. CI checks out `submodules: recursive` for every job, so the pattern is established
upstream — the gap is local. A `scan-workspace` run against an initialised checkout (with those two
git reads permitted) is the natural next step and would settle the two rows marked *unresolved* in
§4. Onboarding the submodules on inherited description alone would be onboarding a claim, not a
codebase.

**Q4 — Where do the ADRs, the PRD and the GTM README live?**
See F4. Are they in a separate planning repository, in the wakalix record under a cycle this
organisation has not opened yet, or outside the system entirely? `artifact_read()` returned an empty
list, so they are not in the wakalix record today.

---

## 7. What I did not do

Per `say-what-you-did-not-do` and `grounding-reads-and-does-not-change`.

- **I changed nothing.** No source file, workflow, config, submodule pointer or dependency was
  touched. The only file this run creates is this document.
- **I did not fix F1, F2 or F3.** Grounding produces a manifest, not a change; each is a finding for
  the plan.
- **I did not read any submodule.** Seven of eight components are described here from the meta repo's
  own account of them, and that provenance is marked everywhere it applies.
- **I did not read the submodule gitlink pins.** The commands that would show them were refused.
- **I did not produce a `grounding-manifest`.** This packet's contract is
  `produces: [intake-brief]`; the manifest is the other half of the role and belongs to a later
  dispatch, informed by Q1–Q4.
- **I did not plan.** No sequencing, no exit checks, no work items. The onboarding agent does not
  plan, and the catalog says so.
- **I did not run any build, test or lint.** Nothing here rests on an execution result, only on
  reading.
- **I did not answer Q1–Q4 by assumption.** Q1 carries my reading, explicitly labelled as a reading;
  Q2, Q3 and Q4 carry nothing.

---

## 8. What was missing or ambiguous in what I was given

Per `report-what-was-missing`. Stated even where the gap turned out not to block.

- **No acceptance criteria.** `tail.acceptance_criteria` was `[]`, so nothing defines what a
  successfully onboarded `yaagents` looks like. I could not check this brief against a bar; there
  was none.
- **No prior intake and no work item.** `tail.intake` and `tail.work_item_id` were both `null`, and
  `prior` was empty on both keys. Consistent with a first dispatch, and the reason this brief starts
  from the repository rather than from a record.
- **The record is empty.** `artifact_read()` returned zero artifacts, so there was no earlier
  wakalix decision, requirement or manifest to be consistent with. Everything here is first-generation.
- **The `grounding-manifest` dimension vocabulary is not published.** `registry_read` gives type
  descriptions, not schemas. The dimension names in §4 are mine and may not match what
  `scan-workspace` expects. Flagged in place rather than presented as canonical.
- **`context.paths` and `context.excerpts` were both empty.** The packet pointed at nothing, so the
  choice of what to read was entirely mine; §2 is the full list, and a reader should judge coverage
  against it.
- **The intent line is one sentence with a typo (`yaagent`) and no scope boundary.** It does not say
  whether onboarding means the engine should be able to *change* all eight repos or merely *know*
  about them. That is Q2, and it is the single most consequential thing this brief could not settle.
- **The workspace is a worktree with submodules absent**, which is a materially weaker basis for
  onboarding a meta repo than an initialised checkout would have been. That is Q3.
