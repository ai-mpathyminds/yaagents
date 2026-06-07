# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for financial-risk-screening agentic flows.

Covers:
  license-check gate (403)   — community tier rejected
  approval_required   (412)  — high-risk claim exceeds threshold
  created             (201)  — low-risk claim approved
  health endpoints
"""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from main import app  # pythonpath=["."] set in pyproject.toml

# Common headers for licensed callers
_PROF_HEADERS = {
    "X-Tenant-ID": "tenant-001",
    "X-License-Tier": "professional",
}
_ENT_HEADERS = {
    "X-Tenant-ID": "tenant-001",
    "X-License-Tier": "enterprise",
}


@pytest.fixture()
def client() -> TestClient:
    return TestClient(app, raise_server_exceptions=True)


# ── license-check gate (403) ──────────────────────────────────────────────────


def test_community_tier_returns_403(client: TestClient) -> None:
    """Community tier receives 403 application/vnd.yaagents.error+json."""
    r = client.post(
        "/claims/clm-001/risk-screens",
        json={"amount": 500, "claimant_history": "good"},
        headers={"X-Tenant-ID": "tenant-001", "X-License-Tier": "community"},
    )
    assert r.status_code == 403
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.error+json"
    )
    data = r.json()
    assert data["type"] == "forbidden"
    assert data["code"] == "LICENSE_TIER_INSUFFICIENT"
    assert "trace" in data


def test_missing_license_tier_defaults_to_community_403(
    client: TestClient,
) -> None:
    """No X-License-Tier header → defaults to community → 403."""
    r = client.post(
        "/claims/clm-002/risk-screens",
        json={"amount": 500, "claimant_history": "good"},
        headers={"X-Tenant-ID": "tenant-001"},
    )
    assert r.status_code == 403


# ── approval_required (412) ───────────────────────────────────────────────────


def test_approval_required_high_amount(client: TestClient) -> None:
    """High-amount neutral claim exceeds threshold → 412 approval_required."""
    r = client.post(
        "/claims/clm-010/risk-screens",
        json={"amount": 15000, "claimant_history": "neutral"},
        headers=_PROF_HEADERS,
    )
    assert r.status_code == 412
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.approval-required+json"
    )
    data = r.json()
    assert data["type"] == "approval_required"
    assert data["code"] == "APPROVAL_REQUIRED"
    assert "approvalToken" in data
    assert "trace" in data


def test_approval_required_bad_history(client: TestClient) -> None:
    """Adverse claimant history amplifies the risk score → 412."""
    r = client.post(
        "/claims/clm-011/risk-screens",
        json={"amount": 6000, "claimant_history": "bad"},
        headers=_ENT_HEADERS,
    )
    # 6000/10000 * 1.4 = 0.840 > 0.7
    assert r.status_code == 412


def test_approval_token_includes_claim_id(client: TestClient) -> None:
    """approvalToken includes the claimId for traceability."""
    r = client.post(
        "/claims/clm-trace/risk-screens",
        json={"amount": 15000, "claimant_history": "neutral"},
        headers=_PROF_HEADERS,
    )
    assert r.status_code == 412
    assert "clm-trace" in r.json()["approvalToken"]


# ── created (201) ─────────────────────────────────────────────────────────────


def test_created_low_risk_returns_201(client: TestClient) -> None:
    """Low-risk claim (good history, small amount) → 201 with screen result."""
    r = client.post(
        "/claims/clm-020/risk-screens",
        json={"amount": 500, "claimant_history": "good"},
        headers=_PROF_HEADERS,
    )
    assert r.status_code == 201
    assert r.headers["content-type"].startswith("application/json")
    data = r.json()
    assert "screen" in data
    assert data["screen"]["claimId"] == "clm-020"
    assert data["screen"]["status"] == "approved"
    assert data["screen"]["riskScore"] <= 0.7
    assert "trace" in data


def test_created_enterprise_tier_passes(client: TestClient) -> None:
    """Enterprise tier also passes the license gate."""
    r = client.post(
        "/claims/clm-021/risk-screens",
        json={"amount": 100, "claimant_history": "good"},
        headers=_ENT_HEADERS,
    )
    assert r.status_code == 201


def test_created_risk_score_in_response(client: TestClient) -> None:
    """riskScore is returned in the approved screen body."""
    r = client.post(
        "/claims/clm-022/risk-screens",
        json={"amount": 1000, "claimant_history": "good"},
        headers=_PROF_HEADERS,
    )
    assert r.status_code == 201
    assert isinstance(r.json()["screen"]["riskScore"], float)


# ── health endpoints ──────────────────────────────────────────────────────────


def test_healthz(client: TestClient) -> None:
    r = client.get("/healthz")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


def test_readyz(client: TestClient) -> None:
    r = client.get("/readyz")
    assert r.status_code == 200
    assert r.json()["profile"] == "v0.3"


# ── resource-oriented endpoint check (parity-review §11) ─────────────────────


def test_endpoint_is_resource_oriented(client: TestClient) -> None:
    """Route is /claims/{id}/risk-screens (resource sub-collection).

    Must NOT be /agents/risk-screening/invoke (parity-review §11).
    """
    r = client.post(
        "/claims/clm-999/risk-screens",
        json={"amount": 100, "claimant_history": "good"},
        headers=_PROF_HEADERS,
    )
    assert r.status_code != 404, (
        "Endpoint /claims/{id}/risk-screens must exist (resource-oriented, §11)"
    )
