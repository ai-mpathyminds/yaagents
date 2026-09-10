# YAAgents — grounding manifest

**Product:** yaagents
**Produced by:** onboarding-agent / confirm-grounding
**Work packet:** `01a0898d-1a74-775e-a748-15d610c22497`
**Skill run:** `01a08983-b116-7e75-98ac-8f3c7b7b3b48`
**Date:** 2026-09-10
**Repository home:** `github.com/ai-mpathyminds/yaagents` at `ebf8a71`, submodules fetched at
the pointers that commit records.

This manifest is what grounding learned. It records a value, a confidence and a source for
every dimension, and it names the ones a person settled rather than the repository. It
changes nothing: every gap below is a finding for the plan that accompanies this manifest,
never an edit made here.

---

## Artifacts read

Per *cite-what-you-read*, everything this manifest rests on, by id:

| Id | What | Where it came from |
|---|---|---|
| `01a0898d-1a74-775e-a748-15d610c22497` | This work packet — rules, contract, `tail.input.intent`, `tail.intake` | the dispatch |
| `01a08983-b116-7e75-98ac-8f3c7b7b3b48` | The skill run this manifest closes | the dispatch |
| — | Agent catalogue: `agent-builder`, `architect`, `developer`, `onboarding-agent`, `product-manager`, `reviewer`, `scrum-master`, each read for its skills | `catalog_read` |
| — | Artifact-type registry: 10 types, `grounding-manifest` among them | `registry_read` |

**No portfolio artifact was read, because there are none.** `artifact_read` with no cycle
returned `{"artifacts": []}` — this organisation has produced nothing for yaagents yet, and
`prior.artifacts` in the packet is empty to match. Every finding below therefore comes from
the workspace or from the operator, and nothing from a prior run.

---

## The ten dimensions

Confidence is `settled` (the source is unambiguous), `settled-with-drift` (the source is
unambiguous and something in the tree disagrees with it), or `unclear`.

### 1. Repositories — settled — **inspected**

Eight public repositories under `ai-mpathyminds`: a meta-repo and seven submodules, wired in
`.gitmodules`.

| Path | Repository | Language |
|---|---|---|
| *(root)* | `yaagents` | spec, schemas, OpenAPI, docs site, examples, CI |
| `gateway/` | `yaagents-gateway` | Go |
| `sdk-go/` | `yaagents-sdk-go` | Go |
| `sdk-fastapi/` | `yaagents-sdk-fastapi` | Python |
| `client-python/` | `yaagents-client-python` | Python |
| `client-ts/` | `yaagents-client-ts` | TypeScript |
| `client-go/` | `yaagents-client-go` | Go |
| `cli/` | `yaagents-cli` | Python |

All seven were fetched before grounding, as `tail.input.intent` required. Every one resolved
over public HTTPS with no credential, which is the check that the public flip actually
happened; `tail.intake` §3 says so, and the fetch agrees with it.

**Convention a coding agent gets wrong by default:** a change to component code lands in the
component repository first, and the meta-repo then takes a separate pointer-bump commit. The
log is full of them — `745b0fe` "Bump client-go submodule pointer", `42d8e95` "Bump pointers:
gateway PR#2 + sdk-fastapi PR#1 + cli PR#1". An agent that edits `gateway/` from the
meta-repo checkout and commits there has written to a detached submodule worktree, not to
`yaagents-gateway`.

### 2. Services — settled — **inspected**, deployment posture **asked**

One deployable — the gateway, a Go binary shipped as `ghcr.io/ai-mpathyminds/yaagents-gateway`
(`docker/gateway/`, `gateway/Dockerfile`). Everything else is a library or a tool: two server
SDKs (`sdk-go`, `sdk-fastapi`), three clients (`client-python`, `client-ts`, `client-go`), one
CLI (`cli`). The docs site is an Astro Starlight build under `docs/`, published to GitHub
Pages under the `/yaagents/` base path.

There is no service this organisation runs. See dimension 8.

### 3. Branch model — settled-with-drift — **inspected**

`main` is the default and the base for pull requests; `origin/HEAD` points at it. Work happens
on short-lived topic branches merged by PR — `#25 gotest-advisory`, `#24 govuln-gateway-advisory`,
`#23 lint-degraded`. Long-lived cycle branches (`pi1-yaa`, `pi3-yaa`, `pi4-yaa`) still exist on
the remote and are historical.

