# Campaign API -- YAAgents Agentic REST Profile v0.1 reference example
# Multi-stage build: dependency layer -> app layer
# Supports YAAgents Profile v0.1 (spec/agentic-rest-profile.md)
#
# Build context: yaagents/ repo root (docker compose sets context: ../..)
# All COPY paths are relative to the yaagents root.

ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-slim AS base

WORKDIR /app

# Install dependencies in a separate layer for cache efficiency
COPY examples/campaign-api/pyproject.toml ./
# Install the SDK from the local source tree (for demo; production would use PyPI)
COPY sdk-fastapi /opt/yaagents-fastapi/
RUN pip install --no-cache-dir /opt/yaagents-fastapi/ && \
    pip install --no-cache-dir "fastapi>=0.111" "uvicorn[standard]>=0.29"

# App layer (README.md required by hatchling metadata validator)
COPY examples/campaign-api/README.md ./
COPY examples/campaign-api/src/ ./src/
RUN pip install --no-cache-dir -e . --no-deps

EXPOSE 8121

# Non-root user
RUN adduser --disabled-password --gecos "" appuser
USER appuser

ENV PORT=8121
CMD ["python", "-m", "uvicorn", "campaign_api.app:app", "--host", "0.0.0.0", "--port", "8121"]
