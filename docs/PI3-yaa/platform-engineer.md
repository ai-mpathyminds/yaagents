# PI3-yaa — NFR / Supply-Chain Pass (platform-engineer, A-4)

> Authored: 2026-06-02. Agent: platform-engineer (workspace-root .claude/agents/platform-engineer.md).
> All WIs below land at [READY] — A-4 is the NFR-gate that flips the whole roadmap.
> PRD ref: `yaagents/system-refs/yaagents-v0.3_detailed.md §12 NFR Seeds`.
> bin artifact: `bin/yaagents-public-mirror-verify.sh` (TRACK SCRUB acceptance; authored this turn).

**[FIN] N/A** — PI3-yaa makes ZERO AWS substrate changes. GitHub Pages + PyPI + npm + GHCR +
Go module proxy are all free-tier public services. No TF edits. No new IAM roles. No ECS tasks.
Portfolio run-rate remains ≤$150/mo. (`portfolio/infrastructure/` not in `target_services`.)

---

## Sprint 2 — sdk-go CI gates

### WI-3yaa.NFR-SEC-1: sdk-go CI — govulncheck + zero-core-deps + reproducible-codegen + ≥80% coverage [READY] — Sprint 2
service: yaagents/sdk-go (.github/workflows/sdk-go-ci.yml)
parent_feature: F-SDKGO
sprint: S2
brief: |
  Author `.github/workflows/sdk-go-ci.yml` in the meta-repo (runs on push/PR paths `sdk-go/**`).
  Four gate steps — all must pass green before sdk-go developer WIs beyond SG-2 can merge:

  1. **govulncheck**: `go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./sdk-go/...`
     Fails on any known Go vulnerability in sdk-go or its transitive deps.

  2. **Zero non-stdlib core deps**: verify `sdk-go/go.mod` declares no external `require` lines
     (only the stdlib is permitted per PRD §5.10; adapter sub-packages are tested separately).
     Check: `grep -E '^require ' sdk-go/go.mod | grep -v '^\s*//' | wc -l` must equal 0.
     Adapter packages (`sdk-go/adapters/chi/`, `sdk-go/adapters/gin/`, `sdk-go/adapters/echo/`)
     have their own go.mod or live in the sdk-go module but ONLY import their router framework
     at the adapter layer — the core `sdkgo/` package must remain zero-external-dep.

  3. **Reproducible codegen**: re-run the schema codegen script (`go generate ./sdk-go/...`) then
     `git diff --exit-code sdk-go/`. Any diff means the committed generated code is stale; CI fails.
     Requires codegen tooling (quicktype or custom generator per SG-2) is pinned by version in the
     workflow.

  4. **Coverage ≥80%**: `go test -coverprofile=cov.out ./sdk-go/sdkgo/...`
     `go tool cover -func=cov.out | awk '/total/{gsub(/%/,""); if ($3+0 < 80) {print "Coverage "$3"% < 80% threshold"; exit 1}}'`
     Must include all 10 response-type `Status()` + `ContentType()` unit tests per PRD §12.
acceptance:
- `.github/workflows/sdk-go-ci.yml` present; `on: push: paths: [sdk-go/**]` + `on: pull_request: paths: [sdk-go/**]`.
- All 4 gates defined as sequential steps in one job.
- Workflow runs green on a known-passing sdk-go state (SG-5 green = this gate green).
- govulncheck step fails the job on any high-severity finding.
- Zero-deps step fails if any external `require` entry in `sdk-go/go.mod` core module.
- Reproducible-codegen step fails if `git diff` is non-empty post-regeneration.
- Coverage step fails if total < 80%.
library_justify: "CI gate infrastructure; no external library adoption — govulncheck is the Go toolchain's own vulnerability scanner; no third-party testing framework."
depends_on: [WI-3yaa.SG-2]

---

## Sprint 3 — campaign-api-go Compose smoke CI

