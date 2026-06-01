# PI2-yaa — Component: Go client SDK (`client-go/`) — NEW

Owner lane: **go-developer**. Sprint 3. Published as Go module
`github.com/ai-mpathyminds/yaagents/client-go` (tag-driven via
`client-go/v0.2.0` per PRD §9.4; publish WI REL-4 in
`release-and-publish.md`).

> **Library gate:** every WI carries `library_justify: novel Go client;
> idiomatic Go analog to client-python/client-ts; stdlib net/http only
> (zero non-stdlib runtime deps per PRD §5.9 design constraints).` Gate 4
> duplication: GOC-* vs PYC-* vs TSC-* are intentional dual-language
> clients of one contract; override noted in this header — the
> Levenshtein-on-brief heuristic does not fire on cross-language pairs.

Design constraints (PRD §5.9):
- Zero non-stdlib runtime dependencies. `net/http`, `encoding/json`,
  `context`, `errors`, `fmt`, `crypto/rand` (UUID v4), `sync` only.
- `context.Context` propagated through every call.
- Minimum Go version: `go 1.22` (matches gateway).
- Module-within-monorepo: `client-go/go.mod` declares `module
  github.com/ai-mpathyminds/yaagents/client-go`.

---

### WI-2yaa.GOC-1: Client + options + headers [DONE] — Sprint 3
service: yaagents/client-go
parent_feature: F-GOCLIENT
brief: Create `client-go/` Go module (`client-go/go.mod` with module path
`github.com/ai-mpathyminds/yaagents/client-go` + `go 1.22`). Implement the
constructor and option set per PRD §5.9:
```go
func New(baseURL string, opts ...Option) *Client
func WithToken(token string) Option
func WithTenantID(id string) Option
func WithCorrelationID(id string) Option
func WithHTTPClient(c *http.Client) Option
```
Default header injection on every request:
- `Authorization: Bearer {token}` (if `WithToken` set)
- `X-Tenant-ID: {tenantID}` (if `WithTenantID` set)
- `X-Correlation-ID: {auto UUID v4}` (override via `WithCorrelationID`)
- `Content-Type: application/json` on bodied requests

UUID v4 generation uses `crypto/rand` directly (no external uuid library) —
12 lines of stdlib code, matches PRD §5.9 zero-non-stdlib constraint.
The default `http.Client` is a value-copy of `http.DefaultClient` with a
30-s timeout; `WithHTTPClient` replaces it entirely (allows TLS pinning
per PRD §10 `[SEC]` client-go).
acceptance:
- `go build ./...` clean; `golangci-lint` clean; `go vet` clean
- `go.mod` declares zero non-stdlib `require` entries (`grep -cE "^require " client-go/go.mod` ≤ 1 toolchain line — verify no `require ( ... )` block has external deps)
- Unit test: `New("http://x", WithToken("t"), WithTenantID("tn"))` then issue a request to a test HTTP server; request headers contain `Authorization: Bearer t`, `X-Tenant-ID: tn`, `X-Correlation-ID: <uuid-v4-shape>`
- Unit test: `WithCorrelationID("custom")` overrides the auto UUID
- Unit test: `WithHTTPClient(&http.Client{Timeout: 1*time.Second})` is used by `(*Client).do` (verified by injecting a transport that records)
- ≥85% coverage on the client + options
library_justify: novel Go client; idiomatic Go analog to client-python/client-ts; stdlib net/http only (zero non-stdlib runtime deps per PRD §5.9).
depends_on: [WI-2yaa.LIC-1]

