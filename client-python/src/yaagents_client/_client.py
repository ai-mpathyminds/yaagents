# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""YaAgentsClient — sync HTTP client for the YAAgents Agentic REST Profile v0.1."""

from __future__ import annotations

import uuid
from typing import Any

import httpx

from ._resources import CampaignResource

__all__ = ["YaAgentsClient"]

#: Profile version supported by this client build.
PROFILE_VERSION = "v0.1"


class YaAgentsClient:
    """Minimal synchronous client for the YAAgents Agentic REST Profile.

    Args:
        base_url: Root URL of the YAAgents gateway (trailing slash stripped).
        token: Bearer token for ``Authorization`` header.
        tenant_id: Tenant identifier injected as ``X-Tenant-ID`` on every request.
        _transport: Optional httpx transport for testing; not part of the public API.
    """

    def __init__(
        self,
        base_url: str,
        token: str,
        tenant_id: str,
        *,
        _transport: httpx.BaseTransport | None = None,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._token = token
        self._tenant_id = tenant_id
        self._http = httpx.Client(
            base_url=self._base_url,
            headers={
                "Authorization": f"Bearer {token}",
                "X-Tenant-ID": tenant_id,
            },
            transport=_transport,
        )

    # ------------------------------------------------------------------
    # Internal request helper (used by resource sub-classes)
    # ------------------------------------------------------------------

    def _request(
        self,
        method: str,
        path: str,
        body: dict[str, Any] | None = None,
        *,
        correlation_id: str | None = None,
    ) -> httpx.Response:
        """Send an authenticated request, auto-generating a correlation ID."""
        headers = {"X-Correlation-ID": correlation_id or str(uuid.uuid4())}
        return self._http.request(method, path, json=body, headers=headers)

    # ------------------------------------------------------------------
    # Resource factory
    # ------------------------------------------------------------------

    def campaigns(self, campaign_id: str) -> CampaignResource:
        """Return a fluent ``CampaignResource`` accessor for *campaign_id*."""
        return CampaignResource(client=self, campaign_id=campaign_id)

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def close(self) -> None:
        """Close the underlying HTTP connection pool."""
        self._http.close()

    def __enter__(self) -> YaAgentsClient:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()
