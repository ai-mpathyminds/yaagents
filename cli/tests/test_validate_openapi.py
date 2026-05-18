"""Tests for yaagents validate-openapi (WI-1yaa.CLI-2).

AC:
  - Passes on openapi/yaagents-response-profile.yaml
  - Fails on a fixture with wrong Content-Type
  - Detects missing x-yaagents
  - Detects dangling $ref
  - --json mode
"""

from __future__ import annotations

import json
import pathlib
import textwrap

import pytest
import yaml

from yaagents_cli.__main__ import main
from yaagents_cli._validate_openapi import validate_openapi

# Path to the canonical openapi file (two dirs up from cli/)
_OPENAPI_DIR = (
    pathlib.Path(__file__).parent.parent.parent / "openapi"
)
_PROFILE_YAML = _OPENAPI_DIR / "yaagents-response-profile.yaml"
_SCHEMAS_DIR = (
    pathlib.Path(__file__).parent.parent.parent / "schemas" / "v0.1"
)


def _write_yaml(tmp_path: pathlib.Path, name: str, content: str) -> str:
    """Write *content* as YAML to *tmp_path/name* and return the path string."""
    p = tmp_path / name
    p.write_text(textwrap.dedent(content), encoding="utf-8")
    return str(p)


def _minimal_agentic_op(
    status: str,
    content_type: str,
    ref: str = "../schemas/v0.1/clarification-required.schema.json",
    include_x_yaagents: bool = True,
    operationKind: str = "recommendation",
) -> dict:  # type: ignore[type-arg]
    """Build a minimal OpenAPI 3.1 dict with one agentic POST operation."""
    op: dict = {  # type: ignore[type-arg]
        "operationId": "testOp",
        "responses": {
            status: {
                "description": "test",
                "content": {
                    content_type: {
                        "schema": {"$ref": ref},
                    }
                },
            }
        },
    }
    if include_x_yaagents:
        op["x-yaagents"] = {
            "resource": "Test",
            "operationKind": operationKind,
            "deterministic": False,
            "mutating": False,
        }
    return {
        "openapi": "3.1.0",
        "info": {"title": "Test", "version": "0.1"},
        "paths": {"/test": {"post": op}},
    }


# ── golden fixture: openapi/yaagents-response-profile.yaml → PASS ────────────


class TestGoldenFixture:
    def test_profile_yaml_passes(self) -> None:
        assert _PROFILE_YAML.exists(), (
            f"openapi/yaagents-response-profile.yaml not found at {_PROFILE_YAML}"
        )
        result = validate_openapi(str(_PROFILE_YAML))
        assert result.passed, (
            "yaagents-response-profile.yaml expected PASS but got FAIL:\n"
            + "\n".join(
                f"  [{f.pointer}] {f.message}" for f in result.findings
            )
        )

    def test_profile_yaml_pass_exit0(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["validate-openapi", str(_PROFILE_YAML)])
        out = capsys.readouterr().out
        assert rc == 0
        assert out.startswith("PASS:")


# ── check (b): wrong Content-Type → FAIL ─────────────────────────────────────


