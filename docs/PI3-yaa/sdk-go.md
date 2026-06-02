# PI3-yaa — Component: Go server SDK (`sdk-go/`) — NEW in v0.3

Owner lane: **go-developer**. Sprints 1–3 (scaffold → core → example) + S6 (publish tag).
Module path (post split): `github.com/ai-mpathyminds/yaagents-sdk-go` (per ADR PI3-yaa-0002).
Published as Go module via tag `v0.3.0` (SG-7); independent submodule repo
`github.com/ai-mpathyminds/yaagents-sdk-go` (per ADR PI3-yaa-0001).

> **Library gate (Gate 3) — applies to every SG-* WI in this file**: `library_justify: novel
> Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10
> design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).`
> Gate 4 duplication: sdk-go vs sdk-fastapi are intentional cross-language analogs of one contract;
> Levenshtein-on-brief heuristic does not fire on cross-language pairs.

Design constraints (PRD §5.10):
- **Zero non-stdlib runtime dependencies.** `net/http`, `encoding/json`, `context`, `errors`,
  `fmt`, `crypto/rand` (UUID v4 if needed) only. Adapter packages (chi/gin/echo) MAY import
  their respective router (those are adapter-only build deps, not core sdkgo deps).
- **Context-propagation throughout.** Every method takes `context.Context` or `AgenticContext`.
- **Vendor types GENERATED** from `schemas/v0.3/*.json` at build time (codegen step in SG-2).
  Generation is reproducible (same schema input → byte-identical generated Go).
- **Router-agnostic core.** `sdkgo/` package imports zero router framework code; adapters live
  in separate sub-packages `adapters/chi/`, `adapters/gin/`, `adapters/echo/`.
- Minimum Go version: `go 1.22`.

## Cross-lane edge (A-3b)

PRD §11 OQ-6 + planning runbook A-3b: ai-platform/agent-api adopts `sdk-go` on **one** resource
endpoint as canary. **SG-5 (≥80% coverage on `sdkgo/` core; all 10 response-type Status()+
ContentType() unit-tested) MUST be green before ai-platform-side canary WIs dispatch.**
Sequencing: yaagents-side SG-1..SG-5 land in S2 → ai-platform-architect-authored canary WIs
(`ai-platform/docs/PI3-yaa/agent-api-canary.md`) land in S3 or later.

---

### WI-3yaa.SG-1: sdk-go module scaffold + AgenticContext + FromRequest [DRAFT] — Sprint 1
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Create `sdk-go/` directory at meta-repo root (will become standalone repo
`github.com/ai-mpathyminds/yaagents-sdk-go` at RP-SDKGO-INIT in S3). Initialize Go module:
```
sdk-go/
├── go.mod              # module github.com/ai-mpathyminds/yaagents-sdk-go; go 1.22; zero require entries
├── go.sum              # empty (no deps)
├── LICENSE             # Apache 2.0 verbatim
└── sdkgo/
    ├── profile.go      # const ProfileVersion = "v0.3"
    └── context.go      # AgenticContext + FromRequest()
```
`profile.go`:
```go
package sdkgo
// ProfileVersion is the YAAgents Agentic REST Profile version this SDK supports.
const ProfileVersion = "v0.3"
```
`context.go`:
```go
package sdkgo

import "net/http"

type AgenticContext struct {
    CorrelationID string
    RequestID     string
    ActorTenant   string
    Principal     string
}

// FromRequest extracts gateway-injected context from request headers per PRD §5.10.1.
// Header names match the gateway (token-validator + tenant-injector) injection contract:
//   X-Correlation-ID, X-Request-ID, X-Tenant-ID, X-Actor-Principal.
func FromRequest(r *http.Request) AgenticContext { /* ... */ }
```
SPDX header on every source file per ADR PI2-yaa-0003 carry: `// SPDX-License-Identifier: Apache-2.0\n// Copyright 2026 AimpathyMinds`.
acceptance:
- `go build ./...` clean in `sdk-go/`; `golangci-lint run` clean; `go vet ./...` clean.
- `grep -cE "^require [^/]" sdk-go/go.mod` returns 0 (no non-stdlib deps).
- Unit test: `FromRequest(r)` extracts all 4 headers verbatim into `AgenticContext`; missing
  header → empty-string field (no panic).
