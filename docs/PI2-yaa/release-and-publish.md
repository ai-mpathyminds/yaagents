# PI2-yaa — Component: Re-publish at 0.2.0 + repo scaffolding updates (`release-and-publish`)

> **SCOPE AMENDMENT 2026-06-02 (user-direct, chief-architect-proxied).**
> Prod public-registry publish stages for REL-1 (PyPI ×3) / REL-2 (npm) /
> REL-3 (GHCR public) / REL-4 (Go module tag) **AND** the un-park step in
> REL-PRECHECK and PI-GATE acceptance criteria #1-4 (the
> `pip install`/`npm install`/`docker pull`/`go get` from prod public
> registries) are **DEFERRED to PI3-yaa** per the public-launch scope
> consolidation. PI2-yaa source repo stays PRIVATE through close;
> PI3-yaa is the public-launch PI (sdk-go + repo restructure + portfolio
> scrub + Pages site at yaagents.dev + prod publish + repo public-flip).
> See `yaagents/system-refs/yaagents-v0.3.seed.md` + `portfolio/INTAKE/PI3-yaa-intake.md`.
>
> **What PI2-yaa REL-* WIs still do**: build the artifacts; publish to
> TestPyPI (REL-1 sub-step) as a packaging-validation gate; publish to
> GHCR with **private** package visibility under `ai-mpathyminds` org
> (REL-3 sub-step; verifies multi-arch build + SBOM attachment); skip
> REL-4 entirely (Go module proxy publish requires public repo).
> REL-PRECHECK runs the **5 automated checks + manual checklist**
> verifying account-side readiness for PI3-yaa launch, but DOES NOT
> un-park the publish workflows (workflows remain in `workflow_dispatch:
> {}` only mode through PI2-yaa close).
>
> **PI-GATE PI2-yaa acceptance amended**: steps 1-4 (prod public-registry
> install from PyPI/npm/Docker Hub/Go proxy) become STAGED in PI3-yaa
> close gate; steps 5-7 (compose demos + license-clean scan) still
> execute as PI2-yaa close criteria.

Owner lane: **platform-engineer** (supply-chain / NFR / publishing
pipelines). Sprints 1 + 5. ADRs: PI2-yaa-0003 (Apache 2.0 license
metadata everywhere), PI1-yaa-0005 (OIDC trusted publishing carries
forward unchanged), PI1-yaa-0003 (Hatch + dual ESM/CJS carries forward).

> **Library gate:** REL-S carries `library_ref: ADR PI2-yaa-0003`.
> REL-1/2/3/4 carry `library_ref: ADR PI1-yaa-0005 + ADR PI2-yaa-0003`
> (OIDC trusted publishing unchanged from PI1-yaa; license-metadata
> field changes per PI2-yaa-0003). PI-GATE is a conformance gate, not
> a library-consuming WI — `library_justify: PI gate; no library import`.

The PRD describes these as **re-publish** WIs: the pipelines from PI1-yaa
REL-3/4/5/6 already exist; PI2-yaa amends each to publish the 0.2.0
artefacts with Apache 2.0 license metadata. The OIDC trust relationships
(PyPI Trusted Publisher, npm Provenance, GHCR OIDC) carry forward as-is —
no new IAM / org-config work.

---

### WI-2yaa.REL-S: README license-badge swap + CONTRIBUTING update [DONE] — Sprint 1
service: yaagents/(root)
parent_feature: F-LICENSE
brief: Update `README.md` license badge from "source-available / fair-code"
to Apache 2.0 badge (`![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)`).
Update README hero text per PRD §8.3 verbatim. `CONTRIBUTING.md` retains
the `legal-review-pending` banner verbatim per ADR PI2-yaa-0003
§Consequences (removal is PC-6 non-engineering checklist item, NOT this
WI); the CLA placeholder remains until counsel sign-off. Issue templates
(`.github/ISSUE_TEMPLATE/`) update wording from "source-available" to
"Apache 2.0".
acceptance:
- README badge URL is Apache-2.0; hero text matches PRD §8.3
- `grep -nE "open source|OSI" README.md` still returns 0 hits (Apache 2.0 IS OSI; but the directive is to avoid claims like "the most open license" — wording stays factual)
- `CONTRIBUTING.md` legal-review-pending banner unchanged
- Issue template wording updated
- Sprint-1 timing: lands BEFORE LIC-2 sweep so the new README example snippets are part of the sweep
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-1]

