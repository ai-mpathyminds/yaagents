# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for AgenticResponse factory — WI-1yaa.SDK-1.

Verifies that each of the 10 factory methods emits the exact HTTP status code
and Content-Type mandated by spec/agentic-rest-profile.md §4, and that every
vendor-typed body carries a populated ``trace`` block (spec §5).
"""

from __future__ import annotations

import json
from typing import Any

import pytest

from yaagents_fastapi.response import AgenticResponse

# ── fixtures ──────────────────────────────────────────────────────────────────

CORR = "corr-test-001"
REQ = "req-test-001"


def _body(resp: Any) -> dict[str, Any]:
    """Decode the response body bytes to a dict."""
    assert resp.body is not None
    return json.loads(resp.body)  # type: ignore[no-any-return]


# ── §4 row 1: success ─────────────────────────────────────────────────────────


def test_success_status_and_content_type() -> None:
    r = AgenticResponse.success({"result": "ok"})
    assert r.status_code == 200
    assert r.media_type == "application/json"


def test_success_trace_injected_when_ids_provided() -> None:
    r = AgenticResponse.success({"x": 1}, correlation_id=CORR, request_id=REQ)
    body = _body(r)
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


def test_success_no_trace_when_ids_omitted() -> None:
    r = AgenticResponse.success({"x": 1})
    body = _body(r)
    assert "trace" not in body


# ── §4 row 2: created ─────────────────────────────────────────────────────────


def test_created_status_and_content_type() -> None:
    r = AgenticResponse.created({"id": "new-1"})
    assert r.status_code == 201
    assert r.media_type == "application/json"


def test_created_trace_injected() -> None:
    r = AgenticResponse.created({"id": "new-1"}, correlation_id=CORR, request_id=REQ)
    body = _body(r)
    assert body["trace"]["correlationId"] == CORR
    assert body["trace"]["requestId"] == REQ


# ── §4 row 3: accepted ────────────────────────────────────────────────────────


def test_accepted_status_and_content_type() -> None:
    r = AgenticResponse.accepted(
        "op-1", "/ops/op-1/status", correlation_id=CORR, request_id=REQ
    )
    assert r.status_code == 202
    assert r.media_type == "application/vnd.yaagents.operation+json"


def test_accepted_body_shape() -> None:
    r = AgenticResponse.accepted(
        "op-abc", "/ops/op-abc/status", correlation_id=CORR, request_id=REQ
    )
    body = _body(r)
    assert body["type"] == "operation_accepted"
    assert body["code"] == "OPERATION_ACCEPTED"
    assert body["operationId"] == "op-abc"
    assert body["statusUrl"] == "/ops/op-abc/status"
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


# ── §4 row 4: clarification_required ─────────────────────────────────────────


def test_clarification_required_status_and_content_type() -> None:
    inputs = [{"name": "metric", "location": "body", "type": "string",
               "required": True, "question": "Which metric?"}]
    r = AgenticResponse.clarification_required(
        inputs, correlation_id=CORR, request_id=REQ
    )
    assert r.status_code == 400
    assert r.media_type == "application/vnd.yaagents.clarification+json"


def test_clarification_required_body_shape() -> None:
    inputs = [{"name": "metric", "location": "body", "type": "string",
               "required": True, "question": "Which metric?"}]
    r = AgenticResponse.clarification_required(
        inputs, correlation_id=CORR, request_id=REQ
    )
    body = _body(r)
    assert body["type"] == "clarification_required"
    assert body["code"] == "CLARIFICATION_REQUIRED"
    assert body["requiredInputs"] == inputs
    assert body["trace"]["correlationId"] == CORR


# ── §4 row 5: validation_failed ───────────────────────────────────────────────


def test_validation_failed_status_and_content_type() -> None:
    r = AgenticResponse.validation_failed(
        [{"field": "budget", "message": "Must be > 0"}],
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 422
    assert r.media_type == "application/vnd.yaagents.validation-error+json"


def test_validation_failed_body_shape() -> None:
    errors = [{"field": "budget", "message": "Must be > 0"}]
    r = AgenticResponse.validation_failed(errors, correlation_id=CORR, request_id=REQ)
    body = _body(r)
    assert body["type"] == "validation_failed"
    assert body["code"] == "VALIDATION_FAILED"
    assert body["errors"] == errors
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


# ── §4 row 6: approval_required ───────────────────────────────────────────────


def test_approval_required_status_and_content_type() -> None:
    r = AgenticResponse.approval_required(
        "tok-xyz",
        message="Approval needed.",
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 412
    assert r.media_type == "application/vnd.yaagents.approval-required+json"


def test_approval_required_body_shape() -> None:
    r = AgenticResponse.approval_required(
        "tok-xyz",
        message="Approval needed.",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert body["type"] == "approval_required"
    assert body["approvalToken"] == "tok-xyz"
    assert body["trace"]["requestId"] == REQ


# ── §4 row 7: forbidden ───────────────────────────────────────────────────────


def test_forbidden_status_and_content_type() -> None:
    r = AgenticResponse.forbidden(
        code="PERMISSION_DENIED",
        message="No access.",
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 403
    assert r.media_type == "application/vnd.yaagents.error+json"


def test_forbidden_body_shape() -> None:
    r = AgenticResponse.forbidden(
        code="PERMISSION_DENIED",
        message="No access.",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert body["type"] == "forbidden"
    assert body["code"] == "PERMISSION_DENIED"
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


# ── §4 row 8: conflict ────────────────────────────────────────────────────────


def test_conflict_status_and_content_type() -> None:
    r = AgenticResponse.conflict(
        code="CAMPAIGN_LOCKED",
        message="Locked.",
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 409
    assert r.media_type == "application/vnd.yaagents.conflict+json"


def test_conflict_body_with_resource_id() -> None:
    r = AgenticResponse.conflict(
        code="CAMPAIGN_LOCKED",
        message="Locked.",
        conflicting_resource_id="camp-001",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert body["type"] == "conflict"
    assert body["conflictingResourceId"] == "camp-001"
    assert body["trace"]["correlationId"] == CORR


def test_conflict_body_without_resource_id() -> None:
    r = AgenticResponse.conflict(
        code="STATE_CONFLICT",
        message="Conflict.",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert "conflictingResourceId" not in body


# ── §4 row 9: failed_dependency ───────────────────────────────────────────────


def test_failed_dependency_status_and_content_type() -> None:
    r = AgenticResponse.failed_dependency(
        code="UPSTREAM_UNAVAILABLE",
        message="Upstream down.",
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 424
    assert r.media_type == "application/vnd.yaagents.error+json"


def test_failed_dependency_body_shape() -> None:
    r = AgenticResponse.failed_dependency(
        code="UPSTREAM_UNAVAILABLE",
        message="Upstream down.",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert body["type"] == "failed_dependency"
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


# ── §4 row 10: error ──────────────────────────────────────────────────────────


def test_error_status_and_content_type() -> None:
    r = AgenticResponse.error(
        code="INTERNAL_ERROR",
        message="Something went wrong.",
        correlation_id=CORR,
        request_id=REQ,
    )
    assert r.status_code == 500
    assert r.media_type == "application/vnd.yaagents.error+json"


def test_error_body_shape() -> None:
    r = AgenticResponse.error(
        code="INTERNAL_ERROR",
        message="Something went wrong.",
        correlation_id=CORR,
        request_id=REQ,
    )
    body = _body(r)
    assert body["type"] == "error"
    assert body["code"] == "INTERNAL_ERROR"
    assert body["trace"] == {"correlationId": CORR, "requestId": REQ}


# ── parametrised: all vendor-typed methods have non-empty trace ───────────────

@pytest.mark.parametrize(
    "resp",
    [
        AgenticResponse.accepted(
            "op-1", "/ops/op-1/status", correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.clarification_required(
            [{"name": "x", "location": "body", "type": "string",
              "required": True, "question": "?"}],
            correlation_id=CORR,
            request_id=REQ,
        ),
        AgenticResponse.validation_failed(
            [{"field": "f", "message": "m"}], correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.approval_required(
            "tok", message="msg", correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.forbidden(
            code="DENIED", message="no", correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.conflict(
            code="CON", message="c", correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.failed_dependency(
            code="DEP", message="d", correlation_id=CORR, request_id=REQ
        ),
        AgenticResponse.error(
            code="ERR", message="e", correlation_id=CORR, request_id=REQ
        ),
    ],
)
def test_all_vendor_typed_responses_have_trace(resp: Any) -> None:
    """Every application/vnd.yaagents.* body MUST carry a trace block (spec §5)."""
    body = _body(resp)
    assert "trace" in body, f"Missing trace in {body}"
    assert body["trace"]["correlationId"] == CORR
    assert body["trace"]["requestId"] == REQ
