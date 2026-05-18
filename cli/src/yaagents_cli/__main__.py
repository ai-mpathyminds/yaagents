"""yaagents CLI — entry point.

Usage:
    yaagents validate-response <file.json> [--json]
    yaagents validate-openapi  <file.yaml> [--json]
    yaagents init fastapi      [target-dir] [--force] [--json]
"""

from __future__ import annotations

import argparse
import json
import sys

from yaagents_cli._init_fastapi import init_fastapi
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


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)

    if args.command == "validate-response":
        return _cmd_validate_response(args)
    if args.command == "validate-openapi":
        return _cmd_validate_openapi(args)
    if args.command == "init":
        return _cmd_init(args)

    parser.print_help()
    return 2


def _entry() -> None:
    sys.exit(main())


if __name__ == "__main__":
    _entry()
