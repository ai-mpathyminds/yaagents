"""yaagents-fastapi — FastAPI SDK for the YAAgents Agentic REST Profile v0.1.

Supports YAAgents Profile v0.1 (spec/agentic-rest-profile.md).
"""

from yaagents_fastapi.__about__ import __profile__, __version__
from yaagents_fastapi.response import AgenticResponse

__all__ = [
    "__version__",
    "__profile__",
    "AgenticResponse",
]
