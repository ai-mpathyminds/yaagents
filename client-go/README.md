# yaagents Go client SDK

Go client for the **YAAgents Agentic REST Profile** — idiomatic resource-oriented
access to agentic operations with zero non-stdlib runtime dependencies.

Supports **YAAgents Profile v0.2** · `ProfileVersion = "v0.2"`

---

## Install

```bash
go get github.com/ai-mpathyminds/yaagents/client-go@v0.2.0
```

Go 1.22+ required. Zero non-stdlib runtime dependencies (`net/http`, `encoding/json`,
`crypto/rand`, `context` only).

---

## Idiomatic usage

```go
import "github.com/ai-mpathyminds/yaagents/client-go"

client := yaagentsclient.New(
    "http://localhost:8120",
    yaagentsclient.WithToken("my-jwt"),
    yaagentsclient.WithTenantID("tenant-001"),
)

result, err := client.Campaigns().ByID("cmp-123").Optimizations().Create(ctx, map[string]any{
    "goal": "reduce_cost_per_lead",
})
if err != nil {
    var clarify *yaagentsclient.ClarificationRequired
    if errors.As(err, &clarify) {
        for _, input := range clarify.RequiredInputs {
            fmt.Printf("Required: %s — %s\n", input.Name, input.Question)
        }
        return
    }
    return err
}
fmt.Printf("Created: %s\n", result.Resource)
```

---

## Constructor & options

| Symbol | Description |
|--------|-------------|
| `New(baseURL string, opts ...Option) *Client` | Root client; one instance per service |
| `WithToken(token string) Option` | `Authorization: Bearer {token}` header |
| `WithTenantID(id string) Option` | `X-Tenant-ID: {id}` header |
| `WithCorrelationID(id string) Option` | Overrides per-request auto UUID v4 |
| `WithHTTPClient(c *http.Client) Option` | Custom transport / TLS pinning |

Headers injected on every request:
- `Authorization: Bearer {token}` (when `WithToken` is set)
- `X-Tenant-ID: {tenantID}` (when `WithTenantID` is set)
- `X-Correlation-ID: {UUID v4}` (auto-generated via `crypto/rand`; override with `WithCorrelationID`)
- `Content-Type: application/json` (on requests with a body)

---

## Resource accessor chain

```
client.Campaigns()                              → CampaignsResource
  .ByID(id string)                              → CampaignResource
  .Optimizations()                              → OptimizationsResource
    .Create(ctx, body) (*AgenticResult, error)  // POST /campaigns/{id}/optimizations
    .Get(ctx, id) (*AgenticResult, error)       // GET  /campaigns/{id}/optimizations/{id}
  .Assets()                                     → AssetsResource
    .Generate(ctx, body) (*AgenticResult, error)// POST /campaigns/{id}/assets:generate
```

---

## Response types

All methods return `(*AgenticResult, error)`. The response follows the
[Agentic REST Response Profile](../spec/agentic-rest-profile.md) §4 10-row table:

| `result.Type` | HTTP | Error type |
|---------------|-----:|------------|
| `"success"` | 200 | nil |
| `"created"` | 201 | nil |
| `"accepted"` | 202 | nil — poll `result.OperationID` |
| `"clarification_required"` | 400 | `*ClarificationRequired` |
| `"validation_failed"` | 422 | `*ValidationFailed` |
| `"approval_required"` | 412 | `*ApprovalRequired` |
| `"forbidden"` | 403 | `*AgenticForbidden` |
| `"conflict"` | 409 | `*Conflict` |
| `"failed_dependency"` | 424 | `*FailedDependency` |
| `"error"` | 500 | `*AgenticError` |

Use `errors.As` for typed dispatch; use `result.Type` for switch-style dispatch.
`result.Err()` returns nil for success/created/accepted, the typed error otherwise.

---

## License

Source-available · Community + Commercial editions.
See [`LICENSE`](../LICENSE) and [`CONTRIBUTING.md`](../CONTRIBUTING.md).