- Unit test: `sdkgo.ProfileVersion == "v0.3"`.
- SPDX header present on every `.go` file (`grep -L "SPDX-License-Identifier: Apache-2.0" sdk-go/sdkgo/*.go` returns empty).
library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10 design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).
depends_on: [WI-3yaa.SP-1]

### WI-3yaa.SG-2: Vendor type codegen from `schemas/v0.3/*.json` [DRAFT] — Sprint 1
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Generate Go vendor types from `schemas/v0.3/*.json` at build time (PRD §5.10.1 vendor-types
table). Generation MUST be reproducible: same schema input → byte-identical Go output. Place
generated code at `sdk-go/sdkgo/types.go` (single file; not committed as separate per-schema files).

Code-gen approach (chosen for zero-runtime-deps constraint): single `tools/gen/main.go` Go
program in the meta-repo (NOT a runtime dep of `sdkgo`) that reads `schemas/v0.3/*.json` and
emits a deterministically-formatted `types.go`. Invocation: `go run ./tools/gen ./sdk-go/sdkgo/types.go`
(or via `go generate` directive at top of `sdkgo/types.go`).

Generated Go types (one struct per schema per PRD §5.10.1):
- `OperationAccepted` ← `operation-accepted.schema.json`
- `ClarificationRequiredBody` ← `clarification-required.schema.json`
- `ValidationFailedBody` ← `validation-failed.schema.json`
- `ApprovalRequiredBody` ← `approval-required.schema.json`
- `ConflictBody` ← `conflict.schema.json`
- `AgenticErrorBody` ← `agentic-error.schema.json`
- Plus supporting structs: `RequiredInput`, `ValidationError`, `Trace`.

Each generated type has `json:` tags matching the schema property names.
acceptance:
- `go generate ./...` (or explicit `go run ./tools/gen`) produces `sdk-go/sdkgo/types.go` byte-identical to the committed file (CI gate: regenerate + `git diff --exit-code`).
- All 6 schema-derived types + 3 supporting structs present + exported.
- `go build ./...` clean post-generation.
- Unit test: round-trip `json.Marshal` + `json.Unmarshal` on `ClarificationRequiredBody` populated with `RequiredInput` slice; output JSON matches the PRD §4.1 canonical body shape byte-equivalent (ignoring whitespace).
- Each generated `.go` file carries SPDX + `// Code generated by tools/gen; DO NOT EDIT.` first-line directive.
library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10 design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).
depends_on: [WI-3yaa.SG-1, WI-3yaa.SP-1]

### WI-3yaa.SG-3: AgenticResponse factory + Write helper + AgenticWritable interface [DRAFT] — Sprint 2
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Implement the 10 response-type factory methods per PRD §5.10.1 + the `Write()` helper +
the `AgenticWritable` interface. Two files: `sdk-go/sdkgo/response.go` (factory) + `sdk-go/sdkgo/write.go` (writer).

