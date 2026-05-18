# PI1-yaa — Component: Campaign reference example + Compose demo (`examples/campaign-api/`)

Owner lane: python-developer (FastAPI server + routes); platform-engineer
NFR-refines the Compose file at A-4. Sprint 4 (server) + Sprint 5 (Compose
e2e + the reserved conformance gate). ADR: PI1-yaa-0001, PI1-yaa-0002.
Ports: gateway `8120`, campaign-api `8121` (portfolio table).

---

### WI-1yaa.EX-1: Campaign FastAPI server (uses `yaagents-fastapi`) [DONE]
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

### WI-1yaa.EX-2: Gateway `routes.yaml` for the example [DONE]
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

### WI-1yaa.EX-3: Docker Compose demo [DONE] — Sprint 5
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

### WI-1yaa.EX-4: End-to-end conformance gate (PI gate) [READY] — Sprint 5
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

---

## NFR Addendum — A-4 platform-engineer pass (2026-05-17)

Compose file authored at A-4: `examples/campaign-api/docker-compose.yml`
compose-linter result: **11/11 checks PASS** (2026-05-17).

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SEC] no real secrets in compose | **NFR WI below** | WI-1yaa.NFR-EX-2 |
| [SRE] health checks in compose | feature WI | EX-3 (both services) |
| [SRE] resource limits in compose | **NFR WI below** | WI-1yaa.NFR-EX-1 |
| [SRE] named volumes in compose | feature WI | EX-3 (`campaign-data`) |
| [SUPPLY-CHAIN] N/A | N/A | reference example; not published to a registry |
| [FIN] FinOps WI | **N/A** | dev-host/CI product; no cloud run-rate in PI1-yaa |

### WI-1yaa.NFR-EX-1: Compose resource limits [DONE]
service: yaagents/examples/campaign-api
brief: [SRE] `examples/campaign-api/docker-compose.yml` MUST declare
`deploy.resources.limits.memory` + `deploy.resources.limits.cpus` for every
service. Limits: `yaagents-gateway`: memory 128m, cpus 0.50;
`campaign-api`: memory 256m, cpus 0.50. Total ≤ 384 MB — safely within
the 16 GB dev ceiling (PRD §9.1). The compose-linter skill (run at A-4,
must also be clean at EX-3 acceptance) enforces check 3.
acceptance:
- `docker compose up` completes without OOM on a 4 GB Docker Desktop allocation
- `docker stats` shows both services within their configured limits under load
- compose-linter check 3 (resource limits) PASS on the file
library_justify: novel; standalone OSS surface (demo)
depends_on: [WI-1yaa.EX-3]

### WI-1yaa.NFR-EX-2: Demo token discipline in compose [DONE]
service: yaagents/examples/campaign-api
brief: [SEC] The Compose demo MUST make explicit that `GATEWAY_JWT_SECRET`
is a hardcoded demo-only value. Concrete requirements: (1) compose file
carries a `SECURITY NOTE` comment at the top (authored at A-4 — present);
(2) `README.md` quick-start block warns "demo token — never use in production;
set GATEWAY_JWT_JWKS_URL for production deployments" (ADR PI1-yaa-0001 §3);
(3) no `.env` file with a real secret is committed (`git-secrets` / pre-commit
hook in REL-1 SECURITY.md guards this); (4) `GATEWAY_JWT_SECRET` default value
MUST be the literal string `demo-secret-not-for-production` — never a valid-looking
random token that could be mistaken for a real credential.
acceptance:
- `grep -r GATEWAY_JWT_SECRET .env` fails (no .env in repo)
- README quick-start block contains "never use in production" near the demo curl
- `GATEWAY_JWT_SECRET` value in compose = `demo-secret-not-for-production` (string-checked in CI)
library_justify: novel; standalone OSS surface (demo)
depends_on: [WI-1yaa.EX-3, WI-1yaa.REL-1]
