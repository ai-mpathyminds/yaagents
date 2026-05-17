---
name: retro-generator
description: Produce a PI retrospective file from roadmap status tokens, git log across product repos, and tokens.ndjson — runs stale-DoD cross-check; enforces retro template.
---

# retro-generator

Invoked by `scrum-master` at PI close. Produces a retrospective markdown file adhering to the template in the scrum-master agent spec.

## Steps the skill performs

1. Load `portfolio/PRIORITIES.md` → extract intended PI scope.
2. For each product with `docs/PI{n}/`, load `roadmap.md` + all `{service}.md` files. Extract: WIs, current status tokens, PRD references.
3. In each product repo, run `git log --since=<PI-start-date> --pretty=format:'%H|%ad|%s|%b' --date=short` → parse `Agent:` and `WI:` trailers.
4. Load `portfolio/METRICS/tokens.ndjson` if present; aggregate by `agent` and `product` over the PI window.
5. **Run the feedback.ndjson schema validation pass** (added 2026-05-03 per PROCESS [ADOPTED] "Runbook feedback entry labeling schema"; **authoritative as of PC-5-04 / pi14-plt-aip-postmortem**). Do **NOT** hand-roll a regex here — the prior hand-rolled `^B-\d{2}$` awk diverged from the authoritative write-time `CANON` and made the PI14 retro report 0 violations while 22 existed (the exact evidence/report drift PC-5-04 fixes). Instead, invoke the single source-of-truth validator:
   - `bash bin/validate-feedback-schema.sh --pi PI{n}`  (it extracts the canonical `CANON='...'` from `.claude/skills/runbook-feedback/SKILL.md` §Schema rules — the same pattern the FATAL write-time gate uses — so retro and write-time can never disagree).
   - Exit 0 ⇒ no schema violations. Exit 1 ⇒ ≥1 violation; each `SCHEMA-VIOLATION runbook_entry=… ts=… agent=… wi=…` line it prints becomes one bullet in the retro's "Stale DoD check" section. Exit 2/3 ⇒ tooling error (feedback missing / CANON not found) — surface as a retro blocker, do not silently report "0 violations".
   - The validator also reports `non_utf8_lines` / `malformed_json` counts — copy these verbatim into the retro as data-quality findings (PC-5-05 governance-audit input). It is report-only (never mutates feedback.ndjson); write-time enforcement remains `runbook-feedback`'s FATAL gate.
6. Compute the **stale-DoD diff (structural)**:
   - `claimed_done = {WI : token == [DONE]}`
   - `committed = {WI : any commit in window has WI: trailer}`
   - Report `claimed_done - committed` (claimed but no commit) and `committed - claimed_done` (commit but not marked done).
