# PI3-yaa — Component: Pages site (`docs/` — Astro Starlight on GitHub Pages)

Owner lane: **frontend-developer** (Astro / TypeScript / MDX content) + **python-developer**
(for any cross-language code-snippet examples; rare). Sprint 5. Meta-repo root surface (NOT a
submodule per ADR PI3-yaa-0001).

Target URL: `https://ai-mpathyminds.github.io/yaagents/` (project-pages mode; ADR PI3-yaa-0004).
TLS: GitHub-managed `*.github.io` (zero operator config).
Custom domain (`yaagents.dev`): **DEFERRED** post-traction (intake §A-5 §Migration — 1-WI follow-up).

> **Library gate (Gate 3) — applies to every PG-* WI in this file**: `library_ref: ADR
> PI3-yaa-0004` (GitHub Pages + Astro Starlight choice) + `library_justify: Astro Starlight is
> the canonical Astro docs framework; default theme = no custom component design needed at v0.3
> MVP scope (intake §A-5 + PRD §5.11); no portfolio shared library applies (Pages site is
> yaagents-public-only; web/corporate uses Astro static but no shared theme today; cross-product
> design-system reuse is a v0.4+ consideration).`

Design constraints (PRD §5.11):
- Static site generator: **Astro Starlight** (zero-runtime JS on rendered pages by default).
- Default theme — no custom component design at v0.3 MVP.
- Accessibility target: WCAG AA (Starlight default theme is AA-compliant out-of-box).
- Performance target: Lighthouse ≥90 Perf + A11y (gated by CI; platform-engineer A-4).
- Privacy: zero analytics, zero cookies, zero third-party trackers.

> **A-2.5 UX-architect station**: SKIPPED with evidence per planning runbook
> `context.ux_a25_skip_evidence` — no custom UI components in PI3-yaa scope; Starlight default
> theme suffices.

---

### WI-3yaa.PG-1: Astro Starlight scaffold + Astro config [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Initialize the Astro Starlight project at meta-repo root `docs/`. Per Starlight scaffold
+ Astro config tuned for GitHub Pages project-pages mode (ADR PI3-yaa-0004).

Files to author:
```
docs/
├── package.json              # name: "yaagents-docs"; private: true; license: Apache-2.0
├── pnpm-lock.yaml            # pnpm-managed (matches portfolio-conventions per Gate 5)
├── astro.config.mjs          # site: 'https://ai-mpathyminds.github.io', base: '/yaagents'
├── tsconfig.json             # Starlight default
├── src/
│   ├── content/
│   │   ├── config.ts         # Starlight collection config
│   │   └── docs/             # MDX content files (PG-2..PG-7 populate)
│   └── env.d.ts
├── public/
│   ├── favicon.ico           # AimpathyMinds branding (small)
│   └── og-image.png          # 1200x630 OG image for social sharing
└── README.md                 # local dev: pnpm install && pnpm dev
```

`package.json`:
```json
{
  "name": "yaagents-docs",
  "private": true,
  "license": "Apache-2.0",
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview"
  },
  "dependencies": {
    "astro": "^4.x",
    "@astrojs/starlight": "^0.x",
    "sharp": "^0.x"
  }
}
```

`astro.config.mjs`:
```js
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://ai-mpathyminds.github.io',
  base: '/yaagents',
  integrations: [
    starlight({
      title: 'YAAgents',
      description: 'Source-available Agentic REST Profile',
      social: {
        github: 'https://github.com/ai-mpathyminds/yaagents',
      },
      sidebar: [
        { label: 'Quick Start', link: '/quickstart/' },
        { label: 'Why YAAgents', link: '/why/' },
        { label: 'Profile Spec', link: '/profile/' },
        { label: 'SDK Quickstarts', autogenerate: { directory: 'sdks' } },
        { label: 'Examples', autogenerate: { directory: 'examples' } },
        { label: 'Plugin Authoring', link: '/plugins/' },
        { label: 'Public Roadmap', link: '/roadmap/' },
        { label: 'Contributing', link: '/contributing/' },
        { label: 'Community', link: '/community/' },
      ],
    }),
  ],
});
```

