# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""@agentic_operation decorator, AgenticResponses, and AgenticRouter — WI-1yaa.SDK-3.

``@agentic_operation`` decorates an endpoint function with two effects:

1. **Signature injection** — adds
   ``ctx: Annotated[AgenticContext, Depends(AgenticContext)]`` to the
   function's ``__signature__`` if the caller has not declared it, so
   FastAPI resolves the gateway-propagated context automatically.

2. **Metadata storage** — stores two attributes on the decorated function:

   * ``__agentic_openapi_extra__`` — ``{"x-yaagents": {…}}`` dict ready
     for FastAPI's ``openapi_extra=`` route-decorator kwarg.
   * ``__agentic_responses__`` — FastAPI-compatible ``responses=`` dict
     with correct vendor ``Content-Type`` + schema ``$ref`` per declared
     type (status codes and media types from ``spec/agentic-rest-profile.md
     §4``; schema ``$ref`` URIs from ``schemas/v0.1/`` canonical ``$id``
     values per ADR PI1-yaa-0002 §3).

``AgenticRouter`` (wraps ``APIRouter``) provides a convenience ``post()``
method that accepts the agentic metadata kwargs and passes computed
``openapi_extra`` + ``responses`` straight to FastAPI's route registration.

``agentic_route_kwargs(func)`` is a helper for ad-hoc registration via
``app.add_api_route``.
"""

from __future__ import annotations

import inspect
from collections.abc import Callable
from dataclasses import dataclass
from typing import (
    Annotated,
    Any,
    Literal,
    TypeVar,
)

from fastapi import APIRouter, Depends

from yaagents_fastapi.context import AgenticContext

F = TypeVar("F", bound=Callable[..., Any])

OperationKind = Literal[
    "recommendation", "generation", "mutation", "analysis"
]

# Attribute names stored on decorated functions.
_OPENAPI_EXTRA_ATTR = "__agentic_openapi_extra__"
_RESPONSES_ATTR = "__agentic_responses__"


# ---------------------------------------------------------------------------
# AgenticResponses helper
# ---------------------------------------------------------------------------


@dataclass
class AgenticResponses:
    """Declares which response types an agentic endpoint may return.

    Pass to :func:`agentic_operation`.  The selected types drive both
    ``openapi_extra`` generation (vendor ``Content-Type`` + schema ``$ref``)
    and documentation.

    Status codes and ``Content-Type`` values: ``spec/agentic-rest-profile.md
    §4`` (used here as MIME-type identifiers, not table redefinition).
    Schema ``$ref`` URIs: ``schemas/v0.1/`` canonical ``$id`` values
    (ADR PI1-yaa-0002 §3).
    """

    success: bool = False
    created: bool = False
    accepted: bool = False
    clarification_required: bool = False
    validation_failed: bool = False
    approval_required: bool = False
    forbidden: bool = False
    conflict: bool = False
    failed_dependency: bool = False
    error: bool = False

    def to_openapi_responses(self) -> dict[int | str, dict[str, Any]]:
        """Build the FastAPI ``responses=`` dict for all declared types."""
        result: dict[int | str, dict[str, Any]] = {}

        if self.success:
            result[200] = {
                "description": "Success (spec §7.1).",
                "content": {"application/json": {"schema": {}}},
            }

        if self.created:
            result[201] = {
                "description": "Resource created (spec §7.2).",
                "content": {"application/json": {"schema": {}}},
            }

        if self.accepted:
            result[202] = {
                "description": (
                    "Accepted for async processing (spec §7.3)."
                    " Polling runtime is v0.2 scope (ADR PI1-yaa-0002 §4)."
                ),
                "content": {
                    "application/vnd.yaagents.operation+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/operation-accepted.schema.json"
                            )
                        }
                    }
                },
            }

        if self.clarification_required:
            result[400] = {
                "description": (
                    "Agent requires additional inputs (spec §7.4 / §6)."
                ),
                "content": {
                    "application/vnd.yaagents.clarification+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/clarification-required.schema.json"
                            )
                        }
                    }
                },
            }

        if self.forbidden:
            result[403] = {
                "description": "Caller lacks permission (spec §7.7).",
                "content": {
                    "application/vnd.yaagents.error+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/agentic-error.schema.json"
                            )
                        }
                    }
                },
            }

        if self.conflict:
            result[409] = {
                "description": "Conflicting resource state (spec §7.8).",
                "content": {
                    "application/vnd.yaagents.conflict+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/conflict.schema.json"
                            )
                        }
                    }
                },
            }

        if self.approval_required:
            result[412] = {
                "description": "Human approval required (spec §7.6).",
                "content": {
                    "application/vnd.yaagents.approval-required+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/approval-required.schema.json"
                            )
                        }
                    }
                },
            }

        if self.validation_failed:
            result[422] = {
                "description": (
                    "Request inputs failed validation (spec §7.5)."
                ),
                "content": {
                    "application/vnd.yaagents.validation-error+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/validation-failed.schema.json"
                            )
                        }
                    }
                },
            }

        if self.failed_dependency:
            result[424] = {
                "description": "Upstream dependency failed (spec §7.9).",
                "content": {
                    "application/vnd.yaagents.error+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/agentic-error.schema.json"
                            )
                        }
                    }
                },
            }

        if self.error:
            result[500] = {
                "description": "Internal server error (spec §7.10).",
                "content": {
                    "application/vnd.yaagents.error+json": {
                        "schema": {
                            "$ref": (
                                "https://yaagents.io/schemas/v0.1"
                                "/agentic-error.schema.json"
                            )
                        }
                    }
                },
            }

        return result


# ---------------------------------------------------------------------------
# Private helpers
# ---------------------------------------------------------------------------


def _inject_context_param(func: Callable[..., Any]) -> None:
    """Append ``ctx: AgenticContext`` to *func*'s ``__signature__`` if absent.

    FastAPI reads ``__signature__`` for dependency resolution.  This lets
    the caller omit ``ctx`` from the function definition — the decorator
    adds it transparently.
    """
    sig = inspect.signature(func)
    if "ctx" in sig.parameters:
        return  # Already declared; respect the caller's own annotation.
    ctx_param = inspect.Parameter(
        "ctx",
        kind=inspect.Parameter.KEYWORD_ONLY,
        annotation=Annotated[AgenticContext, Depends(AgenticContext)],
    )
    func.__signature__ = sig.replace(  # type: ignore[attr-defined]
        parameters=[*sig.parameters.values(), ctx_param]
    )


def _build_openapi_extra(
    resource: str,
    operation_kind: OperationKind,
    mutating: bool,
) -> dict[str, Any]:
    """Build the ``x-yaagents`` extension block.

    Shape: ``openapi/yaagents-response-profile.yaml §XYaagentsExtension``.
    """
    return {
        "x-yaagents": {
            "resource": resource,
            "operationKind": operation_kind,
            "deterministic": False,
            "mutating": mutating,
        }
    }


# ---------------------------------------------------------------------------
# Public decorator
# ---------------------------------------------------------------------------


def agentic_operation(
    resource: str,
    operation_kind: OperationKind,
    *,
    mutating: bool = False,
    roles: list[str] | None = None,
    responses: AgenticResponses,
) -> Callable[[F], F]:
    """Decorator: inject ``AgenticContext`` + store x-yaagents OpenAPI metadata.

    The decorated function gains two attributes consumed by FastAPI route
    registration helpers::

        @agentic_operation(
            resource="Campaign",
            operation_kind="recommendation",
            mutating=False,
            responses=AgenticResponses(
                clarification_required=True,
                validation_failed=True,
                failed_dependency=True,
            ),
        )
        async def create_optimization() -> Response:
            # ctx is injected by FastAPI via __signature__
            ...

        # Register with FastAPI (two styles):
        app.add_api_route(
            "/path",
            create_optimization,
            methods=["POST"],
            **agentic_route_kwargs(create_optimization),
        )
        # — or — use AgenticRouter (see below)

    ``roles`` is stored for future gateway RBAC enforcement; currently
    informational only.
    """

    def decorator(func: F) -> F:
        _inject_context_param(func)
        setattr(func, _OPENAPI_EXTRA_ATTR, _build_openapi_extra(
            resource, operation_kind, mutating
        ))
        setattr(func, _RESPONSES_ATTR, responses.to_openapi_responses())
        return func

    return decorator


# ---------------------------------------------------------------------------
# Route-kwargs helper
# ---------------------------------------------------------------------------


def agentic_route_kwargs(func: Any) -> dict[str, Any]:
    """Extract FastAPI route kwargs from a function decorated with @agentic_operation.

    Usage::

        app.add_api_route(
            "/path",
            my_endpoint,
            methods=["POST"],
            **agentic_route_kwargs(my_endpoint),
        )
    """
    return {
        "openapi_extra": getattr(func, _OPENAPI_EXTRA_ATTR, {}),
        "responses": getattr(func, _RESPONSES_ATTR, {}),
    }


# ---------------------------------------------------------------------------
# AgenticRouter convenience class
# ---------------------------------------------------------------------------


class AgenticRouter(APIRouter):
    """``APIRouter`` subclass with agentic-aware ``post()`` / ``get()`` etc.

    Passes ``x-yaagents`` metadata and vendor-typed responses to FastAPI's
    route registration in a single decorator call::

        router = AgenticRouter()

        @router.post(
            "/campaigns/{campaign_id}/optimizations",
            resource="Campaign",
            operation_kind="recommendation",
            mutating=False,
            responses=AgenticResponses(
                clarification_required=True,
                validation_failed=True,
            ),
        )
        async def create_optimization(
            ctx: Annotated[AgenticContext, Depends(AgenticContext)],
        ) -> Response:
            ...

        app.include_router(router)
    """

    def post(  # type: ignore[override]
        self,
        path: str,
        *,
        resource: str,
        operation_kind: OperationKind,
        mutating: bool = False,
        roles: list[str] | None = None,
        responses: AgenticResponses,
        **route_kwargs: Any,
    ) -> Callable[[F], F]:
        """Register a POST route with full agentic metadata."""
        return super().post(
            path,
            openapi_extra=_build_openapi_extra(
                resource, operation_kind, mutating
            ),
            responses=responses.to_openapi_responses(),
            **route_kwargs,
        )
