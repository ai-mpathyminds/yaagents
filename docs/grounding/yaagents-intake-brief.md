# YAAgents — Grounding Intake Brief

| Field | Value |
|---|---|
| Product | `yaagents` |
| Repository | `github.com/ai-mpathyminds/yaagents` |
| Produced by | onboarding-agent · skill `confirm-grounding` |
| Artifact type | `intake-brief` (`application/vnd.wakalix.intake-brief+json; v=1`) |
| Work packet | `01a07098-45c1-74e8-aff7-b18418c3586e` |
| Skill run | `01a07098-0e13-7c47-b49a-00ba5d374675` |
| Branch inspected | `wakalix/run-01a070980e13` (base `main`), at commit `ebf8a71` |
| Date | 2026-09-05 |

This is the intake a grounding cycle opens from. It records what was asked, what
inspection of the workspace settled, and — by name — what it did not settle.
It changes nothing in the workspace: every gap below is a finding for the plan,
not an edit made now.

---

## 1. The ask, as given

The operator's intent, verbatim from the work packet:

> The repo is meta repo and it contains multiple sub modules , use all to do the grounding

Nothing else was supplied. `tail.acceptance_criteria` was an empty list,
`tail.intake` was `null`, `tail.work_item_id` was `null`, and
`context.excerpts` and `context.paths` were both empty. The ask is therefore
read as: **ground the `yaagents` product across the meta-repository and all
seven of its submodules.**

## 2. What was read

**Wakalix record.** The portfolio artifact store was queried and is empty:
`artifact_read` with no arguments returned `{"artifacts": []}`. The packet's
`prior.artifacts` and `prior.attempts` were likewise both empty. **No prior
artifact exists for this product, so none is cited below — every finding in
this brief comes from repository inspection, not from a previous cycle.** The
agent catalog was read for role `onboarding-agent`, and the artifact-type
registry was read to confirm what an `intake-brief` is.

**Repository files inspected**, in full:

| Path | What it settled |
|---|---|
| `README.md` | Product statement, response-profile table, repository layout, install surface |
| `.gitmodules` | The seven submodule paths and their remotes |
| `CONTRIBUTING.md` | Dev setup, PR process, code style, release process, plugin contribution path |
| `SECURITY.md` | Supported versions, disclosure SLA, scope, security defaults, scanning |
| `.github/workflows/ci.yml` | Every quality gate, and which are currently advisory |
| `docs/astro.config.mjs`, `docs/src/content/config.ts` | Docs site build scope and navigation |
| `docs/src/content/docs/start-here/overview.mdx` | Positioning, layering, explicit non-goals |
| `docs/src/content/docs/roadmap.mdx` | v0.3 shipped inventory, v0.4 plan, community asks |
| `docs/src/content/docs/sdks/index.mdx`, `docs/src/content/docs/sdks/cli.mdx` | SDK target list; CLI command surface |
| `spec/VERSION` | Profile version `0.3` |
| `bin/yaagents-public-mirror-verify.sh` | Public-mirror hygiene gate over the eight repos |

**Directory listings taken** (paths enumerated, contents not read): `spec/`,
`schemas/`, `openapi/`, `docs/`, `examples/`, `tests/`, `tools/`, `docker/`,
`.github/`.

## 3. Dimension table

Source is one of **inspected** (read in this workspace), **asked** (supplied by
the operator), or **assumed**. There are no `assumed` rows: where inspection did
not settle a dimension it is marked `unsettled` and carried to §5 rather than
being resolved silently.

| # | Dimension | Source | Confidence | Settled? |
|---|---|---|---|---|
| 1 | Product identity and purpose | inspected | high | yes — §4.1 |
| 2 | Repository topology | inspected + asked | high | yes — §4.2 |
| 3 | Public contract (Profile / schemas / OpenAPI) | inspected | high | yes — §4.3 |
| 4 | Languages and toolchains | inspected | medium | partly — §4.4, gap in §5.4 |
| 5 | Build and test commands | inspected | medium | partly — §4.4, gap in §5.4 |
| 6 | CI gates and current health | inspected | high | yes — §4.5 |
| 7 | Release and distribution | inspected | medium | partly — §4.6, gap in §5.5 |
| 8 | Security posture | inspected | high | yes — §4.7 |
| 9 | Governance and contribution | inspected | high | yes — §4.8 |
| 10 | Docs site | inspected | high | yes — §4.9 |
| 11 | Examples and reference stacks | inspected | high | yes — §4.10 |
| 12 | **Submodule internals** | — | **none** | **no — §5.1** |
| 13 | **Runtime environments / deployment targets** | — | **low** | **no — §5.2** |
| 14 | **Ownership and on-call** | — | **none** | **no — §5.3** |
| 15 | **Goal of this grounding** | — | **none** | **no — §5.6** |

