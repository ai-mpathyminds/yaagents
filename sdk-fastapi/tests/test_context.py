# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for AgenticContext and RequiredInput — WI-1yaa.SDK-2.

AC:
- AgenticContext injectable via Depends; all 4 fields populated from headers.
- RequiredInput round-trips into a §4.1-shaped requiredInputs[] entry.
"""

from __future__ import annotations

import re
from typing import Annotated, Any

import pytest
from fastapi import Depends, FastAPI
from fastapi.responses import Response
from fastapi.testclient import TestClient

from yaagents_fastapi.context import AgenticContext, RequiredInput
from yaagents_fastapi.response import AgenticResponse

# ── test app ──────────────────────────────────────────────────────────────────

app = FastAPI()


@app.get("/ctx-dump")
def ctx_dump(
    ctx: Annotated[AgenticContext, Depends(AgenticContext)],
) -> dict[str, str]:
    """Echo all four context fields as JSON so tests can inspect them."""
    return {
        "tenant_id": ctx.tenant_id,
        "actor_id": ctx.actor_id,
        "correlation_id": ctx.correlation_id,
        "request_id": ctx.request_id,
    }


@app.post("/clarify")
def clarify_endpoint(
    ctx: Annotated[AgenticContext, Depends(AgenticContext)],
) -> Response:
    """Return a clarification_required response using context IDs."""
    inputs = [
        RequiredInput(
            name="successMetric",
            location="body",
            type="string",
            required=True,
            question="Which success metric?",
            allowed_values=["ctr", "cpl"],
        )
    ]
    return AgenticResponse.clarification_required(
        [ri.to_dict() for ri in inputs],
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )


client = TestClient(app, raise_server_exceptions=True)

_UUID4_RE = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)


# ── AgenticContext — header extraction ────────────────────────────────────────


def test_all_four_headers_extracted() -> None:
    resp = client.get(
        "/ctx-dump",
        headers={
            "X-Tenant-ID": "tenant-abc",
            "X-Actor-Subject": "user-123",
            "X-Correlation-ID": "corr-xyz",
            "X-Request-ID": "req-xyz",
        },
    )
    assert resp.status_code == 200
    body = resp.json()
    assert body["tenant_id"] == "tenant-abc"
    assert body["actor_id"] == "user-123"
    assert body["correlation_id"] == "corr-xyz"
    assert body["request_id"] == "req-xyz"


def test_correlation_id_generated_when_absent() -> None:
    resp = client.get(
        "/ctx-dump",
        headers={
            "X-Tenant-ID": "t1",
            "X-Actor-Subject": "u1",
            # No X-Correlation-ID
            "X-Request-ID": "req-001",
        },
    )
    body = resp.json()
    assert _UUID4_RE.match(body["correlation_id"]), (
        f"expected UUID v4, got {body['correlation_id']!r}"
    )


def test_request_id_generated_when_absent() -> None:
    resp = client.get(
        "/ctx-dump",
        headers={
            "X-Correlation-ID": "corr-001",
            # No X-Request-ID
        },
    )
    body = resp.json()
    assert _UUID4_RE.match(body["request_id"]), (
        f"expected UUID v4, got {body['request_id']!r}"
    )


def test_both_trace_ids_generated_when_all_headers_absent() -> None:
    resp = client.get("/ctx-dump")
    body = resp.json()
    assert _UUID4_RE.match(body["correlation_id"])
    assert _UUID4_RE.match(body["request_id"])
    # Two separate requests get distinct IDs
    resp2 = client.get("/ctx-dump")
    body2 = resp2.json()
    assert body["correlation_id"] != body2["correlation_id"]
    assert body["request_id"] != body2["request_id"]


def test_tenant_and_actor_empty_when_headers_absent() -> None:
    resp = client.get("/ctx-dump")
    body = resp.json()
    assert body["tenant_id"] == ""
    assert body["actor_id"] == ""


# ── RequiredInput — to_dict wire shape ────────────────────────────────────────


def test_required_input_full_dict() -> None:
    ri = RequiredInput(
        name="successMetric",
        location="body",
        type="string",
        required=True,
        question="Which metric?",
        allowed_values=["ctr", "cpl"],
    )
    d = ri.to_dict()
    assert d == {
        "name": "successMetric",
        "location": "body",
        "type": "string",
        "required": True,
        "question": "Which metric?",
        "allowedValues": ["ctr", "cpl"],
    }


def test_required_input_omits_allowed_values_when_none() -> None:
    ri = RequiredInput(
        name="budget",
        location="query",
        type="integer",
        required=True,
        question="Enter the budget.",
    )
    d = ri.to_dict()
    assert "allowedValues" not in d
    assert d["type"] == "integer"
    assert d["location"] == "query"


@pytest.mark.parametrize("location", ["body", "query", "path", "header"])
def test_required_input_all_locations(location: str) -> None:
    ri = RequiredInput(
        name="x",
        location=location,  # type: ignore[arg-type]
        type="string",
        required=False,
        question="?",
    )
    assert ri.to_dict()["location"] == location


# ── RequiredInput → clarification_required round-trip (§4.1 shape) ────────────


def test_required_input_feeds_clarification_required() -> None:
    resp = client.post(
        "/clarify",
        headers={"X-Correlation-ID": "corr-rt", "X-Request-ID": "req-rt"},
    )
    assert resp.status_code == 400
    assert resp.headers["content-type"].startswith(
        "application/vnd.yaagents.clarification+json"
    )
    body = resp.json()
    assert body["type"] == "clarification_required"
    assert body["code"] == "CLARIFICATION_REQUIRED"
    assert len(body["requiredInputs"]) == 1
    ri_wire = body["requiredInputs"][0]
    assert ri_wire["name"] == "successMetric"
    assert ri_wire["type"] == "string"
    assert ri_wire["required"] is True
    assert ri_wire["allowedValues"] == ["ctr", "cpl"]
    # trace IDs propagated from gateway headers
    assert body["trace"]["correlationId"] == "corr-rt"
    assert body["trace"]["requestId"] == "req-rt"


def test_required_input_dict_matches_spec_4_1_example() -> None:
    """to_dict() output is structurally identical to spec §4.1 example element."""
    ri = RequiredInput(
        name="successMetric",
        location="body",
        type="string",
        required=True,
        question="Which success metric should be optimized?",
        allowed_values=["ctr", "cpl", "conversion_rate", "lead_quality"],
    )
    expected: dict[str, Any] = {
        "name": "successMetric",
        "location": "body",
        "type": "string",
        "required": True,
        "question": "Which success metric should be optimized?",
        "allowedValues": ["ctr", "cpl", "conversion_rate", "lead_quality"],
    }
    assert ri.to_dict() == expected