**The drift:** `CONTRIBUTING.md` §PR process mandates `feature/<description>`,
`fix/<description>`, `docs/<description>`. No branch in the recent log uses those prefixes.
The documented standard and the practised one disagree, and nothing in the repository says
which wins.

### 4. CI — settled — **inspected**

Ten workflows in `.github/workflows/`. `ci.yml` is the pull-request gate and is the authority
on what blocks; the rest are triggered by tag push or by path.

Blocking today:

- `e2e-tokenvalidator` (6 scenarios), `e2e-tenantinjector` (5), `e2e-licensecheck` (6),
  `e2e-promptsanitize` (5), `e2e-otelaudit` (5) — each preceded by a `no t.Skip` assertion.
- `ci-python` × 3 packages — `ruff check`, `mypy` strict, `pip-audit`, `pytest` at ≥80%.
- `ci-ts` — `eslint`, `tsc --noEmit`, `pnpm audit --audit-level=high`, `vitest` at ≥80%.
- `no-mathr-rand-client-go`, `no-dynamic-load-scan`, `no-secret-in-dockerfile`,
  `no-env-files-in-repo`.

Advisory (`continue-on-error: true`), all degraded on 2026-06-09 with the reason written into
`ci.yml` at the step:

| Step | Why it was degraded | What re-promotes it |
|---|---|---|
| `golangci-lint` | golangci-lint v2 config schema rejects every shape tried | the v2 schema settling, then re-authoring `gateway/.golangci.yml` |
| `govulncheck-gateway`, `govulncheck-plugins`, `govulncheck-client-go` | Go 1.25.0 stdlib CVEs GO-2026-5037 / GO-2026-5039, unpatchable here | a Go 1.25.x patch release and a `go.mod` bump |
| `go test` | 17 pre-existing failures: 16 YAML fixture-indent bugs in `loader_test.go` + `routes_test.go`, and `TestBUMP3_SpecVersion` expecting `0.2` against an actual `0.3` | fixing the 16 fixtures and the bump3 expectation |
| `Enforce ≥80% coverage gate` | depends on `go test` producing a clean profile | re-promoting `go test` |

`license-clean-scan` is advisory by design — it runs on `push` only, deliberately, so fork
contributors are not blocked by a check they cannot run locally.

**This is the intended release posture, not an accident.** `tail.intake` §4 settles it: the
release ships with those gates advisory and the debt carried as GitHub issues. Recorded as
**asked**, not inspected — the workflow says a step is advisory, but only the operator says
that shipping that way is the decision.

### 5. Tests — settled-with-drift — **inspected**

Go: `go test -race -covermode=atomic` with an ≥80% statement gate. Python: `pytest` with
`--cov-fail-under=80` in each `pyproject.toml`. TypeScript: `vitest` with ≥80% thresholds in
`vitest.config.ts`. Conformance suites exist at `sdk-go/conformance/`, `client-go/conformance_test.go`
and `tests/golden/`.

**The drift:** the Go floor is not currently enforced — see dimension 4. Python and TypeScript
floors are.

### 6. Conventions — settled-with-drift — **inspected**

`CONTRIBUTING.md` is the standard: Conventional Commits with a component scope, a DCO
`Signed-off-by:` on every commit, no CLA. Formatters and linters per language —
`gofmt` + `golangci-lint`, `ruff format` + `ruff check` + `mypy --strict`, `prettier` +
`eslint`, and a clean Astro build for MDX. `.markdownlint.json` sits at the root and **no
workflow invokes it** — it is a configured standard with no enforcement.

Maintainer commits additionally carry `Agent:` and `WI:` trailers; `CONTRIBUTING.md` §PR
process states explicitly that external contributors are not required to add them.

**The drift, and it is the sharpest one in the tree:** there are three different contribution
documents claiming authority over the same eight repositories.

- Root `CONTRIBUTING.md` (206 lines) says *"Use GitHub Issues for all bug reports"*, pointing
  at `ai-mpathyminds/yaagents/issues`.
- `gateway/CONTRIBUTING.md` (18 lines) says *"File bugs and feature requests on this repo's
  issue tracker"*, pointing at `yaagents-gateway/issues`. `sdk-go/CONTRIBUTING.md` is the same
  shape.
