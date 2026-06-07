# Financial Risk Screening — YAAgents Agentic REST Profile v0.3 example
# Multi-stage build; non-root user; port 8123 (yaagents band 8120-8129)
# Supports YAAgents Profile v0.3 (spec/agentic-rest-profile.md)
#
# Build context: yaagents/ repo root (docker compose sets context: ../..)
# All COPY paths are relative to the yaagents root.

ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-slim AS base

WORKDIR /app

# Install dependencies in a separate layer for cache efficiency.
# Install the SDK from the local source tree (demo; production uses PyPI).
COPY sdk-fastapi /opt/yaagents-fastapi/
RUN pip install --no-cache-dir /opt/yaagents-fastapi/ && \
    pip install --no-cache-dir "fastapi>=0.111" "uvicorn[standard]>=0.29"

# App layer
COPY examples/financial-risk-screening/main.py ./

EXPOSE 8123

# Non-root user
RUN adduser --disabled-password --gecos "" appuser
USER appuser

ENV PORT=8123
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8123"]
