# PI3-yaa — Component: Repo restructure (meta-repo + 7 submodules)

Owner lane: **operator-driven for new GitHub repo creation** (B-01-style mechanical prep) +
**go-developer / python-developer / frontend-developer** for per-submodule orphan-baseline +
package-metadata migration commits. Sprints 3–4.

Target topology (ADR PI3-yaa-0001):

```
github.com/ai-mpathyminds/yaagents                  ← META-REPO (public)
├── README, LICENSE, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT, .gitmodules
├── spec/        schemas/        openapi/        examples/        docs/      ← root, NOT submodules
└── submodules ×7:
    ├── gateway/        → github.com/ai-mpathyminds/yaagents-gateway
    ├── sdk-fastapi/    → github.com/ai-mpathyminds/yaagents-sdk-fastapi
    ├── sdk-go/         → github.com/ai-mpathyminds/yaagents-sdk-go          [NEW PI3-yaa]
    ├── client-python/  → github.com/ai-mpathyminds/yaagents-client-python
    ├── client-ts/      → github.com/ai-mpathyminds/yaagents-client-ts
    ├── client-go/      → github.com/ai-mpathyminds/yaagents-client-go
    └── cli/            → github.com/ai-mpathyminds/yaagents-cli
```

> **Library gate (Gate 3) — applies to every RP-* WI in this file**: `library_ref: ADR
> PI3-yaa-0001` (meta-repo + 7 submodule shape) + `library_ref: ADR PI3-yaa-0002` (Go module path
> migration; applies to RP-GATEWAY-INIT, RP-SDKGO-INIT, RP-CLIGO-INIT) + `library_ref: ADR
> PI3-yaa-0003` (orphan-baseline squashed history).
>
> **Gate 4 duplication finding (7-way INIT WIs)**: Gate 4 **FIRES** — RP-GATEWAY-INIT,
> RP-SDKFASTAPI-INIT, RP-SDKGO-INIT, RP-CLIPY-INIT, RP-CLITS-INIT, RP-CLIGO-INIT, RP-CLI-INIT
> differ only in their per-submodule target. **ARCHITECT OVERRIDE per architect judgment +
> intake §A-1 + PRD §6.2**: each submodule carries distinct per-language package metadata (Go
> `go.mod` module path; Python `pyproject.toml` package name + classifiers + Hatch build target;
> TypeScript `package.json` name + scope + ESM/CJS build config; Docker image metadata) and
> distinct per-component cross-reference updates. Extracting a shared "orphan-baseline-and-migrate"
> abstraction would couple 4 different language toolchains into one shared module — net
> complexity ↑, not ↓.
>
> `duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod /
> Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate
> command sequence is the only superficially shared structure and is 4 shell lines per INIT.
> Extraction would couple 4 language toolchain writers — net complexity ↑.`
>
> The override line above is carried verbatim in each `RP-*-INIT` WI body below per
> `.claude/rules/library-gates.md §Gate 4` requirement.

## Operator pre-condition (mechanical-entry B-prep before any RP-* WI dispatches)

Before any RP-*-INIT WI lands, the operator creates the 7 empty submodule repos on GitHub at
`github.com/ai-mpathyminds/yaagents-{gateway,sdk-fastapi,sdk-go,client-python,client-ts,client-go,cli}`
(empty, PRIVATE for now; flip to PUBLIC at LA-PUBLIC-FLIP). This is operator-driven (B-01-style
mechanical entry; NOT a roadmap WI). Architect surfaces this via the execution runbook
mechanical-entry pattern (`model: none`) authored at A-6.

---

### WI-3yaa.RP-META: Meta-repo public skeleton + community health files [DRAFT] — Sprint 3
service: yaagents (meta-repo root)
parent_feature: F-REPO
brief: Author the meta-repo public-facing skeleton at `github.com/ai-mpathyminds/yaagents` root.
This is the surface OSS users see when they land on the GitHub page.

Files to author / verify at meta-repo root:
- `README.md` — community-facing overview; install commands for all 5 package targets per PRD
  §1 Goals; badges (license / version / build status); link to `docs/` Pages site.
- `LICENSE` — Apache 2.0 verbatim (carry-forward from ADR PI2-yaa-0003).
- `CONTRIBUTING.md` — DCO sign-off requirement; PR checklist; plugin contribution path; carries
  the verbatim "legal-review-pending disclaimer" banner per ADR PI2-yaa-0003 §3.
