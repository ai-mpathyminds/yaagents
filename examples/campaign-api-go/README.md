# campaign-api-go

> **Looking for the recommended starter?** See
> [`examples/store-go/`](../store-go/) — the Go ecommerce product-recommendations
> example. `campaign-api-go/` is preserved as an alternative example showing
> the same Profile shape applied to a different domain.

Go reference example for the [YAAgents Agentic REST Profile v0.3](../../spec/agentic-rest-profile.md).

Mirrors `examples/campaign-api/` (Python FastAPI) using the **Go server SDK**
(`sdk-go`) with `net/http` — no router framework. Demonstrates all five §8.1 / §13.2
demo flows for `POST /campaigns/{id}/optimizations`.

**SDK:** [`github.com/ai-mpathyminds/yaagents-sdk-go`](https://github.com/ai-mpathyminds/yaagents-sdk-go)  
**Port:** `8121` (direct) · `8120` (via gateway)  
**Profile:** `v0.3` (`X-YAAgents-Profile: v0.3` on every response)

---

## Quick start (PRD §8.2)

```bash
cd examples/campaign-api-go
docker compose up
```

Once healthy, run the PRD §8.2 verbatim curl:

```bash
# Optimization request — happy path (flow 2: created)
curl -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal": "ctr"}'
# → 201 application/json  X-YAAgents-Profile: v0.3
```

---

## Five demo flows

The handler at `POST /campaigns/{id}/optimizations` demonstrates all five
PRD §13.2 / §8.1 flows using the pure sdk-go sequence:

```
sdkgo.FromRequest(r) → sdkgo.AgenticResponse{} → sdkgo.Write(w, ...)
```

### Flow 1 — Clarification (400)

Missing `goal` field → `400 application/vnd.yaagents.clarification+json`

```bash
curl -s -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{}'
# → 400 application/vnd.yaagents.clarification+json
# Body: {"type":"clarification_required","requiredInputs":[{"name":"goal",...}],...}
```

### Flow 2 — Created (201)

Valid body with `goal` field → `201 application/json`

```bash
curl -s -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal": "ctr"}'
# → 201 application/json  X-YAAgents-Profile: v0.3
```

### Flow 3 — Accepted / async (202)

`Prefer: respond-async` header → `202 application/vnd.yaagents.operation+json`

```bash
curl -s -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -H "Prefer: respond-async" \
  -d '{"goal": "ctr"}'
# → 202 application/vnd.yaagents.operation+json
# Body: {"type":"operation_accepted","operationId":"op-cmp-123-ctr",...}
```

### Flow 4 — Validation failed (422)

Invalid field type (number instead of string) → `422 application/vnd.yaagents.validation-error+json`

```bash
curl -s -X POST http://localhost:8121/campaigns/cmp-123/optimizations \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal": 42}'
# → 422 application/vnd.yaagents.validation-error+json
# Body: {"type":"validation_failed","errors":[{"field":"goal","message":"must be a string..."}],...}
```

### Flow 5 — Auth failure (401 from gateway)

No `Authorization` header through the gateway → `401`

```bash
curl -s -X POST http://localhost:8120/campaigns/cmp-123/optimizations \
  -H "Content-Type: application/json" \
  -d '{"goal": "ctr"}'
# → 401 (token-validator plugin; no upstream involvement)
```

---

## Running the CI test

```bash
# Requires services already running via docker compose up
cd examples/campaign-api-go
bash test_e2e.sh
# Exits 0 on all-pass; non-zero on any flow regression.
```

---

## Building locally (without Docker)

```bash
cd examples/campaign-api-go
go build ./...
# Binary: campaign-api-go (or campaign-api-go.exe on Windows)
PORT=8121 ./campaign-api-go
```

Requires Go 1.22+. The `sdk-go` dependency is resolved via the `replace`
directive in `go.mod` pointing to `../../sdk-go`.

---

## Compose topology

```
host:8120 ──► yaagents-gateway:8080 ──► campaign-api-go:8121 (Docker network)
host:8121 ─────────────────────────────► campaign-api-go:8121 (direct, for demo)
```

The gateway (`yaagents-gateway`) validates JWTs, enforces RBAC
(`roles: [campaign:optimize]`), and injects tenant context.
Direct access to port 8121 is used by the quick-start curl and by
`test_e2e.sh` flows 1–4; flow 5 uses the gateway path to demonstrate 401.

---

## License

Apache 2.0 — see [../../LICENSE](../../LICENSE).
