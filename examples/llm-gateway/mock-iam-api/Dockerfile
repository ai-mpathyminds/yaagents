# Multi-stage Alpine build — non-root, CGO_ENABLED=0, arm64-capable.
# Image: ghcr.io/ai-mpathyminds/yaagents-mock-iam-api:demo (demo-only; not published).

FROM golang:1.23-alpine AS builder
WORKDIR /build
ENV CGO_ENABLED=0 GOOS=linux
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /mock-iam-api .

FROM alpine:3.21
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /mock-iam-api /app/mock-iam-api
# Default tenants file; override by mounting a custom file via compose volumes.
COPY mock-tenants.yaml /app/mock-tenants.yaml
USER appuser
ENV TENANTS_FILE=/app/mock-tenants.yaml PORT=8122
EXPOSE 8122
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8122/healthz || exit 1
ENTRYPOINT ["/app/mock-iam-api"]
