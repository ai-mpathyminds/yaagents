"""validate-openapi logic for yaagents-cli.

Three checks per WI-1yaa.CLI-2:
  (a) x-yaagents metadata present + well-formed on agentic operations
  (b) each declared vendor Content-Type matches the §4 normative table
  (c) schema $ref values that point to schemas/v0.1 files resolve on disk
"""

from __future__ import annotations

import pathlib
from collections.abc import Generator
from typing import Any

import yaml

from yaagents_cli._validate import FindingEntry, ValidationError

# ── §4 normative table ────────────────────────────────────────────────────────
# HTTP status code → expected vendor Content-Type for that status.
# Only vendor (vnd.yaagents.*) content-types are policed; plain
# application/json (200/201) is left to the service's own validation.
_STATUS_TO_VENDOR_CT: dict[str, str] = {
    "202": "application/vnd.yaagents.operation+json",
    "400": "application/vnd.yaagents.clarification+json",
    "412": "application/vnd.yaagents.approval-required+json",
    "422": "application/vnd.yaagents.validation-error+json",
    "403": "application/vnd.yaagents.error+json",
    "409": "application/vnd.yaagents.conflict+json",
    "424": "application/vnd.yaagents.error+json",
    "500": "application/vnd.yaagents.error+json",
}

# Reverse: vendor CT → set of valid HTTP statuses.
_VENDOR_CT_TO_STATUSES: dict[str, frozenset[str]] = {}
for _status, _ct in _STATUS_TO_VENDOR_CT.items():
    _VENDOR_CT_TO_STATUSES.setdefault(_ct, set()).add(_status)  # type: ignore[arg-type]
_VENDOR_CT_TO_STATUSES = {
    k: frozenset(v) for k, v in _VENDOR_CT_TO_STATUSES.items()
}

_VENDOR_PREFIX = "application/vnd.yaagents."

# ── x-yaagents schema constraints ────────────────────────────────────────────
_OPERATION_KINDS = frozenset(
    {"recommendation", "generation", "mutation", "analysis"}
)
_X_YAAGENTS_REQUIRED = ("resource", "operationKind", "deterministic", "mutating")

# Schemas/v0.1 path fragment — any $ref containing this is checked for disk
# existence.
_SCHEMA_PATH_FRAGMENT = "schemas/v0.1/"

_MAX_FILE_BYTES = 10 * 1024 * 1024  # 10 MB


# ── result types (reuse FindingEntry from _validate) ──────────────────────────


class OpenApiResult:
    """Outcome of a single validate-openapi call."""

    def __init__(
        self,
        file: str,
        passed: bool,
        findings: list[FindingEntry] | None = None,
        error: str | None = None,
    ) -> None:
        self.file = file
        self.passed = passed
        self.findings: list[FindingEntry] = findings or []
        self.error = error

    def to_dict(self) -> dict[str, Any]:
        d: dict[str, Any] = {
            "file": self.file,
            "result": "PASS" if self.passed else "FAIL",
        }
        if self.findings:
            d["errors"] = [f.to_dict() for f in self.findings]
        if self.error:
            d["error"] = self.error
        return d


# ── helpers ───────────────────────────────────────────────────────────────────


def _walk_refs(
    node: Any, path: str = ""
) -> Generator[tuple[str, str], None, None]:
    """Yield (json-pointer-ish path, $ref value) for every $ref in *node*."""
    if isinstance(node, dict):
        for k, v in node.items():
            child_path = f"{path}/{k}"
            if k == "$ref" and isinstance(v, str):
                yield child_path, v
            else:
                yield from _walk_refs(v, child_path)
    elif isinstance(node, list):
        for i, item in enumerate(node):
            yield from _walk_refs(item, f"{path}/{i}")


def _load_yaml(file_path: str) -> Any:
    """Load YAML from *file_path* with path-safety + size check."""
    p = pathlib.Path(file_path)
    if ".." in p.parts:
        raise ValidationError("path traversal not allowed")
    try:
        size = p.stat().st_size
    except FileNotFoundError as exc:
        raise ValidationError(f"file not found: {file_path}") from exc
    if size > _MAX_FILE_BYTES:
        raise ValidationError("file exceeds 10 MB limit")
    return yaml.safe_load(p.read_text(encoding="utf-8"))


