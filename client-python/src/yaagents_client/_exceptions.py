"""Typed exception hierarchy for the YAAgents Agentic REST Profile v0.1.

Exception classes map the vendor content-types declared in the normative
response table (spec/agentic-rest-profile.md §4) to Python exceptions with
parsed, attribute-rich payloads so callers never have to decode raw dicts.
"""

from __future__ import annotations

from typing import Any

__all__ = [
    "AgenticError",
    "ClarificationRequired",
    "ValidationFailed",
    "FailedDependency",
    "AgenticForbidden",
]


class AgenticError(Exception):
    """Base exception for every non-success agentic REST response.

    All typed exceptions extend this class so callers can write a single
    ``except AgenticError`` catch-all while still branching on the specific
    sub-type for recoverable conditions.

    Attributes:
        status_code: HTTP status code of the response.
        content_type: ``Content-Type`` header value (params stripped).
        code: Service-specific error code from the response body, if present.
        body: Parsed JSON body (dict) for vendor-typed responses, raw text for
              unrecognised content-types, or ``None`` when unavailable.
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        content_type: str,
        code: str | None = None,
        body: dict[str, Any] | str | None = None,
    ) -> None:
        super().__init__(message)
        self.status_code = status_code
        self.content_type = content_type
        self.code = code
        self.body = body


class ClarificationRequired(AgenticError):
    """Raised on HTTP 400 ``application/vnd.yaagents.clarification+json``.

    The agent needs additional inputs before it can proceed.

    Attributes:
        required_inputs: List of input descriptors parsed from the
            ``requiredInputs`` array in the response body (spec §6).
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        content_type: str,
        code: str | None = None,
        body: dict[str, Any],
        required_inputs: list[dict[str, Any]],
    ) -> None:
        super().__init__(
            message,
            status_code=status_code,
            content_type=content_type,
            code=code,
            body=body,
        )
        self.required_inputs = required_inputs


class ValidationFailed(AgenticError):
    """Raised on HTTP 422 ``application/vnd.yaagents.validation-error+json``.

    The request inputs failed schema or business-rule validation.

    Attributes:
        errors: List of ``{field, message}`` dicts from the ``errors`` array.
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        content_type: str,
        code: str | None = None,
        body: dict[str, Any],
        errors: list[dict[str, Any]],
    ) -> None:
        super().__init__(
            message,
            status_code=status_code,
            content_type=content_type,
            code=code,
            body=body,
        )
        self.errors = errors


class FailedDependency(AgenticError):
    """Raised on HTTP 424 ``application/vnd.yaagents.error+json``.

    Discriminated by ``type: "failed_dependency"`` in the response body.

    An upstream dependency (model service, data provider, …) was unavailable.

    Attributes:
        dependency: Full parsed response body as a dict; callers can inspect
            ``dependency["code"]`` / ``dependency["message"]`` for details.
    """

    def __init__(
        self,
        message: str,
        *,
        status_code: int,
        content_type: str,
        code: str | None = None,
        body: dict[str, Any],
        dependency: dict[str, Any],
    ) -> None:
        super().__init__(
            message,
            status_code=status_code,
            content_type=content_type,
            code=code,
            body=body,
        )
        self.dependency = dependency


class AgenticForbidden(AgenticError):
    """Raised on HTTP 403 ``application/vnd.yaagents.error+json`` (type: forbidden).

    The caller lacks permission to perform the requested operation.
    """