- `SECURITY.md` — vulnerability reporting address (`security@aimpathyminds.com` or GitHub
  Security Advisories); 90-day disclosure SLA.
- `CODE_OF_CONDUCT.md` — Contributor Covenant v2.1 verbatim.
- `.github/ISSUE_TEMPLATE/` — bug-report + feature-request + plugin-proposal templates.
- `.github/PULL_REQUEST_TEMPLATE.md` — DCO checkbox + linked-issue + changelog-entry checklist.
- Root remains the canonical home for `spec/`, `schemas/`, `openapi/`, `examples/`, `docs/`
  per ADR PI3-yaa-0001 (those are NOT submodules).
acceptance:
- All 5 community health files present at meta-repo root.
- `LICENSE` is Apache 2.0 byte-verbatim (cmp against PI2-yaa LIC-1).
- `README.md` install commands match PRD §1 Goals (5 package targets verbatim).
- `gh repo view ai-mpathyminds/yaagents --json licenseInfo` returns `{"name":"Apache License 2.0"}` (after RP-XREF lands and the repo is recognized).
- `CONTRIBUTING.md` carries the legal-review-pending banner verbatim from ADR PI2-yaa-0003.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI2-yaa-0003 (Apache 2.0 license posture carry-forward).
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2]

### WI-3yaa.RP-SUBMOD: `.gitmodules` + submodule pointers [DRAFT] — Sprint 3
service: yaagents (meta-repo root)
parent_feature: F-REPO
brief: After the 7 submodule repos exist (operator-created empty) AND each has its orphan-baseline
commit (RP-*-INIT WIs landed), add `.gitmodules` to meta-repo root pointing at the 7 submodule
repos. Run `git submodule add github.com/ai-mpathyminds/yaagents-{X}.git <X>/` for each X.

`.gitmodules` shape:
```
[submodule "gateway"]
    path = gateway
    url = https://github.com/ai-mpathyminds/yaagents-gateway.git
[submodule "sdk-fastapi"]
    path = sdk-fastapi
    url = https://github.com/ai-mpathyminds/yaagents-sdk-fastapi.git
...
```

Each submodule pointer is to the orphan-baseline commit SHA from the respective RP-*-INIT WI.
Commit the meta-repo update with message "Add 7 submodule pointers at v0.3.0 orphan-baseline SHAs".
acceptance:
- `.gitmodules` present at meta-repo root with 7 `[submodule "X"]` blocks.
- `git submodule status` returns 7 lines, each pointing at the respective orphan-baseline SHA.
- `git clone --recurse-submodules https://github.com/ai-mpathyminds/yaagents.git` (post LA-PUBLIC-FLIP) succeeds and produces a working tree matching the meta-repo + 7 submodules.
- `examples/campaign-api-go/` + `examples/campaign-api/` + `examples/llm-gateway/` Compose files still resolve to the correct submodule sources (cross-ref updated in RP-XREF).
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape).
depends_on: [WI-3yaa.RP-GATEWAY-INIT, WI-3yaa.RP-SDKFASTAPI-INIT, WI-3yaa.RP-SDKGO-INIT, WI-3yaa.RP-CLIPY-INIT, WI-3yaa.RP-CLITS-INIT, WI-3yaa.RP-CLIGO-INIT, WI-3yaa.RP-CLI-INIT]

### WI-3yaa.RP-GATEWAY-INIT: `yaagents-gateway` orphan-baseline + Go module path migration [DRAFT] — Sprint 4
service: yaagents/gateway (→ github.com/ai-mpathyminds/yaagents-gateway)
parent_feature: F-REPO
brief: Initialize the standalone `yaagents-gateway` submodule repo. Steps (in a working copy of
the new empty `github.com/ai-mpathyminds/yaagents-gateway` repo):

1. `cp -r <internal-monorepo>/yaagents/gateway/* .` (working-tree mirror).
2. Edit `go.mod`: `module github.com/ai-mpathyminds/yaagents-gateway` (was
   `github.com/ai-mpathyminds/yaagents/gateway`).
3. Sweep all internal `import "github.com/ai-mpathyminds/yaagents/gateway/..."` references to the new path:
   `find . -name '*.go' -exec sed -i 's|github.com/ai-mpathyminds/yaagents/gateway|github.com/ai-mpathyminds/yaagents-gateway|g' {} +`
