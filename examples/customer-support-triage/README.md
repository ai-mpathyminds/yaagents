# Customer Support Triage — YAAgents Reference Example

FastAPI reference implementation of the [YAAgents Agentic REST Profile v0.3](../../spec/)
demonstrating the **`clarification_required`** agentic flow.

A ticket-triage service that asks for more information when `severity` or
`category` are missing from the request body — illustrating how agents request
structured clarification instead of returning a generic error.

---

## Example table

| Example | Purpose | Port | Main endpoint | Required headers | Success status |
|---|---|---|---|---|---|
| customer-support-triage | Triage support tickets; clarification flow | 8122 | POST /tickets/{ticketId}:triage | X-Tenant-ID + Authorization | 201 |

---

## Quick start (Docker Compose demo)

> **SECURITY NOTE — demo token only.**
> `GATEWAY_JWT_SECRET=demo-secret-not-for-production` is hard-coded for local
> demos. **Never use this value in production.** Set `GATEWAY_JWT_JWKS_URL`
> instead (ADR PI1-yaa-0001 §3).

```bash
cd examples/customer-support-triage
docker compose up --build
```

Both services start healthy (gateway waits for the triage service to pass `/healthz`).

### Pre-generated demo token (HS256 / `demo-secret-not-for-production`)

```bash
DEMO_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZXYiLCJyb2xlcyI6WyJjYW1wYWlnbjpvcHRpbWl6ZSJdLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcxNjAwMDAwMH0.7sivhIA1seepaiHvyOnjhIMKiiExMfnMK7mv99NpDBg
```

This token is signed with the demo secret, expires year 2286, and is safe to
commit — it only works against a gateway running the demo secret.

---

### Clarification flow (400): missing severity and category

```bash
curl -s -X POST http://localhost:8122/tickets/t-123:triage \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{}' \
  | python3 -m json.tool
# -> HTTP 400, Content-Type: application/vnd.yaagents.clarification+json
# requiredInputs lists severity + category with allowedValues
```

### Created flow (201): high-severity ticket

```bash
curl -s -X POST http://localhost:8122/tickets/t-123:triage \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"severity":"high","category":"technical","description":"Service down"}' \
  | python3 -m json.tool
# -> HTTP 201, Content-Type: application/json
# ticket.status = "escalated", recommendedOwner = "support-team-level-2@example.com"
```

### Created flow (201): low-severity ticket

```bash
curl -s -X POST http://localhost:8122/tickets/t-456:triage \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"severity":"low","category":"billing","description":"Invoice question"}' \
  | python3 -m json.tool
# -> HTTP 201, Content-Type: application/json
# ticket.status = "auto-resolved", autoresolveHint included
```

### Via gateway (port 8120)

```bash
curl -s -X POST http://localhost:8120/tickets/t-123:triage \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"severity":"high","category":"technical"}' \
  | python3 -m json.tool
```

---

## Endpoint

| Method | Path | Agentic flows |
|--------|------|---------------|
| POST | `/tickets/{ticketId}:triage` | clarification_required (400), created (201) |

Service port: `http://localhost:8122` (direct)
Gateway port: `http://localhost:8120` (proxied)

---

## Local dev (no Docker)

```bash
pip install -e ".[test]" ../../sdk-fastapi
uvicorn main:app --port 8122
```

OpenAPI docs: `http://localhost:8122/docs`

---

## Running tests

```bash
pip install -e ".[test]" ../../sdk-fastapi
pytest
```
