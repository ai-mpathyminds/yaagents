# PI2-yaa — Component: LLM gateway reference example (`examples/llm-gateway/`)

Owner lane: **go-developer** (mock-LLM backend + config) + **platform-engineer**
(Compose file at A-4). Sprints 4 + 5. ADR: PI2-yaa-0002 (Option A
layer-atop; the example proves the absorbed LLM specialisation has a
working consumer).

> **Library gate:** all WIs carry `library_ref: ADR PI2-yaa-0002` (the
> absorbed LLM specialisation is the abstraction being demonstrated here).

Topology (PRD §7.2):
```
yaagents-gateway (port 8122) ←→ llm-api (internal mock LLM backend)
```
Port 8122 chosen per the campaign-api retention at 8121 (PRD §7.2 +
PRD §11 OQ-4). Platform-engineer confirms at A-4 compose-linter pass.

---

### WI-2yaa.EX-LLM-1: Mock LLM backend + routes.yaml + plugins.yaml [DONE] — Sprint 4
service: yaagents/examples/llm-gateway
parent_feature: F-LLM
brief: Create `examples/llm-gateway/` directory with:
- `mock-llm-api/` — Go binary (could be Python; Go preferred for stdlib +
  no extra runtime dep). HTTP server implementing two endpoints:
  - `POST /completions` (JSON): returns canned LLM response per the
    `model` field in the request; supports `stream: false` (one-shot
    JSON) and `stream: true` (SSE chunks). Can simulate streaming delay
    (configurable) and timeout (sleep > execution_timeout). Returns the
    standard PRD §4 `201 application/json` for success.
  - `GET /healthz`, `GET /readyz` (always 200 once ready).
- `routes.yaml` — declares `/completions` route with `mode: sse`, target
  `http://llm-api:8123`, `tenantRequired: true`, `executionTimeoutSeconds: 30`,
  CORS plugin enabled with `allowed_origins: [http://localhost:3000]`.
- `plugins.yaml` — enables token-validator (HS256 test_mode with a known
  demo secret), tenant-injector (allowlist `tenant-001, tenant-002`),
  license-check disabled (commented-out + rationale: demo skips license),
  prompt-sanitize disabled (stub), otel-audit disabled (stub),
  CORS plugin enabled.
- `mock-llm-api/Dockerfile` — multi-stage Alpine, non-root, `CGO_ENABLED=0`
  + SPDX header on source files.
acceptance:
- `cd examples/llm-gateway && go run ./mock-llm-api/` then `curl -X POST
  http://localhost:8123/completions -d '{"stream":false,"prompt":"hi"}'`
  returns a 201 JSON body with a mocked completion
- SSE mode (`-H "Accept: text/event-stream" -d '{"stream":true,...}'`) emits
  progressive `text/event-stream` chunks (verified with `curl -N`)
- routes.yaml validates against the gateway's route schema (gateway boots
  cleanly with this routes.yaml)
- plugins.yaml validates against gateway's plugins schema; token-validator
  enabled cannot be disabled
- mock-llm-api Dockerfile builds; image runs as non-root
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.LLM-4]

### WI-2yaa.EX-LLM-2: docker-compose.yml + dual-network topology [READY] — Sprint 4
service: yaagents/examples/llm-gateway
parent_feature: F-LLM
brief: `examples/llm-gateway/docker-compose.yml` with two services:
- `yaagents-gateway` (`image: ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0`):
  port `8122:8120` (gateway internal port 8120; host port 8122 per OQ-4 +
  PRD §7.2); mounts `./routes.yaml` + `./plugins.yaml`; env
  `GATEWAY_PORT=8120 GATEWAY_ROUTES_FILE=/cfg/routes.yaml
  GATEWAY_PLUGINS_FILE=/cfg/plugins.yaml GATEWAY_JWT_SECRET=<demo-secret>`;
  health check `GET /healthz`; depends_on `llm-api` healthy.
- `llm-api` (built locally from `./mock-llm-api/Dockerfile`): port `8123`
  (internal-only — NOT exposed to host); health check `GET /healthz`;
  named volume `mock-llm-data` (if mock keeps state); resource limits
  (`cpus: 0.5`, `mem: 256m` advisory — platform-engineer tunes at A-4).
