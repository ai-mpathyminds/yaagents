# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Validate examples/campaign-api/routes.yaml against WI-1yaa.EX-2 AC.

These tests do NOT require the gateway binary — they parse the YAML
directly and assert the schema constraints the gateway enforces at load time
(routes.Load logic, gateway/internal/routes/routes.go).

AC verified:
  - All 5 §6.1 paths present with correct method + target
  - optimizations: roles:[campaign:optimize], tenantRequired:true, audit:true
  - All routes have tenantRequired:true
  - Non-optimizations routes have audit:false (default)
  - YAML loads without schema error (ids unique, methods valid, targets valid URLs,
    paths well-formed)
"""

from __future__ import annotations

import pathlib
import re

import pytest
import yaml

ROUTES_YAML = (
    pathlib.Path(__file__).parent.parent / "routes.yaml"
)

# Expected §6.1 paths  →  (method, target_host)
_EXPECTED_PATHS = {
    "/campaigns": ("POST", "http://campaign-api:8121"),
    "/campaigns/{campaignId}": ("GET", "http://campaign-api:8121"),
    "/campaigns/{campaignId}/optimizations": ("POST", "http://campaign-api:8121"),
    "/campaigns/{campaignId}/optimizations/{opId}": (
        "GET",
        "http://campaign-api:8121",
    ),
    "/campaigns/{campaignId}/assets:generate": ("POST", "http://campaign-api:8121"),
}

_OPTIMIZATION_PATHS = {
    "/campaigns/{campaignId}/optimizations",
    "/campaigns/{campaignId}/optimizations/{opId}",
}

# Demo-only routes (prefixed /_demo/) are excluded from §6.1 profile assertions.
# They are intentionally NOT part of the YAAgents Profile and may have relaxed
# tenantRequired / audit settings (EX-4 flow-4 gate support).
_DEMO_PATH_PREFIX = "/_demo/"


def _profile_routes(routes: list[dict]) -> list[dict]:  # type: ignore[type-arg]
    """Return only the §6.1 profile routes (exclude /_demo/* internal routes)."""
    return [r for r in routes if not r.get("path", "").startswith(_DEMO_PATH_PREFIX)]

_PLACEHOLDER_RE = re.compile(r"\{([^{}]+)\}")
_VALID_METHODS = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}


# ── fixture ───────────────────────────────────────────────────────────────────


@pytest.fixture(scope="module")
def route_list() -> list[dict]:  # type: ignore[type-arg]
    data = yaml.safe_load(ROUTES_YAML.read_text(encoding="utf-8"))
    assert isinstance(data, dict), "routes.yaml root must be a mapping"
    assert "routes" in data, "routes.yaml must have a top-level 'routes' key"
    routes = data["routes"]
    assert isinstance(routes, list), "'routes' must be a list"
    return routes  # type: ignore[return-value]


# ── schema / uniqueness checks ────────────────────────────────────────────────


def test_ids_are_unique(route_list: list[dict]) -> None:  # type: ignore[type-arg]
    ids = [r["id"] for r in route_list]
    assert len(ids) == len(set(ids)), f"Duplicate route IDs: {ids}"


def test_all_routes_have_required_fields(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    for r in route_list:
        assert r.get("id"), f"Missing id on route: {r}"
        assert r.get("method"), f"Missing method on route {r['id']}"
        assert r.get("path"), f"Missing path on route {r['id']}"
        assert r.get("target"), f"Missing target on route {r['id']}"


def test_methods_are_valid(route_list: list[dict]) -> None:  # type: ignore[type-arg]
    for r in route_list:
        assert r["method"].upper() in _VALID_METHODS, (
            f"Route {r['id']}: invalid method {r['method']!r}"
        )


def test_targets_are_valid_urls(route_list: list[dict]) -> None:  # type: ignore[type-arg]
    import urllib.parse

    for r in route_list:
        parsed = urllib.parse.urlparse(r["target"])
        assert parsed.scheme in ("http", "https"), (
            f"Route {r['id']}: target {r['target']!r} is not an absolute URL"
        )
        assert parsed.netloc, (
            f"Route {r['id']}: target {r['target']!r} missing netloc"
        )


def test_path_placeholders_are_non_empty(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    for r in route_list:
        for m in _PLACEHOLDER_RE.finditer(r["path"]):
            assert m.group(1), (
                f"Route {r['id']}: empty placeholder {{}} in path {r['path']!r}"
            )


# ── §6.1 path coverage ────────────────────────────────────────────────────────


def test_all_section61_paths_present(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Every PRD §6.1 path + method is present in routes.yaml."""
    by_path_method = {
        (r["path"], r["method"].upper()): r for r in route_list
    }
    for path, (method, _) in _EXPECTED_PATHS.items():
        key = (path, method)
        assert key in by_path_method, (
            f"Missing route for {method} {path}"
        )


def test_all_targets_point_to_campaign_api(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    for r in route_list:
        assert r["target"] == "http://campaign-api:8121", (
            f"Route {r['id']}: target must be http://campaign-api:8121, "
            f"got {r['target']!r}"
        )


# ── RBAC checks (§6.2 gateway-RBAC 403 demo flow) ────────────────────────────


def test_optimization_routes_have_campaign_optimize_role(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Both optimization routes require the 'campaign:optimize' role."""
    opt_routes = [r for r in route_list if r["path"] in _OPTIMIZATION_PATHS]
    assert len(opt_routes) == 2, (
        f"Expected 2 optimization routes, found {len(opt_routes)}"
    )
    for r in opt_routes:
        roles = r.get("roles") or []
        assert "campaign:optimize" in roles, (
            f"Route {r['id']}: must require role 'campaign:optimize', "
            f"got roles={roles!r}"
        )


def test_non_optimization_routes_have_no_roles(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Non-optimization routes are open to all authenticated callers (no roles)."""
    for r in route_list:
        if r["path"] not in _OPTIMIZATION_PATHS:
            roles = r.get("roles") or []
            assert not roles, (
                f"Route {r['id']}: expected no roles, got {roles!r}"
            )


# ── tenantRequired checks (§6.2 tenant-missing rejection demo) ───────────────


def test_all_routes_have_tenant_required(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Every §6.1 profile route requires a tenant header.

    Demo-only /_demo/* routes are excluded — they are internal toggle endpoints
    not subject to the YAAgents Profile tenant enforcement requirement.
    """
    for r in _profile_routes(route_list):
        assert r.get("tenantRequired") is True, (
            f"Route {r['id']}: tenantRequired must be true, "
            f"got {r.get('tenantRequired')!r}"
        )


# ── audit checks (§6.2 audit event demo) ────────────────────────────────────


def test_optimization_routes_have_audit_true(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Optimization routes emit audit events."""
    opt_routes = [r for r in route_list if r["path"] in _OPTIMIZATION_PATHS]
    for r in opt_routes:
        assert r.get("audit") is True, (
            f"Route {r['id']}: audit must be true, got {r.get('audit')!r}"
        )


def test_non_optimization_routes_have_audit_false(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Non-optimization §6.1 routes do not emit audit events by default."""
    for r in _profile_routes(route_list):
        if r["path"] not in _OPTIMIZATION_PATHS:
            # audit defaults to false when absent
            audit_val = r.get("audit", False)
            assert audit_val is False, (
                f"Route {r['id']}: audit should be false, got {audit_val!r}"
            )


# ── route count sanity check ─────────────────────────────────────────────────


def test_exactly_five_profile_routes(
    route_list: list[dict],  # type: ignore[type-arg]
) -> None:
    """Exactly 5 §6.1 profile routes (demo-only /_demo/* routes excluded)."""
    profile = _profile_routes(route_list)
    assert len(profile) == 5, (
        f"Expected exactly 5 §6.1 routes, found {len(profile)}: "
        f"{[r['path'] for r in profile]}"
    )
