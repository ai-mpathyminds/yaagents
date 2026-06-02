# PI3-yaa — Component: Public launch (publish wave + PRIVATE→PUBLIC flip + PI-GATE)

Owner lane: **operator-driven / CI tag-push** for publish WIs (`model: none` mechanical-entry
pattern at execution-runbook time) + **chief-architect / operator** for the PRIVATE→PUBLIC repo
flip. Sprints 6–7. **All LA-* WIs depend on B-01 PRECHECK PASS** (intake §A-8 + PRD §10.6).

B-01 (first Phase B entry; mechanical `model: none` operator-driven publish-prep) is NOT a
roadmap WI — it is authored at A-6 by `execution-runbook-generator` per planning runbook
`context.first_b_entry_intent`. Its 5 checks (PRD §10.6):

1. PyPI Trusted Publisher configured for `yaagents-fastapi` / `yaagents-client` / `yaagents-cli` on the new submodule repo paths.
2. npm `@aimpathyminds` scope ownership + Granular Access Token + `NODE_AUTH_TOKEN` secret on `yaagents-client-ts` repo.
3. GitHub Environments `pypi` / `testpypi` / `npm` with tag-ref protection on each relevant submodule repo.
4. GHCR org package settings allow public-visibility flip for `yaagents-gateway`.
5. `bin/yaagents-public-mirror-verify.sh` returns 0 hits on dry-run scrub.

Gate: operator confirms `PRECHECK_MANUAL_OK=1` before any LA-* WI dispatches.

> **Library gate (Gate 3) — applies to every LA-* WI**: `library_ref: ADR PI1-yaa-0005`
> (OIDC trusted publishing carries forward unchanged across PI3-yaa); `library_ref: ADR
> PI2-yaa-0003` (Apache 2.0 metadata fields carry forward); `library_ref: ADR PI3-yaa-0002`
> (new Go module paths for sdk-go + client-go); `library_ref: ADR PI3-yaa-0001` (per-submodule
> repo structure determines per-publish-workflow surface).
>
> **Gate 4 duplication on LA-PYPI×3**: 3 publish WIs differ only in package name. **Architect
> override**: each WI is a 3-line GitHub Actions tag-push that delegates to the PyPI Trusted
> Publisher OIDC flow — the publish action is OWNED by the registry-side trust config, not
> shared yaagents code. No code to extract.
>
> `duplication_override: each LA-PYPI-* WI is a tag-push trigger only; OIDC Trusted Publisher
> OWNS the publish action; no shared portfolio code to extract.`

---

### WI-3yaa.LA-PYPI-FASTAPI: PyPI publish `yaagents-fastapi@0.3.0` [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-sdk-fastapi
parent_feature: F-LAUNCH
brief: Trigger PyPI publish of `yaagents-fastapi==0.3.0` via OIDC Trusted Publisher tag-driven
workflow on `yaagents-sdk-fastapi` submodule repo. Pre-conditions: B-01 PRECHECK PASS (PyPI
Trusted Publisher configured + `pypi` GitHub Environment exists with tag-ref protection);
RP-SDKFASTAPI-INIT landed; `pyproject.toml` declares `version = "0.3.0"`.

Operator/CI action sequence:
1. From `yaagents-sdk-fastapi` repo `main` (post RP-SDKFASTAPI-INIT): `git tag v0.3.0`
2. `git push origin v0.3.0`
3. GHA `.github/workflows/pypi-publish.yml` triggers on tag push, environment `pypi` (requires
   manual approval if configured per B-01 check 3); runs Hatch build → TestPyPI publish
   (validation) → PyPI prod publish via OIDC.
