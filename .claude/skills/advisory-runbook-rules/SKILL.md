---
name: advisory-runbook-rules
description: Companion to execution-runbook-generator: verbatim advisory runbook rules 11-13 (library annotation, duplication detection, pkg-manager check) + rule-12 algorithm. Warn-only; never halts authorship.
---

# advisory-runbook-rules

Companion to `execution-runbook-generator`. Holds the **verbatim** advisory rules §11–§13 and the rule-§12 duplication-detection algorithm, extracted from the main skill at WI-15.GOV-7 (PI15-plt-aip, PC-5-05 #14) to keep the main skill's hard-rule core lean. **Numbering and semantics are unchanged** — the main skill still evaluates §11–§13 during authorship (Pass-1 step 8.5 / step 9 render) and emits the same `context.advisory_warnings:` / `context.duplication_findings:` blocks; this file is the source-of-truth rule text it applies. Hard rules §1–§10 (which halt authorship) remain in `execution-runbook-generator/SKILL.md`.

## Advisory rules (warn-only — surface for architect choice; never halt authorship)

Per `.claude/rules/library-gates.md` (PI10 PC-4 carry-forward). These rules SURFACE findings into the runbook + AUDIT + postmortem inputs so the architect's choice is informed and audited; they do NOT block runbook authorship. Each warn becomes a postmortem candidate at PC-4.

| # | Rule | Surface form | Source-of-truth |
|---|------|--------------|-----------------|
| 11 | Per-WI library annotation. Every feature entry SHOULD carry `library_ref:` (cite a `LIBRARIES.md` row) OR `library_justify:` (one line "no catalog entry applies because…"). | `> warn: rule-§11 entry <id> WI <wi-id>: missing library_ref/library_justify` printed at end of authorship + appended to `context.advisory_warnings:` block in the runbook. | `library-gates.md` Gate 3 |
| 12 | Same-PI duplication detection. For every pair of `[DRAFT]` WIs whose `service:` differs but whose `brief:` first-paragraph token-Levenshtein similarity ≥0.7 (after stripping service-name + stop-tokens), emit a `[PROPOSED]` extraction-WI suggestion in the runbook's `context.duplication_findings:` block AND append a `duplication-found` row to `portfolio/AUDIT.md` BEFORE the duplicate WIs leave `[DRAFT]`. Architect may accept (extract WI joins roadmap) OR override per-WI with a `duplication_override: <one line>` field — both routes are valid. | `context.duplication_findings:` array in the runbook (1 entry per matched pair); AUDIT `duplication-found` row; postmortem rolls up acceptance/override ratio. | `library-gates.md` Gate 4 |
| 13 | Package-manager validation. When an entry's `brief:` references `pnpm`/`npm`/`yarn`, verify the named tool matches the lockfile present in the target product's UI dir (`pnpm-lock.yaml` / `package-lock.json` / `yarn.lock`). | `> warn: rule-§13 entry <id>: brief references <tool>, lockfile is <other>` + `context.advisory_warnings:` append. | `library-gates.md` Gate 5 |

**Distinction from §1–§10**: hard rules halt authorship; advisory rules write the runbook + record the finding. The runbook still authors with the warnings included; the architect reads them in `## Advisory warnings` (auto-prepended to the file's top comment block by the generator) and decides the response. **The rule-of-three trade-off remains a valid architect choice** (PI10-user-input.md §Counter-positions explicit) — these rules surface the choice, they do not pre-empt it.

### Algorithm — rule §12 (duplication detection)

For every unordered pair `(wi_i, wi_j)` in the runbook:

1. Skip if same `service:` field (intra-service duplication is out of scope; that's `code-reviewer`'s lane).
2. Tokenize `brief.first_paragraph` for each: lowercase, split on `[\s,.:;()/]+`, drop stop-tokens (`the|a|an|and|or|to|for|of|in|on|with|service|wi|brief`).
3. Strip occurrences of either WI's `service:` value from both token lists (so "campaigns RS256 middleware" vs "assets RS256 middleware" doesn't carry trivial vocabulary divergence).
4. Compute token-edit distance (Levenshtein on the token sequences). Similarity = `1 - dist / max(len_i, len_j)`.
5. If similarity ≥ 0.7 ⇒ emit a `context.duplication_findings:` entry: `{pair: [wi_i, wi_j], similarity: <float>, recommend: extract-to-shared-lib, owner: platform-librarian}`.
6. After the pass: if the count of findings ≥ 1, prepend an `## Advisory warnings` comment block to the YAML naming each pair + the catalog-row-or-not status (cross-reference `LIBRARIES.md §Extraction proposals (open)` — a finding whose pair already has an open proposal row is double-flagged with `existing_catalog_row: true` + the row text).

## What this skill is NOT

- Not a separately-dispatched skill — it carries no procedure of its own; `execution-runbook-generator` is the sole caller and applies these rules inline during A-6 authorship.
- Not a behavioral change — extraction is verbatim text relocation only (WI-15.GOV-7 AC: "no behavioral change to runbook authoring; advisory semantics preserved").
- Not the home of hard rules §1–§10 or the rule-§8 algorithm — those stay in `execution-runbook-generator/SKILL.md` because they halt authorship.
