"""Tests for @agentic_operation, AgenticResponses, AgenticRouter — WI-1yaa.SDK-3.

AC:
- Decorated endpoint OpenAPI has x-yaagents + per-response vendor content-types.
- agentic_route_kwargs() returns openapi_extra + responses dicts.
- AgenticRouter.post() registers a route with full agentic metadata.
- AgenticContext is injected (ctx present in __signature__ after decoration).
"""

from __future__ import annotations

import inspect
from typing import Annotated, Any

import pytest
from fastapi import Depends, FastAPI
from fastapi.responses import Response
from fastapi.testclient import TestClient

from yaagents_fastapi.context import AgenticContext
from yaagents_fastapi.decorator import (
    AgenticResponses,
    AgenticRouter,
    agentic_operation,
    agentic_route_kwargs,
)
from yaagents_fastapi.response import AgenticResponse

# ── helpers ───────────────────────────────────────────────────────────────────


def _get_op(openapi: dict[str, Any], method: str, path: str) -> dict[str, Any]:
    """Extract operation dict from generated OpenAPI."""
    return openapi["paths"][path][method]  # type: ignore[no-any-return]


# ── AgenticResponses.to_openapi_responses ─────────────────────────────────────


def test_agentic_responses_empty() -> None:
    ar = AgenticResponses()
    assert ar.to_openapi_responses() == {}


def test_agentic_responses_clarification_required() -> None:
    ar = AgenticResponses(clarification_required=True)
    result = ar.to_openapi_responses()
    assert 400 in result
    content = result[400]["content"]
    assert "application/vnd.yaagents.clarification+json" in content
    ref = content["application/vnd.yaagents.clarification+json"]["schema"]["$ref"]
    assert "clarification-required.schema.json" in ref


def test_agentic_responses_validation_failed() -> None:
    ar = AgenticResponses(validation_failed=True)
    result = ar.to_openapi_responses()
    assert 422 in result
    content = result[422]["content"]
    assert "application/vnd.yaagents.validation-error+json" in content


def test_agentic_responses_accepted() -> None:
    ar = AgenticResponses(accepted=True)
    result = ar.to_openapi_responses()
    assert 202 in result
    content = result[202]["content"]
    assert "application/vnd.yaagents.operation+json" in content
    ref = content["application/vnd.yaagents.operation+json"]["schema"]["$ref"]
    assert "operation-accepted.schema.json" in ref


def test_agentic_responses_error_types_use_agentic_error_schema() -> None:
    """forbidden / failed_dependency / error all reference agentic-error.schema.json."""
    ar = AgenticResponses(forbidden=True, failed_dependency=True, error=True)
    result = ar.to_openapi_responses()
    for code in (403, 424, 500):
        assert code in result
        ct_key = "application/vnd.yaagents.error+json"
        assert ct_key in result[code]["content"]
        assert "agentic-error.schema.json" in (
            result[code]["content"][ct_key]["schema"]["$ref"]
        )


@pytest.mark.parametrize(
    "flag,status,ct_fragment",
    [
        ("approval_required", 412, "approval-required+json"),
        ("conflict", 409, "conflict+json"),
        ("success", 200, "application/json"),
        ("created", 201, "application/json"),
    ],
)
def test_agentic_responses_individual_types(
    flag: str, status: int, ct_fragment: str
) -> None:
    ar = AgenticResponses(**{flag: True})  # type: ignore[arg-type]
    result = ar.to_openapi_responses()
    assert status in result
    assert any(ct_fragment in ct for ct in result[status]["content"])


# ── @agentic_operation decorator ──────────────────────────────────────────────


def test_agentic_operation_stores_openapi_extra() -> None:
    @agentic_operation(
        "Campaign",
        "recommendation",
        responses=AgenticResponses(clarification_required=True),
    )
    async def _endpoint() -> Response:
        return Response()

    extra = getattr(_endpoint, "__agentic_openapi_extra__", None)
    assert extra is not None
    assert "x-yaagents" in extra
    x = extra["x-yaagents"]
    assert x["resource"] == "Campaign"
    assert x["operationKind"] == "recommendation"
    assert x["mutating"] is False
    assert "deterministic" in x


def test_agentic_operation_stores_responses() -> None:
    @agentic_operation(
        "Campaign",
        "mutation",
        mutating=True,
        responses=AgenticResponses(validation_failed=True, error=True),
    )
    async def _endpoint() -> Response:
        return Response()

    stored = getattr(_endpoint, "__agentic_responses__", {})
    assert 422 in stored
    assert 500 in stored


def test_agentic_operation_injects_ctx_into_signature() -> None:
    @agentic_operation(
        "Campaign",
        "generation",
        responses=AgenticResponses(clarification_required=True),
    )
    async def _endpoint() -> Response:
        return Response()

    sig = inspect.signature(_endpoint)
    assert "ctx" in sig.parameters
    ann = sig.parameters["ctx"].annotation
    # Annotated[AgenticContext, Depends(...)]
    assert hasattr(ann, "__metadata__")
    assert ann.__args__[0] is AgenticContext


def test_agentic_operation_respects_existing_ctx() -> None:
    """If ctx is already declared, the decorator leaves it untouched."""

    @agentic_operation(
        "Campaign",
        "analysis",
        responses=AgenticResponses(success=True),
    )
    async def _endpoint(
        ctx: Annotated[AgenticContext, Depends(AgenticContext)],
    ) -> Response:
        return Response()

    sig = inspect.signature(_endpoint)
    params = list(sig.parameters.keys())
    assert params.count("ctx") == 1  # not duplicated


