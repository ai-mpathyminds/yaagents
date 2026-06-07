# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Inventory Agent — YAAgents Agentic REST Profile v0.3 inter-agent example (AF-1).

Serves GET /inventory/{productId}/stock — returns stock availability to callers.
Demonstrates Profile v0.3 outcome propagation: 424 failed_dependency when the
backend is unavailable (toggled via INVENTORY_UNAVAILABLE env var).

Inter-agent demonstration points:
- actor_id in logs reflects the service-account token (set by gateway-B via
  X-Actor-Subject), NOT the end-user identity — proof of service-account auth.
- X-Tenant-ID propagated from the end-user's original request (forwarded by
  recommendation-agent; gateway-B passes it through).
- Idempotency-Key dedup: same (product, key) pair returns cached result;
  "stock check performed" log line appears only once per unique key.
"""

from __future__ import annotations

import logging
import os
from typing import Annotated

from fastapi import Depends, FastAPI, Header
from fastapi.responses import Response
from yaagents_fastapi import AgenticContext, AgenticResponse

# ── structured logging ─────────────────────────────────────────────────────────

_LOG_FMT = (
    '{"time":"%(asctime)s","level":"%(levelname)s",'
    '"service":"inventory-agent","msg":"%(message)s"}'
)
logging.basicConfig(level=logging.INFO, format=_LOG_FMT)
log = logging.getLogger("inventory-agent")

# ── app ────────────────────────────────────────────────────────────────────────

_PROFILE_VERSION = "v0.3"

app = FastAPI(
    title="Inventory Agent",
    version="0.1.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 inter-agent example (AF-1). "
        "GET /inventory/{productId}/stock — called by recommendation-agent."
    ),
)

# ── demo data ──────────────────────────────────────────────────────────────────

# Products that are out of stock (all others return in_stock=true).
_OUT_OF_STOCK: frozenset[str] = frozenset({"p-99", "p-out"})

# In-memory idempotency cache: (product_id, idempotency_key) → in_stock result.
# Demo only — never expires; process-local.
_idempotency_cache: dict[tuple[str, str], bool] = {}

# ── dependency alias ──────────────────────────────────────────────────────────

CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


# ── GET /inventory/{productId}/stock ─────────────────────────────────────────


@app.get(
    "/inventory/{product_id}/stock",
    summary="Check stock availability for a product",
    tags=["inventory"],
    responses={
        200: {"description": "Stock status for the product."},
        424: {"description": "failed_dependency — inventory backend unavailable."},
    },
)
async def get_stock(
    product_id: str,
    ctx: CtxDep,
    idempotency_key: Annotated[str | None, Header(alias="Idempotency-Key")] = None,
) -> Response:
    """Return stock availability for a product.

    Demo flows:

    - **INVENTORY_UNAVAILABLE=true** (env) → 424 ``failed_dependency``
      (simulates backend down; recommendation-agent must propagate this upstream).
    - **product_id in {p-99, p-out}** → ``in_stock=false``.
    - **Idempotency-Key dedup**: same key for the same product returns the
      cached result; the "stock check performed" log line is emitted only once.
    """
    cache_key = (product_id, idempotency_key) if idempotency_key else None

    # ── idempotency cache hit ─────────────────────────────────────────────────
    if cache_key and cache_key in _idempotency_cache:
        in_stock = _idempotency_cache[cache_key]
        log.info(
            "idempotency cache hit | product=%s in_stock=%s "
            "tenant=%s actor=%s key=%s",
            product_id,
            in_stock,
            ctx.tenant_id,
            ctx.actor_id,
            idempotency_key,
        )
        return AgenticResponse.success(
            body={
                "productId": product_id,
                "inStock": in_stock,
                "source": "idempotency-cache",
            },
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # ── simulated backend unavailability ─────────────────────────────────────
    if os.environ.get("INVENTORY_UNAVAILABLE", "false").lower() == "true":
        log.warning(
            "inventory backend unavailable | product=%s tenant=%s",
            product_id,
            ctx.tenant_id,
        )
        return AgenticResponse.failed_dependency(
            code="INVENTORY_BACKEND_DOWN",
            message="Inventory backend is currently unavailable. Retry later.",
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # ── stock check ───────────────────────────────────────────────────────────
    in_stock = product_id not in _OUT_OF_STOCK
    log.info(
        "stock check performed | product=%s in_stock=%s "
        "tenant=%s actor=%s key=%s",
        product_id,
        in_stock,
        ctx.tenant_id,
        ctx.actor_id,
        idempotency_key or "none",
    )

    # Cache result for this idempotency key.
    if cache_key:
        _idempotency_cache[cache_key] = in_stock

    return AgenticResponse.success(
        body={"productId": product_id, "inStock": in_stock},
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


# ── toggle endpoint (demo only) ───────────────────────────────────────────────


@app.post("/_demo/inventory-down", include_in_schema=False)
async def toggle_inventory_down(enabled: bool = True) -> dict:
    """Toggle INVENTORY_UNAVAILABLE for demo purposes."""
    os.environ["INVENTORY_UNAVAILABLE"] = "true" if enabled else "false"
    return {"inventoryUnavailable": enabled}


# ── entry-point ───────────────────────────────────────────────────────────────

if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", "8127"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
