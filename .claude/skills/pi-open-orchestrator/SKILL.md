---
name: pi-open-orchestrator
description: A-5 pi-open ceremony (scrum-master+operator): run portfolio.sh pi-open, solo-mode residue reconcile (gitignore junk, commit orphan work w/ original-owner trailer), then branch+tag.
---

# pi-open-orchestrator

> Authored at PI16-plt-aip A-5 (mid-execution escalation) by chief-architect after
> the `model: none` mechanical A-5 entry dead-halted on dirty-tree residue it had no
> judgment to triage. Revised same PI to the **solo-mode single-agent** residue model
> (user-direct) — see "Collaboration mode". Wraps `./portfolio.sh pi-open` as an
> agent-facing ceremony per `.claude/rules/runbook-driven-dispatch.md` §"Agents may
> use skills" — the symmetric counterpart to that rule's hypothetical
> `pi-close-orchestrator`.

**Primary caller**: `scrum-master` at planning station **A-5** (the nominal pi-open
owner per `.claude/rules/portfolio-conventions.md` writable-paths table).
**Escalation caller**: `chief-architect` (escalation receiver) when residue resolution
needs cross-lane governance judgment.

## Why this is a skill, not a `model: none` trigger
`./portfolio.sh pi-open <PI>` is two phases: (1) deterministic target resolution +
branch/tag creation; (2) a dirty-tree pre-flight that, on residue, emits an
**ATTRIBUTION MANIFEST** routing each uncommitted path to its owning agent. Phase 2
needs an agent to *act* on the manifest. A mechanical entry can only re-print the
halt. This skill is that agent's playbook.

## Collaboration mode — solo (default) vs copilot
Per `portfolio/system-refs/agent-definition-model.md` Part IV, every action is
`(Principal, Agent)`. Residue resolution depends on the operator's mode:

