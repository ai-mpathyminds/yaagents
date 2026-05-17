---
name: wi-context-clusterer
description: Cluster a PI's [READY] WIs by shared context at A-3 — extends library-gates §12 similarity primitive; feeds execution-runbook-generator Pass-2. Report-only; never merges.
---

# wi-context-clusterer

Design-time (A-3) clustering of a single PI's `[READY]` WIs by **shared context**, so
`execution-runbook-generator` Pass-2 (PC-5-01) can consolidate same-context WIs into
fewer, larger entries — cutting the ~4–6K-token-per-session fixed tax that a high WI
count imposes even with `session_strategy` in place (PI14-plt-aip-user-input.md §2).

**This skill EXTENDS the `library-gates.md` §12 / execution-runbook-generator rule §12
similarity primitive — it does NOT reinvent it, and is NOT a second `duplication-scanner`.**

## Boundary — three adjacent things, kept distinct

| Concern | Owner | Axis |
|---|---|---|
| "Are these two WI briefs near-duplicate mirrors (same PI, differ only in `service:`)?" | library-gates §12 (pairwise, τ≥0.7) | the **primitive this skill reuses** |
| "Does the same code pattern exist in ≥2 **product repos**?" | `duplication-scanner` (cross-product, post-hoc, code files) | different artifact, different lifecycle |
| "Which `[READY]` WIs in **this one PI roadmap** share enough context to merge at runbook time?" | **this skill** (N-way cluster, design-time, WI bodies) | unique value |

## Caller / when

Invoked at **A-3** by `platform-librarian` (Gate-1/design-time consultation) or the
`{product}-architect` while the roadmap is `[READY]`. Output is consumed by
`chief-architect` at **A-6** in `execution-runbook-generator` Pass-2. Report-only.

## Inputs (deterministic)

1. `{product}/docs/PI{n}/roadmap.md` + per-service WI files — the `[READY]` WI set.
2. Each WI body's: `brief` first paragraph, `service:`, `parent_feature:`, cited ADR
   ids, cited file/dir paths, `library_ref:`, and **all acceptance-criteria lines**.

## Algorithm (extends §12 — cite, don't re-derive)

1. **Brief-similarity component** — apply the **verbatim §12 primitive** (tokenize brief
   first paragraph: lowercase, split `[\s,.:;()/]+`, drop stop-tokens, strip the WI's
   `service:` value; token-level Levenshtein; `sim = 1 − dist/max(len)`). Do not modify
   the primitive — import its result.
2. **Composite context score** for each WI pair (cap 1.0):
   `0.40·brief_sim + 0.25·same_parent_feature + 0.15·shared_ADR + 0.10·shared_path_prefix
   + 0.10·same_service`. (Weights are the skill default; an architect may override with a
   one-line rationale — mirrors the §12 `duplication_override` precedent.)
3. **Cluster**: build a graph with an edge where composite ≥ **τ = 0.60** (deliberately
   *below* §12's 0.7 pairwise-dup bar — we cluster shared-context, not identical, WIs and
   the merge is architect-reviewed, never automatic). Connected components → clusters of
   size ≥2. Singletons go to `unclustered` (Pass-2 leaves them untouched).
4. For each cluster emit the **union of all member acceptance criteria, deduped** — this
   is the GUARDRAIL artifact: Pass-2 MUST carry every one into the merged entry.

## Output contract (STABLE — PC-5-01 Pass-2 consumes this verbatim)

Write to stdout AND `{product}/docs/PI{n}/wi-context-clusters.md`:

```yaml
## WI context clusters — PI{n} — YYYY-MM-DD
tau: 0.60
clusters:
  - id: CL-1
    members: [WI-n.X, WI-n.Y, ...]            # ≥2
    shared_context: ["parent_feature:FX.PE", "adr:PIn-00k", "path:portfolio/infrastructure/roots/x"]
    composite_scores: {"WI-n.X~WI-n.Y": 0.74}
    union_acceptance_criteria:                 # EVERY member AC, deduped
      - "<verbatim AC line>"
    est_session_reduction: "<m members → 1 entry; ~<(m-1)×4–6>K tokens saved"
    recommend: merge-candidate                 # | keep-separate
unclustered: [WI-n.Z, ...]                     # singletons — Pass-2 must not touch
guardrail: "Pass-2 MUST preserve every union_acceptance_criteria item, every test gate,
  and every review step of every member. Dropping any is a HARD FAIL (user-input §2)."
```

## Validation (assert before write)

1. Every `[READY]` WI appears exactly once (in a cluster OR `unclustered`) — no WI dropped.
2. `union_acceptance_criteria` of a cluster ⊇ the multiset of every member's ACs (count check).
3. No cluster crosses `product:` boundaries (single-PI, single-roadmap scope).
4. `tau` and weights printed in output (reproducibility).

## What this skill does NOT do

- **Does not merge or rewrite WIs** — report-only. Merge is `execution-runbook-generator`
  Pass-2 (PC-5-01), architect-reviewed; an architect may reject a cluster with a one-line
  `cluster_override:` rationale (documented choice, not a blocked one — §12 precedent).
- **Does not reimplement §12 or duplication-scanner** — reuses §12's primitive; defers
  cross-product/code-file scope to `duplication-scanner`.
- Does not drop or weaken acceptance criteria, test gates, or review depth — the whole
  point of `union_acceptance_criteria` is to make Pass-2's GUARDRAIL mechanically checkable.
- Does not run cross-product (single PI roadmap only).

## Exit signal — REQUIRED

The run is complete only when the agent prints, verbatim, the sentinel line:

`WI-CONTEXT-CLUSTERER: COMPLETE — clusters=<n> unclustered=<m> validation=PASS|FAIL`

Absence of this line means the skill did NOT run — it is a HARD STOP, never a silent
fallback to hand-clustering (PI14-plt-aip-postmortem PC-5-06; the skill-execution
contract this thread's meta-lesson demands). `validation=FAIL` blocks Pass-2.
