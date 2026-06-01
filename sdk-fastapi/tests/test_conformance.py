# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""WI-1yaa.SDK-4: Schema-conformance tests.

One parametrised group per vendor media type:
  1. Corpus valid fixtures   → schema MUST validate (PASS)
  2. Corpus invalid fixtures → schema MUST reject   (FAIL)
  3. SDK factory output      → schema MUST validate (PASS)

Schemas:  schemas/v0.1/*.schema.json  (JSON Schema Draft-07)
Corpus:   spec/examples/v0.1/*.json
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import pytest
from jsonschema import Draft7Validator
from jsonschema.exceptions import ValidationError

from yaagents_fastapi.response import AgenticResponse

# ── path anchors ─────────────────────────────────────────────────────────────
_REPO = Path(__file__).parent.parent.parent  # yaagents/
_SCHEMAS = _REPO / "schemas" / "v0.1"
_CORPUS = _REPO / "spec" / "examples" / "v0.1"


# ── helpers ───────────────────────────────────────────────────────────────────


def _load_schema(name: str) -> dict[str, Any]:
    return json.loads((_SCHEMAS / f"{name}.schema.json").read_text())


def _load_fixture(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text())


def _sdk_body(response: Any) -> dict[str, Any]:
    """Extract the JSON body dict from an AgenticResponse factory return value."""
    return json.loads(response.body)


def _validator(schema_name: str) -> Draft7Validator:
    return Draft7Validator(_load_schema(schema_name))


def _corpus_valid(schema_name: str) -> list[Path]:
    return sorted(_CORPUS.glob(f"{schema_name}.valid.*.json"))


def _corpus_invalid(schema_name: str) -> list[Path]:
    return sorted(_CORPUS.glob(f"{schema_name}.invalid.*.json"))


# ── 1. clarification-required ─────────────────────────────────────────────────

_CLARIF_SCHEMA = "clarification-required"


@pytest.mark.parametrize("fixture", _corpus_valid(_CLARIF_SCHEMA), ids=lambda p: p.name)
def test_clarification_required_corpus_valid(fixture: Path) -> None:
    _validator(_CLARIF_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_CLARIF_SCHEMA), ids=lambda p: p.name
)
def test_clarification_required_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_CLARIF_SCHEMA).validate(_load_fixture(fixture))


def test_clarification_required_sdk_output() -> None:
    resp = AgenticResponse.clarification_required(
        required_inputs=[
            {
                "name": "successMetric",
                "location": "body",
                "type": "string",
                "required": True,
                "question": "Which metric?",
                "allowedValues": ["ctr", "cpl"],
            }
        ],
        message="Please clarify.",
        correlation_id="c-001",
        request_id="r-001",
    )
    _validator(_CLARIF_SCHEMA).validate(_sdk_body(resp))


# ── 2. validation-failed ──────────────────────────────────────────────────────

_VALFAIL_SCHEMA = "validation-failed"


@pytest.mark.parametrize(
    "fixture", _corpus_valid(_VALFAIL_SCHEMA), ids=lambda p: p.name
)
def test_validation_failed_corpus_valid(fixture: Path) -> None:
    _validator(_VALFAIL_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_VALFAIL_SCHEMA), ids=lambda p: p.name
)
def test_validation_failed_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_VALFAIL_SCHEMA).validate(_load_fixture(fixture))


def test_validation_failed_sdk_output() -> None:
    resp = AgenticResponse.validation_failed(
        errors=[{"field": "budget", "message": "Must be > 0."}],
        correlation_id="c-002",
        request_id="r-002",
    )
    _validator(_VALFAIL_SCHEMA).validate(_sdk_body(resp))


# ── 3. approval-required ──────────────────────────────────────────────────────

_APPROVAL_SCHEMA = "approval-required"


@pytest.mark.parametrize(
    "fixture", _corpus_valid(_APPROVAL_SCHEMA), ids=lambda p: p.name
)
def test_approval_required_corpus_valid(fixture: Path) -> None:
    _validator(_APPROVAL_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_APPROVAL_SCHEMA), ids=lambda p: p.name
)
def test_approval_required_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_APPROVAL_SCHEMA).validate(_load_fixture(fixture))


def test_approval_required_sdk_output() -> None:
    resp = AgenticResponse.approval_required(
        approval_token="tok-abc123",
        message="Manager approval required.",
        correlation_id="c-003",
        request_id="r-003",
    )
    _validator(_APPROVAL_SCHEMA).validate(_sdk_body(resp))


# ── 4. conflict ───────────────────────────────────────────────────────────────

_CONFLICT_SCHEMA = "conflict"


@pytest.mark.parametrize(
    "fixture", _corpus_valid(_CONFLICT_SCHEMA), ids=lambda p: p.name
)
def test_conflict_corpus_valid(fixture: Path) -> None:
    _validator(_CONFLICT_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_CONFLICT_SCHEMA), ids=lambda p: p.name
)
def test_conflict_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_CONFLICT_SCHEMA).validate(_load_fixture(fixture))


def test_conflict_sdk_output_with_resource_id() -> None:
    resp = AgenticResponse.conflict(
        code="RESOURCE_ALREADY_EXISTS",
        message="Campaign already exists.",
        conflicting_resource_id="camp-007",
        correlation_id="c-004",
        request_id="r-004",
    )
    _validator(_CONFLICT_SCHEMA).validate(_sdk_body(resp))


def test_conflict_sdk_output_no_resource_id() -> None:
    resp = AgenticResponse.conflict(
        code="RESOURCE_ALREADY_EXISTS",
        message="Campaign already exists.",
        correlation_id="c-005",
        request_id="r-005",
    )
    body = _sdk_body(resp)
    assert "conflictingResourceId" not in body
    _validator(_CONFLICT_SCHEMA).validate(body)


# ── 5. agentic-error (forbidden / failed_dependency / error) ──────────────────

_ERROR_SCHEMA = "agentic-error"


@pytest.mark.parametrize("fixture", _corpus_valid(_ERROR_SCHEMA), ids=lambda p: p.name)
def test_agentic_error_corpus_valid(fixture: Path) -> None:
    _validator(_ERROR_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_ERROR_SCHEMA), ids=lambda p: p.name
)
def test_agentic_error_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_ERROR_SCHEMA).validate(_load_fixture(fixture))


def test_forbidden_sdk_output() -> None:
    resp = AgenticResponse.forbidden(
        code="PERMISSION_DENIED",
        message="No permission.",
        correlation_id="c-006",
        request_id="r-006",
    )
    _validator(_ERROR_SCHEMA).validate(_sdk_body(resp))


def test_failed_dependency_sdk_output() -> None:
    resp = AgenticResponse.failed_dependency(
        code="UPSTREAM_UNAVAILABLE",
        message="Upstream is down.",
        correlation_id="c-007",
        request_id="r-007",
    )
    _validator(_ERROR_SCHEMA).validate(_sdk_body(resp))


def test_error_sdk_output() -> None:
    resp = AgenticResponse.error(
        code="INTERNAL_ERROR",
        message="Unexpected failure.",
        correlation_id="c-008",
        request_id="r-008",
    )
    _validator(_ERROR_SCHEMA).validate(_sdk_body(resp))


# ── 6. operation-accepted ─────────────────────────────────────────────────────

_ACCEPTED_SCHEMA = "operation-accepted"


@pytest.mark.parametrize(
    "fixture", _corpus_valid(_ACCEPTED_SCHEMA), ids=lambda p: p.name
)
def test_operation_accepted_corpus_valid(fixture: Path) -> None:
    _validator(_ACCEPTED_SCHEMA).validate(_load_fixture(fixture))


@pytest.mark.parametrize(
    "fixture", _corpus_invalid(_ACCEPTED_SCHEMA), ids=lambda p: p.name
)
def test_operation_accepted_corpus_invalid(fixture: Path) -> None:
    with pytest.raises(ValidationError):
        _validator(_ACCEPTED_SCHEMA).validate(_load_fixture(fixture))


def test_accepted_sdk_output() -> None:
    resp = AgenticResponse.accepted(
        operation_id="op-camp-001-20260517",
        status_url="/operations/op-camp-001-20260517/status",
        correlation_id="c-009",
        request_id="r-009",
    )
    _validator(_ACCEPTED_SCHEMA).validate(_sdk_body(resp))