### WI-2yaa.GOC-2: Resource accessors (Campaigns / ByID / Optimizations / Assets) [WIP] — Sprint 3
service: yaagents/client-go
parent_feature: F-GOCLIENT
brief: Implement the resource-accessor chain per PRD §5.9 API surface:
```go
client.Campaigns()                              // → CampaignsResource
  .ByID(id string)                              // → CampaignResource
  .Optimizations()                              // → OptimizationsResource
    .Create(ctx, body) (*AgenticResult, error)  // POST /campaigns/{id}/optimizations
    .Get(ctx, id) (*AgenticResult, error)       // GET  /campaigns/{id}/optimizations/{id}
  .Assets()                                     // → AssetsResource
    .Generate(ctx, body) (*AgenticResult, error)// POST /campaigns/{id}/assets:generate
```
Resource types are small structs carrying `*Client + parent IDs` only —
no heavyweight builder pattern. Each terminal method:
1. Builds the URL by interpolating parent IDs into the path template.
2. Serialises `body` via `encoding/json`.
3. Calls `(*Client).do(ctx, method, path, body)`.
4. Returns the parsed `*AgenticResult` (parsing logic lives in GOC-3) + a typed error.

`ctx` is passed verbatim to the request (`req.WithContext(ctx)`); caller
cancels on context cancellation.
acceptance:
- `go build ./...` clean
- Unit tests (against `httptest.NewServer`):
  - `Campaigns().ByID("c1").Optimizations().Create(ctx, body)` → `POST /campaigns/c1/optimizations` with JSON body
  - `Campaigns().ByID("c1").Optimizations().Get(ctx, "o1")` → `GET /campaigns/c1/optimizations/o1`
  - `Campaigns().ByID("c1").Assets().Generate(ctx, body)` → `POST /campaigns/c1/assets:generate`
- All 6 PRD §5.9 resource symbols present + tested
- Context cancellation propagates to the in-flight request (server sees client disconnect)
- ≥85% coverage on resources
library_justify: novel Go client; idiomatic Go analog to client-python/client-ts; stdlib net/http only.
depends_on: [WI-2yaa.GOC-1]

### WI-2yaa.GOC-3: AgenticResult + typed errors [READY] — Sprint 3
service: yaagents/client-go
parent_feature: F-GOCLIENT
brief: Implement `AgenticResult` struct + typed-error types per PRD §5.9:
```go
type AgenticResult struct {
    Type           string          // "created" | "success" | "accepted" | "clarification_required" | ...
    Status         int             // HTTP status
    Resource       json.RawMessage // populated for created/success
    RequiredInputs []RequiredInput // clarification_required
    OperationID    string          // accepted
    Message        string
    Trace          Trace
}
func (r *AgenticResult) Err() error
```
Result-parsing logic (in `(*Client).do`):
1. Read body + `Content-Type` from response.
2. Switch on `Content-Type` (matched against the 10-row PRD §4 status×media-type table):
   - `application/json` + 2xx → `Type: "success"|"created"`; `Resource: body`
   - `application/vnd.yaagents.operation+json` → `Type: "accepted"`; populate `OperationID`
   - `application/vnd.yaagents.clarification+json` → `Type: "clarification_required"`; populate `RequiredInputs`; return `&ClarificationRequired{...}` as the error
   - `application/vnd.yaagents.validation-error+json` → `&ValidationFailed{...}`
   - `application/vnd.yaagents.error+json` + 403 → `&AgenticForbidden{...}`
   - `application/vnd.yaagents.error+json` + 424 → `&FailedDependency{...}`
   - `application/vnd.yaagents.error+json` + 500 → generic `&AgenticError{...}` (carries Message + Trace)
   - `application/vnd.yaagents.approval-required+json` → typed `&ApprovalRequired{...}` (additive over PRD §5.9 — covers PRD §4 412 row)
   - `application/vnd.yaagents.conflict+json` → `&Conflict{...}` (covers PRD §4 409 row)
3. `AgenticResult.Err()` returns the typed error for non-success results,
   `nil` otherwise — supports both error-style (`if err != nil`) and
   result-style (`switch result.Type`) callers per PRD §5.9.

Implement four typed errors fully per PRD §5.9 + two more (Conflict +
ApprovalRequired) for full §4 coverage. Each error implements `error`
interface with a useful `Error()` message and carries the `Trace` block
(`CorrelationID`, `RequestID`).
acceptance:
- All 10 PRD §4 response types parse to the correct `AgenticResult.Type` value (golden-corpus driven — see GOC-4)
- `errors.As(err, &target)` works for each typed error
- `result.Err()` returns nil for 2xx and 202; non-nil for 4xx/5xx
- Body-parse errors (truncated JSON, missing fields) → `&AgenticError{Code: "DESERIALIZE_ERROR", Message: <wrapped>}` (no panic)
- ≥85% coverage on result + errors
library_justify: novel Go client; idiomatic Go analog to client-python/client-ts; stdlib net/http only.
depends_on: [WI-2yaa.GOC-2]