### WI-2yaa.REL-PRECHECK: Publish-account & workflow-park readiness audit [BLOCKED] — Sprint 5 (gates REL-1 + REL-2)
> blocked by: (1) operator — create GH environments `pypi`/`testpypi`/`npm` on `ai-mpathyminds/yaagents` (Settings → Environments, tag-ref protection `v*.*.*`); (2) operator — add `NODE_AUTH_TOKEN` Granular Access Token secret (Settings → Secrets → Actions); (3) scope-amendment-2026-06-02 — workflow un-park deferred to PI3-yaa (check #5 PARKED intentionally). PyPI names FREE ✓, npm FREE ✓.
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: Before REL-1 (PyPI ×3) and REL-2 (npm) fire in Sprint 5, run a
5-item readiness audit that converts the previously-undocumented
"external setup carries forward from PI1-yaa" assumption into a tracked
gate. The PI1-yaa-era publish setup was never re-verified after the
PI2-yaa scope-amendment (2026-05-30 added Apache 2.0 license flip +
cross-lane ai-gateway absorption); a Sprint-5 dispatch of REL-1/REL-2
without this gate risks hitting "Trusted Publisher not configured" /
"NODE_AUTH_TOKEN missing" failures only at publish time. Run
`bin/pi2-yaa-publish-precheck.sh` (platform-engineer authors at A-4
dispatch; chief-architect committed at A-6 amendment) — exits 0 iff all
5 items pass.

The 5 items:

1. **PyPI package name status** (3 packages: `yaagents-fastapi`,
   `yaagents-client`, `yaagents-cli`) — HTTP GET `pypi.org/pypi/<name>/json`
   to determine whether each name is FREE (first publish will claim it),
   CLAIMED-BY-YOU (Trusted Publisher must be configured on owning
   account), or CLAIMED-BY-OTHER (name-squatting; escalate before
   proceeding). Same check repeated for `test.pypi.org` (TestPyPI is a
   separate registry; the gate WI also requires TestPyPI publish to
   succeed before prod stage).
2. **npm package + scope status** — HTTP GET
   `registry.npmjs.org/@aimpathyminds/yaagents-client`. Verify
   `@aimpathyminds` scope is owned by an organisation account, NOT a
   personal account (bus-factor risk — captured in `risks.npm_scope_org`
   addendum to planning runbook context). Operator runs `npm org ls
   aimpathyminds` locally to confirm.
3. **GitHub Environments configured** — `pypi`, `testpypi`, `npm` all
   exist on `ai-mpathyminds/yaagents` repo with tag-ref protection
   (deployments restricted to refs matching `v*.*.*`). Verified via
   `gh api repos/ai-mpathyminds/yaagents/environments/<env>`.
4. **GitHub repo secret `NODE_AUTH_TOKEN` present** — verified via
   `gh secret list --repo ai-mpathyminds/yaagents | grep NODE_AUTH_TOKEN`.
   The script CANNOT verify the token's scope or expiration (npm doesn't
   expose either via API); operator confirms manually that the token is
   a Granular Access Token scoped to ONLY
   `@aimpathyminds/yaagents-client` with `publish` permission, expiration
   ≤ 90 days. Operator MUST also verify `NPM_TOKEN` (the misnamed legacy
   secret) is ABSENT — REL-2 acceptance criterion explicitly rejects it.
5. **Publish workflow un-park status** — `grep "^# PUBLISH PARKED"` on
   `pypi-publish.yml` and `npm-publish.yml`. Both files carry a PARKED
   comment dated 2026-05-24 (yaagents was private). Sprint 5 MUST un-park
   them — remove the comment block, uncomment the tag trigger
   (`push.tags: ['v[0-9]+.[0-9]+.[0-9]+']`) — committed as part of this
   WI (in the same PR as the script).

**Manual checklist** (printed by the script; operator confirms via
`PRECHECK_MANUAL_OK=1` env var before the script exits 0):

- PyPI Trusted Publisher configured for each of the 3 packages → repo
  `ai-mpathyminds/yaagents`, workflow `pypi-publish.yml`, env `pypi`
  (and `testpypi` for TestPyPI). Configuration done on pypi.org → Account
  → Publishing → "Add a new pending publisher".
