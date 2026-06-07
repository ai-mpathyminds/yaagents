# Recommendation Agent

**Part of the `agent-graph-ecom` multi-service example** ([`../`](..)). Serves
`POST /products/{productId}/recommendations` — filters candidate recommendations
by calling `inventory-agent` before responding.

## Purpose

| Field | Value |
|---|---|
| **Name** | recommendation-agent |
| **Purpose** | Product recommendations filtered by live inventory availability |
| **Port** (direct) | 8125 |
| **Port** (via gateway-A) | 8124 |
| **Main endpoint** | `POST /products/{productId}/recommendations` |
| **Required headers** | `Authorization: Bearer <end-user-token>` · `X-Tenant-ID` · `Content-Type: application/json` |
| **Success status** | 200 `application/json` |

## Demo flows

| Trigger | Result |
|---|---|
| Valid request for `p-1` | 200 with candidates minus out-of-stock (`p-99`, `p-out`) |
| Stop inventory-agent container | 424 `failed_dependency` propagated |
| `p-out` as sole candidate | 200 with empty recommendations list |
| Same `Idempotency-Key` twice | Second call uses cached inventory results |

## Inter-agent auth

recommendation-agent mints a short-lived service-account JWT (HS256, signed
with `RECOMMENDATION_SVC_SECRET`) at startup. When calling inventory-agent via
gateway-B, it presents **this service-account token** — not the end-user token.
The `actor_id` logged by inventory-agent will be `recommendation-agent-svc`,
distinguishable from end-user subjects.

See the [top-level README](../README.md) for the full docker-compose quickstart.
