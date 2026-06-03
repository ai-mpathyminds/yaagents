# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents CLI — entry point.

Usage:
    yaagents validate-response <file.json> [--json]
    yaagents validate-openapi  <file.yaml> [--json]
    yaagents init fastapi      [target-dir] [--force] [--json]
    yaagents conformance-test  <base-url>  [--jwt-secret S] [--tenant T] [--json]
"""

from __future__ import annotations

import argparse
import json
import sys

from yaagents_cli.__about__ import __version__
from yaagents_cli._conformance import DEMO_JWT_SECRET, conformance_test
from yaagents_cli._init_fastapi import init_fastapi
from yaagents_cli._validate import validate_response
from yaagents_cli._validate_openapi import validate_openapi


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="yaagents",
        description="YAAgents CLI — Agentic REST Profile v0.3 validator.",
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"%(prog)s {__version__}",
    )
    sub = parser.add_subparsers(dest="command", metavar="<command>")

    vr = sub.add_parser(
        "validate-response",
        help="Validate a response body JSON file against the v0.3 schema.",
    )
    vr.add_argument(
        "file",
        metavar="<file.json>",
        help="Path to the response body JSON file.",
    )
    vr.add_argument(
        "--json",
        dest="json_output",
        action="store_true",
        default=False,
        help="Emit machine-readable JSON output.",
    )

    vo = sub.add_parser(
        "validate-openapi",
        help=(
            "Validate an OpenAPI YAML file for YAAgents profile conformance: "
            "x-yaagents metadata, §4 Content-Type table, and $ref resolution."
        ),
    )
    vo.add_argument(
        "file",
        metavar="<file.yaml>",
        help="Path to the OpenAPI YAML file.",
    )
    vo.add_argument(
        "--json",
        dest="json_output",
        action="store_true",
        default=False,
        help="Emit machine-readable JSON output.",
    )

    # ── init ──────────────────────────────────────────────────────────────────
    init_p = sub.add_parser(
        "init",
        help="Scaffold a new project from a YAAgents template.",
    )
    init_p.add_argument(
        "framework",
        choices=["fastapi"],
        metavar="<framework>",
        help="Target framework: fastapi",
    )
    init_p.add_argument(
        "target_dir",
        metavar="[target-dir]",
        nargs="?",
        default="campaign-api",
        help="Directory to scaffold into (default: campaign-api).",
    )
    init_p.add_argument(
        "--force",
        action="store_true",
        default=False,
        help="Overwrite an existing non-empty target directory.",
    )
    init_p.add_argument(
        "--json",
        dest="json_output",
        action="store_true",
        default=False,
        help="Emit machine-readable JSON output.",
    )

    # ── conformance-test ──────────────────────────────────────────────────────
    ct = sub.add_parser(
        "conformance-test",
        help=(
            "Exercise mandatory response types against a live YAAgents gateway "
            "and emit a PRD §5.8 conformance report."
        ),
    )
    ct.add_argument(
        "base_url",
        metavar="<base-url>",
        help="Gateway base URL, e.g. http://localhost:8120",
    )
    ct.add_argument(
        "--jwt-secret",
        dest="jwt_secret",
        default=DEMO_JWT_SECRET,
        metavar="SECRET",
        help=(
            "HS256 JWT secret for the target gateway "
            "(default: Compose demo secret)"
        ),
    )
    ct.add_argument(
        "--tenant",
        dest="tenant_id",
        default="t-conformance",
        metavar="TENANT",
        help="X-Tenant-ID value to use in test requests (default: t-conformance).",
    )
    ct.add_argument(
        "--require-plugin",
        dest="require_plugins",
        action="append",
        default=[],
        metavar="PLUGIN",
        help=(
            "Assert that a named plugin is active (repeatable). "
            "Supported: token-validator, tenant-injector."
        ),
    )
    ct.add_argument(
        "--json",
        dest="json_output",
        action="store_true",
        default=False,
        help="Emit machine-readable JSON output.",
    )

    return parser


def _cmd_validate_response(args: argparse.Namespace) -> int:
    result = validate_response(args.file)

    if args.json_output:
        print(json.dumps(result.to_dict(), indent=2))
        return 0 if result.passed else 1

    # ── human-readable output ────────────────────────────────────────────────
    if result.error:
        print(f"ERROR: {result.error}", file=sys.stderr)
        return 1

    if result.passed:
        schema_info = f" [{result.schema_name}]" if result.schema_name else ""
        print(f"PASS: {result.file}{schema_info}")
        return 0

    print(f"FAIL: {result.file}")
    for finding in result.findings:
        print(f"  [{finding.pointer}] {finding.message}")
    return 1


def _cmd_validate_openapi(args: argparse.Namespace) -> int:
    result = validate_openapi(args.file)

    if args.json_output:
        print(json.dumps(result.to_dict(), indent=2))
        return 0 if result.passed else 1

    # ── human-readable output ────────────────────────────────────────────────
    if result.error:
        print(f"ERROR: {result.error}", file=sys.stderr)
        return 1

    if result.passed:
        print(f"PASS: {result.file}")
        return 0

    print(f"FAIL: {result.file}")
    for finding in result.findings:
        print(f"  [{finding.pointer}] {finding.message}")
    return 1


def _cmd_init(args: argparse.Namespace) -> int:
    result = init_fastapi(args.target_dir, force=args.force)

    if args.json_output:
        print(json.dumps(result.to_dict(), indent=2))
        return 0 if result.succeeded else 1

    if result.error:
        print(f"ERROR: {result.error}", file=sys.stderr)
        return 1

    print(f"OK: scaffolded {args.framework!r} project in {result.target_dir!r}")
    for f in result.created:
        print(f"  + {f}")
    return 0


def _cmd_conformance_test(args: argparse.Namespace) -> int:
    result = conformance_test(
        args.base_url,
        jwt_secret=args.jwt_secret,
        tenant_id=args.tenant_id,
        require_plugins=args.require_plugins or [],
    )

    if args.json_output:
        print(json.dumps(result.to_dict(), indent=2))
        return 0 if result.passed else 1

    # ── human-readable PRD §5.8 report format ────────────────────────────────
    if result.error:
        print(f"ERROR: {result.error}", file=sys.stderr)
        return 1

    print("YAAgents Conformance Report")
    print()
    for check in result.checks:
        mark = "✓" if check.passed else "✗"
        line = f"{mark} {check.name}"
        if not check.passed and check.detail:
            line += f" — {check.detail}"
        print(line)

    # ── Content-Type matrix summary table (PRD §4, 10 rows) ──────────────────
    if result.matrix:
        print()
        print("Content-Type matrix (PRD §4):")
        col_w = [8, 48, 48, 6]
        hdr = (
            f"  {'status':<{col_w[0]}} | {'requested':<{col_w[1]}} | "
            f"{'observed':<{col_w[2]}} | pass"
        )
        sep = (
            "  " + "-" * col_w[0] + "-+-"
            + "-" * col_w[1] + "-+-"
            + "-" * col_w[2] + "-+------"
        )
        print(hdr)
        print(sep)
        for row in result.matrix:
            pass_str = "PASS" if row.passed else "FAIL"
            print(
                f"  {row.status:<{col_w[0]}} | {row.requested:<{col_w[1]}} | "
                f"{row.observed:<{col_w[2]}} | {pass_str}"
            )

    print()
    overall = "PASS" if result.passed else "FAIL"
    print(f"Overall: {overall}")
    return 0 if result.passed else 1


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)

    if args.command == "validate-response":
        return _cmd_validate_response(args)
    if args.command == "validate-openapi":
        return _cmd_validate_openapi(args)
    if args.command == "init":
        return _cmd_init(args)
    if args.command == "conformance-test":
        return _cmd_conformance_test(args)

    parser.print_help()
    return 2


def _entry() -> None:
    sys.exit(main())


if __name__ == "__main__":
    _entry()