### WI-3yaa.NFR-SRE-1: campaign-api-go Compose smoke CI [READY] — Sprint 3
service: yaagents/examples/campaign-api-go (.github/workflows/sdk-go-smoke.yml)
parent_feature: F-SDKGO
sprint: S3
brief: |
  Author `.github/workflows/sdk-go-smoke.yml` in the meta-repo (triggered on push/PR to
  `sdk-go/**` and `examples/campaign-api-go/**`). Validates the reference example end-to-end
  in CI using Docker Compose.

  Steps:
  1. `cd examples/campaign-api-go && docker compose up --detach --wait --wait-timeout 60`
  2. Happy path: `curl -sf -X POST http://localhost:8121/campaigns/c1/optimizations \`
     `-H 'Content-Type: application/json' -d '{"brief":"test"}' | grep '"status":201'`
  3. Clarification path: `curl -sf -X POST http://localhost:8121/campaigns/c1/optimizations \`
     `-H 'Content-Type: application/json' -d '{"brief":"?"}' | grep '"status":400'`
     Response Content-Type must match `application/vnd.yaagents.clarification+json`.
  4. `docker compose down --volumes`

  Port 8121 is the yaagents `campaign-api` reference example port per portfolio conventions table.
  Smoke steps mirror the PRD §13.2 Compose end-to-end flows (5 flows; happy + clarification are
  the minimum automated subset; the full 5 are manual-verified at LA-PI-GATE).
acceptance:
- `.github/workflows/sdk-go-smoke.yml` present with trigger paths and 4 steps above.
- CI step exits 0 on a working campaign-api-go service (SG-6 green).
- curl returns HTTP 201 on the happy path and HTTP 400 on the clarification path.
- `docker compose down --volumes` runs unconditionally (even on failure) via `if: always()` step.
library_justify: "CI smoke test using Docker Compose (stdlib infra); no third-party test library."
depends_on: [WI-3yaa.SG-6]

---

## Sprint 5 — Pages Lighthouse CI gate

### WI-3yaa.NFR-SRE-2: Pages Lighthouse ≥90 Perf+A11y CI gate [READY] — Sprint 5
service: yaagents/.github/workflows/pages.yml
parent_feature: F-PAGES
sprint: S5
brief: |
  Wrap the NFR gates around the PG-8 workflow stub (`.github/workflows/pages.yml`).
  PG-8 provides the Astro build + deploy skeleton; this WI ADDS the Lighthouse budget step.

  Addition to pages.yml (as a separate `lighthouse` job that runs after the `build` job on PRs;
  does NOT run on push to main — deploy-only on main):

  ```yaml
  lighthouse:
    runs-on: ubuntu-24.04
    needs: build
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with: { name: github-pages, path: ./dist }
      - uses: treosh/lighthouse-ci-action@v12
        with:
          uploadArtifacts: true
          budgetPath: .github/lighthouse-budget.json
  ```

  And the budget file `.github/lighthouse-budget.json`:
  ```json
  [
    {
      "path": "/*",
      "resourceSizes": [],
      "scores": [
        { "id": "performance",  "minScore": 0.9 },
        { "id": "accessibility","minScore": 0.9 }
      ]
    }
  ]
  ```

  `treosh/lighthouse-ci-action@v12` is the canonical LHCI GitHub Action; no long-lived tokens;
  runs headless Chromium inside the GitHub Actions runner; uses the pre-built dist artifact from
  the `build` job.
acceptance:
- `.github/lighthouse-budget.json` present with `performance` + `accessibility` minScore ≥ 0.9.
- `lighthouse` job added to `pages.yml`; triggers on PRs touching `docs/**`.
- Job fails the PR if either score drops below 0.9 on any route.
- Action uses `treosh/lighthouse-ci-action@v12` (pinned major version) — no API key required.
- `build` job failure gates out the `lighthouse` job (transitive via `needs: build`).
library_justify: "treosh/lighthouse-ci-action is the standard Lighthouse CI integration for GitHub Actions; no npm package adoption — workflow-level action only."
depends_on: [WI-3yaa.PG-8]

---

## Sprint 6 — Pre-launch supply-chain gates

