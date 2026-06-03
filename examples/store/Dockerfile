# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
# yaagents store example — FastAPI reference image
#
# Build context: examples/store/ (docker compose default)
# Uses yaagents-fastapi from PyPI (v0.3.0+).
#
# SECRET HYGIENE: No ENV instruction sets credentials or secrets.
# All configuration is injected at runtime via the container orchestrator.

ARG PYTHON_VERSION=3.12
FROM python:${PYTHON_VERSION}-slim

WORKDIR /app

# Install Python dependencies before copying source (better layer caching).
COPY pyproject.toml .
RUN pip install --no-cache-dir "fastapi>=0.115" "uvicorn[standard]>=0.32" "yaagents-fastapi>=0.3.0"

# Copy source after deps (source changes don't invalidate the deps layer).
COPY src/ src/
COPY data/ data/

# Expose src/ on PYTHONPATH so 'import store' resolves to src/store/
# without requiring a package install step (keeps data path relative to source).
ENV PYTHONPATH=/app/src

# Non-root runtime user (security best practice).
RUN adduser --disabled-password --gecos "" appuser && \
    chown -R appuser:appuser /app

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
  CMD python3 -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/healthz')" || exit 1

USER appuser

# Portfolio port allocation: yaagents store example = 8121 (host), 8080 (container)
EXPOSE 8080

CMD ["uvicorn", "store.app:app", "--host", "0.0.0.0", "--port", "8080"]
