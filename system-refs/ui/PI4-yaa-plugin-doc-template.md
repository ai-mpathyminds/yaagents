# PI4-yaa Goal 2 PLG-doc-* — Per-Plugin Pages Doc Template — UX Spec
Status: [READY]
Product: yaagents (`yaagents/docs/src/content/docs/plugins/{plugin}.mdx` — 5 new files)
Surface: Astro Starlight Pages site
Touches PRD: `yaagents/system-refs/yaagents-v0.4_detailed.md` §6 Goal 2 + §6.2 plugin contracts + §6.3 acceptance
Design-system version: `portfolio/design-system/PRINCIPLES.md` v1.0 (2026-04-22) — mandatory Input #1
Spec authored: 2026-06-06 (ux-architect, A-2.5)

> Five plugin doc pages (token-validator, tenant-injector, license-check, prompt-sanitize, otel-audit) MUST share one shape — so platform readers can scan all five with one mental model. The template here defines the shape; per-plugin content is the architect/PM call at A-3. This spec is the **shape contract**, not the content.

---

## 1. Why a shared template

- **Repeat-scan optimization** (PRINCIPLES.md §1, operator mindset): platform engineers configuring all 5 plugins read all 5 pages in one sitting. Shape consistency lets them speed-skim §2 (Config) across all five and immediately spot the differences. Five different shapes force five re-orientations.
- **Authoring efficiency**: architect at A-3 authors 5 stub bodies against one shape; per-plugin filling at execution time is mechanical, not creative.
- **Stable-badge audit**: PRD §6.3 requires "Coming in v0.4" markers removed and Stable badge present. A shared template means the audit is a one-shot grep, not a five-page eyeball check.

---

## 2. The 6-section shape (mandatory ordering)

Every plugin page has these six `##` sections in this exact order:

```
{plugin-name}.mdx
├── §1 Purpose         (what problem this plugin solves; ≤3 sentences)
├── §2 Config          (YAML config block + field-by-field table)
├── §3 Request/Response  (what the plugin reads from the request; what it writes back; status codes)
├── §4 Security & privacy  (auth surface, secrets handling, PII boundary, threat-model notes)
├── §5 Observability   (what spans/logs/metrics this plugin emits; correlation-id propagation)
└── §6 Failure modes   (what happens when X fails; configurable behavior; what the client sees)
```

The TOC right rail will show 6 anchors. All 6 anchors are exactly the names above. **No section is omitted**, even if its content is "N/A — this plugin doesn't touch PII" — write the "N/A" sentence and move on. Omitting a section breaks the scan-across-5 affordance.

---

## 3. Section-by-section template

### §1 — Purpose

**Content shape**:
- One paragraph (≤3 sentences) — what the plugin does, who it's for, what the alternative would be (do-it-yourself in every service vs. one gateway plugin).
- One `<Badge>` (Starlight built-in) above the H1: maturity badge — `Stable` for PI4-yaa (per PRD §6 acceptance). Color: Starlight's default `success` variant (PRINCIPLES.md §3 — semantic role, not raw color).
- Plugin pipeline position diagram (small Mermaid `flowchart LR` showing where this plugin sits in the 5-plugin chain — highlight the current plugin). Same diagram source on all 5 pages, only the highlighted node differs. Authored once, copied with one-token edit per page.

**IA constraints**:
- No code block in §1. Code starts in §2.
- "Purpose" is reader-state framing — answer "should I read this page or skip it?" in 30 seconds.

### §2 — Config

**Content shape**:
1. **`<Code>` block** — full YAML config example, all fields populated with realistic values (NOT lorem-ipsum, NOT `<your-value-here>` placeholders). Copy from PRD §6.2.{N} "Config shape" YAML for the starting point; architect at A-3 finalizes against discovered implementation.
2. **Field-by-field table** — every key in the YAML gets a row:

| Field | Type | Default | Required | Description |
|---|---|---|---|---|
| `webhook_url` | string (URL) | — | yes | Webhook endpoint the gateway calls to resolve tenant ID. |
| `cache_ttl` | duration | `60s` | no | How long to cache resolved tenant IDs per principal. |
| ... | ... | ... | ... | ... |

3. One `<Aside type="caution">` for any field whose misconfiguration would silently degrade security (e.g. `on_failure: pass-through` on tenant-injector).

