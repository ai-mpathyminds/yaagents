# Extend the yaagents store example — Cursor

**How to use**: Add the `.cursorrules` snippet below to the project root, then
use the Cursor Composer (`Cmd/Ctrl+I`) or Chat (`Cmd/Ctrl+L`) with the starter prompts.

---

## `.cursorrules` snippet

Add this to `.cursorrules` in the `examples/store/` directory (or the yaagents repo root):

```
## yaagents store example

When working in examples/store/, you are extending a yaagents reference implementation.

Core files:
- src/store/recommender.py  — mock_recommend() is the function to replace with LLM calls
- src/store/app.py          — FastAPI endpoints; return shape must stay unchanged
- data/products.json        — 20 static products; recommendations must come from here

YAAgents Profile v0.3 contract (DO NOT CHANGE):
- Return AgenticResponse.success(body=...) for success (200)
- body shape: {"seed_product_id": str, "recommendations": list[dict], "reasoning": str}
- product dicts in recommendations must be exact entries from products.json
- X-Customer-Id header support must be preserved
- /healthz and /readyz unchanged

For 404s: raise HTTPException(status_code=404, detail="...")
For 5xx: raise HTTPException(status_code=503, detail="...") or AgenticResponse.error()
```

---

## Cursor Composer — Starter Prompt 1 (inline edit)

Open `src/store/recommender.py` in Cursor, select the `mock_recommend` function,
then in Composer (`Cmd/Ctrl+I`):

```
Replace this mock_recommend() function with a real call to Anthropic Claude.
Use the anthropic Python package. Pass the seed product and customer context.
Ask Claude to pick 3 products from @data/products.json and return them with reasoning.

Keep the signature: mock_recommend(product_id, customer_id, limit, exclude_purchased) → (list[dict], str)
Read ANTHROPIC_API_KEY from os.environ; fall back to the original mock if the key is absent.
Add anthropic to @pyproject.toml dependencies.
```

---

## Cursor Chat — Starter Prompt 2 (add endpoint)

```
In @src/store/app.py, add POST /carts/{id}:checkout-assist.

Request: CartAssistBody with cart_items: list[str] (product IDs).
Call Claude to return upsell_suggestions (list of product dicts from @data/products.json)
and abandonment_copy (str).
Return via AgenticResponse.success(body={...}).

Then update @routes.yaml to add the route targeting http://store:8080.
```

---

## Cursor Chat — Starter Prompt 3 (Go mirror)

```
Look at @examples/store-go/recommender.go.
Replace mockRecommend() with a call to the Anthropic Go SDK
(github.com/anthropics/anthropic-sdk-go).
Keep func signature: mockRecommend(products, seed, customer, limit, excludePurchased) ([]Product, string)
Read ANTHROPIC_API_KEY from os.Getenv; fall back to mock if absent.
Update @examples/store-go/go.mod.
```
