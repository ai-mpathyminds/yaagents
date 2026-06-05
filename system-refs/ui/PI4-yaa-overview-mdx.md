# PI4-yaa Goal 3 DOC-02 — `start-here/overview.mdx` — UX Spec
Status: [READY]
Product: yaagents (`yaagents/docs/src/content/docs/start-here/overview.mdx` — new file)
Surface: Astro Starlight Pages site (`ai-mpathyminds.github.io/yaagents`)
Touches PRD: `yaagents/system-refs/yaagents-v0.4_detailed.md` §7 Goal 3 + §7.2 DOC-02
Design-system version: `portfolio/design-system/PRINCIPLES.md` v1.0 (2026-04-22) — mandatory Input #1
Spec authored: 2026-06-06 (ux-architect, A-2.5)

> Overview.mdx is **independent of the README** (FF-3 user-direct, PRD §7.1). README is a 45-second GitHub funnel; overview.mdx is a 5-minute Pages read with a left sidebar, search, table-of-contents on the right, and the rest of the docs one click away. The two surfaces share NOTHING by lockstep — but they share one canonical Mermaid diagram, kept in sync because the source is identical, not because the surfaces are coupled.

---

## 1. Surface constraints (read these first)

- **Rendering**: Astro Starlight (theme = `starlight`). Page template is `splash` for the index; overview.mdx uses the default `doc` template (sidebar + TOC + content).
- **Width**: content column ~720px on desktop; sidebar 240px left; TOC 200px right.
- **Audience**: a developer who clicked the README "Read the docs" link, or who Googled "yaagents docs", and is now committing 3–5 minutes. They are NOT first-touch; they have already accepted the README hook.
- **Starlight components available** (used elsewhere in this site — `index.mdx`, `plugins.mdx`): `<Card>`, `<CardGrid>`, `<Aside>` (note/tip/caution/danger), `<Code>` (syntax-highlighted), `<Tabs>` + `<TabItem>`, `<Steps>` (ordered list with visual step numbers), `<LinkCard>`, `<LinkButton>`.
- **What overview.mdx is NOT**: it is not a tutorial (Quick Start owns that), not a spec (Profile owns that), not an architecture deep-dive (Reference Architecture page owns that). It is the **mental-model on-ramp** for the deep-reader who landed on the Pages site.

---

## 2. Narrative arc (Pages-reader funnel)

Five sections, each maps to one Starlight `##` heading. The TOC right rail surfaces these five anchors — that IS the navigation contract.

```
overview.mdx
├── §1 What YAAgents is (one-paragraph definition + the contrast block)
├── §2 Where it fits (the architecture diagram + 5-layer walkthrough)
├── §3 What you get with it (4-card use-case grid)
├── §4 What it is not (boundary-setting — keeps it honest)
└── §5 Where to go next (3 LinkCards branching to Quick Start / Patterns / Reference)
```

The reader can scroll-read end-to-end in ~4 minutes OR jump to §5 directly via TOC. Both must work.

### Anti-patterns being removed (PRD §7.1 + `yaagents_docs_structure_review.md` §"What is not working")

- ❌ Definition-before-motivation. Current site has no `start-here/overview.mdx` — readers fall into `index.mdx` (splash) → must choose between Quick Start / Why / Profile. Overview.mdx is the **missing on-ramp** between splash and the deep pages.
- ❌ Treating overview.mdx as a "what is this" landing page. The README already did that job. Overview.mdx is the **second** explanation, deeper, scoped to a reader who has 5 minutes not 45 seconds.
- ❌ Duplicating the README's 3-persona table. Overview.mdx targets a different reader-state (committed reader, not first-touch) — use cards, not table (see §3.3).

---

## 3. Section-by-section IA

### §1 — What YAAgents is (one paragraph + the contrast block)

**Goal**: in ~3 sentences, ground the reader who skipped the README.

**Content shape**:
1. **One paragraph** (≤4 sentences) defining YAAgents in production terms: gateway + SDKs + Profile, framework-neutral, governs the application↔agent boundary.
2. **`<Code>` block** showing the contrast — same shape as the README has today, but the Pages version uses Starlight `<Code>` (syntax-highlighted) instead of a fenced block. Three lines:
   ```http
   POST /campaigns/{id}/optimizations    ← YAAgents pattern
   POST /tickets/{id}:triage             ← YAAgents pattern
   POST /agents/invoke                   ← what YAAgents replaces
   ```
