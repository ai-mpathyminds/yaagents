---
name: handoff-router
description: Emit standardized Handoff block + append AUDIT.md row at end of any portfolio-agent turn that produced an artifact or moved a WI state. See body for required-callers list.
---

# handoff-router

Every agent ends its turn with two side effects:
1. A `## Handoff` block visible to the user (and usable by `./portfolio.sh next`).
2. One appended line in `portfolio/AUDIT.md` so the trail survives the conversation.

## Required output format

Before ending your turn, print this block at the bottom of your response:

```
## Handoff
next: <agent-name>
artifact: <absolute path to the artifact the next agent should read>
intent: <one sentence — what the next agent should do>
cwd: <optional — directory to cd into before invoking>
```

The four fields are load-bearing. Omit `cwd` only when the next agent runs at the portfolio root.

## Required AUDIT.md entry

Append exactly one line to `"${PORTFOLIO_ROOT:-$(git rev-parse --show-toplevel)}/portfolio/AUDIT.md"` (workspace-relative; resolves from the env var if set, else the repo root — never a hardcoded OS path):

```
<YYYY-MM-DD HH:MM> | <your-agent-name> | <product or "portfolio"> | <event verb> | <artifact path>
```

Event verbs: `seed-written`, `prd-written`, `roadmap-written`, `nfr-appended`, `wi-ready`, `wi-wip`, `wi-done`, `wi-blocked`, `wi-vetoed`, `retro-written`, `process-delta`, `governance-finding`, `handoff`.

Example:
```
2026-04-20 14:32 | chief-architect | platform-services | seed-written | platform-services/system-ref/config/tenant-override.seed.md
```

## When to skip

- Pure read-only turns (user asked a question, you answered). No artifact changed; no handoff owed.
- Interrupt / clarification turns where you're waiting on the user.

## Anti-patterns

- Don't invent a next agent just to "keep the chain going." If the work is done, set `next: <user>` and say so.
- Don't rewrite past AUDIT.md lines. Append-only.
- Don't bundle multiple events into one line. One event per line.
- Don't log a `handoff` event when another verb fits better (e.g. `wi-done` is more useful than a generic `handoff`).