**IA constraints**:
- Table is the right primitive (PRINCIPLES.md §1 — rows when scanning >n fields, which all 5 plugins will have).
- Tabular numerics on the `Default` column where values are numeric (`60s`, `300s`) — Starlight tables inherit `font-variant-numeric: tabular-nums` from PRINCIPLES.md §2 by way of the docs site CSS.
- The YAML block goes ABOVE the table, not below. Reader scans the YAML for the field they care about, then drops to the table row for detail.

### §3 — Request/Response

**Content shape**:
- **Reads from request** subsection (`###`): a 1–3 row table listing what headers / claims / body fields the plugin reads.
- **Writes to request** subsection (`###`): a 1–3 row table listing what headers / fields the plugin sets on the request before forwarding upstream. Empty table is permitted — say "this plugin does not modify the forwarded request" if so.
- **Writes to response** subsection (`###`): a 1–3 row table for the same on the response side. Some plugins (e.g. otel-audit) only emit observability, not response modifications — write the explicit "no response modification" sentence.
- **Status codes the plugin can return early** subsection (`###`): a table aligned to PRD §9 normative response table. Each row: HTTP status + media type + when this plugin returns it.

  Example for token-validator:
  | Status | Media type | When |
  |---|---|---|
  | `401` | `application/vnd.yaagents.error+json` | No bearer token; expired token; signature invalid. |
  | `403` | `application/vnd.yaagents.error+json` | Issuer not in configured `issuers:` list; claim mappings fail. |

**IA constraints**:
- Four `###` subsections inside §3 — they are scan targets, not deep prose. (PRINCIPLES.md §5 — visual weight matches importance; here all four are equal-weight reference info.)
- Status-code rows MUST reference the normative table in PRD §9 / response-profile spec — never invent new status/media-type pairings here.

### §4 — Security & privacy

**Content shape**:
- **What this plugin trusts** (`###`): a flat bulleted list — every input the plugin consumes without re-validation (e.g., "the gateway trusts the JWT issuer field after JWKS-signature verification, but does NOT trust unsigned claims").
- **What this plugin protects** (`###`): a flat bulleted list — every attack surface the plugin closes (e.g., "tenant-injector closes the attack vector where a client claims a tenant header directly").
- **PII boundary** (`###`): one paragraph or table — does this plugin see/log/forward PII? If yes, what is the redaction posture?
- **Secrets handling** (`###`): how the plugin receives secrets (env var, file mount, Parameter Store, …) and what it never logs.

**IA constraints**:
- Four `###` subsections, parallel to §3 structure — this is the most-skipped section in plugin docs in practice; the consistent 4-subsection layout makes the "did we forget secrets handling?" check a grep.
- An `<Aside type="danger">` is appropriate for any KNOWN misconfiguration that would create a CVE-class risk (e.g., `on_failure: pass-through` on tenant-injector when `tenantRequired: true` routes exist). Reserve `danger` aside for actual danger — overuse degrades the signal (PRINCIPLES.md §3 — color/iconography is the third channel, not the first).

### §5 — Observability

**Content shape**:
- **Spans / events emitted** subsection (`###`): table — span name, attributes, when emitted. (otel-audit will have rows; other plugins may emit "configuration-loaded" event at boot, etc.)
- **Log lines** subsection (`###`): the structured log lines this plugin emits at INFO / WARN / ERROR. Example log line in a `<Code>` block.
- **Metrics** subsection (`###`): Prometheus-flavored metric names + types + labels.
- **Correlation-id propagation** subsection (`###`): one sentence each — does this plugin read `X-Correlation-ID` from the request, does it write it to spans/logs, does it forward it upstream. (Per PRD §9 gateway responsibilities, every plugin SHOULD propagate.)

**IA constraints**:
- Four parallel `###` subsections again. Pattern-consistency across §3/§4/§5 is the operator-mindset payoff (PRINCIPLES.md §1).
- For plugins that emit no observability of their own beyond the gateway baseline, write the explicit "this plugin emits only the gateway baseline spans/logs/metrics; see `architecture/audit-and-observability.mdx` for the baseline" — DO NOT omit the section.

### §6 — Failure modes

**Content shape**: a table — one row per identifiable failure mode:

| Failure | Configurable behavior | What the client sees |
|---|---|---|
| Webhook unreachable (tenant-injector) | `on_failure: reject` or `pass-through` | `503 application/vnd.yaagents.error+json` (reject) or upstream success (pass-through, logged WARN) |
| JWKS endpoint unreachable (token-validator) | (none — fail closed) | `503 application/vnd.yaagents.error+json` |
| Pattern match (prompt-sanitize) | `strategy: reject` or `redact` | `400 clarification+json` (reject) or redacted body forwarded (redact) |