3. **One `<Aside type="note">`** clarifying scope: "YAAgents is not an agent framework. It is the REST/gateway layer around whatever agent framework you choose."

**IA constraints**:
- ≤200 words total in this section. The reader is being grounded, not taught.
- The `<Aside>` is the section's anti-positioning safeguard — without it, half the reader-confusion in PI3-yaa retro recurs.

### §2 — Where it fits (the architecture diagram)

**Goal**: one diagram, one walkthrough — the mental model.

**Content shape**:
1. **One Mermaid `flowchart TD`** — 6 nodes vertically stacked: Application → Gateway → Agentic API Service → Agent Implementation → Optional A2A/AGNTCY/MCP/Tools. (Source diagram from `yaagents_site_update_instructions.md` §3 Mermaid version, simplified — drop the per-layer leaf nodes like `B1/B2/B3` to keep the diagram below 8 boxes per PRINCIPLES.md §1 information-density via simplicity.)
   - **This is the SAME Mermaid source as README Band 3** (Spec 1 §3 / cross-surface link). Identical text. Updated in lockstep by the v0.4 DOC-02 WI.
2. **Five-layer walkthrough** as a `<Steps>` block (Starlight numbered-step component), one step per architectural layer (Application / Gateway / Service / Agent / Optional ecosystem). Each step is 2–3 lines.
3. **One caption line below the diagram**: "YAAgents governs the application-to-agent boundary. Agent frameworks, A2A, AGNTCY, and MCP can operate behind that boundary." (Verbatim from `yaagents_site_update_instructions.md` §3 caption.)

**IA constraints**:
- ONE diagram in this section. (`yaagents_site_update_instructions.md` proposes 6 diagrams across docs; overview.mdx gets exactly ONE — the architecture/fit diagram. The other five live on dedicated pages — see §5 link-out and Spec 3 for plugin pages.)
- `<Steps>` over prose paragraphs (PRINCIPLES.md §1 information-dense).

### §3 — What you get with it (4-card use-case grid)

**Goal**: the reader sees themselves in one of the four cards.

**Content shape**: a `<CardGrid>` with 4 `<Card>` children, 2×2 on desktop, single-column on mobile (Starlight default). One card per persona/use-case:

| Card title | Body (≤25 words) | Footer link |
|---|---|---|
| Add agents behind existing product APIs | Keep your domain endpoints; expose AI optimization or recommendation as new resource operations. | → Quick Start |
| Govern many agent services in one place | Centralize auth, tenancy, audit, license — your agent services stay simple. | → Plugin overview |
| Keep OpenAPI + SDK discipline | Typed responses, generated clients, no free-form-text parsing. | → Profile spec |
| Wrap any agent framework | LangGraph, Pydantic AI, Semantic Kernel, or custom — YAAgents is the boundary. | → `examples/` |

**IA constraints**:
- 4 cards is the right count per PRINCIPLES.md §1 ("cards only when each item needs distinct identity"). Each persona is a distinct **reader-shape**, not a comparison row — cards are correct here (vs. the README §Band 4 table, which is correct *there*).
- One link per card; lead the reader to ONE next surface, not a menu.
- No icon decoration (PRINCIPLES.md §1, §7 — icons only when paired with text and meaningful). Starlight `<Card>` accepts an `icon` prop; **omit it** for these four cards. Cards are text-led.

### §4 — What it is not (boundary-setting)

**Goal**: prevent the misreading that has dogged PI1-yaa through PI3-yaa.

**Content shape**: a flat bulleted list, 6–8 bullets, each one a "not". Lifted/adapted from PRD-README and `yaagents_site_update_instructions.md` §16:

> YAAgents is **not** an agent framework. Not a chatbot framework. Not an A2A replacement. Not an AGNTCY replacement. Not an MCP replacement. Not an agent reasoning engine. Not a model provider abstraction.

Close with one `<Aside type="tip">` linking forward: "Comparing YAAgents to A2A, AGNTCY, MCP, or LangGraph? See the comparison page." → links to `explanation/yaagents-vs-a2a-agntcy-mcp.mdx` (PRD §8.3 DOCS-02 sibling page).

