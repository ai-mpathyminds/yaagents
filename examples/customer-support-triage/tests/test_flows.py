# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for customer-support-triage agentic flows.

Covers:
  clarification_required (400) — missing severity and/or category
  created / high-severity  (201) — escalated to level-2 support
  created / low-severity   (201) — autoresolve hint returned
  health endpoints
"""

from __future__ import annotations

import pytest
from fastapi.testclient import TestClient

from main import app  # pythonpath=["."] set in pyproject.toml


@pytest.fixture()
def client() -> TestClient:
    return TestClient(app, raise_server_exceptions=True)


# ── clarification_required ────────────────────────────────────────────────────


def test_clarification_required_empty_body(client: TestClient) -> None:
    """Empty body triggers 400 clarification_required for both missing fields."""
    r = client.post(
        "/tickets/t-001:triage",
        json={},
        headers={"X-Tenant-ID": "tenant-001"},
    )
    assert r.status_code == 400
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.clarification+json"
    )
    data = r.json()
    assert data["type"] == "clarification_required"
    assert data["code"] == "CLARIFICATION_REQUIRED"
    names = {inp["name"] for inp in data["requiredInputs"]}
    assert "severity" in names
    assert "category" in names
    assert "trace" in data


def test_clarification_required_missing_severity(client: TestClient) -> None:
    """Missing severity triggers clarification_required."""
    r = client.post(
        "/tickets/t-002:triage",
        json={"category": "billing"},
        headers={"X-Tenant-ID": "tenant-001"},
    )
    assert r.status_code == 400
    assert r.headers["content-type"].startswith(
        "application/vnd.yaagents.clarification+json"
    )
    data = r.json()
    names = {inp["name"] for inp in data["requiredInputs"]}
    assert "severity" in names
    assert "category" not in names


def test_clarification_required_missing_category(client: TestClient) -> None:
    """Missing category triggers clarification_required."""
    r = client.post(
        "/tickets/t-003:triage",
        json={"severity": "low"},
        headers={"X-Tenant-ID": "tenant-001"},
    )
    assert r.status_code == 400
    data = r.json()
    names = {inp["name"] for inp in data["requiredInputs"]}
    assert "category" in names
    assert "severity" not in names


def test_clarification_required_inputs_have_allowed_values(
    client: TestClient,
) -> None:
    """requiredInputs entries carry allowedValues."""
    r = client.post("/tickets/t-004:triage", json={})
    data = r.json()
    for inp in data["requiredInputs"]:
        assert "allowedValues" in inp
        assert len(inp["allowedValues"]) >= 2


# ── created — high severity ───────────────────────────────────────────────────


def test_created_high_severity_returns_201(client: TestClient) -> None:
    """Valid high-severity body returns 201 with escalated status."""
    r = client.post(
        "/tickets/t-010:triage",
        json={"severity": "high", "category": "technical"},
        headers={"X-Tenant-ID": "tenant-001"},
    )
    assert r.status_code == 201
    assert r.headers["content-type"].startswith("application/json")
    data = r.json()
    assert "ticket" in data
    assert data["ticket"]["ticketId"] == "t-010"
    assert data["ticket"]["status"] == "escalated"
    assert "recommendedOwner" in data["ticket"]
    assert data["ticket"]["priority"] == "P1"
    assert "trace" in data


def test_created_high_severity_preserves_category(client: TestClient) -> None:
    """Category is echoed back in the triage result."""
    r = client.post(
        "/tickets/t-011:triage",
        json={"severity": "high", "category": "billing"},
    )
    assert r.status_code == 201
    assert r.json()["ticket"]["category"] == "billing"


# ── created — low / medium severity ──────────────────────────────────────────


def test_created_low_severity_returns_201_autoresolve(
    client: TestClient,
) -> None:
    """Valid low-severity body returns 201 with autoresolve hint."""
    r = client.post(
        "/tickets/t-020:triage",
        json={"severity": "low", "category": "shipping"},
    )
    assert r.status_code == 201
    data = r.json()
    assert data["ticket"]["status"] == "auto-resolved"
    assert "autoresolveHint" in data["ticket"]
    assert data["ticket"]["priority"] == "P3"


def test_created_medium_severity_returns_201_autoresolve(
    client: TestClient,
) -> None:
    """Medium severity also falls into the autoresolve path."""
    r = client.post(
        "/tickets/t-021:triage",
        json={"severity": "medium", "category": "technical"},
    )
    assert r.status_code == 201
    assert r.json()["ticket"]["status"] == "auto-resolved"


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
    """Route is /tickets/{id}:triage (resource + colon-action), not /agents/invoke."""
    # If the app were using /agents/triage/invoke, this path would 404.
    r = client.post(
        "/tickets/t-999:triage",
        json={"severity": "low", "category": "billing"},
    )
    assert r.status_code != 404, (
        "Endpoint /tickets/{id}:triage must exist (resource-oriented, §11)"
    )