- npm `@aimpathyminds` scope owned by an org account (`npm org ls
  aimpathyminds` shows operator as admin/owner).
- `NODE_AUTH_TOKEN` value is a Granular Access Token, NOT a
  publish-everything legacy token.
- Repo visibility intent matches Sprint 5 plan (private now → public at
  Apache 2.0 launch, or stay private with workflows explicitly un-parked
  for manual `workflow_dispatch`).

acceptance:
- `bin/pi2-yaa-publish-precheck.sh` exits 0 with `PRECHECK_MANUAL_OK=1`
  set, after operator confirms the manual checklist items.
- All 5 automated checks PASS (PyPI name status logged as FREE or
  CLAIMED-BY-YOU; npm package status logged; 3 GH environments present;
  `NODE_AUTH_TOKEN` present; `# PUBLISH PARKED` comment absent from both
  workflows).
- Workflow un-park commit included in the same PR as the script — tag
  trigger uncommented on both `pypi-publish.yml` and `npm-publish.yml`.
- AUDIT row `publish-precheck-passed` appended on success; row cites
  GH-environments + PyPI-name + npm-scope evidence per STATUS-CLAIM
  DISCIPLINE.
- B-31 (REL-1) and B-32 (REL-2) `depends_on:` are updated to include
  this WI's runbook entry id.
library_justify: PI gate; no library import (operator script + manual checklist for external-registry account configuration).
depends_on: [WI-2yaa.LIC-1]

### WI-2yaa.REL-1: PyPI re-publish ×3 @ 0.2.0 (OIDC) [WIP] — Sprint 5
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: Re-publish `yaagents-fastapi`, `yaagents-client`, `yaagents-cli` at
version `0.2.0` via the existing PyPI Trusted Publisher OIDC workflow
(PI1-yaa REL-3 carry-forward). Each package's `pyproject.toml` has been
bumped by BUMP-1a/b/c. CI workflow: build → **TestPyPI publish (gate)** →
PyPI publish on tag `v0.2.0` push. No long-lived tokens. Pre-publish gate:
`grep -rn "YAAgents Community License" yaagents-fastapi yaagents-client
yaagents-cli` returns 0 hits in the source dist (license-flip
verification).
acceptance:
- `pip install yaagents-fastapi==0.2.0` succeeds from public PyPI
- `pip install yaagents-client==0.2.0` succeeds from public PyPI
- `pip install yaagents-cli==0.2.0` succeeds from public PyPI
- All three wheels report `License: Apache-2.0` in metadata (`pip show <pkg>`)
- All three wheels declare `Supports-YAAgents-Profile: v0.2`
- Zero long-lived token in repo CI (verified by inspecting workflow files —
  no `PYPI_TOKEN`/`PYPI_API_KEY` secret references)
- TestPyPI stage gates prod stage (TestPyPI publish runs first; failure halts)
library_ref: ADR PI1-yaa-0005, ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-1a, WI-2yaa.BUMP-1b, WI-2yaa.BUMP-1c]

### WI-2yaa.REL-2: npm re-publish @ 0.2.0 (OIDC provenance) [WIP] — Sprint 5
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: Re-publish `@aimpathyminds/yaagents-client` at version `0.2.0` via
the existing npm Provenance OIDC workflow (PI1-yaa REL-4 carry-forward).
`client-ts/package.json` bumped by BUMP-2. CI workflow: build dual ESM+CJS
→ `npm publish --provenance` via OIDC. Pre-publish gate: `grep -nE "Community License|source-available"
client-ts/` returns 0 hits in built artefacts.
acceptance:
- `npm install @aimpathyminds/yaagents-client@0.2.0` succeeds from public npm
- Provenance attached (visible on npm's package page)
- `npm view @aimpathyminds/yaagents-client@0.2.0 license` returns `Apache-2.0`
- No `NPM_TOKEN` secret in repo (verified by inspecting workflow files)
- Profile-version constant exports as `v0.2` from both ESM and CJS bundles
library_ref: ADR PI1-yaa-0005, ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-2]