**IA constraints**:
- The list is **flat**, not a table — these are negations, not comparisons. Tables imply axis comparison; bullets imply enumeration.
- The Aside is the bridge to deeper comparison. Without it, this section reads as defensive; with it, it reads as a scope statement followed by a "here's the deeper read."

### §5 — Where to go next (3 LinkCards)

**Goal**: branch the reader to the right next surface based on intent.

**Content shape**: 3 `<LinkCard>` blocks in a `<CardGrid>`:

| LinkCard | When to choose | Target |
|---|---|---|
| Try it in 10 minutes | "I want to run it." | `/quickstart/` |
| Learn the patterns | "I want to design my agent's API surface." | `/patterns/agentic-rest-endpoint-design/` (PRD §8.3 DOCS sibling, may be PI5-yaa) |
| Read the profile | "I want the normative spec." | `spec/agentic-rest-profile.md` (frozen at v0.3) |

**IA constraints**:
- 3 links, not 5. (PI3-yaa retro: too many next-step choices → analysis paralysis.)
- LinkCards over prose (PRINCIPLES.md §5 — one primary action per surface; here the surface offers three branched paths and the reader picks one).
- If the `/patterns/` page does not yet exist at PI4-yaa close (deferred to PI5-yaa per PRD §4 Out-of-scope reading), this LinkCard MUST be conditionally hidden — not left dangling. Architect at A-3 decides whether to scope-creep `/patterns/` into PI4-yaa DOC track OR omit this LinkCard until PI5-yaa.

---

## 4. Sidebar + TOC contract

Starlight auto-generates the left sidebar from frontmatter and config; the TOC from `##` headings.

- **Sidebar slot**: `start-here/overview.mdx` MUST appear under a new sidebar group `Start Here` (PRD §7.2 implicit; explicit IA call here). Group title from `astro.config.mjs` sidebar config. Order within the group: `Overview` (this page) → `Why YAAgents?` (existing `why.mdx` moved into group) → future siblings.
- **TOC right rail**: 5 anchors auto-generated from the 5 `##` headings (§1–§5 above). All 5 anchors MUST be ≤4 words so they render on a single line in the 200px TOC rail.

---

## 5. Component breakdown (per section, tagged)

| Section | Component | Tag |
|---|---|---|
| §1 | `<Code>` + `<Aside type="note">` | `starlight:` (Starlight built-in) |
| §2 | Mermaid `flowchart TD` block + `<Steps>` | `starlight:` (Mermaid is rendered by the Starlight expressive-code integration) + `gh-md:` (Mermaid block is plain Markdown) |
| §3 | `<CardGrid>` containing 4 `<Card>` | `starlight:` |
| §4 | bulleted list + `<Aside type="tip">` | `starlight:` |
| §5 | `<CardGrid>` containing 3 `<LinkCard>` | `starlight:` |

**Component lane** — all components used are Starlight built-ins. None are from `portfolio/design-system/components/` (which targets shadcn/Radix React apps — wrong stack for a docs site). None are net-new.

---

## 6. Net-new components proposed

None at the portfolio design-system level — the docs site is a separate stack (Astro Starlight, not React+Tailwind+shadcn). Portfolio design system does not govern this surface.

**Future signal** (not in PI4-yaa scope): if YAAgents docs grow a custom Starlight extension component (e.g., a "Profile maturity badge" component used on the per-plugin pages — see Spec 3 §3.6), that would be a yaagents-local component, NOT a portfolio promotion candidate (Starlight ≠ shadcn).

---

## 7. Accessibility

