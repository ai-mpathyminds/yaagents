# YAAgents MVP PRD

## Product

**YAAgents** is an Agentic REST Profile for building deterministic, resource-oriented agentic APIs.

It lets teams expose agentic capabilities through ordinary business resources such as `/campaigns`, `/orders`, `/tickets`, or `/claims` instead of generic chat endpoints or framework-specific agent APIs.

The core idea is simple:

> Keep your domain resources. Make selected operations agentic.

YAAgents does not introduce a new agent runtime in the MVP. Agent implementations can be built using any framework, including LangGraph, Semantic Kernel, Pydantic AI, LangChain, direct LLM SDKs, rules, workflows, or custom application logic.

The MVP focuses on the missing enterprise interface layer:

- Resource-oriented agentic API contracts
- Standard response types
- Standard content types
- OpenAPI response generation
- Gateway-based access control
- Native client handling
- Kubernetes-compatible deployment

---

## Problem

Agentic capabilities are often exposed as generic chat interfaces or framework-specific invocation endpoints.

This creates problems for real product and enterprise systems:

- Inputs and outputs are loosely controlled
- Access control is hard to apply consistently
- Client applications need custom parsing logic
- Error handling is inconsistent
- Clarification flows are unstructured
- APIs are difficult to document and test
- Existing API gateway and microservice practices are bypassed
- Agent frameworks leak into the external integration model

YAAgents solves this by making agentic behavior look and behave like normal REST APIs.

---

## Core Thesis

Do not expose agents as the primary API surface.

Expose domain resources.

For example:

```http
POST /campaigns/{campaignId}/optimizations
POST /campaigns/{campaignId}/assets:generate
POST /campaigns/{campaignId}/readiness-checks
```

Not:

```http
POST /agents/campaign-agent/invoke
```

The agent remains an implementation detail behind a controlled resource operation.

---

## MVP Scope

The MVP contains:

1. **Agentic REST Response Profile**
2. **Go-based AI Gateway**
3. **Python FastAPI server SDK**
4. **Python client SDK**
5. **TypeScript client SDK**
6. **OpenAPI reusable components**
7. **JSON schemas**
8. **CLI validator**
9. **Campaign API reference example**
10. **Docker Compose demo**
11. **Kubernetes deployment manifests**

The MVP does not include:

- Custom agent runtime
- Workflow orchestrator
- Multi-agent planning
- Agent registry
- Memory framework
- Visual builder
- Tool marketplace
- Kubernetes operator
- LangGraph-specific runtime wrapper

---

## Architecture

```text
+-------------------------+
| Python / TS Client SDK  |
+-----------+-------------+
            |
            v
+-------------------------+
| YAAgents Gateway        |
| Go                      |
| auth, RBAC, tenant,     |
| audit, routing          |
+-----------+-------------+
            |
            v
+-------------------------+
| Campaign Agentic API    |
| FastAPI + YAAgents SDK  |
+-----------+-------------+
            |
            v
+-------------------------+
| Agent Implementation    |
| any framework/custom    |
+-------------------------+

Runtime substrate: Kubernetes
```

---

## Component 1: Agentic REST Response Profile

YAAgents defines standard response types for agentic REST operations.

| Response Type | HTTP Status | Content-Type |
|---|---:|---|
| `success` | `200` | `application/json` |
| `created` | `201` | `application/json` |
| `accepted` | `202` | `application/vnd.yaagents.operation+json` |
| `clarification_required` | `400` | `application/vnd.yaagents.clarification+json` |
| `validation_failed` | `422` | `application/vnd.yaagents.validation-error+json` |
| `approval_required` | `412` | `application/vnd.yaagents.approval-required+json` |
| `forbidden` | `403` | `application/vnd.yaagents.error+json` |
| `conflict` | `409` | `application/vnd.yaagents.conflict+json` |
| `failed_dependency` | `424` | `application/vnd.yaagents.error+json` |
| `error` | `500` | `application/vnd.yaagents.error+json` |

The most important response for agentic APIs is `clarification_required`.

Example:

```http
400 Bad Request
Content-Type: application/vnd.yaagents.clarification+json
```

```json
{
  "type": "clarification_required",
  "code": "CLARIFICATION_REQUIRED",
  "message": "Additional information is required.",
  "requiredInputs": [
    {
      "name": "successMetric",
      "location": "body",
      "type": "string",
      "required": true,
      "question": "Which success metric should be optimized?",
      "allowedValues": [
        "ctr",
        "cpl",
        "conversion_rate",
        "lead_quality"
      ]
    }
  ],
  "trace": {
    "correlationId": "corr-123",
    "requestId": "req-456"
  }
}
```

---

## Component 2: Go Gateway

The YAAgents Gateway is a lightweight API gateway for agentic REST APIs.

Responsibilities:

- Authenticate incoming requests
- Extract tenant and actor context
- Enforce route-level RBAC
- Inject standard context headers
- Route requests to backend services
- Normalize response headers
- Emit audit logs
- Preserve typed agentic responses