`response.go`:
```go
type AgenticResponse struct{} // zero-value usable

func (AgenticResponse) Accepted(ctx AgenticContext, operationID string) AgenticWritable     // 202 application/vnd.yaagents.operation+json
func (AgenticResponse) Done(ctx AgenticContext, body any) AgenticWritable                    // 200 application/json
func (AgenticResponse) Created(ctx AgenticContext, body any) AgenticWritable                 // 201 application/json
func (AgenticResponse) Failed(ctx AgenticContext, message string) AgenticWritable            // 500 application/vnd.yaagents.error+json
func (AgenticResponse) ClarificationRequired(ctx AgenticContext, inputs []RequiredInput) AgenticWritable
                                                                                              // 400 application/vnd.yaagents.clarification+json
func (AgenticResponse) ValidationFailed(ctx AgenticContext, errors []ValidationError) AgenticWritable
                                                                                              // 422 application/vnd.yaagents.validation-error+json
func (AgenticResponse) ApprovalRequired(ctx AgenticContext, approvers []string, reason string) AgenticWritable
                                                                                              // 412 application/vnd.yaagents.approval-required+json
func (AgenticResponse) Forbidden(ctx AgenticContext, message string) AgenticWritable          // 403 application/vnd.yaagents.error+json
func (AgenticResponse) Conflict(ctx AgenticContext, message string) AgenticWritable           // 409 application/vnd.yaagents.conflict+json
func (AgenticResponse) FailedDependency(ctx AgenticContext, dependency, message string) AgenticWritable
                                                                                              // 424 application/vnd.yaagents.error+json

type AgenticWritable interface {
    Status() int
    ContentType() string
    Body() ([]byte, error)
}
```
`write.go`:
```go
// Write serializes resp to w. Sets status, Content-Type, and X-YAAgents-Profile: v0.3 header.
func Write(w http.ResponseWriter, resp AgenticWritable) error {
    b, err := resp.Body()
    if err != nil { return err }
    w.Header().Set("Content-Type", resp.ContentType())
    w.Header().Set("X-YAAgents-Profile", ProfileVersion)
    w.WriteHeader(resp.Status())
    _, err = w.Write(b)
    return err
}
```
Each method returns an internal struct implementing `AgenticWritable`. The struct carries the
`AgenticContext` so `Trace{CorrelationID, RequestID}` is populated in the response body.
acceptance:
- All 10 factory methods implemented per PRD §5.10.1; each returns a non-nil `AgenticWritable`.
- `Write()` sets status, Content-Type, and `X-YAAgents-Profile: v0.3` header on every response (golden tests via `httptest.NewRecorder()`).
- Per-method status × media-type matches PRD §4 normative table byte-equivalent (table covers all 10).
- `Trace` populated in response body when `AgenticContext` carries `CorrelationID` + `RequestID` (verified by `json.Unmarshal` of recorded body).
- `go build ./...` clean; `go vet` clean; ≥80% line coverage on `response.go` + `write.go`.
library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10 design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).
depends_on: [WI-3yaa.SG-2]

### WI-3yaa.SG-4: Router adapters (chi/gin/echo) [DRAFT] — Sprint 2
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Implement thin adapter sub-packages per PRD §5.10.1 package layout. Each adapter wraps
the `net/http`-native core with router-specific URL-param extraction; the `AgenticContext` /
`AgenticResponse` / `Write` API is identical across all three.

```
sdk-go/adapters/
├── chi/
│   ├── go.mod          # module github.com/ai-mpathyminds/yaagents-sdk-go/adapters/chi
│   │                   # require github.com/go-chi/chi/v5 (build-only; not a core dep)
│   ├── adapter.go
│   └── adapter_test.go
├── gin/
│   ├── go.mod          # require github.com/gin-gonic/gin
│   ├── adapter.go
│   └── adapter_test.go
└── echo/
    ├── go.mod          # require github.com/labstack/echo/v4
    ├── adapter.go
    └── adapter_test.go
```
Each adapter exports:
- `URLParam(r *http.Request, name string) string` — wraps the router's path-param extractor.
- Re-exports of `sdkgo.AgenticContext`, `sdkgo.AgenticResponse`, `sdkgo.Write` (alias style).