7. Compute the **stale-DoD semantic check** (added 2026-05-03 per PI10 PC-5-07; PROCESS [ADOPTED] rationale: structural check passes any [DONE] WI whose log carries a `WI: <id>` trailer regardless of whether the commit's content matches the WI's AC — PI10 WI-10.NFR-INFRA-2 commit 6eefa72 is the canonical seed case). For every WI in `claimed_done ∩ committed` (i.e. structurally clean):
   - Open the WI's roadmap entry in `{product}/docs/PI{n}/{service}.md`. Walk from the `### <wi-id>` heading to the next `### ` or EOF — that window is the WI body.
   - Extract **AC tokens** from the body via two regexes (union):
     - **Path tokens**: `(oppor|ai-platform|platform-services|portfolio)/[\w./-]+\.(go|py|ts|tsx|md|yaml|yml|json|sql|sh|toml|mod)` (matches absolute portfolio-relative paths cited in the AC; covers all 4 products + extension set spanning Go/Python/TS/docs/SQL/IaC).
     - **Test-name tokens**: `Test[A-Z]\w*` (Go test functions); `test_[a-z]\w*` (Python pytest functions); `it\(['"][^'"]+['"]` (Vitest/Jest specs — first match captures the test description for substring search).
   - For each `WI: <id>` commit SHA in the PI window, run `git show --stat <sha>` (in the relevant product repo per the WI's `service:`) and collect the changed-files list.
   - **Cross-check rule**: at least one AC path token must appear as a substring in at least one changed-file path, OR at least one AC test-name token must appear as a substring in at least one changed-file path or in `git show <sha>` body. Substring (not exact-equality) accommodates relative-vs-absolute path conventions across the products' nested repo layout.
   - If **zero matches** across all the WI's in-window commits ⇒ surface as a **semantic stale-DoD finding**. Format: `- <wi-id> (<product>/<service>) — [DONE] with commits [<sha1>, <sha2>, ...] carrying WI: trailer; AC cited tokens [<path-or-test-names>] but git show --stat shows no file matching any token. Cross-reference for <product>/<lang>-developer audit.`
   - **Performance budget**: AC-token extraction is regex-only (no LLM call); `git show --stat` is ~10ms per commit; for a 26-WI PI with ~30 commits, the whole pass runs in <1s. Stays in haiku-tier mechanical model — implementation is shell + awk, no model invocation needed for this step.
   - **Additive, not replacement**: structural findings from step 6 still land; semantic findings join them. A WI may surface in BOTH lists (no commit at all + AC-mismatch on a commit elsewhere) — list under both with cross-reference.
8. Render into `portfolio/RETROS/PI{n}.md` using the scrum-master template. Emit step 5 schema violations as bullets under `## Stale DoD check` (`Schema violation: runbook_entry="<value>" (row ts <iso>; agent <agent>; wi <wi>) should match ^B-\d{2}$ per PROCESS 2026-05-03 [ADOPTED]`); emit step 6 structural findings under `## Stale DoD check` (existing convention); emit step 7 semantic findings under a new sibling subsection `## Stale DoD check (semantic)` (always present even if empty — empty section reads "No semantic AC-vs-commit mismatches detected this PI."). If ≥1 schema violation found, draft a [PROPOSED] PROCESS delta in step 9 covering the affected agent (typical remedy: "patch `<agent-file>.md` inline-spec template so its `runbook_entry` value derives from `--field id` of the runbook entry, not the WI id"). If ≥1 semantic finding found, draft a [PROPOSED] PROCESS delta candidate naming the affected WI for next-PI postmortem PC-4 input (do NOT re-open the closed WI from this skill — that's a developer-audit decision, not a retro-generator decision).
9. For each non-trivial finding, draft a proposal in `portfolio/PROCESS.md` under a new `###` heading with token `[PROPOSED]`.
10. Append one line to `portfolio/AUDIT.md` with verb `retro-written`.

## Stale-DoD decision rules

| Finding | Action |
|---------|--------|
| `[DONE]` but no commit | Check roadmap WI for commit reference in body; if missing, flip to `[WIP]` and note in retro |
| Commit with `WI:` but token is `[WIP]` | Flip token to `[DONE]` only if commit body says so explicitly; otherwise flag in retro for dev attention |
| Commit with `WI:` but token is `[DRAFT]` or `[READY]` | Always flag — dev skipped the token lifecycle |
| Commit with no `WI:` trailer | Count as trailer-compliance violation in retro; don't guess WI mapping |
| **Schema violation** in `feedback.ndjson` (any row failing the authoritative `CANON` per `bin/validate-feedback-schema.sh`; NOT a hand-rolled `^B-\d{2}$`) | List one bullet per `SCHEMA-VIOLATION` line the validator prints, in the retro's "Stale DoD check" subsection, naming the row's `agent`, `wi`, `runbook_entry`, and the canonical regex it violated. If ≥1 violation surfaces, draft a [PROPOSED] PROCESS delta for that PI naming the affected agent (typical remedy: patch the agent's inline-spec template so `runbook_entry` resolves from runbook entry id, not WI id). PI10 seed regression: rows `runbook_entry:"WI-10.K1"` and `runbook_entry:"WI-10.NFR-FE-3"` (both go-developer + frontend-developer respectively). |
| **Semantic AC-vs-commit mismatch** (added 2026-05-03 per PI10 PC-5-07) — WI is `[DONE]` AND its commits carry `WI:` trailer (structurally clean) BUT `git show --stat <sha>` for those commits touches zero files matching the AC-cited path tokens / test-name tokens | List one bullet per offending WI under `## Stale DoD check (semantic)` subsection naming the WI id, product/service, commit SHAs, AC-extracted tokens, and the empty-overlap conclusion. Cross-reference the appropriate `<lang>-developer` for next-PI audit; do NOT re-open the WI from this skill (developer audit is a separate authority). PI10 seed regression: WI-10.NFR-INFRA-2 commit 6eefa72 — AC cites `oppor/services/campaigns/pkg/keycloak_idempotency_test.go` + `TestBootstrapIdempotency*` but commit's changed-files list shows neither (PC-5-05 is the corresponding go-developer audit entry). |

## Stale-DoD semantic check — algorithm reference

The semantic check upgrades the structural check from "commit carries `WI:` trailer ⇒ WI met" to "commit carries `WI:` trailer AND commit's changed files overlap WI's AC-cited artifacts ⇒ WI met". Origin: PI10 demonstrated a WI can satisfy the trailer convention (structural pass) without satisfying its AC (e.g. WI-10.NFR-INFRA-2 commit 6eefa72 carries the trailer but the cited idempotency-test artifact is not in the commit's diff).

### Token extraction (regex-only — no LLM)

For each `[DONE]` WI, locate the WI body in `{product}/docs/PI{n}/{service}.md` (heading-line match `^### <wi-id>(:|\s|—|$)`; walk to next `### ` or EOF). Within the body, prefer the AC subsection (heading match `^### Acceptance criteria` OR `^\*\*Acceptance criteria\*\*:` OR `^Criteria:` — any of these conventions used across products); if no AC subsection found, fall back to whole-body scan.

| Token class | Regex | Match scope |
|-------------|-------|-------------|
| **Path** (portfolio-relative) | `(oppor\|ai-platform\|platform-services\|portfolio)/[\w./-]+\.(go\|py\|ts\|tsx\|md\|yaml\|yml\|json\|sql\|sh\|toml\|mod)` | Captures e.g. `oppor/services/campaigns/internal/auth/jwks.go`. Substring-matches against `git show --stat` lines (handles relative-path conventions like `services/campaigns/internal/auth/jwks.go` in commits inside the per-service repo). |
| **Go test name** | `Test[A-Z]\w*` | E.g. `TestBootstrapIdempotency`. Substring-matches against changed-file paths (test files include the test name) AND against `git show <sha>` body (catches inline-defined helpers). |
| **Python test name** | `test_[a-z]\w+` | E.g. `test_bootstrap_idempotency`. Same dual-match strategy. |
| **TS/JS spec name** | `it\(['"]\K[^'"]+(?=['"])` (PCRE — captures spec description) | First-paragraph spec descriptions; substring-match against `git show <sha>` body. |

### Verification call

Per WI, per `WI: <id>` commit in window:

```bash
# Inside the relevant product repo (per WI's service:)
git show --stat <sha> | awk 'NR>1 && /\|/ {print $1}'   # changed-file paths
git show <sha>                                             # full diff for body-substring matches
```

Cross-check: the WI passes the semantic check iff EITHER (a) ≥1 path-token appears as substring in ≥1 changed-file path, OR (b) ≥1 test-name-token appears as substring in ≥1 changed-file path or in the diff body. Else: surface as semantic stale-DoD finding.

### Reporting format

Surface in the retro under `## Stale DoD check (semantic)`. Always emit the section even when empty (empty body reads `_No semantic AC-vs-commit mismatches detected this PI._`). Per finding:

```markdown
- **WI-{id}** ({product}/{service}) — [DONE] with commits [{sha1}, {sha2}] carrying `WI:` trailer; AC cited tokens [{token1}, {token2}, ...] but `git show --stat` for those commits shows no file matching any token. Cross-reference for {product}/{lang}-developer audit; consider as PC-4 postmortem candidate at next PI close.
```

The retro-generator does NOT re-open the WI; that's the developer's audit decision (PC-5 dispatch territory). The retro-generator's job is to surface the discrepancy with enough evidence (commits + tokens + match outcome) for the developer to act.

### Cross-reference: structural vs semantic findings

| Finding location | Detects | Origin |
|------------------|---------|--------|
| `## Stale DoD check` (structural) | `[DONE]` without commit OR commit without `[DONE]` token | original retro-generator (pre-2026-05-03) |
| `## Stale DoD check (semantic)` | `[DONE]` + commit + trailer BUT commit content misses AC | added 2026-05-03 per PI10 PC-5-07 |

A WI may appear in BOTH lists (e.g. zero commits at all + a separate commit elsewhere with empty AC overlap) — list under both with cross-reference. The two checks are complementary, not redundant.

## What this skill does NOT do

- Does not mutate code or PRDs.
- Does not assign blame; it reports patterns.
- Does not set `[ADOPTED]` on its own process deltas.
- Does not cross products in a single commit — scrum-master commits the retro in the portfolio root repo only.

## Exit signal — REQUIRED (skill-execution contract)

Per `.claude/skills/SKILL-EXECUTION-CONTRACT.md` (PC-5-06). This skill's no-op
is the failure that motivated the contract: it once "loaded" but produced no
output and the agent silently hand-authored the retro from the template.

The run is complete **only when** the agent prints, verbatim, as its final line:

`RETRO-GENERATOR: COMPLETE -- pi=PI{n} violations=<v> stale_dod=<s> validation=PASS|FAIL`

(`violations` = count from `bin/validate-feedback-schema.sh`; `stale_dod` =
count of stale-DoD findings.) **Absence of this line means the skill did NOT
run — it is a HARD STOP**: the invoking agent emits `> blocker: retro-generator
produced no COMPLETE sentinel — treat as NOT RUN`, appends an AUDIT `skill-noop`
row, and MUST NOT hand-author `RETROS/PI{n}.md` from the template as a silent
fallback. `validation=FAIL` also halts (do not ship a retro built on a failed
schema/stale-DoD pass).
