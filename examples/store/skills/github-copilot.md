# Extend the yaagents store example — GitHub Copilot

**How to use**: This file can be used as a Copilot custom instructions file or
as a Copilot Chat prompt. For VS Code, open Copilot Chat (`Ctrl+Shift+I`) and
paste the prompts below.

---

## Option A — `.github/copilot-instructions.md` (project-wide context)

Create `.github/copilot-instructions.md` in the yaagents repo with:

```markdown
## yaagents store example context

This repo contains a yaagents reference example at examples/store/.
The store service has one agentic endpoint: POST /products/{id}/recommendations.

The current mock recommender is in examples/store/src/store/recommender.py.
When asked to extend the recommender, always:
1. Keep AgenticResponse.success(body=...) as the return type from app.py
2. Keep the body shape: {"seed_product_id": ..., "recommendations": [...], "reasoning": "..."}
3. Recommendations must be real product dicts from examples/store/data/products.json
4. Support X-Customer-Id header for personalisation

Profile spec: spec/agentic-rest-profile.md
```

---

## Copilot Chat — Starter Prompt 1 (open recommender.py first)

```
Open examples/store/src/store/recommender.py.

Replace mock_recommend() with a real call to the Anthropic Claude API.
Use the anthropic Python package. Pass the seed product details and the customer's
purchase history as context. Ask Claude to select 3 products from the products.json
catalog (share the file path) and return them with reasoning.

Requirements:
- ANTHROPIC_API_KEY from os.environ; fallback to mock if unset
- Return type must still be tuple[list[dict], str]
- Product dicts must be valid entries from products.json (no made-up products)

Also update examples/store/pyproject.toml to add anthropic as a dependency.
```

---

## Copilot Chat — Starter Prompt 2 (add checkout assist)

```
Add a new endpoint to examples/store/src/store/app.py:
POST /carts/{id}:checkout-assist

Request body (Pydantic model):
  cart_items: list[str]  # product IDs currently in the cart

The endpoint should call an LLM to suggest upsell products and generate
abandonment-prevention copy.

Response (use AgenticResponse.success from yaagents_fastapi):
{
  "upsell_suggestions": [/* up to 3 product dicts */],
  "abandonment_copy": "string"
}

Also add the route to examples/store/routes.yaml targeting http://store:8080.
```

---

## Copilot Chat — Starter Prompt 3 (Go mirror)

```
Look at examples/store-go/recommender.go.

Replace mockRecommend() with a real call to the Anthropic API using the
Go anthropic-sdk-go package (github.com/anthropics/anthropic-sdk-go).

Keep the return type as ([]Product, string).
Add the dependency to examples/store-go/go.mod.
Read ANTHROPIC_API_KEY from os.Getenv("ANTHROPIC_API_KEY"); return the
mock result as fallback if the key is absent.
```
