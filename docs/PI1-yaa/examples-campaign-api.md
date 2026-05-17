# PI1-yaa — Component: Campaign reference example + Compose demo (`examples/campaign-api/`)

Owner lane: python-developer (FastAPI server + routes); platform-engineer
NFR-refines the Compose file at A-4. Sprint 4 (server) + Sprint 5 (Compose
e2e + the reserved conformance gate). ADR: PI1-yaa-0001, PI1-yaa-0002.
Ports: gateway `8120`, campaign-api `8121` (portfolio table).

---

### WI-1yaa.EX-1: Campaign FastAPI server (uses `yaagents-fastapi`) [DRAFT]
service: yaagents/examples/campaign-api
brief: FastAPI app exposing PRD §6.1 endpoints (`POST /campaigns`,
`GET /campaigns/{id}`, `POST /campaigns/{id}/optimizations`,
`GET .../optimizations/{opId}`, `POST /campaigns/{id}/assets:generate`) via
`@agentic_operation`. Deterministic in-memory state. Demonstrates the §6.2
flows: success/created, clarification_required (missing `successMetric`),
validation_failed (bad types), failed_dependency (simulated LLM-down toggle).
acceptance:
- All 5 endpoints respond with correct status+vendor content-type per §4
- The 4 flows reproducible via documented curl; `/openapi.json` passes `validate-openapi`
library_justify: novel; standalone OSS surface (reference example)
depends_on: [WI-1yaa.SDK-3]

### WI-1yaa.EX-2: Gateway `routes.yaml` for the example [DRAFT]
service: yaagents/examples/campaign-api
brief: `examples/campaign-api/routes.yaml` mapping the §6.1 paths to
`http://campaign-api:8121`, with `roles:` on the optimizations route (so a
token missing the role demonstrates the §6.2 gateway-RBAC 403 flow),
`tenantRequired:true`, `audit:true` on optimizations.
acceptance:
- Loads clean in the gateway; RBAC-fail path returns `403 application/vnd.yaagents.error+json`
- tenant-missing path rejected; audit event emitted for optimizations
library_ref: ADR PI1-yaa-0001
depends_on: [WI-1yaa.GW-5, WI-1yaa.EX-1]

### WI-1yaa.EX-3: Docker Compose demo [DRAFT] — Sprint 5
service: yaagents/examples/campaign-api
brief: `examples/campaign-api/docker-compose.yml` — `yaagents-gateway` (image
or built; `8120:8080`) ↔ `campaign-api` (internal `8121`). HS256 demo token
(ADR 0001 §3), `routes.yaml` bind-mount, health checks, named volumes,
ports in the 8120–8129 band. `docker compose up` brings both green; PRD §6.4
quick-start curls work. platform-engineer NFR-refines (compose-linter) at A-4.
acceptance:
- `docker compose up` → both healthy; §6.4 clarification curl returns 400 vendor body
- Runs within the 16 GB dev ceiling; `compose-linter` clean (A-4)
library_justify: novel; standalone OSS surface (demo)
depends_on: [WI-1yaa.EX-2, WI-1yaa.GW-5]

### WI-1yaa.EX-4: End-to-end conformance gate (PI gate) [DRAFT] — Sprint 5
service: yaagents/examples/campaign-api
brief: The PI1-yaa acceptance gate. With the Compose demo up: run
`yaagents conformance-test http://localhost:8120` → `Overall: PASS`; exercise
all 4 §6.2 flows through the gateway with RBAC enforced + correlation-id
propagated; run the Python + TS client against the live demo asserting typed
handling. Verify the **12 PRD §12 success criteria** as a checklist; verify
no component redefines the §4 table (grep `spec/` is sole source).
acceptance:
- `conformance-test` PASS; all 4 flows green through gateway; both clients handle clarification natively
- PRD §12 1–9 checklist all ✓ (10–12 publish criteria verified by REL-3/4/5 in Sprint 5)
- grep proves §4 table appears only in `spec/` (no paraphrase elsewhere)
library_justify: novel; standalone OSS surface (conformance gate)
depends_on: [WI-1yaa.CLI-4, WI-1yaa.PYC-3, WI-1yaa.TSC-3, WI-1yaa.EX-3]
