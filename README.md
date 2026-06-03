# YAAgents

> **Build the agent however you want. Expose it like a governed API.**

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go 1.22+](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg?logo=go)](https://golang.org/)
[![Python 3.11+](https://img.shields.io/badge/python-3.11%2B-3776AB.svg?logo=python)](https://python.org/)
[![CI](https://github.com/ai-mpathyminds/yaagents/actions/workflows/ci.yml/badge.svg)](https://github.com/ai-mpathyminds/yaagents/actions)
[![Gateway image](https://img.shields.io/badge/ghcr.io-yaagents--gateway-0d1117.svg?logo=github)](https://ghcr.io/ai-mpathyminds/yaagents-gateway)
[![Profile version](https://img.shields.io/badge/YAAgents%20Profile-v0.3-blueviolet)](spec/)

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
| Framework-specific client code | Native Python, TypeScript, and Go client SDKs |
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
│  Go net/http + yaagents-sdk-go   ← Go       │
│  or Spring Boot · ASP.NET Core · Express …  │
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

### Install

```bash
# Python SDK + client + CLI (server SDK, consumer client, validator)
pip install yaagents-fastapi yaagents-client yaagents-cli

# TypeScript / Node client
npm install @aimpathyminds/yaagents-client

# Go client SDK
go get github.com/ai-mpathyminds/yaagents-client-go

# Go server SDK
go get github.com/ai-mpathyminds/yaagents-sdk-go

# Gateway container image
docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0
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
├── spec/                    ← Agentic REST Response Profile (normative)
├── schemas/                 ← JSON schemas for all response types
├── openapi/                 ← Reusable OpenAPI components
├── docs/                    ← Pages site source (Astro Starlight)
├── gateway/                 ← Go gateway source                     [submodule]
├── sdk-fastapi/             ← Python FastAPI server SDK (yaagents-fastapi) [submodule]
├── sdk-go/                  ← Go server SDK (yaagents-sdk-go)       [submodule]
├── client-python/           ← Python client (yaagents-client)       [submodule]
├── client-ts/               ← TypeScript client (@aimpathyminds/yaagents-client) [submodule]
├── client-go/               ← Go client SDK (yaagents-client-go)    [submodule]
├── cli/                     ← CLI validator (yaagents-cli)           [submodule]
└── examples/                ← Reference examples + Docker Compose demos
```

---

## Published Artifacts

| Artifact | Registry | Install |
|---|---|---|
| Gateway image | GHCR | `docker pull ghcr.io/ai-mpathyminds/yaagents-gateway:0.3.0` |
| Python FastAPI SDK | PyPI | `pip install yaagents-fastapi==0.3.0` |
| Python client | PyPI | `pip install yaagents-client==0.3.0` |
| CLI validator | PyPI | `pip install yaagents-cli==0.3.0` |
| TypeScript client | npm | `npm install @aimpathyminds/yaagents-client@0.3.0` |
| Go client SDK | Go modules | `go get github.com/ai-mpathyminds/yaagents-client-go@v0.3.0` |
| Go server SDK | Go modules | `go get github.com/ai-mpathyminds/yaagents-sdk-go@v0.3.0` |

---

## Documentation

- [Full documentation site](https://ai-mpathyminds.github.io/yaagents/)
- [Response Profile spec](spec/)
- [JSON schemas](schemas/)
- [OpenAPI components](openapi/)
- [Gateway configuration](gateway/README.md)
- [FastAPI SDK guide](sdk-fastapi/README.md)
- [Go server SDK guide](sdk-go/README.md)
- [Campaign API example (Python)](examples/campaign-api/README.md)
- [Campaign API example (Go)](examples/campaign-api-go/README.md)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

Contributions are welcome under the Apache 2.0 license via the
Developer Certificate of Origin (DCO) sign-off process. All pull requests
must carry a `Signed-off-by:` trailer. See CONTRIBUTING.md for the full
checklist, plugin contribution path, and legal disclaimer.

---

## License

> **License:** Apache 2.0 — see `LICENSE`. v0.1.x packages shipped under the YAAgents Community License remain under that license (non-retroactive). v0.2.x+ ships under Apache 2.0. For questions about historical v0.1.x usage, contact bhaskar@aimpathyminds.com.

---

## Security

See [SECURITY.md](SECURITY.md) for our responsible-disclosure policy.

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
