---
description: Extend the yaagents store mock recommender into a real LLM-driven agent
model: claude-sonnet-4-5
tools:
  - Read
  - Edit
  - Bash
---

# Claude Code Agent — Extend the yaagents store example

## What this agent does

You are helping a developer extend the `examples/store/` yaagents reference example.
The current implementation uses a mock recommender that returns same-category products.
Your task is to replace the mock with real LLM-driven recommendations while keeping
the YAAgents Profile v0.3 contract intact.

## Profile contract — what you MUST preserve

- **Response shape**: return `AgenticResponse.success(body=..., ...)` from sdk-fastapi
  (200 OK). The `body` dict can contain anything; keep `seed_product_id`, `recommendations`,
  and `reasoning` keys for compatibility with existing clients.
- **Header `X-YAAgents-Profile: v0.3`**: set automatically by sdk-fastapi — do not touch.
- **Error responses**: use `raise HTTPException(status_code=..., detail=...)` for 4xx errors
  OR use `AgenticResponse.error()` for 500-class errors.
- **Customer context**: honour the `X-Customer-Id` header and `exclude_purchased` body param.

## Repo orientation

- Service code: `examples/store/src/store/`
- Mock to replace: `examples/store/src/store/recommender.py` → `mock_recommend()`
- App entrypoint: `examples/store/src/store/app.py`
- Product data: `examples/store/data/products.json`
- Profile spec: `spec/agentic-rest-profile.md`
- SDK helpers: `sdk-fastapi/src/yaagents_fastapi/response.py`

## Starter prompt 1 — Wire in Anthropic Claude

Read `examples/store/src/store/recommender.py`.
Replace `mock_recommend()` with a real call to Anthropic Claude using the Anthropic
Python SDK (`anthropic` package). Provide the seed product name, category, and the
customer's purchased product names as context. Ask the LLM to return 3 product
recommendations from the static catalog with personalised reasoning.

Requirements:
- Install `anthropic` and add it to `pyproject.toml` dependencies.
- Read `ANTHROPIC_API_KEY` from the environment; raise `HTTPException(503)` if absent.
- Keep the response shape: `{"seed_product_id": ..., "recommendations": [...], "reasoning": "..."}`.
- The `recommendations` list must contain valid product dicts from `products.json`
  (do not invent new products).
- Fallback to `mock_recommend()` when `ANTHROPIC_API_KEY` is unset (graceful degradation).

## Starter prompt 2 — Add tool use

Extend the recommender with two tools:
- `check_inventory(product_id: str) -> bool` — returns True if in stock (mock: always True)
- `get_similar_brand_products(brand: str) -> list[dict]` — returns products from the same brand

Ask the LLM to call these tools as needed before returning recommendations.
Filter out out-of-stock products from the final list.

Wire the tools using the Anthropic SDK `tools` parameter. Keep the existing
`AgenticResponse.success()` response shape.

## Starter prompt 3 — Add a new endpoint

Add `POST /carts/{id}:checkout-assist` to `examples/store/src/store/app.py`.
The request body should accept `{"cart_items": [...product_ids...]}`.
The LLM should return:
- `upsell_suggestions`: 2–3 complementary products not in the cart
- `abandonment_copy`: a one-sentence message to encourage checkout completion

Follow the same Profile pattern: `AgenticResponse.success(body={...})`.
Update `examples/store/routes.yaml` and `examples/store/docker-compose.yml`
to expose the new route through the gateway.

## What success looks like

```bash
curl -sX POST http://localhost:8120/products/p-1/recommendations \
  -H 'Content-Type: application/json' \
  -H 'X-Customer-Id: c-1' \
  -d '{"limit": 3}' | jq
# Returns AgenticResponse with 3 LLM-driven recommendations + personalised reasoning
# X-YAAgents-Profile: v0.3 header present in response
```