## 4. What inspection settled

### 4.1 Product identity and purpose

YAAgents is a gateway, SDK layer, and response contract for exposing agentic
capabilities as governed domain resource operations
(`docs/src/content/docs/start-here/overview.mdx:10`). It governs the
application-to-agent boundary — auth, tenant context, audit, typed responses,
OpenAPI contracts — while leaving the agent framework to the user.

The overview page states explicit non-goals (`overview.mdx:63-71`): not an agent
framework, not a chatbot framework, not a replacement for A2A, AGNTCY or MCP,
not a model-provider abstraction, and not a router or load balancer. These are
recorded here because a grounding cycle that mistakes any of them for in-scope
work would plan against the wrong product.

### 4.2 Repository topology

A meta-repository with seven submodules, declared in `.gitmodules`:

| Path | Remote | Role per `README.md:150-156` |
|---|---|---|
| `gateway` | `yaagents-gateway.git` | Go gateway source |
| `sdk-fastapi` | `yaagents-sdk-fastapi.git` | Python FastAPI server SDK (`yaagents-fastapi`) |
| `sdk-go` | `yaagents-sdk-go.git` | Go server SDK (`yaagents-sdk-go`) |
| `client-python` | `yaagents-client-python.git` | Python client (`yaagents-client`) |
| `client-ts` | `yaagents-client-ts.git` | TypeScript client (`@aimpathyminds/yaagents-client`) |
| `client-go` | `yaagents-client-go.git` | Go client SDK (`yaagents-client-go`) |
| `cli` | `yaagents-cli.git` | CLI validator (`yaagents-cli`) |

All remotes are under the `ai-mpathyminds` GitHub organisation.
`docs/src/content/docs/roadmap.mdx:33` states the public footprint is **8 repos —
this meta-repository plus the 7 component submodules**, and
`bin/yaagents-public-mirror-verify.sh:46-47` names the same eight by name.

The meta-repository itself owns, directly and not via any submodule: `spec/`,
`schemas/`, `openapi/`, `docs/`, `examples/`, `tests/golden/`, `tools/gen/`,
`docker/gateway/`, `bin/`, and `.github/`.

### 4.3 Public contract

The normative contract is the **Agentic REST Response Profile v0.3**
(`spec/agentic-rest-profile.md`; `spec/VERSION` reads `0.3`). `README.md:88-99`
gives the response-type table — ten outcomes, each with a fixed HTTP status and
a dedicated media type:

`success` 200 · `created` 201 · `accepted` 202 · `clarification_required` 400 ·
`validation_failed` 422 · `approval_required` 412 · `forbidden` 403 ·
`conflict` 409 · `failed_dependency` 424 · `error` 500.

Supporting material in the meta-repository:

- `schemas/v0.1/`, `schemas/v0.2/`, `schemas/v0.3/` — six JSON schemas per
  version (`agentic-error`, `approval-required`, `clarification-required`,
  `conflict`, `operation-accepted`, `validation-failed`). All three versions are
  retained side by side.
- `openapi/yaagents-components.yaml`, `openapi/yaagents-response-profile.yaml`.
- `spec/examples/v0.1/` — valid and deliberately-invalid conformance fixtures,
  indexed by `spec/examples/INDEX.md`.
- `tests/golden/` — ten golden response bodies, one per outcome type.
- Responses carry an `X-YAAgents-Profile: v0.3` header
  (`docs/src/content/docs/roadmap.mdx:20`) and a mandatory `trace` block
  (`README.md:101`).

Five first-party gateway plugins are named consistently across
`README.md:12`, `docs/astro.config.mjs:27-33` and the CI job list:
`token-validator`, `tenant-injector`, `license-check`, `prompt-sanitize`,
`otel-audit`. A `cors` plugin is additionally listed at `roadmap.mdx:19`.
`README.md:12` marks all five as Stable.

### 4.4 Languages, toolchains, build and test

Declared minimums (`README.md:4-5`): Go 1.22+, Python 3.11+.
Versions CI actually pins (`.github/workflows/ci.yml`): Go resolved from each
module's `go.mod`, Python 3.12, Node.js 22, pnpm 9.

Per-component commands, from `CONTRIBUTING.md:45-75`:

