# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Deterministic in-memory state store for the campaign-api reference example.

Intentionally simple: plain dicts protected by a single module-level lock.
No database, no persistence — state resets on process restart.

Toggling ``LLM_DOWN = True`` causes all optimizations to return
``failed_dependency`` (PRD §6.2 demo flow 4).
"""

from __future__ import annotations

import threading
import uuid
from typing import Any

# ── demo toggle: set to True to simulate LLM-down (flow 4) ───────────────────
LLM_DOWN: bool = False

_lock = threading.Lock()

# { campaign_id: { id, name, budget, targetAudience, successMetric?, status } }
_campaigns: dict[str, dict[str, Any]] = {}

# { campaign_id: { optimization_id: { id, status, suggestions? } } }
_optimizations: dict[str, dict[str, dict[str, Any]]] = {}


# ── campaigns ─────────────────────────────────────────────────────────────────


def create_campaign(name: str, budget: float, target_audience: str,
                    success_metric: str | None) -> dict[str, Any]:
    with _lock:
        cid = str(uuid.uuid4())
        record: dict[str, Any] = {
            "id": cid,
            "name": name,
            "budget": budget,
            "targetAudience": target_audience,
            "status": "active",
        }
        if success_metric is not None:
            record["successMetric"] = success_metric
        _campaigns[cid] = record
        return dict(record)


def get_campaign(campaign_id: str) -> dict[str, Any] | None:
    with _lock:
        rec = _campaigns.get(campaign_id)
        return dict(rec) if rec else None


# ── optimizations ─────────────────────────────────────────────────────────────


def create_optimization(campaign_id: str,
                        suggestions: list[dict[str, Any]]) -> dict[str, Any]:
    with _lock:
        oid = str(uuid.uuid4())
        record: dict[str, Any] = {
            "id": oid,
            "campaignId": campaign_id,
            "status": "completed",
            "suggestions": suggestions,
        }
        _optimizations.setdefault(campaign_id, {})[oid] = record
        return dict(record)


def get_optimization(campaign_id: str,
                     op_id: str) -> dict[str, Any] | None:
    with _lock:
        ops = _optimizations.get(campaign_id, {})
        rec = ops.get(op_id)
        return dict(rec) if rec else None
