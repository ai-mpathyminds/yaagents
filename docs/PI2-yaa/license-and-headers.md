# PI2-yaa — Component: License flip & SPDX header sweep (`LICENSE` + repo-wide)

Owner lane: **chief-architect** (LIC-1 LICENSE-file authority; v0.1.x boundary
clause is user-direct) + go-developer / python-developer / frontend-developer
(LIC-2 SPDX sweep across their respective dirs). ADR: PI2-yaa-0003
(supersedes PI1-yaa-0004). Sprint 1 — these are blocking prerequisites for
every subsequent SPDX-bearing artefact.

> **Library gate:** both WIs carry `library_ref: ADR PI2-yaa-0003`. The
> license flip is novel to PI2-yaa; no portfolio shared license-tooling
> library applies.

---

### WI-2yaa.LIC-1: `LICENSE` → Apache 2.0 verbatim + `COMMERCIAL.md` retirement [DONE]
service: yaagents/(root)
parent_feature: F-LICENSE
brief: Replace `LICENSE` (currently YAAgents Community License v0.1) with the
Apache License, Version 2.0 text **byte-verbatim from PRD §8.2**. Delete
`COMMERCIAL.md` (`git rm`). Update `README.md` License section to the
pointer block in PRD §8.3 (verbatim — mentions v0.1.x non-retroactive
boundary, contact email). `CONTRIBUTING.md` retains the `legal-review-pending`
banner verbatim per ADR PI2-yaa-0003 §Consequences until counsel sign-off
(removal is a PC-6 non-engineering checklist item, NOT this WI).
acceptance:
- `LICENSE` content matches PRD §8.2 byte-verbatim (diff against PRD §8.2 must be empty)
- `COMMERCIAL.md` absent from git tree (`git ls-files COMMERCIAL.md` empty)
- `README.md` License section matches PRD §8.3 verbatim; no "open source"/"OSI" claim (`grep -nE "open source|OSI"` returns 0 hits)
- `CONTRIBUTING.md` retains legal-review-pending banner unchanged
- One commit; message `Agent: chief-architect` (license-file authority); WI trailer `WI-2yaa.LIC-1`
library_ref: ADR PI2-yaa-0003
depends_on: []

### WI-2yaa.LIC-2: SPDX header sweep — all source dirs (single commit) [DONE]
service: yaagents/(root cross-cutting — go-developer + python-developer + frontend-developer lanes)
parent_feature: F-LICENSE
brief: Add the SPDX header (PRD §8.4) to **every source file** in:
`gateway/`, `sdk-fastapi/`, `client-python/`, `client-ts/`, `cli/`,
`client-go/` (created in S3 — LIC-2 amends client-go separately if S3 lands
before S1's LIC-2 commit; otherwise included), `examples/`, and any
plugin-package source under `gateway/internal/plugins/` (created S1+).
Header form per language:
```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds
```
```python
# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
```
```typescript
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds
```
YAML/JSON files: omit (no standard comment syntax — PRD §8.4 explicit).
**Single commit** (the entire sweep) — partial-sweep states in git history
are explicitly forbidden by PRD §8.4. The commit author trailer is the
language-lane developer whose dir contributes the largest LOC; co-authored
via `Co-Authored-By:` blocks for the others. Three-lane orchestration: each
language developer prepares their dir staged; chief-architect-proxy
coordinates the merge into one commit.
acceptance:
- `grep -rL "SPDX-License-Identifier: Apache-2.0" gateway/ sdk-fastapi/ client-python/ client-ts/ cli/ client-go/ examples/ --include='*.go' --include='*.py' --include='*.ts' --include='*.tsx'` returns 0 files (every source file has the header)
- All v0.1.x "YAAgents Community License" SPDX identifiers removed from v0.2.0 source tree (`grep -rn "YAAgents Community License" --include='*.go' --include='*.py' --include='*.ts'` returns 0 hits)
- Single commit; commit subject `PI2-yaa LIC-2: SPDX header sweep — Apache-2.0`
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-1]

---

## Cross-WI notes

- **client-go SPDX timing**: `client-go/` directory does not exist until S3
  GOC-1 lands. Two acceptable patterns: (a) GOC-1..4 source files include the
  SPDX header as a coding standard at authoring time (preferred — no
  retroactive sweep) and a small follow-up commit (`LIC-2-clientgo`) confirms;
  (b) defer the `client-go/` portion of LIC-2 to a follow-up commit after
  GOC-4 lands. Choice belongs to platform-engineer at A-4 sequencing review;
  default = (a).
- **Plugin packages**: `gateway/internal/plugins/*` directories created in S1
  PLG-3..5 and S2 PLG-7 — same rule as client-go: include the SPDX header at
  source-authoring time.
- **examples/llm-gateway/**: same — mock-LLM-backend source files (S4 EX-LLM-1)
  include the SPDX header at authoring time.

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] Community-license CI grep gate | **NFR WI below** | WI-2yaa.NFR-LIC-1 |
| [SUPPLY] Apache-2.0 license metadata in pyproject.toml ×3, package.json, go.mod | feature WI | BUMP-1a/1b/1c (sdk-fastapi/client-python/cli), BUMP-2 (client-ts), LIC-1 (repo root LICENSE) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI2-yaa (no TF edits; license file change has zero AWS cost implication) |

### WI-2yaa.NFR-LIC-1: Community-license grep CI gate [READY]
service: yaagents/(ci)
parent_feature: F-LICENSE
brief: [SEC] Wire a pre-merge CI grep gate in `.github/workflows/ci.yml` (or
equivalent) that runs on every push to `main` **after** the LIC-2 sweep
merge. The gate runs:
```bash
grep -rn "YAAgents Community License" \
  --include='*.go' --include='*.py' --include='*.ts' .
```
and exits non-zero (CI FAIL) on any match. This ensures the v0.1.x
Community License string never reappears in v0.2.0+ source. The gate is
also invoked by `WI-2yaa.PI-GATE` step 7 (`bin/pi2-yaa-gate.sh`).
acceptance:
- CI workflow step `license-clean-scan` added to `.github/workflows/ci.yml`; runs on `push` to `main` (not PRs, to avoid blocking forks — advisory)
- Step grep command matches the pattern above; step exits 1 on any hit (verified by injecting a dummy `// YAAgents Community License` comment in a test file on a branch, observing CI FAIL, then removing it)
- `grep -rn "YAAgents Community License" --include='*.go' --include='*.py' --include='*.ts' .` returns 0 hits on the v0.2.0 tagged commit (verified in PI-GATE step 7)
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-2]
