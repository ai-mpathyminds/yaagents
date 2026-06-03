# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Mock product recommender for the yaagents store example.

Returns same-category products as recommendations.
Replace mock_recommend() with a real LLM call to make this a genuine
AI-powered recommender. See examples/store/skills/ for starter prompts.
"""

from __future__ import annotations

from .data import get_customer, get_product, load_products

def mock_recommend(
    product_id: str,
    customer_id: str | None = None,
    limit: int = 3,
    exclude_purchased: bool = True,
) -> tuple[list[dict], str]:
    """Return same-category recommendations for a product.

    Args:
        product_id: ID of the seed product.
        customer_id: Optional customer ID; used to filter purchased items.
        limit: Maximum number of recommendations to return.
        exclude_purchased: When True and customer_id is set, filter out
            products already purchased by this customer.

    Returns:
        A (recommendations, reasoning) tuple. recommendations is a list
        of product dicts; reasoning is a human-readable explanation string.
    """
    seed = get_product(product_id)
    if not seed:
        return [], "Product not found"

    all_products = load_products()
    candidates = [
        p
        for p in all_products
        if p["category"] == seed["category"] and p["id"] != product_id
    ]

    if customer_id and exclude_purchased:
        customer = get_customer(customer_id)
        if customer:
            purchased = set(customer.get("purchased", []))
            candidates = [p for p in candidates if p["id"] not in purchased]

    recommendations = candidates[:limit]
    reason = f"Same category as '{seed['name']}' ({seed['category']})"
    if customer_id:
        reason += f"; personalised for customer {customer_id} (purchased items excluded)"
    return recommendations, reason
