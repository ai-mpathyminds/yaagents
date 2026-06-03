# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents-cli — Agentic REST Profile v0.3 validator CLI.

Supports-YAAgents-Profile: v0.3
"""

from yaagents_cli.__about__ import __version__
from yaagents_cli.__main__ import main
from yaagents_cli._init_fastapi import InitResult, init_fastapi
from yaagents_cli._validate import ValidateResult, validate_response
from yaagents_cli._validate_openapi import OpenApiResult, validate_openapi

__all__ = [
    "__version__",
    "main",
    "InitResult",
    "init_fastapi",
    "ValidateResult",
    "validate_response",
    "OpenApiResult",
    "validate_openapi",
]