External dep vetting (per `.claude/rules/library-gates.md`): Astro + Starlight are
**well-established OSS** (Astro 4.x stable; Starlight is the canonical Astro docs framework,
adopted by Cloudflare/Bun docs); no governance-auditor co-sign required.
acceptance:
- `pnpm install` clean from `docs/` (no peer-dep warnings on critical paths; pnpm Gate 5 honored).
- `pnpm build` succeeds; output in `docs/dist/` with `index.html` at `/yaagents/` path.
- `pnpm dev` serves locally at `http://localhost:4321/yaagents/`.
- `astro.config.mjs` has `site` + `base` set per ADR PI3-yaa-0004 (verified by grep).
- All 9 sidebar entries declared (Quick Start, Why, Profile, SDK Quickstarts, Examples, Plugins, Roadmap, Contributing, Community).
- LICENSE file at `docs/` is Apache 2.0; `package.json` `license: "Apache-2.0"`.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages at ai-mpathyminds.github.io/yaagents/; Astro Starlight choice).
depends_on: [WI-3yaa.SP-1]

### WI-3yaa.PG-2: Hero + "Why yaagents" content [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Author the landing-page Hero + "Why yaagents" content per PRD §5.11. PRD §11 OQ-3
resolution: product-manager turn drafts Quick Start + Why copy; this WI lifts that prose into
MDX files.

Files:
- `docs/src/content/docs/index.mdx` — landing page with Hero component (Starlight built-in)
  carrying:
  - **Title:** "YAAgents"
  - **Tagline:** "Agentic REST Profile — govern AI agents like ordinary APIs"
  - **Install commands** for all 5 package targets per PRD §1 Goals:
    ```bash
    pip install yaagents-fastapi yaagents-client yaagents-cli
    npm install @aimpathyminds/yaagents-client
    go get github.com/ai-mpathyminds/yaagents-sdk-go
    go get github.com/ai-mpathyminds/yaagents-client-go
    docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0
    ```
  - **GitHub badges:** license (Apache 2.0), latest release version, build status.
- `docs/src/content/docs/why.mdx` — 3-bullet value proposition:
  1. **Domain resources, not `/agents/invoke`**: agentic operations expressed as ordinary
     REST resource actions (`POST /campaigns/{id}/optimizations`); standard HTTP semantics.
  2. **Typed responses for every agentic outcome**: 10 normative response types
     (clarification_required, validation_failed, approval_required, conflict, ...) per the
     Profile spec; not arbitrary JSON blobs.
  3. **Gateway-controlled access + framework-neutral runtime**: token-validator +
     tenant-injector + license-check plugins enforce governance; bring any agent framework
     behind the gateway.
acceptance:
- `docs/src/content/docs/index.mdx` carries the hero block + 5 install commands verbatim per PRD §1.
- `docs/src/content/docs/why.mdx` carries the 3 bullets above.
- `pnpm build` renders both pages without MDX parse errors.
- Pages render in `pnpm preview` mode; manual visual check (frontend-developer self-verifies).
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight).
depends_on: [WI-3yaa.PG-1]

### WI-3yaa.PG-3: Quick Start (10-minute walkthrough) [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Author the Quick Start walkthrough per PRD §5.11. Goal: a developer reading this for the
first time gets a running agentic endpoint in ≤10 minutes. Path: install gateway + sdk-fastapi
+ Python client → run campaign-api example → first agentic call.

File: `docs/src/content/docs/quickstart.mdx`.

Sections:
1. **Prerequisites** — Docker; Python 3.11+; `pip`.
2. **Install** — `pip install yaagents-fastapi yaagents-client` + `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0`.
3. **Run the campaign-api example** — `git clone https://github.com/ai-mpathyminds/yaagents.git && cd yaagents/examples/campaign-api && docker compose up -d`.
4. **First agentic call** (clarification flow):
   ```bash
   curl -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
     -H "Authorization: Bearer demo-token" \
     -H "X-Tenant-ID: tenant-001" \
     -H "Content-Type: application/json" \
     -d '{}'
   # → 400 application/vnd.yaagents.clarification+json with requiredInputs array
   ```
5. **Happy path** — POST with `{"goal":"ctr"}` → 201 application/json.
6. **Next steps** — link to SDK Quickstarts (PG-5) + Examples (PG-6) + Profile Spec (PG-4).
acceptance:
- `docs/src/content/docs/quickstart.mdx` covers all 6 sections above.
- All curl commands return the expected status × Content-Type when run against `examples/campaign-api/` Compose stack (verified manually + via campaign-api CI smoke).
- Walkthrough completes in ≤10 minutes for a developer with Docker + Python + pip installed (informal estimate; honored as a soft acceptance bar).
- Links to PG-5, PG-6, PG-4 resolve.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight).
depends_on: [WI-3yaa.PG-1, WI-3yaa.PG-2]

