# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Pydantic request/response models for the campaign-api example.

Uses Pydantic v2 model syntax (FastAPI >=0.111 ships Pydantic v2).
"""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, field_validator


class CreateCampaignRequest(BaseModel):
    """POST /campaigns request body."""

    name: str
    budget: float
    targetAudience: str
    # successMetric is intentionally optional — omitting it triggers
    # clarification_required (PRD §6.2 demo flow 2).
    successMetric: str | None = None

    @field_validator("budget")
    @classmethod
    def budget_positive(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("budget must be positive")
        return v


class CreateOptimizationRequest(BaseModel):
    """POST /campaigns/{id}/optimizations request body.

    ``objectives`` must be a list[str] — passing a plain string triggers
    validation_failed (PRD §6.2 demo flow 3).
    """

    objectives: list[str]
    maxSuggestions: int = 3

    @field_validator("objectives")
    @classmethod
    def objectives_non_empty(cls, v: list[str]) -> list[str]:
        if not v:
            raise ValueError("objectives must contain at least one item")
        return v


class GenerateAssetsRequest(BaseModel):
    """POST /campaigns/{id}/assets:generate request body."""

    assetTypes: list[str]
    tone: str = "professional"
    extraParams: dict[str, Any] = {}
