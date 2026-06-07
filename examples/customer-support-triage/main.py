# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Customer Support Triage — YAAgents Agentic REST Profile v0.3 example.

Demonstrates the ``clarification_required`` (400) agentic flow via
``POST /tickets/{ticketId}:triage``.

Supports YAAgents Profile v0.3 (spec/agentic-rest-profile.md).
"""

from __future__ import annotations

import os
from typing import Annotated

from fastapi import Depends, FastAPI
from fastapi.responses import Response
from pydantic import BaseModel
from yaagents_fastapi import (
    AgenticContext,
    AgenticResponse,
    RequiredInput,
)

# ── app setup ─────────────────────────────────────────────────────────────────

_PROFILE_VERSION = "v0.3"

app = FastAPI(
    title="Customer Support Triage",
    version="0.1.0",
    description=(
        "YAAgents Agentic REST Profile v0.3 example. "
        "POST /tickets/{ticketId}:triage — clarification_required flow."
    ),
)


# ── request model ─────────────────────────────────────────────────────────────


class TriageRequest(BaseModel):
    """Body for POST /tickets/{ticketId}:triage.

    Both ``severity`` and ``category`` are optional in the request so the
    handler can demonstrate the ``clarification_required`` agentic flow when
    either is absent.
    """

    severity: str | None = None  # "high" | "medium" | "low"
    category: str | None = None  # "billing" | "technical" | "shipping"
    description: str = ""


# ── dependency alias ──────────────────────────────────────────────────────────

CtxDep = Annotated[AgenticContext, Depends(AgenticContext)]


# ── POST /tickets/{ticketId}:triage ───────────────────────────────────────────


@app.post(
    "/tickets/{ticket_id}:triage",
    summary="Triage a support ticket",
    tags=["tickets"],
    responses={
        201: {"description": "Ticket triaged successfully."},
        400: {"description": "clarification_required — severity or category missing."},
    },
)
async def triage_ticket(
    ticket_id: str,
    body: TriageRequest,
    ctx: CtxDep,
) -> Response:
    """Triage a support ticket.

    Demo flows:
    - **clarification_required** (400): omit ``severity`` and/or ``category``.
    - **created / high-severity** (201): ``severity="high"`` → escalate to
      level-2 support team.
    - **created / low-severity** (201): ``severity="low"`` or ``"medium"`` →
      autoresolve hint returned.
    """
    # ── clarification_required — missing required fields ──────────────────────
    missing: list[RequiredInput] = []
    if not body.severity:
        missing.append(
            RequiredInput(
                name="severity",
                location="body",
                type="string",
                required=True,
                question="What is the urgency level of this ticket?",
                allowed_values=["high", "medium", "low"],
            )
        )
    if not body.category:
        missing.append(
            RequiredInput(
                name="category",
                location="body",
                type="string",
                required=True,
                question="What type of issue does this ticket describe?",
                allowed_values=["billing", "technical", "shipping"],
            )
        )

    if missing:
        return AgenticResponse.clarification_required(
            required_inputs=[ri.to_dict() for ri in missing],
            message="Ticket severity and category are required to triage.",
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    # ── created ───────────────────────────────────────────────────────────────
    if body.severity == "high":
        triage_result: dict = {
            "ticketId": ticket_id,
            "status": "escalated",
            "recommendedOwner": "support-team-level-2@example.com",
            "priority": "P1",
            "category": body.category,
        }
    else:  # medium or low
        triage_result = {
            "ticketId": ticket_id,
            "status": "auto-resolved",
            "autoresolveHint": (
                "Please check https://help.example.com/faq "
                "for self-service options."
            ),
            "priority": "P3",
            "category": body.category,
        }

    return AgenticResponse.created(
        body={"ticket": triage_result},
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

    port = int(os.environ.get("PORT", "8122"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
