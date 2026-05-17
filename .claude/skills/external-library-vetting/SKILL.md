---
name: external-library-vetting
description: Vet a third-party dependency for platform-librarian — pass/fail verdict against license, CVE, maintenance, size, stability. Escalates to governance-auditor for critical-path packages.
---

# external-library-vetting

Offline-first vet of a third-party library. Uses package metadata + public CVE indexes + published release notes. Output is a go/no-go verdict table.

## Inputs
- Package name + ecosystem (e.g. `github.com/pb33f/libopenapi` Go; `structlog` Python; `@tanstack/query` npm)
- Candidate version (or "latest")
- Intended use (critical-path? auth/crypto/network/data-at-rest? UI-only?)
- Adopter product(s)

## Checks

### 1. License
- **PASS**: MIT, Apache-2.0, BSD (any), ISC, Unlicense
- **CAVEAT** (flag for legal): MPL-2.0, LGPL-3.0, EPL-2.0 (weak copyleft; compatible with commercial per usage)
- **FAIL** (reject absent explicit legal approval): GPL-3.0, AGPL-3.0, SSPL, proprietary, unclear/missing

### 2. CVE history
- Check the ecosystem's vulnerability database (Go: `govulncheck`; Python: PyPI Advisory DB / pip-audit; npm: npm audit / GitHub Advisory)
- **FAIL**: any open high/critical CVE with no fix available in the candidate version or later
- **CAVEAT**: any CVE (any severity) in the last 12 months, even if patched — document
- **PASS**: clean for the candidate version

### 3. Maintenance activity
- Last commit date on primary branch
- Release cadence (latest 3 releases' dates)
- Open/closed issue ratio + median close time (best effort via repo API)
- **FAIL**: >18 months since last commit and no ongoing discussion
- **CAVEAT**: 6–18 months quiet OR single-maintainer (bus factor)
- **PASS**: active (commits in last 6 months; multi-maintainer or strong community)

### 4. Size footprint
- Go: compiled binary delta (estimate: library LoC × 1.5–3 KB per KLoC after stripping). Target: <10% of service's current binary.
- Python: installed size of the wheel (`pip show` / PyPI metadata). Target: <10% of service image size budget (<200 MB).
- Frontend: bundle delta via bundlephobia or explicit build measurement. Target: <10% of route bundle budget.
- **CAVEAT** (document): any dimension exceeding 10% but <25%.
- **FAIL**: any dimension > 25% unless core to the feature.

### 5. API stability
- Major version changes in last 24 months
- Breaking-change frequency in minor versions
- **FAIL**: >2 major versions in 24 months (high churn); or pre-1.0 with active breaking-change history AND intended for a critical path
- **CAVEAT**: pre-1.0 but narrowly-scoped use; or rapid minor-version breaks
- **PASS**: stable major (≥1.0) with semver discipline; or pre-1.0 for non-critical path with low breakage

### 6. Critical-path handoff (conditional)
If the library enters a critical path (auth, crypto, network transport, data-at-rest, or IAM primitives):
- Escalate to `governance-auditor` for security sign-off before adoption.
- Verdict cannot be ADOPT without governance-auditor approval noted in the handoff.

### 7. Alternatives scan
Name 1–2 credible alternatives. State why this one wins (feature coverage, community, perf, license compatibility). If no alternatives exist, say so explicitly.

## Output

```markdown
## External vetting — {package}@{version} — YYYY-MM-DD

**Intended use**: {critical-path? | UI-only | internal helper | etc.}
**Adopter(s)**: {products}

| Dimension | Finding | Verdict |
|---|---|---|
| License | {SPDX} | PASS / CAVEAT / FAIL |
| CVE | {n open / n recent} | PASS / CAVEAT / FAIL |
| Maintenance | {last commit; release cadence} | PASS / CAVEAT / FAIL |
| Size | {measurement} | PASS / CAVEAT / FAIL |
| API stability | {summary} | PASS / CAVEAT / FAIL |
| Critical path? | {yes/no; gov sign-off required?} | — |

**Alternatives considered**: {1-2 with rejection rationale each}

**Overall verdict**: **ADOPT** | **ADOPT-WITH-CAVEATS** | **REJECT**

**Rationale**: {one paragraph}

**Proposed catalog entry** (platform-librarian to add to portfolio/LIBRARIES.md):
- Name / ecosystem / version / owner / primary adopter / notes
```

## What this skill does NOT do

- Install the library (developer WI).
- Run dynamic/behavioral security tests (governance-auditor + CI territory).
- Decide adoption — platform-librarian recommends, architect decides.
- Re-vet on version bumps automatically — that's a quarterly governance sweep trigger (or an on-demand re-vet when the adopter bumps the version).