def test_agentic_operation_mutating_flag() -> None:
    @agentic_operation(
        "Asset",
        "mutation",
        mutating=True,
        responses=AgenticResponses(created=True),
    )
    async def _endpoint() -> Response:
        return Response()

    extra = _endpoint.__agentic_openapi_extra__  # type: ignore[attr-defined]
    assert extra["x-yaagents"]["mutating"] is True


# ── agentic_route_kwargs helper ───────────────────────────────────────────────


def test_agentic_route_kwargs_returns_correct_keys() -> None:
    @agentic_operation(
        "Lead",
        "recommendation",
        responses=AgenticResponses(success=True),
    )
    async def _ep() -> Response:
        return Response()

    kwargs = agentic_route_kwargs(_ep)
    assert set(kwargs.keys()) == {"openapi_extra", "responses"}
    assert "x-yaagents" in kwargs["openapi_extra"]


def test_agentic_route_kwargs_on_plain_function() -> None:
    """Falls back gracefully when called on a non-decorated function."""

    async def plain() -> Response:
        return Response()

    kwargs = agentic_route_kwargs(plain)
    assert kwargs["openapi_extra"] == {}
    assert kwargs["responses"] == {}


# ── OpenAPI generation via agentic_route_kwargs ───────────────────────────────


def _make_app_with_route_kwargs() -> FastAPI:
    @agentic_operation(
        "Campaign",
        "recommendation",
        responses=AgenticResponses(
            clarification_required=True, validation_failed=True
        ),
    )
    async def create_optimization() -> Response:
        return AgenticResponse.success({"ok": True})

    app = FastAPI()
    app.add_api_route(
        "/campaigns/{campaign_id}/optimizations",
        create_optimization,
        methods=["POST"],
        **agentic_route_kwargs(create_optimization),
    )
    return app


def test_openapi_has_x_yaagents_via_route_kwargs() -> None:
    app = _make_app_with_route_kwargs()
    openapi = app.openapi()
    op = _get_op(openapi, "post", "/campaigns/{campaign_id}/optimizations")
    assert "x-yaagents" in op
    assert op["x-yaagents"]["resource"] == "Campaign"
    assert op["x-yaagents"]["operationKind"] == "recommendation"


def test_openapi_has_vendor_content_types_via_route_kwargs() -> None:
    app = _make_app_with_route_kwargs()
    openapi = app.openapi()
    op = _get_op(openapi, "post", "/campaigns/{campaign_id}/optimizations")
    responses = op["responses"]
    assert "400" in responses
    assert "application/vnd.yaagents.clarification+json" in (
        responses["400"]["content"]
    )
    assert "422" in responses
    assert "application/vnd.yaagents.validation-error+json" in (
        responses["422"]["content"]
    )


# ── OpenAPI generation via AgenticRouter ──────────────────────────────────────


def _make_app_with_agentic_router() -> FastAPI:
    router = AgenticRouter()

    @router.post(
        "/campaigns/{campaign_id}/optimizations",
        resource="Campaign",
        operation_kind="recommendation",
        mutating=False,
        responses=AgenticResponses(
            clarification_required=True,
            validation_failed=True,
            failed_dependency=True,
        ),
    )
    async def create_optimization(
        ctx: Annotated[AgenticContext, Depends(AgenticContext)],
    ) -> Response:
        return AgenticResponse.clarification_required(
            [{"name": "metric", "location": "body", "type": "string",
              "required": True, "question": "?"}],
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )

    app = FastAPI()
    app.include_router(router)
    return app


def test_agentic_router_openapi_has_x_yaagents() -> None:
    app = _make_app_with_agentic_router()
    openapi = app.openapi()
    op = _get_op(openapi, "post", "/campaigns/{campaign_id}/optimizations")
    assert "x-yaagents" in op
    assert op["x-yaagents"]["resource"] == "Campaign"
    assert op["x-yaagents"]["operationKind"] == "recommendation"
    assert op["x-yaagents"]["mutating"] is False


def test_agentic_router_openapi_has_vendor_content_types() -> None:
    app = _make_app_with_agentic_router()
    openapi = app.openapi()
    op = _get_op(openapi, "post", "/campaigns/{campaign_id}/optimizations")
    responses = op["responses"]
    assert "400" in responses
    assert "application/vnd.yaagents.clarification+json" in (
        responses["400"]["content"]
    )
    assert "422" in responses
    assert "application/vnd.yaagents.validation-error+json" in (
        responses["422"]["content"]
    )
    assert "424" in responses
    assert "application/vnd.yaagents.error+json" in responses["424"]["content"]


def test_agentic_router_endpoint_executes() -> None:
    """The endpoint registered via AgenticRouter is callable end-to-end."""
    app = _make_app_with_agentic_router()
    client = TestClient(app)
    resp = client.post(
        "/campaigns/camp-001/optimizations",
        headers={
            "X-Correlation-ID": "corr-sdk3",
            "X-Request-ID": "req-sdk3",
        },
    )
    assert resp.status_code == 400
    body = resp.json()
    assert body["type"] == "clarification_required"
    assert body["trace"]["correlationId"] == "corr-sdk3"
