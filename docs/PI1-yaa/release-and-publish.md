# PI1-yaa — Component: Release, publishing & repo scaffolding (`release-and-publish`)

Owner lane: **platform-engineer** (supply-chain / NFR). The architect
enumerates these as real first-class WIs (publishing is not an afterthought —
runbook rule 5); platform-engineer refines NFR/[SUPPLY-CHAIN] bodies + picks
SBOM format at A-4. REL-1/2 land Sprint 1 (scaffolding parallel to the
contract); REL-3..6 land Sprint 5 (publish after artifacts are green).
ADR: PI1-yaa-0003 (Hatch / dual-bundle), PI1-yaa-0004 (license),
PI1-yaa-0005 (OIDC publishing).

---

### WI-1yaa.REL-1: Repo scaffolding [DRAFT] — Sprint 1
service: yaagents/(root)
brief: `README.md` (overview, quick-start, architecture diagram, badge set —
source-available wording per ADR 0004, never "open source"),
`CONTRIBUTING.md` (CLA placeholder + "contributions not accepted until legal
review" banner), `SECURITY.md`, `CODE_OF_CONDUCT.md` (Contributor Covenant),
`.github/ISSUE_TEMPLATE/` (bug / feature / adapter-request).
acceptance:
- All 5 docs present; README uses source-available/fair-code wording only (grep: no "open source"/"OSI")
- Issue templates render on GitHub
library_justify: novel; standalone OSS surface
depends_on: []

### WI-1yaa.REL-2: Dual-license artifacts [DRAFT] — Sprint 1
service: yaagents/(root)
brief: `LICENSE` = YAAgents Community License v0.1 (drafted verbatim from GTM
README §14), `COMMERCIAL.md` = commercial terms + contact. Carry the
GTM-§Appendix **legal-review-pending disclaimer verbatim**. Combined-AND
threshold placeholder (`<10 employees AND <USD 1M revenue`) per ADR 0004
(final lock = chief-architect/user at PC-6).
acceptance:
- `LICENSE` + `COMMERCIAL.md` present; disclaimer text byte-verbatim from GTM §Appendix
- No "OSI"/"open source" claims; threshold matches ADR 0004 placeholder
library_ref: ADR PI1-yaa-0004
depends_on: []

### WI-1yaa.REL-3: PyPI publish (×3, OIDC Trusted Publisher) [DRAFT] — Sprint 5
service: yaagents/(ci)
brief: Hatch `pyproject.toml` for `yaagents-fastapi`/`yaagents-client`/
`yaagents-cli` (ADR 0003); each declares `Supports-YAAgents-Profile: v0.1`.
GH Actions workflow: build → **TestPyPI publish (gate)** → PyPI publish via
**PyPI Trusted Publisher OIDC** (`id-token: write`, no API token). One
publisher registration per project (external setup = Phase-B dependency).
acceptance:
- `pip install yaagents-fastapi|yaagents-client|yaagents-cli` from public PyPI succeeds (PRD §12.10)
- Zero long-lived token in repo CI; TestPyPI stage gates prod stage
library_ref: ADR PI1-yaa-0005
depends_on: [WI-1yaa.SDK-4, WI-1yaa.PYC-3, WI-1yaa.CLI-4]

### WI-1yaa.REL-4: npm publish (OIDC provenance) [DRAFT] — Sprint 5
service: yaagents/(ci)
brief: `@yaagents/client` dual ESM+CJS package (ADR 0003) with
`yaagents.profile`+`PROFILE_VERSION`. GH Actions: build → `npm publish
--provenance` via OIDC (no `NPM_TOKEN`).
acceptance:
- `npm install @yaagents/client` from public npm succeeds (PRD §12.11); provenance attached
- No `NPM_TOKEN` secret in repo
library_ref: ADR PI1-yaa-0005
depends_on: [WI-1yaa.TSC-3]

### WI-1yaa.REL-5: GHCR gateway image (multi-arch, OIDC, SBOM) [DRAFT] — Sprint 5
service: yaagents/(ci)
brief: Multi-stage Alpine Dockerfile (non-root, `CGO_ENABLED=0`). GH Actions
`docker/build-push-action` multi-arch (`linux/amd64`+`linux/arm64`) →
`ghcr.io/yaagents/gateway:0.1.0`+`:latest` via `GITHUB_TOKEN` OIDC
(`packages: write`). **SBOM generated at publish + attached to GitHub
Release** (format = platform-engineer A-4 pick per ADR 0005/OQ-5). Cosign =
PI2-yaa (note only).
acceptance:
- `docker pull ghcr.io/yaagents/gateway:0.1.0` succeeds, both arches (PRD §12.12)
- SBOM artifact on the GitHub Release; `trivy` image scan in CI; no PAT used
library_ref: ADR PI1-yaa-0005
depends_on: [WI-1yaa.GW-5]

### WI-1yaa.REL-6: CI test/lint matrix + spec release archive [DRAFT] — Sprint 5
service: yaagents/(ci)
brief: GH Actions CI: Go (`golangci-lint`+`govulncheck`+test), Python
(`ruff`+`mypy`+`pip-audit`+pytest), TS (`eslint`+`tsc --noEmit`+`npm
audit`+vitest), each ≥80% coverage gate. Attach `schemas/` + `openapi/` as a
versioned archive to the `v0.1.0` GitHub Release. Heavy/live tests CI-only
(dev ceiling 16 GB).
acceptance:
- CI green on all three language jobs with coverage gates enforced
- `schemas`+`openapi` archive attached to the v0.1.0 Release
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.EX-4]
