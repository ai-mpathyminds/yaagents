---
name: component-library-designer
description: Author or update a portfolio-level component spec under portfolio/design-system/components/ for ux-architect — enforces spec schema + product-local → portfolio-shared promotion criteria.
---

# component-library-designer

Produces one component spec file per invocation at `portfolio/design-system/components/{component-name}.md`. Does NOT write component implementation code — that's frontend-developer.

## Inputs
- Component name + slug
- Consumers: which products/features need it
- Base primitives already in `portfolio/design-system/` (check via Glob first)
- Existing one-off versions in `{product}/ui-*/src/components/` (grep for component-name across products)

## Steps

1. Check for existing spec at `portfolio/design-system/components/{name}.md`. If present: append a changelog entry, don't rewrite.
2. Check existing product-local implementations (grep `oppor/ui-oppor/src/components/**/*.tsx`, `ai-platform/ui-ai-platform/**/*.tsx`). Note signatures.
3. Derive canonical prop surface as the union of reasonable variations — name the product source for each prop so frontend-developer can reconcile during migration.
4. Render the spec using the template below.
5. If multiple products have divergent versions, add a "Migration Plan" section listing which product owns the current source of truth (based on maturity + test coverage) and what breaking changes the others will need.

## Template

```markdown
# {Component Name}

**Type**: ds: (portfolio design-system primitive) | ds: (composite)
**Status**: [DRAFT] | [READY]
**Consumers**: {list of products/features}
**Version**: 1.0 (bumped on breaking prop change; minor on additive)

## Purpose
{one sentence on what the component does, user-facing}

## Props
| Name | Type | Required | Default | Description |
|---|---|---|---|---|

## Variants
{list with visual rationale — when to use which}

## States
- default, hover, focus, active, disabled, loading, error

## Composition
- Children: {what can be nested}
- Slots: {named slots if any}

## Accessibility
- Role / aria-attributes
- Keyboard: {tab order, shortcuts}
- Focus management
- Screen-reader copy

## Stories (spec — implementation is a WI)
- story: default
- story: each variant
- story: each state combination of interest
- play function: {interaction scripted test}

## Design tokens used
{list token keys from portfolio/design-system/tokens/}

## Migration Plan (if multiple product-local versions exist today)
- Source of truth: {product}/src/components/{name}
- {other product}: {list breaking changes to align}
- Timeline: {PI target}

## Changelog
- YYYY-MM-DD v1.0 {change}
```

## Promotion criteria (reject if not met)

Before writing a ds: spec for a component, verify at least one is true:
- 2+ products need the same component with compatible props, OR
- it's a primitive (Button, Input, Badge, Dialog, form field) obviously generalizable, OR
- a11y/token regression would result from keeping it product-local.

If none apply, return "should stay product-local" with rationale — do not write the spec.

## What this skill does NOT do

- Write React code / CSS — out of lane (frontend-developer).
- Perform the actual migration of product-local code to ds: — that's a frontend-developer WI.
- Invent new tokens — that's `design-token-curator`.
