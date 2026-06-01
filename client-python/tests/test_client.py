# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Tests for WI-1yaa.PYC-1 — YaAgentsClient + resource fluent accessors.

AC:
  - Headers injected on every request; correlation-id auto + overridable
  - Resource accessors build correct method + path + body
"""

from __future__ import annotations

import json
import uuid
from collections.abc import Callable

import httpx
import pytest

from yaagents_client import YaAgentsClient  # noqa: E402

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_HandlerFn = Callable[[httpx.Request], httpx.Response]


def _ok_transport(handler: _HandlerFn | None = None) -> httpx.MockTransport:
    """Return a MockTransport that replies 200 OK and records the request."""

    def _default(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"ok": True})

    return httpx.MockTransport(handler or _default)


def _capture_transport() -> tuple[list[httpx.Request], httpx.MockTransport]:
    """Return (captured_requests, transport) pair."""
    requests: list[httpx.Request] = []

    def handler(request: httpx.Request) -> httpx.Response:
        requests.append(request)
        return httpx.Response(200, json={"ok": True})

    return requests, httpx.MockTransport(handler)


def _make_client(**kwargs: object) -> YaAgentsClient:
    requests, transport = _capture_transport()
    client = YaAgentsClient(
        base_url="http://localhost:8120",
        token="test-token",
        tenant_id="tenant-abc",
        _transport=transport,
        **kwargs,  # type: ignore[arg-type]
    )
    return client


# ---------------------------------------------------------------------------
# Header injection tests
# ---------------------------------------------------------------------------


def test_authorization_header_injected() -> None:
    """Authorization: Bearer <token> is present on every request."""
    captured, transport = _capture_transport()
    with YaAgentsClient(
        "http://localhost:8120",
        "my-secret-token",
        "tenant-xyz",
        _transport=transport,
    ) as client:
        client.campaigns("c1").optimizations.create({"goal": "ctr"})

    assert len(captured) == 1
    req = captured[0]
    assert req.headers["authorization"] == "Bearer my-secret-token"


def test_tenant_id_header_injected() -> None:
    """X-Tenant-ID is present on every request."""
    captured, transport = _capture_transport()
    with YaAgentsClient(
        "http://localhost:8120",
        "tok",
        "tenant-99",
        _transport=transport,
    ) as client:
        client.campaigns("c2").assets.generate({"style": "bold"})

    req = captured[0]
    assert req.headers["x-tenant-id"] == "tenant-99"


def test_correlation_id_auto_generated() -> None:
    """X-Correlation-ID is auto-generated (valid UUID) when not supplied."""
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c1").optimizations.create({})

    corr = captured[0].headers["x-correlation-id"]
    # Must be a valid UUID
    parsed = uuid.UUID(corr)
    assert str(parsed) == corr


def test_correlation_id_unique_per_request() -> None:
    """Each request without an explicit correlation-id gets a fresh UUID."""
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c1").optimizations.create({})
        c.campaigns("c1").optimizations.create({})

    ids = [r.headers["x-correlation-id"] for r in captured]
    assert ids[0] != ids[1]


def test_correlation_id_overridable_optimizations() -> None:
    """Caller-supplied correlation-id is passed through unchanged (optimizations)."""
    custom = "my-trace-00000001"
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c1").optimizations.create({}, correlation_id=custom)

    assert captured[0].headers["x-correlation-id"] == custom


def test_correlation_id_overridable_assets() -> None:
    """Caller-supplied correlation-id is passed through unchanged (assets)."""
    custom = "my-trace-00000002"
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c2").assets.generate({}, correlation_id=custom)

    assert captured[0].headers["x-correlation-id"] == custom


def test_headers_on_assets_generate_too() -> None:
    """All three default headers are present on assets:generate request."""
    captured, transport = _capture_transport()
    with YaAgentsClient(
        "http://localhost:8120", "bearer-tok", "t-42", _transport=transport
    ) as c:
        c.campaigns("camp").assets.generate({"format": "png"})

    req = captured[0]
    assert req.headers["authorization"] == "Bearer bearer-tok"
    assert req.headers["x-tenant-id"] == "t-42"
    assert "x-correlation-id" in req.headers


# ---------------------------------------------------------------------------
# Resource accessor / routing tests
# ---------------------------------------------------------------------------


def test_optimizations_create_method_and_path() -> None:
    """optimizations.create() sends POST /campaigns/{id}/optimizations."""
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("camp-001").optimizations.create({"goal": "cpl"})

    req = captured[0]
    assert req.method == "POST"
    assert req.url.path == "/campaigns/camp-001/optimizations"


def test_assets_generate_method_and_path() -> None:
    """assets.generate() sends POST /campaigns/{id}/assets:generate."""
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("camp-002").assets.generate({"style": "minimal"})

    req = captured[0]
    assert req.method == "POST"
    assert req.url.path == "/campaigns/camp-002/assets:generate"


def test_optimizations_create_sends_body() -> None:
    """Request body is serialised as JSON and sent verbatim."""
    body = {"goal": "conversion_rate", "budget": 500}
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c3").optimizations.create(body)

    req = captured[0]
    assert req.headers["content-type"].startswith("application/json")
    assert json.loads(req.content) == body


def test_assets_generate_sends_body() -> None:
    """Request body for assets:generate is serialised as JSON."""
    body = {"format": "svg", "theme": "dark"}
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("c4").assets.generate(body)

    req = captured[0]
    assert req.headers["content-type"].startswith("application/json")
    assert json.loads(req.content) == body


def test_campaigns_returns_campaign_resource() -> None:
    """campaigns() returns a CampaignResource with the correct sub-resources."""
    _, transport = _capture_transport()
    client = YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport)
    try:
        cr = client.campaigns("c5")
        from yaagents_client._resources import (
            AssetsResource,
            CampaignResource,
            OptimizationsResource,
        )

        assert isinstance(cr, CampaignResource)
        assert isinstance(cr.optimizations, OptimizationsResource)
        assert isinstance(cr.assets, AssetsResource)
        assert cr._campaign_id == "c5"
    finally:
        client.close()


def test_different_campaign_ids_route_independently() -> None:
    """Two CampaignResource instances with different IDs do not share paths."""
    captured, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport) as c:
        c.campaigns("alpha").optimizations.create({})
        c.campaigns("beta").optimizations.create({})

    paths = [r.url.path for r in captured]
    assert paths[0] == "/campaigns/alpha/optimizations"
    assert paths[1] == "/campaigns/beta/optimizations"


# ---------------------------------------------------------------------------
# Context manager lifecycle
# ---------------------------------------------------------------------------


def test_context_manager_closes_cleanly() -> None:
    """__enter__/__exit__ work; no error raised on normal exit."""
    _, transport = _capture_transport()
    with YaAgentsClient("http://localhost:8120", "t", "ten", _transport=transport):
        pass  # should not raise


def test_base_url_trailing_slash_stripped() -> None:
    """Trailing slash on base_url does not double-slash the path."""
    captured, transport = _capture_transport()
    with YaAgentsClient(
        "http://localhost:8120/", "t", "ten", _transport=transport
    ) as c:
        c.campaigns("c6").optimizations.create({})

    assert captured[0].url.path == "/campaigns/c6/optimizations"


# ---------------------------------------------------------------------------
# Package metadata
# ---------------------------------------------------------------------------


def test_profile_version_exposed() -> None:
    """PROFILE_VERSION constant is 'v0.1'."""
    from yaagents_client import PROFILE_VERSION

    assert PROFILE_VERSION == "v0.1"


def test_package_version_present() -> None:
    """__version__ is importable."""
    from yaagents_client import __version__

    assert __version__  # non-empty string


@pytest.mark.parametrize("name", ["YaAgentsClient", "CampaignResource", "__version__"])
def test_public_api_exportable(name: str) -> None:
    """Key symbols are accessible from the top-level package."""
    import yaagents_client

    assert hasattr(yaagents_client, name)
