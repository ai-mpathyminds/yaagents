# Campaign API -- YAAgents Reference Example

> **Looking for the recommended starter?** See
> [`examples/store/`](../store/) — the ecommerce product-recommendations
> example used in the Quick Start. `campaign-api/` is preserved as an
> alternative example showing the same Profile shape applied to a
> different domain (martech / campaign optimization with RBAC, tenant
> context, and audit logging).

FastAPI reference implementation of the [YAAgents Agentic REST Profile v0.1](../../spec/).

Demonstrates all four PRD §6.2 flows (`created`, `clarification_required`,
`validation_failed`, `failed_dependency`) through the yaagents gateway with
HS256 auth, route-level RBAC, tenant context, and audit logging.

See `yaagents/docs/PI1-yaa/examples-campaign-api.md` for WI details.

---

## Quick start (Docker Compose demo)

> **SECURITY NOTE — demo token only.**
> `GATEWAY_JWT_SECRET=demo-secret-not-for-production` is hard-coded for local
> demos. **Never use this value in production.** Set `GATEWAY_JWT_JWKS_URL`
> instead (ADR PI1-yaa-0001 §3). No `.env` file with a real secret should
> ever be committed.

```bash
cd examples/campaign-api
docker compose up --build
```

Both services start healthy (gateway waits for campaign-api to pass `/healthz`).

### Pre-generated demo tokens (HS256 / `demo-secret-not-for-production`)

```bash
# Full access (includes campaign:optimize role)
DEMO_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZXYiLCJyb2xlcyI6WyJjYW1wYWlnbjpvcHRpbWl6ZSJdLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcxNjAwMDAwMH0.7sivhIA1seepaiHvyOnjhIMKiiExMfnMK7mv99NpDBg

# No roles (triggers RBAC 403 on optimizations routes)
NO_ROLE_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZXYiLCJyb2xlcyI6W10sImV4cCI6OTk5OTk5OTk5OSwiaWF0IjoxNzE2MDAwMDAwfQ.53NBy9kHHTwJKkFCb8xpxc4wAyatdnZVmI7lUOpP-yM
```

These tokens are signed with the demo secret, expire year 2286, and are
safe to commit -- they only work against a gateway running the demo secret.

### §6.2 flow 1: created (POST /campaigns with all fields)

```bash
curl -s -X POST http://localhost:8120/campaigns \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"name":"Summer Sale","budget":5000,"targetAudience":"18-35","successMetric":"ctr"}' \
  | python3 -m json.tool
# -> HTTP 201, Content-Type: application/vnd.yaagents.created+json
# Save the campaignId from the response for subsequent calls
```

### §6.2 flow 2: clarification_required (omit successMetric)

```bash
curl -s -X POST http://localhost:8120/campaigns \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"name":"Summer Sale","budget":5000,"targetAudience":"18-35"}' \
  | python3 -m json.tool
# -> HTTP 400, Content-Type: application/vnd.yaagents.clarification_required+json
# requiredInputs lists successMetric with allowedValues
```

### §6.4 quick-start (PRD) -- clarification_required via gateway

```bash
# This is the canonical §6.4 curl from the PRD
curl -s -X POST http://localhost:8120/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"objectives":[],"maxSuggestions":3}' \
  | python3 -m json.tool
# Campaign cmp-123 does not exist -> 404; create one first (flow 1 above)
```

### §6.2 flow -- gateway RBAC 403 (token lacks campaign:optimize)

```bash
# First create a campaign (get the campaignId from the response):
CAMPAIGN_ID=$(curl -s -X POST http://localhost:8120/campaigns \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"name":"RBAC Demo","budget":1000,"targetAudience":"all","successMetric":"roas"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['campaign']['id'])")

# Now call optimizations with the NO_ROLE_TOKEN -- gateway returns 403
curl -s -X POST http://localhost:8120/campaigns/${CAMPAIGN_ID}/optimizations \
  -H "Authorization: Bearer $NO_ROLE_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"objectives":["reduce_cpl"],"maxSuggestions":2}' \
  | python3 -m json.tool
# -> HTTP 403, Content-Type: application/vnd.yaagents.error+json
```

### §6.2 flow 4: failed_dependency (LLM down toggle)

```bash
# Toggle LLM-down (demo-only endpoint, excluded from OpenAPI spec)
curl -s -X POST "http://localhost:8121/campaigns/x" 2>/dev/null || true
# Use the internal port 8121 directly for the toggle (gateway doesn't route /_demo)
# OR toggle via docker exec:
docker exec $(docker compose ps -q campaign-api) \
  python3 -c "import requests; requests.post('http://localhost:8121/_demo/llm-down?enabled=true')"

# Now call optimizations through gateway -- returns failed_dependency
curl -s -X POST http://localhost:8120/campaigns/${CAMPAIGN_ID}/optimizations \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"objectives":["reduce_cpl"],"maxSuggestions":2}' \
  | python3 -m json.tool
# -> HTTP 424, Content-Type: application/vnd.yaagents.failed_dependency+json
```

---

## Endpoints

| Method | Path | §6.2 flows |
|--------|------|------------|
| POST | `/campaigns` | created, clarification_required, validation_failed |
| GET | `/campaigns/{id}` | success (200) |
| POST | `/campaigns/{id}/optimizations` | created, failed_dependency; RBAC-gated |
| GET | `/campaigns/{id}/optimizations/{opId}` | success (200); RBAC-gated |
| POST | `/campaigns/{id}/assets:generate` | created, failed_dependency |

Gateway entry: `http://localhost:8120`
Campaign-API internal: `http://localhost:8121` (not published to host)

---

## Local dev (no Docker)

```bash
pip install -e ".[test]" ../../../sdk-fastapi
uvicorn campaign_api.app:app --port 8121
```

OpenAPI docs: `http://localhost:8121/docs`

---

## Running tests

```bash
pip install -e ".[test]"
pytest
```