class TestWrongContentType:
    def test_wrong_ct_for_400(self, tmp_path: pathlib.Path) -> None:
        """HTTP 400 should use clarification+json, not error+json."""
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.error+json",
            ref=str(_SCHEMAS_DIR / "agentic-error.schema.json"),
        )
        f = tmp_path / "wrong_ct.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any("§4 table" in fi.message for fi in result.findings), (
            "Expected §4-table violation finding"
        )

    def test_wrong_ct_for_422(self, tmp_path: pathlib.Path) -> None:
        """HTTP 422 should use validation-error+json, not clarification+json."""
        doc = _minimal_agentic_op(
            status="422",
            content_type="application/vnd.yaagents.clarification+json",
            ref=str(_SCHEMAS_DIR / "clarification-required.schema.json"),
        )
        f = tmp_path / "wrong_ct_422.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any("§4 table" in fi.message for fi in result.findings)

    def test_correct_ct_202(self, tmp_path: pathlib.Path) -> None:
        """HTTP 202 with correct vendor CT should produce no CT finding."""
        doc = _minimal_agentic_op(
            status="202",
            content_type="application/vnd.yaagents.operation+json",
            ref=str(_SCHEMAS_DIR / "operation-accepted.schema.json"),
        )
        f = tmp_path / "correct_202.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        # No CT violation (may still pass overall)
        ct_findings = [
            fi for fi in result.findings if "§4 table" in fi.message
        ]
        assert not ct_findings

    @pytest.mark.parametrize(
        "status,ct",
        [
            ("202", "application/vnd.yaagents.operation+json"),
            ("400", "application/vnd.yaagents.clarification+json"),
            ("412", "application/vnd.yaagents.approval-required+json"),
            ("422", "application/vnd.yaagents.validation-error+json"),
            ("403", "application/vnd.yaagents.error+json"),
            ("409", "application/vnd.yaagents.conflict+json"),
            ("424", "application/vnd.yaagents.error+json"),
            ("500", "application/vnd.yaagents.error+json"),
        ],
    )
    def test_correct_ct_no_violation(
        self, tmp_path: pathlib.Path, status: str, ct: str
    ) -> None:
        """Every §4-correct (status, CT) pair produces no CT finding."""
        # Use a real schema file for the $ref
        _vnd = "application/vnd.yaagents."
        schema_map = {
            _vnd + "operation+json": "operation-accepted.schema.json",
            _vnd + "clarification+json": "clarification-required.schema.json",
            _vnd + "approval-required+json": "approval-required.schema.json",
            _vnd + "validation-error+json": "validation-failed.schema.json",
            _vnd + "error+json": "agentic-error.schema.json",
            _vnd + "conflict+json": "conflict.schema.json",
        }
        schema_file = schema_map[ct]
        doc = _minimal_agentic_op(
            status=status,
            content_type=ct,
            ref=str(_SCHEMAS_DIR / schema_file),
        )
        f = tmp_path / f"correct_{status}.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        ct_findings = [
            fi for fi in result.findings if "§4 table" in fi.message
        ]
        assert not ct_findings, (
            f"Unexpected CT finding for HTTP {status} + {ct}: "
            + "; ".join(fi.message for fi in ct_findings)
        )


# ── check (a): missing x-yaagents → FAIL ─────────────────────────────────────


class TestMissingXYaagents:
    def test_missing_x_yaagents_detected(self, tmp_path: pathlib.Path) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.clarification+json",
            ref=str(_SCHEMAS_DIR / "clarification-required.schema.json"),
            include_x_yaagents=False,
        )
        f = tmp_path / "no_x_yaagents.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any(
            "x-yaagents extension is required" in fi.message
            for fi in result.findings
        )

    def test_malformed_operationKind(self, tmp_path: pathlib.Path) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.clarification+json",
            ref=str(_SCHEMAS_DIR / "clarification-required.schema.json"),
            operationKind="not_a_real_kind",
        )
        f = tmp_path / "bad_op_kind.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any(
            "operationKind" in fi.message for fi in result.findings
        )

    def test_missing_required_field_in_x_yaagents(
        self, tmp_path: pathlib.Path
    ) -> None:
        """x-yaagents present but missing 'resource' field."""
        doc: dict = {  # type: ignore[type-arg]
            "openapi": "3.1.0",
            "info": {"title": "T", "version": "0.1"},
            "paths": {
                "/t": {
                    "post": {
                        "operationId": "t",
                        "x-yaagents": {
                            # 'resource' missing
                            "operationKind": "generation",
                            "deterministic": True,
                            "mutating": True,
                        },
                        "responses": {
                            "400": {
                                "description": "c",
                                "content": {
                                    "application/vnd.yaagents.clarification+json": {
                                        "schema": {
                                            "$ref": str(
                                                _SCHEMAS_DIR
                                                / "clarification-required.schema.json"
                                            )
                                        }
                                    }
                                },
                            }
                        },
                    }
                }
            },
        }
        f = tmp_path / "missing_resource.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any(
            "resource" in fi.message for fi in result.findings
        )


# ── check (c): dangling $ref → FAIL ──────────────────────────────────────────


