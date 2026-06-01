# YAAgents

> **Build the agent however you want. Expose it like a governed API.**

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg?logo=go)](https://golang.org/)
[![Python 3.11+](https://img.shields.io/badge/python-3.11%2B-3776AB.svg?logo=python)](https://python.org/)
[![CI](https://github.com/ai-mpathyminds/yaagents/actions/workflows/ci.yml/badge.svg)](https://github.com/ai-mpathyminds/yaagents/actions)
[![Gateway image](https://img.shields.io/badge/ghcr.io-yaagents--gateway-0d1117.svg?logo=github)](https://ghcr.io/ai-mpathyminds/yaagents-gateway)
[![Profile version](https://img.shields.io/badge/YAAgents%20Profile-v0.2-blueviolet)](spec/)

**YAAgents** is an **Agentic REST Profile** — a response contract, gateway, and client layer for exposing agentic capabilities as ordinary domain resource operations.

```http
POST /campaigns/{id}/optimizations       ← YAAgents pattern
POST /tickets/{id}:triage
POST /claims/{id}/reviews

POST /agents/invoke                      ← what YAAgents replaces
```

Typed responses. Gateway RBAC. OpenAPI contracts. Native clients. Framework-neutral — bring any agent runtime.

---

## Why YAAgents?

Production systems are resource-oriented. Chat is a poor integration surface for business workflows. Agentic capabilities should live behind domain resource operations, not generic `/agents/invoke` endpoints.

YAAgents standardizes:

| Problem today | YAAgents solution |
|---|---|
| Free-form text responses | Typed `application/vnd.yaagents.*` media types |
| Ad hoc clarification | `clarification_required` — a machine-readable status |
| Weak access control | Gateway RBAC with tenant context |
| Framework-specific client code | Native Python and TypeScript client SDKs |
| Hard-to-document agent APIs | OpenAPI-first response contracts |
| Vendor-locked runtimes | Framework-neutral: LangGraph, Pydantic AI, Semantic Kernel, or custom logic |

---

## Architecture

```
┌─────────────────────────────────────────────┐
│  Application / Consumer                     │
│  Python client  ·  TypeScript client  ·  cURL│
└────────────────────┬────────────────────────┘
                     │  HTTP (typed media types)
                     ▼
┌─────────────────────────────────────────────┐
│  YAAgents Gateway  (Go)                     │
│  Authentication · RBAC · Tenant routing     │
│  Audit logging · Correlation-id propagation │
│  /healthz  /readyz  /metrics                │
└────────────────────┬────────────────────────┘
                     │  upstream HTTP
                     ▼
┌─────────────────────────────────────────────┐
│  Agentic API Service  (any language/runtime)│
│  FastAPI + yaagents-fastapi SDK  ← Python   │
│  or Spring Boot · ASP.NET Core · Go Gin …   │
└────────────────────┬────────────────────────┘
                     │  internal call
                     ▼
┌─────────────────────────────────────────────┐
│  Agent Implementation  (your choice)        │
│  LangGraph · Pydantic AI · Semantic Kernel  │
│  LangChain · direct LLM SDK · custom logic  │
└─────────────────────────────────────────────┘
```

### Response status × media-type table (normative)

| HTTP status | Meaning | Media type |
|---|---|---|
| `200 OK` | Synchronous result | `application/vnd.yaagents.result+json` |
| `202 Accepted` | Async task accepted | `application/vnd.yaagents.accepted+json` |
| `206 Partial Content` | Streaming progress | `application/vnd.yaagents.progress+json` |
| `400 Bad Request` | Malformed input | `application/vnd.yaagents.error+json` |
| `412 Precondition Failed` | Clarification required | `application/vnd.yaagents.clarification+json` |
| `422 Unprocessable Entity` | Validation failure | `application/vnd.yaagents.error+json` |

---

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Make (optional, recommended)

### Run the campaign-api demo in one command

```bash
git clone https://github.com/ai-mpathyminds/yaagents.git
cd yaagents/examples/campaign-api
docker compose up
```

Services start on:

| Service | URL |
|---|---|
| YAAgents Gateway | `http://localhost:8120` |
| Campaign API (reference) | `http://localhost:8121` |

### Try a request

```bash
# Trigger an optimization — returns 200 (result) or 412 (clarification required)
curl -X POST http://localhost:8120/campaigns/c-001/optimizations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer dev-token" \
  -d '{"budget_delta": 500}'
```

### Install the clients

```bash
# Python
pip install yaagents-client

# TypeScript / Node
npm install @aimpathyminds/yaagents-client
```

### Python client example

```python
from yaagents_client import YAAgentsClient

client = YAAgentsClient(base_url="http://localhost:8120", token="dev-token")
result = client.post("/campaigns/c-001/optimizations", json={"budget_delta": 500})

if result.is_clarification():
    print("Needs clarification:", result.clarification_fields())
elif result.is_result():
    print("Optimization result:", result.data())
```

### TypeScript client example

```typescript
import { YAAgentsClient } from "@aimpathyminds/yaagents-client";

const client = new YAAgentsClient({ baseUrl: "http://localhost:8120", token: "dev-token" });
const result = await client.post("/campaigns/c-001/optimizations", { budget_delta: 500 });

if (result.isClarification()) {
  console.log("Needs clarification:", result.clarificationFields());
} else if (result.isResult()) {
  console.log("Optimization result:", result.data());
}
```

---

## Repository Layout

```
yaagents/
├── spec/                  ← Agentic REST Response Profile (normative)
├── schemas/               ← JSON schemas for all response types
├── openapi/               ← Reusable OpenAPI components
├── gateway/               ← Go gateway source
├── sdk-fastapi/           ← Python FastAPI server SDK (yaagents-fastapi)
├── client-python/         ← Python client (yaagents-client)
├── client-ts/             ← TypeScript client (@aimpathyminds/yaagents-client)
├── cli/                   ← CLI validator (yaagents-cli)
└── examples/campaign-api/ ← Reference example + Docker Compose demo
```

---

## Published Artifacts

| Artifact | Registry | Install |
|---|---|---|
| Gateway image | GHCR | `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.1.0` |
| Python FastAPI SDK | PyPI | `pip install yaagents-fastapi` |
| Python client | PyPI | `pip install yaagents-client` |
| CLI validator | PyPI | `pip install yaagents-cli` |
| TypeScript client | npm | `npm install @aimpathyminds/yaagents-client` |

---

## Documentation

- [Response Profile spec](spec/)
- [JSON schemas](schemas/)
- [OpenAPI components](openapi/)
- [Gateway configuration](gateway/README.md)
- [FastAPI SDK guide](sdk-fastapi/README.md)
- [Campaign API example walkthrough](examples/campaign-api/README.md)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

**Note:** External contributions are not accepted until legal review of the license is complete. See the banner in CONTRIBUTING.md for details.

---

## License

> **License:** Apache 2.0 — see `LICENSE`. v0.1.x packages shipped under the YAAgents Community License remain under that license (non-retroactive). For questions about historical v0.1.x usage, contact bhaskar@aimpathyminds.com.

---

## Security

See [SECURITY.md](SECURITY.md) for our responsible-disclosure policy.

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
