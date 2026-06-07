# Financial Risk Screening — YAAgents Reference Example

FastAPI reference implementation of the [YAAgents Agentic REST Profile v0.3](../../spec/)
demonstrating the **`approval_required`** agentic flow.

A claims risk-screening service that halts processing and requests human
approval when a claim's risk score exceeds a configurable threshold — showing
how agents surface high-stakes decisions rather than acting silently. Also
demonstrates a **license-tier gate**: community-tier callers receive 403 before
any screening logic runs.

---

## Example table

| Example | Purpose | Port | Main endpoint | Required headers | Success status |
|---|---|---|---|---|---|
| financial-risk-screening | Screen claims; approval flow | 8123 | POST /claims/{claimId}/risk-screens | X-Tenant-ID + Authorization + X-License-Tier | 201 |

---

## Quick start (Docker Compose demo)

> **SECURITY NOTE — demo token only.**
> `GATEWAY_JWT_SECRET=demo-secret-not-for-production` is hard-coded for local
> demos. **Never use this value in production.** Set `GATEWAY_JWT_JWKS_URL`
> instead (ADR PI1-yaa-0001 §3).

```bash
cd examples/financial-risk-screening
docker compose up --build
```

Both services start healthy (gateway waits for the screening service to pass `/healthz`).

### Pre-generated demo token (HS256 / `demo-secret-not-for-production`)

```bash
DEMO_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZXYiLCJyb2xlcyI6WyJjYW1wYWlnbjpvcHRpbWl6ZSJdLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcxNjAwMDAwMH0.7sivhIA1seepaiHvyOnjhIMKiiExMfnMK7mv99NpDBg
```

---

### License-check gate (403): community tier

```bash
curl -s -X POST http://localhost:8123/claims/clm-001/risk-screens \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "X-License-Tier: community" \
  -H "Content-Type: application/json" \
  -d '{"amount":500,"claimant_history":"good"}' \
  | python3 -m json.tool
# -> HTTP 403, Content-Type: application/vnd.yaagents.error+json
# code: LICENSE_TIER_INSUFFICIENT
```

### Approval-required flow (412): high-risk claim

```bash
curl -s -X POST http://localhost:8123/claims/clm-002/risk-screens \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "X-License-Tier: professional" \
  -H "Content-Type: application/json" \
  -d '{"amount":15000,"claimant_history":"neutral"}' \
  | python3 -m json.tool
# -> HTTP 412, Content-Type: application/vnd.yaagents.approval-required+json
# riskScore > 0.7; approvalToken returned
```

### Created flow (201): low-risk claim

```bash
curl -s -X POST http://localhost:8123/claims/clm-003/risk-screens \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "X-License-Tier: enterprise" \
  -H "Content-Type: application/json" \
  -d '{"amount":500,"claimant_history":"good"}' \
  | python3 -m json.tool
# -> HTTP 201, Content-Type: application/json
# screen.status = "approved", riskScore < 0.7
```

### Via gateway (port 8120)

```bash
curl -s -X POST http://localhost:8120/claims/clm-003/risk-screens \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "X-License-Tier: enterprise" \
  -H "Content-Type: application/json" \
  -d '{"amount":500,"claimant_history":"good"}' \
  | python3 -m json.tool
```

---

## Risk-score model (demo only)

```
score = min(amount / 10_000 * history_multiplier, 1.0)
history_multiplier: good=0.7  neutral=1.0  bad=1.4
threshold: 0.7
```

Examples that exceed the threshold:
- `amount=10001, claimant_history=neutral` (score ≈ 1.000 → approval required)
- `amount=5001,  claimant_history=bad`     (score ≈ 0.700 → right at threshold)

---

## Endpoint

| Method | Path | Agentic flows |
|--------|------|---------------|
| POST | `/claims/{claimId}/risk-screens` | forbidden (403), approval_required (412), created (201) |

Service port: `http://localhost:8123` (direct)
Gateway port: `http://localhost:8120` (proxied)

---

## Local dev (no Docker)

```bash
pip install -e ".[test]" ../../sdk-fastapi
uvicorn main:app --port 8123
```

OpenAPI docs: `http://localhost:8123/docs`

---

## Running tests

```bash
pip install -e ".[test]" ../../sdk-fastapi
pytest
```
