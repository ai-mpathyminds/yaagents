# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents store example FastAPI service — YAAgents Agentic REST Profile v0.3.

Exposes POST /products/{id}/recommendations — returns mock product recommendations
from the same category as the seed product.

To extend with a real LLM, open examples/store/skills/<your-ai-tool>.md and
follow the starter prompts. The Profile contract (response shape, headers) is
preserved by the sdk-fastapi helpers used here.
"""

from __future__ import annotations

from fastapi import FastAPI, Header, HTTPException, Request
from pydantic import BaseModel

from yaagents_fastapi import AgenticResponse

from .data import get_product
from .recommender import mock_recommend

app = FastAPI(
    title="yaagents store example",
    version="0.3.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 ecommerce reference example. "
        "POST /products/{id}/recommendations returns mock product recommendations."
    ),
)

class RecommendBody(BaseModel):
    """Request body for POST /products/{id}/recommendations."""

    limit: int = 3
    """Maximum number of recommendations to return."""
    exclude_purchased: bool = True
    """When true and X-Customer-Id is set, filter already-purchased products."""

@app.post(
    "/products/{product_id}/recommendations",
    summary="Get product recommendations",
    responses={
        200: {"description": "Recommendations returned."},
        404: {"description": "Product not found."},
    },
)
async def recommend(
    product_id: str,
    body: RecommendBody,
    request: Request,
    x_customer_id: str | None = Header(None),
) -> dict:
    """Return same-category product recommendations for a given product.

    - Pass ``X-Customer-Id`` header to personalise (filters purchased items).
    - Replace ``mock_recommend()`` in ``recommender.py`` with a real LLM call;
      see ``examples/store/skills/`` for AI-tool starter prompts.
    """
    if not get_product(product_id):
        raise HTTPException(status_code=404, detail=f"Product '{product_id}' not found.")

    recommendations, reason = mock_recommend(
        product_id=product_id,
        customer_id=x_customer_id,
        limit=body.limit,
        exclude_purchased=body.exclude_purchased,
    )

    return AgenticResponse.success(
        body={
            "seed_product_id": product_id,
            "recommendations": recommendations,
            "reasoning": reason,
        },
        correlation_id=request.headers.get("X-Correlation-ID", ""),
        request_id=request.headers.get("X-Request-ID", ""),
    )

@app.get("/healthz", include_in_schema=False)
async def healthz() -> dict[str, str]:
    """Liveness probe."""
    return {"status": "ok"}

@app.get("/readyz", include_in_schema=False)
async def readyz() -> dict[str, str]:
    """Readiness probe — reports Profile version."""
    return {"status": "ok", "profile": "v0.3"}