class TestDanglingRef:
    def test_dangling_ref_detected(self, tmp_path: pathlib.Path) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.clarification+json",
            ref="../schemas/v0.1/does-not-exist.schema.json",
        )
        f = tmp_path / "dangling.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert any("dangling $ref" in fi.message for fi in result.findings)

    def test_valid_ref_no_finding(self, tmp_path: pathlib.Path) -> None:
        """A $ref that resolves to an existing file → no dangling finding."""
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.clarification+json",
            ref=str(_SCHEMAS_DIR / "clarification-required.schema.json"),
        )
        f = tmp_path / "good_ref.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        dangling = [
            fi for fi in result.findings if "dangling" in fi.message
        ]
        assert not dangling

    def test_fragment_ref_not_flagged(self, tmp_path: pathlib.Path) -> None:
        """Internal #/... refs are never flagged as dangling."""
        doc: dict = {  # type: ignore[type-arg]
            "openapi": "3.1.0",
            "info": {"title": "T", "version": "0.1"},
            "components": {
                "schemas": {"Foo": {"type": "object"}},
            },
            "paths": {
                "/t": {
                    "post": {
                        "operationId": "t",
                        "x-yaagents": {
                            "resource": "Foo",
                            "operationKind": "analysis",
                            "deterministic": True,
                            "mutating": False,
                        },
                        "responses": {
                            "400": {
                                "description": "c",
                                "content": {
                                    "application/vnd.yaagents.clarification+json": {
                                        "schema": {
                                            "$ref": "#/components/schemas/Foo"
                                        }
                                    }
                                },
                            }
                        },
                    }
                }
            },
        }
        f = tmp_path / "fragment_ref.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        dangling = [
            fi for fi in result.findings if "dangling" in fi.message
        ]
        assert not dangling


# ── --json output mode ────────────────────────────────────────────────────────


class TestJsonMode:
    def test_json_pass(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["validate-openapi", str(_PROFILE_YAML), "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert data["result"] == "PASS"
        assert rc == 0

    def test_json_fail(
        self, tmp_path: pathlib.Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.error+json",
            ref=str(_SCHEMAS_DIR / "agentic-error.schema.json"),
        )
        f = tmp_path / "fail.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        rc = main(["validate-openapi", str(f), "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert data["result"] == "FAIL"
        assert "errors" in data
        assert rc == 1


# ── path-safety + error paths ─────────────────────────────────────────────────


class TestErrorPaths:
    def test_file_not_found(self, tmp_path: pathlib.Path) -> None:
        result = validate_openapi(str(tmp_path / "nonexistent.yaml"))
        assert not result.passed
        assert result.error is not None

    def test_dotdot_rejected(self) -> None:
        result = validate_openapi("../../../etc/hosts")
        assert not result.passed
        assert result.error is not None
        assert "path traversal" in result.error

    def test_non_dict_root(self, tmp_path: pathlib.Path) -> None:
        f = tmp_path / "scalar.yaml"
        f.write_text("just a string\n", encoding="utf-8")
        result = validate_openapi(str(f))
        assert not result.passed
        assert result.error is not None

    def test_human_output_pass(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["validate-openapi", str(_PROFILE_YAML)])
        out = capsys.readouterr().out
        assert rc == 0
        assert out.startswith("PASS:")

    def test_human_output_fail(
        self, tmp_path: pathlib.Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.error+json",
            ref=str(_SCHEMAS_DIR / "agentic-error.schema.json"),
        )
        f = tmp_path / "fail_human.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        rc = main(["validate-openapi", str(f)])
        out = capsys.readouterr().out
        assert rc == 1
        assert out.startswith("FAIL:")
        assert "[" in out  # pointer present


# ── to_dict() shape ───────────────────────────────────────────────────────────


class TestResultDict:
    def test_pass_dict(self) -> None:
        result = validate_openapi(str(_PROFILE_YAML))
        d = result.to_dict()
        assert d["result"] == "PASS"
        assert "file" in d
        assert "errors" not in d

    def test_fail_dict(self, tmp_path: pathlib.Path) -> None:
        doc = _minimal_agentic_op(
            status="400",
            content_type="application/vnd.yaagents.error+json",
            ref=str(_SCHEMAS_DIR / "agentic-error.schema.json"),
        )
        f = tmp_path / "fail_dict.yaml"
        f.write_text(yaml.dump(doc), encoding="utf-8")
        result = validate_openapi(str(f))
        d = result.to_dict()
        assert d["result"] == "FAIL"
        assert isinstance(d["errors"], list)
        assert all("pointer" in e and "message" in e for e in d["errors"])