### WI-2yaa.REL-3: GHCR gateway image re-publish @ 0.2.0 (multi-arch + SBOM) [WIP] — Sprint 5
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: Re-publish the gateway image at tag `0.2.0` + update `:latest` to
0.2.0 (PI1-yaa REL-5 carry-forward). Multi-stage Alpine Dockerfile
unchanged in structure; rebuild with all PI2-yaa code (plugin chain + LLM
specialisation). `docker/build-push-action` multi-arch
(`linux/amd64`+`linux/arm64`) → `ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0`
+ `:latest` via `GITHUB_TOKEN` OIDC (`packages: write`). **SBOM**: Syft
SPDX 2.3 JSON (carry-forward from PI1-yaa OQ-5 resolution) attached to
GitHub Release for `v0.2.0`. Image OCI metadata:
`org.opencontainers.image.licenses=Apache-2.0` (was `…=LicenseRef-YAAgents-Community-License-v0.1`
in PI1-yaa — flip per ADR PI2-yaa-0003). Cosign signing remains deferred
to PI3-yaa (note only).
acceptance:
- `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` succeeds, both architectures
- `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:latest` resolves to the 0.2.0 digest
- `docker inspect ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0 | jq '.[0].Config.Labels["org.opencontainers.image.licenses"]'` returns `Apache-2.0`
- SBOM artefact (SPDX 2.3 JSON) attached to the GitHub Release for `v0.2.0`
- `trivy image ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` reports 0 HIGH/CRITICAL
- No PAT used; verified by inspecting workflow file
library_ref: ADR PI1-yaa-0005, ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-3, WI-2yaa.PLG-6, WI-2yaa.LLM-4]

### WI-2yaa.REL-4: Go module tag `client-go/v0.2.0` + proxy.golang.org verify [READY] — Sprint 5
service: yaagents/(ci)
parent_feature: F-GOCLIENT
brief: Publish `client-go` via Go modules tag-driven release per PRD §9.4:
`git tag client-go/v0.2.0` + `git push --tags` on the yaagents repository
post-merge of GOC-4. **No CI workflow needed** — Go modules pull from the
git tag once `proxy.golang.org` indexes it (~30 min). Verification step
in CI: a small workflow `verify-go-module.yml` that runs after the tag
push, waits 30 min (or polls `proxy.golang.org/github.com/ai-mpathyminds/yaagents/client-go/@v/v0.2.0.info`
with backoff), then runs `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0`
in a scratch module and asserts the import works. Carries `library_ref:
ADR PI1-yaa-0005` even though Go modules is not OIDC — the spirit of the
ADR (no long-lived registry tokens) is satisfied trivially: tag pushes use
the standard `git` auth (same GitHub OIDC for the workflow that pushes the
tag).
acceptance:
- Tag `client-go/v0.2.0` exists on the yaagents repository (visible at
  `https://github.com/ai-mpathyminds/yaagents/releases/tag/client-go%2Fv0.2.0`)
- Within 30 min of tag push, `https://proxy.golang.org/github.com/ai-mpathyminds/yaagents/client-go/@v/v0.2.0.info`
  returns a JSON document with the tag
- `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` in a fresh
  Go module succeeds and resolves to the published `go.mod`
- `verify-go-module.yml` workflow runs and exits 0
library_ref: ADR PI1-yaa-0005 (no-long-lived-token spirit), ADR PI2-yaa-0003 (license-metadata in go.mod LICENSE file at module root)
depends_on: [WI-2yaa.GOC-4]

### WI-2yaa.PI-GATE: PI2-yaa close gate — installability + e2e + license-clean scan [READY] — Sprint 5
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: The PI close gate (architect-authored; platform-engineer runs).
Aggregates the 10 success criteria from `roadmap.md` into one orchestrated
script `bin/pi2-yaa-gate.sh`:
1. `pip install yaagents-fastapi==0.2.0 yaagents-client==0.2.0 yaagents-cli==0.2.0` from fresh venv
2. `npm install @aimpathyminds/yaagents-client@0.2.0` in fresh `npm init`
3. `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` both arches
4. `go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0` in fresh module
5. `cd examples/campaign-api && docker compose up -d && yaagents conformance-test http://localhost:8121 && docker compose down`
6. `cd examples/llm-gateway && docker compose up -d && ./test-e2e.sh && yaagents conformance-test http://localhost:8122 --require-plugin token-validator --require-plugin tenant-injector && docker compose down`
7. **License-clean repo scan**: `grep -rn "YAAgents Community License" --include='*.go' --include='*.py' --include='*.ts'` returns 0 hits
8. Output: PASS table + per-step PASS/FAIL; exit 0 iff all 8 steps PASS.
acceptance:
- `bin/pi2-yaa-gate.sh` exits 0
- The script prints a 10-row PASS table mapping to the 10 success criteria
  in `roadmap.md` §"Success criteria"
