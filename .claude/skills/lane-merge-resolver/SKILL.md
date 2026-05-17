---
name: lane-merge-resolver
description: Resolve cross-lane file conflicts at PC-2 for scrum-master — append-only chronological sort, per-lane scope wins, semantic-edit escalation. Idempotent; supersedes audit-merge-resolve.sh.
---

# lane-merge-resolver

> Skeleton + resolution-rules table authored at PI14-plt-aip A-3a by chief-architect. Full body + backing script land in WI-14.SKILL-4 (Sprint 3 execution).

**Primary caller**: `scrum-master` at PC-2 (branch-merge ops at PI close per `.claude/rules/portfolio-conventions.md` writable-paths table).
**Escalation caller**: portfolio-tier `platform-engineer` for the mechanical bits when LLM judgment is not needed (or when scrum-master delegates).

**Status**: skeleton — DO NOT INVOKE UNTIL WI-14.SKILL-4 [DONE].

**Supersedes**: the never-authored `bin/audit-merge-resolve.sh` referenced in `portfolio/PROCESS.md` 2026-05-11 [PROPOSED] "PI-close ceremony for two-lane model". The script never landed; this skill replaces it with broader scope (all shared-file conflict classes, not just `AUDIT.md`). PROCESS.md annotation update is part of WI-14.SKILL-4's PR.

## Resolution rules (12 file classes)

| File class | Owner / scope | Resolution |
|---|---|---|
| `portfolio/AUDIT.md` | append-only ledger | Mechanical: chronological sort on `YYYY-MM-DD HH:MM` prefix; dedupe identical rows |
| `portfolio/METRICS/tokens.ndjson` | append-only NDJSON | Mechanical: sort by `ts` field; dedupe |
| `portfolio/METRICS/feedback.ndjson` | append-only NDJSON | Mechanical: sort by `ts` field; dedupe |
| `portfolio/PROCESS.md` | both lanes append `[PROPOSED]` deltas | Union — keep both deltas; `[ADOPTED]`/`[REJECTED]` flips disambiguate later |
| `portfolio/PRIORITIES.md` | both lanes may add/re-rank | Semantic: Lane A + Lane B rank-1 rows coexist; flag if both edit the SAME row |
| `portfolio/RUNBOOKS/pi{n}-{lane}-*.yml` | per-lane filename | No collision — pure union (filenames differ) |
| `portfolio/LIBRARIES.md` | both lanes add catalog entries | Union; flag if same row touched |
| `.claude/agents/*.md` | both lanes may edit | Semantic: same-hunk both sides → escalate; otherwise apply both diffs |
| `.claude/skills/*/SKILL.md` | same | Same |
| `.claude/rules/*.md` | both lanes may edit | Same |
| `bin/**` mechanical scripts | both lanes may touch | Semantic: same-file conflicts → escalate; otherwise apply both |
| `portfolio.sh` | rare; both lanes may touch | Semantic: same-line → escalate (this file is chief-architect-sacred) |
| `oppor/**` (per-product) | Lane A scope only | Lane A wins; Lane B touch flagged as lane-violation bug |
| `platform-services/**`, `ai-platform/**`, `portfolio/infrastructure/**` | Lane B scope only | Lane B wins; Lane A touch flagged as bug |
| Anything else | unknown | Escalate — emit chunked diff to operator; no auto-resolve |

## Behavior contract

1. Operates on a `merge-pi{n}-{src-lane}-into-{dst-lane}` branch — never writes directly to `main`.
2. Detects append-only file conflicts via `git diff --merge-base`; the bash layer (`bin/lane-merge-resolve.sh`) handles class 1–3 + 6–7 + 13–14 mechanically.
3. Semantic classes (4–5, 8–12) escalate to the LLM layer (this skill body); the skill reads both sides of each conflicting hunk and applies the rule.
4. Emits one `merge-resolved | <file> | <class> | <hunk-count>` AUDIT row per resolved file (yes — into the very `AUDIT.md` it just resolved; the row lands at the bottom and the next run's sort puts it in chronological place).
5. Lists genuine semantic conflicts in a top-level stdout summary; operator reviews + lands the calls manually.
6. **Idempotent**: re-running on the same merge state is a NOOP except for AUDIT row deduplication.
7. **First run** (per-workspace): proposes adding `.gitattributes merge=union` annotations for `portfolio/AUDIT.md` + `portfolio/METRICS/tokens.ndjson` + `portfolio/METRICS/feedback.ndjson` as a one-time structural fix (reduces future mechanical merges to zero-intervention git operations). Subsequent runs detect existing annotations and skip the proposal.

## Backing script — bin/lane-merge-resolve.sh

Non-LLM layer (per `.claude/rules/runbook-driven-dispatch.md` mechanical-entry pattern):
1. Detect append-only conflicts via `git diff --merge-base`.
2. Apply timestamp-sort + dedupe for AUDIT.md + NDJSON files.
3. Emit a conflict-manifest JSON for the skill's LLM layer to review.
4. Check out the merged result to the working branch.

LLM layer in this skill body handles the semantic classes the script cannot.

## Escalation policy (ADR PI14-plt-aip-0005)

When same-hunk conflicts surface in workspace-tier files (`.claude/agents/*.md`, `.claude/skills/*/SKILL.md`, `.claude/rules/*.md`, `bin/**`), the skill:
- Emits the chunked diff (both sides + base) to stdout
- Tags each conflict `ESCALATE: <file>:<hunk-line-range>`
- Does NOT auto-resolve under any heuristic
- AUDIT row uses verb `merge-escalated` (distinct from `merge-resolved`)

Operator (scrum-master or chief-architect if delegated) reviews each escalation and lands the manual resolution before pushing the merge branch to `main`.

See `portfolio/system-refs/adr/PI14-plt-aip-0005.md` for the full escalation rationale + when LLM judgment is allowed vs. forbidden.

## Reads

- `.claude/rules/runbook-driven-dispatch.md` (mechanical-entry pattern)
- `.claude/rules/portfolio-conventions.md §"Two-lane parallel PI model"` (defines lane scope)
- `portfolio/system-refs/adr/PI14-plt-aip-0005.md` (escalation policy)

## Test fixture

`.claude/skills/lane-merge-resolver/test/` (lands in WI-14.SKILL-4):
- 5 racing append-only entries in AUDIT.md (3 Lane A + 2 Lane B; overlapping timestamps)
- 1 semantic conflict on `.claude/agents/chief-architect.md` (both lanes added a different line to a "Reads" section)
- 1 per-lane code edit (`platform-services/audit/handler.go` Lane B + `oppor/campaigns/handler.go` Lane A — no collision)

Expected: auto-resolve 6 of 7 (append-only sort + per-lane wins); escalate 1 (`.claude/agents/chief-architect.md`) with chunked diff.

## What this skill is NOT

- NOT a general-purpose 3-way merge tool — scope is the two-lane PC-2 ritual only.
- NOT a chief-architect-decision substitute — semantic escalations require human judgment.
- NOT a continuous merge tool — invoked at PC-2 ceremonies (lane B → main, lane A → main; or rebase across lanes mid-PI in rare cases).
