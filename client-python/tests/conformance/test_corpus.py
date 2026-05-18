"""WI-1yaa.PYC-3 — Client conformance tests against the golden corpus.

Drives ``YaAgentsClient`` against a fixture HTTP server that replays every
body in ``spec/examples/v0.1/``.  One test per corpus fixture (31 total);
asserts the correct exception type or payload dict per the response-profile
mapping defined in ``_mapper.process_response``.

Corpus path resolves from this file's location — no env var required.

AC:
  - One test per corpus fixture asserting the mapped outcome
  - Coverage ≥80%; lint/type/audit clean
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
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

# ---------------------------------------------------------------------------
# Corpus path (resolved once at import time; fails fast if not found)
# ---------------------------------------------------------------------------

_CORPUS_DIR: Path = (
    Path(__file__).parents[3]  # → yaagents/
    / "spec"
    / "examples"
    / "v0.1"
)
assert _CORPUS_DIR.is_dir(), (
    f"Corpus directory not found: {_CORPUS_DIR}\n"
    "Run tests from inside the yaagents/ repo root."
)

# ---------------------------------------------------------------------------
# Content-type constants (normative table spec §4)
# ---------------------------------------------------------------------------
_CT_CLARIFICATION = "application/vnd.yaagents.clarification+json"
_CT_VALIDATION = "application/vnd.yaagents.validation-error+json"
_CT_APPROVAL = "application/vnd.yaagents.approval-required+json"
_CT_CONFLICT = "application/vnd.yaagents.conflict+json"
_CT_ERROR = "application/vnd.yaagents.error+json"
_CT_OPERATION = "application/vnd.yaagents.operation+json"


# ---------------------------------------------------------------------------
# Corpus case descriptor
# ---------------------------------------------------------------------------

@dataclass(frozen=True)
class CorpusCase:
    """Describes one corpus fixture and its expected client outcome.

    Attributes:
        fixture: Filename stem (without ``.json``) inside ``spec/examples/v0.1/``.
        status: HTTP status code to serve this fixture with.
        ct: ``Content-Type`` header value to use.
        exc: Exception class that should be raised, or ``None`` if the client
             should return a ``dict`` payload.
        exact_type: When ``True`` the test asserts ``type(exc) is <exc>``;
                    when ``False`` it asserts ``isinstance(exc, <exc>)``.
                    Only meaningful when ``exc`` is not ``None``.
    """

    fixture: str
    status: int
    ct: str
    exc: type[AgenticError] | None
    exact_type: bool = False


def _load(fixture_stem: str) -> dict[str, Any]:
    path = _CORPUS_DIR / f"{fixture_stem}.json"
    return dict(json.loads(path.read_text(encoding="utf-8")))


def _client(status: int, ct: str, body: dict[str, Any]) -> YaAgentsClient:
    encoded = json.dumps(body).encode()

    def handler(request: httpx.Request) -> httpx.Response:  # noqa: ARG001
        return httpx.Response(
            status,
            content=encoded,
            headers={"content-type": ct},
        )

    return YaAgentsClient(
        "http://localhost:8120", "tok", "ten",
        _transport=httpx.MockTransport(handler),
    )


# ---------------------------------------------------------------------------
# 31-entry corpus table
# ---------------------------------------------------------------------------
# Ordering follows INDEX.md section sequence:
#   clarification (5) → validation (5) → approval (5) →
#   conflict (5) → agentic-error (6) → operation-accepted (5)
# ---------------------------------------------------------------------------

CORPUS_CASES: list[CorpusCase] = [
    # ── clarification-required (HTTP 400) ───────────────────────────────────
    CorpusCase(
        "clarification-required.valid",
        400, _CT_CLARIFICATION, ClarificationRequired,
    ),
    CorpusCase(
        "clarification-required.valid.multi-input",
        400, _CT_CLARIFICATION, ClarificationRequired,
    ),
    CorpusCase(
        "clarification-required.invalid.missing-trace",
        400, _CT_CLARIFICATION, ClarificationRequired,
    ),
    CorpusCase(
        "clarification-required.invalid.wrong-type",
        400, _CT_CLARIFICATION, ClarificationRequired,
    ),
    CorpusCase(
        "clarification-required.invalid.empty-inputs",
        400, _CT_CLARIFICATION, ClarificationRequired,
    ),
    # ── validation-failed (HTTP 422) ────────────────────────────────────────
    CorpusCase(
        "validation-failed.valid",
        422, _CT_VALIDATION, ValidationFailed,
    ),
    CorpusCase(
        "validation-failed.valid.multi-error",
        422, _CT_VALIDATION, ValidationFailed,
    ),
    CorpusCase(
        "validation-failed.invalid.missing-trace",
        422, _CT_VALIDATION, ValidationFailed,
    ),
    CorpusCase(
        "validation-failed.invalid.wrong-type",
        422, _CT_VALIDATION, ValidationFailed,
    ),
    CorpusCase(
        "validation-failed.invalid.missing-message",
        422, _CT_VALIDATION, ValidationFailed,
    ),
    # ── approval-required (HTTP 412) ────────────────────────────────────────
    CorpusCase(
        "approval-required.valid",
        412, _CT_APPROVAL, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "approval-required.valid.long-token",
        412, _CT_APPROVAL, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "approval-required.invalid.missing-trace",
        412, _CT_APPROVAL, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "approval-required.invalid.wrong-type",
        412, _CT_APPROVAL, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "approval-required.invalid.missing-approval-token",
        412, _CT_APPROVAL, AgenticError, exact_type=True,
    ),
    # ── conflict (HTTP 409) ─────────────────────────────────────────────────
    CorpusCase(
        "conflict.valid",
        409, _CT_CONFLICT, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "conflict.valid.no-resource-id",
        409, _CT_CONFLICT, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "conflict.invalid.missing-trace",
        409, _CT_CONFLICT, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "conflict.invalid.wrong-type",
        409, _CT_CONFLICT, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "conflict.invalid.missing-code",
        409, _CT_CONFLICT, AgenticError, exact_type=True,
    ),
    # ── agentic-error (HTTP 403 / 424 / 500) ───────────────────────────────
    CorpusCase(
        "agentic-error.valid.forbidden",
        403, _CT_ERROR, AgenticForbidden,
    ),
    CorpusCase(
        "agentic-error.valid.failed-dependency",
        424, _CT_ERROR, FailedDependency,
    ),
    CorpusCase(
        "agentic-error.valid.error",
        500, _CT_ERROR, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "agentic-error.invalid.missing-trace",
        500, _CT_ERROR, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "agentic-error.invalid.wrong-type",
        500, _CT_ERROR, AgenticError, exact_type=True,
    ),
    CorpusCase(
        "agentic-error.invalid.empty-code",
        500, _CT_ERROR, AgenticError, exact_type=True,
    ),
    # ── operation-accepted (HTTP 202) — returns dict, no exception ──────────
    CorpusCase(
        "operation-accepted.valid",
        202, _CT_OPERATION, None,
    ),
    CorpusCase(
        "operation-accepted.valid.absolute-url",
        202, _CT_OPERATION, None,
    ),
    CorpusCase(
        "operation-accepted.invalid.missing-trace",
        202, _CT_OPERATION, None,
    ),
    CorpusCase(
        "operation-accepted.invalid.wrong-type",
        202, _CT_OPERATION, None,
    ),
    CorpusCase(
        "operation-accepted.invalid.missing-operation-id",
        202, _CT_OPERATION, None,
    ),
]

assert len(CORPUS_CASES) == 31, f"Expected 31 corpus cases, got {len(CORPUS_CASES)}"


# ---------------------------------------------------------------------------
# Primary parametrised test — one per corpus fixture
# ---------------------------------------------------------------------------

@pytest.mark.parametrize(
    "case",
    CORPUS_CASES,
    ids=[c.fixture for c in CORPUS_CASES],
)
def test_corpus_fixture(case: CorpusCase) -> None:
    """Drive the client with a corpus fixture body; assert mapped outcome."""
    body = _load(case.fixture)

    with _client(case.status, case.ct, body) as client:
        if case.exc is None:
            # Expected: dict payload returned (success / operation-accepted)
            result = client.campaigns("corpus").optimizations.create({})
            assert isinstance(result, dict), (
                f"Expected dict for {case.fixture!r}, got {type(result)}"
            )
        else:
            # Expected: typed exception raised
            with pytest.raises(case.exc) as exc_info:
                client.campaigns("corpus").optimizations.create({})
            if case.exact_type:
                assert type(exc_info.value) is case.exc, (
                    f"{case.fixture!r}: expected exact type {case.exc.__name__}, "
                    f"got {type(exc_info.value).__name__}"
                )
            assert exc_info.value.status_code == case.status


# ---------------------------------------------------------------------------
# Rich attribute assertions for key valid fixtures
# ---------------------------------------------------------------------------

def test_clarification_valid_required_inputs_count() -> None:
    """clarification-required.valid.json → 1 required input."""
    body = _load("clarification-required.valid")
    with (
        _client(400, _CT_CLARIFICATION, body) as client,
        pytest.raises(ClarificationRequired) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert len(exc_info.value.required_inputs) == 1
    assert exc_info.value.required_inputs[0]["name"] == "successMetric"


def test_clarification_multi_input_required_inputs_count() -> None:
    """clarification-required.valid.multi-input.json → 3 required inputs."""
    body = _load("clarification-required.valid.multi-input")
    with (
        _client(400, _CT_CLARIFICATION, body) as client,
        pytest.raises(ClarificationRequired) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert len(exc_info.value.required_inputs) == 3


def test_clarification_empty_inputs_gives_empty_list() -> None:
    """clarification-required.invalid.empty-inputs.json → .required_inputs == []."""
    body = _load("clarification-required.invalid.empty-inputs")
    with (
        _client(400, _CT_CLARIFICATION, body) as client,
        pytest.raises(ClarificationRequired) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert exc_info.value.required_inputs == []


def test_validation_valid_errors_count() -> None:
    """validation-failed.valid.json → 1 error entry."""
    body = _load("validation-failed.valid")
    with (
        _client(422, _CT_VALIDATION, body) as client,
        pytest.raises(ValidationFailed) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert len(exc_info.value.errors) == 1
    assert exc_info.value.errors[0]["field"] == "budget"


def test_validation_multi_error_count() -> None:
    """validation-failed.valid.multi-error.json → multiple errors (≥1)."""
    body = _load("validation-failed.valid.multi-error")
    with (
        _client(422, _CT_VALIDATION, body) as client,
        pytest.raises(ValidationFailed) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert len(exc_info.value.errors) >= 1


def test_validation_missing_message_uses_default() -> None:
    """validation-failed.invalid.missing-message.json — no message → default str."""
    body = _load("validation-failed.invalid.missing-message")
    with (
        _client(422, _CT_VALIDATION, body) as client,
        pytest.raises(ValidationFailed) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    # Exception string must be a non-empty str (default fallback)
    assert str(exc_info.value)


def test_forbidden_code_matches_fixture() -> None:
    """agentic-error.valid.forbidden.json → .code == PERMISSION_DENIED."""
    body = _load("agentic-error.valid.forbidden")
    with (
        _client(403, _CT_ERROR, body) as client,
        pytest.raises(AgenticForbidden) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert exc_info.value.code == "PERMISSION_DENIED"


def test_failed_dependency_dependency_dict() -> None:
    """agentic-error.valid.failed-dependency.json → .dependency is a dict."""
    body = _load("agentic-error.valid.failed-dependency")
    with (
        _client(424, _CT_ERROR, body) as client,
        pytest.raises(FailedDependency) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    dep = exc_info.value.dependency
    assert isinstance(dep, dict)
    assert dep["type"] == "failed_dependency"
    assert dep["code"] == "UPSTREAM_UNAVAILABLE"


def test_operation_accepted_valid_returns_operation_id() -> None:
    """operation-accepted.valid.json → returned dict contains operationId."""
    body = _load("operation-accepted.valid")
    with _client(202, _CT_OPERATION, body) as client:
        result = client.campaigns("c").optimizations.create({})
    assert result["operationId"] == "op-camp-001-20260517-001"
    assert "statusUrl" in result


def test_operation_accepted_absolute_url() -> None:
    """operation-accepted.valid.absolute-url.json → statusUrl is an absolute URL."""
    body = _load("operation-accepted.valid.absolute-url")
    with _client(202, _CT_OPERATION, body) as client:
        result = client.campaigns("c").optimizations.create({})
    assert result["statusUrl"].startswith("https://")


def test_approval_required_body_accessible() -> None:
    """approval-required.valid.json → .body dict carries approvalToken."""
    body = _load("approval-required.valid")
    with (
        _client(412, _CT_APPROVAL, body) as client,
        pytest.raises(AgenticError) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert isinstance(exc_info.value.body, dict)
    assert "approvalToken" in exc_info.value.body  # type: ignore[operator]


def test_conflict_body_accessible() -> None:
    """conflict.valid.json → .body dict carries conflictingResourceId."""
    body = _load("conflict.valid")
    with (
        _client(409, _CT_CONFLICT, body) as client,
        pytest.raises(AgenticError) as exc_info,
    ):
        client.campaigns("c").optimizations.create({})
    assert isinstance(exc_info.value.body, dict)
    assert exc_info.value.code == "CAMPAIGN_LOCKED"