### WI-2yaa.GOC-4: Golden corpus conformance tests + idiomatic example [READY] — Sprint 3
service: yaagents/client-go
parent_feature: F-GOCLIENT
brief: Wire the shared golden corpus (PI1-yaa SPEC-5 — `tests/golden/`)
into client-go conformance tests. For each canonical response fixture
(one per the 10 PRD §4 rows), spin up an `httptest.NewServer` that returns
the fixture body + status + content-type; assert that the client parses it
into the expected `AgenticResult.Type` + `Status` + typed-error result.
Add `client-go/examples_test.go` with the idiomatic-usage example from
PRD §5.9 (verified-compileable example doc). Update the `Supports-YAAgents-Profile`
declaration via build-tag-injected constant `ProfileVersion = "v0.2"` in
`client-go/internal/version.go` — clients can read this to assert compatibility.
acceptance:
- 10 golden-fixture tests pass (one per PRD §4 row)
- `go test ./... -race` passes
- `examples_test.go` compiles + runs (verifies docs against real API)
- `client-go.ProfileVersion == "v0.2"` (verified by test)
- README.md added under `client-go/README.md` with PRD §5.9 idiomatic-usage block verbatim + install instructions
library_justify: novel Go client; idiomatic Go analog to client-python/client-ts; stdlib net/http only.
depends_on: [WI-2yaa.GOC-3]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] govulncheck on client-go module | **NFR WI below** | WI-2yaa.NFR-GOC-1 |
| [SEC] TLS root pool injection | feature WI | GOC-1 (`WithHTTPClient` option; caller injects transport with custom root pool / pinning; documented in PRD §10) |
| [SEC] crypto/rand for UUID v4 (no math/rand) | **NFR WI below** | WI-2yaa.NFR-GOC-1 (CI grep gate) |
| [SUPPLY] Go module tag + proxy.golang.org verify | feature WI | REL-4 (`client-go/v0.2.0` tag; `verify-go-module.yml` workflow) |
| [SUPPLY] go.mod module path matches PRD §5.9 | feature WI | GOC-1 (acceptance criterion: `go.mod` `module github.com/ai-mpathyminds/yaagents/client-go`) |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; Go module proxy is free; zero cloud cost implication |

### WI-2yaa.NFR-GOC-1: govulncheck + math/rand absence gate [READY]
service: yaagents/client-go
parent_feature: F-GOCLIENT
brief: [SEC] Two CI checks for the client-go module:
(1) **`govulncheck`**: run `govulncheck ./client-go/...` on every PR +
main push; target 0 HIGH/CRITICAL findings. Expected to be trivially clean
(zero non-stdlib runtime deps per PRD §5.9) but the gate must be wired
regardless so any future transitive dep vulnerability is caught
automatically.
(2) **`math/rand` absence grep**: `grep -rn "math/rand" client-go/`
returns 0 hits; CI FAILs on any match. The correlation-id UUID v4 MUST use
`crypto/rand` (GOC-1 brief; PRD §10 secure-randomness floor) — this gate
prevents accidental `math/rand` adoption in any future patch.
acceptance:
- CI step `govulncheck-client-go` added; exits 1 on HIGH/CRITICAL; passes on v0.2.0 tagged commit
- CI step `no-mathr-rand-client-go` grep exits 0 on clean codebase; exits 1 when `math/rand` injected in a test branch
- `govulncheck ./client-go/...` output contains zero findings on the v0.2.0 commit
library_justify: novel Go client; zero non-stdlib runtime deps per PRD §5.9 — govulncheck expected trivially clean; gate is belt-and-suspenders for future patches.
depends_on: [WI-2yaa.GOC-1]
