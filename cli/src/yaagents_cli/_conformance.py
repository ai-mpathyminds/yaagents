# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""conformance-test for WI-1yaa.CLI-4.

Exercises the live YAAgents gateway at ``<base-url>``, asserts mandatory
response-type behaviour, and emits the PRD §5.8 report format:

  YAAgents Conformance Report

  ✓ X-YAAgents-Profile header on proxied response
  ✓ Clarification response uses correct content type
  ✓ 400 response matches clarification schema
  ✓ Correlation ID propagated
  ✓ Gateway route requires tenant context

  Overall: PASS

Non-zero exit when any check fails.  Uses only Python stdlib — no heavy deps.

Demo-mode JWT is minted inline with HMAC-SHA256 (stdlib ``hmac`` + ``hashlib``).
The HS256 secret defaults to the Compose demo value; pass ``--jwt-secret`` to
override for a custom deployment.
"""

from __future__ import annotations

import base64
import hashlib
import hmac as _hmac
import json
import pathlib
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass, field
from typing import Any

import jsonschema

# ── profile constants (spec §2 + §4) ─────────────────────────────────────────

_PROFILE_HEADER = "X-YAAgents-Profile"
_PROFILE_VERSION = "v0.2"
_CT_CLARIFICATION = "application/vnd.yaagents.clarification+json"
_CT_ERROR = "application/vnd.yaagents.error+json"
_CT_JSON = "application/json"

# Default JWT secret for the Compose demo (GATEWAY_JWT_SECRET in docker-compose.yml).
# Never use in production — see compose security note.
DEMO_JWT_SECRET = "demo-secret-not-for-production"

_DEFAULT_TENANT = "t-conformance"
_CORR_ID_SENTINEL = "yaagents-conformance-test-id"
_REQUEST_TIMEOUT = 10  # seconds

# Bundled schema directory (same layout as _validate.py)
_SCHEMA_DIR = pathlib.Path(__file__).parent / "_schemas" / "v0.1"


# ── data classes ──────────────────────────────────────────────────────────────


@dataclass
class CheckResult:
    """Outcome of a single conformance check."""

    name: str
    passed: bool
    detail: str = ""


@dataclass
class ConformanceResult:
    """Aggregate result of all conformance checks."""

    checks: list[CheckResult] = field(default_factory=list)
    error: str | None = None

    @property
    def passed(self) -> bool:
        return self.error is None and bool(self.checks) and all(
            c.passed for c in self.checks
        )

    def to_dict(self) -> dict[str, Any]:
        result: dict[str, Any] = {
            "result": "PASS" if self.passed else "FAIL",
            "checks": [
                {
                    "name": c.name,
                    "result": "PASS" if c.passed else "FAIL",
                    **({"detail": c.detail} if c.detail else {}),
                }
                for c in self.checks
            ],
        }
        if self.error:
            result["error"] = self.error
        return result


# ── JWT helpers (stdlib only, no PyJWT dep) ───────────────────────────────────


def _b64url(data: bytes) -> str:
    """URL-safe base64 with no padding — RFC 4648 §5."""
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()


def make_jwt(
    secret: str,
    roles: list[str] | None = None,
    exp_delta: int = 300,
) -> str:
    """Mint a minimal HS256 JWT signed with *secret*.

    Valid for *exp_delta* seconds (default 5 min).  ``roles`` defaults to
    ``["campaign:optimize"]`` (the role required by the demo optimizations route).
    """
    header = _b64url(json.dumps({"alg": "HS256", "typ": "JWT"}).encode())
    payload = _b64url(
        json.dumps(
            {
                "sub": "conformance-tester",
                "roles": (
                    roles if roles is not None else ["campaign:optimize"]
                ),
                "exp": int(time.time()) + exp_delta,
            }
        ).encode()
    )
    signing_input = f"{header}.{payload}"
    sig = _hmac.new(
        secret.encode(),
        signing_input.encode(),
        hashlib.sha256,
    ).digest()
    return f"{signing_input}.{_b64url(sig)}"


# ── HTTP helper ───────────────────────────────────────────────────────────────


class ConformanceError(Exception):
    """Raised when a fatal pre-check fails (bad URL scheme, connection refused)."""


def _http(
    method: str,
    url: str,
    *,
    headers: dict[str, str] | None = None,
    body: bytes | None = None,
    timeout: int = _REQUEST_TIMEOUT,
) -> tuple[int, dict[str, str], bytes]:
    """Make an HTTP request; return ``(status, headers, body)``.

    HTTP 4xx/5xx responses are NOT raised as exceptions — the caller inspects
    the status code.  Network/connection errors raise :class:`ConformanceError`.
    """
    req = urllib.request.Request(url, data=body, method=method)
    for k, v in (headers or {}).items():
        req.add_header(k, v)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            # Normalise header keys to lowercase so lookups are case-insensitive.
            # Go's net/http canonicalises headers (e.g. "X-YAAgents-Profile" →
            # "X-Yaagents-Profile"), which breaks exact-case dict.get() calls.
            hdrs = {k.lower(): v for k, v in resp.headers.items()}
            return resp.status, hdrs, resp.read()
    except urllib.error.HTTPError as e:
        return e.code, {k.lower(): v for k, v in e.headers.items()}, e.read()
    except urllib.error.URLError as e:
        raise ConformanceError(f"connection failed: {e.reason}") from e


# ── schema helpers ────────────────────────────────────────────────────────────


def _load_schema(name: str) -> Any:
    path = _SCHEMA_DIR / name
    return json.loads(path.read_text(encoding="utf-8"))


def _validate_body(
    body_bytes: bytes, schema_name: str
) -> list[str]:
    """Return a list of error messages; empty list means valid."""
    try:
        body = json.loads(body_bytes.decode("utf-8"))
    except (json.JSONDecodeError, UnicodeDecodeError) as exc:
        return [f"body is not valid JSON: {exc}"]
    schema = _load_schema(schema_name)
    validator = jsonschema.Draft7Validator(schema)
    return [
        e.message
        for e in sorted(validator.iter_errors(body), key=lambda e: e.json_path)
    ]


def _ct_base(headers: dict[str, str]) -> str:
    """Extract the media-type portion of a Content-Type header (strip params)."""
    ct = headers.get("Content-Type", headers.get("content-type", ""))
    return ct.split(";")[0].strip()


# ── individual checks ─────────────────────────────────────────────────────────


def _check_profile_header(
    base: str,
    token: str,
    tenant: str,
) -> CheckResult:
    """Check 1: X-YAAgents-Profile header on a proxied response (spec §2)."""
    name = "X-YAAgents-Profile header on proxied response"
    payload = json.dumps(
        {
            "name": "Conformance Test Campaign",
            "budget": 1000,
            "targetAudience": "developers",
            "successMetric": "ctr",
        }
    ).encode()
    hdrs = {
        "Authorization": f"Bearer {token}",
        "X-Tenant-ID": tenant,
        "Content-Type": _CT_JSON,
    }
    status, resp_hdrs, _ = _http(
        "POST", f"{base}/campaigns", headers=hdrs, body=payload
    )
    got = resp_hdrs.get(_PROFILE_HEADER, resp_hdrs.get(_PROFILE_HEADER.lower(), ""))
    if got == _PROFILE_VERSION:
        return CheckResult(name=name, passed=True)
    detail = (
        f"status {status}; "
        f"{_PROFILE_HEADER}: {got!r} (expected {_PROFILE_VERSION!r})"
    )
    return CheckResult(name=name, passed=False, detail=detail)


def _check_clarification_ct(
    base: str,
    token: str,
    tenant: str,
) -> tuple[CheckResult, bytes]:
    """Check 2: POST /campaigns without successMetric → 400 vendor content-type."""
    name = "Clarification response uses correct content type"
    payload = json.dumps(
        {"name": "Test", "budget": 500, "targetAudience": "all"}
    ).encode()
    hdrs = {
        "Authorization": f"Bearer {token}",
        "X-Tenant-ID": tenant,
        "Content-Type": _CT_JSON,
    }
    status, resp_hdrs, body = _http(
        "POST", f"{base}/campaigns", headers=hdrs, body=payload
    )
    ct = _ct_base(resp_hdrs)
    if status == 400 and ct == _CT_CLARIFICATION:
        return CheckResult(name=name, passed=True), body
    detail = f"status {status}; Content-Type: {ct!r} (expected {_CT_CLARIFICATION!r})"
    return CheckResult(name=name, passed=False, detail=detail), body


def _check_clarification_schema(body: bytes) -> CheckResult:
    """Check 3: 400 clarification body matches clarification-required schema."""
    name = "400 response matches clarification schema"
    errors = _validate_body(body, "clarification-required.schema.json")
    if not errors:
        return CheckResult(name=name, passed=True)
    detail = "; ".join(errors[:3])  # cap detail length
    return CheckResult(name=name, passed=False, detail=detail)


def _check_correlation_id(
    base: str,
    token: str,
    tenant: str,
) -> CheckResult:
    """Check 4: X-Correlation-ID sent in request is echoed in response (spec §5)."""
    name = "Correlation ID propagated"
    payload = json.dumps(
        {
            "name": "Corr-ID Test Campaign",
            "budget": 100,
            "targetAudience": "robots",
            "successMetric": "conversion_rate",
        }
    ).encode()
    hdrs = {
        "Authorization": f"Bearer {token}",
        "X-Tenant-ID": tenant,
        "X-Correlation-ID": _CORR_ID_SENTINEL,
        "Content-Type": _CT_JSON,
    }
    _, resp_hdrs, _ = _http("POST", f"{base}/campaigns", headers=hdrs, body=payload)
    echoed = resp_hdrs.get(
        "X-Correlation-ID", resp_hdrs.get("x-correlation-id", "")
    )
    if echoed == _CORR_ID_SENTINEL:
        return CheckResult(name=name, passed=True)
    detail = f"sent {_CORR_ID_SENTINEL!r}; got {echoed!r}"
    return CheckResult(name=name, passed=False, detail=detail)


def _check_tenant_required(
    base: str,
    token: str,
) -> CheckResult:
    """Check 5: request without X-Tenant-ID is rejected 403 vendor-error (spec §5.2)."""
    name = "Gateway route requires tenant context"
    payload = json.dumps(
        {"name": "NoTenant", "budget": 1, "targetAudience": "x", "successMetric": "ctr"}
    ).encode()
    hdrs = {
        "Authorization": f"Bearer {token}",
        # Intentionally no X-Tenant-ID
        "Content-Type": _CT_JSON,
    }
    status, resp_hdrs, _ = _http(
        "POST", f"{base}/campaigns", headers=hdrs, body=payload
    )
    ct = _ct_base(resp_hdrs)
    if status == 403 and ct == _CT_ERROR:
        return CheckResult(name=name, passed=True)
    detail = f"status {status}; Content-Type: {ct!r} (expected 403 + {_CT_ERROR!r})"
    return CheckResult(name=name, passed=False, detail=detail)


# ── public entry-point ────────────────────────────────────────────────────────


def conformance_test(
    base_url: str,
    *,
    jwt_secret: str = DEMO_JWT_SECRET,
    tenant_id: str = _DEFAULT_TENANT,
) -> ConformanceResult:
    """Run the YAAgents conformance suite against *base_url*.

    Returns a :class:`ConformanceResult`; never raises (callers inspect `.passed`).

    URL scheme MUST be ``http`` or ``https`` — ``file://`` and other schemes are
    rejected per NFR-CLI-1.
    """
    parsed = urllib.parse.urlparse(base_url)
    if parsed.scheme not in ("http", "https"):
        return ConformanceResult(
            error=f"URL scheme must be http or https; got {parsed.scheme!r}"
        )

    base = base_url.rstrip("/")
    token = make_jwt(jwt_secret)

    checks: list[CheckResult] = []

    try:
        # Check 1 — profile header
        checks.append(_check_profile_header(base, token, tenant_id))

        # Checks 2 + 3 share one HTTP call
        c2, clarif_body = _check_clarification_ct(base, token, tenant_id)
        checks.append(c2)
        checks.append(_check_clarification_schema(clarif_body))

        # Check 4 — correlation ID
        checks.append(_check_correlation_id(base, token, tenant_id))

        # Check 5 — tenant enforcement
        checks.append(_check_tenant_required(base, token))

    except ConformanceError as exc:
        return ConformanceResult(checks=checks, error=str(exc))

    return ConformanceResult(checks=checks)