Close with one `<Aside type="tip">` linking to the integration-test discipline rule (`.claude/rules/integration-test-discipline.md`) — every failure mode listed here SHOULD have an e2e test exercising it (PRD §14 NFR).

**IA constraints**:
- Table column order is fixed across all 5 pages (Failure → Configurable → Client sees). Operator scans the third column to know the contract surface.
- "What the client sees" entries MUST use the normative status × media-type pairings from PRD §9. Inventing new pairings here is a spec violation, not a doc choice.

---

## 4. Cross-page sidebar contract

All 5 plugin pages live under a new sidebar group `Plugins` in the Starlight sidebar (`astro.config.mjs`). Group order:

1. `Plugin overview` (the existing `plugins.mdx` — keep at top of group as the authoring guide / interface contract)
2. `token-validator`
3. `tenant-injector`
4. `license-check`
5. `prompt-sanitize`
6. `otel-audit`

Order matches the gateway pipeline execution order (PRD §6 §"Plugin pipeline (v0.4 Stable)"). The sidebar IS the pipeline diagram, vertically.

---

## 5. Component breakdown (per section, tagged)

All Starlight built-ins. No new components required.

| Section | Components used |
|---|---|
| Page header | `<Badge>` (maturity), Mermaid `flowchart LR` (pipeline position) |
| §2 | `<Code>` (YAML), Markdown table, `<Aside type="caution">` |
| §3 | 4× Markdown tables (one per subsection) |
| §4 | 4× bullet lists or tables, `<Aside type="danger">` (conditional) |
| §5 | 4× Markdown tables + `<Code>` (log line example) |
| §6 | Markdown table + `<Aside type="tip">` |

Tag: `starlight:` for everything. No `ds:` (portfolio design system), no `new:`. No promotion candidates.

---

## 6. Accessibility

- **WCAG target**: AA. Starlight default theme is AA; per-plugin pages must not introduce custom CSS that regresses contrast (PRINCIPLES.md §3, §10 — no raw hex outside tokens).
- **Heading hierarchy**: page H1 from frontmatter; §1–§6 are `##`; subsections inside §3/§4/§5 are `###`. Strict, no level-skipping.
- **Tables**: each table MUST have a header row with semantic `<th>` (Markdown `|---|` syntax gives this automatically). Avoid empty cells — use `—` (em dash) for "not applicable" (em dash announces as "dash" via screen reader; better than empty cell silence).
- **`<Aside>` semantics**: `caution`, `danger`, `tip` map to ARIA roles via Starlight; do not override.
- **`<Badge>` for maturity**: text content `Stable` is the accessible label; color (semantic `success`) is decorative. The word "Stable" is the channel; color reinforces (PRINCIPLES.md §3 — color never sole channel).
- **Mermaid alt text**: source block IS the fallback; caption sentence below each diagram provides semantic context.

---

## 7. Per-plugin variance (what each page diverges on)

The template defines shape; the 5 plugins differ in content. Variance summary for the architect at A-3:

| Plugin | §1 distinguishing line | §4 PII subsection lead | §6 dominant failure mode |
|---|---|---|---|
| token-validator | "Validate JWTs from any configured OIDC/OAuth2 issuer." | "Reads JWT claims; does NOT log raw token." | JWKS endpoint unreachable; signature invalid. |
| tenant-injector | "Resolve tenant ID from principal via configurable webhook." | "Forwards `X-Tenant-ID` header — tenant ID is treated as semi-sensitive." | Webhook unreachable; `on_failure` configurable. |
| license-check | "Gate requests against tenant entitlements." | "Reads `X-Tenant-ID`; license tier may be logged at INFO." | License backend unreachable; unlicensed tenant. |
| prompt-sanitize | "Detect and reject or redact harmful content in payloads." | "INSPECTS request body — PII surface is highest; redaction posture controls log exposure." | Pattern match (configurable reject/redact). |
| otel-audit | "Emit OpenTelemetry spans for every gateway-proxied request." | "Span attributes may carry tenant_id + actor_id — span exporter trust boundary applies." | OTLP exporter unreachable (configurable: fail-open vs queue+drop). |

The architect at A-3 finalizes per-plugin content against discovered implementations (PRD §6.1 user-direct: "architect discovers existing contracts before authoring close WIs").

---

## 8. Maturity-badge contract