| Component | Build / test |
|---|---|
| `gateway`, `sdk-go` | `go build ./... && go test ./...` |
| `sdk-fastapi`, `client-python`, `cli` | `pip install -e ".[dev]" && hatch run test` |
| `client-ts` | `pnpm install && pnpm test` |
| `docs` | `pnpm install && pnpm run dev` → `http://localhost:4321/yaagents/` |
| Demos | `docker compose up` in `examples/campaign-api` or `examples/campaign-api-go` |

Lint and type-check tooling (`CONTRIBUTING.md:121-142`): Go — `gofmt`,
`golangci-lint`; Python — `ruff format`, `ruff check`, `mypy --strict`;
TypeScript — `prettier`, `eslint`; MDX — a clean `pnpm run build`.

Coverage floors are 80% in all three language families: `ci.yml:116-125` (Go),
`--cov-fail-under=80` in each Python package's `pyproject.toml`
(`ci.yml:455-457`), and `vitest.config.ts` thresholds (`ci.yml:559-561`).

Note the toolchain rows are marked *medium* confidence, not high: these are the
commands the meta-repository **documents**, and none could be executed or
cross-checked against a submodule manifest here. See §5.1 and §5.4.

### 4.5 CI gates, and the health of them

`.github/workflows/ci.yml` defines these jobs: `ci-go`, `ci-go-client`,
`e2e-tokenvalidator` (6 scenarios), `e2e-tenantinjector` (5),
`e2e-licensecheck` (6), `e2e-promptsanitize` (5), `e2e-otelaudit` (5),
`ci-python` (matrix over `sdk-fastapi`, `client-python`, `cli`), `ci-ts`,
`no-dynamic-load-scan`, `license-clean-scan`, `no-secret-in-dockerfile`, and
`spec-archive` (tag pushes only). Nine further workflows exist: `bench.yml`,
`gateway-image.yml`, `npm-publish.yml`, `pages.yml`, `pypi-publish.yml`,
`sdk-go-ci.yml`, `sdk-go-smoke.yml`, `supply-chain-audit.yml`,
`verify-go-module.yml`.

Every checkout in `ci.yml` uses `submodules: recursive` — CI is the only place
the eight repositories are assembled into one tree.

**Six gates are currently degraded to advisory** (`continue-on-error: true`),
each with an inline comment dated 2026-06-09 explaining why:

| Gate | `ci.yml` | Stated reason |
|---|---|---|
| `golangci-lint` | :59-72 | golangci-lint v2 config schema rejects every config shape tried; waiting for the schema to settle |
| `govulncheck-gateway` | :87-89 | Go 1.25.0 stdlib CVEs (`GO-2026-5037`, `GO-2026-5039`) the project cannot patch; awaiting a Go 1.25.x patch |
| `govulncheck-plugins` | :96-98 | same stdlib CVEs |
| `go test` (gateway) | :109-111 | 17 pre-existing failures — 16 YAML fixture-indent bugs in `loader_test.go` / `routes_test.go`, plus `TestBUMP3_SpecVersion` still expecting `spec/VERSION` `0.2` |
| ≥80% coverage gate | :116-125 | depends on `go test` producing a clean profile |
| `govulncheck-client-go` | :166-173 | same stdlib CVEs |

The two most recent commits on this branch's history corroborate this:
`56d129d` "Degrade go test + coverage gate to advisory (pre-existing
fixture/spec bugs)" and `58a1bd8` "Degrade govulncheck-gateway +
govulncheck-plugins to advisory". **The remediation work is already named in
those comments** — fix the 16 fixture YAMLs to 2-space indent, update the
version-bump test to expect `0.3`, re-author the linter config against the
settled v2 schema, and re-promote all six gates once the Go stdlib CVEs clear.
This is the largest single block of known, undisputed outstanding work in the
workspace.

Gates that remain blocking and are worth preserving: `no-mathr-rand-client-go`
(`ci.yml:179-190`, enforcing a `crypto/rand` floor for UUID v4),
`no-dynamic-load-scan` (`ci.yml:570-596`, forbidding `plugin.Open`/`dlopen` —
community plugins are compiled in via module imports), `no-secret-in-dockerfile`
and `no-env-files-in-repo` (`ci.yml:627-652`), and the five e2e plugin gates,
each of which additionally asserts that no `t.Skip*` call has been introduced
into its e2e test file.

### 4.6 Release and distribution

