# Extend the yaagents store example — Claude.ai (web/desktop)

**How to use**: Open [claude.ai](https://claude.ai), start a new conversation, paste
the System Prompt below, then use one of the Starter Prompts.

---

## System Prompt

Paste this as the system prompt or first message in your Claude conversation:

```
You are helping a developer extend the `examples/store/` yaagents reference example
(https://github.com/ai-mpathyminds/yaagents).

The store service exposes POST /products/{id}/recommendations and currently returns
mock same-category recommendations. Your task is to help replace the mock with real
LLM-driven recommendations.

Context files you should ask the developer to share or read:
- examples/store/src/store/recommender.py   ← the mock to replace
- examples/store/src/store/app.py           ← FastAPI app
- examples/store/data/products.json         ← 20 static products

YAAgents Profile v0.3 contract you MUST preserve:
- The endpoint returns AgenticResponse.success(body=...) from yaagents-fastapi SDK
- Body shape: {"seed_product_id": "...", "recommendations": [...], "reasoning": "..."}
- Recommendations must be real product dicts from products.json (no invented products)
- Keep X-Customer-Id header support and exclude_purchased behaviour
- Keep /healthz and /readyz endpoints unchanged

SDK reference:
  AgenticResponse.success(body: dict, correlation_id: str = "", request_id: str = "")
  → returns 200 application/json FastAPI Response

For errors: raise HTTPException(status_code=..., detail="...") is fine for 4xx.
```

---

## Starter Prompt 1 — Wire in a real LLM

```
Here is my current mock recommender (paste contents of recommender.py).

Replace mock_recommend() with a real call to Anthropic Claude using the anthropic
Python package. The LLM receives:
- The seed product name, category, and price
- The customer's purchased products (names + categories)

Ask the LLM to pick 3 products from the attached products.json list (share the file)
that the customer would most likely want, with a one-sentence reasoning.

Requirements:
1. Read ANTHROPIC_API_KEY from os.environ; if absent, fall back to mock_recommend()
2. Parse the LLM response to return valid product dicts (match product IDs from the list)
3. Keep the response shape exactly as-is

Show me the updated recommender.py and any changes needed to pyproject.toml.
```

---

## Starter Prompt 2 — Structured LLM output

```
Improve the LLM recommender from the previous step to use structured output.
Instead of asking the LLM to return free text and parsing it, use Anthropic's
tool_use feature to force the model to return a structured list of product IDs.

Define a tool called select_products with a schema:
  {"product_ids": ["p-1", "p-2", "p-3"], "reasoning": "..."}

The model should be forced to call this tool. The function should validate that
all returned IDs exist in products.json before returning.

Show me the updated recommender.py.
```

---

## Starter Prompt 3 — Add a new endpoint

```
Add POST /carts/{id}:checkout-assist to app.py.

Request body: {"cart_items": ["p-1", "p-5"]}

The LLM should return:
{
  "upsell_suggestions": [/* 2-3 product dicts */],
  "abandonment_copy": "One sentence encouraging checkout"
}

Use AgenticResponse.success(body=...) for the response.
Update routes.yaml to add the new route through the gateway.
Show me the changes to app.py and routes.yaml.
```
