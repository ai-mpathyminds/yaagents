# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents store example FastAPI service — YAAgents Agentic REST Profile v0.3.

Exposes POST /products/{id}/recommendations — returns mock product
recommendations from the same category as the seed product.

To extend with a real LLM, see examples/store/skills/.
"""

from __future__ import annotations

from typing import Annotated, Any

from fastapi import Depends, FastAPI, HTTPException
from fastapi.responses import Response
from pydantic import BaseModel

from yaagents_fastapi import (
    AgenticContext,
    AgenticResponse,
    AgenticResponses,
    AgenticRouter,
    agentic_operation,
    agentic_route_kwargs,
)

from .data import get_product
from .recommender import mock_recommend


_PROFILE_VERSION = "v0.3"

app = FastAPI(
    title="yaagents store example",
    version="0.3.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 ecommerce reference example. "
        "POST /products/{id}/recommendations returns mock product recommendations."
    ),
)

router = AgenticRouter()


CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


class RecommendBody(BaseModel):
    """Request body for POST /products/{id}/recommendations."""

    limit: int = 3
    exclude_purchased: bool = True


@agentic_operation(
    resource="ProductRecommendations",
    operation_kind="recommendation",
    mutating=False,
    responses=AgenticResponses(
        success=True,
    ),
)
async def recommend_products(
    product_id: str,
    body: RecommendBody,
    ctx: CtxDep,
) -> Response:
    """Return 3 product recommendations for the given seed product.

    Mock implementation returns same-category products excluding the seed
    and (when X-Actor-Subject identifies a known customer) any in that
    customer's purchase history. Real implementations replace
    ``mock_recommend()`` with LLM calls — see examples/store/skills/.
    """
    if get_product(product_id) is None:
        raise HTTPException(status_code=404, detail=f"Product {product_id} not found.")

    recommendations, reasoning = mock_recommend(
        product_id=product_id,
        customer_id=ctx.actor_id,
        limit=body.limit,
        exclude_purchased=body.exclude_purchased,
    )

    return AgenticResponse.success(
        body={
            "seed_product_id": product_id,
            "recommendations": recommendations,
            "reasoning": reasoning,
        },
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


app.add_api_route(
    "/products/{product_id}/recommendations",
    recommend_products,
    methods=["POST"],
    **agentic_route_kwargs(recommend_products),
    summary="Get product recommendations",
    tags=["products"],
)


@app.get("/healthz")
async def healthz() -> dict[str, str]:
    return {"status": "ok"}
