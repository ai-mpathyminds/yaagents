# Intake brief — grounding the YAAgents product, meta repository and all submodules

**Artifact type:** `intake-brief` (`application/vnd.wakalix.intake-brief+json; v=1`) — rendered here as a document.
**Produced by:** onboarding-agent, skill `confirm-grounding`.
**Packet:** `01a07191-8327-7b46-a993-bd7b444e3383` · **Skill run:** `01a07191-597f-7013-89f8-3c5fe5ec0fad`
**Product:** `yaagents` · **Repository:** `github.com/ai-mpathyminds/yaagents`
**Branch:** `wakalix/run-01a07191597f` (base `main`) · **Date:** 2026-09-05

> This brief records what was asked and what the workspace already answers. It agrees nothing.
> Grounding reads and does not change: every gap below is a finding for the cycle that follows,
> not an edit made now.

---

## 1. The ask, as it arrived

Verbatim, from the packet's `tail.input.intent`:

> This is a meta repository and there are sub modules, We will ground all including submodules

The packet carried no `work_item_id`, no `intake`, no acceptance criteria, and no context
excerpts or paths. The sentence above is the entire input.

---

## 2. What I read

**From wakalix, by id:**

| Source | What it gave |
|---|---|
| `catalog_read(role="onboarding-agent")` | The role's purpose, its two skills (`confirm-grounding`, `scan-workspace`), and what it produces (`intake-brief`, `grounding-manifest`). |
| `registry_read()` | The 12 registered artifact types, including the definition of `intake-brief` this document claims to be. |
| `artifact_read()` (whole organisation, no filter) | **Empty.** `{"artifacts": []}`. Nothing has been produced for this organisation yet. |
| Packet `prior` | `attempts: []`, `artifacts: []`. This ask has not been dispatched before. |

There were no prior artifacts to read, so nothing below is a re-derivation of existing work.

**From the workspace, by path** (worktree of `wakalix/run-01a07191597f`):

`.gitmodules` · `README.md` · `CONTRIBUTING.md` · `.gitignore` · `.markdownlint.json` ·
`spec/VERSION` · `docs/astro.config.mjs` · `docs/.gitignore` ·
`.github/workflows/supply-chain-audit.yml` ·
directory listings of `spec/`, `schemas/`, `openapi/`, `tests/`, `tools/`, `bin/`, `docker/`,
`examples/`, `docs/`, `.github/`, and each of the seven submodule paths ·
`git ls-files -s` over the seven submodule paths.

---

## 3. What the ask appears to mean

Stated as a reading, not as an agreement. Section 6 lists what a person must confirm.

The requester wants a grounding cycle whose scope is the **whole eight-repository YAAgents
product**, not the meta repository alone. "Ground all including submodules" is read as: the
grounding-manifest that follows must carry rows for each submodule's own repository, not only
for the meta repository that pins them.

---

## 4. Scope, as the workspace declares it

`.gitmodules` declares seven submodules. `.github/workflows/supply-chain-audit.yml` (lines
179–188) enumerates the same seven plus the meta repository as "all 8 repos", which is the
workspace's own name for the full set. The pinned commits below are from `git ls-files -s`
at the branch head.

| Path | Upstream repository | Pinned commit | Working tree here |
|---|---|---|---|
| *(root)* | `ai-mpathyminds/yaagents` | `ebf8a71` (branch head) | populated |
| `gateway/` | `ai-mpathyminds/yaagents-gateway` | `d99fa7226804cea62c62c4a1294173b7f9f9402b` | **empty** |
| `sdk-go/` | `ai-mpathyminds/yaagents-sdk-go` | `27580d90ec20a5dfd411a2d7ff1b83b1ff1af028` | **empty** |
| `sdk-fastapi/` | `ai-mpathyminds/yaagents-sdk-fastapi` | `062b0e9269b30132679f94937740a7a9bb253223` | **empty** |
| `client-python/` | `ai-mpathyminds/yaagents-client-python` | `3c6d180362eda8f8d88c2c7fb8092e1ee2ea0362` | **empty** |
| `client-ts/` | `ai-mpathyminds/yaagents-client-ts` | `7a11174a8ef7a1f5642f09c7b10479fc389702e5` | **empty** |
| `client-go/` | `ai-mpathyminds/yaagents-client-go` | `89a628f5cbc450083c8684670dd9e440dba02a5c` | **empty** |
| `cli/` | `ai-mpathyminds/yaagents-cli` | `e77fb8e84b74c2e82291e29f5976c450b512cbed` | **empty** |

**This is the load-bearing finding of this intake.** The gitlinks are pinned, but the seven
submodule working trees are not checked out in this worktree — every one of the seven
directories is empty. See §7.

---

## 5. What inspection settled

Every row names its source. `inspected` means a file in this workspace says it; nothing below
is assumed, and nothing below was asked, because nothing has been asked yet.