4. `git checkout --orphan public-main`
5. `git add -A`
6. `git commit -m "Initial public release of yaagents-gateway v0.3.0 — squashed mirror of internal monorepo at ai-mpathyminds/yaagents-internal@<sha>"` (orphan-baseline; commit message includes the internal SHA as provenance per ADR PI3-yaa-0003 §3).
7. `git branch -M public-main main`; `git push origin main`.
8. Verify `LICENSE` (Apache 2.0), `CONTRIBUTING.md` (pointer to meta-repo), per-repo `README.md`
   (component overview + install command + link back to meta-repo for community docs).
acceptance:
- New repo `github.com/ai-mpathyminds/yaagents-gateway` has exactly 1 commit on `main` (orphan-baseline).
- `go.mod` declares `module github.com/ai-mpathyminds/yaagents-gateway`.
- `grep -rn "github.com/ai-mpathyminds/yaagents/gateway" .` returns 0 hits (all imports migrated).
- `go build ./...` clean.
- Commit message includes provenance reference: `squashed mirror of internal monorepo at ai-mpathyminds/yaagents-internal@<sha>`.
- LICENSE present at repo root (Apache 2.0).
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0002 (Go module path migration); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-SDKFASTAPI-INIT: `yaagents-sdk-fastapi` orphan-baseline + pyproject migration [DRAFT] — Sprint 4
service: yaagents/sdk-fastapi (→ github.com/ai-mpathyminds/yaagents-sdk-fastapi)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-sdk-fastapi` submodule repo. Steps:
1. `cp -r <internal-monorepo>/yaagents/sdk-fastapi/* .`
2. Edit `pyproject.toml`: keep `name = "yaagents-fastapi"` + `version = "0.3.0"`; bump
   `Supports-YAAgents-Profile` metadata to `v0.3`; verify `license = "Apache-2.0"` classifier
   present + `License :: OSI Approved :: Apache Software License` PyPI classifier.
3. Edit `[tool.hatch.build.targets.wheel]` `packages = ["yaagents_fastapi"]` (or whatever the
   actual Hatch target is — preserve verbatim from PI2-yaa).
4. Orphan-baseline commit per RP-GATEWAY-INIT pattern (steps 4–8).
5. Add per-repo `README.md` + `LICENSE` + `CONTRIBUTING.md` pointing back to meta-repo.
acceptance:
- New repo `github.com/ai-mpathyminds/yaagents-sdk-fastapi` has exactly 1 commit on `main`.
- `pyproject.toml` declares `name = "yaagents-fastapi"`, `version = "0.3.0"`, `Supports-YAAgents-Profile = "v0.3"`, `license = "Apache-2.0"`.
- `python -m build` produces a `.whl` whose `METADATA` shows `License: Apache-2.0` + `Supports-YAAgents-Profile: v0.3`.
- `pip install dist/*.whl && python -c "import yaagents_fastapi; print(yaagents_fastapi.__version__)"` returns `0.3.0`.
- LICENSE present at repo root.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-SDKGO-INIT: `yaagents-sdk-go` orphan-baseline + new Go module path [DRAFT] — Sprint 4
service: yaagents/sdk-go (→ github.com/ai-mpathyminds/yaagents-sdk-go)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-sdk-go` submodule repo (NEW component — no internal
predecessor). Steps:
1. `cp -r <internal-monorepo>/yaagents/sdk-go/* .` (working tree built up by SG-1..SG-6 in S1–S3).
2. Verify `go.mod` declares `module github.com/ai-mpathyminds/yaagents-sdk-go` (set in SG-1; no
   migration needed since the module path was new in v0.3 per ADR PI3-yaa-0002).
3. Orphan-baseline commit per RP-GATEWAY-INIT pattern (steps 4–8).
4. Adapter sub-modules (`adapters/chi/`, `adapters/gin/`, `adapters/echo/`) carry their own
   `go.mod` declaring `module github.com/ai-mpathyminds/yaagents-sdk-go/adapters/{chi,gin,echo}`
   (verify post-mirror — these were authored in SG-4).
5. Add per-repo `README.md` (sdk-go quickstart per PRD §5.10.1 idiomatic example) + `LICENSE` +
   `CONTRIBUTING.md` pointing back to meta-repo.
acceptance:
- New repo `github.com/ai-mpathyminds/yaagents-sdk-go` has exactly 1 commit on `main`.
- `go.mod` declares `module github.com/ai-mpathyminds/yaagents-sdk-go`; `grep -cE "^require [^/]" go.mod` returns 0 (zero non-stdlib deps in core).
- `adapters/chi/go.mod`, `adapters/gin/go.mod`, `adapters/echo/go.mod` present with correct sub-module paths.
- `go build ./...` clean from the repo root + each adapter sub-module.
- `go test ./sdkgo/... -cover` reports ≥80% (SG-5 carry-over).
- LICENSE present at repo root.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0002 (new Go module path `github.com/ai-mpathyminds/yaagents-sdk-go`); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META, WI-3yaa.SG-6]

### WI-3yaa.RP-CLIPY-INIT: `yaagents-client-python` orphan-baseline + pyproject migration [DRAFT] — Sprint 4
service: yaagents/client-python (→ github.com/ai-mpathyminds/yaagents-client-python)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-client-python` submodule repo. Steps mirror
RP-SDKFASTAPI-INIT with the `yaagents-client` Python package metadata:
1. `cp -r <internal-monorepo>/yaagents/client-python/* .`
2. `pyproject.toml`: `name = "yaagents-client"`, `version = "0.3.0"`, `Supports-YAAgents-Profile = "v0.3"`, `license = "Apache-2.0"`.
3. Orphan-baseline commit (steps 4–8 of RP-GATEWAY-INIT).
4. Per-repo `README.md` + `LICENSE` + `CONTRIBUTING.md`.
acceptance:
- 1 commit on `main`; `pyproject.toml` declares `name = "yaagents-client"`, `version = "0.3.0"`, `Supports-YAAgents-Profile = "v0.3"`, `license = "Apache-2.0"`.
- `python -m build` produces a `.whl` whose `METADATA` shows correct license + profile.
- `pip install dist/*.whl && python -c "from yaagents_client import YaAgentsClient"` succeeds.
- LICENSE present.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-CLITS-INIT: `yaagents-client-ts` orphan-baseline + package.json migration [DRAFT] — Sprint 4
service: yaagents/client-ts (→ github.com/ai-mpathyminds/yaagents-client-ts)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-client-ts` submodule repo. Steps:
1. `cp -r <internal-monorepo>/yaagents/client-ts/* .`
2. `package.json`: `"name": "@aimpathyminds/yaagents-client"`, `"version": "0.3.0"`, `"license": "Apache-2.0"`. ESM primary + CJS compat bundle build config (carry-forward from PI2-yaa).
3. Orphan-baseline commit (steps 4–8 of RP-GATEWAY-INIT).
4. Per-repo `README.md` + `LICENSE` + `CONTRIBUTING.md`.
acceptance:
- 1 commit on `main`; `package.json` declares correct name + scope + version + license.
- `pnpm install && pnpm build` produces ESM + CJS bundles in `dist/`.
- `npm pack && tar tf yaagents-yaagents-client-0.3.0.tgz | grep package.json` shows package.json in the tarball.
- LICENSE present.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-CLIGO-INIT: `yaagents-client-go` orphan-baseline + Go module path migration [DRAFT] — Sprint 4
service: yaagents/client-go (→ github.com/ai-mpathyminds/yaagents-client-go)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-client-go` submodule repo. Steps mirror RP-GATEWAY-INIT
with the client-go module path:
1. `cp -r <internal-monorepo>/yaagents/client-go/* .`
2. `go.mod`: `module github.com/ai-mpathyminds/yaagents-client-go` (was `github.com/ai-mpathyminds/yaagents/client-go`).
3. Sweep imports: `find . -name '*.go' -exec sed -i 's|github.com/ai-mpathyminds/yaagents/client-go|github.com/ai-mpathyminds/yaagents-client-go|g' {} +`
4. Orphan-baseline commit (steps 4–8 of RP-GATEWAY-INIT).
5. Per-repo `README.md` + `LICENSE` + `CONTRIBUTING.md`.
acceptance:
- 1 commit on `main`; `go.mod` declares `module github.com/ai-mpathyminds/yaagents-client-go`.
- `grep -rn "github.com/ai-mpathyminds/yaagents/client-go" .` returns 0 hits.
- `go build ./...` clean; existing PI2-yaa unit tests pass post-migration (carry-forward).
- LICENSE present.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0002 (Go module path migration: yaagents/client-go subpath → yaagents-client-go own module); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-CLI-INIT: `yaagents-cli` orphan-baseline + pyproject migration [DRAFT] — Sprint 4
service: yaagents/cli (→ github.com/ai-mpathyminds/yaagents-cli)
parent_feature: F-REPO
brief: Initialize standalone `yaagents-cli` submodule repo. Steps:
1. `cp -r <internal-monorepo>/yaagents/cli/* .`
2. `pyproject.toml`: `name = "yaagents-cli"`, `version = "0.3.0"`, `Supports-YAAgents-Profile = "v0.3"`, `license = "Apache-2.0"`. CLI entry-point: `[project.scripts] yaagents = "yaagents_cli:main"`.
3. Orphan-baseline commit (steps 4–8 of RP-GATEWAY-INIT).
4. Per-repo `README.md` + `LICENSE` + `CONTRIBUTING.md`.
acceptance:
- 1 commit on `main`; pyproject metadata correct.
- `pip install dist/*.whl && yaagents --help` works; `yaagents validate-openapi` + `yaagents validate-response` + `yaagents conformance-test` + `yaagents init fastapi` subcommands present (carry from PI2-yaa).
- LICENSE present.
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0003 (orphan-baseline squashed history).
duplication_override: per-submodule package-metadata + per-language toolchain (Go go.mod / Python pyproject / TS package.json / Docker image) are distinct; orphan-baseline-and-migrate command sequence is the only superficially shared structure and is 4 shell lines per INIT. Extraction would couple 4 language toolchain writers — net complexity ↑.
depends_on: [WI-3yaa.SC-1, WI-3yaa.SC-2, WI-3yaa.RP-META]

### WI-3yaa.RP-XREF: Cross-reference sweep across all 8 repos [DRAFT] — Sprint 4
service: yaagents (meta-repo + 7 submodules)
parent_feature: F-REPO
brief: Update every cross-reference that previously pointed at the monorepo paths to the new
submodule paths. Per PRD §6.4 list:

1. **Go import paths** in `gateway/` + `client-go/` + `sdk-go/` + `examples/campaign-api-go/` —
   covered per-WI in RP-{GATEWAY,CLIGO,SDKGO}-INIT but RP-XREF does final cross-component grep
   `find . -name '*.go' -exec grep -l "github.com/ai-mpathyminds/yaagents/" {} +` returns 0.
2. **Python `pyproject.toml`** in sdk-fastapi + client-python + cli — verified per-WI; XREF
   double-checks `Supports-YAAgents-Profile = "v0.3"` consistent across 3 packages.
3. **TypeScript `package.json`** in client-ts — version + scope + license verified per-WI.
4. **GH Actions workflow files** in each submodule repo — registry URLs reflect new package
   names (e.g. `pypi-publish` workflow targets the new repo path; `npm publish` targets
   `@aimpathyminds/yaagents-client` not the old monorepo). RP-XREF audit:
   `grep -rn "ai-mpathyminds/yaagents/" .github/` across all 8 repos returns 0 hits.
5. **`examples/` Compose files** — image references use the new GHCR path
   `ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` (was `ghcr.io/ai-mpathyminds/yaagents/gateway`).
6. **`docs/` Pages site** SDK quickstart install commands cite new package names + module paths
   (sweep is also done at PG-5 SDK Quickstarts WI; RP-XREF is the final pre-PUBLIC-flip
   verification).
7. **`README.md`** badge URLs (CI status badges + version badges) point at the new submodule
   repos.
acceptance:
- `grep -rn "github.com/ai-mpathyminds/yaagents/" .` across all 8 repos returns 0 hits in source files (excluding LICENSE / CHANGELOG / historical references which are legitimate).
- `grep -rn "ai-mpathyminds/yaagents/gateway\|ai-mpathyminds/yaagents/client-go" .` across all 8 repos returns 0 hits.
- All 4 `pyproject.toml` files carry consistent `Supports-YAAgents-Profile = "v0.3"`.
- All GH Actions workflow files reference correct registry + package paths.
- examples/ Compose files reference `ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` (note new path; no more sub-path).
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0002 (Go module path migration).
depends_on: [WI-3yaa.RP-GATEWAY-INIT, WI-3yaa.RP-SDKFASTAPI-INIT, WI-3yaa.RP-SDKGO-INIT, WI-3yaa.RP-CLIPY-INIT, WI-3yaa.RP-CLITS-INIT, WI-3yaa.RP-CLIGO-INIT, WI-3yaa.RP-CLI-INIT]
