# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
# store-go — YAAgents Go reference example (WI-3yaa.STORE-EXAMPLE-AND-SKILLS)
# Multi-stage Alpine build; non-root; CGO_ENABLED=0 (fully static binary).
#
# Build context: yaagents/ repo root (docker compose sets context: ../..)
# All COPY paths are relative to the yaagents root.
#
# SECRET HYGIENE: No ENV instruction may set credentials or secrets.

FROM golang:1.25-alpine AS builder
WORKDIR /build

# Copy the sdk-go source first (required by replace directive ../../sdk-go).
COPY sdk-go/ ./sdk-go/

# Copy go.mod + go.sum to resolve dependencies before full source copy.
COPY examples/store-go/go.mod examples/store-go/go.sum ./examples/store-go/

# Copy application source + data files.
COPY examples/store-go/ ./examples/store-go/

WORKDIR /build/examples/store-go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /store-go .

# ── runtime image ─────────────────────────────────────────────────────────────
FROM alpine:3.20 AS runtime

# Non-root runtime user.
RUN addgroup -S yaagents && adduser -S store -G yaagents

WORKDIR /app
COPY --from=builder /store-go /app/store-go

# Copy static data files.
COPY --from=builder /build/examples/store-go/data/ /app/data/

# Liveness/readiness probes use wget (busybox built-in on Alpine).
HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:${PORT:-8121}/healthz || exit 1

USER store

# Portfolio port allocation: yaagents store-go = 8121 (container)
EXPOSE 8121

ENV PORT=8121

ENTRYPOINT ["/app/store-go"]
