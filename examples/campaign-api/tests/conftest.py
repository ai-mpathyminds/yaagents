# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Shared fixtures for campaign-api tests (gate)."""

from __future__ import annotations

import os
import urllib.error
import urllib.request

import pytest

_DEFAULT_GATEWAY = "http://localhost:8120"

def _is_gateway_up(base_url: str) -> bool:
    """Return True if the gateway healthz endpoint responds."""
    try:
        req = urllib.request.Request(f"{base_url}/healthz")
        with urllib.request.urlopen(req, timeout=3) as resp:
            return resp.status == 200
    except Exception:
        return False

@pytest.fixture(scope="session")
def gateway_base_url() -> str:
    """Base URL of the running gateway (env: GATEWAY_URL; default localhost:8120)."""
    return os.environ.get("GATEWAY_URL", _DEFAULT_GATEWAY).rstrip("/")

@pytest.fixture(scope="session")
def gateway_live(gateway_base_url: str) -> str:
    """Skip all tests in the session if the gateway is not reachable."""
    if not _is_gateway_up(gateway_base_url):
        pytest.skip(
            f"Gateway not reachable at {gateway_base_url} — "
            "start the Compose demo first: docker compose up --build -d"
        )
    return gateway_base_url