### WI-3yaa.PG-4: Profile Spec MDX render [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Render `spec/agentic-rest-profile.md` (the canonical normative Profile spec) inline on the
Pages site at `/profile/`. Use Starlight MDX import to pull the markdown from meta-repo root
`spec/` rather than copy/paste (single source of truth for the normative content).

File: `docs/src/content/docs/profile.mdx`.

```mdx
---
title: Agentic REST Response Profile (v0.3)
description: Normative specification — status × Content-Type × body shape.
---

import SpecContent from '../../../../../spec/agentic-rest-profile.md?raw';

<div dangerouslySetInnerHTML={{ __html: SpecContent }} />
```
(Or Starlight's native `<Code>` / markdown import — frontend-developer picks idiomatic Astro
4.x + Starlight MDX import pattern.)

The Profile Spec page MUST render the §4 status × Content-Type table byte-equivalent with the
canonical PRD §4 + `spec/agentic-rest-profile.md` table.
acceptance:
- `docs/src/content/docs/profile.mdx` renders the spec inline (no copy/paste — verified by editing `spec/agentic-rest-profile.md` and seeing the change reflected on `pnpm build`).
- The status × Content-Type table renders correctly (visual check; 10 rows).
- The Profile version banner declares `v0.3`.
- Page link from sidebar resolves to `/profile/`.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight).
depends_on: [WI-3yaa.PG-1, WI-3yaa.SP-1]

### WI-3yaa.PG-5: 6-target SDK Quickstarts [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Author SDK Quickstart pages for 6 targets per PRD §5.11. Each is a single MDX file with:
install command + minimal-viable-snippet + link to the component's submodule repo for full API
reference.

Files in `docs/src/content/docs/sdks/`:
- `python-server.mdx` — sdk-fastapi quickstart (`pip install yaagents-fastapi==0.3.0`; FastAPI `@agentic_operation` example).
- `go-server.mdx` — sdk-go quickstart (`go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0`; net/http example mirroring PRD §5.10.1 idiomatic example).
- `python-client.mdx` — client-python quickstart (`pip install yaagents-client==0.3.0`; `YaAgentsClient(...)` example with clarification handling).
- `ts-client.mdx` — client-ts quickstart (`npm install @aimpathyminds/yaagents-client@0.3.0`; ESM import + `result.type === 'clarification_required'` discriminator example).
- `go-client.mdx` — client-go quickstart (`go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0`; `client.Campaigns().ByID(...).Optimizations().Create(ctx, body)` example).
- `cli.mdx` — cli quickstart (`pip install yaagents-cli==0.3.0`; `yaagents validate-openapi` + `yaagents conformance-test` + `yaagents init fastapi` subcommand examples).

Each snippet is **copy-pasteable** + runnable against the campaign-api example after `docker compose up`.
acceptance:
- All 6 SDK quickstart MDX files present + render.
- Each install command cites the correct registry + version `0.3.0` (or `v0.3.0` for Go modules).
- Each snippet exercises the AgenticResponse / AgenticResult flow per PRD §5.5..§5.10 API surfaces.
- Sidebar autogenerated entry from `sdks/` directory shows all 6 (Astro Starlight `autogenerate`).
- Internal links to Examples (PG-6) resolve.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight).
depends_on: [WI-3yaa.PG-1, WI-3yaa.RP-XREF]

### WI-3yaa.PG-6: Examples walkthroughs (campaign-api + campaign-api-go) [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Author example walkthroughs in `docs/src/content/docs/examples/`. Two examples per PRD
§8 + §5.11:

- `campaign-api.mdx` — Python `examples/campaign-api/` walkthrough; 5 PRD §13.2 / §8.1 flows
  (clarification, created, accepted, validation-failed, auth-failure); curl example per flow;
  expected response body shape per flow.
- `campaign-api-go.mdx` — Go `examples/campaign-api-go/` walkthrough mirroring the Python one;
  PRD §8.2 quickstart curl example; same 5 flows; emphasizes sdk-go API surface.
- `llm-gateway.mdx` (optional — carry-forward from PI2-yaa LLM example) — link to
  `examples/llm-gateway/` for SSE / per-tenant concurrency / execution-timeout flows.
acceptance:
- `campaign-api.mdx` + `campaign-api-go.mdx` present.
- Each walkthrough exercises all 5 PRD §13.2 flows with curl examples + expected responses (status + Content-Type + body shape).
- Internal links to PG-5 SDK Quickstarts resolve.
- `llm-gateway.mdx` exists as a stub linking to the meta-repo `examples/llm-gateway/` directory (carry-forward acknowledgment).
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight).
depends_on: [WI-3yaa.PG-1, WI-3yaa.SG-6, WI-3yaa.RP-XREF]