Releases are tag-driven: pushing a `v[0-9]+.[0-9]+.[0-9]+` tag triggers publish
(`CONTRIBUTING.md:148-152`, `ci.yml:24-25`). Published artifacts are at
**v0.3.0** across all seven distribution channels (`roadmap.mdx:22-29`):
`yaagents-fastapi`, `yaagents-client`, `yaagents-cli` (PyPI);
`@aimpathyminds/yaagents-client` (npm);
`github.com/ai-mpathyminds/yaagents-sdk-go`,
`github.com/ai-mpathyminds/yaagents-client-go` (Go modules);
`ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` (GHCR, multi-arch).

A major bump must tag **all eight components at the same version in the same
cycle**, in the order `gateway` → `sdk-*` → `client-*` → `cli` → meta-repository,
each publish succeeding before the next is tagged (`CONTRIBUTING.md:160-164`).
This cross-repository release wave is the sharpest coupling constraint in the
product and should survive into any plan.

`README.md:12` states v0.4 is in progress and will ship alongside the Helm chart
and the full publish wave.

### 4.7 Security posture

`SECURITY.md` supports `0.3.x` and backports fixes to `0.2.x`; `0.1.x` and
earlier are unsupported. Private reporting via `security@aimpathyminds.com` or a
GitHub Security Advisory; acknowledgement within 3 business days, initial
assessment within 7, and a **90-day coordinated-disclosure SLA** applied
universally regardless of severity. No bug bounty.

In scope: gateway authn/authz bypass, tenant-isolation failure, RBAC bypass,
secret exposure via logs/errors/image layers, supply-chain CVEs, injection and
DoS input-validation gaps. Out of scope: the user's own agent implementation,
theoretical issues, physical access, social engineering, and resource-exhaustion
DoS against a correctly-configured user deployment.

Gateway security defaults (`SECURITY.md:92-99`): non-root container user,
read-only filesystem where applicable, no baked-in secrets, JWT validation
required on all upstream routes except in explicit dev mode, structured JSON
logs with no credential values, graceful shutdown.

Licensing is Apache 2.0 across all packages and repositories
(`roadmap.mdx:32`, `README.md:170`). v0.1.x packages shipped under the previous
source-available licence remain under it, non-retroactively; `license-clean-scan`
(`ci.yml:603-621`) blocks regression to the old licence identifier in source.
`CONTRIBUTING.md:5-9` carries a verbatim legal disclaimer that the licensing
strategy is not legal advice.

`bin/yaagents-public-mirror-verify.sh` is a hygiene gate run across all eight
working trees before any private→public flip: it fails on internal planning
directories or filename patterns, and greps source and docs for five internal
path markers, with an allowlist that permits legitimate architecture-decision
citations. Anything written into these repositories has to stay clean against
it — **including this document**, which is why the cycle markers in `ci.yml` and
in that script are referenced above by path and line rather than quoted.

### 4.8 Governance and contribution

DCO `Signed-off-by:` on every commit, no CLA (`roadmap.mdx:35`,
`CONTRIBUTING.md:93-109`); PRs without it fail the DCO check and cannot be
merged. Conventional Commits with a component scope. Branch naming
`feature/…`, `fix/…`, `docs/…`. Seven issue labels are defined, and four issue
templates exist (`adapter-request`, `bug`, `feature`, `plugin-proposal`).

`CONTRIBUTING.md:111-115` records an internal maintainer convention: commits from
the AimpathyMinds development workflow carry additional `Agent:` and `WI:`
trailers, which are routing metadata only and are **not** required of external
contributors. This matters for grounding — it is the existing, documented
convention for machine-authored commits in these repositories.

Community plugins follow a five-step path (`CONTRIBUTING.md:183-199`): propose,
implement at `gateway/plugins/<name>/plugin.go` registering via `init()`,
publish standalone as `github.com/<you>/yaagents-plugin-<name>`, ship unit tests,
document the YAML config schema.

### 4.9 Docs site

Astro Starlight, deployed to GitHub Pages at
`https://ai-mpathyminds.github.io/yaagents/` — `site` and `base: '/yaagents'` in
`docs/astro.config.mjs:5-6`, published by `.github/workflows/pages.yml`.
The base prefix is load-bearing: the most recent commit on this branch,
`ebf8a71`, is a hotfix restoring the `/yaagents/` base prefix on hero CTAs.

**The site builds only `docs/src/content/docs/**`** — that is the sole content
collection (`docs/src/content/config.ts`), and the sidebar is explicit apart from
two autogenerated directories, `sdks` and `examples`
(`docs/astro.config.mjs:49-50`). This document sits at `docs/grounding/`,
outside `docs/src/`, and is therefore not part of the published site.