- README.md at `examples/llm-gateway/README.md` with the PRD §7.2 quick-start
  block verbatim (the four `curl` examples).
acceptance:
- `docker compose config` exits 0; both services have health checks
- Named volume present
- `compose-linter` (portfolio skill, run by platform-engineer at A-4) clean
  on the file (pinned image tags, health checks, resource limits, named
  volumes — `portfolio-conventions.md`)
- README quick-start commands work copy-paste once images are pulled
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.EX-LLM-1]

### WI-2yaa.EX-LLM-3: Compose e2e — all 5 §13.2 flows green [READY] — Sprint 5
service: yaagents/examples/llm-gateway
parent_feature: F-LLM
brief: End-to-end demo + automated test harness that exercises all five
PRD §13.2 flows against the running `docker compose up` topology:
1. **Standard call** — JWT + tenant + non-streaming → 201 JSON.
2. **SSE streaming** — JWT + tenant + `Accept: text/event-stream` → progressive SSE chunks.
3. **SSE concurrency exceeded** — 11 concurrent SSE requests from `tenant-001`
   (limit 10) → 11th returns 429 vendor-error body.
4. **Execution timeout** — request that causes mock-llm-api to sleep > 30 s
   → 500 vendor-error with `code: EXECUTION_TIMEOUT`.
5. **CORS preflight** — OPTIONS from `Origin: http://localhost:3000` →
   200 with `Access-Control-Allow-Origin: http://localhost:3000`.
6. **(Bonus) Community plugin** — operator builds a custom gateway binary
   that imports an example community plugin (`examples/llm-gateway/community-plugin/`)
   and runs the same compose; verify the plugin appears in the chain
   (its custom response header is observed downstream). This proves PRD
   §6.6 — community plugin authoring contract.

Implementation: shell script `examples/llm-gateway/test-e2e.sh` that runs
the 6 flows via `curl`/`bash` and exits 0 on all-pass, 1 otherwise. Also
run `yaagents conformance-test http://localhost:8122` (CLI-CONF) and
assert exit 0. PI-GATE in `release-and-publish.md` invokes both.
acceptance:
- `docker compose up -d` brings both services healthy within 60 s
- `./test-e2e.sh` exits 0 with all 6 flows PASS
- `yaagents conformance-test http://localhost:8122 --require-plugin token-validator --require-plugin tenant-injector` exits 0
- The bonus community-plugin path produces a build with the plugin loaded — verified by `docker logs` showing the plugin's registration log line
library_ref: ADR PI2-yaa-0002
depends_on: [WI-2yaa.EX-LLM-2, WI-2yaa.CLI-CONF, WI-2yaa.REL-3]

---

## NFR Addendum — A-4 platform-engineer pass (2026-06-01)

### OQ-4 port resolution (CONFIRMED)

`examples/campaign-api/` retains host port **8121** (PI1-yaa allocation; unchanged).
`examples/llm-gateway/` host port = **8122** (next free in the 8120–8129
yaagents band per `.claude/rules/portfolio-conventions.md` Port Allocation).
The compose file `examples/llm-gateway/docker-compose.yml` (EX-LLM-2) MUST
map `8122:8120` (gateway internal port 8120; host port 8122). OQ-4 is
closed. **Portfolio port table action**: row `8122 | yaagents llm-gateway (ref example) | yaagents`
must be added to `.claude/rules/portfolio-conventions.md` — this requires
chief-architect authority (outside yaagents platform-engineer writable lane);
surfaced in `## Handoff` below.

### compose-linter status

`examples/llm-gateway/docker-compose.yml` does NOT yet exist (created by
WI-2yaa.EX-LLM-2 in Sprint 4). Compose-linter MUST be run against that
file before EX-LLM-2 acceptance is declared; it is wired as a CI step
(`compose-lint-llm-gateway`) and as an explicit acceptance criterion in
EX-LLM-2 above. The linter must return 0 FAIL on all 6 checks (image
pinning for `ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` pinned ✓;
health checks on both services ✓; resource limits on both services ✓;
named volumes ✓; host port 8122 in portfolio table — pending chief-arch
table update ✓; named network `yaagents-llm-net` ✓).