| Dimension | What the workspace says | Source |
|---|---|---|
| Product identity | YAAgents — a gateway plus SDKs that expose an AI agent behind a governed REST API. | inspected — `README.md` §"What problem does yaagents solve?" |
| Shape | Meta repository holding the normative spec, schemas, OpenAPI components, docs site, and examples; seven implementation repositories attached as submodules. | inspected — `README.md` §"Repository layout", `.gitmodules` |
| Normative contract | Agentic REST Response Profile v0.3, `spec/agentic-rest-profile.md`; ten response types with fixed HTTP status and content type. | inspected — `spec/VERSION` (`0.3`), `README.md` §"Response Profile" |
| Schema versions carried | `schemas/v0.1`, `schemas/v0.2`, `schemas/v0.3`, six schemas each. | inspected — `schemas/` listing |
| Release state | Published artifacts v0.3.0; v0.4 in progress; v0.4.0 targeted at "PI5-yaa" alongside a Helm chart and a full publish wave. | inspected — `README.md` §banner and §License |
| Languages and floors | Go 1.22+, Python 3.11+, Node 20+ (pnpm), TypeScript. | inspected — `README.md` badges, `CONTRIBUTING.md` §"Dev setup" |
| Release mechanics | Tag-driven. A major bump must tag all eight components at the same version in one cycle, in order `gateway` → `sdk-*` → `client-*` → `cli` → meta-repo, each publish succeeding before the next. | inspected — `CONTRIBUTING.md` §"Release process" |
| Contribution rules | Conventional Commits with component scope; branch names `feature/`, `fix/`, `docs/`; DCO `Signed-off-by:` required on every commit; maintainer commits additionally carry `Agent:` and `WI:` trailers from internal tooling. | inspected — `CONTRIBUTING.md` §"PR process" |
| Quality gates | Go: `golangci-lint`. Python: `ruff check`, `ruff format --check`, `mypy --strict`. TS: `eslint`, `prettier`. Docs: clean `pnpm run build`. | inspected — `CONTRIBUTING.md` §"Code style" |
| CI present at meta level | 11 workflows: `ci`, `bench`, `pages`, `gateway-image`, `npm-publish`, `pypi-publish`, `sdk-go-ci`, `sdk-go-smoke`, `supply-chain-audit`, `verify-go-module`. | inspected — `.github/workflows/` listing |
| Cross-repo audit already exists | `supply-chain-audit.yml` already clones and scans all eight repositories as one unit. A grounding cycle is not the first thing to treat the eight as a set. | inspected — `.github/workflows/supply-chain-audit.yml` lines 179–217 |
| Docs surface | Astro Starlight site under `docs/`, published at `https://ai-mpathyminds.github.io/yaagents/` with base path `/yaagents`. Only `docs/src/content/docs/` is published. | inspected — `docs/astro.config.mjs` |
| Examples | Seven runnable examples: `store`, `store-go`, `campaign-api`, `campaign-api-go`, `agent-graph-ecom`, `customer-support-triage`, `financial-risk-screening`, `llm-gateway`. | inspected — `examples/` listing |
| Licensing | Apache 2.0. v0.1.x packages remain under the previous source-available terms, non-retroactively. `CONTRIBUTING.md` carries a verbatim legal disclaimer citing ADR PI2-yaa-0003 §3. | inspected — `LICENSE` reference in `README.md` §License, `CONTRIBUTING.md` §disclaimer |
| A hard content constraint | `supply-chain-audit.yml` check C-1 **fails the build** on the literal string naming the superseded source-available licence, in any `*.go`, `*.py`, `*.ts`, `*.md`, or `*.toml` file across all eight repositories. Any document a grounding cycle writes is in scope for it. | inspected — `.github/workflows/supply-chain-audit.yml` lines 190–217 |
| Prior wakalix record | None. No artifacts, no cycles, no prior attempts. This is the product's first contact with the engine. | inspected — `artifact_read()`, packet `prior` |

---

## 6. What a person must settle

Named individually, because a manifest silent about a dimension reads as one with no answer.
None of these was resolved silently.

**6.1 — One product or eight?**
The ask says "ground all including submodules". That is consistent with two different shapes,
and the workspace does not choose between them: (a) one product `yaagents` whose grounding
manifest carries per-component rows for eight repositories; or (b) eight products, each
grounded on its own, with `yaagents` as an umbrella. The packet's `tail.product` is the single
value `yaagents` and `delivery.repository` is the single meta repository, which leans toward
(a) — but that is the dispatcher's framing, not the requester's stated intent. **Unclear.**

**6.2 — Are the submodules in scope as sources, or as targets?**
"Ground all" could mean *read* all eight to describe the product, or *establish a grounding
cycle for* each of the eight so that later work can be dispatched into them. These imply
different delivery configurations downstream. **Unclear.**

