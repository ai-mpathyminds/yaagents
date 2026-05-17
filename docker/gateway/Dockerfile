# yaagents gateway — multi-stage Alpine build
# Published as: ghcr.io/ai-mpathyminds/yaagents-gateway:0.1.0
# Multi-arch: linux/amd64 + linux/arm64 (see WI-1yaa.REL-5)
# Non-root; CGO_ENABLED=0 for fully static binary.
# SBOM generated at publish via syft (SPDX 2.3) — platform-engineer A-4 pick.
#
# SECRET HYGIENE (NFR-GW-1): No ENV instruction may set GATEWAY_JWT_SECRET,
# GATEWAY_JWT_JWKS_URL, or any other credential. All secrets are runtime-only
# environment variables injected by the container orchestrator.

FROM golang:1.23-alpine3.20 AS builder
WORKDIR /build

# Cache deps separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/gateway

# ── runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.20 AS runtime

# Non-root runtime user (govulncheck + trivy: no unnecessary capabilities)
RUN addgroup -S yaagents && adduser -S gateway -G yaagents

WORKDIR /app
COPY --from=builder /gateway /app/gateway

# Liveness/readiness probes use wget (busybox built-in on Alpine)
HEALTHCHECK --interval=10s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:${GATEWAY_PORT:-8120}/healthz || exit 1

USER gateway
# Portfolio port allocation: yaagents-gateway = 8120
# (ref: .claude/rules/portfolio-conventions.md §Port Allocation)
EXPOSE 8120

ENTRYPOINT ["/app/gateway"]
