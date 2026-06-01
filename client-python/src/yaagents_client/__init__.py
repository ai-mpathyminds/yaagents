# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents-client — Python client for the YAAgents Agentic REST Profile v0.2."""

from .__about__ import __profile__, __version__
from ._client import PROFILE_VERSION, YaAgentsClient
from ._exceptions import (
    AgenticError,
    AgenticForbidden,
    ClarificationRequired,
    FailedDependency,
    ValidationFailed,
)
from ._resources import CampaignResource

__all__ = [
    "__version__",
    "__profile__",
    "PROFILE_VERSION",
    "YaAgentsClient",
    "CampaignResource",
    # exceptions
    "AgenticError",
    "AgenticForbidden",
    "ClarificationRequired",
    "FailedDependency",
    "ValidationFailed",
]