- **Solo mode (default today — one human principal behind every agent):** the invoking
  ceremony agent owns the ENTIRE residue resolution, working interactively with the
  operator. It MAY commit another lane's orphan work **on behalf of that lane**,
  carrying the **original owner's `Agent:` trailer** (capability attribution intact)
  with the **operator as git-author** (accountability). This is RBAC =
  principal-entitlement × agent-lane (Part IV #9): a solo operator-principal is
  entitled to all lanes. The `lane-enforcement` hook is **trailer-keyed**, so a
  `product-manager`-trailered commit on PRD paths passes unchanged — **no hook code
  carve-out**; the carve-out is *documented operator authorization*, recorded as a
  `pi-open-reconcile` AUDIT row per commit. This is NOT the PI14 reconcile
  anti-pattern: that was an UN-attributed bulk sweep with no principal in the loop;
  here the operator confirms each item and the trailer stays truthful.
- **Copilot / multi-principal mode:** when a DIFFERENT human owns a lane, that human
  must vouch for their own work → ROUTE that lane's orphan authored work via handoff
  (the strict fan-out). Junk/hygiene + the invoker's own lane are still handled
  directly. The operator declares copilot mode at invocation (no machine mode-flag
  yet — Part IV decision #10 is future work).

## Inputs
- `PI` — lane-suffixed id passed to pi-open (e.g. `16-plt-aip`).
- `mode` — `solo` (default) or `copilot:<lane>=<human>` declarations.
- Reads: live `git status` of every target repo; the pi-open ATTRIBUTION MANIFEST;
  `.claude/rules/portfolio-conventions.md` lane table; root + per-repo `.gitignore`.

## Steps
1. **Resolve + pre-flight.** Run `./portfolio.sh pi-open <PI>`. Clean trees →
   branches+tags created; go to 6.
2. **On `Pre-flight FAILED`, classify every residue path** via the table below (the
   manifest's `→ agent` column is the same lane mapping; verify against the table).
3. **Present the triage plan to the operator** — grouped:
   (a) **junk → gitignore** (list the exact patterns; **flag anything that might be
   real** — e.g. `*.py`/`*.csv` generator scripts, a new product dir — for an explicit
   keep/ignore call; never assume junk);
   (b) **orphan authored work → commit-on-behalf** (list `path → Agent: <owner>`
   trailer that will be used);
   (c) **separate-repo dirs / unknown → escalate**.
   Get the operator's confirm/adjust. Do NOT proceed past any keep/ignore ambiguity.
4. **Execute.**
   - *Solo mode:* append patterns to the OWNING `.gitignore` (root = workspace,
     chief-architect; per-product = that product). For each confirmed orphan
     artifact, `git -C <repo> commit` with the **original owner's** `Agent:`/`WI:`
     trailers; commit body notes "reconciled at pi-open, operator-confirmed". Append
     one `pi-open-reconcile` AUDIT row per commit (path · original-owner trailer ·
     operator-confirmed).
   - *Copilot mode:* same for junk + own-lane; emit one `## Handoff` per other-human
     lane and HALT for those.
5. **Re-run `./portfolio.sh pi-open <PI>`**; loop 1–4 until every target repo is clean.
6. **Branch + tag created.** Append the `pi-open` AUDIT row (agent = invoking agent).
   Print the exit sentinel.

## Residue triage table (each path → exactly one action; first match wins)
| Class | Match | Action |
|---|---|---|
| Whitelisted ledger | `portfolio/METRICS/{tokens,feedback}.ndjson` | Ignore — pi-open already whitelists; rolls into the next commit. |
| Secret-risk file | `ssm_params.json`, `**/*params*.json`, `*.pem`, `*.key` | **gitignore** + run `secret-scanner`; if already tracked, flag governance-auditor. NEVER commit. |
| Harness lockfile | `**/.claude/*.lock` | **gitignore** — harness-local, machine-bound. |
| TF plan output | `**/plan-*.out`, `portfolio/infrastructure/**/*.out` | **gitignore** — transient. |
| Report/build artifact | `portfolio/REPORTS/**.{pdf,png,pptx,csv}`, `~$*`, `*.zip` | **gitignore** (operator-confirmed). Never `rm`. **Flag `*.py`/`*.csv` as possibly-real** → ask before ignoring. |
| Separate/unregistered repo dir | `corporate-site/`, `aim-starter/`, any dir owning its own `.git` | **Escalate** to `platform-engineer` (repos.yml steward): register as submodule or gitignore. Never commit into the parent. |
| Orphan authored work | PRD / seed / service code with a clear owning lane | **Solo:** commit-on-behalf with the ORIGINAL owner's `Agent:`/`WI:` trailer, operator-confirmed (git-author = operator). **Copilot:** if a different human owns the lane → ROUTE via handoff; else commit-on-behalf. |
| Unknown | none of the above | **Escalate** to operator — list the path; no auto-action. |

## --force policy (break-glass only)
`--force` orphans residue onto the new PI branch (the "PI15-on-main" defect class).
This skill **never** passes `--force` itself. Only the operator may, with
`--force-rationale "<reason>"`; the skill records a `pi-open-forced` break-glass AUDIT
row and sets `forced=true` in the sentinel.

## What this skill is NOT
- Not a tree-cleaner that `rm`s files — junk is gitignored, never deleted.
- Not an UN-attributed bulk sweep — every reconcile commit carries the original
  owner's `Agent:` trailer + an operator-confirmed `pi-open-reconcile` AUDIT row.
  That attribution + operator-in-the-loop is what separates it from the PI14
  reconcile anti-pattern.
- Not permission to skip same-turn commits at prior stations — residue is a discipline
  lapse (PI16-plt-aip intake §1 evidence-based station-advance gate); this skill is the
  safety net, not a license to create residue.
- Not a replacement for `lane-merge-resolver` (that is PC-2 merge-conflict resolution).
- Not authorized to invent target scope — targets come from the intake `[targets:]`
  line via `_resolve_pi_targets`; a missing/wrong scope is a chief-architect A-1 fix.

## Exit signal — REQUIRED (per `.claude/skills/SKILL-EXECUTION-CONTRACT.md`)
Final action MUST print, ASCII-only, verbatim:

`PI-OPEN-ORCHESTRATOR: COMPLETE -- pi=<PI> branched=<n>/<total> reconciled=<r> routed=<x> forced=<true|false> validation=PASS|FAIL`

`validation=PASS` iff every target repo's `pi<PI>` branch + `pi<PI>-start` tag exist
AND no target repo is dirty (excluding whitelisted ledgers). The next agent
HARD-STOPS on a missing sentinel per the contract.
