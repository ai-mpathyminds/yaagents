# store — YAAgents ecommerce reference example

A universally-relatable reference implementation of the
[YAAgents Agentic REST Profile v0.3](https://ai-mpathyminds.github.io/yaagents/profile/).

The `store` service exposes one agentic endpoint:

```
POST /products/{id}/recommendations
```

It returns mock product recommendations from the same category as the seed product.
The mock is intentionally simple so you can replace it with a real LLM call using the
starter prompts in `skills/`.

---

## Quick start (Docker Compose)

> **SECURITY NOTE — demo token only.**
> `GATEWAY_JWT_SECRET=demo-secret-not-for-production` is hard-coded for local demos.
> **Never use this value in production.**

```bash
cd examples/store
docker compose up --build
```

Both services start healthy. The gateway listens on `localhost:8120`; the store service
listens on `localhost:8121` (also accessible directly for local testing).

### Get recommendations for product p-1 (wireless headphones)

```bash
curl -s -X POST http://localhost:8120/products/p-1/recommendations \
  -H "Content-Type: application/json" \
  -d '{"limit": 3, "exclude_purchased": true}' \
  | python3 -m json.tool
```

**Expected response — 200 `application/json`:**

```json
{
  "seed_product_id": "p-1",
  "recommendations": [
    {"id": "p-2", "name": "Bluetooth speaker", "category": "electronics", "price": 49.99, "brand": "AudioCore"},
    {"id": "p-3", "name": "Smart watch", "category": "electronics", "price": 199.99, "brand": "TickTock"},
    {"id": "p-5", "name": "Mechanical keyboard", "category": "electronics", "price": 129.99, "brand": "KeyMaster"}
  ],
  "reasoning": "Same category as 'Wireless headphones' (electronics)"
}
```

### Personalise with customer purchase history

Pass `X-Customer-Id: c-1` to filter out products already purchased by Alice
(who has bought `p-1` and `p-4`):

```bash
curl -s -X POST http://localhost:8120/products/p-1/recommendations \
  -H "Content-Type: application/json" \
  -H "X-Customer-Id: c-1" \
  -d '{"limit": 3}' \
  | python3 -m json.tool
```

Alice already owns `p-1` (the seed) and `p-4`. The response filters `p-4` and
returns the next three electronics products.

### Product not found

```bash
curl -s -X POST http://localhost:8120/products/p-99/recommendations \
  -H "Content-Type: application/json" \
  -d '{}' \
  | python3 -m json.tool
# -> 404 Not Found
```

---

## Make it real — use an AI-tool skill

The mock recommender is a starting point. Open `skills/<your-ai-tool>.md` to get a
starter prompt that instructs your AI coding tool to replace the mock with real
LLM-driven recommendations.

| AI tool | Skill file |
|---------|-----------|
| Claude Code | `skills/claude-code-agent.md` |
| Claude.ai (web/desktop) | `skills/claude-web.md` |
| GitHub Copilot | `skills/github-copilot.md` |
| OpenAI / ChatGPT | `skills/openai-codex.md` |
| Cursor | `skills/cursor.md` |
| Any other tool | `skills/generic-prompt.md` |

In ~10–20 minutes, the AI tool will extend the mock into a real LLM-driven recommender
while keeping the Profile contract intact.

---

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/products/{id}/recommendations` | Get recommendations for a product |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe (reports Profile version) |

Gateway entry: `http://localhost:8120`
Store service internal: `http://localhost:8121`

OpenAPI docs (store direct): `http://localhost:8121/docs`

---

## Local dev (no Docker)

```bash
# From the examples/store/ directory:
pip install -e "."
uvicorn store.app:app --reload --port 8080
```

Then hit `http://localhost:8080/products/p-1/recommendations` directly.

---

## Sample data

| File | Contents |
|------|----------|
| `data/products.json` | 20 products across 5 categories |
| `data/customers.json` | 3 customers with purchase histories |

---

## See also

- [Quick Start tutorial](https://ai-mpathyminds.github.io/yaagents/tutorials/quick-start/)
  — full walkthrough using this example
- [Profile Spec](https://ai-mpathyminds.github.io/yaagents/profile/) — normative
  response shape + status codes
- [`examples/store-go/`](../store-go/) — Go mirror using sdk-go
- [`examples/campaign-api/`](../campaign-api/) — alternative example (same Profile,
  different domain)