Every page header includes `<Badge text="Stable" variant="success" />` (Starlight syntax). PRD §6.3 requires Preview→Stable flip for all 5 plugins. Until a plugin's PI4-yaa Stable-bar criteria are met (feature-complete + tested + documented per PRD §6.1 OQ-2), the page MUST carry `<Badge text="Preview" variant="caution" />` instead.

**The badge IS the audit signal**. `grep -l 'Badge text="Preview"' yaagents/docs/src/content/docs/plugins/*.mdx` at PI4-yaa close MUST return zero results (per PRD §6.3 acceptance criterion + "Coming in v0.4" removal). Spec is the audit recipe.

---

## 9. Net-new components proposed

None. All five pages use existing Starlight built-ins.

**Anti-suggestion**: do NOT promote `<Badge text="Stable" />` into a yaagents-custom `<MaturityBadge>` wrapper component. It would add a Starlight extension for one prop. The plain `<Badge>` is sufficient and is the same surface used elsewhere in Starlight docs across the web (familiar to reader).

---

## 10. Open questions (for architect at A-3)

| ID | Question | Resolution path |
|---|---|---|
| OQ-UX-1 | Should the 6 sections be six `##` headings (as specified) or collapse §3+§5 into one "Runtime contract" section (5 sections)? Five sections compresses but conflates request-shape with telemetry-shape. | Default: keep 6 sections. Architect override at A-3 if a strong reason emerges; spec is explicit because operator scan-across-5 needs shape constancy. |
| OQ-UX-2 | Should §6 failure-mode tables be moved INTO §4 Security (since most failure modes ARE security postures)? | Recommend keeping §6 separate — failure modes include non-security failures (OTLP exporter unreachable is reliability, not security). Conflating loses signal. |
| OQ-UX-3 | The pipeline-position diagram in §1 is shared across all 5 pages, with only the highlighted node differing. Should this be a single shared Mermaid include OR 5 duplicates (one per page)? Starlight supports neither native include AT MDX level (only via `<Fragment>` with limits). | Default: 5 duplicates, kept in sync at PR-review time. Single Mermaid diff per page change is small; over-engineering an include for 5 files is not warranted. |
| OQ-UX-4 | Should the per-plugin pages live at `plugins/{name}.mdx` or `reference/plugins/{name}.mdx`? Current `plugins.mdx` is at top-level and serves as the plugin-authoring guide. | Recommend top-level `plugins/` group (mirroring `examples/`, `sdks/`). Architect at A-3 confirms; touches `astro.config.mjs` sidebar. |

---

## 11. Acceptance criteria (UX spec — aligns to PRD §6.3)

- [ ] Five files exist at `yaagents/docs/src/content/docs/plugins/{token-validator,tenant-injector,license-check,prompt-sanitize,otel-audit}.mdx`.
- [ ] Every file has exactly 6 `##` sections in the order §1–§6 specified above (`grep -c '^## ' {file}` returns 6 on each).
- [ ] Every file has `<Badge text="Stable" variant="success" />` in the page header (or appropriate maturity if a plugin slips Stable bar — flagged at A-4 NFR review).
- [ ] Every file has the pipeline-position Mermaid `flowchart LR` in §1 with the current plugin highlighted.
- [ ] Every §2 has a YAML `<Code>` block + a field-by-field table.
- [ ] Every §3 has the 4 `###` subsections (reads / writes-request / writes-response / status-codes).
- [ ] Every §4 has the 4 `###` subsections (trusts / protects / PII / secrets).
- [ ] Every §5 has the 4 `###` subsections (spans / logs / metrics / correlation-id).
- [ ] Every §6 has the 3-column failure-modes table.
- [ ] Sidebar group `Plugins` in `astro.config.mjs` lists the 6 entries (1 overview + 5 plugins) in the pipeline-execution order specified in §4 here.
- [ ] `grep -l 'Coming in v0.4' yaagents/docs/src/content/docs/**/*.mdx` returns zero results (PRD §6.3 + this spec §8 audit).
- [ ] `bin/yaagents-pages-link-audit.sh` returns zero broken links across the 5 new pages + the cross-links from `production-checklist`, `plugin-pipeline`, `reference-architecture`.
- [ ] All 5 plugin pages render without Mermaid parse errors (Starlight build is clean — A-4 platform-engineer NFR gate).
- [ ] No raw `#hex` color in any Mermaid block (PRINCIPLES.md §3 — Mermaid default theme inherits Starlight's; do not override per-diagram).
- [ ] PRD §6.3 seed success signal §2 satisfied: all 5 plugins ship Stable badge + per-plugin Pages docs + e2e green.
