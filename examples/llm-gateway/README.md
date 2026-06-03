# yaagents LLM Gateway — Reference Example

Demonstrates the yaagents gateway acting as an LLM proxy with the plugin chain
(token-validator → tenant-injector → CORS) and SSE streaming support.

**ADR:** PI2-yaa-0002 (Option A layer-atop; SSE + concurrency + CORS absorbed from ai-platform/ai-gateway)

## Architecture

```
Client → yaagents-gateway (port 8122) → llm-api (internal :8123, mock LLM backend)
                                      → mock-iam-api (internal :8122, tenant directory)
```

The gateway validates the JWT, resolves the tenant via mock-iam-api, and proxies
`POST /completions` to the mock LLM backend using SSE pipe-and-flush semantics.

## Quick Start

```bash
cd examples/llm-gateway
docker compose up
```

> **SECURITY — demo tokens only.** `GATEWAY_JWT_SECRET` is set to a hardcoded
> demo value (`demo-secret-not-for-production`) in `docker-compose.yml` and
> `plugins.yaml`. **Never use this value in production.** For production
> deployments, set `GATEWAY_JWT_JWKS_URL` and remove `test_mode: true` from
> the token-validator plugin (ADR PI1-yaa-0001 §3).

The demo uses a pre-signed HS256 JWT (sub: `user-alice@example.com`, tenant: `tenant-001`):

```bash
DEMO_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLWFsaWNlQGV4YW1wbGUuY29tIiwiZXhwIjo5OTk5OTk5OTk5fQ.uqNGt4LWuwmb_Ky_cBVZu_gu1CKTkWqLmbCXHyQsX4Y"
```

### 1 — Standard LLM call (non-streaming)

```bash
curl -X POST http://localhost:8122/completions \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Generate a campaign headline", "stream": false}'
```

Expected: `201 application/json` with a mocked completion body.

### 2 — SSE streaming response

```bash
curl -N -X POST http://localhost:8122/completions \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"prompt": "Generate a campaign headline", "stream": true}'
```

Expected: progressive `text/event-stream` chunks, terminated by `data: [DONE]`.

### 3 — CORS preflight

```bash
curl -i -X OPTIONS http://localhost:8122/completions \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST"
```

Expected: `200` with `Access-Control-Allow-Origin: http://localhost:3000`; no upstream call.

### 4 — Execution timeout simulation

```bash
curl -X POST http://localhost:8122/completions \
  -H "Authorization: Bearer $DEMO_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "timeout test", "stream": false, "simulate_timeout": true}'
```

Expected: `500 application/vnd.yaagents.error+json` with `code: EXECUTION_TIMEOUT`
after `executionTimeoutSeconds: 30` (routes.yaml).

> **Production note:** replace the demo token with a real JWKS-backed JWT.
> Never use `demo-secret-not-for-production` or `test_mode: true` in any
> environment that handles real data.

## Services

| Service | Internal port | Host port | Purpose |
|---------|--------------|-----------|---------|
| `yaagents-gateway` | 8120 | **8122** | yaagents gateway; proxies /completions |
| `llm-api` | 8123 | — (internal) | Mock LLM backend; canned completions + SSE |
| `mock-iam-api` | 8122 | — (internal) | Tenant directory stub; resolves JWT sub → tenant |

## Plugin chain (plugins.yaml)

1. **token-validator** — HS256 test mode; `demo-secret-not-for-production`
2. **tenant-injector** — resolves `sub` claim → tenant via mock-iam-api; injects `X-Tenant-ID`
3. **cors** — handles `OPTIONS` preflight; injects `Access-Control-Allow-Origin`

Disabled (commented-out): `license-check`, `prompt-sanitize`, `otel-audit`.

## Resource limits (NFR-EX-LLM-1)

| Service | Memory | CPUs |
|---------|--------|------|
| yaagents-gateway | 128 MB | 0.50 |
| llm-api | 256 MB | 0.50 |
| mock-iam-api | 64 MB | 0.25 |

Tested against 4 GB Docker Desktop allocation.

## Conformance

```bash
yaagents conformance-test http://localhost:8122 \
  --require-plugin token-validator \
  --require-plugin tenant-injector
```

(Requires `yaagents` CLI; see `cli/` directory.)
