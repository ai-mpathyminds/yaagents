---
name: postmortem-runbook-generator
description: Author portfolio/RUNBOOKS/pi{n}-postmortem.yml at Phase-C step PC-4 — dispatch-ready remediation runbook rolled up from user retro + feedback.ndjson + RETROS/PI{n}.md + AUDIT slice.
---

# postmortem-runbook-generator

Invoked by `chief-architect` once Phase-C PC-1 (retro) and PC-3 (delta ratification) have landed. Produces the postmortem runbook that Phase-C PC-5 then dispatches.

## Inputs (deterministic — all required)

| # | Input | Path | Filter |
|---|-------|------|--------|
| 1 | **User retro** (free-text) | `portfolio/RETROS/PI{n}-user-input.md` | whole file |
| 2 | **Agent friction signal** | `portfolio/METRICS/feedback.ndjson` | `pi == "PI{n}"` |
| 3 | **Scrum-master retro** | `portfolio/RETROS/PI{n}.md` | whole file |
| 4 | **Activity ledger slice** | `portfolio/AUDIT.md` | date window `[PI{n} open … PI{n} close]` |

If any input is missing, **do not author a partial runbook** — emit a Handoff block with `> blocker: <missing input>` and return. Postmortem depends on all four streams to avoid survivorship bias.

## Steps the skill performs

1. **Load user retro.** Parse into three buckets via markdown headings: `What worked well`, `As-is`, `To improve`. If headings absent, fall back to narrative parse — but flag in the runbook preamble as `input_quality: narrative`.
2. **Aggregate feedback.ndjson.** Group by `agent`, `category`, and `AC`. Rank findings by `count × distinct_WIs`. Hotspots surface first. **Sprint grouping** (when sprint context is helpful for the postmortem): bucket each row into its containing sprint by looking up `runbook_entry` against `pi{n}-execution.yml` `sprints[].entries[].id`. **Ungrouped rows handling** (per `runbook-feedback/SKILL.md` §Schema rules + PROCESS 2026-05-03 [ADOPTED]): any phase-B row whose `runbook_entry` does NOT match `^B-\d{2}$` cannot be assigned to a sprint deterministically — collect these into a separate **"Ungrouped feedback rows (schema violation)"** subsection of the postmortem `preamble:` block. Each ungrouped-row entry lists `ts`, `agent`, `wi`, the offending `runbook_entry` value, and a one-line note ("schema violation — does not match `^B-\d{2}$`; row preserved for postmortem accounting but not aggregable into sprint groups"). Ungrouped rows still feed individual entry synthesis at step 5; only the sprint-bucket aggregation skips them. PI10 seed regression cases (for validation tests when this skill is exercised against PI10 retroactively): rows `runbook_entry:"WI-10.K1"` (go-developer, 2026-05-02) and `runbook_entry:"WI-10.NFR-FE-3"` (frontend-developer, 2026-05-03) — both should land in the ungrouped subsection, not in their notional sprint buckets.
3. **Cross-read scrum-master retro.** Extract: `Stale-DoD` findings, `[PROPOSED]` deltas authored this retro, token-burn outliers. These become candidate postmortem entries.
4. **AUDIT slice scan.** Count `governance-finding`, `wi-blocked`, `contract-drift`, and any `handoff` events (routing violations per Phase-B protocol). These are process-health signals, not plan-drift.
5. **Synthesize entries.** Map each distinct friction signal to one postmortem entry. Merge duplicates across streams (e.g. a `help_needed` in `feedback.ndjson` that matches a `contract-drift` AUDIT event becomes **one** entry, not two).
6. **Pick owner per entry** by target-file lane lookup (see `portfolio-conventions.md` lane table):
   - `.claude/agents/*.md` changes → `governance-auditor` (multi-agent edits) or `chief-architect` (portfolio-tier agents only).
   - `.claude/skills/*/SKILL.md` changes or new skills → `chief-architect`.
   - `.claude/hooks/**`, workspace `.claude/settings.json`, `bin/**`, `portfolio.sh`, root `docker-compose.yml` (include graph) → `chief-architect`.
   - `portfolio/PROCESS.md` new proposals → `scrum-master`.
   - `.claude/rules/*.md` edits → `governance-auditor`.
   - `{product}/docs/PI{n}/*.md` retroactive fixes → `{product}-architect` or `platform-engineer` by section.
7. **Set `cwd:` to align with the agent's home** (this is a hard constraint — `bin/dispatch-entry.sh` resolves the agent file by walking up from `cwd` and scanning each `.claude/agents/<basename>.md`):
   - **Workspace-tier agents** (`chief-architect`, `scrum-master`, `governance-auditor`, `ux-architect`, `platform-librarian`) live at `PORTFOLIO_ROOT/.claude/agents/`. Set `cwd: <portfolio-root>`.
   - **Product-tier agents** (`platform-engineer`, `product-manager`, `{product}-architect`, `{lang}-developer`) live at `<product>/.claude/agents/`. Set `cwd: <portfolio-root>/<product>` even when the work touches workspace-root paths — the brief uses absolute paths for the workspace-root touch-points. If multiple products carry the same agent name, the dispatcher rejects the entry as ambiguous; pick a specific product.
   - **Never** pair a product-tier agent with `cwd: <portfolio-root>` — the dispatcher cannot resolve it (PI8 PC-5-01/PC-5-02 precedent, fixed 2026-04-29).