### WI-3yaa.NFR-SUP-1: OIDC-only registry publish audit + scrub-verify CI integration [READY] — Sprint 6
service: yaagents/.github/workflows/supply-chain-audit.yml
parent_feature: F-LAUNCH
sprint: S6
brief: |
  Author `.github/workflows/supply-chain-audit.yml` — a meta-repo CI job that runs two
  independent checks before LA-PUBLIC-FLIP is allowed to proceed.

  **Check A — OIDC-only publish (no long-lived registry tokens)**
  Grep all 8 repos' `.github/workflows/` directories for disallowed patterns:
  - Any `REGISTRY_PASSWORD`, `PYPI_API_TOKEN`, `NPM_TOKEN` secret references
    (old-style tokens; forbidden; OIDC Trusted Publisher replaces them).
  - Any `docker login --password` with a static secret (GHCR login must use
    `permissions: packages: write` + OIDC `docker/login-action` with `registry: ghcr.io`).
  Allowed: `GITHUB_TOKEN`, `NODE_AUTH_TOKEN` (npm provenance OIDC flow only).
  Implementation: `grep -rn "REGISTRY_PASSWORD\|PYPI_API_TOKEN\|NPM_TOKEN\|--password.*secrets"` across
  `.github/workflows/` in each checked-out repo must return 0 hits. Exit 1 on any match.

  **Check B — `bin/yaagents-public-mirror-verify.sh` CI gate**
  Run `bin/yaagents-public-mirror-verify.sh` against each of the 8 locally-cloned working trees
  (the workflow checks them out as steps). Exits non-zero if any internal artifact leaks are found.
  This wires the bin script (authored at A-4) into CI so it runs automatically on every push to
  meta-repo main after S4 (RP-XREF) lands.

  Workflow trigger: `on: push: branches: [main]` (after TRACK REPO completes) + `on: workflow_dispatch`.
  `workflow_dispatch` allows operator to re-run pre-flip as a manual gate.
acceptance:
- `.github/workflows/supply-chain-audit.yml` present; both Check A + Check B steps defined.
- Check A grep returns 0 hits on a clean repo set (no PYPI_API_TOKEN / REGISTRY_PASSWORD).
- Check B (`bin/yaagents-public-mirror-verify.sh`) returns exit 0 on a clean post-scrub working tree.
- `workflow_dispatch` trigger enabled so operator can run manually before LA-PUBLIC-FLIP.
- No long-lived registry tokens introduced by this workflow itself.
library_justify: "Pure shell grep + bin script; no external library adoption."
depends_on: [WI-3yaa.RP-XREF]

### WI-3yaa.NFR-SUP-2: Apache-2.0 license-clean CI gate across all source files [READY] — Sprint 6
service: yaagents/.github/workflows/supply-chain-audit.yml
parent_feature: F-LAUNCH
sprint: S6
brief: |
  Add a third check to `.github/workflows/supply-chain-audit.yml` (same file as NFR-SUP-1):

  **Check C — Apache-2.0 license-clean scan**
  Grep all source files across all 8 repos for "Community License" (the old v0.1.x dual-license
  term that must not appear in v0.3.0 artifacts per PRD §2 + ADR PI2-yaa-0003):

  ```bash
  grep -rn "Community License" \
    --include="*.go" --include="*.py" --include="*.ts" \
    --include="*.md" --include="*.toml" \
    "$REPO_PATH" 2>/dev/null
  ```

  Must return 0 hits. Exit 1 on any match.

  Also verify SPDX Apache-2.0 header presence on Go source files (sdk-go + gateway):
  ```bash
  find "$REPO_PATH" -name "*.go" | xargs grep -L "SPDX-License-Identifier: Apache-2.0" | head -5
  ```
  Any Go file lacking the SPDX identifier is a WARN (not FAIL at v0.3.0 — header adoption is
  progressive; the FAIL trigger is "Community License" text only). Warn count reported to operator.

  This gate fires pre-launch (Sprint 6) to catch any accidental Community License text in the
  newly-authored sdk-go source, the orphan-baseline commits, or the Pages site content.
acceptance:
- "Community License" grep returns 0 hits across all .go, .py, .ts, .md, .toml files in all 8 repos.
- SPDX Apache-2.0 WARN count ≤ threshold communicated to operator (non-blocking for this gate).
- CI step is part of `supply-chain-audit.yml` (Check C; runs in same job or parallel step to A + B).
- Gate runs on push to meta-repo main (Sprint 6+) and on manual `workflow_dispatch`.
library_justify: "Pure grep; no license-scanning tool adoption — direct text match on the forbidden license string per ADR PI2-yaa-0003."
depends_on: [WI-3yaa.RP-XREF, WI-3yaa.NFR-SUP-1]
