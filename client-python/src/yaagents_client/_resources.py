"""Fluent resource accessors for YAAgents Agentic REST Profile v0.1."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

import httpx

if TYPE_CHECKING:
    from ._client import YaAgentsClient

__all__ = ["CampaignResource"]


class OptimizationsResource:
    """Optimizations sub-resource scoped to one campaign."""

    def __init__(self, client: YaAgentsClient, campaign_id: str) -> None:
        self._client = client
        self._campaign_id = campaign_id

    def create(
        self,
        body: dict[str, Any],
        *,
        correlation_id: str | None = None,
    ) -> httpx.Response:
        """POST /campaigns/{id}/optimizations — create a new optimization run."""
        return self._client._request(
            "POST",
            f"/campaigns/{self._campaign_id}/optimizations",
            body,
            correlation_id=correlation_id,
        )


class AssetsResource:
    """Assets sub-resource scoped to one campaign."""

    def __init__(self, client: YaAgentsClient, campaign_id: str) -> None:
        self._client = client
        self._campaign_id = campaign_id

    def generate(
        self,
        body: dict[str, Any],
        *,
        correlation_id: str | None = None,
    ) -> httpx.Response:
        """POST /campaigns/{id}/assets:generate — trigger asset generation."""
        return self._client._request(
            "POST",
            f"/campaigns/{self._campaign_id}/assets:generate",
            body,
            correlation_id=correlation_id,
        )


class CampaignResource:
    """Fluent resource accessor for a single campaign."""

    def __init__(self, client: YaAgentsClient, campaign_id: str) -> None:
        self._client = client
        self._campaign_id = campaign_id
        self.optimizations: OptimizationsResource = OptimizationsResource(
            client, campaign_id
        )
        self.assets: AssetsResource = AssetsResource(client, campaign_id)
