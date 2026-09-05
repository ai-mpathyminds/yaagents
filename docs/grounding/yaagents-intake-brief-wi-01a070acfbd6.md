# YAAgents — Grounding Intake Brief (cycle `wi-01a070acfbd6`)

| Field | Value |
|---|---|
| Product | `yaagents` |
| Repository | `github.com/ai-mpathyminds/yaagents` |
| Produced by | onboarding-agent · skill `confirm-grounding` |
| Artifact type | `intake-brief` (`application/vnd.wakalix.intake-brief+json; v=1`) |
| Work packet | `01a070ac-ffb7-7b2c-9d5b-eaddf8152dda` |
| Work item | `01a070ac-fbd6-7cf3-ab3e-dabe22983923` |
| Branch inspected | `wakalix/wi-01a070acfbd6` (base `main`), at commit `ebf8a71` |
| Date | 2026-09-05 |

This is the intake this grounding cycle opens from. It changes nothing in the
workspace: every gap below is a finding for the plan, not an edit made now.

**It is not the first intake for this product.** A prior `intake-brief` exists in
the record and is cited throughout as **artifact
`01a0709d-8c9b-74f2-be10-7c961fcb1112`** (v1, `docs/grounding/yaagents-intake-brief.md`,
commit `f58412c`). This brief does not re-derive what that one established. It
re-verifies the facts this cycle depends on, carries the rest forward by
reference in §5, and spends its effort on the rows that artifact left open —
three of which are now closed in §4.

---

## 1. The ask, as given

The operator's intent, verbatim from the work packet:

> Open this grounding cycle: run confirm-grounding

Nothing else was supplied. `tail.input`, `tail.intake` and `tail.skill_run_id`
were all `null`; `tail.acceptance_criteria` was an empty list; `context.paths`
and `context.excerpts` were both empty; and `prior.artifacts` and
`prior.attempts` were both empty.

The intent names an action and not a subject. It is read as: **produce the
intake for a new grounding cycle over the `yaagents` product, on this branch.**
What the grounding is *for* is still not stated — see §6.5, which is the same
row artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` §5.6 raised and which the
operator has not yet answered.

## 2. What was read

**The record.** The artifact store was listed with `artifact_read` (no
arguments) and returned exactly one artifact:
`01a0709d-8c9b-74f2-be10-7c961fcb1112`, type `intake-brief`, version 1. It was
then read in full by id. It is the only artifact this organisation holds for
this product, and it is the only one cited in this brief. Note that this
packet's own `prior.artifacts` was **empty** — the prior brief was found by
querying the record, not by being handed over. The agent catalog was read for
role `onboarding-agent`, and the artifact-type registry was read to confirm what
an `intake-brief` is.

**Files read in full in this workspace, this run:**

| Path | What it was read for |
|---|---|
| `.gitmodules` | Re-verify the seven submodule paths and remotes |
| `spec/VERSION` | Re-verify the Profile version |
| `.github/workflows/gateway-image.yml` | §4.1 — unread by the prior cycle |
| `.github/workflows/pypi-publish.yml` | §4.1 — unread by the prior cycle |
| `.github/workflows/npm-publish.yml` | §4.1 — unread by the prior cycle |
| `.github/workflows/supply-chain-audit.yml` | §4.1, §4.3 — unread by the prior cycle |
| `bin/yaagents-public-mirror-verify.sh` | Establish the constraints this document itself must satisfy (§7) |

**Targeted searches run:** every `continue-on-error: true` in
`.github/workflows/ci.yml` with context; every `actions/checkout@v4` and
`submodules:` line across all ten workflows; `release.yml` and `scrub-verify.yml`
across all Markdown; `store-go` in `README.md`; the scanning claims in
`SECURITY.md`; the v0.4 security-hardening block in
`docs/src/content/docs/roadmap.mdx`. Directory listings taken: repository root,
`examples/`, `.github/workflows/`, and each of the seven submodule paths.

**Not re-read this run**, because artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112`
already records them and nothing in this run contradicted it: `README.md`,
`CONTRIBUTING.md`, `SECURITY.md` (beyond the four lines searched), `ci.yml`
(beyond the searches above), the docs-site config and content pages, and
`spec/examples/`.

## 3. Dimension table

