# PI1-yaa — Component: CLI validator (`cli/`)

Owner lane: python-developer. Published as PyPI `yaagents-cli`. CLI-1/2/3 land
Sprint 3 (depend on the contract); **CLI-4 conformance-test is Sprint 5** (the
reserved Compose-e2e + CLI-conformance gate). ADR: PI1-yaa-0002, PI1-yaa-0003.

---

### WI-1yaa.CLI-1: CLI skeleton + `validate-response` [DONE]
service: yaagents/cli
brief: `yaagents` CLI (Typer or argparse — zero heavy deps preferred).
`yaagents validate-response <file.json>`: infer media type from the body's
`type` field, validate against the matching `schemas/v0.1` schema, print
pass/fail + per-error JSON-pointer findings; non-zero exit on fail.
acceptance:
- Every `spec/examples/v0.1` valid fixture → PASS exit 0; every invalid → FAIL exit 1 with findings
- `--json` machine-readable output mode
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-2, WI-1yaa.SPEC-5]

### WI-1yaa.CLI-2: `validate-openapi` [DONE]
service: yaagents/cli
brief: `yaagents validate-openapi <file.yaml>`: assert (a) `x-yaagents`
metadata present + well-formed on agentic operations, (b) each declared
response type uses the correct vendor `Content-Type` per the §4 table, (c)
schema `$ref`s resolve to the `schemas/v0.1` set. pass/fail + findings.
acceptance:
- Passes on `openapi/yaagents-response-profile.yaml`; fails on a fixture with wrong content-type
- Detects missing `x-yaagents` and dangling `$ref`
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-4]

### WI-1yaa.CLI-3: `init fastapi` scaffold [DONE]
service: yaagents/cli
brief: `yaagents init fastapi`: generate a minimal FastAPI starter
(`main.py` with one `@agentic_operation` endpoint correctly wired to
`AgenticResponse`, `pyproject.toml` depending on `yaagents-fastapi`,
`routes.yaml` stub for the gateway). Idempotent (refuses to overwrite without
`--force`).
acceptance:
- Generated project runs (`uvicorn`) and its `/openapi.json` passes `validate-openapi`
- Refuses non-empty target dir without `--force`
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.CLI-2, WI-1yaa.SDK-3]

### WI-1yaa.CLI-4: `conformance-test <base-url>` [WIP] — Sprint 5 gate
service: yaagents/cli
brief: `yaagents conformance-test <base-url>`: exercise the mandatory response
types against a live service (via the gateway), assert correct status +
vendor content-type, `X-YAAgents-Profile` header, and correlation-id
propagation/echo. Emit the **exact PRD §5.8 report format** (✓ lines +
`Overall: PASS`); non-zero exit on any FAIL.
acceptance:
- Run against the Compose demo (`http://localhost:8120`) → `Overall: PASS`, exit 0
- Report text matches PRD §5.8 lines; a deliberately broken route → FAIL exit 1
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.CLI-1, WI-1yaa.EX-3]

---

## NFR Addendum — A-4 platform-engineer pass (2026-05-17)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] input-validation hardening | **NFR WI below** | WI-1yaa.NFR-CLI-1 |
| [SEC] dependency audit (`pip-audit`) | feature WI | REL-6 CI matrix (Python job) |
| [SRE] health/readiness/logs | N/A | CLI tool, not a running service |
| [SUPPLY-CHAIN] OIDC publish — no long-lived token | feature WI | REL-3 (PyPI Trusted Publisher) |
| [SUPPLY-CHAIN] reproducible builds (Hatch) | feature WI | REL-3 + ADR PI1-yaa-0003 |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI1-yaa |

### WI-1yaa.NFR-CLI-1: CLI input-validation hardening [READY]
service: yaagents/cli
brief: [SEC] Harden all CLI entry points against malicious input.
`validate-response <file>` and `validate-openapi <file>`: reject paths
containing `..` or absolute paths outside the working directory (path-traversal
defence); enforce a max file size of 10 MB before reading (DoS-from-local
defence). `conformance-test <base-url>`: validate the URL scheme is `http` or
`https` only — reject `file://`, `ftp://`, or any other scheme with exit 1 +
clear error; parse with `urllib.parse`/Node URL, do not string-match. `init
fastapi <target-dir>`: reject target paths with `..` traversal. All error
messages MUST NOT echo raw user input back as shell-interpretable text
(no f-string/template injection into shell calls).
acceptance:
- `yaagents validate-response ../../../etc/passwd` → exit 1 "path traversal not allowed"
- `yaagents conformance-test file:///etc/hosts` → exit 1 "scheme must be http or https"
- 5 MB file within limit → processed; 11 MB file → exit 1 "file exceeds 10 MB limit"
- `ruff` + `mypy` clean; no `subprocess(shell=True)` with user input
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.CLI-1, WI-1yaa.CLI-2, WI-1yaa.CLI-3, WI-1yaa.CLI-4]
