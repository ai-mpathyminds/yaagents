# agent-graph-ecom — Multi-Service Inter-Agent Example

> **YAAgents Agentic REST Profile v0.3 · WI-4yaa.AF-1**

Two yaagents services that call each other, demonstrating the four
inter-agent patterns agentforge needs to convert 20 internal agents to
Profile v0.3-conformant HTTP services.

```
client ──► gateway-A:8124 ──► recommendation-agent:8125
                                       │
                              service-account JWT
                                       │
                                       ▼
                              gateway-B:8126 ──► inventory-agent:8127
```

## What this example demonstrates

| Concept | Where it shows up |
|---|---|
| **Inter-agent auth** | recommendation-agent mints a HS256 service-account token (distinct from the end-user token) and presents it to gateway-B. inventory-agent logs `actor=recommendation-agent-svc`. |
| **Tenant context propagation** | `X-Tenant-ID` from the end-user request is forwarded by recommendation-agent on every inventory call. inventory-agent logs the same `tenant=tenant-001`. |
| **Profile v0.3 outcome propagation** | If inventory-agent returns `424 failed_dependency`, recommendation-agent propagates `424` (not `500`). If inventory-agent returns `400 clarification_required`, recommendation-agent propagates `400`. |
| **Idempotency-Key forwarding** | recommendation-agent forwards `Idempotency-Key: {outer}::{product}` on each inventory call. Identical keys return cached results — only one "stock check performed" log line per product per key. |

## Per-example table (parity-review §8)

| Field | Value |
|---|---|
| **Name** | agent-graph-ecom |
| **Purpose** | 2-agent inter-agent call: recommendations filtered by inventory |
| **gateway-A port** | 8124 (end-user) |
| **recommendation-agent port** | 8125 (direct, internal only) |
| **gateway-B port** | 8126 (service-to-service) |
| **inventory-agent port** | 8127 (direct, internal only) |
| **Main endpoint** | `POST /products/{productId}/recommendations` via gateway-A |
| **Required headers** | `Authorization: Bearer <end-user-token>` · `X-Tenant-ID` · `Content-Type: application/json` |
| **Success status** | `200 application/json` |

## Quick start

### 1. Start all four containers

```bash
cd examples/agent-graph-ecom
docker compose up --build -d
docker compose ps          # all four should be healthy within ~30 s
```

### 2. Generate a demo end-user token

```bash
python3 - <<'EOF'
import jwt, time
token = jwt.encode(
    {"sub": "user@example.com", "iss": "demo-user-idp",
     "iat": int(time.time()), "exp": int(time.time()) + 3600},
    "demo-user-secret", algorithm="HS256"
)
print(token)
EOF
```

Store the output as `$USER_TOKEN`.

### 3. Happy path — in-stock recommendations

```bash
curl -s -X POST http://localhost:8124/products/p-1/recommendations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"limit": 3}' | python3 -m json.tool
```

Expected: `200` with `recommendations` containing `p-2` and `p-3` (candidates
`p-99` and `p-out` are filtered because they are out of stock).

### 4. Verify inter-agent auth (service-account token)

```bash
docker compose logs inventory-agent | grep "stock check"
# Look for: actor=recommendation-agent-svc  (NOT user@example.com)
```

### 5. Verify tenant propagation

```bash
docker compose logs inventory-agent | grep "tenant="
# Look for: tenant=tenant-001  (same value the end-user sent)
```

### 6. Verify Idempotency-Key deduplication

```bash
# First call — triggers stock checks; logs show "stock check performed"
curl -s -X POST http://localhost:8124/products/p-1/recommendations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Idempotency-Key: demo-idem-001" \
  -H "Content-Type: application/json" \
  -d '{"limit": 3}' | python3 -m json.tool

# Second call with SAME key — cached; logs show "idempotency cache hit"
curl -s -X POST http://localhost:8124/products/p-1/recommendations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Idempotency-Key: demo-idem-001" \
  -H "Content-Type: application/json" \
  -d '{"limit": 3}' | python3 -m json.tool

docker compose logs inventory-agent | grep -E "stock check|cache hit"
# "stock check performed" appears only once per product per key
```

### 7. Failed-dependency propagation

```bash
# Stop inventory-agent to simulate unavailability
docker compose stop inventory-agent

curl -s -X POST http://localhost:8124/products/p-1/recommendations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"limit": 3}'
# Expected: 424 application/vnd.yaagents.error+json

# Restart
docker compose start inventory-agent
```

### 8. Out-of-stock filtering

```bash
# p-out is always out-of-stock in inventory-agent
curl -s -X POST http://localhost:8124/products/p-2/recommendations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"limit": 5}' | python3 -m json.tool
```

## File layout

```
agent-graph-ecom/
├── docker-compose.yml           # 4-service compose (gateways + agents)
├── README.md                    # this file
├── gateway-A-routes.yaml        # routes POST /products/{id}/recommendations
├── gateway-A-plugins.yaml       # token-validator end-user IdP
├── gateway-B-routes.yaml        # routes GET /inventory/{id}/stock
├── gateway-B-plugins.yaml       # token-validator service-account IdP
├── recommendation-agent/
│   ├── main.py                  # FastAPI app; calls inventory-agent
│   ├── pyproject.toml
│   ├── Dockerfile
│   └── README.md
└── inventory-agent/
    ├── main.py                  # FastAPI app; stock check + idempotency cache
    ├── pyproject.toml
    ├── Dockerfile
    └── README.md
```

## Architecture notes

- **Framework-neutral design** — the agent-A → agent-B pattern is plain HTTP
  with Profile v0.3 typed outcomes; no proprietary agent-framework vocabulary.
- Gateway-A and gateway-B use **different** `jwt_secret` values (demo). In
  production configure separate JWKS issuers in token-validator's `issuers:`
  list (ADR PI4-yaa-0002 multi-IDP plug-point).
- Idempotency cache in inventory-agent is **in-memory / process-local** (demo
  only). Production implementations use a shared cache (Redis, DynamoDB, etc.).
- See `architecture/inter-agent-calls.mdx` in the docs for full pattern
  documentation including RFC 8693 token-exchange and retry guidance.