### WI-3yaa.PG-7: Plugin Authoring + Public Roadmap + Contributing + Community [DRAFT] — Sprint 5
service: yaagents/docs
parent_feature: F-PAGES
brief: Author the remaining 4 sidebar sections per PRD §5.11.

Files:
- `docs/src/content/docs/plugins.mdx` — Plugin Authoring brief. Cover: implement the `Plugin`
  interface (Init/Handler/Shutdown lifecycle, carried from PI2-yaa ADR PI2-yaa-0001); register
  via Go `init()` import side-effect (no `plugin.Open`/`dlopen`); YAML config schema;
  community plugin Go module path convention. Link to `gateway/internal/plugin/` reference
  implementation in the meta-repo.
- `docs/src/content/docs/roadmap.mdx` — Public Roadmap per PRD §11 OQ-4 resolution: features
  framed by quarter (not internal PI IDs). Initial content: "v0.3 shipped (this release)"
  bullet list + "Planned for v0.4 / Q3 2026" bullet list (K8s/Helm, Cosign hardening,
  prompt-sanitize impl, otel-audit impl, multi-backend license-check) + "Community-requested"
  pointer to GitHub Discussions. NO internal PI/WI IDs. NO internal ADR references.
- `docs/src/content/docs/contributing.mdx` — `CONTRIBUTING.md` from meta-repo rendered inline
  (Starlight MDX import like PG-4 Profile Spec). Cover: DCO sign-off; PR checklist; plugin
  contribution path. Carries the legal-review-pending disclaimer banner verbatim per ADR
  PI2-yaa-0003.
- `docs/src/content/docs/community.mdx` — Community Hub: GitHub Discussions link
  (`github.com/ai-mpathyminds/yaagents/discussions`); Code of Conduct link
  (`CODE_OF_CONDUCT.md` rendered inline); "Discord/Slack deferred" callout per PRD §11 OQ-5
  resolution.
acceptance:
- All 4 MDX files present + render without MDX parse errors.
- `roadmap.mdx` contains 0 hits for `PI[0-9]*-yaa` (verified by grep — the public roadmap is community-framed).
- `contributing.mdx` renders the legal-review-pending banner from meta-repo `CONTRIBUTING.md`.
- `community.mdx` links to GitHub Discussions (enabled at LA-PUBLIC-FLIP).
- Sidebar entries resolve.
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages + Astro Starlight); ADR PI2-yaa-0003 (legal-review-pending banner carry-forward).
depends_on: [WI-3yaa.PG-1, WI-3yaa.RP-META]

### WI-3yaa.PG-8: `.github/workflows/pages.yml` build+deploy workflow stub [DRAFT] — Sprint 5
service: yaagents/.github/workflows
parent_feature: F-PAGES
brief: Author the GitHub Actions workflow that builds the Astro site + publishes to GitHub Pages
per ADR PI3-yaa-0004 §Build. **This WI is a stub** — platform-engineer at A-4 wraps the NFR
gates (Lighthouse ≥90 budget, build-failure-blocks-PR) around this workflow scaffolding.

File: `.github/workflows/pages.yml` (at meta-repo root).

```yaml
name: Deploy Pages site
on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - 'spec/**'
      - 'schemas/**'
      - 'openapi/**'
      - '.github/workflows/pages.yml'
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: pnpm
          cache-dependency-path: docs/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
        working-directory: docs
      - run: pnpm build
        working-directory: docs
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: docs/dist
      - id: deployment
        uses: actions/deploy-pages@v4
```
acceptance:
- `.github/workflows/pages.yml` present at meta-repo root.
- Workflow triggers on push to `main` paths-filter for `docs/**` + `spec/**` + `schemas/**` + `openapi/**` + the workflow file itself + manual dispatch.
- Permissions block matches `actions/deploy-pages@v4` requirements (contents:read, pages:write, id-token:write).
- Concurrency group prevents overlapping deploys.
- Workflow first run succeeds on a feature branch (manual dispatch) and `gh-pages` deployment URL returns the rendered site (post LA-PUBLIC-FLIP; in-repo workflow dry-run on a feature branch verifies build-only behavior pre-flip).
- platform-engineer A-4 appends Lighthouse CI step + ≥90 Perf+A11y budget gate (NFR pass; not in this WI's acceptance).
library_ref: ADR PI3-yaa-0004 (Pages on GitHub Pages; `actions/deploy-pages@v4` deploy path).
depends_on: [WI-3yaa.PG-1]
