# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Core validation logic for yaagents-cli validate-response."""

from __future__ import annotations

import json
import pathlib
from typing import Any

import jsonschema
import jsonschema.exceptions

# ── schema bundle bundled with the package ───────────────────────────────────
_SCHEMA_DIR = pathlib.Path(__file__).parent / "_schemas" / "v0.1"

# Maps the body `type` discriminator → schema filename.
_TYPE_TO_SCHEMA: dict[str, str] = {
    "operation_accepted": "operation-accepted.schema.json",
    "clarification_required": "clarification-required.schema.json",
    "validation_failed": "validation-failed.schema.json",
    "approval_required": "approval-required.schema.json",
    "conflict": "conflict.schema.json",
    # agentic-error covers three sub-types
    "forbidden": "agentic-error.schema.json",
    "failed_dependency": "agentic-error.schema.json",
    "error": "agentic-error.schema.json",
}

_MAX_FILE_BYTES = 10 * 1024 * 1024  # 10 MB


class ValidationError(Exception):
    """Raised when pre-validation checks fail (path traversal, size, etc.)."""


class FindingEntry:
    """A single schema-validation finding."""

    def __init__(self, pointer: str, message: str) -> None:
        self.pointer = pointer
        self.message = message

    def to_dict(self) -> dict[str, str]:
        return {"pointer": self.pointer, "message": self.message}


class ValidateResult:
    """Outcome of a single validate-response call."""

    def __init__(
        self,
        file: str,
        passed: bool,
        schema_name: str | None = None,
        findings: list[FindingEntry] | None = None,
        error: str | None = None,
    ) -> None:
        self.file = file
        self.passed = passed
        self.schema_name = schema_name
        self.findings: list[FindingEntry] = findings or []
        self.error = error

    def to_dict(self) -> dict[str, Any]:
        d: dict[str, Any] = {
            "file": self.file,
            "result": "PASS" if self.passed else "FAIL",
        }
        if self.schema_name:
            d["schema"] = self.schema_name
        if self.findings:
            d["errors"] = [f.to_dict() for f in self.findings]
        if self.error:
            d["error"] = self.error
        return d


def _safe_path(file_path: str) -> pathlib.Path:
    """Resolve path; reject ``..`` traversal components (NFR-CLI-1).

    Rejects any path component that is ``..`` (prevents directory traversal).
    Absolute paths are permitted — callers pass explicit, fully-qualified
    paths in legitimate use.  The ``..`` guard is the essential protection
    against traversal attacks (e.g. ``../../../etc/passwd``).
    """
    p = pathlib.Path(file_path)
    if ".." in p.parts:
        raise ValidationError("path traversal not allowed")
    return p


def _load_json_file(file_path: str) -> Any:
    """Load JSON from *file_path* with size and path-safety checks."""
    p = _safe_path(file_path)
    try:
        size = p.stat().st_size
    except FileNotFoundError as exc:
        raise ValidationError(f"file not found: {file_path!r}") from exc
    if size > _MAX_FILE_BYTES:
        raise ValidationError("file exceeds 10 MB limit")
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise ValidationError(f"invalid JSON: {exc}") from exc


def _infer_schema(body: Any) -> tuple[str, str]:
    """Return (schema_name, schema_file_path) from the body `type` field."""
    if not isinstance(body, dict):
        raise ValidationError("body must be a JSON object")
    type_val = body.get("type")
    if not isinstance(type_val, str) or not type_val:
        raise ValidationError(
            "body.type is missing or not a string; cannot infer schema"
        )
    schema_file = _TYPE_TO_SCHEMA.get(type_val)
    if schema_file is None:
        known = ", ".join(sorted(_TYPE_TO_SCHEMA))
        raise ValidationError(f"unknown type '{type_val}'; known types: {known}")
    return schema_file, str(_SCHEMA_DIR / schema_file)


def _load_schema(schema_path: str) -> Any:
    return json.loads(pathlib.Path(schema_path).read_text(encoding="utf-8"))


def _make_pointer(abs_path: Any) -> str:
    """Build an RFC 6901 JSON Pointer from a jsonschema absolute_path deque."""
    if not abs_path:
        return "/"
    return "/" + "/".join(str(p) for p in abs_path)


def validate_response(file_path: str) -> ValidateResult:
    """Validate *file_path* against the inferred v0.1 schema.

    Returns a :class:`ValidateResult`; never raises (caller examines `.passed`).
    """
    try:
        body = _load_json_file(file_path)
    except ValidationError as exc:
        return ValidateResult(file=file_path, passed=False, error=str(exc))

    try:
        schema_name, schema_path = _infer_schema(body)
        schema = _load_schema(schema_path)
    except ValidationError as exc:
        return ValidateResult(file=file_path, passed=False, error=str(exc))

    validator = jsonschema.Draft7Validator(schema)
    raw_errors = sorted(validator.iter_errors(body), key=lambda e: e.json_path)

    if not raw_errors:
        return ValidateResult(
            file=file_path,
            passed=True,
            schema_name=schema_name,
        )

    findings = [
        FindingEntry(pointer=_make_pointer(e.absolute_path), message=e.message)
        for e in raw_errors
    ]
    return ValidateResult(
        file=file_path,
        passed=False,
        schema_name=schema_name,
        findings=findings,
    )
