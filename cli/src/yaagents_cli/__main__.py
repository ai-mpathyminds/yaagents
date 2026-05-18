"""yaagents CLI — entry point.

Usage:
    yaagents validate-response <file.json> [--json]
"""

from __future__ import annotations

import argparse
import json
import sys

from yaagents_cli._validate import validate_response
from yaagents_cli._validate_openapi import validate_openapi


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="yaagents",
        description="YAAgents CLI — Agentic REST Profile v0.1 validator.",
    )
    sub = parser.add_subparsers(dest="command", metavar="<command>")

    vr = sub.add_parser(
        "validate-response",
        help="Validate a response body JSON file against the v0.1 schema.",
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


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)

    if args.command == "validate-response":
        return _cmd_validate_response(args)
    if args.command == "validate-openapi":
        return _cmd_validate_openapi(args)

    parser.print_help()
    return 2


def _entry() -> None:
    sys.exit(main())


if __name__ == "__main__":
    _entry()