- **WCAG target**: AA. Starlight default theme is AA-compliant out of the box for both light + dark themes — overview.mdx must not regress that.
- **Heading hierarchy**: page H1 from frontmatter `title`; §1–§5 are `##`; sub-anchors are `###`. No skipped levels.
- **Mermaid alt text**: the raw Mermaid source code block IS the alt-text fallback when SVG rendering fails. Screen readers read the code. Caption line (§2 step 3) provides semantic alt for the diagram.
- **Card focus order**: keyboard tab order follows DOM order; Starlight `<Card>` and `<LinkCard>` ship with visible focus rings (PRINCIPLES.md §6 — focus-ring is the one motion that is always essential).
- **Color**: never sole carrier of meaning. The §4 "what it is not" list uses **word** "not" — color (red ✗ etc.) is not used, even though Starlight supports it.
- **Aside semantics**: `<Aside type="note">` (§1) and `<Aside type="tip">` (§4) render with ARIA `role="note"` — appropriate; do not promote to `role="alert"`.
- **Mobile**: Starlight collapses sidebar to a hamburger; TOC moves above content. The 5-section narrative arc still scroll-reads correctly because each `##` is its own scroll target.

---

## 8. Interaction patterns

- **Loading**: N/A — overview.mdx is static MD content; no fetched data.
- **Empty**: N/A — no data-bound surfaces.
- **Error**: N/A unless Mermaid render fails; Starlight fallback shows the source code (acceptable graceful degrade).
- **Optimistic**: N/A.
- **Keyboard navigation**: Starlight defaults — tab cycles through links → cards → sidebar; `Esc` closes mobile sidebar; `/` focuses the search box (Starlight built-in). Do not override.
- **Focus management**: no modal/dialog on this page — focus stays in document flow.

---

## 9. Open questions (for architect at A-3)

| ID | Question | Resolution path |
|---|---|---|
| OQ-UX-1 | The §5 LinkCard "Learn the patterns" points to `/patterns/agentic-rest-endpoint-design/` which is per PRD §4 a PI5-yaa-or-later page. Keep the LinkCard pointing to a 404 (bad UX), OR replace with "Read the plugin overview" (existing `plugins.mdx`), OR omit the third LinkCard until the target lands? | Architect decision at A-3 DOC-02 WI. Default recommendation: **replace with `plugins.mdx` pointer for PI4-yaa**; swap to `/patterns/` when that page lands (PI5-yaa). |
| OQ-UX-2 | The §2 architecture Mermaid is shared with README Band 3 — should the canonical source live in `docs/` or be duplicated in both? | Recommendation: duplicate verbatim. Both surfaces edit independently; a sync drift will be caught at PR review. Single-source via Starlight include is over-engineering for one diagram. |
| OQ-UX-3 | Should overview.mdx have a callout linking to the case-study page (`case-studies/ecommerce-product-recommendations.mdx`, DOC-03)? | Recommendation: NO — overview is the on-ramp, case study is the deep example. Crossing them clutters overview. Case study is reached from `examples/` overview, not from `start-here/overview`. |
| OQ-UX-4 | The current site has a `why.mdx` at top-level. Should it be moved INTO the new `start-here/` group (alongside overview.mdx) or stay at top-level? | Sidebar grouping decision — recommend moving INTO `start-here/`. Architect at A-3 confirms; touches `astro.config.mjs` sidebar config. |

---

## 10. Acceptance criteria (UX spec — aligns to PRD §7.4)

- [ ] File exists at `yaagents/docs/src/content/docs/start-here/overview.mdx`.
- [ ] Page has exactly 5 `##` sections matching the §1–§5 narrative arc here.
- [ ] §2 architecture Mermaid source is identical to README Band 3 Mermaid source (`diff` clean).
- [ ] §3 `<CardGrid>` has exactly 4 `<Card>` entries with no `icon` prop on any of them.
- [ ] §4 "what it is not" list has 6–8 "not X" bullets + one `<Aside type="tip">` linking to the comparison page.
- [ ] §5 has exactly 3 `<LinkCard>` entries (or 2 if OQ-UX-1 resolves to omit).
- [ ] Sidebar shows `Start Here` group containing this page; auto-TOC right rail shows 5 anchors all ≤4 words each.
- [ ] `bin/yaagents-pages-link-audit.sh` returns zero broken links (NFR Pages link-audit gate, PRD §14).
- [ ] No raw `#hex` in any custom CSS (none expected — overview.mdx is plain MDX; PRINCIPLES.md §3).
- [ ] Independent from README per FF-3 (overview.mdx authoring WI is separate from README rewrite WI; touched files do not overlap).
- [ ] PRD §7.4 seed success signal §3 satisfied jointly with Spec 1: 2+ external reviewers report clear understanding reading README → ecom case study → overview.mdx.
