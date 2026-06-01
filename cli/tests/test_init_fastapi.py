# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for `yaagents init fastapi` scaffold command (WI-1yaa.CLI-3)."""

from __future__ import annotations

import io
import json
import pathlib
from contextlib import redirect_stdout

import pytest

from yaagents_cli.__main__ import main
from yaagents_cli._init_fastapi import init_fastapi
from yaagents_cli._validate_openapi import validate_openapi

# ── scaffold content ──────────────────────────────────────────────────────────


class TestScaffoldCreatesFiles:
    def test_creates_four_files(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        assert result.succeeded
        target = pathlib.Path(result.target_dir)
        for name in ("main.py", "pyproject.toml", "routes.yaml", "openapi.yaml"):
            assert (target / name).exists(), f"{name} missing"

    def test_created_list_matches_files(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        assert set(result.created) == {
            "main.py",
            "pyproject.toml",
            "routes.yaml",
            "openapi.yaml",
        }

    def test_main_py_imports_yaagents_fastapi(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "main.py").read_text()
        assert "from yaagents_fastapi import" in content
        assert "AgenticRouter" in content
        assert "AgenticResponse" in content
        assert "AgenticResponses" in content

    def test_main_py_uses_agentic_operation(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "main.py").read_text()
        assert "@router.post" in content
        assert "resource=" in content
        assert "operation_kind=" in content

    def test_pyproject_depends_on_yaagents_fastapi(
        self, tmp_path: pathlib.Path
    ) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "pyproject.toml").read_text()
        assert "yaagents-fastapi" in content
        assert "fastapi" in content
        assert "uvicorn" in content

    def test_routes_yaml_has_x_yaagents(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "routes.yaml").read_text()
        assert "x-yaagents" in content
        assert "routes:" in content

    def test_creates_nested_target_dir(self, tmp_path: pathlib.Path) -> None:
        target = tmp_path / "nested" / "deep" / "app"
        result = init_fastapi(str(target))
        assert result.succeeded
        assert target.exists()


# ── idempotency guard ─────────────────────────────────────────────────────────


class TestIdempotencyGuard:
    def test_refuses_non_empty_dir(self, tmp_path: pathlib.Path) -> None:
        (tmp_path / "existing.txt").write_text("data")
        result = init_fastapi(str(tmp_path))
        assert not result.succeeded
        assert result.error is not None
        assert "non-empty" in result.error
        assert "--force" in result.error

    def test_force_overwrites_non_empty_dir(self, tmp_path: pathlib.Path) -> None:
        (tmp_path / "existing.txt").write_text("data")
        result = init_fastapi(str(tmp_path), force=True)
        assert result.succeeded

    def test_empty_existing_dir_accepted_without_force(
        self, tmp_path: pathlib.Path
    ) -> None:
        target = tmp_path / "empty"
        target.mkdir()
        result = init_fastapi(str(target))
        assert result.succeeded

    def test_nonexistent_dir_accepted_without_force(
        self, tmp_path: pathlib.Path
    ) -> None:
        result = init_fastapi(str(tmp_path / "new"))
        assert result.succeeded

    def test_force_flag_ignored_when_dir_empty(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "clean"), force=True)
        assert result.succeeded

    def test_path_traversal_rejected(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi("../unsafe")
        assert not result.succeeded
        assert result.error is not None
        assert "traversal" in result.error


# ── generated openapi.yaml passes validate-openapi ────────────────────────────


class TestGeneratedOpenapiPassesValidate:
    def test_openapi_yaml_passes_validate_openapi(
        self, tmp_path: pathlib.Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        # chdir so the absolute tmp_path is within CWD (NFR-CLI-1 guard).
        monkeypatch.chdir(tmp_path)
        result = init_fastapi("app")
        assert result.succeeded
        vr = validate_openapi("app/openapi.yaml")
        assert vr.passed, (
            "Generated openapi.yaml FAILED validate-openapi:\n"
            + "\n".join(f"  [{f.pointer}] {f.message}" for f in vr.findings)
        )

    def test_openapi_yaml_has_x_yaagents(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "openapi.yaml").read_text()
        assert "x-yaagents" in content
        assert "resource:" in content
        assert "operationKind:" in content

    def test_openapi_yaml_has_vendor_content_types(
        self, tmp_path: pathlib.Path
    ) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        content = (pathlib.Path(result.target_dir) / "openapi.yaml").read_text()
        assert "application/vnd.yaagents.operation+json" in content
        assert "application/vnd.yaagents.clarification+json" in content
        assert "application/vnd.yaagents.validation-error+json" in content


# ── InitResult.to_dict ────────────────────────────────────────────────────────


class TestInitResultDict:
    def test_succeeded_result_dict(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        d = result.to_dict()
        assert d["result"] == "OK"
        assert "created" in d
        assert isinstance(d["created"], list)

    def test_failed_result_dict(self, tmp_path: pathlib.Path) -> None:
        (tmp_path / "blocker.txt").write_text("x")
        result = init_fastapi(str(tmp_path))
        d = result.to_dict()
        assert d["result"] == "ERROR"
        assert "error" in d
        assert "created" not in d

    def test_target_dir_in_dict(self, tmp_path: pathlib.Path) -> None:
        result = init_fastapi(str(tmp_path / "app"))
        assert "target_dir" in result.to_dict()


# ── CLI integration ───────────────────────────────────────────────────────────


class TestCliInit:
    def test_cli_init_fastapi_exits_zero(self, tmp_path: pathlib.Path) -> None:
        rc = main(["init", "fastapi", str(tmp_path / "app")])
        assert rc == 0

    def test_cli_init_creates_files(self, tmp_path: pathlib.Path) -> None:
        target = tmp_path / "app"
        main(["init", "fastapi", str(target)])
        assert (target / "main.py").exists()
        assert (target / "openapi.yaml").exists()

    def test_cli_init_json_mode(self, tmp_path: pathlib.Path) -> None:
        buf = io.StringIO()
        with redirect_stdout(buf):
            rc = main(["init", "fastapi", str(tmp_path / "app"), "--json"])
        assert rc == 0
        d = json.loads(buf.getvalue())
        assert d["result"] == "OK"
        assert "created" in d

    def test_cli_init_fails_non_empty_exits_one(
        self, tmp_path: pathlib.Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        (tmp_path / "x").write_text("y")
        rc = main(["init", "fastapi", str(tmp_path)])
        assert rc == 1
        captured = capsys.readouterr()
        assert "ERROR" in captured.err

    def test_cli_init_force_overwrites(self, tmp_path: pathlib.Path) -> None:
        (tmp_path / "x").write_text("y")
        rc = main(["init", "fastapi", str(tmp_path), "--force"])
        assert rc == 0

    def test_cli_init_json_mode_error(
        self, tmp_path: pathlib.Path
    ) -> None:
        (tmp_path / "blocker").write_text("x")
        buf = io.StringIO()
        with redirect_stdout(buf):
            rc = main(["init", "fastapi", str(tmp_path), "--json"])
        assert rc == 1
        d = json.loads(buf.getvalue())
        assert d["result"] == "ERROR"

    def test_cli_default_target_dir(self, tmp_path: pathlib.Path) -> None:
        """Default target-dir arg ('campaign-api') used when omitted."""
        target = tmp_path / "campaign-api"
        # Call via explicit path so we don't pollute cwd
        rc = main(["init", "fastapi", str(target)])
        assert rc == 0
        assert (target / "main.py").exists()
