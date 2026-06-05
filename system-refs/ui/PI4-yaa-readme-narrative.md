# PI4-yaa Goal 3 DOC-01 — README Narrative Arc — UX Spec
Status: [READY]
Product: yaagents (meta-repo root `README.md`)
Surface: GitHub-reader (single-screen decision funnel)
Touches PRD: `yaagents/system-refs/yaagents-v0.4_detailed.md` §7 Goal 3 + §7.2 DOC-01 + §7.3 ecom outline
Design-system version: `portfolio/design-system/PRINCIPLES.md` v1.0 (2026-04-22) — mandatory Input #1
Spec authored: 2026-06-06 (ux-architect, A-2.5)

> README is GitHub-flavored Markdown rendered on github.com. There is no React, no Tailwind, no Starlight component library at this surface. PRINCIPLES.md still governs **narrative discipline** — restrained, information-dense, one primary action — even when typography/color/motion are out of our hands. The spec is an IA + copy-deck recommendation; renders as ordinary GitHub Markdown.

---

## 1. Surface constraints (read these first)

- **Rendering surface**: github.com README pane. No JS, no custom CSS, no Starlight. Mermaid IS rendered. Shields.io badges render.
- **Width**: GitHub renders body at ~1012px (sidebar wide) or ~768px (file view). Plan for the narrower.
- **First viewport**: roughly first 600px tall on a standard laptop after the repo header chrome. Treat this as **above the fold** for the decision funnel.
- **Audience**: a developer who Googled "agentic REST", clicked the repo, has ~45 seconds before deciding whether to scroll or close. Repeat-visit operator mindset (PRINCIPLES.md §1) does NOT apply here — README is a first-time-visitor surface.
- **Independence from Pages** (FF-3 user-direct, PRD §7.1): README is **NOT** lockstep with `start-here/overview.mdx`. Two surfaces, two tunings.

---

## 2. Narrative arc (the decision funnel)

The reader moves through 5 stages. Each stage is one screen-band, ordered top-down. The reader can stop at any band; the next band is a deeper commitment.

```
┌──────────────────────────────────────────────────────┐
│ Band 1: TARGET-AUDIENCE CALLOUT (one line, above fold)│  ← 45s decision: stay or close
├──────────────────────────────────────────────────────┤
│ Band 2: WHAT YOU CAN BUILD (the ecom example shown)   │  ← 90s decision: skim or read
├──────────────────────────────────────────────────────┤
│ Band 3: HOW IT WORKS (architecture mermaid + ecom flow)│ ← 3min decision: try or bookmark
├──────────────────────────────────────────────────────┤
│ Band 4: WHO USES IT (3-4 personas; "is this me?")     │ ← 5min decision: install
├──────────────────────────────────────────────────────┤
│ Band 5: NEXT STEPS (install + Pages link + spec)      │ ← exit to deeper surfaces
└──────────────────────────────────────────────────────┘
```

### Anti-patterns being removed (user-direct, PRD §7.1)

