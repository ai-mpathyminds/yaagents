"""PI1-yaa acceptance gate — WI-1yaa.EX-4.

Runs against the live Compose demo (yaagents-gateway at http://localhost:8120).
Tests are automatically skipped when the gateway is unreachable (see conftest.py).

Override the gateway URL:
    GATEWAY_URL=http://localhost:8120 pytest tests/test_e2e_gate.py -v

Verifies PRD §12 success criteria 1–9 (10–12 are publish criteria; see REL-3/4/5):
  1. FastAPI SDK exposes an agentic endpoint  → verified by test_flows.py (unit)
  2. Endpoint returns all 4 §6.2 flows       → test_flow*_through_gateway
  3. SDK maps status + content-type (§4)     → verified by test_flows.py (unit)
  4. OpenAPI includes vendor content-types   → verified by test_flows.py (unit)
  5. Gateway enforces RBAC                   → test_gateway_rbac_403
  6. Gateway propagates tenant/corr-id       → test_gateway_correlation_id_propagated
  7. Both clients handle clarification       → test_python_client_clarification_required
                                                test_ts_client_e2e (Node.js subprocess)
  8. Compose demo runs docker compose up     → gateway_live fixture (healthz up)
  9. conformance-test PASS                   → test_conformance_test_pass
"""

from __future__ import annotations

import json
import pathlib
import subprocess
import sys
import urllib.error
import urllib.request

import pytest

from yaagents_cli._conformance import (
    DEMO_JWT_SECRET,
    conformance_test,
    make_jwt,
)
from yaagents_client import ClarificationRequired, YaAgentsClient
from yaagents_client._mapper import process_response

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

_CT_JSON = "application/json"
_CT_CLARIFICATION = "application/vnd.yaagents.clarification+json"
_CT_ERROR = "application/vnd.yaagents.error+json"
_PROFILE_HEADER = "X-YAAgents-Profile"
_CORR_ID = "e2e-gate-corr-id-001"
_TENANT = "tenant-e2e"
_REQUEST_TIMEOUT = 15

# repo root — two levels up from examples/campaign-api/
_REPO_ROOT = pathlib.Path(__file__).parent.parent.parent.parent.resolve()
_SPEC_DIR = _REPO_ROOT / "spec"
_TS_E2E_SCRIPT = pathlib.Path(__file__).parent.parent / "scripts" / "ts_client_e2e.mjs"


# ---------------------------------------------------------------------------
# HTTP helper (urllib — no extra deps needed at test runtime)
# ---------------------------------------------------------------------------


def _http(
    method: str,
    url: str,
    *,
    headers: dict[str, str] | None = None,
    body: bytes | None = None,
    timeout: int = _REQUEST_TIMEOUT,
) -> tuple[int, dict[str, str], bytes]:
    req = urllib.request.Request(url, data=body, method=method)
    for k, v in (headers or {}).items():
        req.add_header(k, v)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            # Lowercase all header keys — Go's net/http canonicalises
            # "X-YAAgents-Profile" → "X-Yaagents-Profile", so exact lookups fail.
            return resp.status, {k.lower(): v for k, v in resp.headers.items()}, resp.read()
    except urllib.error.HTTPError as e:
        return e.code, {k.lower(): v for k, v in e.headers.items()}, e.read()


def _ct_base(hdrs: dict[str, str]) -> str:
    ct = hdrs.get("Content-Type", hdrs.get("content-type", ""))
    return ct.split(";")[0].strip()


def _auth_headers(
    token: str,
    tenant: str = _TENANT,
    corr_id: str | None = None,
) -> dict[str, str]:
    h = {
        "Authorization": f"Bearer {token}",
        "X-Tenant-ID": tenant,
        "Content-Type": _CT_JSON,
    }
    if corr_id:
        h["X-Correlation-ID"] = corr_id
    return h


# ---------------------------------------------------------------------------
# Session fixture: create one campaign so sub-resource tests can use its id
# ---------------------------------------------------------------------------


@pytest.fixture(scope="module")
def campaign_id(gateway_live: str) -> str:
    """Create a campaign through the gateway; return its id for sub-resource tests."""
    token = make_jwt(DEMO_JWT_SECRET)
    body = json.dumps(
        {
            "name": "E2E Gate Campaign",
            "budget": 5000,
            "targetAudience": "integration-testers",
            "successMetric": "ctr",
        }
    ).encode()
    status, _, resp_bytes = _http(
        "POST",
        f"{gateway_live}/campaigns",
        headers=_auth_headers(token),
        body=body,
    )
    assert status == 201, f"fixture campaign creation failed: {status}"
    cid = json.loads(resp_bytes)["campaign"]["id"]
    return cid


# ---------------------------------------------------------------------------
# §12 criterion 9 — yaagents conformance-test PASS
# ---------------------------------------------------------------------------


