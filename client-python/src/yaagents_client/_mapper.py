"""Response mapper: content-type + status → payload dict or typed exception.

This module is the single translation layer between raw ``httpx.Response``
objects and the public exception hierarchy (``_exceptions.py``).  Resource
sub-classes call ``process_response`` immediately after every HTTP exchange.
"""

from __future__ import annotations

from typing import Any, cast

import httpx

from ._exceptions import (
    AgenticError,
    AgenticForbidden,
    ClarificationRequired,
    FailedDependency,
    ValidationFailed,
)

__all__ = ["process_response"]

# ---------------------------------------------------------------------------
# Vendor content-type constants (normative table spec §4 — do not paraphrase)
# ---------------------------------------------------------------------------
_CT_CLARIFICATION = "application/vnd.yaagents.clarification+json"
_CT_VALIDATION = "application/vnd.yaagents.validation-error+json"
_CT_APPROVAL = "application/vnd.yaagents.approval-required+json"
_CT_CONFLICT = "application/vnd.yaagents.conflict+json"
_CT_ERROR = "application/vnd.yaagents.error+json"
_CT_OPERATION = "application/vnd.yaagents.operation+json"


def _content_type(response: httpx.Response) -> str:
    """Return content-type with parameters stripped (``; charset=utf-8`` etc.)."""
    return response.headers.get("content-type", "").split(";")[0].strip()


def _json(response: httpx.Response) -> dict[str, Any]:
    return cast(dict[str, Any], response.json())


def process_response(response: httpx.Response) -> dict[str, Any]:
    """Translate *response* into a payload dict or raise a typed exception.

    Decision tree (first matching rule wins):

    1. **200 / 201** (``application/json``) → return deserialized body.
    2. **202** ``application/vnd.yaagents.operation+json`` → return body.
    3. **clarification+json** → raise :exc:`ClarificationRequired`.
    4. **validation-error+json** → raise :exc:`ValidationFailed`.
    5. **error+json** discriminated by ``type`` field:
       - ``"forbidden"`` → raise :exc:`AgenticForbidden`
       - ``"failed_dependency"`` → raise :exc:`FailedDependency`
       - anything else → raise :exc:`AgenticError`
    6. **approval-required+json / conflict+json** → raise :exc:`AgenticError`
       with parsed body.
    7. Anything else (unknown vendor type or plain non-2xx) →
       raise :exc:`AgenticError` with raw text body.

    Args:
        response: The raw httpx response to inspect.

    Returns:
        Deserialized JSON payload (``dict``) for 200 / 201 / 202 responses.

    Raises:
        ClarificationRequired: HTTP 400 clarification vendor type.
        ValidationFailed: HTTP 422 validation vendor type.
        AgenticForbidden: HTTP 403 error vendor type (``type: forbidden``).
        FailedDependency: HTTP 424 error vendor type (``type: failed_dependency``).
        AgenticError: All other non-success responses (base catch-all).
    """
    ct = _content_type(response)
    status = response.status_code

    # ── Success paths ────────────────────────────────────────────────────────
    if status in (200, 201):
        return _json(response)

    if ct == _CT_OPERATION:
        return _json(response)

    # ── Clarification required ───────────────────────────────────────────────
    if ct == _CT_CLARIFICATION:
        body = _json(response)
        raise ClarificationRequired(
            str(body.get("message", "Clarification required")),
            status_code=status,
            content_type=ct,
            code=body.get("code"),  # type: ignore[arg-type]
            body=body,
            required_inputs=cast(
                list[dict[str, Any]], body.get("requiredInputs", [])
            ),
        )

    # ── Validation failed ────────────────────────────────────────────────────
    if ct == _CT_VALIDATION:
        body = _json(response)
        raise ValidationFailed(
            str(body.get("message", "Validation failed")),
            status_code=status,
            content_type=ct,
            code=body.get("code"),  # type: ignore[arg-type]
            body=body,
            errors=cast(list[dict[str, Any]], body.get("errors", [])),
        )

    # ── Agentic error (forbidden / failed_dependency / error) ────────────────
    if ct == _CT_ERROR:
        body = _json(response)
        error_type = str(body.get("type", "error"))
        code: str | None = body.get("code")  # type: ignore[assignment]
        msg = str(body.get("message", f"Agentic error (HTTP {status})"))

        if error_type == "forbidden":
            raise AgenticForbidden(
                msg,
                status_code=status,
                content_type=ct,
                code=code,
                body=body,
            )
        if error_type == "failed_dependency":
            raise FailedDependency(
                msg,
                status_code=status,
                content_type=ct,
                code=code,
                body=body,
                dependency=body,
            )
        raise AgenticError(
            msg,
            status_code=status,
            content_type=ct,
            code=code,
            body=body,
        )

    # ── Known vendor types without dedicated exception sub-classes ───────────
    if ct in (_CT_APPROVAL, _CT_CONFLICT):
        try:
            parsed = _json(response)
        except Exception as exc:
            raise AgenticError(
                f"Agentic error (HTTP {status})",
                status_code=status,
                content_type=ct,
                body=response.text,
            ) from exc
        raise AgenticError(
            str(parsed.get("message", f"Agentic error (HTTP {status})")),
            status_code=status,
            content_type=ct,
            code=parsed.get("code"),  # type: ignore[arg-type]
            body=parsed,
        )

    # ── Unknown vendor type / plain non-2xx ──────────────────────────────────
    raise AgenticError(
        f"Unexpected response (HTTP {status})",
        status_code=status,
        content_type=ct,
        body=response.text,
    )