Adapters live in independent Go sub-modules so users who only need `net/http` core do NOT pull
in chi/gin/echo into their build graph (preserves PRD §5.10 zero-runtime-deps for core).
acceptance:
- `go build ./adapters/...` clean in all 3 adapter sub-modules.
- `sdk-go/sdkgo/go.mod` `require` block has **zero** non-stdlib entries (verifies adapters do not leak into core).
- Each adapter test mounts an `httptest.NewServer` with a sample handler that calls `FromRequest()` + `Write(ar.Created(ctx, body))` and verifies the 201 + Content-Type + `X-YAAgents-Profile: v0.3` header round-trip via the respective router.
- Each adapter Go file carries SPDX header + Apache 2.0 LICENSE in the sub-package directory.
library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; chi/gin/echo are the three canonical Go HTTP routers (covers 95%+ of Go server use cases); adapter sub-modules isolate their build deps from the core sdkgo package per PRD §5.10 zero-runtime-deps constraint on core.
depends_on: [WI-3yaa.SG-3]

### WI-3yaa.SG-5: Unit tests + ≥80% coverage on `sdkgo/` [DRAFT] — Sprint 2
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Author the unit-test suite for `sdkgo/` core package. Coverage target ≥80% on `sdkgo/`
(SG-1 context + SG-3 response factory + SG-3 Write helper). All 10 response types exercised; each
verifies `Status() int` + `ContentType() string` + `Body() ([]byte, error)` matches the PRD §4
normative table byte-equivalent.

Files: `sdk-go/sdkgo/response_test.go`, `sdk-go/sdkgo/write_test.go`, `sdk-go/sdkgo/context_test.go`.
Golden corpus: `sdk-go/sdkgo/testdata/` with one JSON file per response type containing the
canonical body shape. Test asserts `json.Marshal` of factory output equals the golden.

Specific tests required (per PRD §12 NFR seed):
- Each of the 10 factory methods produces correct (status, Content-Type) pair per PRD §4 table.
- `Write()` sets `X-YAAgents-Profile: v0.3` on every response (10 sub-tests).
- `FromRequest()` populates all 4 `AgenticContext` fields when headers present; empty string when absent; no panic on nil request body.
- `ClarificationRequired()` round-trips `RequiredInput` slice byte-identically per PRD §4.1 canonical shape.
- `Trace{CorrelationID, RequestID}` populated in every body that includes a `trace` field.
acceptance:
- `go test ./sdk-go/sdkgo/... -cover` reports ≥80% line coverage.
- `go test ./sdk-go/sdkgo/... -race` clean (no data races on concurrent `Write()` calls — verify via 100-goroutine fan-out test).
- All 10 PRD §4 status × Content-Type pairs exercised + asserted.
- Golden corpus: 10 JSON files in `testdata/` + 10 tests asserting `json.Marshal` byte-equivalence.
- CI gate: this WI ships before any ai-platform-side canary WI dispatches (cross-lane edge per roadmap §Cross-lane stretch).
library_justify: novel Go server SDK; idiomatic analog to sdk-fastapi; zero non-stdlib runtime deps per PRD §5.10 design constraints; vendor types generated from canonical schemas/v0.3/ (one source of truth).
depends_on: [WI-3yaa.SG-3, WI-3yaa.SG-4]

### WI-3yaa.SG-6: `examples/campaign-api-go/` reference example [DRAFT] — Sprint 3
service: yaagents/examples/campaign-api-go
parent_feature: F-SDKGO
brief: Author the Go reference example mirroring `examples/campaign-api/` (Python). Uses `sdk-go`
+ `net/http` (no adapter) per PRD §8.2. Port `8121` (carry-forward from PI2-yaa; platform-engineer
confirms at A-4 compose-linter — same port as Python campaign-api since both demos run
mutually-exclusively).

Files:
```
examples/campaign-api-go/
├── go.mod                            # module example.com/campaign-api-go; replace directive for sdk-go local
├── main.go                           # net/http server with /campaigns/{id}/optimizations handler
├── Dockerfile                        # multi-stage Alpine; non-root; CGO_ENABLED=0
├── docker-compose.yml                # gateway (yaagents-gateway:0.3.0) + campaign-api-go upstream
├── gateway-routes.yaml               # route config: POST /campaigns/{id}/optimizations → campaign-api-go
├── gateway-plugins.yaml              # token-validator + tenant-injector + license-check
└── README.md                         # quickstart per PRD §8.2 curl example
```
Handler implements all 5 PRD §13.2 / §8.1 flows:
1. Clarification — missing `goal` field → 400 application/vnd.yaagents.clarification+json
2. Created — valid body → 201 application/json
3. Accepted — `Prefer: respond-async` header → 202 application/vnd.yaagents.operation+json
4. Validation failed — invalid field type → 422 application/vnd.yaagents.validation-error+json
5. Auth failure → 401 from gateway (no upstream involvement; gateway-side test)

