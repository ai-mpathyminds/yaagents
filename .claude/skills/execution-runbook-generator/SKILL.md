---
name: execution-runbook-generator
description: Author portfolio/RUNBOOKS/pi{n}-execution.yml at Phase A station A-6 — dispatch-ready Phase B runbook rolled up from the [READY] PI roadmap + per-service WI files. Body has full schema.
---

# execution-runbook-generator

Invoked by `chief-architect` once Phase A A-4 (NFR pass) has flipped every PI WI to `[READY]` AND A-5 mechanical pi-open ceremony has closed (rule §10 hard-gate; PI10+ numbering — was A-6 pi-open dispatched-out-of-order in PI4..PI9). Produces the execution runbook the user dispatches one entry at a time per PLAYBOOK §Phase B.

Asymmetry note: this skill exists because writing the execution runbook is **mechanical roll-up + 7-rule schema validation** — the kind of work where freehand drift is invisible until S3. The companion **planning runbook is authored freehand** by chief-architect (no skill) because it is judgment-heavy (ADR slate, A-2.5 in/out, capacity sizing). If planning errors recur, promote to a skill — earn the abstraction.

## Inputs (deterministic — all required)

| # | Input | Path | Notes |
|---|-------|------|-------|
| 1 | **PI roadmap** | `{product}/docs/PI{n}/roadmap.md` | Authoritative WI list + dep order |
| 2 | **Per-service WI files** | `{product}/docs/PI{n}/{service}.md` | Source of acceptance criteria for `brief:` |
| 3 | **ADRs (referenced)** | `{product}/docs/adr/PI{n}-*.md` | Cite in `brief:` where a WI implements an ADR decision |
| 4 | **Phase A planning runbook** | `portfolio/RUNBOOKS/pi{n}-planning.yml` | Cross-check `target_services` and `pi{n-1}_carry_forward` are reflected |
| 5 | **Schema template** | `portfolio/RUNBOOKS/TEMPLATE.yml` | `kind: execution` |
| 6 | **Model-tier matrix** | `.claude/rules/token-budget.md` | Per-agent model assignment |

If any input is missing — or any roadmap WI is not `[READY]` — **do not author a partial runbook**. Emit Handoff with `> blocker: <missing input or list of [DRAFT] WIs>` and return.

## Pass-1 — WI roll-up (steps the skill performs)

