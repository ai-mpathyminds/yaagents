# yaagents v0.3 — Public launch + sdk-go + repo restructure + Pages site — PRD Seed
Status: [READY]              # flipped at PI2-yaa PC-6 close 2026-06-02 per fast-close ceremony
Target product: yaagents (+ ai-platform/agent-api canary consumer)
Target services: yaagents (meta-repo restructure + 7 submodule split), yaagents/sdk-go (new), yaagents/docs (Astro Starlight Pages site source), portfolio/yaagents-internal/ (internal planning move-target)
Owner: product-manager (yaagents)

## Problem
yaagents v0.2 (PI2-yaa) ships an internally-ready Apache 2.0 runtime —
gateway + 5 plugins + sdk-fastapi + 3 clients + CLI + LLM specialization —
but does NOT make the source repo public, publish to prod public registries,
ship a Go server SDK, or provide community-facing docs. v0.3 is the
**public launch PI**. Four prerequisite tracks must land together for a
real OSS product: (1) **sdk-go** — AmpathyMinds' agent stack is Go-native
(claude-code-cli, not langchain/ollama Python-heavy), and migration onto
yaagents Profile is blocked without a Go server SDK symmetric to
sdk-fastapi; (2) **repo restructure** — split monorepo into a meta-repo +
per-component public submodules for OSS contributor ergonomics (matches
portfolio submodule pattern); (3) **portfolio scrub** — every internal
planning artifact (.claude/, docs/PI*/, system-refs/seeds, internal ADRs,
CLAUDE.md) MUST move OUT of public surface into `portfolio/yaagents-internal/`;
(4) **Pages site** — community-facing docs at `yaagents.dev` (Astro Starlight
on Cloudflare Pages) so users install + succeed in 10 minutes.

## Why now
PI2-yaa close validates the runtime; PI3-yaa publishes it. All four tracks
must land together — partial launches (publish without docs, split-repo
without sdk-go, scrub without Pages) ship a degraded product. PI2-yaa
REL-* WIs are deferred to PI3-yaa per user-direct scope amendment
2026-06-02; PI2-yaa source repo stays PRIVATE through close. The public
launch story (sdk-go + multi-repo + Pages + prod publish + repo public-flip)
is THE PI3-yaa narrative.

## Rough scope
Four parallel tracks (full per-track scope in `portfolio/INTAKE/PI3-yaa-intake.md §Tracks`):
- **TRACK SDK-GO**: `github.com/ai-mpathyminds/yaagents-sdk-go` Go server SDK —
  AgenticResponse helpers (Accepted / Done / Failed), vendor types **generated**
  from `schemas/*.json`, Context helpers, error envelope, profile-version
  constant, **router-agnostic** (net/http / chi / gin / echo); `examples/campaign-api-go/`
  reference; ai-platform/agent-api canary (one resource endpoint end-to-end).
- **TRACK REPO**: split monorepo into 7-component meta-repo (`gateway` +
  `sdk-fastapi` + `sdk-go` + `client-python` + `client-ts` + `client-go` + `cli`) +
  meta-repo root holding `spec/` + `schemas/` + `openapi/` + `examples/` + `docs/`.
  GitHub org stays `ai-mpathyminds/yaagents-*` (move to standalone `yaagents/`
  org defers to v0.4).
- **TRACK SCRUB**: move internal planning artifacts → `portfolio/yaagents-internal/`;
  public submodule repos each start with squashed-history orphan-baseline
  commit; verification script `bin/yaagents-public-mirror-verify.sh` greps
  for portfolio artifacts (`.claude\|portfolio/\|PI[0-9]*-yaa`) → 0 hits.
- **TRACK PAGES**: Astro Starlight site on **GitHub Pages** at
  `https://ai-mpathyminds.github.io/yaagents/` (project-pages mode; free;
  zero operator setup; custom domain `yaagents.dev` deferred until OSS
  traction warrants — 1-WI migration when ready). MVP sections: hero /
  quick start / Profile spec rendered / SDK quickstarts ×6 / examples /
  plugin authoring brief / public roadmap / contributing / community.
  Polished docs (architecture deep-dive, blog, plugin showcase) defer to v0.4.
- **TRACK LAUNCH** (final sprint): publish prod PyPI ×3, prod npm ×1, GHCR
  public multi-arch, Go module tag ×2; flip 8 repos PRIVATE→PUBLIC; push
  `yaagents.dev` site live.

Out: GitHub org move to `yaagents/`-prefixed (defer v0.4+); LLM-vendor
integration; claude-code-cli coupling; tool-use scaffolding; dynamic
agent registration; v0.4 plugin scope (license-check multi-backend,
cookie JWT transport, JTI revocation, OAuth introspection); paid hosting;
analytics/cookies on yaagents.dev.

## Dependencies
- PI2-yaa PC-6 closure — PI3-yaa A-1 cannot open until PC-6 logs (non-waivable chief-architect gate)
- PI2-yaa REL-* deferred prod-publish stages — PI3-yaa TRACK LAUNCH absorbs them
- ~~yaagents.dev domain purchase~~ — NOT REQUIRED for v0.3 per 2026-06-02 user-direct amendment; GitHub Pages serves at `ai-mpathyminds.github.io/yaagents/` (custom domain is a 1-WI post-traction migration; see intake A-5)
- PyPI / npm account org-vs-personal resolution — bus-factor risk flagged in PI2-yaa REL-PRECHECK (OQ-2 in intake)
- `yaagents/schemas/` (canonical JSON Schemas — sdk-go vendor types generated from here)
- `yaagents/spec/agentic-rest-profile.md` (Profile v0.3 spec — sdk-go MUST match)
- `ai-platform/agent-api` (canary consumer; cross-product WI in TRACK SDK-GO final sprint)

## Success signal
yaagents repos PUBLIC at `github.com/ai-mpathyminds/yaagents` (meta) + 7
submodule repos; ALL 5 component packages installable from prod public
registries (`pip install yaagents-fastapi yaagents-client yaagents-cli`,
`npm install @aimpathyminds/yaagents-client`, `go get github.com/ai-mpathyminds/yaagents-client-go`,
`go get github.com/ai-mpathyminds/yaagents-sdk-go`, `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0`);
`https://ai-mpathyminds.github.io/yaagents/` serves the Pages site with Lighthouse ≥90 on Performance + Accessibility;
`ai-platform/agent-api` runs ≥1 yaagents-shaped resource endpoint via sdk-go (canary);
`bin/yaagents-public-mirror-verify.sh` returns 0 hits for portfolio-artifact greps
across all 8 public repos.
