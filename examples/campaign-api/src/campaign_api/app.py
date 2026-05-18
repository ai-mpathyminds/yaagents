"""Campaign API FastAPI application — YAAgents Agentic REST Profile v0.1.

Five PRD §6.1 endpoints, four §6.2 demo flows.

Supports YAAgents Profile v0.1 (spec/agentic-rest-profile.md).
"""

from __future__ import annotations

import os
from typing import Annotated, Any

from fastapi import Depends, FastAPI, HTTPException
from fastapi.responses import Response
from yaagents_fastapi import (
    AgenticContext,
    AgenticResponse,
    AgenticResponses,
    AgenticRouter,
    RequiredInput,
    agentic_operation,
    agentic_route_kwargs,
)

import campaign_api.store as _store
from campaign_api.models import (
    CreateCampaignRequest,
    CreateOptimizationRequest,
    GenerateAssetsRequest,
)

# ── app setup ─────────────────────────────────────────────────────────────────

_PROFILE_VERSION = "v0.1"

app = FastAPI(
    title="Campaign API",
    version="0.1.0",
    description=(
        "YAAgents Agentic REST Profile v0.1 reference example. "
        "POST /campaigns, GET /campaigns/{id}, "
        "POST /campaigns/{id}/optimizations, "
        "GET /campaigns/{id}/optimizations/{opId}, "
        "POST /campaigns/{id}/assets:generate"
    ),
)

router = AgenticRouter()


# ── helpers ───────────────────────────────────────────────────────────────────


def _ctx_dep() -> type[AgenticContext]:
    return AgenticContext


CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


# ── POST /campaigns ───────────────────────────────────────────────────────────


