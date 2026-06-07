# Inventory Agent

**Part of the `agent-graph-ecom` multi-service example** ([`../`](..)). Serves
`GET /inventory/{productId}/stock` — called by `recommendation-agent` to filter
out-of-stock products before returning recommendations to the end user.

## Purpose

| Field | Value |
|---|---|
| **Name** | inventory-agent |
| **Purpose** | Stock-availability lookup for inter-agent calls |
| **Port** (direct) | 8127 |
| **Port** (via gateway-B) | 8126 |
| **Main endpoint** | `GET /inventory/{productId}/stock` |
| **Required headers** | `Authorization: Bearer <service-account-token>` · `X-Tenant-ID` · `X-Correlation-ID` |
| **Success status** | 200 `application/json` |

## Demo flows

| Trigger | Result |
|---|---|
| `product_id` ∈ `{p-99, p-out}` | `in_stock: false` |
| `INVENTORY_UNAVAILABLE=true` env | `424 application/vnd.yaagents.error+json` (`failed_dependency`) |
| Same `Idempotency-Key` on re-call | Cache hit log; no new stock check |

## Standalone run

```bash
# from yaagents/ root
pip install sdk-fastapi/
PORT=8127 python examples/agent-graph-ecom/inventory-agent/main.py
curl http://localhost:8127/inventory/p-1/stock   # 200 in_stock=true
curl http://localhost:8127/inventory/p-99/stock  # 200 in_stock=false
```

See the [top-level README](../README.md) for the full docker-compose quickstart.