- Output captured as `portfolio/REPORTS/pi2-yaa-gate/PI2-yaa-<sha>.md`
  (platform-engineer wires the report step at A-4)
library_justify: PI gate; no library import (orchestration of installability + conformance + license-cleanliness assertions).
depends_on: [WI-2yaa.REL-1, WI-2yaa.REL-2, WI-2yaa.REL-3, WI-2yaa.REL-4, WI-2yaa.EX-LLM-3]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SUPPLY] Multi-arch image (amd64 + arm64) | feature WI | REL-3 (`docker/build-push-action` `linux/amd64`+`linux/arm64`; GHCR `:0.2.0` + `:latest`) |
| [SUPPLY] SBOM (Syft SPDX 2.3 JSON) | feature WI | REL-3 (carry-forward from PI1-yaa OQ-5; attached to GitHub Release `v0.2.0`) |
| [SUPPLY] OIDC trusted publishing — PyPI | feature WI | REL-1 (PyPI Trusted Publisher; no `PYPI_TOKEN` in repo) |
| [SUPPLY] OIDC trusted publishing — npm | feature WI | REL-2 (`npm publish --provenance` via OIDC; no `NPM_TOKEN` in repo) |
| [SUPPLY] OIDC trusted publishing — GHCR | feature WI | REL-3 (`GITHUB_TOKEN` OIDC; `packages: write`) |
| [SUPPLY] Go module proxy publish — no long-lived token | feature WI | REL-4 (tag-driven; standard `git push --tags`; no token needed) |
| [SUPPLY] Apache 2.0 `LICENSE` at repo root + SPDX sweep | feature WI | LIC-1 + LIC-2 (`license-and-headers.md`) |
| [SEC] Community-license grep CI gate | **NFR WI below** | WI-2yaa.NFR-REL-1 (CI matrix integration; carry-forward from PI1-yaa REL-6) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no TF edits in PI2-yaa; GHCR/PyPI/npm are public registries; Go module proxy is free. No AWS run-rate change. FinOps WI is explicitly N/A for this PI. |
| Cosign image signing | NOTE — deferred | PI3-yaa scope (noted in REL-3 brief; no WI in PI2-yaa) |

### WI-2yaa.NFR-REL-1: CI matrix — community-license grep + SPDX audit [READY]
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: [SEC + SUPPLY] Integrate the two license-hygiene checks from the
NFR pass into the full CI matrix (`release-and-publish.md` REL-6
carry-forward analog):
(1) **Community-license grep** (carry-forward of WI-2yaa.NFR-LIC-1): on
every push to `main`, CI step `license-clean-scan` runs the grep (see
`license-and-headers.md` WI-2yaa.NFR-LIC-1 for the exact command); exits
1 on any hit. This step MUST also run in the REL-1/2/3 publish workflows
as a pre-publish gate (before `hatch build`, `npm publish`, `docker push`
respectively) so a regression can't reach a public registry.
(2) **SPDX metadata audit**: for each published artefact (three Python
wheels, one npm tarball, one Docker image), CI logs the license field
(`pip show`, `npm view`, `docker inspect … labels`) and asserts it equals
`Apache-2.0`. This is a belt-and-suspenders check on top of BUMP-1a/b/c,
BUMP-2, and REL-3 OCI-label assertions that are already in each WI's
acceptance criteria.
acceptance:
- `license-clean-scan` step present in `.github/workflows/ci.yml` AND in each of `pypi-publish.yml`, `npm-publish.yml`, `ghcr-publish.yml`; all exit 1 on any Community-License hit
- Each publish workflow logs the license field of the artefact before declaring success
- All four artefact-license assertions return `Apache-2.0` on the v0.2.0 tagged commit
library_ref: ADR PI2-yaa-0003, ADR PI1-yaa-0005
depends_on: [WI-2yaa.LIC-2, WI-2yaa.REL-1, WI-2yaa.REL-2, WI-2yaa.REL-3, WI-2yaa.REL-4]