- `client-ts/CONTRIBUTING.md` (194 lines) is a near-complete second copy of the root document
  — a fork that can drift silently.
- `cli/`, `client-go/`, `client-python/`, `sdk-fastapi/` carry a 9-line pointer stub.

Four shapes, two of which contradict each other about where an issue goes. Nothing in the
repository says which wins.

### 7. Existing agent setup — settled — **inspected**

**There is none.** No `.claude/` directory, no `CLAUDE.md`, no `AGENTS.md`, no `.mdc` rule
file, anywhere in the meta-repo or in any of the seven submodules. No knowledge base exists
for any of the seven agents `catalog_read` returns.

This is the largest single gap between this product and grounded, and the plan's `K-1` is the
entry that closes it.

### 8. Deployment model and environments — settled — **asked**

There is no deployment and there are no environments. `tail.input.intent` states it plainly:
*"its a direct release of artifacts to opensource."* The release surface is five public
registries plus a site:

| Artifact | Registry | Workflow | Trigger |
|---|---|---|---|
| `yaagents-fastapi`, `yaagents-client`, `yaagents-cli` | PyPI, via OIDC Trusted Publisher | `pypi-publish.yml` | tag `v*.*.*` |
| `@aimpathyminds/yaagents-client` | npm, with provenance | `npm-publish.yml` | tag `v*.*.*` |
| `yaagents-gateway` image | GHCR, Cosign + Syft SBOM | `gateway-image.yml` | tag `v*.*.*` |
| `yaagents-sdk-go`, `yaagents-client-go` | Go module proxy | `verify-go-module.yml` | tag `client-go/v*.*.*` |
| spec archive (`schemas/` + `openapi/`) | GitHub Release asset | `ci.yml` job `spec-archive` | tag, after all gates |
| docs site | GitHub Pages | `pages.yml` | push to `main` touching `docs/`, `spec/`, `schemas/` |

A version bump is a **release wave**: `CONTRIBUTING.md` §Release process requires all eight
components tagged at the same version in the same cycle, in the order
`gateway` → `sdk-*` → `client-*` → `cli` → meta-repo, each publish succeeding before the next
tag. No trusted-publisher or registry credential could be verified from a checkout; the
workflows are wired, and whether each registry's side of the trust relationship is live is
outside what a read of the repository can answer.

**A gap in the written process, not in the wiring:** `CONTRIBUTING.md` §Release process names
`.github/workflows/release.yml` as the workflow that publishes everything. That file does not
exist. Publishing is done by the four separate workflows above. Anyone — person or agent —
following the documented release process looks for a file that is not there.

### 9. Documentation — settled — **asked**

`tail.intake` §1: documentation lives in the same repository, at `docs/`. §2: the earlier
documents are old and retired, and the same `docs/` folder is to hold new ADRs and other
artifacts.

What is at `docs/` today is an Astro Starlight **site**, not a document folder — `package.json`,
`astro.config.mjs` and `src/content/docs/**` of published `.mdx`. Its sidebar already has an
*Architecture* section (`src/content/docs/architecture/`). This manifest is written to
`docs/grounding/`, which sits outside `src/` and is therefore not built into the site; note
that `pages.yml` watches `docs/**`, so anything added under `docs/` triggers a site rebuild
even when it adds no page.

**What "retired" makes stale, and it is not cosmetic.** Every authority document the tree
cites by name is one of the retired internal ones, and none of them is in this repository:

- `ci.yml` cites ADR PI1-yaa-0005 §6, ADR PI2-yaa-0001 §3, ADR PI2-yaa-0003, ADR PI4-yaa-0003.
- `CONTRIBUTING.md` cites ADR PI2-yaa-0003 §3 and the GTM README appendix for its verbatim
  legal disclaimer, and the PI4-yaa launch-readiness runbook for the release-wave pattern.
- `bin/yaagents-public-mirror-verify.sh` cites `yaagents/docs/PI3-yaa/roadmap.md`,
  `yaagents/docs/PI3-yaa/platform-engineer.md` and `portfolio/RUNBOOKS/pi3-yaa-planning.yml`.

