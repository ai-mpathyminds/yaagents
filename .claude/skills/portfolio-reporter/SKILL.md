---
name: portfolio-reporter
description: Author portfolio/REPORTS/* pitch decks (customer/VC/investor), PI metrics reports (DORA + tokens + WI velocity), and ad-hoc PI history Q&A from git+AUDIT+METRICS. Reports only.
---

# portfolio-reporter

Reporting skill — three modes, one entry point. Wraps the existing `portfolio/REPORTS/gen_*.py` + `measure_dora.py` generators (do not re-implement them) and the existing `token-budget-audit` / `portfolio-diff` skills. Read-only on every input; the only writes are to `portfolio/REPORTS/**` and the AUDIT row.

## When to invoke

| Trigger | Mode | Typical caller |
|---|---|---|
| "produce a pitch deck for <audience>" / "refresh the investor deck" / "warm-circle flyer" | `pitch` | chief-architect (user-direct) |
| "PI{n} metrics report" / "what did PI{n} cost / ship / break" | `pi-metrics` | chief-architect, scrum-master (at retro) |
| ad-hoc Q on PI history / agent cost / velocity / blockers | `pi-qa` | chief-architect, user-direct |

## Invariants (read before any mode)

- **Read-only on the system under report.** Never modifies WIs, roadmaps, ADRs, PRDs, runbooks, or service code. Writes only under `portfolio/REPORTS/**` (and one AUDIT row).
- **No fabrication.** Every claim cites a file path + line range, a commit SHA, an AUDIT row, an NDJSON ts, or a chart preimage. "Looks like ~30 commits" without a `git log | wc -l` is forbidden.
- **One question class per Q&A invocation budget cell** — see §pi-qa budget at the bottom.
- **Generator scripts are versioned content, not parametrized CLIs.** The current `gen_*.py` files hard-code their output filename + date inside `main()`. Treat each new run as: (1) bump the version/date string at the bottom of the file, (2) run, (3) commit both the source bump and the output.
- **ASCII-only in sentinel + console output** (smart quotes / em-dash render as `?` on the Windows console — already a real defect class this PI).

## Mode A — `pitch`

### Audience → generator routing

| Audience | Default generator(s) | Notes |
|---|---|---|
| VC / investor (long-form) | `gen_proposal_deck.py` + `gen_exec_summary.py` | 17-slide deck + 2-page exec PDF as leave-behind |
| Enterprise customer / CTO (1-hour meeting) | `gen_proposal_deck.py` + `proposal-handout-*.pdf` (emitted by same script) | Per the v6 frame: 10-min pitch + 50-min demo |
| Warm-circle / peer / lightweight intro | `gen_flyer.py` | 1-page flyer; no metrics-heavy framing |
| Internal review / portfolio readout | `gen_exec_summary.py` only | PI process exec summary, no sales surface |

If the requested audience does not fit, ask one clarifying question before generating — do not pick a default silently.

### Run protocol (per generator chosen)

1. **Confirm inputs are fresh.** Run `python portfolio/REPORTS/measure_dora.py` first if any DORA chart will be embedded AND `dora_measured.csv` is older than the most recent PI close in `portfolio/RETROS/PI*.md`. (Without this, the deck embeds stale DORA numbers and a customer / VC catches it.)
2. **Bump the version string** at the bottom of the chosen generator's `main()`:
   - `gen_proposal_deck.py` → `proposal-deck-<YYYY-MM-DD>-v<N+1>.pptx` and `proposal-handout-<YYYY-MM-DD>-v<N+1>.pdf` (next sequential version against the existing `proposal-deck-*.pptx` files in `portfolio/REPORTS/`)
   - `gen_exec_summary.py` → `exec-summary-<YYYY-MM-DD>.pdf`
   - `gen_flyer.py` → `flyer-<topic>-<YYYY-MM-DD>.pdf` (preserve the topic slug; today's slug is `agent-onboarding`)
3. **Close any open `.pptx` in PowerPoint first** — Office holds a write lock; the script will fail with `PermissionError` on the `.pptx` write. The `~$proposal-deck-*.pptx` files in `portfolio/REPORTS/` are the lock turds left behind; if present, PowerPoint still has the file open.
4. **Run the generator** from the workspace root:
   ```
   python portfolio/REPORTS/gen_proposal_deck.py
   python portfolio/REPORTS/gen_exec_summary.py
   python portfolio/REPORTS/gen_flyer.py
   ```
   Expect `OK: <path>` lines for each artifact. Halt on any non-zero exit; do NOT emit a `validation=PASS` sentinel against a failed run.
5. **Spot-check the chart preimages** that were regenerated alongside (`_agents.png`, `_dora.png`, `_velocity.png`, `_radar.png`, `_quality.png`, `_roadmap.png`, `_lifecycle.png`). If any preimage is older than its source script's `git log -1`, regeneration was skipped — investigate before claiming PASS.
6. **Open the output** and visually confirm: cover page audience-correct, no `2026-05-21` dates leaking into a fresh 2026-05-26 deck, no `[Company]` / `[CTO]` placeholders unfilled.

### Single-product variant

The current `gen_proposal_deck.py` is portfolio-scoped (7 product cards on the "The portfolio behind these numbers" slide). For a **single-product** pitch:

- **Today**: edit the `PRODUCT_CARDS` list in `gen_proposal_deck.py` to the target product only (e.g. `yaagents` alone) + edit the cover + the demo-script slide; bump version to `v<N+1>-<product>`; run.
- **Future**: if single-product pitches become recurring, file a `[READY]` WI to parametrize `gen_proposal_deck.py` with `--product <name>` and a `PRODUCT_CARDS` config dict. Do NOT inline-fork the script into `gen_proposal_deck_<product>.py` — that creates a duplication-scanner finding the moment the second one lands.

### AUDIT + commit

After generation:
- `git add portfolio/REPORTS/proposal-deck-*-v<N+1>.pptx portfolio/REPORTS/proposal-handout-*-v<N+1>.pdf portfolio/REPORTS/gen_proposal_deck.py portfolio/REPORTS/_*.png`
- Commit message: `report(pitch): proposal deck v<N+1> — audience=<vc|customer|warm>, PI window through PI<n>` with trailers `Agent: chief-architect`, `WI: n/a` (reports have no WI; rationale string in body cites this).
- AUDIT verb: `pitch-generated`.

## Mode B — `pi-metrics`

Single PI metrics report → `portfolio/REPORTS/PI<n>-metrics.md` (new convention; PI-scoped, distinct from the date-stamped exec summary).

### Inputs (read-only)

1. `portfolio/RETROS/PI<n>.md` — "What shipped" table; PI close date.
2. `portfolio/AUDIT.md` — `wi-wip`, `wi-done`, `wi-blocked`, escalation rows for the PI window.
3. `portfolio/METRICS/tokens.ndjson` — filter `pi == "PI<n>"` (note: some entries carry `pi=PI<n>-<lane>` after the two-lane model adopted at PI14).
4. `portfolio/METRICS/feedback.ndjson` — filter `pi` field same way.
5. `portfolio/REPORTS/dora_measured.csv` — re-run `measure_dora.py` first if PI is the current or just-closed one; otherwise the CSV row is canonical.
6. Per-product `git log` for the PI window — bracketed by branch creation at A-5 (pi-open) and PC merge to main, or by ISO date range from RETROS.
7. PI roadmap docs `{product}/docs/PI<n>/*.md` — `[DONE]` vs `[BLOCKED]` token counts.

### Steps

1. **Compute the PI window.** ISO start = first `pi-open` AUDIT row for PI<n>; ISO end = either `phase-c-closed` AUDIT row OR (if PI is open) "now".
2. **Velocity.** Count: WIs `[DONE]`, WIs `[BLOCKED]`, WIs `[VETOED]`, total commits with `WI:` trailer (sum across all product repos via `portfolio-diff` skill — invoke it, don't re-implement).
3. **DORA.** Read the PI's row from `dora_measured.csv` (do NOT recompute the formulas; trust `measure_dora.py`). Cite confidence flags verbatim.
4. **Cost.** Invoke `token-budget-audit` skill scoped to `pi=PI<n>`. Sum input + output + cache_read tokens; multiply by per-model unit cost (cite the unit cost source in the report; budget-side-of-truth lives in `.claude/rules/token-budget.md`). Roll up per-agent.
5. **Quality signals.** From `feedback.ndjson`: count deviations + `help_needed` entries; classify by `category` (infra/tool/contract/scope). From AUDIT: count escalations (`> blocker:` rows), `skill-noop` rows, `commit-skip` rows, governance vetoes.
6. **Carry-forward.** Cross-check `RETROS/PI<n>.md §Carry-forward` against `INTAKE/PI<n+1>-intake.md` — flag any carry-forward item not honored in the next PI's intake.

### Output format (markdown)

```markdown
## PI<n> metrics report

Window: <ISO start> → <ISO end> (<duration days>)
Status: <open | closed at <PC-6 date>>

### Velocity
| metric | value | source |
|---|---|---|
| WIs [DONE] | <n> | <product>/docs/PI<n>/*.md |
| WIs [BLOCKED] | <n> | … |
| WIs [VETOED] | <n> | … |
| commits with WI: trailer | <n> | portfolio-diff PI<n>-start → PI<n>-end |

### DORA (from dora_measured.csv)
| metric | value | confidence |
|---|---|---|
| Deployment Frequency | <x>/day | HARD/SOFT/NA |
| Lead Time | <h>h | … |
| Change Failure Rate | <%>% | … |
| MTTR | <h>h | … |

### Cost
| agent | invocations | input tokens | output tokens | cache_read | est USD | budget? |
|---|---|---|---|---|---|---|
| chief-architect | <n> | … | … | … | $<x> | <PASS|OVER> |

### Quality signals
- Escalations (`> blocker:`): <n> — top causes: <cat1, cat2>
- skill-noop rows: <n>
- commit-skip rows: <n>
- governance vetoes: <n>
- feedback.ndjson deviations: <n> across <m> entries

### Carry-forward audit
- From RETROS/PI<n>.md: <list>
- Honored in INTAKE/PI<n+1>-intake.md: <list>
- Slipped: <list — flagged for next retro>
```

### AUDIT + commit
- File: `portfolio/REPORTS/PI<n>-metrics.md`. Commit trailers `Agent: <caller>`, `WI: n/a`. AUDIT verb: `pi-metrics-generated`.

## Mode C — `pi-qa`

Ad-hoc Q&A grounded in primary sources. Each answer cites at least one path:line, commit SHA, NDJSON ts, or AUDIT row. No answer without a cite.

### Recipe library

Each row = one question class + its exact resolution incantation. Add new rows here when a novel question class is asked twice.

| Q class | Source | Incantation |
|---|---|---|
| "Which agent burned the most tokens in PI<n>?" | `tokens.ndjson` | `grep '"pi":"PI<n>' portfolio/METRICS/tokens.ndjson \| jq -s 'group_by(.agent) \| map({agent: .[0].agent, in: (map(.input_tokens) \| add), out: (map(.output_tokens) \| add), cache: (map(.cache_read) \| add)}) \| sort_by(-.out)'` — note `pi` field is `PI<n>` pre-PI14 and `PI<n>-<lane>` for PI14+, so use `"pi":"PI<n>` as a prefix match. |
| "How long did WI-<id> stay in [WIP]?" | `AUDIT.md` | `grep -E 'WI-<id>\|wi-(wip\|done\|blocked)' portfolio/AUDIT.md` → diff first wi-wip ts vs first wi-done ts |
| "Which sprints overran capacity in PI<n>?" | execution runbook + AUDIT | `portfolio/RUNBOOKS/pi<n>-execution.yml` sprint_size vs AUDIT wi-done rows per sprint window |
| "What was the most common blocker category in PI<n>?" | `feedback.ndjson` | `grep '"pi":"PI<n>' portfolio/METRICS/feedback.ndjson \| jq -c '.help_needed[]?.category' \| sort \| uniq -c \| sort -rn` |
| "Which product shipped the most code in PI<n>?" | git log per product | Invoke `portfolio-diff` skill with `since=PI<n>-start` |
| "Which WIs are carry-forward candidates for PI<n+1>?" | roadmap docs + RETROS | `grep -lE '\[BLOCKED\]\|\[WIP\]' {product}/docs/PI<n>/*.md` cross-check `RETROS/PI<n>.md §Carry-forward` |
| "Where did chief-architect overspend its ≤5-invocations/lane PI budget?" | `tokens.ndjson` | `grep '"agent":"chief-architect"' tokens.ndjson \| jq -c '{ts,pi,model}' \| group by pi+lane`; flag any group with count > 5 |
| "How many integration tests silently skipped in PI<n>?" | git log + grep gate | `git log --since=<PI start> --until=<PI end> -p -- '*_integration_test.go' \| grep -c 't.Skip'` (per `.claude/rules/integration-test-discipline.md`) |
| "What did skill X cost us when it no-op'd?" | AUDIT.md | `grep 'skill-noop\|<skill-name>' portfolio/AUDIT.md` (skill-execution-contract failure pattern) |
| "How many TF resources are drifted vs ground truth?" | cloud-state-grounder report | `portfolio/REPORTS/cloud-state/PI<n>-alignment-matrix.md` → count `DRIFT-FAIL` cells |

### Recipe additions

When a question doesn't fit an existing row:
1. Run the most plausible incantation; cite the result.
2. If the question is likely to recur, append a new row to the table in this SKILL.md in the same turn (chief-architect lane).
3. If the question requires cross-cutting data not in any single file, escalate to a `[READY]` WI proposing a derived snapshot — do NOT silently fabricate.

### Q&A budget

- Soft cap: **≤ 5 questions per invocation** before suggesting `/compact` or `fresh` session-strategy. Beyond that the context bloats and quality drops (per `.claude/rules/agent-context-discipline.md` Layer 2).
- Every answer ends with `Source: <cite>`. Multi-source answers list each cite on its own line. **No cite, no answer** — `> blocker: cannot ground answer in primary source` instead.
- pi-qa mode does NOT write any artifact (the answer goes back to the caller); the sentinel still fires.

## Exit signal — REQUIRED

Per `.claude/skills/SKILL-EXECUTION-CONTRACT.md`. Final line of every invocation, verbatim, ASCII only:

```
PORTFOLIO-REPORTER: COMPLETE -- mode=<pitch|pi-metrics|pi-qa> artifact=<path|n/a> pi=<PI<n>|n/a> questions=<n|n/a> validation=PASS|FAIL
```

- `mode` — exact value chosen.
- `artifact` — output path for `pitch` / `pi-metrics`; `n/a` for `pi-qa`.
- `pi` — PI scope of the report; `n/a` for portfolio-wide pitch decks that span all PIs.
- `questions` — count of Q&A questions answered this invocation; `n/a` for `pitch` / `pi-metrics`.
- `validation` — `PASS` iff (a) every claim in the artifact / answer carries a cite, (b) every generator exited 0, (c) no `~$*.pptx` lock turds left behind, (d) for `pi-metrics` every input source listed in §Inputs was actually read. `FAIL` otherwise — caller halts per skill-execution contract §2.

## Handoff + AUDIT

Every invocation ends with:
1. The `PORTFOLIO-REPORTER: COMPLETE …` sentinel above.
2. A `## Handoff` block per the chief-architect agent body (4 fields).
3. One `portfolio/AUDIT.md` append per the matching verb:
   - `pitch-generated` (Mode A)
   - `pi-metrics-generated` (Mode B)
   - `pi-qa-answered` (Mode C; one row per session, not per question)
4. `runbook-feedback` NDJSON line iff this invocation closed a Phase-B or Phase-C runbook entry (per `runbook-feedback` skill).

## What this skill does NOT do

- ❌ Does NOT re-implement the generator scripts. They live at `portfolio/REPORTS/gen_*.py` and `measure_dora.py`; this skill orchestrates them.
- ❌ Does NOT compute DORA formulas — `measure_dora.py` is the only source of truth; reading `dora_measured.csv` is the only legal path.
- ❌ Does NOT replace `retro-generator`, `token-budget-audit`, `portfolio-diff`, or `postmortem-runbook-generator` — those are PI-cycle-binding artifacts; this skill is for external-facing or ad-hoc reporting.
- ❌ Does NOT write outside `portfolio/REPORTS/**` + the AUDIT row.
- ❌ Does NOT answer a Q&A question without a citable source. "I think roughly N commits" is forbidden; if the data isn't grounded, the answer is `> blocker: <reason>`.
- ❌ Does NOT silently fall back to manual artifact authorship if a generator no-ops — emit the sentinel `validation=FAIL` with the no-op reason, per skill-execution contract §2.
- ❌ Does NOT enforce token budgets (that is `governance-auditor`'s lane via `token-budget-audit`).
- ❌ Does NOT branch / tag / push. Reports stay local until the caller commits them.
