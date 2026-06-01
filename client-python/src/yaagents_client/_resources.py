# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Fluent resource accessors for YAAgents Agentic REST Profile v0.1."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from ._client import YaAgentsClient

from ._mapper import process_response

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
    ) -> dict[str, Any]:
        """POST /campaigns/{id}/optimizations — create a new optimization run.

        Returns:
            Deserialized payload for 200/201/202 responses.

        Raises:
            ClarificationRequired: Agent needs additional inputs (HTTP 400).
            ValidationFailed: Request inputs failed validation (HTTP 422).
            AgenticForbidden: Caller lacks permission (HTTP 403).
            FailedDependency: Upstream service unavailable (HTTP 424).
            AgenticError: All other non-success responses.
        """
        response = self._client._request(
            "POST",
            f"/campaigns/{self._campaign_id}/optimizations",
            body,
            correlation_id=correlation_id,
        )
        return process_response(response)


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
    ) -> dict[str, Any]:
        """POST /campaigns/{id}/assets:generate — trigger asset generation.

        Returns:
            Deserialized payload for 200/201/202 responses.

        Raises:
            ClarificationRequired: Agent needs additional inputs (HTTP 400).
            ValidationFailed: Request inputs failed validation (HTTP 422).
            AgenticForbidden: Caller lacks permission (HTTP 403).
            FailedDependency: Upstream service unavailable (HTTP 424).
            AgenticError: All other non-success responses.
        """
        response = self._client._request(
            "POST",
            f"/campaigns/{self._campaign_id}/assets:generate",
            body,
            correlation_id=correlation_id,
        )
        return process_response(response)


class CampaignResource:
    """Fluent resource accessor for a single campaign."""

    def __init__(self, client: YaAgentsClient, campaign_id: str) -> None:
        self._client = client
        self._campaign_id = campaign_id
        self.optimizations: OptimizationsResource = OptimizationsResource(
            client, campaign_id
        )
        self.assets: AssetsResource = AssetsResource(client, campaign_id)
