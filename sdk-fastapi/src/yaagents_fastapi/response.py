"""AgenticResponse — factory for YAAgents Agentic REST Profile v0.1 typed responses.

Status codes and Content-Type values are taken verbatim from the normative table in
``spec/agentic-rest-profile.md §4`` (ADR PI1-yaa-0002 §1 — sole authoritative source).
This file does **not** redefine or paraphrase that table; every media-type string used
below is a Content-Type header value (a reference, not a definition).

Profile version: ``spec/VERSION`` = ``0.1``
"""

from __future__ import annotations

import json
from typing import Any

from fastapi.responses import Response


def _build(body: dict[str, Any], *, status_code: int, media_type: str) -> Response:
    """Serialise *body* to UTF-8 JSON bytes and wrap in a FastAPI Response."""
    return Response(
        content=json.dumps(body, default=str).encode(),
        status_code=status_code,
        media_type=media_type,
    )


def _trace(correlation_id: str, request_id: str) -> dict[str, str]:
    """Build the mandatory ``trace`` block (spec §5)."""
    return {"correlationId": correlation_id, "requestId": request_id}


class AgenticResponse:
    """Factory class for the 10 YAAgents Profile v0.1 response types.

    Each class method returns a :class:`fastapi.responses.Response` with the
    exact HTTP status code and ``Content-Type`` mandated by the normative table
    in ``spec/agentic-rest-profile.md §4``.

    The ``trace`` block (``correlationId`` + ``requestId``) is injected into
    every vendor-typed response body (spec §5).  For ``success``/``created``
    (``application/json``), ``trace`` is **recommended** and injected only when
    the caller supplies non-empty IDs.
    """

    # ── §4 row 1: success ──────────────────────────────────────────────────

    @classmethod
    def success(
        cls,
        body: dict[str, Any],
        *,
        correlation_id: str = "",
        request_id: str = "",
    ) -> Response:
        """200 ``application/json`` — spec §4 row 1 / §7.1.

        Body shape is domain-defined.  ``trace`` is injected when IDs are
        supplied (spec §5: recommended, not mandatory for plain JSON types).
        """
        if correlation_id or request_id:
            body = {**body, "trace": _trace(correlation_id, request_id)}
        return _build(body, status_code=200, media_type="application/json")

    # ── §4 row 2: created ──────────────────────────────────────────────────

    @classmethod
    def created(
        cls,
        body: dict[str, Any],
        *,
        correlation_id: str = "",
        request_id: str = "",
    ) -> Response:
        """201 ``application/json`` — spec §4 row 2 / §7.2.

        Body shape is domain-defined.  ``trace`` is injected when IDs are
        supplied (spec §5: recommended, not mandatory for plain JSON types).
        """
        if correlation_id or request_id:
            body = {**body, "trace": _trace(correlation_id, request_id)}
        return _build(body, status_code=201, media_type="application/json")

    # ── §4 row 3: accepted ─────────────────────────────────────────────────

    @classmethod
    def accepted(
        cls,
        operation_id: str,
        status_url: str,
        *,
        message: str = "Operation accepted for asynchronous processing.",
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """202 ``application/vnd.yaagents.operation+json`` — spec §4 row 3 / §7.3.

        Schema ships in PI1-yaa; async polling runtime is v0.2 scope
        (ADR PI1-yaa-0002 §4 — no polling endpoint built here).
        """
        body: dict[str, Any] = {
            "type": "operation_accepted",
            "code": "OPERATION_ACCEPTED",
            "message": message,
            "operationId": operation_id,
            "statusUrl": status_url,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=202,
            media_type="application/vnd.yaagents.operation+json",
        )

    # ── §4 row 4: clarification_required ──────────────────────────────────

    @classmethod
    def clarification_required(
        cls,
        required_inputs: list[dict[str, Any]],
        *,
        message: str = "Additional information is required.",
        code: str = "CLARIFICATION_REQUIRED",
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """400 ``application/vnd.yaagents.clarification+json`` — spec §4 row 4.

        See also spec §4.1 and §6.
        ``required_inputs`` MUST contain at least one element (spec §6).
        """
        body: dict[str, Any] = {
            "type": "clarification_required",
            "code": code,
            "message": message,
            "requiredInputs": required_inputs,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=400,
            media_type="application/vnd.yaagents.clarification+json",
        )

    # ── §4 row 5: validation_failed ────────────────────────────────────────

    @classmethod
    def validation_failed(
        cls,
        errors: list[dict[str, Any]],
        *,
        message: str = "The request inputs failed validation.",
        code: str = "VALIDATION_FAILED",
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """422 ``application/vnd.yaagents.validation-error+json`` — spec §4 / §7.5.
        """
        body: dict[str, Any] = {
            "type": "validation_failed",
            "code": code,
            "message": message,
            "errors": errors,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=422,
            media_type="application/vnd.yaagents.validation-error+json",
        )

    # ── §4 row 6: approval_required ────────────────────────────────────────

    @classmethod
    def approval_required(
        cls,
        approval_token: str,
        *,
        message: str,
        code: str = "APPROVAL_REQUIRED",
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """412 ``application/vnd.yaagents.approval-required+json`` — spec §4 / §7.6.
        """
        body: dict[str, Any] = {
            "type": "approval_required",
            "code": code,
            "message": message,
            "approvalToken": approval_token,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=412,
            media_type="application/vnd.yaagents.approval-required+json",
        )

    # ── §4 row 7: forbidden ────────────────────────────────────────────────

    @classmethod
    def forbidden(
        cls,
        *,
        code: str,
        message: str,
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """403 ``application/vnd.yaagents.error+json`` — spec §4 row 7 / §7.7."""
        body: dict[str, Any] = {
            "type": "forbidden",
            "code": code,
            "message": message,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=403,
            media_type="application/vnd.yaagents.error+json",
        )

    # ── §4 row 8: conflict ─────────────────────────────────────────────────

    @classmethod
    def conflict(
        cls,
        *,
        code: str,
        message: str,
        conflicting_resource_id: str | None = None,
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """409 ``application/vnd.yaagents.conflict+json`` — spec §4 row 8 / §7.8."""
        body: dict[str, Any] = {
            "type": "conflict",
            "code": code,
            "message": message,
            "trace": _trace(correlation_id, request_id),
        }
        if conflicting_resource_id is not None:
            body["conflictingResourceId"] = conflicting_resource_id
        return _build(
            body,
            status_code=409,
            media_type="application/vnd.yaagents.conflict+json",
        )

    # ── §4 row 9: failed_dependency ────────────────────────────────────────

    @classmethod
    def failed_dependency(
        cls,
        *,
        code: str,
        message: str,
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """424 ``application/vnd.yaagents.error+json`` — spec §4 row 9 / §7.9."""
        body: dict[str, Any] = {
            "type": "failed_dependency",
            "code": code,
            "message": message,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=424,
            media_type="application/vnd.yaagents.error+json",
        )

    # ── §4 row 10: error ───────────────────────────────────────────────────

    @classmethod
    def error(
        cls,
        *,
        code: str,
        message: str,
        correlation_id: str,
        request_id: str,
    ) -> Response:
        """500 ``application/vnd.yaagents.error+json`` — spec §4 row 10 / §7.10."""
        body: dict[str, Any] = {
            "type": "error",
            "code": code,
            "message": message,
            "trace": _trace(correlation_id, request_id),
        }
        return _build(
            body,
            status_code=500,
            media_type="application/vnd.yaagents.error+json",
        )
