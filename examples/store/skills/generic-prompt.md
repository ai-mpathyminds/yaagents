# Extend the yaagents store example — Generic AI-tool prompt

Use this prompt with any AI coding assistant: Gemini, Mistral, CodeWhisperer,
Tabnine, JetBrains AI, or any other tool that accepts natural language instructions.

---

## Context to provide your AI tool

Before using the prompts, share or describe these files to your AI tool:

1. **`src/store/recommender.py`** — the mock `mock_recommend()` function you want to replace
2. **`src/store/app.py`** — the FastAPI app showing how `mock_recommend()` is called
3. **`data/products.json`** — the 20 static products the recommendations must come from

---

## What your AI tool needs to know (paste this as context)

```
I have a yaagents store example service (https://github.com/ai-mpathyminds/yaagents).
It exposes POST /products/{id}/recommendations and returns mock product recommendations.

I want to replace the mock with a real LLM call.

Key constraints:
1. The function signature to replace:
   mock_recommend(product_id: str, customer_id: str | None, limit: int, exclude_purchased: bool)
   → returns (list[dict], str)  # (recommendations, reasoning)

2. Return type must stay the same. Recommendations must be product dicts from products.json
   (exact structure: {"id": "p-1", "name": "...", "category": "...", "price": ..., "brand": "..."})

3. The app.py wraps the result in:
   AgenticResponse.success(body={"seed_product_id": ..., "recommendations": ..., "reasoning": ...})
   Do NOT change app.py's response wrapping.

4. Customer context: if customer_id is set and exclude_purchased is True, filter out products
   that appear in the customer's "purchased" list in customers.json.
```

---

## Starter Prompt 1 — Replace the mock

```
Replace the mock_recommend() function in src/store/recommender.py with a real LLM call.

Use [your LLM API of choice]. Provide the seed product details and the customer's
purchase history as context. Ask the LLM to select up to {limit} products from the
products.json catalog with personalised reasoning.

Requirements:
- Read the API key from an environment variable (e.g. LLM_API_KEY)
- If the key is absent, fall back to the original mock behavior
- Return type must stay: tuple[list[dict], str]
- Product dicts must be exact entries from products.json
- Add the LLM SDK to pyproject.toml dependencies

Show me the complete updated recommender.py and pyproject.toml.
```

---

## Starter Prompt 2 — Add tool use / function calling

```
Extend the LLM recommender to use function calling / tool use.

Define two tools for the LLM:
1. filter_by_category(category: str) → list of product IDs in that category
2. filter_purchased(customer_id: str) → list of already-purchased product IDs

Wire them so the LLM can call them before returning its final recommendation list.
Show the updated recommender.py.
```

---

## Starter Prompt 3 — New endpoint

```
Add POST /carts/{id}:checkout-assist to src/store/app.py.

Request body: {"cart_items": ["p-1", "p-5"]}

Use your LLM to return:
- upsell_suggestions: 2-3 complementary products not in the cart
- abandonment_copy: one sentence encouraging checkout

Wrap the result with:
  return AgenticResponse.success(body={
      "upsell_suggestions": [...],
      "abandonment_copy": "..."
  })

Also update routes.yaml to add the route through the gateway.
Show app.py changes and updated routes.yaml.
```

---

## What success looks like

```bash
# Start the compose stack
docker compose up --build

# Test the LLM-powered endpoint
curl -sX POST http://localhost:8120/products/p-1/recommendations \
  -H 'Content-Type: application/json' \
  -H 'X-Customer-Id: c-1' \
  -d '{"limit": 3}' | python3 -m json.tool

# Expected: 200 OK with personalised LLM reasoning instead of
# "Same category as 'Wireless headphones' (electronics)"
```