Example route config:

```yaml
routes:
  - id: create-campaign-optimization
    method: POST
    path: /campaigns/{campaignId}/optimizations
    target: http://campaign-api:8080
    roles:
      - campaign.manager
    tenantRequired: true
    audit: true
```

The gateway is published as a Docker image:

```bash
docker pull ghcr.io/yaagents/gateway:0.1.0
```

---

## Component 3: Python FastAPI SDK

The Python SDK helps developers expose agentic REST operations using FastAPI.

Example:

```python
@router.post("/campaigns/{campaign_id}/optimizations")
@agentic_operation(
    resource="CampaignOptimization",
    operation_kind="recommendation",
    mutating=False,
    roles=["campaign.manager"],
    responses=[
        AgenticResponses.created("CampaignOptimization"),
        AgenticResponses.clarification_required(),
        AgenticResponses.validation_failed(),
        AgenticResponses.failed_dependency()
    ]
)
async def create_optimization(
    campaign_id: str,
    request: OptimizationRequest,
    context: AgenticContext
):
    if not request.success_metric:
        return AgenticResponse.clarification_required(
            message="Additional information is required.",
            required_inputs=[
                RequiredInput(
                    name="successMetric",
                    location="body",
                    type="string",
                    question="Which success metric should be optimized?",
                    allowed_values=[
                        "ctr",
                        "cpl",
                        "conversion_rate",
                        "lead_quality"
                    ],
                    required=True
                )
            ]
        )

    result = await optimizer.recommend(campaign_id, request)
    return AgenticResponse.created(result)
```

The SDK handles:

- Response builders
- HTTP status mapping
- Content-Type mapping
- Trace metadata
- OpenAPI response generation
- Standard error conversion
- Agentic context injection

Published as:

```bash
pip install yaagents-fastapi
```

---

## Component 4: Python Client

The Python client consumes YAAgents-compatible APIs.

Example:

```python
from yaagents_client import YaAgentsClient, ClarificationRequired

client = YaAgentsClient(
    base_url="http://localhost:8080",
    token="demo-token",
    tenant_id="tenant-001"
)

try:
    optimization = client.campaigns("cmp-123").optimizations.create({
        "goal": "reduce_cost_per_lead"
    })
except ClarificationRequired as e:
    print(e.required_inputs)
```

Published as:

```bash
pip install yaagents-client
```

---

## Component 5: TypeScript Client

The TypeScript client consumes YAAgents-compatible APIs from frontend or Node.js applications.

Example:

```ts
import { YaAgentsClient } from "@yaagents/client";

const client = new YaAgentsClient({
  baseUrl: "http://localhost:8080",
  token: "demo-token",
  tenantId: "tenant-001"
});

const result = await client.campaigns
  .byId("cmp-123")
  .optimizations()
  .create({
    goal: "reduce_cost_per_lead"
  });

if (result.type === "clarification_required") {
  renderClarificationForm(result.requiredInputs);
}

if (result.type === "created") {
  console.log(result.resource);
}
```

Published as:

```bash
npm install @yaagents/client
```

---

## Component 6: OpenAPI Components

YAAgents publishes reusable OpenAPI components.

Location:

```text
openapi/
  yaagents-components.yaml
  yaagents-response-profile.yaml
```

These include:

- Standard headers
- Standard response schemas
- Standard media types
- Standard error responses
- `x-yaagents` operation metadata

Example:

```yaml
paths:
  /campaigns/{campaignId}/optimizations:
    post:
      operationId: createCampaignOptimization
      x-yaagents:
        resource: Campaign
        operationKind: recommendation
        deterministic: true
        mutating: false
      responses:
        "201":
          description: Optimization created
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CampaignOptimization"

        "400":
          description: Clarification required
          content:
            application/vnd.yaagents.clarification+json:
              schema:
                $ref: "#/components/schemas/ClarificationRequired"
```

---

## Component 7: JSON Schemas

YAAgents publishes JSON schemas for all standard response types.

Location:

```text
schemas/
  clarification-required.schema.json
  validation-failed.schema.json
  approval-required.schema.json
  conflict.schema.json
  agentic-error.schema.json
  operation-accepted.schema.json
```

These schemas are used by:

- Server SDKs
- Client SDKs
- Gateway
- CLI validator
- Conformance tests
- OpenAPI components

---

## Component 8: CLI

The CLI validates YAAgents-compatible APIs and responses.

Example commands:

```bash
yaagents validate-openapi openapi.yaml
yaagents validate-response clarification-response.json
yaagents conformance-test http://localhost:8080
yaagents init fastapi
```

Expected output:

```text
YAAgents Conformance Report

✓ OpenAPI includes x-yaagents metadata
✓ Clarification response uses correct content type
✓ 400 response matches clarification schema
✓ Correlation ID propagated
✓ Gateway route requires tenant context

Overall: PASS
```