4. Verify: `pip install yaagents-fastapi==0.3.0` from public PyPI in a fresh venv succeeds.
acceptance:
- `v0.3.0` tag present on `yaagents-sdk-fastapi` repo.
- `pip index versions yaagents-fastapi` shows `0.3.0`.
- `pip install yaagents-fastapi==0.3.0 && python -c "import yaagents_fastapi; print(yaagents_fastapi.__version__)"` returns `0.3.0` in a fresh venv.
- Package METADATA shows `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3` (RP-SDKFASTAPI-INIT carry).
- No long-lived PyPI token used (verified via workflow `permissions: id-token: write` + Trusted Publisher OIDC flow).
library_ref: ADR PI1-yaa-0005 (OIDC trusted publishing); ADR PI2-yaa-0003 (Apache 2.0 license metadata); ADR PI3-yaa-0001 (per-submodule repo structure).
duplication_override: each LA-PYPI-* WI is a tag-push trigger only; OIDC Trusted Publisher OWNS the publish action; no shared portfolio code to extract.
depends_on: [WI-3yaa.RP-SDKFASTAPI-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-PYPI-CLIPY: PyPI publish `yaagents-client@0.3.0` [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-client-python
parent_feature: F-LAUNCH
brief: Mirror LA-PYPI-FASTAPI for `yaagents-client` on `yaagents-client-python` submodule repo.
Tag push `v0.3.0` → OIDC Trusted Publisher → PyPI prod.
acceptance:
- `v0.3.0` tag on `yaagents-client-python` repo.
- `pip install yaagents-client==0.3.0` from public PyPI succeeds.
- Package METADATA: `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3`.
- No long-lived token used.
library_ref: ADR PI1-yaa-0005; ADR PI2-yaa-0003; ADR PI3-yaa-0001.
duplication_override: each LA-PYPI-* WI is a tag-push trigger only; OIDC Trusted Publisher OWNS the publish action; no shared portfolio code to extract.
depends_on: [WI-3yaa.RP-CLIPY-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-PYPI-CLI: PyPI publish `yaagents-cli@0.3.0` [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-cli
parent_feature: F-LAUNCH
brief: Mirror LA-PYPI-FASTAPI for `yaagents-cli` on `yaagents-cli` submodule repo.
acceptance:
- `v0.3.0` tag on `yaagents-cli` repo.
- `pip install yaagents-cli==0.3.0` from public PyPI succeeds.
- Package METADATA: `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3`.
- `yaagents --help` works in a fresh venv post-install; subcommands present (validate-openapi, validate-response, conformance-test, init fastapi).
- No long-lived token used.
library_ref: ADR PI1-yaa-0005; ADR PI2-yaa-0003; ADR PI3-yaa-0001.
duplication_override: each LA-PYPI-* WI is a tag-push trigger only; OIDC Trusted Publisher OWNS the publish action; no shared portfolio code to extract.
depends_on: [WI-3yaa.RP-CLI-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-NPM: npm publish `@aimpathyminds/yaagents-client@0.3.0` with provenance [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-client-ts
parent_feature: F-LAUNCH
brief: Trigger npm publish of `@aimpathyminds/yaagents-client@0.3.0` with provenance attestation
via OIDC + `NODE_AUTH_TOKEN` (Granular Access Token from B-01 precheck). Pre-conditions:
B-01 PRECHECK PASS; RP-CLITS-INIT landed; `package.json` declares `"version": "0.3.0"`.

Operator/CI action sequence:
1. `git tag v0.3.0` + `git push origin v0.3.0` on `yaagents-client-ts` repo.
2. GHA `.github/workflows/npm-publish.yml` triggers; runs `pnpm build` → `npm publish --provenance --access public`.
3. Verify: `npm install @aimpathyminds/yaagents-client@0.3.0` from public npm succeeds with provenance attestation reported.
acceptance:
- `v0.3.0` tag on `yaagents-client-ts` repo.
- `npm view @aimpathyminds/yaagents-client@0.3.0` returns the published manifest.
- `npm install @aimpathyminds/yaagents-client@0.3.0` succeeds in a fresh `node_modules`.
- Provenance attestation present: `npm view @aimpathyminds/yaagents-client@0.3.0 dist.attestations` returns non-empty (or `gh attestation verify --owner ai-mpathyminds ...` passes).
- `package.json` `license: "Apache-2.0"` carries through to the published manifest.
library_ref: ADR PI1-yaa-0005 (OIDC trusted publishing + npm provenance); ADR PI2-yaa-0003; ADR PI3-yaa-0001.
depends_on: [WI-3yaa.RP-CLITS-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-GHCR: GHCR publish `yaagents-gateway:0.3.0` (multi-arch + Cosign + Syft SBOM) [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-gateway
parent_feature: F-LAUNCH
brief: Trigger GHCR publish of `ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` (multi-arch
`linux/amd64` + `linux/arm64`) + Cosign signature + Syft SBOM attached. Pre-conditions: B-01
PRECHECK PASS (GHCR org settings allow public visibility flip post-publish); RP-GATEWAY-INIT
landed.

Operator/CI action sequence:
1. `git tag v0.3.0` + `git push origin v0.3.0` on `yaagents-gateway` repo.
2. GHA `.github/workflows/ghcr-publish.yml` triggers; runs `docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0 --push .`
3. Workflow runs Syft SBOM generation: `syft ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0 -o spdx-json | cosign attach sbom --sbom -`
4. Cosign keyless signing (OIDC via GitHub Actions): `cosign sign --yes ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0`
5. Operator flips GHCR package visibility to PUBLIC post-publish (manual step in GHCR org UI).
6. Verify: `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` succeeds; `cosign verify ...` passes; SBOM attached.
acceptance:
- `v0.3.0` tag on `yaagents-gateway` repo.
- `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` succeeds on both `linux/amd64` + `linux/arm64` (verified via `docker manifest inspect`).
- `cosign verify --certificate-identity-regexp '.+' --certificate-oidc-issuer https://token.actions.githubusercontent.com ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` returns PASS (keyless OIDC signature verified).
- `cosign download sbom ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0 | jq .` returns the Syft SBOM (SPDX format).
- OCI label `org.opencontainers.image.licenses=Apache-2.0` present.
- No long-lived registry token (workflow uses `permissions: id-token: write, packages: write` only).
- GHCR package visibility = `public` post-flip.
library_ref: ADR PI1-yaa-0005 (OIDC trusted publishing + GHCR OIDC); ADR PI2-yaa-0003 (Apache 2.0 OCI label); ADR PI3-yaa-0001.
depends_on: [WI-3yaa.RP-GATEWAY-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-GO-CLIENT: Go module publish `yaagents-client-go@v0.3.0` [READY] — Sprint 6
service: github.com/ai-mpathyminds/yaagents-client-go
parent_feature: F-LAUNCH
brief: Tag-driven Go module publish via `proxy.golang.org`. Mirrors SG-7 (which publishes
`yaagents-sdk-go`). Pre-conditions: RP-CLIGO-INIT landed (module path migrated to
`github.com/ai-mpathyminds/yaagents-client-go`).

Operator/CI action sequence:
1. `git tag v0.3.0` + `git push origin v0.3.0` on `yaagents-client-go` repo.
2. Wait ≤30 min for `proxy.golang.org` index propagation (PRD §10.4 SLO).
3. Verify in fresh Go workspace: `go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` succeeds.
acceptance:
- `v0.3.0` tag on `yaagents-client-go` repo.
- `proxy.golang.org` returns 200 for the module at `v0.3.0` within 30 min of push.
- `go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` succeeds in a fresh `GOPATH`.
- `go list -m -json github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` reports correct version + commit SHA.
library_ref: ADR PI3-yaa-0001 (per-submodule repo structure); ADR PI3-yaa-0002 (Go module path migration: yaagents/client-go subpath → yaagents-client-go own module); ADR PI1-yaa-0005 (publishing-discipline consistency — N/A directly for Go modules but cited cross-component).
depends_on: [WI-3yaa.RP-CLIGO-INIT, WI-3yaa.RP-XREF]

### WI-3yaa.LA-PUBLIC-FLIP: Flip 8 repos PRIVATE→PUBLIC on GitHub [READY] — Sprint 7
service: github.com/ai-mpathyminds/yaagents{,-gateway,-sdk-fastapi,-sdk-go,-client-python,-client-ts,-client-go,-cli}
parent_feature: F-LAUNCH
brief: Flip the 8 repos (1 meta + 7 submodules) from PRIVATE to PUBLIC on GitHub. Pre-conditions:
ALL publish WIs (LA-PYPI×3, LA-NPM, LA-GHCR, SG-7, LA-GO-CLIENT) returned PASS (the prod
packages are installable from public registries before the source repos go public — narrow the
"public source but no installable package" window to zero). SCRUB verification (bin/yaagents-public-mirror-verify.sh)
returns 0 hits on all 8 repos (final pre-flip gate).

Operator action sequence (manual; 8 × GitHub Settings → Visibility → Change visibility → Public):
1. `gh repo edit ai-mpathyminds/yaagents --visibility public`
2. `gh repo edit ai-mpathyminds/yaagents-gateway --visibility public`
3. `gh repo edit ai-mpathyminds/yaagents-sdk-fastapi --visibility public`
4. `gh repo edit ai-mpathyminds/yaagents-sdk-go --visibility public`
5. `gh repo edit ai-mpathyminds/yaagents-client-python --visibility public`
6. `gh repo edit ai-mpathyminds/yaagents-client-ts --visibility public`
7. `gh repo edit ai-mpathyminds/yaagents-client-go --visibility public`
8. `gh repo edit ai-mpathyminds/yaagents-cli --visibility public`

Enable GitHub Discussions on the meta-repo (`gh repo edit ai-mpathyminds/yaagents --enable-discussions`) per PRD §11 OQ-5.
acceptance:
- `gh repo view ai-mpathyminds/yaagents --json visibility -q .visibility` returns `PUBLIC` on all 8 repos.
- GitHub Discussions enabled on `ai-mpathyminds/yaagents` meta-repo.
- `git clone https://github.com/ai-mpathyminds/yaagents.git` (no auth) succeeds for an unauthenticated user.
- Final `bin/yaagents-public-mirror-verify.sh` on all 8 repos returns 0 hits (final post-flip verification).
library_ref: ADR PI3-yaa-0001 (per-submodule repo structure); ADR PI3-yaa-0003 (orphan-baseline squashed history; clean public surface).
depends_on: [WI-3yaa.LA-PYPI-FASTAPI, WI-3yaa.LA-PYPI-CLIPY, WI-3yaa.LA-PYPI-CLI, WI-3yaa.LA-NPM, WI-3yaa.LA-GHCR, WI-3yaa.SG-7, WI-3yaa.LA-GO-CLIENT]

### WI-3yaa.LA-PAGES-DEPLOY: Push initial Pages site live [READY] — Sprint 7
service: yaagents (meta-repo)
parent_feature: F-LAUNCH
brief: Trigger the `.github/workflows/pages.yml` build+deploy workflow against meta-repo `main`
post-public-flip; verify the site is reachable at `https://ai-mpathyminds.github.io/yaagents/`.

Operator action:
1. Push a no-op commit (or push the PG-8 workflow file commit) to `main` of meta-repo (post LA-PUBLIC-FLIP) — workflow triggers automatically.
2. OR manual dispatch: `gh workflow run pages.yml -R ai-mpathyminds/yaagents`.
3. Wait for workflow completion (build + deploy via `actions/deploy-pages@v4`).
4. Verify in browser: `https://ai-mpathyminds.github.io/yaagents/` loads with all 9 sidebar entries rendering.
acceptance:
- `gh run list -R ai-mpathyminds/yaagents --workflow pages.yml --limit 1 --json status -q .[0].status` returns `completed`.
- `gh run list -R ai-mpathyminds/yaagents --workflow pages.yml --limit 1 --json conclusion -q .[0].conclusion` returns `success`.
- `curl -sI https://ai-mpathyminds.github.io/yaagents/ | head -1` returns `HTTP/2 200`.
- Manual visual check: hero + sidebar + Quick Start + Profile Spec + 6 SDK Quickstarts + 2 Examples + Plugin Authoring + Public Roadmap + Contributing + Community all render.
- Lighthouse CI gate (platform-engineer A-4 NFR pass): Perf ≥90, A11y ≥90.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages; `actions/deploy-pages@v4` deploy path).
depends_on: [WI-3yaa.PG-1, WI-3yaa.PG-2, WI-3yaa.PG-3, WI-3yaa.PG-4, WI-3yaa.PG-5, WI-3yaa.PG-6, WI-3yaa.PG-7, WI-3yaa.PG-8, WI-3yaa.LA-PUBLIC-FLIP]

### WI-3yaa.LA-PI-GATE: End-of-PI acceptance gate [READY] — Sprint 7
service: yaagents (cross-component)
parent_feature: F-LAUNCH
brief: Run the full prod-install regression check covering PRD §1 Goals (success criteria #1–10
in roadmap.md). The PI-GATE is the binding acceptance gate; PI3-yaa PC-1 does NOT close until
LA-PI-GATE passes.

Checks:
1. **Source repos public**: `gh repo view ai-mpathyminds/yaagents{,-gateway,-sdk-fastapi,-sdk-go,-client-python,-client-ts,-client-go,-cli} --json visibility -q .visibility` returns `PUBLIC` on all 8.
2. **PyPI ×3 prod-installable**: fresh venv `pip install yaagents-fastapi==0.3.0 yaagents-client==0.3.0 yaagents-cli==0.3.0` succeeds; each METADATA shows `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3`.
3. **npm prod-installable with provenance**: `npm install @aimpathyminds/yaagents-client@0.3.0` succeeds; provenance attestation verifies.
4. **GHCR prod-installable + Cosign**: `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` succeeds on both arches; `cosign verify ...` PASS; Syft SBOM attached.
5. **Go modules prod-installable**: `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0` + `go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` succeed within 30 min of tag push (or already propagated post-S6).
6. **Pages site live**: `curl -sI https://ai-mpathyminds.github.io/yaagents/` returns 200; Lighthouse ≥90 Perf+A11y (carried from LA-PAGES-DEPLOY).
7. **ai-platform/agent-api canary green**: ≥1 ai-platform/agent-api resource endpoint runs through sdk-go end-to-end (cross-lane A-3b WI close; reported via paired ai-platform-side acceptance).
8. **Portfolio scrub clean**: `bin/yaagents-public-mirror-verify.sh` on all 8 public repos returns 0 hits.
9. **Apache 2.0 cross-component clean**: `grep -rn "Community License" --include='*.go' --include='*.py' --include='*.ts'` on all 8 repos returns 0 hits at v0.3.0.
10. **`examples/campaign-api-go/` Compose green**: `cd examples/campaign-api-go && docker compose up && curl …` returns all 5 PRD §13.2 flows correct status × Content-Type pairs.

PI-GATE is the **last check before PC-1**. Failure on any check blocks PI3-yaa close.
acceptance:
- All 10 checks above PASS; status-claim discipline (PC-5-11) — each PASS cites concrete evidence (registry URL + commit SHA + DNS lookup + GH repo visibility output).
- Evidence report committed at `portfolio/REPORTS/launch/PI3-yaa-launch-<sha>.md` (operator-authored; format matches PI2-yaa pattern).
- Cross-lane: ai-platform-architect confirms canary WI close before PI-GATE flips green.
library_ref: ADR PI1-yaa-0005 (OIDC trusted publishing); ADR PI2-yaa-0003 (Apache 2.0 license posture); ADR PI3-yaa-0001 (per-submodule repo shape); ADR PI3-yaa-0002 (Go module path migration); ADR PI3-yaa-0003 (orphan-baseline squashed history); ADR PI3-yaa-0004 (Pages on GitHub Pages); ADR PI3-yaa-0005 (org branding).
depends_on: [WI-3yaa.LA-PUBLIC-FLIP, WI-3yaa.LA-PAGES-DEPLOY]