> **Two-pass** (PC-5-01; redesigned PC-5-18, PI15-plt-aip-postmortem). Pass-1 below produces
> one entry per `[READY]` WI exactly as before — **mechanical, schema-validated, unchanged
> behaviour, zero quality change**. Pass-2 (§ below) then consolidates same-context entries to
> cut the ~4–6K-token-per-session fixed tax of a high entry count.
> **Pass-2 is a SELF-DERIVING chief-architect judgment step that ALWAYS runs** (after step 8,
> before step 9 render): the skill derives the clusters itself from the Pass-1 entry set. The
> external `wi-context-clusterer` artifact is an **optional accelerator hint, not a
> precondition** — **there is no silent flat-N fallback** (the PI15 failure: Pass-2 was lost
> entirely → 49 flat entries because the clusterer wasn't run at A-3, AUDIT-1308). If the hint
> is absent, Pass-2 self-derives and the runbook carries an explicit
> `> note: Pass-2 self-derived; clusterer input absent` preamble. Pass-2 may legitimately
> derive **zero** merges — but that is an explicit logged judgment, never a skipped step.

1. **Load roadmap.** Parse the WI table; record dep order. WIs not at `[READY]` are blockers — abort with the list.
2. **Walk each WI.** For each `[READY]` WI, read its body in `{service}.md` to extract: title, acceptance criteria summary (1 sentence), file paths touched, ADR refs.
3. **Map WI → agent**. WI body may carry an explicit `**Agent**:` override line — when present it WINS over the service-type default below (rule §8 enforces consistency: rule §8 fails the runbook if the explicit override is ignored). Default mapping by service-type → developer-language convention:
   - Go service (agent-api, ai-gateway, knowledge-api, tooling-api, iam, audit, notifications, object-lifecycle, config, gateway, campaigns, assets, channels) → `go-developer`
   - Python service (evaluation-service, knowledge-worker, agent-runtime) → `python-developer`
   - React frontend (ui-ai-platform, ui-oppor) → `frontend-developer`
   - **NFR/SRE/SEC/FIN WI**: use the **language-developer of the service the WI touches** (PI6 precedent: `WI-6.NFR-FE-1` Next.js bump → `frontend-developer`, commit `57faee9`). `platform-engineer` only owns NFR WIs whose file scope is exclusively `infra/local/**`, `docker-compose.yml`, or `ai-platform/docs/**` (pure-ops, no service code).
   - Per-service helper bootstrap WI (reviewer / test-writer / openapi-sync / component-auditor / a11y-reviewer / storybook-writer) → the matching helper agent file
   - **Cross-lane WI** (e.g. UI roadmap row touching a `chief-architect`-owned port table; a `platform-engineer`-owned compose change embedded in a service-language section) — author MUST add `- **Agent**: <name>` to the WI body. Rule §8 then enforces.
4. **Compose `brief:`** ≤3 sentences citing the WI id, target service, file paths, and acceptance criteria essence. Cite ADRs by id where applicable.
5. **Wire mandatory skills.** Every entry's `skills_required` MUST include `runbook-feedback` per PROCESS.md 2026-04-24 [ADOPTED]. Additional skills per WI type (helper bootstraps include their reviewer; UI WIs include `component-library-designer` if curating a shared component; service WIs touching OpenAPI include `openapi-sync`).
6. **Resolve `depends_on`.** Translate roadmap dep refs into runbook entry ids (`B-NN` format).
7. **Set `escalate_to: chief-architect`** on every entry per PROCESS.md.
8. **Group into sprints; compute optional `sprint_preamble:` and per-entry `session_strategy:`.** Sprint grouping (consecutive same-agent same-cwd entries with shared cited references — see §sprint_preamble fill heuristic) drives both:
   - **(a) `sprint_preamble:`** (PROCESS.md 2026-04-26 [ADOPTED]; codified PC-5-02 of `pi7-postmortem.yml`). If no sprint has ≥2 same-agent entries, skip — emit the flat `entries:` form.
   - **(b) `session_strategy:`** (per `.claude/rules/agent-context-discipline.md` Layer 1, [ADOPTED] 2026-05-04). Assign per entry: `fresh` for the first entry of each new sprint OR any cross-product agent invocation (different `product:` than the previous entry); `compact-after` every K-th entry within a same-agent same-sprint cluster (K=4 default — i.e. on entries 4, 8, 12 of a long cluster); `continue` otherwise (default — `dispatch-entry.sh` uses `--resume <session-id>`, cache stays warm). Field is OPTIONAL on the entry — omit equals `continue`. Long clusters (>6 same-agent same-sprint entries) SHOULD prefer at least one `compact-after` insertion; warn (not halt) if absent.
8.5 **Pass-2 — self-derived common-context consolidation** (§ "Pass-2" below). ALWAYS runs
   here, on the Pass-1 entry set, BEFORE render — chief-architect derives the clusters inline
   (consuming the optional `wi-context-clusterer` hint if present). Never skipped; a zero-merge
   outcome is an explicit logged judgment, not a no-op skip.
9. **Render** the (Pass-1, or Pass-2-consolidated) entry set into `portfolio/RUNBOOKS/pi{n}-execution.yml` using `TEMPLATE.yml`.
10. **Append** one line to `portfolio/AUDIT.md` with verb `runbook-written`.

## Pass-2 — self-derived common-context consolidation (PC-5-01 guardrail; PC-5-18 redesign)

**Purpose**: reduce the *number of dispatch entries* (each session resume/fresh-start costs
~4–6K tokens regardless of work — PI14-plt-aip-user-input.md §2) by merging WIs that share
context. **It reduces entry count, never rigor.**

**Self-derivation (the default path — PC-5-18).** Pass-2 derives its own clusters from the
Pass-1 entry set; it depends on no external artifact. The derivation reuses two signals the
skill already computes:
1. **Sprint signal** — the step-8 sprint grouping (consecutive entries sharing `agent` + `cwd`
   + ≥1 cited reference; see §sprint_preamble fill heuristic). Entries in one sprint are merge
   candidates.
2. **Brief-similarity signal** — the rule-§12 first-paragraph token-Levenshtein primitive (now
   in the `advisory-runbook-rules` companion skill): entries with similarity ≥0.7 after
   service-name + stop-token stripping are merge candidates even across non-adjacency.

For each candidate cluster the skill reads the member WIs' acceptance criteria (already loaded
in Pass-1 step 2) and computes `union_acceptance_criteria` itself. Clusters whose members
would bloat one working-set beyond coherence, or whose members disagree on `agent`/`cwd`, are
derived as **keep-separate** (one entry each). A derived **zero-merge** outcome is valid and
logged — a judgment, not a skip.