### NFR dimension coverage

| Dimension | Status | Covered by |
|-----------|--------|------------|
| [SRE] compose health checks | feature WI | EX-LLM-2 (both services; `GET /healthz`) |
| [SRE] resource limits | **NFR WI below** | WI-2yaa.NFR-EX-LLM-1 |
| [SRE] named volumes | feature WI | EX-LLM-2 (`mock-llm-data` named volume) |
| [SRE] compose-linter clean | **deferred to EX-LLM-2 landing** | CI step `compose-lint-llm-gateway`; 0 FAIL gate on file creation |
| [SEC] GATEWAY_JWT_SECRET demo discipline | **NFR WI below** | WI-2yaa.NFR-EX-LLM-2 |
| [SEC] mock-llm-api non-root image | feature WI | EX-LLM-1 (Dockerfile: non-root, `CGO_ENABLED=0` Alpine multi-stage) |
| [FIN] FinOps WI | **N/A** | dev-host only; no cloud resources; Docker Desktop/Podman local; zero cloud cost |

### WI-2yaa.NFR-EX-LLM-1: Compose resource limits — llm-gateway demo [READY]
service: yaagents/examples/llm-gateway
parent_feature: F-LLM
brief: [SRE] `examples/llm-gateway/docker-compose.yml` (EX-LLM-2) MUST
declare `deploy.resources.limits.memory` + `deploy.resources.limits.cpus`
for every service. Limits: `yaagents-gateway`: memory 128m, cpus 0.50;
`llm-api`: memory 256m, cpus 0.50. Total ≤ 384 MB — safely within the
16 GB dev ceiling (PRD §9.1). The compose-linter skill enforces check 3
when the file lands. Image tag for `yaagents-gateway` service MUST be
pinned to `ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` (not `:latest`)
per portfolio-conventions.md check 1.
acceptance:
- compose-linter check 3 (resource limits) PASS on `examples/llm-gateway/docker-compose.yml`
- compose-linter check 1 (image pinning) PASS: `ghcr.io/ai-mpathyminds/yaagents-gateway:0.2.0` not `:latest`
- `docker compose up` completes without OOM on a 4 GB Docker Desktop allocation
library_justify: novel; standalone OSS surface (llm-gateway demo).
depends_on: [WI-2yaa.EX-LLM-2]

### WI-2yaa.NFR-EX-LLM-2: Demo token discipline in llm-gateway compose [READY]
service: yaagents/examples/llm-gateway
parent_feature: F-LLM
brief: [SEC] The Compose demo MUST make explicit that `GATEWAY_JWT_SECRET`
is a hardcoded demo-only value. Concrete requirements: (1) compose file
carries a `SECURITY NOTE` comment at the top (matching PI1-yaa campaign-api
precedent); (2) `examples/llm-gateway/README.md` quick-start block warns
"demo token — never use in production; set GATEWAY_JWT_JWKS_URL for
production deployments" (ADR PI1-yaa-0001 §3 carries forward); (3) no
`.env` file with a real secret committed; (4) `GATEWAY_JWT_SECRET` env-var
value in compose MUST be the literal string `demo-secret-not-for-production`
— never a valid-looking random token; (5) the token-validator plugin config
(`plugins.yaml`) sets `test_mode: true` with the same known demo secret —
explicitly commented as demo-only in the YAML file.
acceptance:
- `grep -r GATEWAY_JWT_SECRET .env` fails (no `.env` in repo)
- README quick-start block contains "never use in production" near the demo curl commands
- `GATEWAY_JWT_SECRET` in compose = `demo-secret-not-for-production` (string-checked in CI)
- `plugins.yaml` `test_mode: true` comment present
- compose file carries `SECURITY NOTE` comment at top (same pattern as campaign-api)
library_ref: ADR PI1-yaa-0001 (demo-secret discipline; carries forward)
depends_on: [WI-2yaa.EX-LLM-2]
