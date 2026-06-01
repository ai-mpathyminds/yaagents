# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
#
# Multi-stage Alpine build — non-root, CGO_ENABLED=0, arm64-capable.
# Image: built locally by the llm-gateway compose demo (WI-2yaa.EX-LLM-2).
#
# SECURITY NOTE: this image is for demo/development use only.
# Do not deploy to production without a real authentication layer.

FROM golang:1.23-alpine AS builder
WORKDIR /build
ENV CGO_ENABLED=0 GOOS=linux
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /mock-llm-api .

FROM alpine:3.21
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /mock-llm-api /app/mock-llm-api
USER appuser
ENV PORT=8123 STREAM_CHUNKS=5 STREAM_DELAY_MS=100
EXPOSE 8123
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8123/healthz || exit 1
ENTRYPOINT ["/app/mock-llm-api"]
