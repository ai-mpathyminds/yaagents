# Extend the yaagents store example — OpenAI / ChatGPT

**How to use**: Open [ChatGPT](https://chatgpt.com) (or any OpenAI API interface).
Paste the System Prompt into a new conversation, then use one of the Starter Prompts.
Works with GPT-4o, o3, and Codex-based tools.

---

## System Prompt

```
You are helping a developer extend the yaagents store example
(https://github.com/ai-mpathyminds/yaagents).

The store service (examples/store/) exposes POST /products/{id}/recommendations.
It currently returns mock same-category recommendations. Your task is to help
replace the mock with real LLM-driven recommendations.

Key files:
- examples/store/src/store/recommender.py  ← mock_recommend() to replace
- examples/store/src/store/app.py          ← FastAPI app using sdk-fastapi
- examples/store/data/products.json        ← 20 static products

Contract you MUST preserve (YAAgents Profile v0.3):
- app.py returns: AgenticResponse.success(body={"seed_product_id":..., "recommendations":[...], "reasoning":"..."})
- AgenticResponse.success() is from the yaagents-fastapi package
- "recommendations" must be real product dicts from products.json (exact structure preserved)
- Respect X-Customer-Id header and exclude_purchased body param
- /healthz and /readyz endpoints unchanged
```

---

## Starter Prompt 1 — Wire in OpenAI

```
Here is my current recommender.py (paste file contents).

Replace mock_recommend() with a real call to the OpenAI Chat Completions API
(openai Python package, latest version).

The LLM should receive:
- Seed product: name, category, price
- Customer purchased products: names and categories (if customer_id is set)
- The full product catalog (paste products.json or provide its content inline)

Ask the model to return a JSON array of exactly 3 product IDs from the catalog,
plus a one-sentence reasoning string.

Requirements:
1. Read OPENAI_API_KEY from os.environ; raise HTTPException(503) if absent
2. Validate returned IDs against products.json before returning
3. Fallback to mock_recommend() on any API error
4. Keep return type: tuple[list[dict], str]

Also update pyproject.toml to add openai as a dependency.
Show the complete updated recommender.py.
```

---

## Starter Prompt 2 — Use structured outputs

```
Update the OpenAI recommender to use structured outputs (response_format with
json_schema) so the model is forced to return:
{
  "product_ids": ["p-X", "p-Y", "p-Z"],
  "reasoning": "one sentence"
}

This avoids parsing errors. Show the updated recommender.py using the
openai.beta.chat.completions.parse() method (or the equivalent in the
latest SDK version).
```

---

## Starter Prompt 3 — Add a new endpoint

```
Add POST /carts/{id}:checkout-assist to app.py.

Body: {"cart_items": ["p-1", "p-5"]}

Call the OpenAI API to return:
{
  "upsell_suggestions": [/* 2-3 product dicts */],
  "abandonment_copy": "..."
}

Return this via AgenticResponse.success(body={...}).
Update routes.yaml to expose the new route.
Show app.py changes and the updated routes.yaml.
```