def _check_x_yaagents(
    x_ext: Any, op_path: str, findings: list[FindingEntry]
) -> None:
    """Validate the x-yaagents extension object on one operation."""
    if not isinstance(x_ext, dict):
        findings.append(
            FindingEntry(
                pointer=f"{op_path}/x-yaagents",
                message="x-yaagents must be an object",
            )
        )
        return
    for field in _X_YAAGENTS_REQUIRED:
        if field not in x_ext:
            findings.append(
                FindingEntry(
                    pointer=f"{op_path}/x-yaagents/{field}",
                    message=f"x-yaagents.{field} is required but missing",
                )
            )
    # operationKind enum
    op_kind = x_ext.get("operationKind")
    if op_kind is not None and op_kind not in _OPERATION_KINDS:
        findings.append(
            FindingEntry(
                pointer=f"{op_path}/x-yaagents/operationKind",
                message=(
                    f"operationKind '{op_kind}' is not one of "
                    f"{sorted(_OPERATION_KINDS)}"
                ),
            )
        )
    # resource must be a non-empty string
    resource = x_ext.get("resource")
    if resource is not None and (
        not isinstance(resource, str) or not resource.strip()
    ):
        findings.append(
            FindingEntry(
                pointer=f"{op_path}/x-yaagents/resource",
                message="x-yaagents.resource must be a non-empty string",
            )
        )
    # deterministic + mutating must be bool
    for bool_field in ("deterministic", "mutating"):
        val = x_ext.get(bool_field)
        if val is not None and not isinstance(val, bool):
            findings.append(
                FindingEntry(
                    pointer=f"{op_path}/x-yaagents/{bool_field}",
                    message=(
                        f"x-yaagents.{bool_field} must be a boolean"
                    ),
                )
            )


def _check_responses(
    responses: Any, op_path: str, findings: list[FindingEntry]
) -> bool:
    """Check Content-Types in *responses*. Returns True if any vendor CT found."""
    if not isinstance(responses, dict):
        return False
    has_vendor_ct = False
    for status_code, resp_obj in responses.items():
        status_str = str(status_code)
        if not isinstance(resp_obj, dict):
            continue
        content = resp_obj.get("content", {})
        if not isinstance(content, dict):
            continue
        for ct in content:
            if not ct.startswith(_VENDOR_PREFIX):
                continue
            has_vendor_ct = True
            # Check: is this CT valid for this status code?
            allowed = _VENDOR_CT_TO_STATUSES.get(ct)
            if allowed is None:
                findings.append(
                    FindingEntry(
                        pointer=(
                            f"{op_path}/responses/{status_str}"
                            f"/content/{ct}"
                        ),
                        message=f"unknown vendor Content-Type: {ct!r}",
                    )
                )
            elif status_str not in allowed:
                expected = _STATUS_TO_VENDOR_CT.get(status_str, "(none)")
                findings.append(
                    FindingEntry(
                        pointer=(
                            f"{op_path}/responses/{status_str}"
                            f"/content/{ct}"
                        ),
                        message=(
                            f"HTTP {status_str} must use "
                            f"{expected!r} per §4 table, "
                            f"got {ct!r}"
                        ),
                    )
                )
    return has_vendor_ct


def _check_refs(
    doc: Any, yaml_dir: pathlib.Path, findings: list[FindingEntry]
) -> None:
    """Walk all $ref values; verify schemas/v0.1 refs resolve on disk."""
    for ref_path, ref_val in _walk_refs(doc):
        # Skip JSON-Pointer fragment refs (same-document)
        if ref_val.startswith("#"):
            continue
        if _SCHEMA_PATH_FRAGMENT not in ref_val:
            continue
        # Resolve relative to the YAML file's directory
        resolved = (yaml_dir / ref_val).resolve()
        if not resolved.exists():
            findings.append(
                FindingEntry(
                    pointer=ref_path,
                    message=f"dangling $ref: {ref_val!r} not found at {resolved}",
                )
            )


# ── public entry point ────────────────────────────────────────────────────────


def validate_openapi(file_path: str) -> OpenApiResult:
    """Validate *file_path* against the YAAgents OpenAPI rules.

    Never raises; caller examines ``.passed``.
    """
    try:
        doc = _load_yaml(file_path)
    except ValidationError as exc:
        return OpenApiResult(file=file_path, passed=False, error=str(exc))

    if not isinstance(doc, dict):
        return OpenApiResult(
            file=file_path,
            passed=False,
            error="document root must be a YAML/JSON object",
        )

    yaml_dir = pathlib.Path(file_path).parent
    findings: list[FindingEntry] = []

    paths_obj = doc.get("paths", {})
    if isinstance(paths_obj, dict):
        for path_key, path_item in paths_obj.items():
            if not isinstance(path_item, dict):
                continue
            http_methods = (
                "get", "post", "put", "patch", "delete",
                "head", "options", "trace",
            )
            for method in http_methods:
                op = path_item.get(method)
                if not isinstance(op, dict):
                    continue
                op_pointer = f"/paths/{path_key}/{method}"

                # Check responses (Content-Type vs §4 table)
                has_vendor = _check_responses(
                    op.get("responses", {}), op_pointer, findings
                )

                # Check x-yaagents: required on agentic operations
                # An operation is agentic if it carries ANY vendor CT in
                # its responses OR if it already declares x-yaagents.
                x_ext = op.get("x-yaagents")
                if has_vendor or x_ext is not None:
                    if x_ext is None:
                        findings.append(
                            FindingEntry(
                                pointer=f"{op_pointer}/x-yaagents",
                                message=(
                                    "x-yaagents extension is required on "
                                    "operations that use vendor Content-Types"
                                ),
                            )
                        )
                    else:
                        _check_x_yaagents(x_ext, op_pointer, findings)

    # Check $ref resolution
    _check_refs(doc, yaml_dir, findings)

    return OpenApiResult(
        file=file_path,
        passed=len(findings) == 0,
        findings=findings if findings else None,
    )