def test_conformance_test_pass(gateway_live: str) -> None:
    """PRD §12 crit 9: `yaagents conformance-test <url>` → Overall: PASS."""
    result = conformance_test(gateway_live, jwt_secret=DEMO_JWT_SECRET)
    if not result.passed:
        detail = "\n".join(
            f"  [{c.name}] {c.detail}" for c in result.checks if not c.passed
        )
        pytest.fail(f"conformance-test FAIL:\n{detail}")


# ---------------------------------------------------------------------------
# §12 criterion 8 — Compose demo is running (gateway_live already proves this)
# ---------------------------------------------------------------------------


def test_compose_demo_both_services_healthy(gateway_live: str) -> None:
    """PRD §12 crit 8: docker compose up → gateway + campaign-api healthy."""
    # /healthz at the gateway proves both services are up (depends_on: campaign-api healthy)
    status, _, _ = _http("GET", f"{gateway_live}/healthz")
    assert status == 200, f"gateway /healthz returned {status}"


# ---------------------------------------------------------------------------
# §6.2 flow 1 — success / created through gateway
# ---------------------------------------------------------------------------


def test_flow1_created_through_gateway(gateway_live: str) -> None:
    """Flow 1 through gateway: POST /campaigns (full body) → 201 + X-YAAgents-Profile."""
    token = make_jwt(DEMO_JWT_SECRET)
    body = json.dumps(
        {
            "name": "Flow1 Campaign",
            "budget": 2000,
            "targetAudience": "all",
            "successMetric": "cpl",
        }
    ).encode()
    status, hdrs, _ = _http(
        "POST",
        f"{gateway_live}/campaigns",
        headers=_auth_headers(token),
        body=body,
    )
    assert status == 201, f"expected 201, got {status}"
    profile = hdrs.get(_PROFILE_HEADER.lower(), "")
    assert profile == "v0.1", f"X-YAAgents-Profile: {profile!r} (expected v0.1)"


# ---------------------------------------------------------------------------
# §6.2 flow 2 — clarification_required through gateway
# ---------------------------------------------------------------------------


def test_flow2_clarification_required_through_gateway(gateway_live: str) -> None:
    """Flow 2 through gateway: POST /campaigns (no successMetric) → 400 vendor CT."""
    token = make_jwt(DEMO_JWT_SECRET)
    body = json.dumps(
        {"name": "Flow2", "budget": 1000, "targetAudience": "all"}
    ).encode()
    status, hdrs, resp_bytes = _http(
        "POST",
        f"{gateway_live}/campaigns",
        headers=_auth_headers(token),
        body=body,
    )
    assert status == 400, f"expected 400, got {status}"
    ct = _ct_base(hdrs)
    assert ct == _CT_CLARIFICATION, f"Content-Type: {ct!r}"
    data = json.loads(resp_bytes)
    assert data["type"] == "clarification_required"
    inputs = data.get("requiredInputs", [])
    assert any(i["name"] == "successMetric" for i in inputs), (
        f"successMetric not in requiredInputs: {inputs}"
    )
    assert "trace" in data, "trace block missing from clarification body"


# ---------------------------------------------------------------------------
# §6.2 flow 3 — validation_failed through gateway
# ---------------------------------------------------------------------------


def test_flow3_validation_failed_through_gateway(
    gateway_live: str, campaign_id: str
) -> None:
    """Flow 3 through gateway: POST optimizations (empty objectives) → 422."""
    token = make_jwt(DEMO_JWT_SECRET)
    body = json.dumps({"objectives": []}).encode()
    status, _, _ = _http(
        "POST",
        f"{gateway_live}/campaigns/{campaign_id}/optimizations",
        headers=_auth_headers(token),
        body=body,
    )
    assert status == 422, f"expected 422 (validation_failed), got {status}"


# ---------------------------------------------------------------------------
# §6.2 flow 4 — failed_dependency through gateway (via /_demo/llm-down toggle)
# ---------------------------------------------------------------------------


def test_flow4_failed_dependency_through_gateway(
    gateway_live: str, campaign_id: str
) -> None:
    """Flow 4 through gateway: toggle LLM_DOWN → 424 vendor error body."""
    token = make_jwt(DEMO_JWT_SECRET)

    # Turn LLM_DOWN on
    on_status, _, _ = _http(
        "POST",
        f"{gateway_live}/_demo/llm-down?enabled=true",
        headers={"Authorization": f"Bearer {token}", "Content-Type": _CT_JSON},
    )
    assert on_status == 200, f"/_demo/llm-down?enabled=true returned {on_status}"

    try:
        body = json.dumps({"objectives": ["ctr"]}).encode()
        status, hdrs, resp_bytes = _http(
            "POST",
            f"{gateway_live}/campaigns/{campaign_id}/optimizations",
            headers=_auth_headers(token),
            body=body,
        )
        assert status == 424, f"expected 424 (failed_dependency), got {status}"
        ct = _ct_base(hdrs)
        assert ct == _CT_ERROR, f"Content-Type: {ct!r} (expected {_CT_ERROR!r})"
        data = json.loads(resp_bytes)
        assert data["type"] == "failed_dependency", f"type: {data.get('type')!r}"
        assert data["code"] == "LLM_UNAVAILABLE"
        assert "trace" in data
    finally:
        # Always reset the flag so other tests are not affected
        _http(
            "POST",
            f"{gateway_live}/_demo/llm-down?enabled=false",
            headers={"Authorization": f"Bearer {token}", "Content-Type": _CT_JSON},
        )


