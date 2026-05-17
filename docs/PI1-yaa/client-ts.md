# PI1-yaa — Component: TypeScript client (`client-ts/`)

Owner lane: frontend-developer (a **library**, not a UI app). Sprint 4.
Published as npm `@yaagents/client`. ADR: PI1-yaa-0002, PI1-yaa-0003 (dual
ESM+CJS, zero runtime deps).

> Duplication note: see `client-python.md` — PYC/TSC are intentional
> dual-language clients of one contract; Gate 4 heuristic does not fire.

---

### WI-1yaa.TSC-1: `YaAgentsClient` + fluent resource accessors [DRAFT]
service: yaagents/client-ts
brief: `YaAgentsClient({ baseUrl, token, tenantId })`; `client.campaigns
.byId(id)` → `CampaignResource`; `.optimizations().create(body)` →
`POST /campaigns/{id}/optimizations`; `.assets().generate(body)`. Default
headers (Authorization/X-Tenant-ID/auto X-Correlation-ID). Uses global
`fetch` (Node ≥18 / browser) — **zero runtime deps**.
acceptance:
- Headers injected; correlation-id auto + overridable
- No `dependencies` in package.json (devDeps only)
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.SPEC-1]

### WI-1yaa.TSC-2: `AgenticResult<T>` discriminated union + `strict()` [DRAFT]
service: yaagents/client-ts
brief: `AgenticResult<T>` discriminated union over the response types
(`created`→`resource:T`, `clarification_required`→`requiredInputs`,
`validation_failed`→`errors`, `failed_dependency`→`dependency`,
`accepted`→`operationId`, …). Default = no-throw, caller switches on
`result.type`. `client.strict()` wrapper that throws typed errors on
non-success. First-class `.d.ts` types.
acceptance:
- Exhaustive `switch` over `result.type` type-checks (no `default` needed)
- `strict()` throws typed error matching the vendor type; non-strict never throws
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.TSC-1, WI-1yaa.SPEC-2]

### WI-1yaa.TSC-3: Dual ESM+CJS build + corpus tests [DRAFT]
service: yaagents/client-ts
brief: `tsup` build emitting `dist/index.mjs` + `dist/index.cjs` + bundled
`dist/index.d.ts`; `package.json` `exports` import/require map; `yaagents
.profile = "v0.1"` + exported `PROFILE_VERSION`. vitest suite replaying the
`spec/examples/v0.1` golden corpus via a mock fetch. Coverage ≥80%.
acceptance:
- Both `import` and `require` entrypoints load + resolve types; tree-shakeable ESM
- vitest green vs corpus; coverage ≥80%; `eslint`+`tsc --noEmit`+`npm audit` clean
library_justify: novel; standalone OSS surface
depends_on: [WI-1yaa.TSC-2, WI-1yaa.SPEC-5]
