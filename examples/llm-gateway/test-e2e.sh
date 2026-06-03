#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
#
# test-e2e.sh — E2E acceptance harness for WI-2yaa.EX-LLM-3.
#
# Exercises all 6 PRD §13.2 flows against the running docker compose stack
# (yaagents-gateway:8122, llm-api, mock-iam-api).
#
# Usage:
#   cd examples/llm-gateway
#   docker compose up -d --build
#   ./test-e2e.sh
#
# Exits 0 when all flows PASS, 1 otherwise.
#
# Environment variables (optional overrides):
#   GATEWAY_URL   gateway base URL (default: http://localhost:8122)
#   JWT_SECRET    HS256 secret     (default: demo-secret-not-for-production)

set -euo pipefail

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8122}"
JWT_SECRET="${JWT_SECRET:-demo-secret-not-for-production}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PASS=0
FAIL=0
FAILED_FLOWS=()

# ── helpers ────────────────────────────────────────────────────────────────────

log() { printf '\n[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }
pass() { PASS=$((PASS+1)); printf '  ✓ PASS: %s\n' "$*"; }
fail() { FAIL=$((FAIL+1)); FAILED_FLOWS+=("$*"); printf '  ✗ FAIL: %s\n' "$*"; }

# mint_token: generate a demo HS256 JWT using Python (stdlib only; no PyJWT needed).
# sub defaults to "user-alice@example.com" (mapped to tenant-001 in mock-tenants.yaml).
mint_token() {
  local sub="${1:-user-alice@example.com}"
  local secret="${2:-$JWT_SECRET}"
  python3 - <<EOF
import base64, hashlib, hmac, json, time

def b64url(d):
    return base64.urlsafe_b64encode(d).rstrip(b'=').decode()

secret = '${secret}'
sub    = '${sub}'
h = b64url(json.dumps({'alg':'HS256','typ':'JWT'}).encode())
p = b64url(json.dumps({'sub': sub, 'roles': ['llm:complete'], 'exp': int(time.time()) + 3600}).encode())
sig = hmac.new(secret.encode(), f'{h}.{p}'.encode(), hashlib.sha256).digest()
print(f'{h}.{p}.{b64url(sig)}')
EOF
}

# wait_healthy: wait until GET /healthz returns 200, or time out.
wait_healthy() {
  local url="$1"
  local max_secs="${2:-60}"
  local waited=0
  printf '  Waiting for %s/healthz' "$url"
  while true; do
    if curl -sf "$url/healthz" > /dev/null 2>&1; then
      printf ' (%ds)\n' "$waited"
      return 0
    fi
    if [ "$waited" -ge "$max_secs" ]; then
      printf '\n  TIMEOUT after %ds\n' "$max_secs"
      return 1
    fi
    sleep 2
    waited=$((waited+2))
    printf '.'
  done
}

# ── pre-flight: health check ───────────────────────────────────────────────────

log "Pre-flight: gateway health check"
if ! wait_healthy "$GATEWAY_URL" 60; then
  echo "ERROR: gateway not healthy within 60s — run: docker compose up -d --build" >&2
  exit 1
fi
echo "  Gateway is healthy."

# ── mint demo token ────────────────────────────────────────────────────────────

log "Minting demo JWT (user-alice@example.com → tenant-001)"
TOKEN=$(mint_token "user-alice@example.com" "$JWT_SECRET")
echo "  Token: ${TOKEN:0:40}…"

# ── Flow 1: Standard call (non-streaming → 201 JSON) ──────────────────────────

log "Flow 1: Standard LLM call (non-streaming → 201 JSON)"
RESP=$(curl -s -w '\n%{http_code}' -X POST "$GATEWAY_URL/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Generate a campaign headline", "stream": false}')
BODY=$(echo "$RESP" | head -n -1)
STATUS=$(echo "$RESP" | tail -n 1)

if [ "$STATUS" = "201" ]; then
  # Check profile header via a separate request with -I (head)
  PROFILE=$(curl -s -o /dev/null -w '%{header_json}' -X POST "$GATEWAY_URL/completions" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"prompt": "header check", "stream": false}' 2>/dev/null \
    | python3 -c "import sys,json; h=json.load(sys.stdin); print(h.get('x-yaagents-profile',[''])[0])" 2>/dev/null \
    || echo "")
  echo "  Status: $STATUS | X-YAAgents-Profile: $PROFILE"
  echo "  Body snippet: $(echo "$BODY" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("type","?"), d.get("choices",[{}])[0].get("text","?")[:40])' 2>/dev/null || echo '(raw)')"
  pass "Flow 1 — 201 JSON completion"
else
  echo "  Status: $STATUS | Body: $BODY"
  fail "Flow 1 — expected 201, got $STATUS"
fi

# ── Flow 2: SSE streaming ──────────────────────────────────────────────────────

log "Flow 2: SSE streaming (Accept: text/event-stream)"
SSE_OUT=$(curl -s -N --max-time 15 -X POST "$GATEWAY_URL/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"prompt": "Stream a campaign idea", "stream": true}' 2>/dev/null)

if echo "$SSE_OUT" | grep -q "data: \[DONE\]"; then
  CHUNK_COUNT=$(echo "$SSE_OUT" | grep -c '^data: ' || true)
  echo "  Received $CHUNK_COUNT SSE chunks including [DONE]"
  pass "Flow 2 — SSE streaming with progressive chunks and [DONE]"
elif echo "$SSE_OUT" | grep -q '"type"'; then
  # Non-SSE JSON response (gateway fell back to JSON mode)
  echo "  SSE output: $SSE_OUT"
  fail "Flow 2 — no [DONE] event in SSE stream"
else
  echo "  SSE output: $SSE_OUT"
  fail "Flow 2 — unexpected SSE response"
fi

# ── Flow 3: SSE concurrency exceeded (11th request → 429) ────────────────────

log "Flow 3: SSE concurrency limit (11 concurrent → 11th gets 429)"
# Start 10 slow-SSE connections to fill the per-tenant limit.
PIDS=()
for i in $(seq 1 10); do
  curl -s -N --max-time 30 -X POST "$GATEWAY_URL/completions" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "Accept: text/event-stream" \
    -d "{\"prompt\": \"hold $i\", \"stream\": true, \"hold_open\": true}" \
    > /tmp/sse_hold_$i.txt 2>&1 &
  PIDS+=($!)
done

# Give the connections time to establish their SSE slots.
sleep 3

# 11th request: should be rejected with 429.
STATUS11=$(curl -s -o /tmp/flow3_11.json -w '%{http_code}' \
  --max-time 10 \
  -X POST "$GATEWAY_URL/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "overflow request", "stream": false}')

# Kill background hold-open connections.
for pid in "${PIDS[@]}"; do
  kill "$pid" 2>/dev/null || true
done
wait 2>/dev/null || true

if [ "$STATUS11" = "429" ]; then
  CODE=$(python3 -c "import json; d=json.load(open('/tmp/flow3_11.json')); print(d.get('code','?'))" 2>/dev/null || echo "?")
  echo "  11th request: 429 | code: $CODE"
  pass "Flow 3 — 11th concurrent SSE request got 429"
else
  echo "  11th request: $STATUS11 (expected 429)"
  echo "  Body: $(cat /tmp/flow3_11.json 2>/dev/null)"
  fail "Flow 3 — expected 429 from SSE concurrency limit, got $STATUS11"
fi

# ── Flow 4: Execution timeout (simulate_timeout → 500 EXECUTION_TIMEOUT) ──────

log "Flow 4: Execution timeout (simulate_timeout: true → 500 EXECUTION_TIMEOUT)"
# The gateway's executionTimeoutSeconds=30 + SSE=30 = 60s total.
# The mock-llm-api sleeps 60s; gateway will cancel at ~60s.
# Use a 65s curl timeout to let the gateway fire first.
FLOW4_STATUS=$(curl -s -o /tmp/flow4.json -w '%{http_code}' \
  --max-time 65 \
  -X POST "$GATEWAY_URL/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "timeout test", "stream": false, "simulate_timeout": true}')

FLOW4_CODE=$(python3 -c "import json; d=json.load(open('/tmp/flow4.json')); print(d.get('code','?'))" 2>/dev/null || echo "?")
echo "  Status: $FLOW4_STATUS | code: $FLOW4_CODE"

if [ "$FLOW4_STATUS" = "500" ] && [ "$FLOW4_CODE" = "EXECUTION_TIMEOUT" ]; then
  pass "Flow 4 — 500 EXECUTION_TIMEOUT after gateway timeout"
else
  echo "  Body: $(cat /tmp/flow4.json 2>/dev/null)"
  fail "Flow 4 — expected 500/EXECUTION_TIMEOUT, got $FLOW4_STATUS/$FLOW4_CODE"
fi

# ── Flow 5: CORS preflight ────────────────────────────────────────────────────

log "Flow 5: CORS preflight (OPTIONS from http://localhost:3000)"
CORS_RESP=$(curl -si -X OPTIONS "$GATEWAY_URL/completions" \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  --max-time 10 2>/dev/null)

CORS_STATUS=$(echo "$CORS_RESP" | grep -m1 'HTTP/' | awk '{print $2}')
ACAO=$(echo "$CORS_RESP" | grep -i 'Access-Control-Allow-Origin' | head -1 | sed 's/.*: //' | tr -d '\r')

echo "  Status: $CORS_STATUS | ACAO: $ACAO"
if [ "$CORS_STATUS" = "200" ] && [ "$ACAO" = "http://localhost:3000" ]; then
  pass "Flow 5 — CORS preflight 200 with correct ACAO header"
else
  fail "Flow 5 — expected 200 + ACAO=http://localhost:3000, got $CORS_STATUS / $ACAO"
fi

# ── Flow 6: Community plugin (build + registration verification) ───────────────

log "Flow 6: Community plugin — build and verify registration"
PLUGIN_GW_DIR="$SCRIPT_DIR/community-plugin-gateway"

if ! command -v go &> /dev/null; then
  echo "  WARNING: 'go' not in PATH — skipping community plugin build"
  fail "Flow 6 — Go toolchain not available for community plugin build"
else
  BINARY="$PLUGIN_GW_DIR/community-gw-verify"
  (cd "$PLUGIN_GW_DIR" && go build -o community-gw-verify . 2>&1) || {
    fail "Flow 6 — go build failed for community-plugin-gateway"
  }
  if [ -f "$BINARY" ]; then
    PLUGIN_OUT=$("$BINARY" 2>&1)
    echo "  Binary output: $PLUGIN_OUT"
    if echo "$PLUGIN_OUT" | grep -q "plugin registered: community-example"; then
      pass "Flow 6 — community-example plugin registered in custom gateway binary"
    else
      fail "Flow 6 — 'plugin registered: community-example' not found in output"
    fi
    rm -f "$BINARY"
  fi
fi

# ── Summary ────────────────────────────────────────────────────────────────────

log "Results"
printf '\n  Passed: %d / Failed: %d\n' "$PASS" "$FAIL"
if [ "${#FAILED_FLOWS[@]}" -gt 0 ]; then
  printf '\n  Failed flows:\n'
  for f in "${FAILED_FLOWS[@]}"; do
    printf '    - %s\n' "$f"
  done
  printf '\n'
  exit 1
fi

printf '\n  All %d flows PASS.\n\n' "$PASS"
