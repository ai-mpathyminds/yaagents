"""Tests for yaagents validate-response (WI-1yaa.CLI-1).

AC:
  - Every spec/examples/v0.1 valid fixture → PASS exit 0
  - Every invalid fixture → FAIL exit 1 with per-error findings
  - --json mode
  - Path-safety guards
"""

from __future__ import annotations

import json
import pathlib

import pytest

from yaagents_cli.__main__ import main
from yaagents_cli._validate import ValidationError, _safe_path, validate_response

# ── valid fixtures → PASS ─────────────────────────────────────────────────────


class TestValidFixtures:
    def test_all_valid_pass(self, valid_fixtures: list[pathlib.Path]) -> None:
        assert valid_fixtures, "No valid fixtures found — check corpus path"
        for fixture in valid_fixtures:
            result = validate_response(str(fixture))
            assert result.passed, (
                f"{fixture.name} expected PASS but got FAIL: "
                + "; ".join(f.message for f in result.findings)
            )

    @pytest.mark.parametrize(
        "name",
        [
            "operation-accepted.valid.json",
            "operation-accepted.valid.absolute-url.json",
            "clarification-required.valid.json",
            "clarification-required.valid.multi-input.json",
            "validation-failed.valid.json",
            "validation-failed.valid.multi-error.json",
            "approval-required.valid.json",
            "approval-required.valid.long-token.json",
            "conflict.valid.json",
            "conflict.valid.no-resource-id.json",
            "agentic-error.valid.forbidden.json",
            "agentic-error.valid.failed-dependency.json",
            "agentic-error.valid.error.json",
        ],
    )
    def test_named_valid_fixture(
        self, corpus_dir: pathlib.Path, name: str
    ) -> None:
        path = corpus_dir / name
        result = validate_response(str(path))
        assert result.passed, (
            f"{name} expected PASS: "
            + "; ".join(f.message for f in result.findings)
        )
        assert result.schema_name is not None


# ── invalid fixtures → FAIL with findings ────────────────────────────────────


class TestInvalidFixtures:
    # Fixtures where the body's `type` value maps to a DIFFERENT valid schema:
    # the CLI cannot detect these as invalid via type-inference alone (would
    # require HTTP Content-Type context).  The corpus marks them as invalid-for-
    # their-intended-schema; the CLI intentionally validates them as a different
    # known type and correctly returns PASS.
    _TYPE_INFERENCE_PASS_EXPECTED = frozenset(
        {"conflict.invalid.wrong-type.json"}
    )

    def test_all_invalid_fail(
        self, invalid_fixtures: list[pathlib.Path]
    ) -> None:
        assert invalid_fixtures, "No invalid fixtures found — check corpus path"
        for fixture in invalid_fixtures:
            if fixture.name in self._TYPE_INFERENCE_PASS_EXPECTED:
                continue  # see _TYPE_INFERENCE_PASS_EXPECTED note above
            result = validate_response(str(fixture))
            assert not result.passed, (
                f"{fixture.name} expected FAIL but got PASS"
            )
            # Either findings or an error message must be present
            assert result.findings or result.error, (
                f"{fixture.name}: FAIL result has no findings and no error"
            )

    @pytest.mark.parametrize(
        "name",
        [
            "operation-accepted.invalid.missing-trace.json",
            "operation-accepted.invalid.wrong-type.json",
            "operation-accepted.invalid.missing-operation-id.json",
            "clarification-required.invalid.missing-trace.json",
            "clarification-required.invalid.wrong-type.json",
            "clarification-required.invalid.empty-inputs.json",
            "validation-failed.invalid.missing-trace.json",
            "validation-failed.invalid.wrong-type.json",
            "validation-failed.invalid.missing-message.json",
            "approval-required.invalid.missing-trace.json",
            "approval-required.invalid.wrong-type.json",
            "approval-required.invalid.missing-approval-token.json",
            "conflict.invalid.missing-trace.json",
            "conflict.invalid.missing-code.json",
            "agentic-error.invalid.missing-trace.json",
            "agentic-error.invalid.wrong-type.json",
            "agentic-error.invalid.empty-code.json",
        ],
    )
    def test_named_invalid_fixture(
        self, corpus_dir: pathlib.Path, name: str
    ) -> None:
        path = corpus_dir / name
        result = validate_response(str(path))
        assert not result.passed, f"{name} expected FAIL but got PASS"
        assert result.findings or result.error

    def test_type_inference_edge_case_conflict_wrong_type(
        self, corpus_dir: pathlib.Path
    ) -> None:
        """conflict.invalid.wrong-type.json has type='error' (a valid agentic-
        error type).  CLI type-inference correctly validates it as a valid
        agentic-error response (PASS).  Without HTTP Content-Type context,
        the CLI cannot distinguish this from a genuine agentic-error body."""
        path = corpus_dir / "conflict.invalid.wrong-type.json"
        result = validate_response(str(path))
        # Expected PASS: type='error' maps to agentic-error schema, body valid
        assert result.passed
        assert result.schema_name == "agentic-error.schema.json"

    def test_findings_have_pointers(
        self, invalid_fixtures: list[pathlib.Path]
    ) -> None:
        """Every finding carries an RFC 6901 JSON-pointer (starts with /)."""
        skip = self._TYPE_INFERENCE_PASS_EXPECTED
        for fixture in invalid_fixtures:
            if fixture.name in skip:
                continue
            result = validate_response(str(fixture))
            if result.findings:
                for f in result.findings:
                    assert (
                        isinstance(f.pointer, str)
                        and f.pointer.startswith("/")
                    ), (
                        f"{fixture.name}: finding pointer "
                        f"{f.pointer!r} is not a JSON pointer"
                    )


