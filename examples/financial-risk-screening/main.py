# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Financial Risk Screening — YAAgents Agentic REST Profile v0.3 example.

Demonstrates the ``approval_required`` (412) agentic flow via
``POST /claims/{claimId}/risk-screens``.

Also demonstrates a license-tier gate: community-tier callers receive
403 ``application/vnd.yaagents.error+json`` (license-check plugin gate;
handled at service level in this demo since no real license server is running).

Supports YAAgents Profile v0.3 (spec/agentic-rest-profile.md).
"""

from __future__ import annotations

import os
from typing import Annotated

from fastapi import Depends, FastAPI, Header
from fastapi.responses import Response
from pydantic import BaseModel, Field
from yaagents_fastapi import (
    AgenticContext,
    AgenticResponse,
)

# ── app setup ─────────────────────────────────────────────────────────────────

_PROFILE_VERSION = "v0.3"
_RISK_THRESHOLD = 0.7  # risk-score above this requires human approval

app = FastAPI(
    title="Financial Risk Screening",
    version="0.1.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 example. "
        "POST /claims/{claimId}/risk-screens — approval_required flow. "
        "license-check gate: community tier → 403."
    ),
)


# ── request model ─────────────────────────────────────────────────────────────


class RiskScreenRequest(BaseModel):
    """Body for POST /claims/{claimId}/risk-screens.

    ``amount`` (USD) and ``claimant_history`` drive the mock risk-score
    calculation used to trigger the ``approval_required`` flow.
    """

    amount: float = Field(gt=0, description="Claim amount in USD")
    claimant_history: str = Field(
        default="good",
        description="Claimant risk profile: good | neutral | bad",
    )
    notes: str = ""


# ── dependency alias ──────────────────────────────────────────────────────────

CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


# ── risk-score helper ─────────────────────────────────────────────────────────


def _compute_risk_score(amount: float, history: str) -> float:
    """Return a mock risk score in [0.0, 1.0].

    Deterministic formula for demo purposes — no real ML model involved.
    High claim amounts and adverse claimant history push the score up.
    Threshold: $10 000 at neutral history crosses 0.7.
    """
    history_multiplier = {"good": 0.7, "neutral": 1.0, "bad": 1.4}.get(
        history.lower(), 1.0
    )
    base = min(amount / 10_000.0, 1.0)
    return min(round(base * history_multiplier, 3), 1.0)


# ── POST /claims/{claimId}/risk-screens ───────────────────────────────────────


@app.post(
    "/claims/{claim_id}/risk-screens",
    summary="Screen a claim for financial risk",
    tags=["claims"],
    responses={
        201: {"description": "Claim approved; risk score within threshold."},
        403: {
            "description": (
                "license_tier_insufficient — community tier not allowed."
            )
        },
        412: {
            "description": (
                "approval_required — risk score exceeds threshold."
            )
        },
    },
)
async def screen_claim(
    claim_id: str,
    body: RiskScreenRequest,
    ctx: CtxDep,
    x_license_tier: Annotated[str, Header()] = "community",
) -> Response:
    """Screen a financial claim for risk.

    Demo flows:
    - **license-check gate** (403): omit ``X-License-Tier`` header or set it
      to ``community`` — service returns 403 ``application/vnd.yaagents.error+json``.
    - **approval_required** (412): ``amount > $10 000`` with neutral history
      (or smaller amounts with ``claimant_history="bad"``) exceed the 0.7
      threshold → 412 ``application/vnd.yaagents.approval-required+json``.
    - **created** (201): low-risk claim (amount ≤ $10 000 with good history)
      → 201 ``application/json`` with screening result.
    """
    # ── license-check gate ────────────────────────────────────────────────────
    # In production, the gateway license-check plugin enforces this before the
    # request reaches the service. In this demo the service checks the header
    # directly (plugin is disabled; see gateway-plugins.yaml comments).
    if x_license_tier.lower() == "community":
        return AgenticResponse.forbidden(
            code="LICENSE_TIER_INSUFFICIENT",
            message=(
                "Risk screening requires a professional or enterprise license. "
                "Upgrade your plan to access this operation."
            ),
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # ── risk scoring ──────────────────────────────────────────────────────────
    score = _compute_risk_score(body.amount, body.claimant_history)

    if score > _RISK_THRESHOLD:
        return AgenticResponse.approval_required(
            approval_token=f"appr-{claim_id}-{int(body.amount)}",
            message=(
                "High-risk claim requires human review before processing. "
                f"Risk score: {score:.3f} (threshold: {_RISK_THRESHOLD})."
            ),
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # ── approved ──────────────────────────────────────────────────────────────
    result: dict = {
        "claimId": claim_id,
        "riskScore": score,
        "status": "approved",
        "amount": body.amount,
        "claimantHistory": body.claimant_history,
    }
    return AgenticResponse.created(
        body={"screen": result},
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


# ── app entry-point ───────────────────────────────────────────────────────────

if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", "8123"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
