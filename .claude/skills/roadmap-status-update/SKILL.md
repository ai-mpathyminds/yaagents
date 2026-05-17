---
name: roadmap-status-update
description: Mechanically flip a WI status token ([READY]/[WIP]/[DONE]/[BLOCKED]) in a per-service PI roadmap doc per .claude/rules/status-tokens.md transition grammar.
---

# roadmap-status-update

Every developer agent that closes a WI emits three side effects beyond the commit:
1. Status-token flip on the WI's roadmap-doc heading line — this skill.
2. `handoff-router` block + `AUDIT.md` line.
3. `runbook-feedback` line.

This skill owns side-effect 1; the other two are separate skills invoked at the same turn close.

## Inputs (caller provides)

| Param        | Required | Example                               |
|--------------|----------|---------------------------------------|
| `doc_path`   | yes      | `ai-platform/docs/PI11/agent-api.md`  |
| `wi_id`      | yes      | `WI-11.A4`                            |
| `new_token`  | yes      | `WIP` \| `DONE` \| `BLOCKED` (no brackets in input; the skill writes `[<TOKEN>]`) |

## Steps the skill performs

1. **Locate the WI heading.** Grep for `^### {wi_id}\b` in `doc_path` (anchor at start-of-line; word-boundary after the id so `WI-11.A4` does not match `WI-11.A40`). Capture the matched line. If 0 matches → STOP and surface as `> blocker: roadmap-status-update: WI <wi_id> not found in <doc_path>`. If ≥2 matches → STOP and surface as `> blocker: roadmap-status-update: <wi_id> appears <N> times in <doc_path>; expected 1`.
2. **Verify legal transition** per `.claude/rules/status-tokens.md` §Transitions:
   - First commit: existing token MUST be `[READY]`; new token `[WIP]`.
   - Test-gate pass: existing token MUST be `[WIP]`; new token `[DONE]`.
   - Detour: existing token in `{[READY],[WIP]}` MAY transition to `[BLOCKED]`; the caller is responsible for the `> blocked by:` blockquote in the WI body.
   - Unblock: existing `[BLOCKED]` MAY return to `[READY]` (caller responsibility — typically governance or chief-architect lane, NOT developer).
   Illegal transition (e.g. `[READY]→[DONE]`, `[DRAFT]→[WIP]`, any agent flipping `[VETOED]`) → STOP and surface in the agent's `## Handoff` block as a deviation; do NOT write. Caller must escalate.
3. **Edit the file.** Use the Edit tool to replace the existing `[<OLD>]` token with `[<NEW>]` on the matched heading line. The heading text + WI id + token together make the `old_string` unique within the file. Do NOT use `replace_all`.
4. **Verify the flip.** Re-grep for `^### {wi_id}\b` in `doc_path`; confirm the new token is present and the old token is absent on that line.

## What this skill does NOT do

- Does not commit. Status-token flips ride along on the next commit the developer makes (the commit that closes the WI), but the skill itself just edits the markdown file.
- Does not author the WI's DoD log line. The developer writes DoD prose into the WI body, not the heading.
- Does not flip across files. One WI lives in one `{service}.md` file; cross-service WIs do not exist by convention (cross-service work is split into sibling WIs, one per `{service}.md`).
- Does not handle `[VETOED]` flips — those are `governance-auditor`'s lane only (per `.claude/rules/status-tokens.md` §Transitions). Calling this skill with `new_token: VETOED` from a developer agent → STOP with `> blocker: VETOED is governance-auditor lane`.
- Does not append to `AUDIT.md`. That is `handoff-router`'s job (verb `wi-wip`, `wi-done`, `wi-blocked`).

## Why this skill is mechanical

The token grammar is the contract between developer commits and `scrum-master`'s stale-DoD check at PI close. A drift anywhere (typo in token, illegal transition, file not flipped despite commit) surfaces as a stale-DoD finding in the retro. Making this a skill — not free-form Edit — means the rule lives in one place and every developer agent calls it the same way. Per the 2026-Q2 special audit BLOCKER-3a: 6 developer agents reference this skill; until this file existed, every WI close attempted a non-existent-skill invocation, leaving tokens at `[READY]` despite green commits — exactly the stale-DoD pattern the retro keeps flagging.

## Caller pattern (developer agent — copy verbatim)

After committing a WI:
- First commit: invoke `roadmap-status-update` with `new_token: WIP`.
- Test-gate pass: invoke `roadmap-status-update` with `new_token: DONE`.
- External blocker hit mid-WI: invoke with `new_token: BLOCKED` AND add `> blocked by: <reason>` blockquote to the WI body in the same edit pass.