Published as:

```bash
pip install yaagents-cli
```

or through GitHub Releases.

---

## Reference Example: Campaign API

The reference example demonstrates a resource-oriented agentic API.

Endpoints:

```http
POST /campaigns
GET  /campaigns/{campaignId}

POST /campaigns/{campaignId}/optimizations
GET  /campaigns/{campaignId}/optimizations/{optimizationId}

POST /campaigns/{campaignId}/assets:generate
```

Demonstrated flows:

- Successful optimization creation
- Clarification required
- Validation failed
- Failed dependency
- Gateway RBAC failure
- Client typed response handling
- OpenAPI generation

Run locally:

```bash
cd examples/campaign-api
docker compose up
```

Call the API:

```bash
curl -X POST http://localhost:8080/campaigns/cmp-123/optimizations \
  -H "Authorization: Bearer demo-token" \
  -H "X-Tenant-ID: tenant-001" \
  -H "Content-Type: application/json" \
  -d '{
    "goal": "reduce_cost_per_lead"
  }'
```

---

## Publishing Model

### Docker image

```text
ghcr.io/yaagents/gateway:0.1.0
```

### Python packages

```text
yaagents-fastapi
yaagents-client
yaagents-cli
```

Published to PyPI.

### TypeScript package

```text
@yaagents/client
```

Published to npm.

### OpenAPI components

Published through:

```text
GitHub Releases
/openapi directory in repo
future stable URL
```

### JSON schemas

Published through:

```text
GitHub Releases
/schemas directory in repo
future stable URL
```

### Helm chart

Published through:

```text
GHCR OCI registry
```

Example:

```bash
helm install yaagents-gateway oci://ghcr.io/yaagents/charts/yaagents-gateway
```

---

## Kubernetes Runtime Model

YAAgents does not provide a custom runtime.

Agentic APIs run as normal containerized services on Kubernetes.

Kubernetes provides:

- Deployment
- Scaling
- Service discovery
- ConfigMaps
- Secrets
- Health probes
- Rolling updates
- Ingress/Gateway routing
- Observability integration

YAAgents provides:

- Interface profile
- Response contract
- Gateway controls
- Native clients
- OpenAPI metadata

---

## Internal Agent Development Model

YAAgents is framework-neutral.

Developers can build internals using:

- LangGraph
- Semantic Kernel
- Pydantic AI
- LangChain
- OpenAI SDK
- Azure AI Foundry
- Amazon Bedrock
- Rules engines
- Custom service logic

YAAgents only standardizes how the capability is exposed and consumed.

Positioning:

> Bring your own agent. YAAgents makes it a product-grade API.

---

## Versioning

YAAgents uses a profile version plus package versions.

Example:

```text
YAAgents Profile: v0.1
Gateway: v0.1.0
FastAPI SDK: v0.1.0
Python Client: v0.1.0
TypeScript Client: v0.1.0
CLI: v0.1.0
```

Every package must declare the profile version it supports.

Example:

```text
Supports YAAgents Profile v0.1
```

---

## Success Criteria

The MVP is successful when:

1. A developer can expose an agentic REST endpoint using the FastAPI SDK.
2. The endpoint can return `created`, `clarification_required`, `validation_failed`, and `failed_dependency`.
3. The SDK maps each response to the correct status code and content type.
4. OpenAPI includes response-specific content types and schemas.
5. Gateway enforces route-level RBAC.
6. Gateway propagates tenant, actor, request, and correlation context.
7. Python and TypeScript clients handle clarification responses natively.
8. The campaign example runs with Docker Compose.
9. K8s manifests deploy the gateway and example API.
10. CLI validates the OpenAPI and sample responses.

---

## Roadmap

### v0.1

- Agentic REST response profile
- Go gateway
- FastAPI SDK
- Python client
- TypeScript client
- JSON schemas
- OpenAPI components
- CLI validator
- Campaign API example
- Docker Compose
- K8s manifests

### v0.2

- Spring Boot adapter
- ASP.NET Core adapter
- Enhanced conformance tests
- Async operation profile
- Approval-required flow
- Gateway audit sinks

### v0.3

- OpenTelemetry support
- OPA policy integration
- LangGraph adapter plugin
- Semantic Kernel adapter plugin
- Helm chart hardening
- Documentation site

### v1.0

- Stable Agentic REST Profile
- Multiple framework adapters
- Production-grade gateway
- Conformance certification
- Enterprise deployment guide

---

## Final MVP Statement

YAAgents v0.1 provides the interface layer for resource-oriented agentic APIs.

It does not try to own the agent runtime or internal agent framework. It standardizes how agentic operations are exposed, secured, documented, and consumed.

The first release ships:

```text
Go gateway
Python FastAPI SDK
Python client
TypeScript client
CLI validator
OpenAPI components
JSON schemas
Campaign API example
Docker Compose
Kubernetes manifests
```

The promise:

> Build the agent however you want. Expose it like a governed API.
