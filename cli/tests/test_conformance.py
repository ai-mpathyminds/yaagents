# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for WI-2yaa.CLI-CONF: yaagents conformance-test v0.2.

Covers:
  - Profile header assertion (v0.2 enforced; v0.1 gateway → profile-mismatch FAIL)
  - --require-plugin token-validator PASS / FAIL
  - --require-plugin tenant-injector PASS / FAIL
  - Always-on token-validator assertion (10/10 probes → 403)
  - 10-row Content-Type matrix in ConformanceResult.matrix
  - Summary table rendered in CLI output
  - All PI1-yaa SPEC-5 regression checks (5 existing checks preserved)
  - ≥85% coverage on the conformance command path
"""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

import pytest

from yaagents_cli.__main__ import main
from yaagents_cli._conformance import (
    _CORR_ID_SENTINEL,
    _CT_CLARIFICATION,
    _CT_ERROR,
    _CT_JSON,
    _PROFILE_HEADER,
    _PROFILE_VERSION,
    _TEN_PROBES,
    DEMO_JWT_SECRET,
    MATRIX_ROWS,
    _check_always_on,
    _check_clarification_ct,
    _check_clarification_schema,
    _check_correlation_id,
    _check_correlation_id_llm,
    _check_plugin_tenant_injector,
    _check_plugin_tenant_injector_llm,
    _check_plugin_token_validator,
    _check_profile_header,
    _check_profile_header_and_detect,
    _check_tenant_required,
    _check_tenant_required_llm,
    _make_invalid_jwt,
    conformance_test,
    make_jwt,
)

# ── JWT helpers ───────────────────────────────────────────────────────────────


class TestMakeJwt:
    def test_returns_three_parts(self) -> None:
        token = make_jwt("secret")
        parts = token.split(".")
        assert len(parts) == 3

    def test_header_alg_hs256(self) -> None:
        import base64

        token = make_jwt("secret")
        header_b64 = token.split(".")[0]
        padded = header_b64 + "=" * (-len(header_b64) % 4)
        header = json.loads(base64.urlsafe_b64decode(padded))
        assert header["alg"] == "HS256"
        assert header["typ"] == "JWT"

    def test_custom_roles(self) -> None:
        import base64

        token = make_jwt("s", roles=[])
        payload_b64 = token.split(".")[1]
        padded = payload_b64 + "=" * (-len(payload_b64) % 4)
        payload = json.loads(base64.urlsafe_b64decode(padded))
        assert payload["roles"] == []

    def test_default_roles_include_optimize(self) -> None:
        import base64

        token = make_jwt("s")
        payload_b64 = token.split(".")[1]
        padded = payload_b64 + "=" * (-len(payload_b64) % 4)
        payload = json.loads(base64.urlsafe_b64decode(padded))
        assert "campaign:optimize" in payload["roles"]


class TestMakeInvalidJwt:
    def test_is_valid_jwt_shape(self) -> None:
        token = _make_invalid_jwt()
        assert len(token.split(".")) == 3

    def test_uses_different_secret_than_demo(self) -> None:
        # The invalid JWT should NOT verify with the demo secret
        import base64
        import hashlib
        import hmac as _hmac

        token = _make_invalid_jwt()
        parts = token.split(".")
        signing_input = f"{parts[0]}.{parts[1]}"
        expected_sig = _hmac.new(
            DEMO_JWT_SECRET.encode(),
            signing_input.encode(),
            hashlib.sha256,
        ).digest()
        expected_b64 = base64.urlsafe_b64encode(expected_sig).rstrip(b"=").decode()
        assert parts[2] != expected_b64  # signature mismatch → invalid for demo gateway


# ── URL scheme validation ──────────────────────────────────────────────────────


class TestUrlSchemeValidation:
    def test_file_scheme_rejected(self) -> None:
        result = conformance_test("file:///etc/hosts")
        assert not result.passed
        assert result.error is not None
        assert "http" in result.error

    def test_ftp_scheme_rejected(self) -> None:
        result = conformance_test("ftp://example.com")
        assert not result.passed
        assert result.error is not None

    def test_cli_file_scheme_exit1(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["conformance-test", "file:///etc/hosts"])
        assert rc == 1
        err = capsys.readouterr().err
        assert "http" in err

    def test_empty_scheme_rejected(self) -> None:
        result = conformance_test("localhost:8120")
        assert not result.passed
        assert result.error is not None


# ── mock helpers ──────────────────────────────────────────────────────────────


def _make_mock_response(
    status: int,
    headers: dict[str, str],
    body: bytes,
) -> MagicMock:
    """Build a fake response compatible with urllib.request.urlopen context-manager."""
    mock_resp = MagicMock()
    mock_resp.status = status
    mock_resp.headers = headers
    mock_resp.read.return_value = body
    mock_resp.__enter__ = lambda s: s
    mock_resp.__exit__ = MagicMock(return_value=False)
    return mock_resp


def _clarification_body() -> bytes:
    return json.dumps(
        {
            "type": "clarification_required",
            "code": "CLARIFICATION_REQUIRED",
            "message": "successMetric is required.",
            "requiredInputs": [
                {
                    "name": "successMetric",
                    "location": "body",
                    "type": "string",
                    "required": True,
                    "question": "Which metric?",
                }
            ],
            "trace": {"correlationId": "c1", "requestId": "r1"},
        }
    ).encode()


def _valid_201_headers() -> dict[str, str]:
    return {
        _PROFILE_HEADER: _PROFILE_VERSION,
        "Content-Type": _CT_JSON,
        "X-Correlation-ID": _CORR_ID_SENTINEL,
        "X-Request-ID": "req-1",
    }


def _clarification_headers() -> dict[str, str]:
    return {
        _PROFILE_HEADER: _PROFILE_VERSION,
        "Content-Type": _CT_CLARIFICATION,
        "X-Correlation-ID": "c1",
    }


def _forbidden_headers() -> dict[str, str]:
    return {
        _PROFILE_HEADER: _PROFILE_VERSION,
        "Content-Type": _CT_ERROR,
        "X-Correlation-ID": "c2",
    }


def _forbidden_body() -> bytes:
    return json.dumps({
        "type": "forbidden",
        "code": "TOKEN_INVALID",
        "message": "x",
        "trace": {"correlationId": "c1", "requestId": "r1"},
    }).encode()


def _forbidden_response() -> MagicMock:
    return _make_mock_response(403, _forbidden_headers(), _forbidden_body())


# 6 mock responses for the 6 checks (profile, clarif-ct, corr-id, tenant, always-on×10)
# plus plugin checks when needed.
_SIDE_EFFECTS_BASE = [
    # Check 1: POST /campaigns (full) → 201 with profile header
    _make_mock_response(201, _valid_201_headers(), b'{"campaign":{}}'),
    # Check 2: POST /campaigns (no successMetric) → 400 clarification
    _make_mock_response(400, _clarification_headers(), _clarification_body()),
    # Check 4: POST /campaigns (with X-Correlation-ID) → 201 with echoed header
    _make_mock_response(
        201,
        {**_valid_201_headers(), "X-Correlation-ID": _CORR_ID_SENTINEL},
        b'{"campaign":{}}',
    ),
    # Check 5: POST /campaigns (no X-Tenant-ID) → 403 error
    _make_mock_response(403, _forbidden_headers(), _forbidden_body()),
]

# Always-on 10 probes: all return 403 (invalid JWT intercepted)
_ALWAYS_ON_403 = [_forbidden_response() for _ in range(10)]


def _make_pass_side_effects(
    extra_plugin_responses: list[MagicMock] | None = None,
) -> list[MagicMock]:
    return (
        list(_SIDE_EFFECTS_BASE)
        + list(_ALWAYS_ON_403)
        + (extra_plugin_responses or [])
    )


# ── individual-check unit tests (PI1-yaa regression) ─────────────────────────


class TestCheckProfileHeader:
    def test_pass_when_v02_header_present(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, _valid_201_headers(), b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert result.passed

    def test_fail_when_header_absent(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert not result.passed
        assert "header-absent" in result.detail

    def test_fail_when_v01_gateway(self) -> None:
        """A v0.1 gateway returns X-YAAgents-Profile: v0.1 → profile-mismatch."""
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, {_PROFILE_HEADER: "v0.1"}, b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert not result.passed
        assert "profile-mismatch" in result.detail
        assert "v0.1" in result.detail

    def test_check_name_contains_v02(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, _valid_201_headers(), b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert "v0.2" in result.name


class TestCheckClarificationCt:
    def test_pass_when_correct_ct_and_status(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                400, _clarification_headers(), _clarification_body()
            )
            check, _ = _check_clarification_ct("http://gw", "tok", "t1")
        assert check.passed

    def test_fail_when_wrong_status(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            check, _ = _check_clarification_ct("http://gw", "tok", "t1")
        assert not check.passed

    def test_fail_when_wrong_ct(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                400, {"Content-Type": _CT_JSON}, b"{}"
            )
            check, _ = _check_clarification_ct("http://gw", "tok", "t1")
        assert not check.passed

    def test_ct_with_charset_suffix_passes(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                400,
                {"Content-Type": f"{_CT_CLARIFICATION}; charset=utf-8"},
                _clarification_body(),
            )
            check, _ = _check_clarification_ct("http://gw", "tok", "t1")
        assert check.passed


class TestCheckClarificationSchema:
    def test_pass_on_valid_body(self) -> None:
        result = _check_clarification_schema(_clarification_body())
        assert result.passed

    def test_fail_on_invalid_body(self) -> None:
        bad = json.dumps({"type": "clarification_required"}).encode()
        result = _check_clarification_schema(bad)
        assert not result.passed
        assert result.detail

    def test_fail_on_non_json(self) -> None:
        result = _check_clarification_schema(b"not json")
        assert not result.passed


class TestCheckCorrelationId:
    def test_pass_when_echoed(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201,
                {**_valid_201_headers(), "X-Correlation-ID": _CORR_ID_SENTINEL},
                b"{}",
            )
            result = _check_correlation_id("http://gw", "tok", "t1")
        assert result.passed

    def test_fail_when_not_echoed(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201,
                {"X-Correlation-ID": "different-id"},
                b"{}",
            )
            result = _check_correlation_id("http://gw", "tok", "t1")
        assert not result.passed

    def test_fail_when_absent(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(201, {}, b"{}")
            result = _check_correlation_id("http://gw", "tok", "t1")
        assert not result.passed


class TestCheckTenantRequired:
    def test_pass_when_403_vendor_error(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _forbidden_response()
            result = _check_tenant_required("http://gw", "tok")
        assert result.passed

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_tenant_required("http://gw", "tok")
        assert not result.passed

    def test_fail_when_403_wrong_ct(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                403, {"Content-Type": "text/plain"}, b"forbidden"
            )
            result = _check_tenant_required("http://gw", "tok")
        assert not result.passed


# ── v0.2 new checks ───────────────────────────────────────────────────────────


class TestCheckAlwaysOn:
    def test_pass_when_all_10_return_403(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = [_forbidden_response() for _ in range(10)]
            result = _check_always_on("http://gw", "t1")
        assert result.passed
        assert "10/10" in result.name

    def test_fail_when_one_returns_200(self) -> None:
        responses = [_forbidden_response() for _ in range(9)]
        responses.insert(2, _make_mock_response(200, {"Content-Type": _CT_JSON}, b"{}"))
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = responses
            result = _check_always_on("http://gw", "t1")
        assert not result.passed
        assert result.detail  # has failure detail

    def test_fail_when_one_returns_403_wrong_ct(self) -> None:
        responses = [_forbidden_response() for _ in range(9)]
        responses.insert(0, _make_mock_response(
            403, {"Content-Type": "text/plain"}, b"forbidden"
        ))
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = responses
            result = _check_always_on("http://gw", "t1")
        assert not result.passed

    def test_issues_exactly_10_probes(self) -> None:
        assert len(_TEN_PROBES) == 10

    def test_connection_error_in_probe_counts_as_fail(self) -> None:
        import urllib.error

        responses: list[MagicMock | Exception] = [
            _forbidden_response() for _ in range(9)
        ]
        responses.insert(0, urllib.error.URLError("refused"))
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = responses
            result = _check_always_on("http://gw", "t1")
        assert not result.passed


class TestCheckPluginTokenValidator:
    def test_pass_when_403_vendor_error(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _forbidden_response()
            result = _check_plugin_token_validator("http://gw", "t1")
        assert result.passed
        assert result.name == "plugin:token-validator"

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_plugin_token_validator("http://gw", "t1")
        assert not result.passed
        assert "token-validator" in result.detail

    def test_fail_when_403_wrong_ct(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                403, {"Content-Type": "text/html"}, b"forbidden"
            )
            result = _check_plugin_token_validator("http://gw", "t1")
        assert not result.passed


class TestCheckPluginTenantInjector:
    def test_pass_when_403_vendor_error(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _forbidden_response()
            result = _check_plugin_tenant_injector("http://gw", "valid-token")
        assert result.passed
        assert result.name == "plugin:tenant-injector"

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_plugin_tenant_injector("http://gw", "valid-token")
        assert not result.passed
        assert "tenant-injector" in result.detail


# ── MATRIX_ROWS constant ─────────────────────────────────────────────────────


class TestMatrixRows:
    def test_has_exactly_10_rows(self) -> None:
        assert len(MATRIX_ROWS) == 10

    def test_all_expected_statuses_present(self) -> None:
        statuses = {r[0] for r in MATRIX_ROWS}
        assert statuses == {200, 201, 202, 400, 403, 409, 412, 422, 424, 500}

    def test_forbidden_row_uses_error_ct(self) -> None:
        row = next(r for r in MATRIX_ROWS if r[0] == 403)
        assert row[1] == _CT_ERROR

    def test_clarification_row(self) -> None:
        row = next(r for r in MATRIX_ROWS if r[0] == 400)
        assert row[1] == _CT_CLARIFICATION


# ── full suite (mocked) ───────────────────────────────────────────────────────


class TestConformanceSuite:
    def test_all_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            result = conformance_test(
                "http://localhost:8120", jwt_secret=DEMO_JWT_SECRET
            )
        assert result.passed
        # 6 core checks (profile, clarif-ct, clarif-schema, corr-id, tenant, always-on)
        assert len(result.checks) == 6
        assert all(c.passed for c in result.checks)

    def test_matrix_has_10_rows(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            result = conformance_test("http://localhost:8120")
        assert len(result.matrix) == 10

    def test_broken_profile_header_fails(self) -> None:
        """v0.1 gateway (profile header absent) → result.passed is False."""
        broken_201 = _make_mock_response(
            201,
            {"Content-Type": _CT_JSON},  # no profile header
            b"{}",
        )
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _forbidden_response()
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = (
                [broken_201, clarif, corr, forbid] + list(_ALWAYS_ON_403)
            )
            result = conformance_test("http://localhost:8120")
        assert not result.passed
        assert not result.checks[0].passed
        detail = result.checks[0].detail
        assert "header-absent" in detail or "profile-mismatch" in detail

    def test_v01_gateway_profile_mismatch(self) -> None:
        """Gateway returning X-YAAgents-Profile: v0.1 → profile-mismatch FAIL."""
        v01_headers = {**_valid_201_headers(), _PROFILE_HEADER: "v0.1"}
        broken_201 = _make_mock_response(201, v01_headers, b"{}")
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _forbidden_response()
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = (
                [broken_201, clarif, corr, forbid] + list(_ALWAYS_ON_403)
            )
            result = conformance_test("http://localhost:8120")
        assert not result.passed
        assert "profile-mismatch" in result.checks[0].detail

    def test_connection_error_captured(self) -> None:
        import urllib.error

        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = urllib.error.URLError("Connection refused")
            result = conformance_test("http://localhost:9999")
        assert not result.passed
        assert result.error is not None
        assert "connection" in result.error.lower()

    def test_result_to_dict_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            result = conformance_test("http://localhost:8120")
        d = result.to_dict()
        assert d["result"] == "PASS"
        assert len(d["checks"]) == 6
        assert len(d["matrix"]) == 10
        assert all(c["result"] == "PASS" for c in d["checks"])

    def test_result_to_dict_fail(self) -> None:
        result = conformance_test("file:///etc/hosts")
        d = result.to_dict()
        assert d["result"] == "FAIL"
        assert "error" in d

    def test_require_plugin_token_validator_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects(
                extra_plugin_responses=[_forbidden_response()]
            )
            result = conformance_test(
                "http://localhost:8120",
                require_plugins=["token-validator"],
            )
        assert result.passed
        plugin_checks = [c for c in result.checks if c.name == "plugin:token-validator"]
        assert len(plugin_checks) == 1
        assert plugin_checks[0].passed

    def test_require_plugin_token_validator_fail(self) -> None:
        # Plugin check returns 200 instead of 403 → FAIL
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects(
                extra_plugin_responses=[
                    _make_mock_response(200, {"Content-Type": _CT_JSON}, b"{}")
                ]
            )
            result = conformance_test(
                "http://localhost:8120",
                require_plugins=["token-validator"],
            )
        assert not result.passed
        plugin_checks = [c for c in result.checks if c.name == "plugin:token-validator"]
        assert not plugin_checks[0].passed

    def test_require_plugin_tenant_injector_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects(
                extra_plugin_responses=[_forbidden_response()]
            )
            result = conformance_test(
                "http://localhost:8120",
                require_plugins=["tenant-injector"],
            )
        assert result.passed
        plugin_checks = [c for c in result.checks if c.name == "plugin:tenant-injector"]
        assert plugin_checks[0].passed

    def test_require_both_plugins_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects(
                extra_plugin_responses=[_forbidden_response(), _forbidden_response()]
            )
            result = conformance_test(
                "http://localhost:8120",
                require_plugins=["token-validator", "tenant-injector"],
            )
        assert result.passed
        assert len([c for c in result.checks if c.name.startswith("plugin:")]) == 2

    def test_unknown_plugin_fails(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            result = conformance_test(
                "http://localhost:8120",
                require_plugins=["nonexistent-plugin"],
            )
        assert not result.passed
        plugin_check = next(c for c in result.checks if "plugin:" in c.name)
        assert not plugin_check.passed
        assert "unknown" in plugin_check.detail


# ── CLI integration (report format) ──────────────────────────────────────────


class TestCliOutput:
    def test_pass_emits_report(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            rc = main(
                ["conformance-test", "http://localhost:8120",
                 "--jwt-secret", DEMO_JWT_SECRET]
            )
        out = capsys.readouterr().out
        assert rc == 0
        assert "YAAgents Conformance Report" in out
        assert "Overall: PASS" in out
        assert "✓" in out

    def test_fail_emits_fail(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        broken_201 = _make_mock_response(
            201, {"Content-Type": _CT_JSON}, b"{}"
        )
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _forbidden_response()
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = (
                [broken_201, clarif, corr, forbid] + list(_ALWAYS_ON_403)
            )
            rc = main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert rc == 1
        assert "Overall: FAIL" in out
        assert "✗" in out

    def test_json_mode_pass(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            rc = main(
                ["conformance-test", "http://localhost:8120",
                 "--jwt-secret", DEMO_JWT_SECRET, "--json"]
            )
        out = capsys.readouterr().out
        data = json.loads(out)
        assert rc == 0
        assert data["result"] == "PASS"
        assert len(data["checks"]) == 6
        assert len(data["matrix"]) == 10

    def test_json_mode_fail(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["conformance-test", "file:///etc/hosts", "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert rc == 1
        assert data["result"] == "FAIL"

    def test_summary_table_in_output(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        """Summary table (status | requested | observed | pass) present in output."""
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert "Content-Type matrix" in out
        assert "status" in out
        assert "requested" in out
        assert "observed" in out
        assert "pass" in out
        # All 10 status codes present in table
        for status in (200, 201, 202, 400, 403, 409, 412, 422, 424, 500):
            assert str(status) in out

    def test_require_plugin_flag_pass(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects(
                extra_plugin_responses=[_forbidden_response()]
            )
            rc = main([
                "conformance-test", "http://localhost:8120",
                "--require-plugin", "token-validator",
            ])
        out = capsys.readouterr().out
        assert rc == 0
        assert "plugin:token-validator" in out
        assert "✓" in out

    def test_v01_gateway_exits_1_with_profile_mismatch(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        """Running against a v0.1 gateway → exit 1, 'profile-mismatch' in output."""
        v01_headers = {**_valid_201_headers(), _PROFILE_HEADER: "v0.1"}
        broken_201 = _make_mock_response(201, v01_headers, b"{}")
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _forbidden_response()
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = (
                [broken_201, clarif, corr, forbid] + list(_ALWAYS_ON_403)
            )
            rc = main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert rc == 1
        assert "profile-mismatch" in out

    def test_report_check_names_in_output(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_pass_side_effects()
            main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert "v0.2" in out
        assert "Clarification response uses correct content type" in out
        assert "400 response matches clarification schema" in out
        assert "Correlation ID propagated" in out
        assert "Gateway route requires tenant context" in out
        assert "always-on" in out

    def test_url_scheme_error_exit1(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["conformance-test", "file:///etc/hosts"])
        assert rc == 1
        err = capsys.readouterr().err
        assert "http" in err


# ── make_jwt sub parameter ────────────────────────────────────────────────────


class TestMakeJwtSubParam:
    def test_default_sub_is_conformance_tester(self) -> None:
        import base64

        token = make_jwt("s")
        payload_b64 = token.split(".")[1]
        padded = payload_b64 + "=" * (-len(payload_b64) % 4)
        payload = json.loads(base64.urlsafe_b64decode(padded))
        assert payload["sub"] == "conformance-tester"

    def test_custom_sub(self) -> None:
        import base64

        token = make_jwt("s", sub="custom-user@example.com")
        payload_b64 = token.split(".")[1]
        padded = payload_b64 + "=" * (-len(payload_b64) % 4)
        payload = json.loads(base64.urlsafe_b64decode(padded))
        assert payload["sub"] == "custom-user@example.com"


# ── LLM-gateway mode: _check_profile_header_and_detect ───────────────────────


class TestCheckProfileHeaderAndDetect:
    def test_campaign_api_mode_pass(self) -> None:
        """201 from /campaigns → campaign-api mode, profile header present → PASS."""
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, _valid_201_headers(), b"{}"
            )
            result, is_llm = _check_profile_header_and_detect("http://gw", "tok", "t1")
        assert result.passed
        assert not is_llm

    def test_campaign_api_mode_profile_mismatch(self) -> None:
        """v0.1 gateway (campaign-api) → FAIL, is_llm=False."""
        v01_hdrs = {**_valid_201_headers(), _PROFILE_HEADER: "v0.1"}
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(201, v01_hdrs, b"{}")
            result, is_llm = _check_profile_header_and_detect("http://gw", "tok", "t1")
        assert not result.passed
        assert "profile-mismatch" in result.detail
        assert not is_llm

    def test_llm_gateway_mode_detected_on_404(self) -> None:
        """404 from /campaigns + v0.2 from /completions → PASS, is_llm=True."""
        resp_404 = _make_mock_response(
            404, {"content-type": _CT_ERROR}, b'{"type":"error"}'
        )
        resp_completions = _make_mock_response(201, _valid_201_headers(), b"{}")
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = [resp_404, resp_completions]
            result, is_llm = _check_profile_header_and_detect("http://gw", "tok", "t1")
        assert result.passed
        assert is_llm
        assert "completions" in result.detail

    def test_llm_gateway_no_profile_header_on_completions(self) -> None:
        """LLM gateway: no profile header on /completions → FAIL, is_llm=True."""
        resp_404 = _make_mock_response(404, {}, b'{"type":"error"}')
        resp_completions = _make_mock_response(201, {"content-type": _CT_JSON}, b"{}")
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = [resp_404, resp_completions]
            result, is_llm = _check_profile_header_and_detect("http://gw", "tok", "t1")
        assert not result.passed
        assert "header-absent" in result.detail
        assert is_llm


# ── LLM-gateway mode: _check_correlation_id_llm ──────────────────────────────


class TestCheckCorrelationIdLlm:
    def test_pass_when_echoed(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201,
                {**_valid_201_headers(), "X-Correlation-ID": _CORR_ID_SENTINEL},
                b"{}",
            )
            result = _check_correlation_id_llm("http://gw", "tok", "t1")
        assert result.passed

    def test_fail_when_absent(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            # Headers without the correlation ID (no X-Correlation-ID key)
            mock_open.return_value = _make_mock_response(
                201,
                {"content-type": _CT_JSON, _PROFILE_HEADER: _PROFILE_VERSION},
                b"{}",
            )
            result = _check_correlation_id_llm("http://gw", "tok", "t1")
        assert not result.passed


# ── LLM-gateway mode: _check_tenant_required_llm ─────────────────────────────


class TestCheckTenantRequiredLlm:
    def test_pass_when_unknown_principal_rejected(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _forbidden_response()
            result = _check_tenant_required_llm("http://gw", DEMO_JWT_SECRET)
        assert result.passed

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_tenant_required_llm("http://gw", DEMO_JWT_SECRET)
        assert not result.passed


# ── LLM-gateway mode: _check_plugin_tenant_injector_llm ─────────────────────


class TestCheckPluginTenantInjectorLlm:
    def test_pass_when_403_vendor_error(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _forbidden_response()
            result = _check_plugin_tenant_injector_llm("http://gw", DEMO_JWT_SECRET)
        assert result.passed
        assert result.name == "plugin:tenant-injector"

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": _CT_JSON}, b"{}"
            )
            result = _check_plugin_tenant_injector_llm("http://gw", DEMO_JWT_SECRET)
        assert not result.passed


# ── full LLM-gateway suite (conformance_test in LLM mode) ────────────────────


def _make_llm_pass_side_effects(
    extra_plugin_responses: list[MagicMock] | None = None,
) -> list[MagicMock]:
    """Side-effects for a full PASS run against an LLM gateway.

    Call order:
      1. check 1a: /campaigns → 404 (detect LLM mode)
      2. check 1b: /completions → 201 + profile header (LLM profile confirm)
      3. check 4:  /completions + correlation ID → 201 + corr-id echo
      4. check 5:  /completions (unknown JWT) → 403 (tenant-injector)
      5-14. always-on 10 probes → 403 each
      15+. plugin checks (optional)
    """
    _llm_404 = _make_mock_response(
        404, {"content-type": _CT_ERROR}, b'{"type":"error","code":"ROUTE_NOT_FOUND"}'
    )
    _completions_201 = _make_mock_response(
        201, _valid_201_headers(), b'{"type":"completion"}'
    )
    _corr_201 = _make_mock_response(
        201,
        {**_valid_201_headers(), "X-Correlation-ID": _CORR_ID_SENTINEL},
        b'{"type":"completion"}',
    )
    return (
        [_llm_404, _completions_201, _corr_201, _forbidden_response()]
        + list(_ALWAYS_ON_403)
        + (extra_plugin_responses or [])
    )


class TestConformanceSuiteLlmGateway:
    def test_llm_gateway_all_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_llm_pass_side_effects()
            result = conformance_test(
                "http://localhost:8122", jwt_secret=DEMO_JWT_SECRET
            )
        assert result.passed
        assert len(result.checks) == 6
        assert all(c.passed for c in result.checks)

    def test_llm_gateway_clarification_checks_are_na(self) -> None:
        """Clarification checks are N/A (skipped) in LLM-gateway mode."""
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_llm_pass_side_effects()
            result = conformance_test("http://localhost:8122")
        clarif_checks = [
            c for c in result.checks
            if "Clarification" in c.name or "clarification" in c.name
        ]
        for c in clarif_checks:
            assert c.passed
            assert "N/A" in c.detail

    def test_llm_gateway_with_require_plugins(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_llm_pass_side_effects(
                extra_plugin_responses=[
                    _forbidden_response(),  # token-validator
                    _forbidden_response(),  # tenant-injector (LLM mode)
                ]
            )
            result = conformance_test(
                "http://localhost:8122",
                require_plugins=["token-validator", "tenant-injector"],
            )
        assert result.passed
        plugin_checks = [c for c in result.checks if c.name.startswith("plugin:")]
        assert len(plugin_checks) == 2
        assert all(c.passed for c in plugin_checks)

    def test_llm_gateway_matrix_has_10_rows(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _make_llm_pass_side_effects()
            result = conformance_test("http://localhost:8122")
        assert len(result.matrix) == 10

    def test_llm_gateway_profile_mismatch_fails(self) -> None:
        """LLM gateway returning wrong profile version → FAIL."""
        _llm_404 = _make_mock_response(404, {}, b'{}')
        _completions_v01 = _make_mock_response(
            201, {**_valid_201_headers(), _PROFILE_HEADER: "v0.1"}, b"{}"
        )
        _corr = _make_mock_response(
            201,
            {**_valid_201_headers(), "X-Correlation-ID": _CORR_ID_SENTINEL},
            b"{}",
        )
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = (
                [_llm_404, _completions_v01, _corr, _forbidden_response()]
                + list(_ALWAYS_ON_403)
            )
            result = conformance_test("http://localhost:8122")
        assert not result.passed
        assert "profile-mismatch" in result.checks[0].detail