**6.3 — What is the canonical dimension set for a `grounding-manifest`?**
`catalog_read` says `scan-workspace` "reports a confidence and a source per dimension" but
does not enumerate the dimensions, and no `grounding-manifest` exists in the record to read
one off. §5 above is therefore organised by what the workspace actually volunteered, not by a
schema. If wakalix defines a fixed dimension set, §5 must be re-cut against it before it can
be called complete. **Unclear — and I did not invent a list to fill the gap.**

**6.4 — Which submodule commits should be grounded?**
The pinned commits in §4 are what the meta repository points at as of `ebf8a71`. Each
submodule also has its own moving default branch. Grounding the pins and grounding the
branch heads are different answers, and they will diverge — `README.md` says published
artifacts are v0.3.0 while v0.4 work is in progress. **Unclear.**

**6.5 — Whose statement of intent is this?**
The input is one sentence with no author, no deadline, and no stated outcome the requester
wants grounding to enable. Nothing in the packet says why grounding is being done now.
**Unclear** — and this is the field most likely to change what a good manifest looks like.

---

## 7. What blocks the cycle this brief opens

**The submodule working trees are empty in this workspace.** All seven gitlinks are pinned
(§4), but no submodule content is checked out — `gateway/`, `sdk-go/`, `sdk-fastapi/`,
`client-python/`, `client-ts/`, `client-go/`, and `cli/` each contain zero entries. `.git` is a
file, not a directory: this is a `git worktree`, and `git worktree add` does not populate
submodules.

A `scan-workspace` run dispatched into this worktree as it stands can inspect the meta
repository and **nothing else**. It would report the seven components as unknown and would be
right to.

Two ways forward exist; choosing between them is a decision for the plan, and this brief
makes neither change:

- Populate the submodules in the grounding workspace before scanning — `CONTRIBUTING.md`
  §"Dev setup" gives the intended clone form, `git clone --recurse-submodules`; the worktree
  equivalent is a submodule init/update step, which requires network access to seven private
  or public GitHub repositories and credentials this session was not given.
- Dispatch grounding per repository, one packet per component, and compose the eight results
  into a single manifest at the end.

Related constraint, recorded so the plan does not rediscover it: **check C-1 of
`supply-chain-audit.yml` fails the build** on the literal name of the superseded
source-available licence anywhere in `*.md` across all eight repositories. Documents written
by a grounding cycle are inside that scan's scope.

---

## 8. Why this file is here and not under `docs/src/content/docs/`

`docs/` in this repository is the Astro Starlight package for the **public** documentation
site. Astro publishes only `docs/src/content/docs/`; a file at `docs/grounding/` is version
controlled, is not gitignored (`docs/.gitignore` ignores only `dist/`, `node_modules/`,
`.astro/`), and is not rendered to `ai-mpathyminds.github.io/yaagents`. An internal intake
brief should not appear on the product's public docs site, so it sits outside `src/`.

---

## 9. What I did not do

- **I did not change the workspace.** No submodule was initialised, no configuration touched,
  no defect fixed. Grounding reads. Everything in §6 and §7 is a finding for the plan.
- **I did not inspect any submodule's contents**, because none is present (§7). Every claim in
  §5 is sourced from the meta repository. Where `README.md` or `CONTRIBUTING.md` describes a
  submodule, this brief reports it as *the meta repository's description of that component*,
  verified against nothing inside the component itself.
- **I did not produce a `grounding-manifest`.** That is `scan-workspace`'s output, and it
  cannot honestly be produced while seven of eight repositories are unreadable.
- **I did not resolve the five open items in §6**, and specifically did not invent a
  dimension taxonomy (§6.3) to make §5 look complete.
- **I did not read the git history.** Beyond the branch head named in the packet and the
  `git ls-files -s` gitlinks in §4, no log, tag, or diff was examined; several plumbing
  commands were not permitted in this session. Statements about release cadence and the
  PI-series come from `README.md` and `CONTRIBUTING.md` prose, not from tags.

---

## 10. What was missing or ambiguous in what I was given

- The packet's `context.paths` and `context.excerpts` were both **empty**. Scope was
  reconstructed from `.gitmodules` and the audit workflow, not handed over.
- `tail.acceptance_criteria` was **empty**. There is no stated test for whether this intake is
  good enough, so §6 is my judgement of what a person must settle, not a checklist I was given.
- `tail.work_item_id` and `tail.intake` were **null**. This brief is not attached to a work
  item and does not extend a prior intake.
- The input intent is one sentence. It names a shape ("meta repository", "sub modules") and an
  action ("ground all"), and nothing about purpose, audience, depth, or deadline. §6.5 is the
  consequence.
- The organisation's artifact record is **empty**, so there is no precedent — no prior
  intake-brief, no prior manifest — to match this document's shape or depth against.