**Optional accelerator hint.** If `{product}/docs/PI{n}/wi-context-clusters.md` (the
`wi-context-clusterer` artifact) is present, consume its `clusters[]` as a STARTING hint to
save re-derivation, gated on its sentinel `WI-CONTEXT-CLUSTERER: COMPLETE — … validation=PASS|FAIL`:
- `validation=PASS` → seed the candidate set from the hint, then **still** validate every merge
  against P2-1..P2-4 + the architect's own judgment (the hint accelerates; it never authorizes a
  merge the guardrails would reject).
- `validation=FAIL` → **ignore the hint, self-derive**; emit `> warn: clusterer input FAIL — Pass-2 self-derived instead`.
- absent → **self-derive**; emit `> note: Pass-2 self-derived; clusterer input absent`.

**Under no condition does absence/failure of the hint produce a silent flat-N runbook** — that
regression (AUDIT-1308: PI15 lost Pass-2 entirely, 49 flat entries, token-tax unmitigated,
because the clusterer wasn't run at A-3) is exactly what this redesign removes.

**Procedure** — for each derived (or hint-seeded) merge-candidate cluster that the architect did
**not** annotate `cluster_override:`:
1. Replace its member Pass-1 entries with **one** consolidated entry.
2. `consolidates:` = the member WI ids (new field). The roadmap still tracks each
   member WI individually; the consolidated entry's close transitions all of them.
3. `brief:` cites every member WI id + the cluster's `shared_context`.
4. **GUARDRAIL (user-input §2 hard constraint — mechanically asserted, not aspirational)**:
   the consolidated entry's acceptance/intent MUST carry the cluster's
   `union_acceptance_criteria` **verbatim and complete**; every member's test gate and
   review step survives. Validation P2-1 asserts the merged AC set ⊇ the union (count
   check). Any shortfall = **HARD FAIL, abort authorship** (never silently drop).
5. `depends_on:` = union of members' deps minus intra-cluster edges (keep acyclic).
6. `agent:`/`cwd:` must be identical across members (clusterer guarantees same product;
   if members disagree on agent → do NOT merge that cluster, emit `> warn:` and keep them
   separate — mis-routed dispatch is worse than entry count).
`unclustered` WIs and `recommend: keep-separate` clusters → untouched (one entry each).

**Pass-2 validation (assert before render — additive to the §1–§10 hard rules)**:
- **P2-1** every consolidated entry's AC set ⊇ its cluster `union_acceptance_criteria` (no dropped AC/test/review). HARD FAIL.
- **P2-2** no WI lost: every Pass-1 WI appears either as its own entry OR in exactly one `consolidates:` list. HARD FAIL (this is the rule §1 generalization).
- **P2-3** consolidated members shared `agent:`+`cwd:` (else not merged).
- **P2-4** architect `cluster_override:` on a cluster ⇒ Pass-2 leaves its members un-merged (documented choice; §12 `duplication_override` precedent).

P2-1..P2-4 are what make the self-derived judgment **safe to keep inside the skill** rather than freehand: a self-derived merge that drops an AC, a test gate, or a WI fails P2-1/P2-2 at authorship — the same *drift-invisible-until-S3* guard that justifies this skill's existence (see asymmetry note) now bounds Pass-2's judgment. The judgment is the architect's; the mechanical floor is non-negotiable.

## `sprint_preamble:` fill heuristic (step 8)

The runbook may group entries into sprints with an optional `sprint_preamble:` block listing files all entries in the sprint should preload. `bin/dispatch-entry.sh` prepends this block to the first brief in the batch so entries 2..N read from the warm prompt cache. See `portfolio/docs/PI7-dispatch-pattern-analysis.md` for the read-graph rationale and savings table (~60–65% per follow-up entry in a same-agent batch).

**Sprint grouping**: a sprint = consecutive entries sharing `agent` + `cwd` AND sharing ≥1 cited reference (ADR id, freeze section, or service-local convention file). PI7 worked example: B-09 + B-10 (both `go-developer` in `ai-platform/services/agent-api`, both citing WI-7.A0 freeze + ADR PI7-0004) form one sprint.

**Preamble content** = intersection of references across the sprint's entries. Compute by:

1. For each entry in the candidate sprint, parse `brief:` for path-shaped tokens (`docs/PI{n}/`, `docs/adr/`, `internal/`, `services/`).
2. Intersect: keep paths cited in ≥2 entries OR cited in ≥1 entry AND the architect's roadmap header.
3. Add the per-service convention files (CLAUDE.md, code-reviewer.md, contract-freeze.md if it exists for the PI).
4. Cap at ~10 paths — preamble is meant to seed cache, not enumerate the world.

Skip the wrapper for cohorts where intersection is empty or where every entry stands alone (cross-agent boundaries, single-entry services). Backward-compat: flat `entries:` form remains valid; the `sprints:` wrapper is opt-in.

**Sprint-grouping check** (conditional, applies only when sprint groups are emitted; not numbered alongside the 8 mandatory rules below): every entry in a sprint MUST share `agent` AND `cwd`. Mixing agents inside one sprint defeats the cache assumption.

## Validation rules (asserted before write — all 8 MUST pass)

| # | Rule | Failure mode if missed |
|---|------|------------------------|
| 1 | Every roadmap WI appears in exactly one entry — **either** its own entry **or** as a named member of exactly one consolidated entry's `consolidates:` list (Pass-2). Enforced by Pass-2 rule P2-2. | Silent WI drop → developer never picks it up |
| 2 | Every entry has `runbook-feedback` in `skills_required` | Phase C postmortem starves; PROCESS violation |
| 3 | Every entry has `escalate_to: chief-architect` | Phase B blockers route to /dev/null |
| 4 | `depends_on` chain is acyclic and references existing ids | Dispatch loop deadlock |
| 5 | No entry references a WI in `[DRAFT]` (unreviewed). **`[BLOCKED]` is allowed** if the WI body documents the unblock pathway and the runbook entry's `status: blocked` mirrors the WI token (per `TEMPLATE.yml` allowed values `pending\|in-progress\|done\|blocked`). The architect may intentionally scope a WI to open `[BLOCKED]` (e.g. PI7 WI-7.R3 awaiting librarian verdict — `docs/PI7/roadmap.md §Capacity`). | Developer dispatched against unreviewed WI; OR partial runbook authored when architect's intentional `[BLOCKED]` should have been preserved |
| 6 | `cwd:` matches the service repo path (not portfolio root) for service WIs | Commit lands in wrong repo |
| 7 | `agent:` model tier matches `.claude/rules/token-budget.md` matrix | Cost overrun (haiku WI dispatched to opus, etc.) |
| 8 | When the WI body in `{product}/docs/PI{n}/{service}.md` carries an explicit `- **Agent**: <name>` field, the entry's `agent:` MUST resolve to that name (basename without `.md`). **Mismatch is a hard fail.** Missing `**Agent**:` (legacy WIs not yet annotated) is a soft warn, NOT a fail — graceful migration path. | Mis-routed dispatch — `frontend-developer` invoked against a WI whose body says `**Agent**: chief-architect`. PI8 evidence: B-24 / WI-8.U26 (port-table) was authored as `frontend-developer` but the WI is chief-architect lane; symptom surfaced only at dispatch via `help_needed`, after token spend. |
| 9 | Every entry's `agent:` field uses the **canonical `.claude/agents/<name>.md` form**. Bare names (e.g. `agent: ai-platform-architect`) HARD FAIL. Mechanical entries (`model: none`) follow the same canonical form for their nominal agent. | Bare names skip `bin/dispatch-entry.sh` upward-search resolver (lines 561–604; 2026-04-27 fix); the `*)` fallback only does direct path concat. PI10 evidence: pi10-planning.yml authored 2026-05-01 with bare names broke A-3 dispatch until corrected. |
| 10 | Before authoring `pi{n}-execution.yml`, verify `portfolio/AUDIT.md` contains a `pi-open` row referencing PI{n}. If absent, abort with `> blocker: PI{n} branches not created. Dispatch scrum-master to planning-runbook pi-open mechanical entry (typically A-5 in PI10+ numbering; A-6 in PI4..PI9 numbering); re-invoke this skill after the pi-open AUDIT row lands.` | pi9-no-branches drift: PI9 shipped without ever creating a pi9 branch because pi-open was never dispatched. Phase B commits flowed straight to main. Catching at execution-runbook-authoring time prevents the entire PI from running on the wrong branch. |

Any failure halts authorship and surfaces in the Handoff `> blocker:` block. Partial runbook is never written.

## Advisory rules §11–§13 (warn-only) — see companion skill

Advisory rules **§11** (per-WI library annotation), **§12** (same-PI duplication detection), and **§13** (package-manager validation) are specified **verbatim** in the companion skill **`advisory-runbook-rules`** (`.claude/skills/advisory-runbook-rules/SKILL.md`) — extracted at WI-15.GOV-7 to keep this skill's hard-rule core lean; **numbering and semantics are unchanged**.

These rules SURFACE findings (per `.claude/rules/library-gates.md` Gates 3/4/5, PI10 PC-4 carry-forward) into the runbook + AUDIT + postmortem inputs; they do **NOT** block runbook authorship — contrast §1–§10 hard rules, which halt it. Step 8.5 / step 9 (render) STILL evaluate §11–§13 exactly as before, emit `context.advisory_warnings:` + `context.duplication_findings:`, prepend the `## Advisory warnings` comment block, and append the §12 `duplication-found` AUDIT row. Apply them per the companion skill's rule table + the rule-§12 Levenshtein algorithm; **the rule-of-three trade-off remains a valid architect choice** (PI10-user-input.md §Counter-positions explicit) — these rules surface the choice, they do not pre-empt it. Each warn becomes a postmortem candidate at PC-4.

### Rule §8 — algorithm

For each runbook entry being authored:

1. Open `{product}/docs/PI{n}/{service}.md` (the per-service roadmap doc the entry's `wi:` belongs to).
2. Locate the WI's heading line `^### <wi>(:|\s|—|$)` (matches `### WI-8.U26: ...`, `### WI-8.U26 — ...`, etc.).
3. Walk forward from the heading to the next `^### ` line or EOF. That window is the WI body.
4. Inside the window, scan for the FIRST line matching `^[-*]?\s*\*\*Agent\*\*:\s*(\S+)`. Capture the first whitespace-delimited token after the colon — strip backticks and any trailing parenthetical (the rationale that often follows in the same line, e.g. `chief-architect (rule X is chief-architect lane)`).
5. **If captured value is non-empty** AND captured value does not equal the entry's `agent:` basename (file name without `.md`, leading `.claude/agents/` stripped) ⇒ HARD FAIL. Surface in the Handoff `> blocker:` block as: `rule-§8 mismatch: entry <id> WI <wi-id>: runbook agent='<entry-agent>' but WI body **Agent**:='<wi-agent>'`.
6. **If no `**Agent**:` line found** ⇒ soft warn (logged, but the runbook is still authorable). The default service-type → language-developer mapping from step 3 stands.

Implementation note: rule §8 fires DURING runbook authorship (step 9 `Render`), not at dispatch. The whole point is to catch mis-routings before tokens are spent on a session that ends in `help_needed`. PI8 B-24 / WI-8.U26 — frontend-developer dispatched against a chief-architect WI — is the canonical regression test: with rule §8 in place, that runbook would have failed authorship with a clear error citing the expected agent.

## Entry schema (per `TEMPLATE.yml` `kind: execution`)

| Field | Source |
|-------|--------|
| `id` | generator (`B-01`, `B-02`, … in dep order) |
| `wi` | from roadmap (e.g. `WI-7.A1`) |
| `agent` | step 3 lane mapping |
| `model` | from `.claude/rules/token-budget.md` matrix per agent type |
| `product` | from roadmap header |
| `cwd` | service repo path (e.g. `ai-platform/services/agent-api`) |
| `artifact` | file paths from WI body |
| `intent` | one sentence from WI body |
| `brief` | step 4, ≤3 sentences |
| `skills_allowed` | optional |
| `skills_required` | step 5 (always includes `runbook-feedback`) |
| `depends_on` | step 6 |
| `escalate_to` | `chief-architect` (always) |
| `estimated_effort` | `S | M | L` from WI body |
| `status` | `pending` at authorship |
| `notes` | empty at authorship |
| `library_ref` *(advisory rule §11)* | optional — cite a `LIBRARIES.md` row (e.g. `LIBRARIES.md §Internal:azblob` or `LIBRARIES.md §External:joserfc`) |
| `library_justify` *(advisory rule §11)* | optional — one-line "no catalog entry applies because…" (mutually exclusive with `library_ref`) |
| `duplication_override` *(advisory rule §12)* | optional — one line, present when architect accepts a §12 finding's duplicate WI as-is (rule-of-three or other rationale) |
| `session_strategy` *(per `.claude/rules/agent-context-discipline.md`)* | optional — `continue` (default; omit field) \| `fresh` \| `compact-after`. Step 8b heuristic. |
| `consolidates` *(Pass-2, PC-5-01/PC-5-18)* | optional — list of member WI ids merged into this entry by the self-derived Pass-2. Present ⇒ entry carries the cluster's full `union_acceptance_criteria` (P2-1). Absent ⇒ ordinary single-WI entry. |

## What this skill does NOT do

- Does not dispatch the runbook — user dispatches each entry per PLAYBOOK §Phase B.
- Does not modify the roadmap or WI files. If a WI body is too thin to author a `brief:`, halt and surface as `> blocker:` for the architect to expand.
- Does not author NFR / SRE / SEC / FIN WIs — those came from platform-engineer at A-4.
- Does not wire helper agents as standalone entries unless the roadmap explicitly carries a helper-bootstrap WI (e.g. `WI-N.T0` reviewer bootstrap).
- Does not author the planning or postmortem runbooks — separate concerns; planning is freehand, postmortem has its own skill.

## Exit signal

The skill's output is **complete** when:
- Every roadmap WI is accounted for in dep order — its own entry or exactly one `consolidates:` list (rule §1 / Pass-2 P2-2).
- All ten hard validation rules pass (rule §8 fires per-entry against the WI body's `**Agent**:` override; soft-warns on un-annotated WIs).
- Pass-2 ALWAYS ran (self-derived; PC-5-18) — P2-1..P2-4 all pass and the run prints `EXEC-RUNBOOK-GEN: COMPLETE — entries=<n> consolidated=<c> pass2=PASS(self-derived|clusterer-accelerated) validation=PASS`, where `consolidated=<c>` MAY be 0 (an explicit zero-merge judgment). `pass2=SKIPPED` is no longer a legal value; a flat-N runbook carrying no Pass-2 derivation note, OR sentinel absence, = hard stop (the AUDIT-1308 regression — never silent flat fallback — PC-5-06 skill-execution contract).
- Advisory rules §11–§13 have been evaluated; any findings appear under `context.advisory_warnings:` + `context.duplication_findings:` and in the runbook's top comment block (advisory findings do NOT block authorship).
- `exit_check:` block lists one assertion per WI ("`WI-N.X.Y` `[DONE]` with commit SHA on `main`").
- `AUDIT.md` `runbook-written` line appended with the runbook path; if advisory findings exist, AUDIT also gains one `duplication-found` row per §12 finding (rule §12 explicitly requires the audit row even though authorship proceeds).