The handler is a pure `sdkgo.FromRequest(r) → sdkgo.AgenticResponse{} → sdkgo.Write(w, ...)`
sequence per PRD §5.10.1 idiomatic example. NO router framework (validates `net/http`-only core works
in real example).
acceptance:
- `cd examples/campaign-api-go && go build ./...` clean.
- `docker compose up -d && curl -X POST http://localhost:8121/campaigns/cmp-123/optimizations -H "Authorization: Bearer demo-token" -H "X-Tenant-ID: tenant-001" -H "Content-Type: application/json" -d '{"goal":"ctr"}'` returns `201 application/json` with `X-YAAgents-Profile: v0.3` header (PRD §8.2 verbatim curl).
- Missing-`goal` body returns `400 application/vnd.yaagents.clarification+json` with `requiredInputs` array containing `{"name":"goal", ...}`.
- All 5 PRD §13.2 / §8.1 flows exercised by `examples/campaign-api-go/test_e2e.sh` (or Go test file) — script runs in CI and exits non-zero on any flow regression.
- README.md quickstart steps (clone repo → `docker compose up` → curl) work in a clean clone.
library_justify: novel Go server SDK reference example; idiomatic analog to examples/campaign-api/ (Python). Reference example exists to validate the SDK in a real example — not a candidate for portfolio-shared extraction.
depends_on: [WI-3yaa.SG-5]

### WI-3yaa.SG-7: Go module tag `v0.3.0` on yaagents-sdk-go submodule repo [DRAFT] — Sprint 6
service: yaagents/sdk-go
parent_feature: F-SDKGO
brief: Tag-driven Go module publish via `proxy.golang.org` per PRD §10.4. Pre-conditions:
RP-SDKGO-INIT landed (the submodule repo exists at `github.com/ai-mpathyminds/yaagents-sdk-go`
with orphan-baseline + correct `go.mod` `module github.com/ai-mpathyminds/yaagents-sdk-go`
declaration); B-01 PRECHECK returned PASS (no Trusted Publisher needed for Go modules — tag push
suffices).

Operator/CI action sequence:
1. From `yaagents-sdk-go` submodule repo `main` branch: `git tag v0.3.0` + `git push origin v0.3.0`
2. Wait ≤30 min for `proxy.golang.org` index propagation (PRD §10.4 SLO).
3. Verify in fresh Go workspace: `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0`
   succeeds; `go list -m -json github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0` reports the
   correct version + commit SHA.
acceptance:
- `v0.3.0` tag pushed to `github.com/ai-mpathyminds/yaagents-sdk-go` (verified via `gh release list -R ai-mpathyminds/yaagents-sdk-go` showing the tag).
- `proxy.golang.org` returns 200 for the module at `v0.3.0` within 30 min of push.
- `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0` succeeds in a fresh `GOPATH`.
- `ai-platform/agent-api` (the canary consumer per A-3b cross-lane stretch) can `go get` the published version and rebuild successfully (cross-lane verification gate).
- `LA-PI-GATE` (in `launch.md`) cites this WI as evidence for PRD §1 Goals item "Go modules installable from prod".
library_ref: ADR PI3-yaa-0001 (meta-repo + 7 submodule shape); ADR PI3-yaa-0002 (Go module path migration); ADR PI1-yaa-0005 (OIDC trusted publishing — N/A for Go modules but cited for cross-component publishing-discipline consistency).
depends_on: [WI-3yaa.RP-SDKGO-INIT, WI-3yaa.SG-5, WI-3yaa.SG-6]