# ---------------------------------------------------------------------------
# §12 criterion 5 — RBAC 403 through gateway
# ---------------------------------------------------------------------------


def test_gateway_rbac_403(gateway_live: str, campaign_id: str) -> None:
    """PRD §12 crit 5: token missing campaign:optimize role → 403 vendor error."""
    token_no_role = make_jwt(DEMO_JWT_SECRET, roles=[])
    body = json.dumps({"objectives": ["ctr"]}).encode()
    status, hdrs, resp_bytes = _http(
        "POST",
        f"{gateway_live}/campaigns/{campaign_id}/optimizations",
        headers=_auth_headers(token_no_role),
        body=body,
    )
    assert status == 403, f"expected 403 (RBAC), got {status}"
    ct = _ct_base(hdrs)
    assert ct == _CT_ERROR, f"Content-Type: {ct!r} (expected {_CT_ERROR!r})"
    data = json.loads(resp_bytes)
    assert data.get("type") in ("forbidden", "error"), (
        f"body.type: {data.get('type')!r}"
    )


# ---------------------------------------------------------------------------
# §12 criterion 6 — correlation-id propagated through gateway
# ---------------------------------------------------------------------------


def test_gateway_correlation_id_propagated(gateway_live: str) -> None:
    """PRD §12 crit 6: X-Correlation-ID sent → echoed in response headers."""
    token = make_jwt(DEMO_JWT_SECRET)
    body = json.dumps(
        {
            "name": "CorrID Test",
            "budget": 100,
            "targetAudience": "robots",
            "successMetric": "conversion_rate",
        }
    ).encode()
    _, resp_hdrs, _ = _http(
        "POST",
        f"{gateway_live}/campaigns",
        headers=_auth_headers(token, corr_id=_CORR_ID),
        body=body,
    )
    echoed = resp_hdrs.get("x-correlation-id", "")
    assert echoed == _CORR_ID, f"sent {_CORR_ID!r}; got {echoed!r}"


# ---------------------------------------------------------------------------
# §12 criterion 7a — Python client handles clarification natively
# ---------------------------------------------------------------------------


def test_python_client_clarification_required(gateway_live: str) -> None:
    """PRD §12 crit 7: YaAgentsClient maps 400 vendor body → ClarificationRequired."""
    import httpx

    token = make_jwt(DEMO_JWT_SECRET)
    with httpx.Client(
        base_url=gateway_live,
        headers={
            "Authorization": f"Bearer {token}",
            "X-Tenant-ID": _TENANT,
        },
        timeout=_REQUEST_TIMEOUT,
    ) as http:
        resp = http.post(
            "/campaigns",
            json={"name": "PyClient Test", "budget": 500, "targetAudience": "devs"},
        )
    # process_response maps 400+clarification+json → ClarificationRequired exception
    with pytest.raises(ClarificationRequired) as exc_info:
        process_response(resp)
    exc = exc_info.value
    assert exc.status_code == 400
    assert exc.content_type == _CT_CLARIFICATION
    inputs = exc.required_inputs
    assert any(i["name"] == "successMetric" for i in inputs), (
        f"successMetric not in required_inputs: {inputs}"
    )


# ---------------------------------------------------------------------------
# §12 criterion 7b — TS client handles typed results natively (Node subprocess)
# ---------------------------------------------------------------------------


def test_ts_client_e2e(gateway_live: str) -> None:
    """PRD §12 crit 7: TS client discriminated union + strict ForbiddenError live."""
    node = "node"
    if not _TS_E2E_SCRIPT.exists():
        pytest.fail(f"TS e2e script not found: {_TS_E2E_SCRIPT}")
    result = subprocess.run(
        [node, str(_TS_E2E_SCRIPT), gateway_live],
        capture_output=True,
        text=True,
        timeout=60,
    )
    if result.returncode != 0:
        msg = result.stdout + "\n" + result.stderr
        pytest.fail(f"TS client e2e FAIL (exit {result.returncode}):\n{msg.strip()}")