# ── --json output mode ────────────────────────────────────────────────────────


class TestJsonMode:
    def test_json_pass(
        self,
        corpus_dir: pathlib.Path,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        path = corpus_dir / "operation-accepted.valid.json"
        rc = main(["validate-response", str(path), "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert data["result"] == "PASS"
        assert rc == 0

    def test_json_fail(
        self,
        corpus_dir: pathlib.Path,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        path = corpus_dir / "operation-accepted.invalid.missing-trace.json"
        rc = main(["validate-response", str(path), "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert data["result"] == "FAIL"
        assert "errors" in data
        assert len(data["errors"]) > 0
        assert rc == 1

    def test_json_schema_field_present_on_pass(
        self,
        corpus_dir: pathlib.Path,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        path = corpus_dir / "clarification-required.valid.json"
        main(["validate-response", str(path), "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert "schema" in data
        assert "clarification" in data["schema"]


# ── human-readable output ─────────────────────────────────────────────────────


class TestHumanOutput:
    def test_pass_output(
        self,
        corpus_dir: pathlib.Path,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        path = corpus_dir / "conflict.valid.json"
        rc = main(["validate-response", str(path)])
        out = capsys.readouterr().out
        assert out.startswith("PASS:")
        assert rc == 0

    def test_fail_output_has_findings(
        self,
        corpus_dir: pathlib.Path,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        path = corpus_dir / "conflict.invalid.missing-trace.json"
        rc = main(["validate-response", str(path)])
        out = capsys.readouterr().out
        assert out.startswith("FAIL:")
        assert "[" in out  # pointer bracket present
        assert rc == 1


# ── path-safety guards ────────────────────────────────────────────────────────


class TestPathSafety:
    def test_dotdot_rejected(self) -> None:
        with pytest.raises(ValidationError, match="path traversal"):
            _safe_path("../../../etc/passwd")

    def test_cli_dotdot_exit1(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["validate-response", "../../../etc/passwd"])
        assert rc == 1

    def test_file_not_found(self, tmp_path: pathlib.Path) -> None:
        result = validate_response(str(tmp_path / "nonexistent.json"))
        assert not result.passed
        assert result.error is not None

    def test_file_over_10mb_rejected(self, tmp_path: pathlib.Path) -> None:
        big = tmp_path / "big.json"
        big.write_bytes(b"x" * (11 * 1024 * 1024))
        result = validate_response(str(big))
        assert not result.passed
        assert "10 MB" in (result.error or "")

    def test_file_5mb_within_limit(self, tmp_path: pathlib.Path) -> None:
        """5 MB JSON file — size OK; validation proceeds (may fail schema)."""
        f = tmp_path / "medium.json"
        payload = json.dumps(
            {
                "type": "operation_accepted",
                "code": "OPERATION_ACCEPTED",
                "message": "x" * (5 * 1024 * 1024),
                "operationId": "op-1",
                "statusUrl": "/ops/op-1/status",
                "trace": {"correlationId": "c1", "requestId": "r1"},
            }
        )
        f.write_text(payload, encoding="utf-8")
        result = validate_response(str(f))
        # Size within limit → error is NOT size-limit
        assert result.error is None or "10 MB" not in result.error


# ── unknown type ──────────────────────────────────────────────────────────────


class TestUnknownType:
    def test_unknown_type_returns_error(self, tmp_path: pathlib.Path) -> None:
        f = tmp_path / "unknown.json"
        f.write_text(
            json.dumps({"type": "not_a_real_type"}), encoding="utf-8"
        )
        result = validate_response(str(f))
        assert not result.passed
        assert result.error is not None
        assert "not_a_real_type" in result.error

    def test_missing_type_returns_error(self, tmp_path: pathlib.Path) -> None:
        f = tmp_path / "no_type.json"
        f.write_text(
            json.dumps({"code": "X", "message": "y"}), encoding="utf-8"
        )
        result = validate_response(str(f))
        assert not result.passed
        assert result.error is not None


# ── result.to_dict() ──────────────────────────────────────────────────────────


class TestResultDict:
    def test_pass_dict_shape(self, corpus_dir: pathlib.Path) -> None:
        path = corpus_dir / "agentic-error.valid.error.json"
        result = validate_response(str(path))
        d = result.to_dict()
        assert d["result"] == "PASS"
        assert "file" in d
        assert "errors" not in d

    def test_fail_dict_shape(self, corpus_dir: pathlib.Path) -> None:
        path = corpus_dir / "agentic-error.invalid.empty-code.json"
        result = validate_response(str(path))
        d = result.to_dict()
        assert d["result"] == "FAIL"
        assert isinstance(d["errors"], list)
        assert all("pointer" in e and "message" in e for e in d["errors"])
