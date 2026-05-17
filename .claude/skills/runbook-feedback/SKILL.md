---
name: runbook-feedback
description: Emit one NDJSON line to portfolio/METRICS/feedback.ndjson at end of any turn that closed a Phase-B or Phase-C runbook entry — captures deviations + help-needed signal. Runs alongside handoff-router.
---

# runbook-feedback

Every agent that closes a runbook entry emits **two** turn-end side effects:

1. The existing `handoff-router` block + `AUDIT.md` line (success evidence).
2. One `feedback.ndjson` line (friction signal) — this skill.

The split is deliberate. `handoff-router` answers "what shipped"; `runbook-feedback` answers "where did the plan rub." `postmortem-runbook-generator` at PI close reads the second stream to author the `pi{n}-postmortem.yml` remediation runbook.

## When to invoke (and when NOT to)

**Invoke** when:
- You closed a runbook entry (Phase A, Phase B, or Phase C) — the `status:` on the entry moves from `pending` / `in-progress` to `done` or `blocked`.

**Phase-A scope note** (added 2026-05-09): Phase A IS in scope per `portfolio/PROCESS.md` 2026-04-29 [ADOPTED] "Phase-A telemetry-sinks back-fill duty" — chief-architect MUST verify `feedback.ndjson` fired on every Phase-A station close (A-1, A-1b, A-2, A-3, A-4, A-5, A-5b, A-6). Earlier skill versions excluded Phase A on the grounds that architects emit artifacts not friction signals — PROCESS 2026-04-29 overrode that. PI8 evidence: A-2 product-manager session skipped all three sinks; the back-fill duty is the response. **Empty arrays are valid** for Phase-A rows that genuinely had zero friction.

**Do NOT invoke** when:
- Pure read-only turn (user asked a question).
- Helper-agent verdicts (`code-reviewer`, `test-writer`, `a11y-reviewer`) — they return reviews, not runbook closes.
- You hit a blocker and are escalating to chief-architect — in that case, populate `help_needed[]` richly and still emit; the entry's `status:` is `blocked` and the signal is the point.

**Zero-friction still emits.** If both `deviations` and `help_needed` are empty, the line still lands. Empty-arrays-per-entry is itself useful aggregation signal (agents running clean vs agents hitting friction).

## Required output — one NDJSON line

Append exactly one line to `"${PORTFOLIO_ROOT:-$(git rev-parse --show-toplevel)}/portfolio/METRICS/feedback.ndjson"` (workspace-relative; resolves from the env var if set, else the repo root — never a hardcoded OS path):

```json
{"ts":"2026-04-24T16:45:12Z","pi":"PI6","runbook_entry":"B-14","wi":"WI-6.A11","agent":"go-developer","deviations":[{"WI":"WI-6.A11","AC":"tooling-api /execute endpoint","reason":"endpoint did not exist; used /test per PI4 convention"}],"help_needed":[]}
```

### Envelope fields (always present)

| Field | Type | Notes |
|-------|------|-------|
| `ts` | string (ISO 8601 UTC) | When the entry closed. |
| `pi` | string | `PI6`, `PI7`, … |
| `runbook_entry` | string | The `id:` from the runbook (e.g. `B-14`, `PC-5-01`). |
| `wi` | string | The `WI:` id — or `n/a` for planning / postmortem entries that don't map to a WI. |
| `agent` | string | Your agent name, not file path. |
| `deviations` | array | May be empty `[]`. See shape below. |
| `help_needed` | array | May be empty `[]`. See shape below. |

### `deviations[]` object shape

Use one entry per distinct plan-vs-reality gap. Keep each reason under ~80 chars — if it needs a paragraph, the deviation is actually a blocker; route to `help_needed` instead.