- ❌ A2A vs agentic comparison framing as **the** opener (it is opaque to the reviewer pool — kept only as a "see also" link in Band 4).
- ❌ "What is YAAgents?" as the H1 first content section (definition before motivation — reader hasn't earned the definition yet).
- ❌ Install commands above the fold (asks for commitment before reader knows what they would install).

---

## 3. Band-by-band IA

### Band 1 — Target-audience callout (above the fold)

**Goal**: in one line, tell the reader whether this is for them.

**Recommended copy shape** (final wording is product-manager's call; this is the IA, not the copy):

> *You're building an AI agent that needs to live behind a normal REST API — with auth, tenancy, audit, typed responses, and OpenAPI. YAAgents is the gateway + SDKs that let you keep your agent framework and still ship a governed product API.*

**IA constraints**:
- Single sentence, ≤30 words, ≤2 lines on the 768px-wide GitHub view.
- Identifies the reader (`You're building...`) — not the product (`YAAgents is...`).
- Names two concrete pain points (auth/tenancy/audit/typed) the reader already feels.
- Sits **after** the H1 + tagline + badge row, **before** any code blocks or diagrams. The whole block must fit above the GitHub fold on a 1366×768 laptop.
- No emoji. Tabular numerics N/A. (PRINCIPLES.md §1, §2.)

**Placement vs existing README**: replaces the current `> **Build the agent however you want. Expose it like a governed API.**` blockquote + the `POST /agents/invoke ← what YAAgents replaces` block. Those move to Band 2.

### Band 2 — What you can build (ecom example shown, not described)

**Goal**: the reader sees a real, runnable thing — not an abstraction.

**Content shape**:
1. **One-line use-case lead** — "Here's a product-recommendation API that asks for clarification when it has nothing to recommend." (Final copy: PM.)
2. **The endpoint table** (3 rows, no more — current README has 4; drop the demo `/agents/invoke` row to Band 4 anti-pattern callout):
   ```
   POST /recommendations/{customerId}             → typed clarification when no purchase history
   POST /campaigns/{id}/optimizations             → typed result
   POST /claims/{id}/reviews                      → typed approval-required
   ```
3. **One fenced HTTP snippet** showing a `clarification_required` response (verbatim from the PRD §9 normative table — `400 application/vnd.yaagents.clarification+json`). Five lines of body, not fifteen.
4. **One link line** at the close: "Run it: `examples/store/` (Python) · `examples/store-go/` (Go)." — these are the v0.4 PRD §11 reference flows.

**IA constraints**:
- No architecture diagram in this band (kept for Band 3).
- No conceptual framing ("agentic operations are first-class resources" etc.) — Band 2 is **show, don't tell**. The framing is implicit in the endpoint shape.
- ≤80 lines of rendered README space.

### Band 3 — How it works (architecture + ecom flow walkthrough)

**Goal**: the reader who decided "this is interesting" gets the mental model.

**Content shape** (in this order):
1. **Architecture diagram** — replace the current ASCII box-stack with a **Mermaid** flowchart-LR (5 nodes: Application → Gateway → Service → Agent → optional A2A/MCP/Tools). GitHub renders Mermaid natively. Same diagram appears in `start-here/overview.mdx` (Spec 2) — single canonical figure, kept in sync via the v0.4 narrative-pivot WI.
2. **Ecom flow walkthrough** — 7 numbered steps lifted verbatim from PRD §7.3 (auth → tenant → recommend → clarification → typed-Go-handle → audit). Numbered list, not prose paragraphs.
3. **Normative response table** — the current README §"Response status × media-type table" stays here, unchanged from v0.3 (PRD §9). Reason: it is a frequent grep target for adopters; demoting or removing it would break adopter muscle memory.

**IA constraints**:
- One diagram. Not two. Not the four currently in `yaagents_docs_structure_review.md` §"Required diagrams" — those belong on Pages (Spec 2), not the README.
- The ecom walkthrough is the **only** prose >150 words in the README. Everything else is tables, code, or lists. (PRINCIPLES.md §1 information-dense.)

### Band 4 — Who uses it (3 personas, table form)

**Goal**: the reader self-identifies in <10 seconds.

**Content shape** — a 3-row table (one row per persona):

| Reader | Why YAAgents | First page |
|---|---|---|
| SaaS product team adding AI features | Keep your existing API surface; add agentic operations as new resource endpoints | `examples/store/` |
| Platform team governing many agent services | One gateway for auth + tenancy + audit + license; framework-agnostic | `docs/plugins.mdx` |
| API architect designing an agent product | Typed outcomes, OpenAPI, native clients — no free-form-text contracts | `spec/agentic-rest-profile.md` |

**IA constraints**:
- **3 rows, not 4** (PI3-yaa user-input had 4 personas in the legacy framing; v0.4 reviewer feedback says 4 is too many to scan in the GitHub viewport). Drop "multi-agent builders" (it's a sub-case of platform team).
- Table is the right primitive here (PRINCIPLES.md §1 — rows when scanning >n items, cards when items need distinct identity). 3 personas don't need distinct identity; the reader is choosing one.
- The current README §"How YAAgents differs from A2A, AGNTCY, MCP, frameworks" comparison content moves to **a single see-also line** under this band: `> See also: [YAAgents vs A2A/AGNTCY/MCP/frameworks](docs/explanation/...)` — link goes to the Pages comparison page (Spec 2 references it). Comparison content is not deleted; it's relocated to where deeper readers find it.

### Band 5 — Next steps (install + 3 links + spec)

**Goal**: exit the README to the right next surface.

**Content shape**:
1. **Install commands** — keep the current 5-block install pattern (pip/npm/go/docker) verbatim. README install instructions are a frequent muscle-memory grep target; do not relocate.
2. **3 next-step links** (no more):
   - Pages site (`https://ai-mpathyminds.github.io/yaagents/`) — "Read the docs"
   - Spec (`spec/agentic-rest-profile.md`) — "Profile v0.3"
   - GitHub Issues — "Open issues / contribute"
3. **License + version lineage** as a final 4-line block (Apache 2.0, v0.4.0, profile v0.3 — PRD §13 + §1).

---

## 4. Component breakdown (per band, tagged)

GitHub Markdown renders no React. The "components" below are Markdown primitives + the one renderable extension (Mermaid).

| Band | Primitive | Tag |
|---|---|---|
| 1 | H1 + tagline + badge row + blockquote callout | `gh-md:` (native GitHub Markdown) |
| 2 | Fenced code block (`http`) + table (3 rows) + link list | `gh-md:` |
| 3 | Mermaid `flowchart LR` block + numbered list + table | `gh-md:` (Mermaid is renderable on github.com) |
| 4 | Table (3 rows) + see-also blockquote | `gh-md:` |
| 5 | Fenced code blocks (`bash`) + link list + footer block | `gh-md:` |

There are **no new components proposed** here — README is plain Markdown by necessity. No promotion to `portfolio/design-system/components/` is in scope.

---

## 5. Accessibility

- **WCAG target**: AA where the surface allows it. GitHub renders the README with its own theme; we control structure, not contrast tokens.
- **Heading order**: strict H1 → H2 → H3, no skipped levels. (Current README is compliant; v0.4 rewrite must remain so.)
- **Alt text**: every badge in the badge row has alt text matching the badge label (current README is correct; preserve in v0.4 rewrite).
- **Mermaid diagram**: GitHub's Mermaid renderer produces SVG; the spec MUST include the source as a fenced code block (which is itself the alt text — screen readers read the code). No separate `alt` parameter is available on Mermaid blocks on github.com.
- **Color contrast**: not specifiable — GitHub theme controls it. The narrative arc never relies on color alone to convey state (PRINCIPLES.md §3).
- **Tabular numerics**: N/A (no numeric columns in this README).

---

## 6. Net-new components proposed

None. README is a Markdown-only surface.

---

## 7. Cross-surface links to keep in sync

| README element | Mirror surface | Sync discipline |
|---|---|---|
| Band 3 architecture Mermaid | `start-here/overview.mdx` (Spec 2 §3.1) | Single source of truth: README is canonical; overview.mdx embeds the same Mermaid block. v0.4 DOC-02 WI updates both in one commit. |
| Band 3 ecom 7-step walkthrough | `case-studies/ecommerce-product-recommendations.mdx` (Spec 2 reference; PRD §7.2 DOC-03) | Per PRD §7.1 OQ-4 user-direct: **same content with cross-link**. Case-study page either embeds README §Band-3 verbatim or cross-links. Single source — README is canonical. |
| Band 5 install commands | Quick Start (`quickstart.mdx`) | Quick Start is the deeper surface; README install is short list. Versions MUST match (v0.4.0 across both). |

---

## 8. Open questions (for architect at A-3)

| ID | Question | Resolution path |
|---|---|---|
| OQ-UX-1 | Should the ecom-flow Mermaid in Band 3 also appear in `examples/store/README.md`? PRD §11 mentions both reference flows but doesn't pin the diagram surface. | Architect decision at A-3 — if YES, add `examples/store/README.md` diagram to the v0.4 DOC-01 WI's "files touched" list. |
| OQ-UX-2 | Should Band 4 personas be a 3-row table or a 3-card `<Card>` grid? Cards don't render on github.com (no Starlight) — recommendation is **table for README** (this spec) + **card grid for `start-here/overview.mdx`** (Spec 2). Confirm. | Default: keep this spec's recommendation. Override via architect note in A-3 DOC-01 WI body if a 3-card layout is desired on README (will degrade gracefully but loses density). |
| OQ-UX-3 | Should the v0.3→v0.4 README diff be one commit or two (Bands 1–3 first; Bands 4–5 follow)? | platform-engineer A-4 NFR call (depends on PR-review batch size). Spec is neutral. |

---

## 9. Acceptance criteria (UX spec — aligns to PRD §7.4)

- [ ] README Band 1 target-audience callout sits above the GitHub fold on 1366×768 viewport (manual check at PR review).
- [ ] A2A/agentic comparison framing is no longer the opener (relocated to Band 4 see-also).
- [ ] Band 2 shows the ecom endpoint table + one fenced clarification_required response — no architecture diagram.
- [ ] Band 3 has exactly one architecture diagram (Mermaid `flowchart LR`); the same diagram source appears in `start-here/overview.mdx` (Spec 2 §3.1).
- [ ] Band 4 has a 3-row persona table; see-also link points to the Pages comparison page.
- [ ] Band 5 install block is verbatim parity with `quickstart.mdx` install snippets (v0.4.0).
- [ ] Single H1; H2/H3 hierarchy strict; no skipped heading levels.
- [ ] PRD §7.4 seed success signal §3 satisfied: 2+ external reviewers report clear understanding reading README → ecom case study → overview.mdx.