Content covers Start Here, Quick Start, Profile Spec, five plugin pages, three
Concepts pages, one How-to, six SDK quickstarts, three Examples pages, one Case
Study, two Architecture pages, Roadmap, Contributing, and Community.

### 4.10 Examples and reference stacks

Seven stacks under `examples/`: `campaign-api` (Python FastAPI reference, all
five agentic flows, with `tests/`), `campaign-api-go` (Go `net/http` reference,
same flows), `agent-graph-ecom` (two agents behind two gateways),
`customer-support-triage`, `financial-risk-screening`, `llm-gateway` (mock IAM
and LLM APIs plus a community-plugin sample), and `store`. Each ships a
`docker-compose.yml` and gateway `routes.yaml` / `plugins.yaml` config.

## 5. What was not settled, by name

Each row below is a finding for the grounding plan. None has been resolved by
assumption.

### 5.1 Submodule internals — not inspected at all

**This is the gap that matters most, because it is exactly what the ask named.**

The seven submodule working trees are **empty in the workspace this run was
given**. `sdk-go/`, `gateway/`, `cli/`, `client-go/`, `client-ts/`,
`client-python/` and `sdk-fastapi/` each contain no files here; the
submodules are recorded as gitlinks but are not checked out. The main checkout
of the repository lies outside this run's permitted read scope, and `git`
subcommands that would have reported the pinned submodule commits
(`git submodule status`, `git ls-tree HEAD`) were declined by the sandbox, so
even the pinned SHAs could not be captured.

Everything §4 says about a submodule is therefore **meta-repository evidence
about that submodule** — its remote, its published package name and version, the
commands documented for it, the CI job that builds it — and never a reading of
its source. Specifically **not** established: each submodule's actual module
manifest and dependency set, its real source layout, its test suite and current
pass rate, its `CHANGELOG.md`, its per-component `README.md`, its own workflow
files, and the commit each is pinned to.

Concretely unverifiable from here, though asserted by the meta-repository:
`gateway/go.mod` and `gateway/.golangci.yml` exist (`ci.yml:53-58`);
`gateway/internal/plugins/{tokenvalidator,tenantinjector,licensecheck,promptsanitize,otelaudit}/`
each hold a `*_e2e_test.go`; `client-go` has zero non-stdlib runtime
dependencies and therefore no `go.sum` (`ci.yml:151-157`); the three Python
packages each carry `--cov-fail-under=80`; `client-ts` uses a `pnpm-lock.yaml`
and a `vitest.config.ts`. Note also that `CONTRIBUTING.md:189-192` locates the
plugin interface at `gateway/plugins/plugin.go` while `ci.yml` consistently
addresses plugins under `gateway/internal/plugins/` — the two cannot both be the
whole story, and only the submodule can say which is right.

**Resolution:** the grounding cycle needs the submodules populated
(`git submodule update --init --recursive`, or the clone form at
`CONTRIBUTING.md:47`) inside a workspace the agent can read, and then a
`scan-workspace` pass over all eight trees. Until that happens, no dimension
below the meta-repository layer can be raised above the confidence recorded in
§3.

### 5.2 Runtime environments and deployment targets

The workspace documents how to *build and publish*, and `docker/gateway/Dockerfile`
plus the compose files show how to *run a demo*. It does not name any
environment the product actually runs in — no staging or production target, no
cluster, no account, no promotion path. `roadmap.mdx:55-58` lists a Kubernetes
Helm chart as **planned for v0.4**, which implies no supported Kubernetes
deployment path exists today, but does not describe what is deployed instead.
The one live environment evidenced is the GitHub Pages docs site (§4.9).

### 5.3 Ownership and on-call

No `CODEOWNERS`, no maintainer roster, no on-call rotation was found.
`CONTRIBUTING.md:166` says only that "maintainers with write access to the
`ai-mpathyminds` organization" can publish releases, without naming them.
Two contact addresses exist — `security@aimpathyminds.com` (`SECURITY.md:19`)
and `bhaskar@aimpathyminds.com` for historical licence questions
(`README.md:170`).

### 5.4 Four documentation-vs-CI discrepancies

Each is small, each is a real inconsistency between two files in this
repository, and none was silently resolved:

1. **A release workflow that is referenced but absent.** `CONTRIBUTING.md:154`
   names `.github/workflows/release.yml` as the workflow that publishes to PyPI,
   npm, GHCR and the Go proxy. No such file exists; a repository-wide search for
   the string finds only that one line. The publishing work appears to be split
   across `pypi-publish.yml`, `npm-publish.yml`, `gateway-image.yml` and
   `verify-go-module.yml`. Which of those is authoritative, and whether
   `release.yml` was renamed or never existed, is unsettled.
2. **A broken example link.** `README.md:62` links `examples/store-go/` as the Go
   run-it path. That directory does not exist; the Go reference stack is
   `examples/campaign-api-go/`, which is what `CONTRIBUTING.md:73-74` uses.
3. **Node version.** `CONTRIBUTING.md:63,66` says Node 20+ for `client-ts` and
   `docs`; `ci.yml:530-532` pins Node 22.
4. **npm vs pnpm audit.** `SECURITY.md:113` says every release runs `npm audit`
   on the TypeScript package; `ci.yml:556-557` runs
   `pnpm audit --audit-level=high`. `SECURITY.md:110` also names `trivy` on the
   gateway image, which does not appear in `ci.yml` — it may live in
   `supply-chain-audit.yml` or `gateway-image.yml`, neither of which was read.

### 5.5 Supply-chain claims not verified

`SECURITY.md:115` states an SBOM is attached to each GitHub Release, and
`CONTRIBUTING.md:156-159` states Cosign signing, Syft SBOM and npm provenance are
in place today. `roadmap.mdx:59-63` lists "Full Cosign signing + SLSA
provenance", "SBOM attached to GHCR image at publish time (Syft)" and "Signed
npm provenance attestation" as **planned for v0.4**. These two readings
disagree about what is already shipped. Settling it requires reading
`gateway-image.yml`, `npm-publish.yml`, `pypi-publish.yml` and
`supply-chain-audit.yml`, which this run did not do.

### 5.6 What this grounding is for

The ask says to ground the product across all submodules. It does not say what
the grounding is in service of — a v0.4 release push, the CI-gate remediation of
§4.5, onboarding a new contributor, an audit, or something else. The intent is
unambiguous about *scope* and silent about *purpose*. This does not block
producing this brief, but it does shape what a `scan-workspace` pass should
weight, and it is the single most useful thing an operator could add before the
next step.

## 6. What this run did not do, and why

- **It did not read any submodule source.** The submodule trees are empty in the
  workspace supplied, and the populated checkout is outside this run's permitted
  read scope. This is the §5.1 gap, and it is a limit of the workspace, not a
  choice.
- **It did not run any build, test, or lint command.** Nothing is installed
  against, and with the submodules absent there is nothing to build. Every
  toolchain statement in §4.4 is documentation, not observation.
- **It did not change anything.** Grounding reads and does not change: the CI
  gates of §4.5, the broken link and the four discrepancies of §5.4 are written
  down as findings, not fixed. Fixing them is a plan's business.
- **It did not produce a grounding manifest.** That is the separate
  `scan-workspace` skill, and it should run after §5.1 is resolved.
- **It did not read `spec/agentic-rest-profile.md` in full**, nor the schema
  files, nor `openapi/`, nor the nine non-`ci.yml` workflows, nor the individual
  example stacks. These were enumerated, not opened. The response-profile table
  in §4.3 is taken from `README.md`, and if the normative spec and the README
  ever disagree, the spec wins.
- **It did not capture the submodule pinned commits.** The `git` invocations that
  would have reported them were declined by the sandbox.

## 7. What was missing or ambiguous in the input

Stated plainly, because silence here reads as nothing having been wrong:

1. The workspace did not contain the submodules the ask explicitly named. This
   is the central mismatch between what was asked for and what was provided.
2. No acceptance criteria were supplied, so this brief is written against the
   skill's contract rather than against a stated bar.
3. No prior intake, work item, context paths or context excerpts were supplied,
   and the portfolio artifact store is empty — there was no prior work to build
   on, and this brief starts the record for this product.
4. The purpose behind the grounding was not stated (§5.6).
5. `git` subcommands were unavailable in this sandbox, which cost the submodule
   pins and the exact tree listing.

## 8. Suggested next step

Run `scan-workspace` over all eight trees with the submodules populated and
readable, carrying §5.1–§5.6 in as the rows it must settle. If the submodules
cannot be made readable, `scan-workspace` will reproduce exactly the ceiling
recorded in §3 — the meta-repository layer at high confidence, everything below
it unestablished — and the grounding manifest will inherit this brief's gaps
rather than close them.
