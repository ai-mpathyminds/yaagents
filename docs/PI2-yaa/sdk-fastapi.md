# PI2-yaa — Component: FastAPI server SDK (`sdk-fastapi/`) — version bump

Owner lane: **python-developer**. Sprint 5 (re-publish gate window). PyPI
package `yaagents-fastapi`. ADR: PI2-yaa-0003 (Apache 2.0 supersession).

> **Library gate:** `library_ref: ADR PI2-yaa-0003` for the license-metadata
> change. The API surface itself is unchanged from PI1-yaa (PRD §5.5 explicit
> "version bump only; no API surface changes") — `library_justify: API
> surface unchanged from v0.1.0; license + profile-version metadata only`
> for the no-API-change clause.

---

### WI-2yaa.BUMP-1a: sdk-fastapi 0.2.0 (Apache 2.0 metadata + profile v0.2) [DONE] — Sprint 5
service: yaagents/sdk-fastapi
parent_feature: F-LICENSE
brief: Bump `sdk-fastapi/pyproject.toml` to `version = "0.2.0"`. Update
license metadata to `license = { text = "Apache-2.0" }` (PEP 621) and
classifier `"License :: OSI Approved :: Apache Software License"`. Update
the package's profile-version constant (`PROFILE_VERSION = "v0.2"`).
Confirm all `sdk-fastapi/**/*.py` source files carry the SPDX header
(this is also covered by LIC-2 sweep — BUMP-1a is the version-and-metadata
slice). Update `sdk-fastapi/README.md` install snippet to `pip install
yaagents-fastapi==0.2.0`. **No API surface change** — the decorator
`@agentic_operation`, `AgenticResponse`, `AgenticContext`, `RequiredInput`,
`AgenticResponses` retain v0.1.0 signatures verbatim (PRD §5.5). Update
existing tests' profile-version assertions (`v0.1` → `v0.2`).
acceptance:
- `pyproject.toml` declares `version = "0.2.0"`; license metadata reports `Apache-2.0`
- `python -c "import yaagents_fastapi; print(yaagents_fastapi.PROFILE_VERSION)"` prints `v0.2`
- Built wheel (`hatch build`) inspected via `python -m wheel unpack` reports `License: Apache-2.0` + classifier `Apache Software License`
- Existing PI1-yaa tests pass after profile-version assertion bump; no API-surface change (`git diff sdk-fastapi/src/` shows only SPDX header + version-constant + import-statement changes; no signature changes in the decorator surface)
- README install snippet matches the new version
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-2]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] pip-audit on v0.2.0 wheel | **NFR WI below** | WI-2yaa.NFR-SDK-1 |
| [SUPPLY] `Supports-YAAgents-Profile: v0.2` in wheel metadata | feature WI | BUMP-1a (acceptance criterion: wheel `METADATA` carries the field) |
| [SUPPLY] OIDC trusted publishing | feature WI | REL-1 in `release-and-publish.md` (PyPI Trusted Publisher carry-forward) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; PyPI is a public registry; zero cloud cost |

### WI-2yaa.NFR-SDK-1: pip-audit CI gate for sdk-fastapi [WIP]
service: yaagents/sdk-fastapi
parent_feature: F-LICENSE
brief: [SEC] Run `pip-audit` against the built `yaagents-fastapi==0.2.0`
wheel in CI (carry-forward from PI1-yaa REL-6 CI matrix). Command:
`pip-audit --requirement sdk-fastapi/requirements.txt` (or equivalent
`pip-audit dist/yaagents_fastapi-0.2.0-*.whl`); exits 1 on any
CRITICAL/HIGH advisory. Wired in `.github/workflows/ci.yml` as step
`pip-audit-sdk-fastapi`, running on every PR + main push after the wheel
builds. Zero external runtime deps expected (FastAPI + Starlette are
dev/test only for the SDK; see PRD §5.5); advisory clean is expected
trivially but the gate must be wired.
acceptance:
- CI step `pip-audit-sdk-fastapi` present; exits 1 on HIGH/CRITICAL advisory
- `pip-audit` passes on the v0.2.0 tagged commit (0 findings)
- Gate carries forward from PI1-yaa REL-6 (no regression — gate was already wired in PI1-yaa)
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-1a]
