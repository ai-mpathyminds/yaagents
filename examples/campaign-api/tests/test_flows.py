# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for the four PRD §6.2 demo flows + endpoint smoke tests.

Covers:
  Flow 1 — success/created (POST /campaigns with all fields)
  Flow 2 — clarification_required (missing successMetric)
  Flow 3 — validation_failed (bad budget type / empty objectives)
  Flow 4 — failed_dependency (LLM_DOWN toggle)
  /openapi.json passes validate-openapi
"""

from __future__ import annotations

import json

import pytest
from fastapi.testclient import TestClient

from campaign_api.app import app

@pytest.fixture(autouse=True)
def reset_state() -> None:  # type: ignore[return]
    """Reset in-memory store and LLM flag between tests."""
    import campaign_api.store as s

    s._campaigns.clear()  # noqa: SIM117
    s._optimizations.clear()
    s.LLM_DOWN = False
    yield
    s._campaigns.clear()
    s._optimizations.clear()
    s.LLM_DOWN = False

@pytest.fixture()
def client() -> TestClient:
    return TestClient(app, raise_server_exceptions=True)

# ── helper ────────────────────────────────────────────────────────────────────

_VALID_CAMPAIGN = {
    "name": "Summer Sale",
    "budget": 5000.0,
    "targetAudience": "25-34 urban",
    "successMetric": "ctr",
}

def _post_campaign(client: TestClient, **overrides: object) -> object:
    body = {**_VALID_CAMPAIGN, **overrides}
    return client.post("/campaigns", json=body)

# ── flow 1: success / created ─────────────────────────────────────────────────

def test_flow1_create_campaign_returns_201(client: TestClient) -> None:
    r = _post_campaign(client)
    assert r.status_code == 201
    assert r.headers["content-type"].startswith("application/json")
    data = r.json()
    assert "campaign" in data
    assert data["campaign"]["successMetric"] == "ctr"
    # trace block present (correlation-id generated standalone)
    assert "trace" in data

def test_flow1_get_campaign_returns_200(client: TestClient) -> None:
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.get(f"/campaigns/{cid}")
    assert r.status_code == 200
    assert r.json()["campaign"]["id"] == cid

def test_flow1_create_optimization_returns_201(client: TestClient) -> None:
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.post(
        f"/campaigns/{cid}/optimizations",
        json={"objectives": ["ctr", "cpl"], "maxSuggestions": 2},
    )
    assert r.status_code == 201
    data = r.json()
    assert "optimization" in data
    assert len(data["optimization"]["suggestions"]) == 2

def test_flow1_get_optimization_returns_200(client: TestClient) -> None:
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    opt_r = client.post(
        f"/campaigns/{cid}/optimizations",
        json={"objectives": ["ctr"]},
    )
    oid = opt_r.json()["optimization"]["id"]
    r = client.get(f"/campaigns/{cid}/optimizations/{oid}")
    assert r.status_code == 200
    assert r.json()["optimization"]["id"] == oid

def test_flow1_generate_assets_returns_201(client: TestClient) -> None:
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.post(
        f"/campaigns/{cid}/assets:generate",
        json={"assetTypes": ["banner", "copy"], "tone": "friendly"},
    )
    assert r.status_code == 201
    data = r.json()
    assert len(data["assets"]) == 2
    assert data["assets"][0]["type"] == "banner"

# ── flow 2: clarification_required ───────────────────────────────────────────

def test_flow2_missing_success_metric_returns_400(client: TestClient) -> None:
    r = _post_campaign(client, successMetric=None)
    assert r.status_code == 400
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.clarification+json"
    )
    data = r.json()
    assert data["type"] == "clarification_required"
    assert data["code"] == "CLARIFICATION_REQUIRED"
    inputs = data["requiredInputs"]
    assert len(inputs) >= 1
    assert inputs[0]["name"] == "successMetric"
    assert "allowedValues" in inputs[0]
    assert "trace" in data

# ── flow 3: validation_failed ─────────────────────────────────────────────────

def test_flow3_negative_budget_returns_422(client: TestClient) -> None:
    """Pydantic field_validator rejects negative budget → 422."""
    r = _post_campaign(client, budget=-100)
    assert r.status_code == 422

def test_flow3_empty_objectives_returns_422(client: TestClient) -> None:
    """Empty objectives list fails validation → 422."""
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.post(
        f"/campaigns/{cid}/optimizations",
        json={"objectives": []},
    )
    assert r.status_code == 422

def test_flow3_bad_objectives_type_returns_422(client: TestClient) -> None:
    """objectives must be a list; sending a plain string → 422."""
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.post(
        f"/campaigns/{cid}/optimizations",
        json={"objectives": "ctr"},  # string instead of list
    )
    assert r.status_code == 422

# ── flow 4: failed_dependency ─────────────────────────────────────────────────

def test_flow4_optimization_llm_down_returns_424(client: TestClient) -> None:
    import campaign_api.store as s

    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    s.LLM_DOWN = True
    r = client.post(
        f"/campaigns/{cid}/optimizations",
        json={"objectives": ["ctr"]},
    )
    assert r.status_code == 424
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.error+json"
    )
    data = r.json()
    assert data["type"] == "failed_dependency"
    assert data["code"] == "LLM_UNAVAILABLE"
    assert "trace" in data

def test_flow4_assets_llm_down_returns_424(client: TestClient) -> None:
    import campaign_api.store as s

    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    s.LLM_DOWN = True
    r = client.post(
        f"/campaigns/{cid}/assets:generate",
        json={"assetTypes": ["banner"]},
    )
    assert r.status_code == 424
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.error+json"
    )

# ── health endpoints ──────────────────────────────────────────────────────────

def test_healthz(client: TestClient) -> None:
    r = client.get("/healthz")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"

def test_readyz(client: TestClient) -> None:
    r = client.get("/readyz")
    assert r.status_code == 200
    assert r.json()["profile"] == "v0.1"

# ── /openapi.json passes validate-openapi ─────────────────────────────────────

def test_openapi_json_passes_validate_openapi(
    tmp_path: object,
    client: TestClient,
) -> None:
    """Export /openapi.json and run yaagents-cli validate-openapi against it."""
    import pathlib

    from yaagents_cli._validate_openapi import validate_openapi

    r = client.get("/openapi.json")
    assert r.status_code == 200

    # Write to a temp file for the validator
    assert isinstance(tmp_path, pathlib.Path)
    openapi_file = tmp_path / "openapi.json"
    openapi_file.write_text(json.dumps(r.json()), encoding="utf-8")

    result = validate_openapi(str(openapi_file))
    if not result.passed:
        findings_txt = "\n".join(
            f"  {f.pointer}: {f.message}" for f in result.findings
        )
        pytest.fail(f"validate-openapi FAIL:\n{findings_txt}")

# ── 404 smoke tests ───────────────────────────────────────────────────────────

def test_get_nonexistent_campaign(client: TestClient) -> None:
    r = client.get("/campaigns/does-not-exist")
    assert r.status_code == 404

def test_get_nonexistent_optimization(client: TestClient) -> None:
    create_r = _post_campaign(client)
    cid = create_r.json()["campaign"]["id"]
    r = client.get(f"/campaigns/{cid}/optimizations/does-not-exist")
    assert r.status_code == 404
