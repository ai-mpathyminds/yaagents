"""Tests for WI-1yaa.CLI-4: yaagents conformance-test.

AC:
  - Run against Compose demo → Overall: PASS, exit 0      (live test, skipped in CI)
  - Report text matches PRD §5.8 format (✓ lines + Overall: PASS/FAIL)
  - Deliberately broken route → FAIL exit 1
  - Non-http/https URL → exit 1 with clear error
  - --json mode emits machine-readable JSON
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
    _PROFILE_HEADER,
    _PROFILE_VERSION,
    DEMO_JWT_SECRET,
    _check_clarification_ct,
    _check_clarification_schema,
    _check_correlation_id,
    _check_profile_header,
    _check_tenant_required,
    conformance_test,
    make_jwt,
)

# ── JWT helper ────────────────────────────────────────────────────────────────


class TestMakeJwt:
    def test_returns_three_parts(self) -> None:
        token = make_jwt("secret")
        parts = token.split(".")
        assert len(parts) == 3

    def test_header_alg_hs256(self) -> None:
        import base64

        token = make_jwt("secret")
        header_b64 = token.split(".")[0]
        # add padding
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
        "Content-Type": "application/json",
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


_SIDE_EFFECTS_PASS = [
    # Check 1: POST /campaigns (full) → 201 with profile header
    _make_mock_response(201, _valid_201_headers(), b'{"campaign":{}}'),
    # Check 2: POST /campaigns (no successMetric) → 400 clarification
    _make_mock_response(400, _clarification_headers(), _clarification_body()),
    # Check 4: POST /campaigns (with X-Correlation-ID) → 201 with echoed header
    _make_mock_response(
        201,
        {
            **_valid_201_headers(),
            "X-Correlation-ID": _CORR_ID_SENTINEL,
        },
        b'{"campaign":{}}',
    ),
    # Check 5: POST /campaigns (no X-Tenant-ID) → 403 error
    _make_mock_response(
        403,
        _forbidden_headers(),
        b'{"type":"forbidden","code":"TENANT_REQUIRED","message":"x","trace":{}}',
    ),
]


# ── individual-check unit tests ───────────────────────────────────────────────


class TestCheckProfileHeader:
    def test_pass_when_header_present(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, _valid_201_headers(), b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert result.passed

    def test_fail_when_header_absent(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, {"Content-Type": "application/json"}, b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert not result.passed
        assert _PROFILE_HEADER in result.detail or "expected" in result.detail

    def test_fail_when_wrong_version(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201, {_PROFILE_HEADER: "v0.2"}, b"{}"
            )
            result = _check_profile_header("http://gw", "tok", "t1")
        assert not result.passed


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
                200, {"Content-Type": "application/json"}, b"{}"
            )
            check, _ = _check_clarification_ct("http://gw", "tok", "t1")
        assert not check.passed

    def test_fail_when_wrong_ct(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                400, {"Content-Type": "application/json"}, b"{}"
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
        assert result.detail  # has error detail

    def test_fail_on_non_json(self) -> None:
        result = _check_clarification_schema(b"not json")
        assert not result.passed


class TestCheckCorrelationId:
    def test_pass_when_echoed(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                201,
                {
                    **_valid_201_headers(),
                    "X-Correlation-ID": _CORR_ID_SENTINEL,
                },
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
            mock_open.return_value = _make_mock_response(
                403, _forbidden_headers(), b"{}"
            )
            result = _check_tenant_required("http://gw", "tok")
        assert result.passed

    def test_fail_when_200_returned(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.return_value = _make_mock_response(
                200, {"Content-Type": "application/json"}, b"{}"
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


# ── full suite (mocked) ───────────────────────────────────────────────────────


class TestConformanceSuite:
    def test_all_pass(self) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = _SIDE_EFFECTS_PASS
            result = conformance_test(
                "http://localhost:8120", jwt_secret=DEMO_JWT_SECRET
            )
        assert result.passed
        assert len(result.checks) == 5
        assert all(c.passed for c in result.checks)

    def test_broken_profile_header_fails(self) -> None:
        """Broken route (profile header absent) → result.passed is False."""
        broken_201 = _make_mock_response(
            201,
            {"Content-Type": "application/json"},  # no profile header
            b"{}",
        )
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _make_mock_response(403, _forbidden_headers(), b"{}")
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = [broken_201, clarif, corr, forbid]
            result = conformance_test("http://localhost:8120")
        assert not result.passed
        assert (
            result.checks[0].name
            == "X-YAAgents-Profile header on proxied response"
        )
        assert not result.checks[0].passed

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
            mock_open.side_effect = _SIDE_EFFECTS_PASS
            result = conformance_test("http://localhost:8120")
        d = result.to_dict()
        assert d["result"] == "PASS"
        assert len(d["checks"]) == 5
        assert all(c["result"] == "PASS" for c in d["checks"])

    def test_result_to_dict_fail(self) -> None:
        result = conformance_test("file:///etc/hosts")
        d = result.to_dict()
        assert d["result"] == "FAIL"
        assert "error" in d


# ── CLI integration (report format) ──────────────────────────────────────────


class TestCliOutput:
    def test_pass_emits_report(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = list(_SIDE_EFFECTS_PASS)
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
            201, {"Content-Type": "application/json"}, b"{}"
        )
        clarif = _make_mock_response(
            400, _clarification_headers(), _clarification_body()
        )
        corr = _make_mock_response(
            201, {"X-Correlation-ID": _CORR_ID_SENTINEL}, b"{}"
        )
        forbid = _make_mock_response(403, _forbidden_headers(), b"{}")
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = [broken_201, clarif, corr, forbid]
            rc = main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert rc == 1
        assert "Overall: FAIL" in out
        assert "✗" in out

    def test_json_mode_pass(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = list(_SIDE_EFFECTS_PASS)
            rc = main(
                ["conformance-test", "http://localhost:8120",
                 "--jwt-secret", DEMO_JWT_SECRET, "--json"]
            )
        out = capsys.readouterr().out
        data = json.loads(out)
        assert rc == 0
        assert data["result"] == "PASS"
        assert len(data["checks"]) == 5

    def test_json_mode_fail(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["conformance-test", "file:///etc/hosts", "--json"])
        out = capsys.readouterr().out
        data = json.loads(out)
        assert rc == 1
        assert data["result"] == "FAIL"

    def test_report_check_names_in_output(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("urllib.request.urlopen") as mock_open:
            mock_open.side_effect = list(_SIDE_EFFECTS_PASS)
            main(["conformance-test", "http://localhost:8120"])
        out = capsys.readouterr().out
        assert "X-YAAgents-Profile header on proxied response" in out
        assert "Clarification response uses correct content type" in out
        assert "400 response matches clarification schema" in out
        assert "Correlation ID propagated" in out
        assert "Gateway route requires tenant context" in out

    def test_url_scheme_error_exit1(
        self, capsys: pytest.CaptureFixture[str]
    ) -> None:
        rc = main(["conformance-test", "file:///etc/hosts"])
        assert rc == 1
        err = capsys.readouterr().err
        assert "http" in err
