#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds
#
# test_e2e.sh — CI acceptance gate for campaign-api-go (WI-3yaa.SG-6)
#
# Exercises all 5 PRD §13.2 / §8.1 demo flows against live Docker Compose
# services. Exits 0 on all-pass; exits 1 on first failure.
#
# Prerequisites:
#   docker compose up -d   (from examples/campaign-api-go/ directory)
#   Services must be healthy before this script is invoked.
#
# Usage:
#   cd examples/campaign-api-go
#   docker compose up -d
#   bash test_e2e.sh
#
# Override service addresses:
#   CAMPAIGN_URL=http://localhost:8121   (direct to campaign-api-go)
#   GATEWAY_URL=http://localhost:8120    (via yaagents-gateway)

set -euo pipefail

CAMPAIGN_URL="${CAMPAIGN_URL:-http://localhost:8121}"
GATEWAY_URL="${GATEWAY_URL:-http://localhost:8120}"
WAIT_TIMEOUT="${WAIT_TIMEOUT:-60}"

# ── colours ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

pass()  { echo -e "${GREEN}PASS${NC} $1"; }
fail()  { echo -e "${RED}FAIL${NC} $1"; exit 1; }

# ── wait_healthy: poll /healthz until HTTP 200 or timeout ───────────────────
wait_healthy() {
  local url="$1"
  local label="$2"
  local deadline=$(( SECONDS + WAIT_TIMEOUT ))
  echo "Waiting for $label at $url/healthz ..."
  while (( SECONDS < deadline )); do
    status=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "$url/healthz" 2>/dev/null || echo "000")
    if [[ "$status" == "200" ]]; then
      echo "$label is healthy."
      return 0
    fi
    sleep 2
  done
  echo "ERROR: $label did not become healthy within ${WAIT_TIMEOUT}s" >&2
  exit 1
}

# ── check_response: assert status + content-type + optional header ───────────
# Usage: check_response LABEL EXPECTED_STATUS ACTUAL_STATUS EXPECTED_CT ACTUAL_CT [HEADER_NAME HEADER_VAL]
check_response() {
  local label="$1"
  local exp_status="$2"
  local act_status="$3"
  local exp_ct="$4"
  local act_ct="$5"

  if [[ "$act_status" != "$exp_status" ]]; then
    fail "$label: expected HTTP $exp_status, got $act_status"
  fi
  if [[ -n "$exp_ct" && "$act_ct" != *"$exp_ct"* ]]; then
    fail "$label: expected Content-Type containing '$exp_ct', got '$act_ct'"
  fi
  pass "$label"
}

# ── curl helper: returns "STATUS|CT|PROFILE_HEADER" ─────────────────────────
do_curl() {
  local out
  out=$(curl -s -D - "$@" 2>/dev/null) || true
  local status ct profile
  status=$(echo "$out" | grep -i '^HTTP/' | tail -1 | awk '{print $2}')
  ct=$(echo "$out"     | grep -i '^Content-Type:'       | tail -1 | sed 's/.*: *//' | tr -d '\r')
  profile=$(echo "$out" | grep -i '^X-YAAgents-Profile:' | tail -1 | sed 's/.*: *//' | tr -d '\r')
  echo "${status}|${ct}|${profile}"
}

# ── main ─────────────────────────────────────────────────────────────────────

echo "=== campaign-api-go e2e test suite ==="
echo "Campaign URL : $CAMPAIGN_URL"
echo "Gateway URL  : $GATEWAY_URL"
echo ""

# Wait for both services to be healthy.
wait_healthy "$CAMPAIGN_URL" "campaign-api-go"
wait_healthy "$GATEWAY_URL"  "yaagents-gateway"

echo ""
echo "--- Flow 1: Clarification — missing goal → 400 vnd.yaagents.clarification+json ---"
IFS='|' read -r status ct profile < <(do_curl \
  -X POST "$CAMPAIGN_URL/campaigns/cmp-123/optimizations" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{}')
check_response "flow1:clarification:status" "400" "$status" "" ""
check_response "flow1:clarification:content-type" "400" "$status" "vnd.yaagents.clarification" "$ct"
pass "flow1:clarification:content-type=$ct"

echo ""
echo "--- Flow 2: Created — valid goal → 201 application/json + X-YAAgents-Profile: v0.3 ---"
IFS='|' read -r status ct profile < <(do_curl \
  -X POST "$CAMPAIGN_URL/campaigns/cmp-123/optimizations" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal":"ctr"}')
check_response "flow2:created:status" "201" "$status" "" ""
check_response "flow2:created:content-type" "201" "$status" "application/json" "$ct"
if [[ "$profile" != "v0.3" ]]; then
  fail "flow2:created: X-YAAgents-Profile expected 'v0.3', got '$profile'"
fi
pass "flow2:created:X-YAAgents-Profile=$profile"

echo ""
echo "--- Flow 3: Accepted (async) — Prefer: respond-async → 202 vnd.yaagents.operation+json ---"
IFS='|' read -r status ct profile < <(do_curl \
  -X POST "$CAMPAIGN_URL/campaigns/cmp-123/optimizations" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -H "Prefer: respond-async" \
  -d '{"goal":"ctr"}')
check_response "flow3:accepted:status" "202" "$status" "" ""
check_response "flow3:accepted:content-type" "202" "$status" "vnd.yaagents.operation" "$ct"
pass "flow3:accepted:content-type=$ct"

echo ""
echo "--- Flow 4: Validation failed — invalid type → 422 vnd.yaagents.validation-error+json ---"
IFS='|' read -r status ct profile < <(do_curl \
  -X POST "$CAMPAIGN_URL/campaigns/cmp-123/optimizations" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{"goal":42}')
check_response "flow4:validation-failed:status" "422" "$status" "" ""
check_response "flow4:validation-failed:content-type" "422" "$status" "vnd.yaagents.validation-error" "$ct"
pass "flow4:validation-failed:content-type=$ct"

echo ""
echo "--- Flow 5: Auth failure — no Authorization header via gateway → 401 ---"
IFS='|' read -r status ct profile < <(do_curl \
  -X POST "$GATEWAY_URL/campaigns/cmp-123/optimizations" \
  -H "Content-Type: application/json" \
  -d '{"goal":"ctr"}')
check_response "flow5:auth-failure:status" "401" "$status" "" ""
pass "flow5:auth-failure:status=401"

echo ""
echo "=== All 5 flows PASSED ==="