8. **Render** into `portfolio/RUNBOOKS/pi{n}-postmortem.yml` using `portfolio/RUNBOOKS/TEMPLATE-postmortem.yml`.
9. **Append** one line to `portfolio/AUDIT.md` with verb `postmortem-written`.

## Entry schema (per template)

Postmortem entries carry the full execution-runbook shape **plus** three postmortem-specific fields:

| Field | Source | Notes |
|-------|--------|-------|
| `id` | generator | `PC-5-01`, `PC-5-02`, … |
| `agent` | step 6 | Lane-determined owner. |
| `cwd` | step 6 | Repo-relative dir the owner works in. |
| `brief` | generator | ≤3 sentences, references `problem` + `remedy` by name. |
| `status` | generator | `pending` at authorship. |
| `depends_on` | generator | Other postmortem ids if remedy B needs A first. |
| **`problem`** | inputs | One-line statement of the friction signal, with source citation (e.g. `feedback.ndjson rows 42..44`, `AUDIT line 189`). |
| **`remedy`** | synthesis | Array of atomic actions: `[{target: "<path>", change: "<one line>"}]`. |
| **`owner`** | step 6 | Agent file that owns the lane. Redundant with `agent` but kept for grep-friendliness. |

## Ranking and triage rules

| Signal | Promote if | Deprioritize if |
|--------|-----------|-----------------|
| `feedback.ndjson` `category == external-dep` | Appears across ≥2 distinct WIs | One-off with workaround documented |
| `feedback.ndjson` `category == prd` | Cited in user retro `To improve` | Already flagged as `[PROPOSED]` delta |
| `deviations` with repeat `AC` across WIs | Same AC surfaces ≥2 times | One contract-drift already fixed in retro commit |
| Stale-DoD | Any occurrence | (never deprioritize — always an entry) |
| Token-burn outlier (>2× PI mean) | Opus agent in standard/mechanical tier | Declared override in agent frontmatter |
| Governance finding | Any occurrence not yet closed | Already has `[ADOPTED]` delta remedying it |

## Validation rules (asserted before write — all MUST pass)

| # | Rule | Failure mode if missed |
|---|------|------------------------|
| 1 | Every entry's `agent:` field uses the **canonical `.claude/agents/<name>.md` form** (NOT bare names like `agent: chief-architect`). Resolvable from its `cwd:` via the dispatcher's resolver — i.e. `<cwd>/.claude/agents/<basename>` exists, OR an ancestor of `<cwd>` carries it, OR (for workspace-tier agents) it lives at `PORTFOLIO_ROOT/.claude/agents/`. Cross-check by running `bin/dispatch-entry.sh PI{n} <id> --dry-run` on each entry before authorship completes. **Mechanical entries** (`model: none + triggers:`) follow the same canonical form for their nominal agent. | Dispatcher errors with `canonical agent file not found` (bare names skip the upward-search resolver per 2026-04-27 fix) or `agent ambiguous`; entry is undispatchable. |
| 2 | Every entry has `runbook-feedback` in `skills_required`. | Phase-C feedback stream starves; PROCESS violation. |
| 3 | Every entry has `escalate_to: chief-architect`. | Phase-C blockers route to /dev/null. |
| 4 | `depends_on` chain is acyclic and references existing ids. | Dispatch deadlock. |
| 5 | `agent:` model tier matches `.claude/rules/token-budget.md` (chief-architect/ux-architect → opus; sonnet by default; haiku for mechanical roll-ups). | Cost overrun. |
| 6 | `owner:` equals `agent:` (kept redundant for grep-friendliness). | Lane attribution drifts from dispatcher target. |

Any failure halts authorship and surfaces in the Handoff `> blocker:` block. Partial runbook is never written.

## What this skill does NOT do

- Does not dispatch the resulting runbook — user dispatches each entry per PLAYBOOK PC-5.
- Does not author PROCESS.md deltas. If an entry's remedy is a process change, the entry **dispatches scrum-master** to author a `[PROPOSED]` delta; this skill doesn't bypass the PROCESS authorship lane.
- Does not flip `[ADOPTED]` on any existing delta — that's chief-architect's human-supervised PC-3 / PC-6 judgment.
- Does not re-open a closed PI. Remedies land on `main` in the window between PI close and next PI open (Phase-C adoption gate).
- Does not cross into next PI's scope. If a remedy implies new feature work, flag it in the runbook preamble as `defer_to_next_pi_seed` and surface to chief-architect at A-1.

## Exit signal

The skill's output is **complete** when:
- Every feedback.ndjson hotspot has either an entry or an explicit deprioritization note in the runbook preamble.
- Every stale-DoD finding has a remedy entry.
- Every open governance finding has an entry.
- `exit_check:` block lists one assertion per entry ("remedy landed on main" or "[ADOPTED] delta committed") so governance-auditor's PC-6 sign-off is mechanical.

### REQUIRED completion sentinel (skill-execution contract)

Per `.claude/skills/SKILL-EXECUTION-CONTRACT.md` (PC-5-06). The skill's final
line MUST be, verbatim:

`POSTMORTEM-RUNBOOK-GEN: COMPLETE -- entries=<n> rules=PASS|FAIL validation=PASS|FAIL`

(`rules` = the 6-rule validation gate; `validation` = overall.) **Absence = the
skill did NOT run → HARD STOP**: the invoking agent emits `> blocker:
postmortem-runbook-generator produced no COMPLETE sentinel — treat as NOT RUN`,
appends an AUDIT `skill-noop` row, and MUST NOT hand-author
`pi{n}-postmortem.yml` from the template. `validation=FAIL` also halts.