Source is one of **inspected** (read in this workspace, this run), **carried
forward** (established by artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` and
not contradicted here), **asked** (supplied by the operator), or **assumed**.
There are no `assumed` rows. Where nothing settled a dimension it is marked
unsettled and carried to §6 rather than resolved silently.

| # | Dimension | Source | Confidence | Settled? |
|---|---|---|---|---|
| 1 | Product identity, purpose, non-goals | carried forward | high | yes — §5 |
| 2 | Repository topology | inspected (re-verified) | high | yes — §5 |
| 3 | Public contract (Profile / schemas / OpenAPI) | inspected (version) + carried forward | high | yes — §5 |
| 4 | Languages and toolchains | carried forward | medium | partly — §6.1 |
| 5 | Build and test commands | carried forward | medium | partly — §6.1 |
| 6 | CI gates and current health | inspected (re-verified) | high | yes — §5 |
| 7 | Release and distribution mechanics | **inspected** | **high** | **yes — §4.1** |
| 8 | Supply-chain posture (signing, SBOM, scanning) | **inspected** | **high** | **yes — §4.1** |
| 9 | Security posture (policy, SLA, scope) | carried forward | high | yes — §5 |
| 10 | Governance and contribution | carried forward | high | yes — §5 |
| 11 | Docs site | carried forward | high | yes — §5 |
| 12 | Examples and reference stacks | inspected (re-verified) | high | yes — §5 |
| 13 | **Submodule internals** | — | **none** | **no — §6.1** |
| 14 | **Release-wave executability** | inspected | medium | **no — §4.2** |
| 15 | **Runtime environments / deployment targets** | — | low | no — §6.2 |
| 16 | **Ownership and on-call** | inspected (absence confirmed) | high | **no — §6.3** |
| 17 | **Goal of this grounding** | — | none | **no — §6.5** |

## 4. What this cycle settled that the prior one could not

### 4.1 Supply-chain posture — the prior cycle's §5.5 is closed

Artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` §5.5 recorded a genuine
contradiction it could not resolve: `SECURITY.md` and `CONTRIBUTING.md` describe
Cosign signing, Syft SBOM and npm provenance as **shipped**, while
`docs/src/content/docs/roadmap.mdx:59-62` lists all three as **planned for
v0.4**. It named the four workflows that would settle it and did not read them.

They have now been read. **The documentation is right and the roadmap page is
stale.** All three are implemented in the workspace today:

| Claim | Where it is implemented | Roadmap says |
|---|---|---|
| Cosign keyless OIDC image signing | `gateway-image.yml:110-116` (`cosign-installer`, `cosign sign --yes` by digest) | "planned" (`roadmap.mdx:60`) |
| SBOM, SPDX 2.3 JSON, via Syft | `gateway-image.yml:161-169` (`anchore/sbom-action`), attached to the image at `:175-180` (`cosign attach sbom`) and to the GitHub Release at `:186-191` | "planned" (`roadmap.mdx:61`) |
| Signed npm provenance attestation | `npm-publish.yml:141-144` (`npm publish --provenance` under `id-token: write`) | "planned" (`roadmap.mdx:62`) |

Two further facts the prior cycle flagged as unverifiable are now settled:

- **`trivy` exists, and it is not in `ci.yml`.** `SECURITY.md:111` claims a
  Trivy scan of the gateway image; the prior brief could not find it. It is
  `gateway-image.yml:140-155` — a scan by digest at `CRITICAL,HIGH` with
  `exit-code: 1` and `ignore-unfixed: true`, with SARIF uploaded to the GitHub
  Security tab. It is a **blocking** publish-time gate.
- **The `npm audit` vs `pnpm audit` discrepancy is confirmed and is in two
  places.** `SECURITY.md:113` says `npm audit`; the repository runs
  `pnpm audit --audit-level=high` both in CI and at publish time
  (`npm-publish.yml:73-74`). This is a documentation error, not a missing gate.

**Publishing is token-free by design.** PyPI uses Trusted Publisher OIDC with
`id-token: write` and no API token (`pypi-publish.yml:99-100`, `:278-279`);
GHCR uses `GITHUB_TOKEN` (`gateway-image.yml:66-72`); npm uses a granular access
token deliberately named `NODE_AUTH_TOKEN` rather than `NPM_TOKEN`
(`npm-publish.yml:11-12`). `supply-chain-audit.yml:85-144` enforces exactly this
as "Check A", failing if `secrets.REGISTRY_PASSWORD`, `secrets.PYPI_API_TOKEN`,
`secrets.NPM_TOKEN` or a `--password.*secrets` pattern appears in any workflow
across all eight repositories.

**The PyPI release path has a staged gate.** `pypi-publish.yml` builds all three
wheels, publishes to TestPyPI, then runs `verify-testpypi`
(`pypi-publish.yml:166-268`) — installing from TestPyPI and asserting the three
packages agree on version, all report `Apache-2.0`, all declare
`Supports-YAAgents-Profile: v0.3`, the `[otel]` extra resolves, and the
`yaagents` CLI entrypoint runs. Only then do the production publishes fire.

**The release wave is already authored for v0.4.0.** The header comments of all
three publish workflows are scoped to a 0.4.0 wave (`gateway-image.yml:2`,
`pypi-publish.yml:2`, `npm-publish.yml:2`), and `pypi-publish.yml:232-256`
smoke-tests `AuditEmitter` / `AuditEvent` / `NoopEmitter` symbols it expects
`yaagents-fastapi` 0.4.0 to carry. `roadmap.mdx:47` still frames v0.4 as
"Planned". **v0.4 is further along in the release machinery than the public
roadmap describes** — which is the same staleness as the table above, and the
two should be treated as one finding.

### 4.2 A new finding: the publish workflows do not check out the submodules

This was not visible to the prior cycle, which did not read these files.

Every workflow that builds or scans the product checks out with
`submodules: recursive` — `ci.yml` at thirteen call sites (`:48-50`, `:147-149`,
`:212-214`, `:263-265`, `:316-318`, `:369-371`, `:423-425`, `:469-471`,
`:526-528`, `:576-578`, `:610-612`, `:633-635`, `:668-670`), plus `bench.yml:60-62`,
`sdk-go-ci.yml:40-42` and `sdk-go-smoke.yml:54-56`.

**The three publish workflows do not.** `gateway-image.yml:41-42`,
`pypi-publish.yml:31-32` and `npm-publish.yml:43-44` each use a bare
`actions/checkout@v4` with no `with:` block — yet each builds exclusively from
submodule paths: `context: gateway` (`gateway-image.yml:93`), `hatch build` in
`sdk-fastapi`, `client-python` and `cli` (`pypi-publish.yml:60-70`), and
`working-directory: client-ts` (`npm-publish.yml:38-39`). A default checkout
leaves those directories empty.

Two consequences follow from reading, and they differ in severity:

1. **The builds would fail.** `hatch build` in an empty directory, a Docker
   build context with no source, and `pnpm install` with no manifest all error.
   This is loud, not silent.
2. **The pre-publish licence gates would pass vacuously.** Each is shaped as
   `if grep -rn <pattern> <submodule-path>; then fail; fi`
   (`gateway-image.yml:49-57`, `pypi-publish.yml:38-46`,
   `npm-publish.yml:81-91`). Over an empty directory `grep` matches nothing,
   the branch is not taken, and the step prints `PASS`. A gate that reports
   PASS because it had nothing to look at is the more dangerous half of this.

**This is stated as a static reading, not an observed failure.** No workflow run
was inspected — the run history is outside this workspace. It is possible the
wave has only ever been dispatched in a context where the directories were
populated by other means. That is exactly why it is a finding for the plan and
not a conclusion: **someone should check whether the v0.4.0 release wave has
ever completed end to end, and if it has, what populated those paths.**

### 4.3 A second documentation-vs-reality gap: a workflow named but absent

The prior cycle recorded that `CONTRIBUTING.md:154` names
`.github/workflows/release.yml`, which does not exist. That is re-confirmed: a
search across all Markdown finds that one line and no such file.

A second instance of the same shape was found this run.
`bin/yaagents-public-mirror-verify.sh:22-24` documents its own CI integration as
`.github/workflows/scrub-verify.yml`, "triggered on push to meta-repo main",
passing the eight checked-out paths as arguments. **No `scrub-verify.yml`
exists.** The integration is real but lives elsewhere: it is "Check B" of
`supply-chain-audit.yml:151-169`, which does precisely what the script's header
describes. So the mechanism is present and the pointer is stale — the same
correction shape as `release.yml`, and both are cheap fixes for a plan to pick up.

Note also that `supply-chain-audit.yml:33-79` assembles the eight repositories by
**eight explicit `actions/checkout` steps against named remotes**, not via
submodules. It is therefore the one workflow whose whole-product view does not
depend on the submodule pins at all.

### 4.4 Re-verified, unchanged since the prior cycle

Checked directly this run and found identical to artifact
`01a0709d-8c9b-74f2-be10-7c961fcb1112`:

- **Same tree.** This branch's tip is `ebf8a71`, the same commit that brief
  inspected. No workspace content has changed between the two cycles.
- **Seven submodules**, same paths and same `ai-mpathyminds` remotes
  (`.gitmodules:1-21`).
- **Profile v0.3** (`spec/VERSION`).
- **Six advisory CI gates**, all still `continue-on-error: true` with their
  dated rationale comments intact: `golangci-lint` (`ci.yml:67`),
  `govulncheck-gateway` (`:88`), `govulncheck-plugins` (`:97`), `go test`
  (`:110`), the ≥80% coverage gate (`:117`), and `govulncheck-client-go`
  (`:172`). The named remediation is unchanged — 16 fixture YAMLs to re-indent,
  `TestBUMP3_SpecVersion` to expect `0.3`, a linter config to re-author, and
  the Go stdlib CVEs to clear.
- **Ten workflows**, no more and no fewer.
- **`examples/store-go/` does not exist.** `README.md:62` links it as the Go
  run-it path. The seven directories under `examples/` are `agent-graph-ecom`,
  `campaign-api`, `campaign-api-go`, `customer-support-triage`,
  `financial-risk-screening`, `llm-gateway`, `store`. The broken link stands.
- **No `CODEOWNERS`** anywhere in the repository.

## 5. Carried forward without re-derivation

The following are established by artifact
`01a0709d-8c9b-74f2-be10-7c961fcb1112` and are **not** restated here, because
the tree is byte-identical at the same commit and re-deriving them would add a
second copy that can only drift from the first: product identity and the five
explicit non-goals (§4.1 there); the ten response outcomes and their media types
(§4.3); the schema, OpenAPI, fixture and golden-file inventory (§4.3); the five
first-party gateway plugins (§4.3); declared toolchains and per-component build
and test commands (§4.4); the blocking CI gates worth preserving (§4.5); the
tag-driven eight-component release ordering (§4.6); the security policy, SLA and
scope (§4.7); DCO, Conventional Commits and the community-plugin path (§4.8);
the Astro Starlight docs site and its `/yaagents/` base prefix (§4.9); and the
seven example stacks (§4.10).

A reader who needs any of those should read that artifact. Where this cycle
touched one of them, §4.4 says so explicitly.

## 6. What is still not settled, by name

### 6.1 Submodule internals — still not inspected, for the same reason

The seven submodule working trees are **empty in this workspace too**.
`gateway/`, `sdk-go/`, `sdk-fastapi/`, `client-go/`, `client-python/`,
`client-ts/` and `cli/` each contain no files; the submodules are recorded as
gitlinks and are not checked out. `git submodule status` was declined by the
sandbox in this run, as it was in the last, so **the pinned commits still could
not be captured**.

This is the same ceiling artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` §5.1
recorded, and nothing about it has improved. Every statement about a submodule
in either brief remains **meta-repository evidence about that submodule** — its
remote, its published package name, the commands documented for it, the CI job
that builds it — and never a reading of its source. Still unestablished: each
submodule's module manifest and dependency set, its source layout, its test
suite and current pass rate, its changelog, its own workflows, and its pinned
commit.

The prior brief's open question about where the plugin interface actually lives
— `gateway/plugins/plugin.go` per `CONTRIBUTING.md:189-192` versus
`gateway/internal/plugins/` per `ci.yml` — is likewise still open, and only the
`gateway` submodule can answer it.

**Resolution, unchanged:** populate the submodules
(`git submodule update --init --recursive`) in a workspace the agent can read,
then run `scan-workspace` over all eight trees.

### 6.2 Runtime environments and deployment targets

No environment the product actually runs in is named anywhere: no staging or
production target, no cluster, no account, no promotion path. The Helm chart is
listed as planned for v0.4 (`roadmap.mdx:56-57`), which implies no supported
Kubernetes path exists today without describing what is deployed instead. The
one live environment evidenced remains the GitHub Pages docs site. Note this is
now in mild tension with §4.1: the *publishing* side of v0.4 is substantially
built, while the *deployment* side still shows only a plan.

### 6.3 Ownership and on-call

Confirmed absent this run rather than merely unfound: there is no `CODEOWNERS`
file anywhere in the repository. There is no maintainer roster and no rotation.
`CONTRIBUTING.md:166` says only that "maintainers with write access to the
`ai-mpathyminds` organization" can publish, without naming them. The GitHub
Actions environments `pypi`, `testpypi` and `npm` referenced by the publish
workflows imply protection rules and therefore approvers configured outside the
repository — **who those approvers are is not readable from here**, and it is a
concrete question a person can answer quickly.

### 6.4 Two small carried-forward discrepancies, still open

- **Node version.** `CONTRIBUTING.md:63,66` says Node 20+; `ci.yml:530-532` and
  `npm-publish.yml:47-51` both pin Node 22.
- **`release.yml`.** Whether it was renamed, split, or never existed (§4.3).

### 6.5 What this grounding is for — asked once, still unanswered

The intent says to open a grounding cycle. It does not say what the grounding is
in service of. Artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` §5.6 raised this
as "the single most useful thing an operator could add before the next step",
and this packet supplied no answer, no acceptance criteria and no context.

It does not block this brief, and it has not been resolved by assumption. But it
now has a sharper edge than it did, because §4 has produced candidate purposes
that pull in different directions:

- **the v0.4.0 release wave** — largely built (§4.1), with one unverified
  blocker (§4.2);
- **the six advisory CI gates** — the largest block of named, undisputed
  outstanding work (§4.4);
- **documentation truth** — four stale pointers now identified, each cheap
  (§4.1, §4.3, §4.4, §6.4);
- **submodule-depth grounding** — what the previous cycle's ask actually
  requested, and what §6.1 still blocks.

A one-line answer from the operator would let `scan-workspace` weight its pass.
Without one it will treat all four as equal, which is the least useful of the
available choices.

## 7. What this run did not do, and why

- **It did not read any submodule source.** The trees are empty here; the
  populated checkout is outside this run's read scope. §6.1 is a limit of the
  workspace, not a choice.
- **It did not run any build, test, lint or publish command**, and did not
  inspect any workflow run history. Everything in §4.1 and §4.2 is read from
  workflow definitions, not observed executing. The §4.2 finding in particular
  is a static reading and is labelled as one.
- **It did not re-derive artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112`.**
  §5 carries its findings by reference. Where this run re-checked one of them it
  says so in §4.4; where it did not, §2 lists the file as not re-read.
- **It did not change anything.** Grounding reads and does not change. The six
  advisory gates, the missing `submodules: recursive`, the broken `store-go`
  link, the two absent-workflow pointers, the stale roadmap block and the
  `npm audit` wording are all written down as findings. Fixing them is a plan's
  business, and several of them are one-line fixes precisely because nobody has
  been asked to make them yet.
- **It did not produce a grounding manifest.** That is `scan-workspace`, and it
  should run after §6.1 is resolved.
- **It did not read `spec/agentic-rest-profile.md`, the schema files,
  `openapi/`, `verify-go-module.yml`, `bench.yml`, `pages.yml`, `sdk-go-ci.yml`,
  `sdk-go-smoke.yml`, or the individual example stacks.** These were enumerated,
  not opened. If the normative spec and the README ever disagree, the spec wins.
- **It did not capture the submodule pins.** The `git` invocation that reports
  them was declined by the sandbox, in both cycles now.

This document is written to pass `bin/yaagents-public-mirror-verify.sh`, which
greps Markdown for five internal-path markers and fails on a hit. Internal
identifiers appearing in the workflow headers cited above are therefore
referenced by path and line rather than quoted. It sits at `docs/grounding/`,
outside `docs/src/`, so it is not part of the published docs site
(`docs/src/content/config.ts` defines the only content collection).

## 8. What was missing or ambiguous in the input

Stated plainly, because silence here reads as nothing having been wrong:

1. **The packet's `prior.artifacts` was empty although a prior artifact exists.**
   Artifact `01a0709d-8c9b-74f2-be10-7c961fcb1112` was found only by querying
   the record directly. Had this run trusted the packet, it would have produced a
   duplicate of a brief that already existed. This is the most consequential gap
   in the input and it is worth fixing upstream.
2. **The intent named an action, not a subject.** "Run confirm-grounding" does
   not say over what, or why. Scope was inferred from `tail.product`.
3. **No acceptance criteria** were supplied, so this brief is written against the
   skill's contract rather than a stated bar.
4. **The purpose of the grounding was not supplied** — asked in the prior cycle,
   unanswered here (§6.5).
5. **The workspace still does not contain the submodules** (§6.1), and `git`
   subcommands remain unavailable in the sandbox, which cost the pins a second
   time.

## 9. Suggested next step

Answer §6.5 in one line, then run `scan-workspace` over all eight trees with the
submodules populated and readable, carrying §6.1–§6.4 in as the rows it must
settle and §4.2 in as the first thing to verify against a real workflow run.

If the submodules cannot be made readable, `scan-workspace` will reproduce
exactly the ceiling in §3 for a third time — the meta-repository layer at high
confidence, everything below it unestablished — and the grounding manifest will
inherit these gaps rather than close them. Two cycles have now hit the same
wall; the third will too unless the workspace changes.
