---
name: token-budget-audit
description: Parse portfolio/METRICS/tokens.ndjson and report per-agent per-PI cost — invocations, mean/peak tokens, output-to-input ratio. Reports only; does not enforce.
---

# token-budget-audit

Aggregate the append-only `portfolio/METRICS/tokens.ndjson` into a per-agent cost view.

## Input record (one NDJSON line per Claude Code session)

```json
{
  "ts": "2026-04-20T14:32:01Z",
  "session_id": "...",
  "agent": "chief-architect",
  "product": "platform-services",
  "model": "claude-opus-4-7",
  "input_tokens": 12543,
  "output_tokens": 2187,
  "wi": "WI-2.3.1"
}
```

Records with missing `agent` are counted under `unknown`. Records with missing `product` count as `portfolio`.

## Steps

1. Load NDJSON → filter by `ts` window (e.g., current PI or current quarter).
2. Group by `agent`. For each group compute: `invocations`, `sum_input`, `sum_output`, `mean_total`, `peak_total`, `output_to_input_ratio = sum_output / sum_input`.
3. Compare each agent's metrics to the budget in `.claude/rules/token-budget.md`.
4. Flag any agent that exceeds its budget on ≥ 2 of: invocation count, mean tokens, or output-to-input ratio.

## Output format

```markdown
## Token-budget audit — <window>

| Agent | Model | Invocations | Mean tokens | Peak | Out/In | Budget? |
|-------|-------|-------------|-------------|------|--------|---------|

**Flagged agents**: <list, or "none">

### Recommendations
- <agent>: <specific action — trim prompt, split role, lower model tier>
```

## Budget reference (from token-budget.md)

- `chief-architect`: ≤5 invocations per PI, opus
- `scrum-master`: 1–3 per PI, haiku
- `governance-auditor`: 1 per quarter + on-demand, sonnet/opus
- `*-architect`: 1–3 per PI per product, opus
- `*-developer`: unbounded but 1 commit per WI; commit count > WI count = suspect
- Helpers: post-step only; free-running invocations = suspect

## What this skill does NOT do

- Does not enforce budgets — reports only.
- Does not modify NDJSON (append-only).
- Does not decide role splits — proposes recommendations for governance-auditor / chief-architect to act on.
- Does not infer missing fields — `unknown` is a legal aggregate bucket.