| Field | Type | Notes |
|-------|------|-------|
| `WI` | string | WI that surfaced the drift (may differ from envelope `wi` if a neighbouring WI's plan was discovered wrong during this entry). |
| `AC` | string | Acceptance-criterion id or a short name for the part of the brief that drifted. |
| `reason` | string | One line: observed-vs-planned. No blame, no narrative. |

### `help_needed[]` object shape

Use one entry per unresolved blocker *at close time*. If you unblocked something during the turn, it does **not** go here — that's a `deviation` at most.

| Field | Type | Notes |
|-------|------|-------|
| `category` | enum | `external-dep` \| `process` \| `prd` \| `infra` \| `skill` \| `tool` |
| `blocker_signal` | string | One line: what role or artifact you need. |

## Schema rules (write-time validation — **FATAL as of 2026-05-09**)

Per `portfolio/PROCESS.md` 2026-05-03 [ADOPTED] + 2026-05-09 [REVISED]. The `runbook_entry` field is the field consumers (`retro-generator`, `postmortem-runbook-generator`) group by — inconsistent labels block deterministic aggregation.

**Policy revision (2026-05-09)**: this rule was originally non-fatal ("warn but always write — losing the friction signal is worse than allowing a malformed label"). PI13 evidence forced the revision: **27 of 55 PI13 feedback rows carried non-canonical values** (e.g. `B-OBS-1`, `WI-13.E2E-1`, `B-IAM-2`, `B-OPPOR-CHAN-A`, `A-1-process-delta`) — sprint-bucketing impossible without manual lookup; `postmortem-runbook-generator` aggregation surfaces ~half the PI rows as "Ungrouped". The warn-only stance was load-bearing in name only. **New stance: refuse the write; caller MUST amend before retry.** The skill returns exit 1 on schema violation; the bash idiom below uses an `if` guard so the file is NOT touched on violation. Losing-the-occasional-signal is the lesser harm vs aggregation-blocking drift across half the PI's rows.

### Allowed `runbook_entry` values

| Phase | Regex | Examples (allowed) | Examples (DISALLOWED — refuse write) |
|-------|-------|--------------------|--------------------------------------|
| **Phase A** (planning-runbook entries; in scope since PROCESS 2026-04-29 telemetry back-fill duty) | `^A-[1-9]\d*[a-z]?$` (MUST match — strict; non-zero-padded outer N) | `A-1`, `A-1b`, `A-2`, `A-3`, `A-4a`, `A-4b`, `A-5`, `A-5b`, `A-6`, `A-10`, `A-12c` | `A-1-process-delta` (use `A-1` + meta-row context in `deviations[]`), `a-1` (lowercase), `A-01` (zero-padded — Phase A is non-zero-padded by convention; Phase B is zero-padded), `A-4a-meta` |
| **Phase B** (execution-runbook entries) | `^B-\d{2}$` (MUST match — strict) | `B-01`, `B-02`, `B-26`, `B-54` | `B-OBS-1` (use the numeric entry id — for OBS-1 in pi13-execution.yml that is `B-51`), `WI-13.K1` (this is a WI id; goes in the `wi:` field), `B-1` (not zero-padded), `b-01`, `B-100` (3 digits), `B-IAM-2`, `B-OPPOR-CHAN-A` |
| **Phase C** (postmortem-runbook entries) | `^PC-\d+(-\d+)?$` (MUST match if PC-shaped) **OR** lowercase-kebab descriptive `^[a-z][a-z0-9-]{2,}$` AND NOT `^(a-\d|b-\d|pc-\d)` (kebab veto: typo'd canonical) | `PC-5-01`, `PC-5-02`, `g6-smoke`, `infra-kc-hostname`, `exit-check`, `mid-pi-user-review` | `WI-13.K1` (WI id), `b-01` (lowercase typo of B-01), `pc-5-01` (lowercase typo), `PC5-01` (missing hyphen), `a-4a` (lowercase typo of A-4a) |
| **Meta / non-entry contexts** (skill-shipped, agent-authored, retired-row notes) | `^n/a$` (literal string) | `n/a` | anything else when the row is genuinely not tied to a runbook entry |

**Phase-A / Phase-B / Phase-C asymmetry rationale**: Phase A is low-volume + station-named (A-1..A-6 + letter suffixes for splits) — discipline pays off cheaply. Phase B is mechanical + high-volume (one row per WI close, dozens per PI) — discipline pays off most. Phase C is human-edited + low-volume (entries authored by chief-architect at PC-4 named for what they do) — descriptive labels are clearer than synthetic IDs, so kebab is allowed alongside PC-N-NN.

**Common mistake — WI in `runbook_entry`**: if your row's `runbook_entry` looks like `WI-N.<something>`, you're filling the wrong field. The WI ID goes in the `wi:` field; the runbook entry's ID (e.g. `B-25`, `A-4a`) goes in `runbook_entry:`. The skill validator emits a specific "WI-shaped value in runbook_entry — this is the wi: field's content" error for this case.

### Phase detection (which regex applies)

The phase is derived from the runbook the entry comes from, NOT from the row itself. Convention: if you closed an entry from `pi{n}-planning.yml` ⇒ phase A; if from `pi{n}-execution.yml` ⇒ phase B; if from `pi{n}-postmortem.yml` ⇒ phase C. When in doubt, look at the entry's `id:` in the source runbook — `A-N` / `A-Nb` / `A-Na` is phase A; `B-NN` is phase B; `PC-N-NN` or kebab-descriptive is phase C.

### Write-time validation idiom (copy into agent prompts via inline-spec)

The idiom below uses a guard that REFUSES the append on violation. The caller (the agent's turn-end script) MUST amend the entry value and retry. The non-fatal `printf >> ... ; warn` shape used pre-2026-05-09 is RETIRED.

```bash
# Build the JSON line first
LINE='{"ts":"<ISO8601-UTC>","pi":"<PI>","runbook_entry":"<ID>","wi":"<WI or n/a>","agent":"<your-agent>","deviations":[],"help_needed":[]}'

# Extract runbook_entry value for validation
ENTRY=$(printf '%s' "$LINE" | sed -n 's/.*"runbook_entry":"\([^"]*\)".*/\1/p')

# Canonical regex (union of all 4 valid phases). Phase A + PC outer N are
# non-zero-padded ([1-9]\d*) so 'A-01' / 'PC-05' fail; Phase B is strictly
# zero-padded 2-digit (B-NN). Kebab arm covers Phase-C descriptive labels.
CANON='^(A-[1-9][0-9]*[a-z]?|B-[0-9]{2}|PC-[1-9][0-9]*(-[0-9]+)?|[a-z][a-z0-9-]{2,}|n/a)$'

if ! printf '%s' "$ENTRY" | grep -qE "$CANON"; then
  # Structured WARN on stderr — caller must amend ENTRY and retry.
  printf 'WARN runbook-feedback schema violation: runbook_entry="%s" does not match canonical regex %s (per .claude/skills/runbook-feedback/SKILL.md §Schema rules; fatal as of 2026-05-09).\n' "$ENTRY" "$CANON" >&2
  if printf '%s' "$ENTRY" | grep -qE '^WI-'; then
    printf 'WARN  hint: value is WI-shaped — this is the `wi:` field content. Put the numeric runbook entry id (B-NN, A-N, PC-N-NN) in `runbook_entry:`.\n' >&2
  fi
  printf 'WARN  hint: REFUSED WRITE. Amend the LINE and retry. See SKILL.md §Schema rules for the full allowed-values table.\n' >&2
  exit 1   # FATAL — caller must amend before retry
fi

# Kebab-veto: the kebab arm of $CANON accepts lowercase strings that LOOK like
# typo'd canonical forms (b-01, pc-5-01, a-4a). Reject those — they are almost
# always a developer typo, not an intentional Phase-C label.
if printf '%s' "$ENTRY" | grep -qE '^(a-[0-9]|b-[0-9]|pc-[0-9])'; then
  printf 'WARN runbook-feedback kebab-veto: runbook_entry="%s" looks like a lowercase typo of a canonical Phase A/B/C form. Use uppercase (A-1 / B-01 / PC-5-01) for canonical IDs; Phase-C descriptive labels must NOT begin with a-/b-/pc- followed by a digit.\n' "$ENTRY" >&2
  exit 1
fi

# Write the line (canonical only path)
printf '%s\n' "$LINE" >> portfolio/METRICS/feedback.ndjson
```

**Fatal by design (revised 2026-05-09)**: PI13's 27/55 (49%) non-canonical-rows evidence retired the non-fatal stance. The cost of one refused-write (caller amends ~10 chars and retries — sub-minute friction) is far below the cost of postmortem-runbook-generator failing to aggregate half a PI's rows.

### Historical drift (pre-2026-05-09 rows)

Rows already in `feedback.ndjson` with non-canonical `runbook_entry` values are **read-only history** — do NOT retroactively rewrite. Consumers (`retro-generator`, `postmortem-runbook-generator`) continue to surface them under "Schema violation" / "Ungrouped feedback rows" sections; that section will shrink to zero from PI14 onward as the fatal gate prevents new drift.

### What violations look like in consumers

- `retro-generator` (PI close): violating phase-A/B rows appear in `RETROS/PI{n}.md §Stale DoD check` as `Schema violation: runbook_entry="<value>" should match <canonical regex>`. Each violation surfaces a [PROPOSED] PROCESS delta candidate.
- `postmortem-runbook-generator` (PC-4): violating rows go into a separate "Ungrouped feedback rows (schema violation)" subsection in the postmortem preamble — they don't get grouped into sprints because the label can't resolve to a sprint id.

## Aggregation contract (for consumers)

`postmortem-runbook-generator` reads this file, filters by `pi`, groups by `agent` + `category`, and ranks `(count × distinct_WIs)` to surface the top friction hotspots. Keeping each line **small and structured** (not prose) is what makes that aggregation possible.

## Append, don't pretty-print

One JSON object, one line, no trailing whitespace, no array wrapper. Match the `tokens.ndjson` convention under the same directory.

## What this skill does NOT do

- Does not replace `handoff-router`. Both run at turn end.
- Does not summarize success. Success lives in `AUDIT.md` `wi-done` lines.
- Does not write free-text commentary. If you catch yourself typing prose into `reason` or `blocker_signal`, re-phrase as a signal.
- Does not mutate the runbook. `status:` flips on the entry are the runbook-dispatcher's job, not this skill's.
- Does not cross PIs. One file grows append-only per portfolio; consumers filter by `pi` field.