@agentic_operation(
    resource="Campaign",
    operation_kind="mutation",
    mutating=True,
    responses=AgenticResponses(
        created=True,
        clarification_required=True,
        validation_failed=True,
    ),
)
async def create_campaign(
    body: CreateCampaignRequest,
    ctx: CtxDep,
) -> Response:
    """Create a new campaign.

    Demo flows:
    - §6.2 flow 2 (clarification_required): omit ``successMetric`` in request.
    - §6.2 flow 3 (validation_failed): send ``budget`` as a string / negative value.
    - §6.2 flow 1 (created): valid body with all fields → 201.
    """
    # Flow 2: clarification_required — missing successMetric
    if body.successMetric is None:
        return AgenticResponse.clarification_required(
            required_inputs=[
                RequiredInput(
                    name="successMetric",
                    location="body",
                    type="string",
                    required=True,
                    question=(
                        "Which success metric should this campaign optimize?"
                    ),
                    allowed_values=["ctr", "cpl", "conversion_rate", "roas"],
                ).to_dict()
            ],
            message="successMetric is required to create a campaign.",
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # Flow 1: success/created — all fields present
    record = _store.create_campaign(
        name=body.name,
        budget=body.budget,
        target_audience=body.targetAudience,
        success_metric=body.successMetric,
    )
    return AgenticResponse.created(
        body={"campaign": record},
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


app.add_api_route(
    "/campaigns",
    create_campaign,
    methods=["POST"],
    **agentic_route_kwargs(create_campaign),
    summary="Create a campaign",
    tags=["campaigns"],
)


# ── GET /campaigns/{campaign_id} ──────────────────────────────────────────────


@app.get(
    "/campaigns/{campaign_id}",
    summary="Get a campaign",
    tags=["campaigns"],
    responses={
        200: {"description": "Campaign resource."},
        404: {"description": "Not found."},
    },
)
async def get_campaign(
    campaign_id: str,
    ctx: CtxDep,
) -> Response:
    """Retrieve an existing campaign by ID."""
    record = _store.get_campaign(campaign_id)
    if record is None:
        raise HTTPException(status_code=404, detail="Campaign not found.")
    return AgenticResponse.success(
        body={"campaign": record},
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


# ── POST /campaigns/{campaign_id}/optimizations ───────────────────────────────


@agentic_operation(
    resource="Campaign",
    operation_kind="recommendation",
    mutating=False,
    responses=AgenticResponses(
        created=True,
        clarification_required=True,
        validation_failed=True,
        failed_dependency=True,
    ),
)
async def create_optimization(
    campaign_id: str,
    body: CreateOptimizationRequest,
    ctx: CtxDep,
) -> Response:
    """Request AI-generated optimizations for a campaign.

    Demo flows:
    - §6.2 flow 4 (failed_dependency): toggle ``LLM_DOWN = True`` in store.
    - §6.2 flow 1 (created): valid body, campaign exists, LLM up.
    """
    # 404 guard
    campaign = _store.get_campaign(campaign_id)
    if campaign is None:
        raise HTTPException(status_code=404, detail="Campaign not found.")

    # Flow 4: failed_dependency — LLM service down
    if _store.LLM_DOWN:
        return AgenticResponse.failed_dependency(
            code="LLM_UNAVAILABLE",
            message=(
                "The AI optimisation engine is currently unavailable. "
                "Please retry in a few minutes."
            ),
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # Deterministic suggestions (no real LLM call in this demo)
    suggestions: list[dict[str, Any]] = [
        {
            "id": f"sug-{i + 1}",
            "objective": obj,
            "suggestion": f"Increase bid for '{obj}' by 10% during peak hours.",
            "estimatedImpact": "+12% improvement",
        }
        for i, obj in enumerate(body.objectives[: body.maxSuggestions])
    ]
    record = _store.create_optimization(campaign_id, suggestions)
    return AgenticResponse.created(
        body={"optimization": record},
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


app.add_api_route(
    "/campaigns/{campaign_id}/optimizations",
    create_optimization,
    methods=["POST"],
    **agentic_route_kwargs(create_optimization),
    summary="Create an optimization for a campaign",
    tags=["optimizations"],
)


# ── GET /campaigns/{campaign_id}/optimizations/{op_id} ───────────────────────


@app.get(
    "/campaigns/{campaign_id}/optimizations/{op_id}",
    summary="Get an optimization",
    tags=["optimizations"],
    responses={
        200: {"description": "Optimization resource."},
        404: {"description": "Not found."},
    },
)
async def get_optimization(
    campaign_id: str,
    op_id: str,
    ctx: CtxDep,
) -> Response:
    """Retrieve a specific optimization result."""
    record = _store.get_optimization(campaign_id, op_id)
    if record is None:
        raise HTTPException(status_code=404, detail="Optimization not found.")
    return AgenticResponse.success(
        body={"optimization": record},
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


# ── POST /campaigns/{campaign_id}/assets:generate ────────────────────────────


@agentic_operation(
    resource="Campaign",
    operation_kind="generation",
    mutating=False,
    responses=AgenticResponses(
        created=True,
        failed_dependency=True,
    ),
)
async def generate_assets(
    campaign_id: str,
    body: GenerateAssetsRequest,
    ctx: CtxDep,
) -> Response:
    """Generate creative assets for a campaign.

    Demo flow 4 also applies here — toggle LLM_DOWN to get failed_dependency.
    """
    campaign = _store.get_campaign(campaign_id)
    if campaign is None:
        raise HTTPException(status_code=404, detail="Campaign not found.")

    if _store.LLM_DOWN:
        return AgenticResponse.failed_dependency(
            code="LLM_UNAVAILABLE",
            message=(
                "The AI asset generation engine is currently unavailable. "
                "Please retry in a few minutes."
            ),
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # Deterministic stub assets
    assets = [
        {
            "id": f"asset-{i + 1}",
            "type": asset_type,
            "tone": body.tone,
            "content": (
                f"[Demo] {asset_type.title()} asset for campaign "
                f"'{campaign['name']}' with {body.tone} tone."
            ),
        }
        for i, asset_type in enumerate(body.assetTypes)
    ]
    return AgenticResponse.created(
        body={"assets": assets, "campaignId": campaign_id},
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


app.add_api_route(
    "/campaigns/{campaign_id}/assets:generate",
    generate_assets,
    methods=["POST"],
    **agentic_route_kwargs(generate_assets),
    summary="Generate assets for a campaign",
    tags=["assets"],
)


# ── health endpoints ──────────────────────────────────────────────────────────


@app.get("/healthz", include_in_schema=False)
async def healthz() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/readyz", include_in_schema=False)
async def readyz() -> dict[str, str]:
    return {"status": "ok", "profile": _PROFILE_VERSION}


# ── demo-only toggle endpoint ─────────────────────────────────────────────────


@app.post(
    "/_demo/llm-down",
    include_in_schema=False,
    summary="Toggle LLM-down state (demo only)",
)
async def toggle_llm_down(enabled: bool = True) -> dict[str, Any]:
    """Toggle the LLM-down flag to demo flow 4 (failed_dependency).

    Not part of the Profile; excluded from /openapi.json.
    """
    _store.LLM_DOWN = enabled
    return {"llmDown": _store.LLM_DOWN}


# ── app entry-point ───────────────────────────────────────────────────────────

if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", "8121"))
    uvicorn.run("campaign_api.app:app", host="0.0.0.0", port=port, reload=False)