# ---------------------------------------------------------------------------
# §4 table spec/-only grep proof (no live gateway needed)
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "vendor_type",
    [
        "application/vnd.yaagents.clarification+json",
        "application/vnd.yaagents.operation+json",
        "application/vnd.yaagents.validation-error+json",
        "application/vnd.yaagents.approval-required+json",
        "application/vnd.yaagents.error+json",
        "application/vnd.yaagents.conflict+json",
    ],
)
def test_spec_section4_vendor_types_defined_in_spec(vendor_type: str) -> None:
    """§4 table: each vendor media-type MUST appear in spec/ (normative source)."""
    assert _SPEC_DIR.exists(), f"spec/ directory not found at {_SPEC_DIR}"
    spec_files = list(_SPEC_DIR.rglob("*"))
    found = any(
        vendor_type in f.read_text(encoding="utf-8", errors="ignore")
        for f in spec_files
        if f.is_file()
    )
    assert found, f"vendor type {vendor_type!r} not found in spec/"


def test_spec_section4_table_not_redefined_in_component_source() -> None:
    """grep proof: §4 normative table not embedded in component source code.

    The normative status × media-type table lives in spec/ and the PRD
    (system-refs/).  Component *source code* (*.py, *.go, *.ts, *.js) must
    reference vendor types as constants, not re-embed a table definition.

    We check that no Python, Go, or TypeScript source file outside spec/ contains
    the markdown-table row pattern used in the normative table.

    Documentation (*.md, system-refs/, README files) is intentionally excluded —
    those may legitimately reference or reproduce the table for explanation.
    """
    repo_root = _REPO_ROOT
    # The normative row pattern is identifiable by the backtick-quoted response
    # type in a pipe-delimited table cell.
    _TABLE_PATTERN = "| `clarification_required` |"
    # Scan source code only (not docs/READMEs/PRDs)
    _SOURCE_GLOBS = ("**/*.py", "**/*.go", "**/*.ts", "**/*.js")
    # Exclude spec/ itself (that's the canonical source), node_modules, test dirs
    _SKIP_DIRS = {"spec", ".git", "node_modules", "dist", "__pycache__", ".venv"}
    _SKIP_SUFFIXES = {"_test.go", "_test.py", ".test.ts", ".test.js"}
    _SKIP_NAME_PARTS = {"test_", "tests/"}

    offenders: list[str] = []
    for glob_pattern in _SOURCE_GLOBS:
        for src_file in repo_root.glob(glob_pattern):
            rel = src_file.relative_to(repo_root)
            if rel.parts[0] in _SKIP_DIRS:
                continue
            try:
                content = src_file.read_text(encoding="utf-8", errors="ignore")
            except OSError:
                continue
            # skip test files (they may reference the pattern in assertions)
            name = src_file.name
            if any(name.endswith(s) for s in _SKIP_SUFFIXES):
                continue
            if any(p in str(rel).replace("\\", "/") for p in _SKIP_NAME_PARTS):
                continue
            if _TABLE_PATTERN in content:
                offenders.append(str(rel))

    assert not offenders, (
        "§4 normative table row found in component source (should live in spec/ only):\n"
        + "\n".join(f"  {f}" for f in offenders)
    )


# ---------------------------------------------------------------------------
# PRD §12 criteria 1–9 explicit checklist (cross-reference)
# ---------------------------------------------------------------------------


def test_prd_section12_criteria_checklist(gateway_live: str) -> None:
    """Explicit §12 criteria 1–9 cross-reference map (documentation test).

    Each criterion is mapped to the test(s) that verify it in this gate.
    This test always passes — it documents WHICH tests cover each criterion.
    """
    checklist = {
        "1. FastAPI SDK exposes agentic endpoint": "tests/test_flows.py::test_flow1_*",
        "2. Endpoint returns all 4 §6.2 flows": (
            "test_flow1_*, test_flow2_*, test_flow3_*, test_flow4_*"
        ),
        "3. SDK maps status+CT per §4": "tests/test_flows.py (vendor CT assertions)",
        "4. OpenAPI includes vendor content-types": (
            "tests/test_flows.py::test_openapi_json_passes_validate_openapi"
        ),
        "5. Gateway RBAC 403": "test_gateway_rbac_403",
        "6. Gateway propagates corr-id/tenant": (
            "test_gateway_correlation_id_propagated + conformance check 4+5"
        ),
        "7. Both clients handle clarification": (
            "test_python_client_clarification_required + test_ts_client_e2e"
        ),
        "8. Compose demo docker compose up": "gateway_live fixture (healthz up)",
        "9. conformance-test PASS": "test_conformance_test_pass",
    }
    # All criteria must be mapped — if any key is empty something is wrong
    for criterion, coverage in checklist.items():
        assert coverage, f"§12 criterion unmapped: {criterion}"
