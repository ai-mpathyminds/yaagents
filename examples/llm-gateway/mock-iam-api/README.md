# mock-iam-api

Tiny mock tenant-directory service for the yaagents `examples/llm-gateway/`
compose demo.

Implements the lookup-service contract documented in
so the `tenant-injector` plugin
(b) has a working backend in the OSS demo without an external IAM
dependency. The plugin itself is generic; the mock is just one possible
implementation of the contract.

## Endpoint

```
GET /api/v1/principals/{principal}/tenant

200 OK
Content-Type: application/json
{"principal":"user-alice@example.com","tenant_id":"tenant-001"}

404 Not Found      ← principal not in mock-tenants.yaml
```

`{principal}` is URL-encoded by the calling plugin; the mock decodes and
performs an exact-match lookup in its in-memory map.

## Configuration

Driven by `mock-tenants.yaml` (mounted into the container by the demo
compose file):

```yaml
principals:
  "user-alice@example.com": tenant-001
  "user-bob@example.com":   tenant-002
  "service-account-1":      tenant-001
```

Reloaded on SIGHUP (optional convenience; not required by the contract).

## Build

Multi-stage Alpine, non-root, CGO_ENABLED=0, arm64-capable. Image:
`ghcr.io/ai-mpathyminds/yaagents-mock-iam-api:demo` (built locally by the
demo compose; not published to GHCR — demo-only).

## Source

Implemented at runbook entry B-11a (b) alongside the
tenant-injector v2 plugin. Target: ~80 LOC Go binary; single endpoint;
in-memory map; no external dependencies beyond `net/http` and a YAML
parser.

## Why this exists Decision 2 (default-plugin design principle):
> Demo via mock services — the `examples/llm-gateway/` compose demo
> ships a tiny mock-IAM service so the demo runs green out-of-the-box
> without an external IAM dependency.

OSS adopters using the gateway in production replace this mock with their
own tenant-directory service (Auth0, Cognito, Keycloak userinfo, custom
IAM). The plugin contract — `GET .../{principal}/tenant` returning
`{"tenant_id": "..."}` — is intentionally trivial so any service can
implement it.
