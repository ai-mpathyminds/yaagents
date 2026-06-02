# PI2-yaa — Component: Python client (`client-python/`) — version bump

Owner lane: **python-developer**. Sprint 5. PyPI package `yaagents-client`.
ADR: PI2-yaa-0003 (Apache 2.0 supersession).

> **Library gate:** `library_ref: ADR PI2-yaa-0003` for the license-metadata
> change. API surface unchanged from PI1-yaa (PRD §5.6 explicit).

---

### WI-2yaa.BUMP-1b: client-python 0.2.0 (Apache 2.0 metadata + profile v0.2) [DONE] — Sprint 5
service: yaagents/client-python
parent_feature: F-LICENSE
brief: Bump `client-python/pyproject.toml` to `version = "0.2.0"`. Update
license metadata to `license = { text = "Apache-2.0" }` (PEP 621) and
classifier `"License :: OSI Approved :: Apache Software License"`. Update
`PROFILE_VERSION = "v0.2"` constant. Confirm SPDX headers on all
`client-python/**/*.py` (covered by LIC-2 sweep). Update README install
snippet to `pip install yaagents-client==0.2.0`. **No API surface change**
— `YaAgentsClient`, `ClarificationRequired`, `ValidationFailed`,
`FailedDependency`, `AgenticForbidden` retain v0.1.0 signatures (PRD §5.6).
Existing golden-corpus tests pass with the profile-version assertion
bumped `v0.1` → `v0.2`.
acceptance:
- `pyproject.toml` declares `version = "0.2.0"`; license `Apache-2.0`
- `python -c "import yaagents_client; print(yaagents_client.PROFILE_VERSION)"` prints `v0.2`
- Built wheel metadata reports `License: Apache-2.0`
- Existing PI1-yaa golden-corpus tests pass after profile-version assertion bump
- README install snippet matches new version
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-2]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] pip-audit on v0.2.0 wheel | **NFR WI below** | WI-2yaa.NFR-PYC-1 |
| [SUPPLY] `Supports-YAAgents-Profile: v0.2` in wheel metadata | feature WI | BUMP-1b (acceptance criterion) |
| [SUPPLY] OIDC trusted publishing | feature WI | REL-1 in `release-and-publish.md` |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; PyPI is a public registry; zero cloud cost |

### WI-2yaa.NFR-PYC-1: pip-audit CI gate for client-python [WIP]
service: yaagents/client-python
parent_feature: F-LICENSE
brief: [SEC] Run `pip-audit` against the built `yaagents-client==0.2.0`
wheel in CI (carry-forward from PI1-yaa REL-6 CI matrix). CI step
`pip-audit-client-python` in `.github/workflows/ci.yml`; exits 1 on
HIGH/CRITICAL advisory. Zero external runtime deps expected (pure stdlib
`httpx` or `urllib.request` per PRD §5.6); advisory clean expected
trivially but the gate is mandatory.
acceptance:
- CI step `pip-audit-client-python` present; exits 1 on HIGH/CRITICAL advisory
- `pip-audit` passes on v0.2.0 tagged commit (0 findings)
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-1b]

## Gate 4 (duplication) override

Gate 4 fires across BUMP-1a / BUMP-1b / BUMP-1c (sdk-fastapi / client-python
/ cli) — three near-identical brief shapes differing only in package name.
**Override rationale**: these are three intentionally distinct PyPI
packages with separate `pyproject.toml`s, separate test corpora, and
separate publish workflows. They share a templated brief because the
license-metadata + profile-version + SPDX-touch work is mechanically
identical across the three packages — but the implementation files,
package names, and publish targets are independent. No extraction
candidate (the templated bump is a 6-line `pyproject.toml` change ×3, not
a shared library). The override is recorded in each file's WI body.