The rules those citations point at are still in force — the CI gates that enforce them are
green. What is gone is anything a reader could open to see *why*. Where a new ADR belongs —
published on the site under `architecture/`, or repo-only alongside this manifest — is a
structural choice nobody has recorded, and it is the plan's `D-3`.

### 10. Tracker — settled-with-drift — **asked**

GitHub Issues. `tail.input.intent`: *"the opensource community will use the githu issues
tracking."* Four issue-form templates exist at `.github/ISSUE_TEMPLATE/` — `bug.yml`,
`feature.yml`, `adapter-request.yml`, `plugin-proposal.yml` — and only in the meta-repo; none
of the seven component repositories has one, though `gateway/CONTRIBUTING.md` sends
contributors to its own tracker.

Two smaller disagreements, both citable:

- `plugin-proposal.yml` applies `labels: ["plugin", "triage"]`. `CONTRIBUTING.md`'s label
  table documents `plugin-proposal` and does not mention `triage` at all.
- `docs/src/content/docs/roadmap.mdx` routes feature prioritisation to **GitHub Discussions**
  in three places, and the operator named Issues. Discussions and Issues can coexist, but
  which one a community feature request is supposed to land in is not currently answerable
  from either the docs or `CONTRIBUTING.md`.

### 11 (beyond the ten). Process tooling — settled — **asked**

`tail.input.intent`: *"Use runbooks as process tools."* **There is no runbook anywhere in the
tree** — no `runbook` file in the meta-repo or in any submodule. The only runbook the
repository names, `portfolio/RUNBOOKS/pi3-yaa-planning.yml`, is one of the retired internal
documents. The processes that most need one — the eight-component release wave, and issue
triage now that the community is the one filing — exist today only as prose in
`CONTRIBUTING.md`. The plan's `P-1` is the entry that writes them.

---

## What grounded requires, and where this product stands

| Requirement | Standing |
|---|---|
| Every dimension has a value and a source, and none waits on a person | **Met.** Ten of eleven settled by inspection or by the operator's answers; the eleventh (process tooling) settled by `tail.input.intent`. Nothing below is `unclear`. |
| Every agent has a non-empty knowledge base | **Not met.** No knowledge base exists for any of the seven catalogue agents — dimension 7. → `K-1` |
| The standards are written down | **Partly met.** The coding standard is written (`CONTRIBUTING.md` §Code style) and enforced. The definition of done exists as a PR checklist. **No review bar is written down anywhere** — nothing states what a maintainer must find before approving. → `R-1` |
| Every external system is connected, or a connector is written into their own repository | **Met at the wiring, with two named holes.** PyPI, npm, GHCR, the Go proxy and Pages each have a workflow. `CONTRIBUTING.md` points at a `release.yml` that does not exist, and `bin/yaagents-public-mirror-verify.sh` documents a CI integration at `.github/workflows/scrub-verify.yml` that does not exist either. → `F-1` |
| No agent is blocked from release by an unclear dimension | **Met, once the authority conflicts are recorded.** No dimension is unclear. Three are `settled-with-drift`, and drift is what `D-1`, `D-2` and `D-4` resolve — an agent asked to change `client-ts` today has two contribution standards and no rule for choosing. |

### Drift, in one place

| # | The two sources | Which the plan makes win |
|---|---|---|
| 1 | Root `CONTRIBUTING.md` vs `client-ts/CONTRIBUTING.md` (full second copy) vs 5 pointer stubs vs `gateway`/`sdk-go` "file here" | `D-1` |
| 2 | `spec/VERSION` = `0.3` and `sdk-go/sdkgo/profile.go` = `v0.3`, against `client-go/version.go` = **`v0.2`** | `D-2` |
| 3 | `CONTRIBUTING.md` branch prefixes (`feature/`, `fix/`, `docs/`) vs every branch actually merged | `D-1` |
| 4 | `CONTRIBUTING.md` label `plugin-proposal` vs the template's `plugin` | `F-1` |
| 5 | `roadmap.mdx` — `otel-audit` and `prompt-sanitize` "planned for v0.4" — vs `README.md`, which calls all five plugins Stable, and against `ci.yml`, which has had blocking e2e gates for both since PLG-12/PLG-15 | `F-1` |
| 6 | `roadmap.mdx` claims `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md` and `SECURITY.md` shipped across "8 public repos". **All seven component repos have only `CONTRIBUTING.md`** — no `SECURITY.md`, no `CODE_OF_CONDUCT.md`, in any of them | `C-1` |
| 7 | The PR checklist requires a `CHANGELOG.md` entry under `[Unreleased]`. **`gateway`, `sdk-fastapi`, `client-go` and the meta-repo have no `CHANGELOG.md` at all** — only `cli`, `client-python`, `client-ts` and `sdk-go` do | `F-1` |
| 8 | `roadmap.mdx` sends feature requests to Discussions; the operator named Issues | `D-4` / `F-1` |
| 9 | Source is at `0.4.0` across every package (`__about__.py` × 3, `package.json`); `spec/VERSION` is `0.3`; `README.md` says published artifacts are `v0.3.0` and v0.4.0 ships next | `D-2` — expected mid-wave, but nothing states that `spec/VERSION` tracks the Profile and not the package |

