# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Recommendation Agent — YAAgents Agentic REST Profile v0.3 inter-agent example (AF-1).

Serves POST /products/{productId}/recommendations.

Before returning recommendations, calls inventory-agent (via gateway-B) to
filter out out-of-stock candidates.  Demonstrates four AF-1 concepts:

1. **Inter-agent auth**: outgoing call to gateway-B uses a service-account
   token (signed with RECOMMENDATION_SVC_SECRET / jwt.encode at startup),
   distinct from the end-user's Authorization header.

2. **Tenant context propagation**: X-Tenant-ID from the inbound request is
   forwarded to inventory-agent (NOT re-resolved; already validated upstream
   by gateway-A's tenant-injector plugin).

3. **Profile v0.3 outcome propagation**:
   - inventory-agent 424 → this service returns 424 failed_dependency.
   - inventory-agent 400 clarification_required → this service returns 400
     clarification_required.

4. **Idempotency-Key propagation**: the inbound Idempotency-Key is forwarded
   to each inventory call, keyed per-product (``{outer_key}::{product_id}``).
   On retry the inventory-agent returns cached results.
"""

from __future__ import annotations

import logging
import os
import time
from typing import Annotated

import httpx
import jwt  # PyJWT — build-time only dep; not part of sdk-go core
from fastapi import Depends, FastAPI, Header
from fastapi.responses import Response
from pydantic import BaseModel, Field
from yaagents_fastapi import AgenticContext, AgenticResponse

# ── structured logging ─────────────────────────────────────────────────────────

_LOG_FMT = (
    '{"time":"%(asctime)s","level":"%(levelname)s",'
    '"service":"recommendation-agent","msg":"%(message)s"}'
)
logging.basicConfig(level=logging.INFO, format=_LOG_FMT)
log = logging.getLogger("recommendation-agent")

# ── service-account token (generated at startup) ──────────────────────────────
# recommendation-agent mints its own short-lived service-account JWT using
# RECOMMENDATION_SVC_SECRET (= gateway-B's jwt_secret).
# In production: replace with OAuth 2.0 client_credentials flow.

_SVC_SECRET = os.environ.get("RECOMMENDATION_SVC_SECRET", "demo-svc-secret")
_SVC_TOKEN: str = jwt.encode(
    {
        "sub": "recommendation-agent-svc",
        "iss": "demo-svc-idp",
        "iat": int(time.time()),
        "exp": int(time.time()) + 3600 * 24,  # 24 h; refresh on restart (demo)
    },
    _SVC_SECRET,
    algorithm="HS256",
)
log.info(
    "service-account token minted | sub=recommendation-agent-svc "
    "alg=HS256 exp=24h (demo only)"
)

# ── upstream: inventory-agent via gateway-B ───────────────────────────────────

_INVENTORY_BASE_URL = os.environ.get(
    "INVENTORY_GATEWAY_URL", "http://gateway-b:8080"
)

# ── product catalog (demo) ─────────────────────────────────────────────────────
# Maps each product to a static list of candidate recommendations.
# Out-of-stock candidates (p-99, p-out) will be filtered by inventory-agent.

_CANDIDATES: dict[str, list[dict]] = {
    "p-1": [
        {"id": "p-2", "name": "Wireless Keyboard", "price": 49.99},
        {"id": "p-3", "name": "USB-C Hub", "price": 29.99},
        {"id": "p-99", "name": "Discontinued Headset", "price": 0.00},  # out-of-stock
        {"id": "p-out", "name": "Legacy Adapter", "price": 9.99},       # out-of-stock
    ],
    "p-2": [
        {"id": "p-1", "name": "Laptop Stand", "price": 39.99},
        {"id": "p-3", "name": "USB-C Hub", "price": 29.99},
    ],
}
_DEFAULT_CANDIDATES = [
    {"id": "p-1", "name": "Laptop Stand", "price": 39.99},
    {"id": "p-2", "name": "Wireless Keyboard", "price": 49.99},
]

# ── app ────────────────────────────────────────────────────────────────────────

_PROFILE_VERSION = "v0.3"

app = FastAPI(
    title="Recommendation Agent",
    version="0.1.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 inter-agent example (AF-1). "
        "POST /products/{productId}/recommendations — filters by inventory."
    ),
)


# ── request model ─────────────────────────────────────────────────────────────


class RecommendationRequest(BaseModel):
    """Body for POST /products/{productId}/recommendations."""

    limit: int = Field(default=3, ge=1, le=10, description="Max recommendations")


# ── dependency alias ──────────────────────────────────────────────────────────

CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


# ── inventory call helper ─────────────────────────────────────────────────────


async def _check_stock(
    product_id: str,
    tenant_id: str,
    correlation_id: str,
    idempotency_key: str | None,
) -> tuple[bool | None, str | None, str | None]:
    """Call inventory-agent/gateway-B; return (in_stock, error_code, outcome).

    Returns:
        (True/False, None, None)  — stock result
        (None, error_code, "failed_dependency")  — 424 from inventory
        (None, error_code, "clarification_required")  — 400 from inventory
    """
    # Per-product idempotency key keeps inventory calls idempotent across retries.
    inv_idem_key = f"{idempotency_key}::{product_id}" if idempotency_key else None

    headers: dict[str, str] = {
        "Authorization": f"Bearer {_SVC_TOKEN}",
        "X-Tenant-ID": tenant_id,
        "X-Correlation-ID": correlation_id,
    }
    if inv_idem_key:
        headers["Idempotency-Key"] = inv_idem_key

    log.info(
        "calling inventory-agent | product=%s tenant=%s key=%s",
        product_id,
        tenant_id,
        inv_idem_key or "none",
    )

    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            resp = await client.get(
                f"{_INVENTORY_BASE_URL}/inventory/{product_id}/stock",
                headers=headers,
            )
    except (httpx.ConnectError, httpx.TimeoutException) as exc:
        log.error("inventory-agent unreachable | error=%s", exc)
        return None, "INVENTORY_UNREACHABLE", "failed_dependency"

    if resp.status_code == 200:
        data = resp.json()
        return data.get("inStock", False), None, None

    if resp.status_code == 424:
        body = resp.json() if resp.headers.get("content-type", "").startswith(
            "application/vnd.yaagents"
        ) else {}
        return None, body.get("code", "INVENTORY_FAILED"), "failed_dependency"

    if resp.status_code == 400:
        body = resp.json() if resp.headers.get("content-type", "").startswith(
            "application/vnd.yaagents"
        ) else {}
        return None, body.get("code", "INVENTORY_CLARIFICATION"), "clarification_required"

    # Unexpected status — treat as dependency failure.
    return None, f"INVENTORY_STATUS_{resp.status_code}", "failed_dependency"


# ── POST /products/{productId}/recommendations ────────────────────────────────


@app.post(
    "/products/{product_id}/recommendations",
    summary="Get in-stock product recommendations",
    tags=["products"],
    responses={
        200: {"description": "Filtered recommendations (in-stock only)."},
        400: {"description": "clarification_required from inventory-agent."},
        424: {"description": "failed_dependency — inventory-agent unavailable."},
    },
)
async def get_recommendations(
    product_id: str,
    body: RecommendationRequest,
    ctx: CtxDep,
    idempotency_key: Annotated[str | None, Header(alias="Idempotency-Key")] = None,
) -> Response:
    """Return in-stock product recommendations for the given product.

    Calls inventory-agent for each candidate.  Propagates tenant context and
    idempotency key.  Translates typed outcomes from inventory-agent:

    - ``424 failed_dependency`` → own 424 (inventory is a hard dependency).
    - ``400 clarification_required`` → own 400 (cannot proceed without clarity).
    """
    candidates = _CANDIDATES.get(product_id, _DEFAULT_CANDIDATES)[: body.limit]

    log.info(
        "recommendation request | product=%s tenant=%s actor=%s candidates=%d key=%s",
        product_id,
        ctx.tenant_id,
        ctx.actor_id,
        len(candidates),
        idempotency_key or "none",
    )

    recommendations: list[dict] = []

    for candidate in candidates:
        cid = candidate["id"]
        in_stock, err_code, outcome = await _check_stock(
            product_id=cid,
            tenant_id=ctx.tenant_id,
            correlation_id=ctx.correlation_id,
            idempotency_key=idempotency_key,
        )

        if outcome == "failed_dependency":
            # §4.1 — inventory is a hard dependency; propagate 424 upstream.
            log.warning(
                "inventory failed_dependency | product=%s code=%s",
                cid,
                err_code,
            )
            return AgenticResponse.failed_dependency(
                code=err_code or "INVENTORY_UNAVAILABLE",
                message=(
                    "Inventory service is unavailable; cannot filter recommendations. "
                    "Please retry later."
                ),
                correlation_id=ctx.correlation_id,
                request_id=ctx.request_id,
            )

        if outcome == "clarification_required":
            # §4.2 — translate inventory clarification to own clarification.
            log.warning(
                "inventory clarification_required | product=%s code=%s",
                cid,
                err_code,
            )
            return AgenticResponse.clarification_required(
                required_inputs=[
                    {
                        "name": "productId",
                        "location": "path",
                        "type": "string",
                        "required": True,
                        "question": (
                            f"Product '{cid}' was not found in the inventory catalog. "
                            "Please provide a valid productId."
                        ),
                    }
                ],
                message="One or more candidate products could not be looked up.",
                correlation_id=ctx.correlation_id,
                request_id=ctx.request_id,
            )

        if in_stock:
            recommendations.append(candidate)

    log.info(
        "recommendation done | product=%s tenant=%s returned=%d/%d",
        product_id,
        ctx.tenant_id,
        len(recommendations),
        len(candidates),
    )

    return AgenticResponse.success(
        body={
            "productId": product_id,
            "recommendations": recommendations[: body.limit],
            "totalCandidates": len(candidates),
            "filteredOutOfStock": len(candidates) - len(recommendations),
        },
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


# ── health endpoints ──────────────────────────────────────────────────────────


@app.get("/healthz", include_in_schema=False)
async def healthz() -> dict:
    return {"status": "ok"}


@app.get("/readyz", include_in_schema=False)
async def readyz() -> dict:
    return {"status": "ok", "profile": _PROFILE_VERSION}


# ── entry-point ───────────────────────────────────────────────────────────────

if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", "8125"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
