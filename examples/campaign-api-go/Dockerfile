# campaign-api-go — YAAgents Go reference example (WI-3yaa.SG-6)
# Multi-stage Alpine build; non-root; CGO_ENABLED=0 (fully static binary).
# SPDX-License-Identifier: Apache-2.0
#
# Build context: yaagents/ repo root (docker compose sets context: ../..)
# All COPY paths are relative to the yaagents root.
#
# SECRET HYGIENE: No ENV instruction may set credentials or secrets.
# All configuration is injected at runtime via the container orchestrator.

FROM golang:1.22-alpine3.20 AS builder
WORKDIR /build

# Copy the sdk-go source first (required by replace directive ../../sdk-go).
COPY sdk-go/ ./sdk-go/

# Copy go.mod + go.sum to resolve dependencies before full source copy.
COPY examples/campaign-api-go/go.mod examples/campaign-api-go/go.sum ./examples/campaign-api-go/

# With only a local replace directive (no external deps), no download step needed.
# Build source copy.
COPY examples/campaign-api-go/ ./examples/campaign-api-go/

WORKDIR /build/examples/campaign-api-go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /campaign-api-go .

# ── runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.20 AS runtime

# Non-root runtime user.
RUN addgroup -S yaagents && adduser -S campaign -G yaagents

WORKDIR /app
COPY --from=builder /campaign-api-go /app/campaign-api-go

# Liveness/readiness probes use wget (busybox built-in on Alpine).
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:${PORT:-8121}/healthz || exit 1

USER campaign

# Portfolio port allocation: yaagents campaign-api-go = 8121
# (ref: portfolio conventions §Port Allocation)
EXPOSE 8121

ENV PORT=8121

ENTRYPOINT ["/app/campaign-api-go"]
