# PI2-yaa — Component: CLI validator (`cli/`) — version bump + conformance v0.2

Owner lane: **python-developer**. Sprint 5. PyPI package `yaagents-cli`.
ADR: PI2-yaa-0003 (Apache 2.0 supersession).

> **Library gate:** `library_ref: ADR PI2-yaa-0003` for the metadata change.
> `library_justify: novel conformance harness; standalone OSS surface` for
> CLI-CONF (the v0.2 conformance extensions are net-new validator logic
> for the plugin chain + `X-YAAgents-Profile: v0.2` header). Gate 4
> duplication override w/ BUMP-1a/1b: see `client-python.md` shared
> rationale.

---

### WI-2yaa.BUMP-1c: cli 0.2.0 (Apache 2.0 metadata) [DONE] — Sprint 5
service: yaagents/cli
parent_feature: F-LICENSE
brief: Bump `cli/pyproject.toml` to `version = "0.2.0"`. License metadata
→ `Apache-2.0` + classifier. SPDX headers on all `cli/**/*.py` (covered
by LIC-2 sweep). Update README install snippet to `pip install
yaagents-cli==0.2.0`. **No API surface change to the four `yaagents`
sub-commands** (PRD §5.8: `validate-openapi`, `validate-response`,
`conformance-test`, `init fastapi`); the conformance command's *behaviour*
is extended in CLI-CONF below, but the CLI argument surface itself is
unchanged.
acceptance:
- `pyproject.toml` declares `version = "0.2.0"` + `Apache-2.0` license
- Built wheel metadata reports `License: Apache-2.0`
- `yaagents --version` prints `0.2.0`
- README install snippet matches new version
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-2]

### WI-2yaa.CLI-CONF: `yaagents conformance-test` v0.2 (plugin chain + profile header) [DONE] — Sprint 5
service: yaagents/cli
parent_feature: F-LICENSE
brief: Extend `yaagents conformance-test <base-url>` to validate the v0.2
profile and plugin-chain semantics:
1. **Profile header assertion** — every conformance request asserts
   `X-YAAgents-Profile: v0.2` in the response headers (was `v0.1` in PI1-yaa).
2. **Plugin presence checks** — when the operator passes
   `--require-plugin token-validator --require-plugin tenant-injector`
   (repeatable flag), conformance issues a request known to exercise each
   plugin (e.g. missing token → assert 403 vendor-error body from
   token-validator; missing `X-Tenant-ID` to a tenantRequired route →
   assert 403 from tenant-injector) and reports per-plugin PASS/FAIL.
3. **Always-on assertion** — conformance attempts to validate that
   `token-validator` cannot be disabled by issuing the standard 10
   response-type exercises with an invalid JWT and asserting EVERY
   response is 403 from token-validator (i.e. no other plugin's response
   shape leaks through). The result is reported as
   `token-validator: always-on confirmed (10/10 paths returned 403)` or
   FAIL with details.
4. **`Content-Type` matrix preserved from PI1-yaa** — all 10 PRD §4 rows
   still asserted (response Content-Type matches status × media-type
   table); fixture corpus continues from PI1-yaa SPEC-5 (`tests/golden/`)
   with v0.2 schema-path updates.
5. Output: one summary table at the end (`status | requested | observed |
   pass`); exit 0 on PASS, exit 1 on any FAIL.
acceptance:
- `yaagents conformance-test http://gateway:8120` exits 0 against a v0.2 gateway with default plugin chain
- `--require-plugin token-validator` PASS reported when gateway runs the plugin
- `X-YAAgents-Profile: v0.2` enforced (running against a v0.1 gateway → exits 1 with profile-mismatch FAIL)
- The 10-row PRD §4 matrix asserted against the campaign-api reference example (regression — PI1-yaa fixtures continue to PASS at v0.2)
- ≥85% coverage on the conformance command path
library_justify: novel conformance harness for plugin-chain semantics; net-new validator logic; no portfolio shared library applies.
depends_on: [WI-2yaa.BUMP-1c, WI-2yaa.BUMP-3]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] pip-audit on v0.2.0 wheel | **NFR WI below** | WI-2yaa.NFR-CLI-1 |
| [SEC] CLI input-validation hardening | **N/A (scoped)** | The CLI is a read-only conformance validator: it issues HTTP requests to a target URL and reads responses; user-controlled inputs are the gateway base-URL (parsed by urllib) and flag strings (plain comparison). No exec-style injection surface exists. Scope-statement sufficient; no additional hardening WI needed. |
| [SUPPLY] `Supports-YAAgents-Profile: v0.2` in wheel metadata | feature WI | BUMP-1c (acceptance criterion) |
| [SUPPLY] Used by PI-GATE as conformance harness | feature WI | PI-GATE step 6 (`release-and-publish.md`); CLI-CONF is the gate harness invoked against both examples |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; PyPI is a public registry; zero cloud cost |

### WI-2yaa.NFR-CLI-1: pip-audit CI gate for yaagents-cli [WIP]
service: yaagents/cli
parent_feature: F-LICENSE
brief: [SEC] Run `pip-audit` against the built `yaagents-cli==0.2.0`
wheel in CI (carry-forward from PI1-yaa REL-6 CI matrix). CI step
`pip-audit-cli` in `.github/workflows/ci.yml`; exits 1 on HIGH/CRITICAL
advisory. Zero external runtime deps beyond `click`/`httpx` (or
`requests`) expected; advisory clean expected trivially but gate is
mandatory.
acceptance:
- CI step `pip-audit-cli` present; exits 1 on HIGH/CRITICAL advisory
- `pip-audit` passes on v0.2.0 tagged commit (0 findings)
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-1c]
