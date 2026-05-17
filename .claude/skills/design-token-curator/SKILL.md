---
name: design-token-curator
description: Maintain the portfolio design-token set at portfolio/design-system/tokens/ for ux-architect — add, deprecate (with supersedes pointer), or audit tokens. Never deletes.
---

# design-token-curator

Keeps the portfolio's design tokens (color, spacing, typography, motion, radius, shadow) aligned across products. Single source at `portfolio/design-system/tokens/`.

## Token files (one per category)

- `portfolio/design-system/tokens/color.json`
- `portfolio/design-system/tokens/spacing.json`
- `portfolio/design-system/tokens/typography.json`
- `portfolio/design-system/tokens/motion.json`
- `portfolio/design-system/tokens/radius.json`
- `portfolio/design-system/tokens/shadow.json`

Format: W3C-style Design Tokens (flat JSON with `$value` + `$type` per key). Keeps the door open for `style-dictionary`-style generation.

## Operations

### Add a new token
1. Verify it's not a duplicate — grep the category file for similar `$value`.
2. Append with a conventional key (kebab-case, semantic-first: `color.surface.elevated-1`, not `color.neutral.12`).
3. Log in the category file's `## Changelog` section: `YYYY-MM-DD added {key} value={value} rationale={one line}`.

### Deprecate a token
1. Mark the token `$deprecated: true` and add `$replacedBy: {new-key}` OR `$deprecated: "manual migration; no replacement"`.
2. Never remove the entry — frontend-developer migrates consumers in a WI, then a later quarterly governance pass removes fully-unreferenced deprecated tokens.

### Audit
1. Grep `oppor/ui-oppor/src/**/*.{ts,tsx,css}`, `ai-platform/ui-ai-platform/src/**/*.{ts,tsx,css}` for inline color/spacing values (regex for hex colors, rem values).
2. Any inline value without a matching token → flag as "token bypass" in the audit output.
3. Return a table: Product | File:Line | Inline value | Nearest existing token (or "new token needed").

## Promotion threshold (product-local → portfolio token)

If the audit finds the same inline value in ≥ 2 products, it's a promotion candidate. Propose the key; require ux-architect decision to commit.

## Output formats

For add/deprecate: modify the JSON file + append to its Changelog section.

For audit: markdown table printed inline; do not write a file unless ux-architect explicitly asks.

## What this skill does NOT do

- Modify `.css` / `.tsx` files (frontend-developer lane).
- Delete tokens (deprecate only; quarterly governance handles final removal).
- Invent new token categories — if one is needed, raise via ux-architect decision.
- Rename tokens without a deprecation path.
