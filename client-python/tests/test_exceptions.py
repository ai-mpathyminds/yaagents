# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for WI-1yaa.PYC-2 — typed exception mapping.

AC:
  - Each mandatory error media type raises its specific exception with parsed attributes
  - ``created`` returns payload object; base ``AgenticError`` catches all

Fixtures mirror spec/examples/v0.1/ golden corpus bodies.
"""

from __future__ import annotations

import json
from typing import Any

import httpx
import pytest

from yaagents_client import (
    AgenticError,
    AgenticForbidden,
    ClarificationRequired,
    FailedDependency,
    ValidationFailed,
    YaAgentsClient,
)
from yaagents_client._mapper import process_response

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_CT_CLARIFICATION = "application/vnd.yaagents.clarification+json"
_CT_VALIDATION = "application/vnd.yaagents.validation-error+json"
_CT_APPROVAL = "application/vnd.yaagents.approval-required+json"
_CT_CONFLICT = "application/vnd.yaagents.conflict+json"
_CT_ERROR = "application/vnd.yaagents.error+json"
_CT_OPERATION = "application/vnd.yaagents.operation+json"


def _resp(
    status: int,
    content_type: str,
    body: dict[str, Any] | str,
) -> httpx.Response:
    """Build a synthetic httpx.Response for process_response tests."""
    if isinstance(body, dict):
        content = json.dumps(body).encode()
        ct_header = content_type
    else:
        content = body.encode()
        ct_header = content_type
    return httpx.Response(
        status,
        content=content,
        headers={"content-type": ct_header},
    )


def _client_with_response(
    status: int,
    content_type: str,
    body: dict[str, Any] | str,
) -> YaAgentsClient:
    """Return a YaAgentsClient whose transport always returns *body*."""
    resp_body = json.dumps(body).encode() if isinstance(body, dict) else body.encode()

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(
            status,
            content=resp_body,
            headers={"content-type": content_type},
        )

    transport = httpx.MockTransport(handler)
    return YaAgentsClient(
        "http://localhost:8120", "tok", "ten", _transport=transport
    )


# ---------------------------------------------------------------------------
# process_response — success cases
# ---------------------------------------------------------------------------


def test_200_returns_payload() -> None:
    """HTTP 200 application/json → dict payload returned."""
    payload = {"id": "opt-1", "status": "running"}
    resp = _resp(200, "application/json", payload)
    result = process_response(resp)
    assert result == payload


def test_201_returns_payload() -> None:
    """HTTP 201 application/json → dict payload returned (created)."""
    payload = {"id": "res-42", "created": True}
    resp = _resp(201, "application/json", payload)
    result = process_response(resp)
    assert result == payload


def test_201_content_type_with_charset_returns_payload() -> None:
    """Charset param in content-type is stripped correctly."""
    payload = {"ok": True}
    resp = _resp(201, "application/json; charset=utf-8", payload)
    assert process_response(resp) == payload


def test_202_operation_accepted_returns_payload() -> None:
    """HTTP 202 operation+json → operation dict returned, no exception."""
    payload = {
        "type": "operation_accepted",
        "code": "OPERATION_ACCEPTED",
        "operationId": "op-001",
        "statusUrl": "/operations/op-001/status",
        "trace": {"correlationId": "c1", "requestId": "r1"},
    }
    resp = _resp(202, _CT_OPERATION, payload)
    result = process_response(resp)
    assert result["operationId"] == "op-001"


# ---------------------------------------------------------------------------
# process_response — ClarificationRequired
# ---------------------------------------------------------------------------

_CLARIFICATION_BODY: dict[str, Any] = {
    "type": "clarification_required",
    "code": "CLARIFICATION_REQUIRED",
    "message": "Additional information is required.",
    "requiredInputs": [
        {
            "name": "successMetric",
            "location": "body",
            "type": "string",
            "required": True,
            "question": "Which success metric should be optimized?",
            "allowedValues": ["ctr", "cpl"],
        }
    ],
    "trace": {"correlationId": "corr-001", "requestId": "req-001"},
}


def test_clarification_raises_correct_type() -> None:
    """HTTP 400 clarification+json raises ClarificationRequired."""
    resp = _resp(400, _CT_CLARIFICATION, _CLARIFICATION_BODY)
    with pytest.raises(ClarificationRequired):
        process_response(resp)


def test_clarification_required_inputs_parsed() -> None:
    """.required_inputs contains the parsed requiredInputs list."""
    resp = _resp(400, _CT_CLARIFICATION, _CLARIFICATION_BODY)
    with pytest.raises(ClarificationRequired) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert len(err.required_inputs) == 1
    assert err.required_inputs[0]["name"] == "successMetric"


def test_clarification_attributes() -> None:
    """ClarificationRequired carries status_code, content_type, code, body."""
    resp = _resp(400, _CT_CLARIFICATION, _CLARIFICATION_BODY)
    with pytest.raises(ClarificationRequired) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 400
    assert err.content_type == _CT_CLARIFICATION
    assert err.code == "CLARIFICATION_REQUIRED"
    assert isinstance(err.body, dict)


def test_clarification_is_agentic_error() -> None:
    """ClarificationRequired is a subclass of AgenticError."""
    resp = _resp(400, _CT_CLARIFICATION, _CLARIFICATION_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    assert isinstance(exc_info.value, ClarificationRequired)


def test_clarification_multi_input() -> None:
    """Multiple requiredInputs are all preserved."""
    body: dict[str, Any] = {
        **_CLARIFICATION_BODY,
        "requiredInputs": [
            {
                "name": "a", "location": "body", "type": "string",
                "required": True, "question": "q1",
            },
            {
                "name": "b", "location": "query", "type": "integer",
                "required": False, "question": "q2",
            },
        ],
    }
    resp = _resp(400, _CT_CLARIFICATION, body)
    with pytest.raises(ClarificationRequired) as exc_info:
        process_response(resp)
    assert len(exc_info.value.required_inputs) == 2


# ---------------------------------------------------------------------------
# process_response — ValidationFailed
# ---------------------------------------------------------------------------

_VALIDATION_BODY: dict[str, Any] = {
    "type": "validation_failed",
    "code": "VALIDATION_FAILED",
    "message": "The request inputs failed validation.",
    "errors": [{"field": "budget", "message": "Budget must be greater than 0."}],
    "trace": {"correlationId": "corr-101", "requestId": "req-101"},
}


def test_validation_raises_correct_type() -> None:
    """HTTP 422 validation-error+json raises ValidationFailed."""
    resp = _resp(422, _CT_VALIDATION, _VALIDATION_BODY)
    with pytest.raises(ValidationFailed):
        process_response(resp)


def test_validation_errors_parsed() -> None:
    """.errors contains the parsed errors list."""
    resp = _resp(422, _CT_VALIDATION, _VALIDATION_BODY)
    with pytest.raises(ValidationFailed) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert len(err.errors) == 1
    assert err.errors[0]["field"] == "budget"


def test_validation_attributes() -> None:
    """ValidationFailed carries expected attributes."""
    resp = _resp(422, _CT_VALIDATION, _VALIDATION_BODY)
    with pytest.raises(ValidationFailed) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 422
    assert err.content_type == _CT_VALIDATION
    assert err.code == "VALIDATION_FAILED"


def test_validation_is_agentic_error() -> None:
    """ValidationFailed is caught by ``except AgenticError``."""
    resp = _resp(422, _CT_VALIDATION, _VALIDATION_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    assert isinstance(exc_info.value, ValidationFailed)


def test_validation_multi_error() -> None:
    """Multiple validation errors are all preserved in .errors."""
    body: dict[str, Any] = {
        **_VALIDATION_BODY,
        "errors": [
            {"field": "budget", "message": "must be > 0"},
            {"field": "name", "message": "required"},
        ],
    }
    resp = _resp(422, _CT_VALIDATION, body)
    with pytest.raises(ValidationFailed) as exc_info:
        process_response(resp)
    assert len(exc_info.value.errors) == 2


# ---------------------------------------------------------------------------
# process_response — AgenticForbidden (403)
# ---------------------------------------------------------------------------

_FORBIDDEN_BODY: dict[str, Any] = {
    "type": "forbidden",
    "code": "PERMISSION_DENIED",
    "message": "You do not have permission to create optimizations for this campaign.",
    "trace": {"correlationId": "corr-401", "requestId": "req-401"},
}


def test_forbidden_raises_agentic_forbidden() -> None:
    """HTTP 403 error+json (type: forbidden) raises AgenticForbidden."""
    resp = _resp(403, _CT_ERROR, _FORBIDDEN_BODY)
    with pytest.raises(AgenticForbidden):
        process_response(resp)


def test_forbidden_attributes() -> None:
    """AgenticForbidden carries status_code, content_type, code."""
    resp = _resp(403, _CT_ERROR, _FORBIDDEN_BODY)
    with pytest.raises(AgenticForbidden) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 403
    assert err.content_type == _CT_ERROR
    assert err.code == "PERMISSION_DENIED"


def test_forbidden_is_agentic_error() -> None:
    """AgenticForbidden is caught by ``except AgenticError``."""
    resp = _resp(403, _CT_ERROR, _FORBIDDEN_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    assert isinstance(exc_info.value, AgenticForbidden)


# ---------------------------------------------------------------------------
# process_response — FailedDependency (424)
# ---------------------------------------------------------------------------

_FAILED_DEP_BODY: dict[str, Any] = {
    "type": "failed_dependency",
    "code": "UPSTREAM_UNAVAILABLE",
    "message": "The AI model service is temporarily unavailable.",
    "trace": {"correlationId": "corr-402", "requestId": "req-402"},
}


def test_failed_dependency_raises_correct_type() -> None:
    """HTTP 424 error+json (type: failed_dependency) raises FailedDependency."""
    resp = _resp(424, _CT_ERROR, _FAILED_DEP_BODY)
    with pytest.raises(FailedDependency):
        process_response(resp)


def test_failed_dependency_parsed() -> None:
    """.dependency holds the full parsed body dict."""
    resp = _resp(424, _CT_ERROR, _FAILED_DEP_BODY)
    with pytest.raises(FailedDependency) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.dependency["code"] == "UPSTREAM_UNAVAILABLE"
    assert err.dependency["type"] == "failed_dependency"


def test_failed_dependency_attributes() -> None:
    """FailedDependency carries status_code, content_type, code."""
    resp = _resp(424, _CT_ERROR, _FAILED_DEP_BODY)
    with pytest.raises(FailedDependency) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 424
    assert err.code == "UPSTREAM_UNAVAILABLE"


def test_failed_dependency_is_agentic_error() -> None:
    """FailedDependency is caught by ``except AgenticError``."""
    resp = _resp(424, _CT_ERROR, _FAILED_DEP_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    assert isinstance(exc_info.value, FailedDependency)


# ---------------------------------------------------------------------------
# process_response — AgenticError base (500 error)
# ---------------------------------------------------------------------------

_ERROR_BODY: dict[str, Any] = {
    "type": "error",
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred while processing your request.",
    "trace": {"correlationId": "corr-403", "requestId": "req-403"},
}


def test_500_error_raises_agentic_error() -> None:
    """HTTP 500 error+json (type: error) raises base AgenticError."""
    resp = _resp(500, _CT_ERROR, _ERROR_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    # Must NOT be a sub-class (it IS the base class here)
    assert type(exc_info.value) is AgenticError


def test_500_error_attributes() -> None:
    """Base AgenticError from 500 carries expected attributes."""
    resp = _resp(500, _CT_ERROR, _ERROR_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 500
    assert err.code == "INTERNAL_ERROR"


# ---------------------------------------------------------------------------
# process_response — approval-required (412)
# ---------------------------------------------------------------------------

_APPROVAL_BODY: dict[str, Any] = {
    "type": "approval_required",
    "code": "APPROVAL_REQUIRED",
    "message": "Human approval required.",
    "approvalToken": "apptok-abc123",
    "trace": {"correlationId": "corr-201", "requestId": "req-201"},
}


def test_approval_required_raises_agentic_error() -> None:
    """HTTP 412 approval-required+json raises AgenticError (no dedicated class)."""
    resp = _resp(412, _CT_APPROVAL, _APPROVAL_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 412
    assert err.content_type == _CT_APPROVAL
    assert err.code == "APPROVAL_REQUIRED"
    assert isinstance(err.body, dict)


# ---------------------------------------------------------------------------
# process_response — conflict (409)
# ---------------------------------------------------------------------------

_CONFLICT_BODY: dict[str, Any] = {
    "type": "conflict",
    "code": "CAMPAIGN_LOCKED",
    "message": "Campaign is currently locked.",
    "conflictingResourceId": "camp-001",
    "trace": {"correlationId": "corr-301", "requestId": "req-301"},
}


def test_conflict_raises_agentic_error() -> None:
    """HTTP 409 conflict+json raises AgenticError with parsed body."""
    resp = _resp(409, _CT_CONFLICT, _CONFLICT_BODY)
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 409
    assert err.content_type == _CT_CONFLICT
    assert err.code == "CAMPAIGN_LOCKED"


# ---------------------------------------------------------------------------
# process_response — unknown vendor type → AgenticError with raw body
# ---------------------------------------------------------------------------


def test_unknown_content_type_raises_agentic_error_with_raw_body() -> None:
    """Unrecognised content-type raises AgenticError carrying raw text body."""
    resp = _resp(418, "text/plain", "I'm a teapot")
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    err = exc_info.value
    assert err.status_code == 418
    assert isinstance(err.body, str)


def test_plain_json_non_success_raises_agentic_error() -> None:
    """Plain application/json with 4xx status raises AgenticError (unknown vendor)."""
    resp = _resp(404, "application/json", {"detail": "not found"})
    with pytest.raises(AgenticError) as exc_info:
        process_response(resp)
    assert exc_info.value.status_code == 404


# ---------------------------------------------------------------------------
# Exception hierarchy — isinstance checks
# ---------------------------------------------------------------------------


def test_all_subtypes_are_agentic_error() -> None:
    """Every typed exception is a subclass of AgenticError."""
    assert issubclass(ClarificationRequired, AgenticError)
    assert issubclass(ValidationFailed, AgenticError)
    assert issubclass(FailedDependency, AgenticError)
    assert issubclass(AgenticForbidden, AgenticError)


def test_all_subtypes_exported_from_package() -> None:
    """All exception classes are accessible from the top-level package."""
    import yaagents_client

    for name in ("AgenticError", "AgenticForbidden", "ClarificationRequired",
                 "FailedDependency", "ValidationFailed"):
        assert hasattr(yaagents_client, name), f"missing export: {name}"


# ---------------------------------------------------------------------------
# End-to-end through resource accessors
# ---------------------------------------------------------------------------


def test_resource_200_returns_dict() -> None:
    """optimizations.create() on 200 returns deserialized dict (not httpx.Response)."""
    payload = {"id": "opt-1", "status": "running"}
    with _client_with_response(200, "application/json", payload) as client:
        result = client.campaigns("c1").optimizations.create({})
    assert result == payload


def test_resource_201_returns_dict() -> None:
    """assets.generate() on 201 returns deserialized dict."""
    payload = {"id": "asset-99"}
    with _client_with_response(201, "application/json", payload) as client:
        result = client.campaigns("c1").assets.generate({})
    assert result == payload


def test_resource_raises_clarification_required() -> None:
    """ClarificationRequired propagates from resource accessor."""
    with (
        _client_with_response(400, _CT_CLARIFICATION, _CLARIFICATION_BODY) as client,
        pytest.raises(ClarificationRequired) as exc_info,
    ):
        client.campaigns("c1").optimizations.create({})
    assert len(exc_info.value.required_inputs) == 1


def test_resource_raises_validation_failed() -> None:
    """ValidationFailed propagates from resource accessor."""
    with (
        _client_with_response(422, _CT_VALIDATION, _VALIDATION_BODY) as client,
        pytest.raises(ValidationFailed) as exc_info,
    ):
        client.campaigns("c1").optimizations.create({})
    assert exc_info.value.errors[0]["field"] == "budget"


def test_resource_raises_forbidden() -> None:
    """AgenticForbidden propagates from resource accessor."""
    with (
        _client_with_response(403, _CT_ERROR, _FORBIDDEN_BODY) as client,
        pytest.raises(AgenticForbidden),
    ):
        client.campaigns("c1").assets.generate({})


def test_resource_raises_failed_dependency() -> None:
    """FailedDependency propagates from resource accessor."""
    with (
        _client_with_response(424, _CT_ERROR, _FAILED_DEP_BODY) as client,
        pytest.raises(FailedDependency) as exc_info,
    ):
        client.campaigns("c1").optimizations.create({})
    assert exc_info.value.dependency["code"] == "UPSTREAM_UNAVAILABLE"


def test_resource_base_agentic_error_catches_all() -> None:
    """``except AgenticError`` catches every error from resource accessors."""
    with (
        _client_with_response(500, _CT_ERROR, _ERROR_BODY) as client,
        pytest.raises(AgenticError),
    ):
        client.campaigns("c1").optimizations.create({})