---

## What this run did not do

- **It changed nothing in the product.** The nine drifts above are findings; not one was
  fixed. The only file this run wrote is this manifest.
- **It did not fetch or read the component repositories' own histories** beyond the pinned
  commit each submodule pointer names. Grounding read the tree at those pointers, not each
  repository's `main`.
- **It could not verify any registry credential or trust relationship.** PyPI trusted
  publishing, npm provenance, GHCR write and Cosign keys are asserted by the workflows and
  are not checkable from a checkout. If any of them is stale the release wave fails at the
  first tag, and no reading of this repository would have predicted it.
- **It could not verify GitHub repository settings** — branch protection on `main`, whether
  Discussions is enabled, the label set actually configured. Those live in GitHub, not in the
  tree, and dimension 3, 10 and drift 8 are grounded on what the repository says rather than
  on what the org is configured to do.
- **It did not run any test or gate.** The advisory-vs-blocking table in dimension 4 is read
  from `ci.yml`, not observed from a run.

## What was missing or ambiguous in what this run was given

- **`prior.attempts` records one attempt with an empty `question` and an empty `answer`,**
  yet `tail.intake` is four numbered answers. The questions those four answered are not in the
  packet. This manifest therefore reads §1–§4 against the numbering alone; if the earlier turn
  asked something this reads differently, the misreading would be in dimensions 8, 9 and 10.
- **`tail.intake` §4 is punctuated as a question** — *"the release ship with them advisory and
  the debt tracked as issues?"* — and is taken here as the operator's answer rather than as a
  question back, because `tail.input.intent` points the same way and commits `56d129d`,
  `58a1bd8` and `890fe9a` already made those gates advisory. It is recorded as **asked**. If
  it was in fact a question, dimension 4 and entry `D-4` are what change.
- **`contract.accepts` and `contract.produces` are both empty**, and
  `tail.acceptance_criteria` is empty. Nothing in the packet states what a confirm-grounding
  run must produce. This manifest is written against the artifact-type registry's
  `grounding-manifest` description and the catalogue's statement that this agent produces one.
- **`tail.work_item_id` is null**, so this manifest is attached to a skill run and not to a
  work item.
- **The catalogue holds no knowledge-base-authoring role.** *"Every agent has a knowledge base
  that is not empty"* is a grounding requirement, and none of the seven agents has a skill
  whose job is writing one. The plan routes `K-1` through `developer` / `implement-change`,
  because a knowledge base here is a file on a branch. That is the closest fit in the
  catalogue, not a role that exists. The same is true of writing GitHub issues: `I-1` produces
  a debt register in the repository, which a person or a runbook then mirrors into the tracker,
  because no catalogue skill writes to a tracker.

---

## The plan

The plan this run proposes travels in the work-packet result, not in this repository — the
record holds the plan, and a copy committed here would be the copy that disagrees. It is
twelve entries: four decision records that settle which source wins (`D-1`–`D-4`), a
requirement set for the standards that are not written down and its mechanical check
(`R-1`, `R-2`), five changes that close the gaps (`K-1` the agent knowledge base, `C-1` the
component repos' community-health files, `P-1` the release-wave and issue-triage runbooks,
`F-1` the documented-vs-actual mismatches, `I-1` the CI debt register), and a review (`V-1`).
`plan_check` reports no findings across all twelve.
