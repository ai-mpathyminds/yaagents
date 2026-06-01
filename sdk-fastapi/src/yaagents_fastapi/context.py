# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""AgenticContext and RequiredInput — WI-1yaa.SDK-2.

AgenticContext
--------------
FastAPI class-based dependency that extracts the four gateway-injected headers
(gateway source: ``gateway/internal/tenant/tenant.go``):

  X-Tenant-ID      → :attr:`tenant_id`
  X-Actor-Subject  → :attr:`actor_id`
  X-Correlation-ID → :attr:`correlation_id`  (spec §5.1; UUID v4 generated if absent)
  X-Request-ID     → :attr:`request_id`      (spec §5.1; UUID v4 generated if absent)

``correlation_id`` and ``request_id`` are generated when absent so the SDK
works standalone without the gateway (spec §5.1 gateway-propagation rule).

RequiredInput
-------------
Dataclass for one element of the spec §6 ``requiredInputs[]`` array, used to
feed :meth:`~yaagents_fastapi.response.AgenticResponse.clarification_required`.
"""

from __future__ import annotations

import uuid
from dataclasses import dataclass
from typing import Annotated, Any, Literal

from fastapi import Header


class AgenticContext:
    """Gateway-propagated request context, injectable via FastAPI ``Depends``.

    **Important**: do NOT sub-class ``pydantic.BaseSettings`` — that silently
    sets ``embed=True`` on JSON route bodies.  This class is a plain Python
    class; FastAPI resolves it via ``__init__`` parameter introspection.

    Usage::

        from typing import Annotated
        from fastapi import Depends
        from fastapi.responses import Response
        from yaagents_fastapi import AgenticContext, AgenticResponse

        @app.post("/campaigns/{id}/optimizations")
        def optimize(
            id: str,
            ctx: Annotated[AgenticContext, Depends(AgenticContext)],
        ) -> Response:
            return AgenticResponse.accepted(
                operation_id="op-1",
                status_url=f"/campaigns/{id}/ops/op-1/status",
                correlation_id=ctx.correlation_id,
                request_id=ctx.request_id,
            )
    """

    # Class-level annotations let mypy know the instance attribute types.
    tenant_id: str
    actor_id: str
    correlation_id: str
    request_id: str

    def __init__(
        self,
        x_tenant_id: Annotated[str, Header()] = "",
        x_actor_subject: Annotated[str, Header()] = "",
        x_correlation_id: Annotated[str, Header()] = "",
        x_request_id: Annotated[str, Header()] = "",
    ) -> None:
        """Populate context from request headers.

        FastAPI converts underscores to hyphens (``convert_underscores=True``)
        so parameter names map to headers as follows:

        ================  ====================
        Parameter name    HTTP header matched
        ================  ====================
        x_tenant_id       X-Tenant-ID
        x_actor_subject   X-Actor-Subject
        x_correlation_id  X-Correlation-ID
        x_request_id      X-Request-ID
        ================  ====================
        """
        self.tenant_id = x_tenant_id
        self.actor_id = x_actor_subject
        # Generate UUID v4 when the gateway header is absent — standalone use.
        self.correlation_id = x_correlation_id or str(uuid.uuid4())
        self.request_id = x_request_id or str(uuid.uuid4())


@dataclass
class RequiredInput:
    """One element of the spec §6 ``requiredInputs[]`` array.

    Fields match the spec §6 table exactly.  ``type`` uses the spec name;
    ``allowed_values`` is optional (omit for open-ended inputs).

    To feed :meth:`AgenticResponse.clarification_required`, convert a list
    of instances with ``[ri.to_dict() for ri in inputs]``::

        inputs = [
            RequiredInput(
                name="successMetric",
                location="body",
                type="string",
                required=True,
                question="Which success metric should be optimized?",
                allowed_values=["ctr", "cpl", "conversion_rate"],
            )
        ]
        return AgenticResponse.clarification_required(
            [ri.to_dict() for ri in inputs],
            correlation_id=ctx.correlation_id,
            request_id=ctx.request_id,
        )
    """

    name: str
    location: Literal["body", "query", "path", "header"]
    type: str  # JSON type hint: string | integer | boolean | array | object
    required: bool
    question: str
    allowed_values: list[str] | None = None

    def to_dict(self) -> dict[str, Any]:
        """Serialise to spec §6 wire shape (camelCase keys).

        ``allowedValues`` is omitted when :attr:`allowed_values` is ``None``
        (open-ended input — spec §6: "omit if open-ended").
        """
        d: dict[str, Any] = {
            "name": self.name,
            "location": self.location,
            "type": self.type,
            "required": self.required,
            "question": self.question,
        }
        if self.allowed_values is not None:
            d["allowedValues"] = self.allowed_values
        return d
