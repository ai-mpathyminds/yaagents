# PI3-yaa — Component: Profile contracts (`spec/`, `schemas/`, `openapi/`) — v0.3 bump

Owner lane: **python-developer** (single small commit; cross-language source-of-truth file).
Sprint 1. Meta-repo root surfaces (NOT submodules per ADR PI3-yaa-0001).

> **Library gate (Gate 3)**: `library_justify: profile-version bump only; spec/schemas/openapi
> are the contract canonical source — they DEFINE the abstraction other libraries depend on.
> No portfolio-shared library applies (these files ARE the abstraction).`

---

### WI-3yaa.SP-1: Profile v0.3 bump on `spec/`, `schemas/`, `openapi/` [DRAFT] — Sprint 1
service: yaagents/spec, yaagents/schemas, yaagents/openapi
parent_feature: F-CONTRACTS
brief: Bump the canonical Profile version from `v0.2` to `v0.3` across all three contract surfaces
per PRD §4 + §5.1 + §5.2 + §5.3. Concretely:

1. `spec/agentic-rest-profile.md` — replace every `v0.2` profile-version literal with `v0.3`.
   Add a `### v0.3 changes` subsection at top of the changelog: "`sdk-go` server SDK component
   added (PRD §5.10); no schema body changes; `Supports-YAAgents-Profile` header value bumps to
   `v0.3`; `X-YAAgents-Profile: v0.3` response header MUST be added by every component."
2. `schemas/` — create `schemas/v0.3/` directory and copy the 6 canonical schemas from
   `schemas/v0.2/` verbatim (PRD §5.2: no body changes in v0.3; only the directory prefix
   changes for forward-compat). Each schema file MUST carry `$schema` (Draft-07 minimum), `$id`
   updated to include `/v0.3/`, and `title` unchanged. Files:
   - `schemas/v0.3/clarification-required.schema.json`
   - `schemas/v0.3/validation-failed.schema.json`
   - `schemas/v0.3/approval-required.schema.json`
   - `schemas/v0.3/conflict.schema.json`
   - `schemas/v0.3/agentic-error.schema.json`
   - `schemas/v0.3/operation-accepted.schema.json`
3. `openapi/yaagents-components.yaml` + `openapi/yaagents-response-profile.yaml` — bump every
   `x-yaagents-profile-version: v0.2` literal to `v0.3`. Verify all `$ref:
   "#/components/schemas/..."` references remain valid (no schema renames in v0.3).
acceptance:
- `grep -rn "v0\.2" spec/ schemas/v0.3/ openapi/` returns 0 hits (excluding the changelog
  back-reference and the v0.2 historical schema dir which remains for back-compat).
- `schemas/v0.3/*.schema.json` ×6 present + valid JSON + each carries Draft-07 `$schema`
  declaration + `$id` includes the `/v0.3/` path segment.
- `cli/` (or any local validator) can run `yaagents validate-openapi openapi/yaagents-components.yaml`
  against the bumped components without error (carry-forward PI2-yaa CLI behaviour against the
  v0.3 surface).
- `YAAgents_PRD_README.md` "Profile version" metadata bumped to `v0.3` (PRD §1 acceptance carry).
library_justify: profile-version bump only; spec/schemas/openapi are the contract canonical source — they DEFINE the abstraction other libraries depend on. No portfolio-shared library applies (these files ARE the abstraction).
depends_on: []
