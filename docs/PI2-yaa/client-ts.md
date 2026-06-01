# PI2-yaa — Component: TypeScript client (`client-ts/`) — version bump

Owner lane: **frontend-developer**. Sprint 5. npm package
`@aimpathyminds/yaagents-client`. ADR: PI2-yaa-0003 (Apache 2.0
supersession).

> **Library gate:** `library_ref: ADR PI2-yaa-0003`. API surface unchanged
> from PI1-yaa (PRD §5.7 explicit). `library_justify: API surface
> unchanged; license + profile-version metadata only` applies in spirit
> alongside the ADR reference.

> The client-ts package is a **library, not a UI app** — frontend-developer
> in the yaagents lane authors a typed client SDK. There is no React, no
> bundler beyond `tsup`/`esbuild` for the ESM+CJS dual bundle.

---

### WI-2yaa.BUMP-2: client-ts 0.2.0 (Apache 2.0 metadata + profile v0.2) [READY] — Sprint 5
service: yaagents/client-ts
parent_feature: F-LICENSE
brief: Bump `client-ts/package.json`:
- `"version": "0.2.0"`
- `"license": "Apache-2.0"` (was the v0.1.x source-available string per ADR
  PI1-yaa-0004 — superseded by PI2-yaa-0003)
- Update `"description"` to carry profile `v0.2`
- Update the `yaagents.profile` package-meta field to `v0.2`
- Update the exported `PROFILE_VERSION` constant to `"v0.2"`
Confirm SPDX headers on every `client-ts/src/**/*.ts` source (covered by
LIC-2 sweep). Update README install snippet to `npm install
@aimpathyminds/yaagents-client@0.2.0`. **No API surface change** —
`YaAgentsClient`, `AgenticResult<T>` discriminated union, `result.type
=== 'clarification_required'` discriminant retain v0.1.0 shape (PRD §5.7).
Re-build dual ESM+CJS bundle (ADR PI1-yaa-0003 carries forward unchanged);
both bundles include the updated `PROFILE_VERSION`. Profile-version
assertions in tests bumped `v0.1` → `v0.2`.
acceptance:
- `package.json` declares `"version": "0.2.0"` + `"license": "Apache-2.0"`
- `node -e "import('@aimpathyminds/yaagents-client').then(m => console.log(m.PROFILE_VERSION))"` prints `v0.2` for ESM
- `node -e "console.log(require('@aimpathyminds/yaagents-client').PROFILE_VERSION)"` prints `v0.2` for CJS
- Existing PI1-yaa golden-corpus tests pass after profile-version assertion bump
- `npm pack --dry-run` reports `license: Apache-2.0` in the tarball metadata
- README install snippet matches new version
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.LIC-2]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] pnpm/npm audit on v0.2.0 package | **NFR WI below** | WI-2yaa.NFR-TS-1 |
| [SUPPLY] `"license": "Apache-2.0"` in package.json; no Community-License string | feature WI | BUMP-2 (acceptance criterion); REL-2 pre-publish gate |
| [SUPPLY] OIDC trusted publishing via npm provenance | feature WI | REL-2 in `release-and-publish.md` (`npm publish --provenance` via OIDC carry-forward) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; npm is a public registry; zero cloud cost |

### WI-2yaa.NFR-TS-1: pnpm audit CI gate for client-ts [READY]
service: yaagents/client-ts
parent_feature: F-LICENSE
brief: [SEC] Run `pnpm audit --audit-level high` (or `npm audit --audit-level high`
matching the lockfile tool) against `client-ts/` on every PR + main push
(carry-forward from PI1-yaa REL-6 CI matrix). CI step `pnpm-audit-client-ts`
in `.github/workflows/ci.yml`; exits 1 on HIGH/CRITICAL advisory. The
`client-ts` package is a library with no runtime deps beyond optional
`tsup`/`esbuild` dev tooling; advisory clean is expected trivially but
the gate is mandatory.
acceptance:
- CI step `pnpm-audit-client-ts` (or `npm-audit-client-ts`) present; exits 1 on HIGH/CRITICAL advisory
- Audit passes on v0.2.0 tagged commit (0 findings)
- Tool used (`pnpm` vs `npm`) matches the lockfile present in `client-ts/` (`pnpm-lock.yaml` → pnpm; `package-lock.json` → npm)
library_ref: ADR PI2-yaa-0003
depends_on: [WI-2yaa.BUMP-2]
