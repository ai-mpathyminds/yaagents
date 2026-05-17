---
name: duplication-scanner
description: Find candidate shared code ACROSS product repos for platform-librarian — cross-product aggregation + extraction recommendation. Reports only; never extracts.
---

# duplication-scanner

Thin cross-product aggregator. The per-file "is this a reuse opportunity?" question is already answered by the user-level `simplify` skill — this skill does NOT re-implement that logic. It walks multiple product repos, collects candidates, and lets `simplify` do the quality/reuse evaluation per file.

> **Not** `wi-context-clusterer`. That skill clusters a single PI's `[READY]` **WIs**
> by shared context at **A-3 design-time** (pre-runbook; feeds execution-runbook-generator
> Pass-2). This skill scans **code files across product repos post-hoc** for the
> platform-librarian. Different artifact, different lifecycle — do not conflate.

## Relationship to `simplify`

| Concern | Owner |
|---------|-------|
| "Is this code a reuse opportunity? What's the cleaner form?" | `simplify` (invoke per candidate file) |
| "Does the same pattern exist in ≥ 2 product repos?" | this skill (grep + aggregate) |
| "Which extraction candidate is ready now vs later?" | this skill (score + rank) |
| "Extract and refactor the code" | `{lang}-developer` WI (separate step) |

## Inputs
- Domain hint (e.g. "HTTP middleware", "JWT parsing", "retry+backoff", "error envelope helpers")
- Product set (default: all four portfolio products) — narrow if scoped
- Language filter (`go`, `python`, `typescript`)

## Steps

1. Derive grep patterns from the domain hint (function signatures, type names, distinctive imports).
2. For each product in scope, `Grep` with a tight `glob` (e.g. `**/*.go`, excluding `**/vendor/**`, `**/node_modules/**`).
3. Collect candidate files. For each unique `(product, file)`:
   - Invoke the user-level **`simplify`** skill with the file in focus and the domain hint as the reuse question. Let `simplify` judge: "is this a reuse opportunity? what's the cleaner form?"
   - Capture `simplify`'s verdict: OPPORTUNITY / NOT-AN-OPPORTUNITY, with its suggested canonical form if any.
4. Cross-product aggregation (this skill's unique value):
   - Group candidates by normalized intent (same function purpose across products).
   - For each group with occurrences in ≥ 2 products where `simplify` flagged an opportunity:
     - Probe change frequency: `git log --oneline --since="6 months ago" -- <file>` per occurrence. High churn (>10 commits / 6 months) = risky to extract now.
     - Score: count of products × `simplify` confidence × inverse-churn.
5. Produce the extraction-candidate table below.

## Output

```markdown
## Duplication scan — {domain hint} — YYYY-MM-DD

| Pattern | Products | `simplify` opportunity? | Churn | Recommendation |
|---|---|---|---|---|

### Recommended "extract now"
- {pattern}: {justification; proposed shared module path; first adopter; migration order;
  cite the `simplify` per-file verdicts that grounded this}

### Recommended "extract later"
- {pattern}: {why; what would change the recommendation}

### Leave product-local (even though ≥2 products have it)
- {pattern}: {why; either `simplify` says not-an-opportunity, churn too high, or divergence already real}
```

## Extraction threshold (platform-librarian applies)

Recommend "extract now" only when ALL are true:
- ≥ 2 products have it
- `simplify` flagged OPPORTUNITY on ≥ 2 of them with converging canonical forms
- Low churn (< 10 commits in last 6 months per occurrence)
- No blocking open CVE or deprecation in the enclosing module

Otherwise: "extract later" (flag a revisit trigger) or "leave product-local."

## What this skill does NOT do

- **Evaluate per-file reuse quality** — `simplify` owns that. Calling `simplify` is the required step, not a shortcut.
- Extract or refactor code — that's a developer WI authored by a `{product}-architect` after reviewing this scan.
- Propose the shared-module API surface — architect territory.
- Run across product boundaries that would violate the "no cross-product reads" token-budget rule unless explicitly in scope for the scan (the scan IS the authorized cross-product read for platform-librarian's lane).
