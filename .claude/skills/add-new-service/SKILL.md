---
name: add-new-service
description: Scaffold a new ECS Fargate service in portfolio/infrastructure for platform-engineer — Terraform + ALB + ECR + GHA OIDC + Dockerfile arm64. Idempotent guards; PR-ready commit plan.
---

# add-new-service

Invoked by **portfolio-tier `platform-engineer`** when standing up a brand-new ECS Fargate service. Co-callable by `{product}-architect` at A-3 as a planning aid; commits land in execution.

## Why a skill (not freehand)

A new service touches **7 correlated artifacts across 4 directories** (`roots/compute/`, `roots/network/`, the service repo `.github/workflows/`, the service repo `Dockerfile`). Authoring freehand consistently drops at least one — typical drops: the GHA OIDC deploy role, the listener-rule priority (collision in the band), or the SSM scope on the task role. The skill compresses the 7-artifact set into one commit plan so review is uniform and idempotency is mechanically checkable.

The companion skill `modify-service-tf` handles the **edit** path (correlated-file coherence on shifts). `add-new-service` is **create-only**.

## Inputs (deterministic — all required unless marked optional)

| # | Field | Type | Notes |
|---|-------|------|-------|
| 1 | `service-name` | string (kebab) | e.g. `foo`, `audit`. Lowercase, alphanumeric + dash. Used as Service Connect alias, log group path, resource-name suffix. ≤20 chars. |
| 2 | `product` | enum | `oppor` \| `platform-services` \| `ai-platform` \| `portfolio-shared` |
| 3 | `language` | enum | `go` \| `python` \| `react` — drives Dockerfile template selection |
| 4 | `port` | int | TCP port. MUST be unallocated per `.claude/rules/portfolio-conventions.md §Port Allocation` AND in the product's band (oppor 8080–8089; platform-services 8090–8095; ai-platform 8096–8103) |
| 5 | `feature` | string | Feature tag (e.g. `FX.PE`, `F1.2`) — caller-supplied |
| 6 | `wi-id` | string | Originating WI (e.g. `WI-14.SKILL-1`) — caller-supplied; lands on every resource `wi_id` tag |
| 7 | `repo-owner` | string (optional) | GitHub org segment. Default `ai-mpathyminds`. Used in OIDC `sub` condition |
| 8 | `repo-name` | string (optional) | GitHub repo. Default = `service-name` |
| 9 | `base-path` | string (optional) | ALB path prefix. Default `/api/v1/<service-name>` |
| 10 | `host-header` | string (optional) | If set, listener rule is host-based (Keycloak pattern: `auth.aimpathyminds.com`); priority then moves to 10–19 band |
| 11 | `cpu` / `memory` | int (optional) | Defaults 256 / 512. Values >1024 require an architect ADR — skill rejects and surfaces the ADR requirement |
| 12 | `owner` | string (optional) | Owner tag. Default `platform-engineer` |
| 13 | `health-check-path` | string (optional) | Default `/healthz` (Go/Python). React default `/`. Keycloak-like services override |
| 14 | `enable-execute-command` | bool (optional) | Default false. Set true only for one-time-bootstrap services (Keycloak pattern) |

Inputs that fail pre-flight halt with an error block listing **all** failed checks (not just the first); caller fixes all at once. Partial commit plans are never emitted.

## Generated artifacts — the 7-artifact set

Each numbered artifact below corresponds to one template under `.claude/skills/add-new-service/templates/`. The skill composes templates with input substitution and emits a markdown commit plan with one section per artifact.

### Artifact 1 — `roots/compute/services_<service>.tf` (NEW; Terraform — ecs-fargate-service module call)
Source: `templates/01_ecs_service.tf.tmpl`. Module path `../../modules/ecs-fargate-service`. Inputs filled per the module's `variables.tf` contract: `service_name`, `product`, `cluster_name = aws_ecs_cluster.shared.name`, `container_image = <account-id>.dkr.ecr.<region>.amazonaws.com/ampy-<product-prefix>-prod-ecr-<service>:bootstrap` (`:bootstrap` placeholder replaced on first GHA run), `container_port`, `cpu`/`memory`, `task_role_arn`, `execution_role_arn = aws_iam_role.fargate_execution.arn`, `target_group_arn`, `private_subnet_ids` + `security_group_ids` from network remote state, `service_connect_enabled = true` (Service Connect alias = service name), `health_check_path`, and the 4 caller-supplied tags (`owner`, `feature`, `wi_id`, plus `product` already passed).

### Artifact 2 — `roots/compute/target_groups.tf` (APPEND) + `listener_rules.tf` (APPEND)
Source: `templates/02_target_group.tf.tmpl` + `templates/02_listener_rule.tf.tmpl`. TG `ampy-<product-prefix>-prod-tg-<service>`, `port`, `target_type = ip`, HTTP, default health-check (path = `health-check-path`; `healthy_threshold=2`, `unhealthy_threshold=3`, `interval=30`, `timeout=5`, matcher `200`). Listener-rule priority auto-assigned to next-free integer in the product's band (oppor 20–29; platform-services 10–19; ai-platform 40–49; ui catch-alls 30); band exhaustion → halt. Path-based default `path_pattern = ["{base-path}", "{base-path}/*"]`. Host-header override available via input #10.

### Artifact 3 — `roots/network/security_groups.tf` (APPEND)
Source: `templates/03_sg_rules.tf.tmpl`. Two rules emitted when `service_connect_enabled = true` (current default):

1. **ALB ingress rule** — `aws_security_group_rule.fargate_ingress_<service>`: tcp `port` from `sg_alb`; no public or direct cross-ENI ingress. If `sg_fargate` already carries a blanket `sg_alb → sg_fargate` rule, this is a NO-OP and the idempotency check flags it — caller decides whether to skip (default) or emit for explicitness.
2. **SC self-referential ingress rule** — `aws_security_group_rule.sc_self_<service>`: tcp `port`, `self = true`. Required by `cloud-iac-conventions-aws.md §23` — the Service Connect proxy routes intra-cluster calls to ENI IPs within the same SG; without this rule the SYN is silently dropped at the SG layer (10s × 3 retries → ALB 502). `terraform-conventions-linter` check #14 fails the PR if this rule is absent for any SC-enabled service.

### Artifact 4 — `roots/compute/ecr_<service>.tf` (NEW)
Source: `templates/04_ecr_repo.tf.tmpl`. Module path `../../modules/ecr-repo`. Repo `ampy-<product-prefix>-prod-ecr-<service>`. Lifecycle policy inherits module defaults: keep last 10 semver-tagged (`v*`), expire untagged after 7 days, expire any remaining after 90 days. IMMUTABLE tags + scan-on-push + AES256 (all module defaults; not overridable here — modify-service-tf path for changes).

### Artifact 5 — `roots/compute/iam.tf` (APPEND; task role + GHA OIDC deploy role)
Source: `templates/05_iam_roles.tf.tmpl`. **Task role** `ampy-<product-prefix>-prod-iam-task-<service>`: trust `ecs-tasks.amazonaws.com`; inline policy = `ssm:GetParameter*` on `/ampy/<product>/prod/<service>/*` + `kms:Decrypt` on `alias/aws/ssm` via `kms:ViaService` condition. **GHA deploy role** `ampy-<product-prefix>-prod-iam-gha-deploy-<service>`: trust `token.actions.githubusercontent.com` with `sub` pinned to `repo:<repo-owner>/<repo-name>:ref:refs/heads/main`; inline policy = `ecr:GetAuthorizationToken` (global) + ECR push actions scoped to the service's ECR repo ARN + `ecs:UpdateService`/`DescribeServices` scoped to the cluster's service ARN. **No long-lived AWS keys; OIDC only.**

### Artifact 6 — `<service-repo>/.github/workflows/deploy.yml` (NEW)
Source: `templates/06_deploy_workflow.yml.tmpl`. Triggers: `push` on `main` with paths filter scoped to the service directory. Permissions: `id-token: write`, `contents: read` (OIDC requirement). Steps: checkout → `aws-actions/configure-aws-credentials@v4` (assume GHA deploy role via OIDC) → docker buildx (arm64) → ECR login → push with SHA tag → `aws ecs update-service --force-new-deployment` → `aws ecs wait services-stable`. **No `${{ secrets.AWS_ACCESS_KEY_ID }}`** — only `AWS_ACCOUNT_ID` (non-secret) needed; OIDC role-to-assume is the auth surface.

### Artifact 7 — `<service-repo>/Dockerfile` (NEW; multi-stage arm64)
Source: language-keyed template (`templates/07_dockerfile_go.tmpl` | `_python.tmpl` | `_react.tmpl`).
- **Go**: `golang:1.22-alpine` build → `alpine:3.19` runtime, `CGO_ENABLED=0 GOARCH=arm64`, non-root user, `HEALTHCHECK` via `wget` against `health-check-path`, `ENTRYPOINT ["/app/server"]`.
- **Python**: `python:3.12-slim-bookworm` build (uv-installed wheels) → same image runtime, non-root, `curl` HEALTHCHECK, `ENTRYPOINT ["python", "-m", "app"]`.
- **React**: `node:20-alpine` build (pnpm default; detect lockfile to swap) → `nginx:alpine` runtime, non-root, read-only fs prep, `nginx.conf` bundled, HEALTHCHECK via `wget`.

## 7-tag set enforcement

Every emitted resource block carries the union of the 7 required tags per `cloud-iac-conventions-aws.md` §10:

| Tag | Source |
|-----|--------|
| `product` | input #2 |
| `environment` | provider `default_tags` (`prod` for PI13–PI15) — NOT in resource block |
| `owner` | input #12 (default `platform-engineer`) |
| `cost_center` | provider `default_tags` — NOT in resource block |
| `managed_by` | provider `default_tags` (literal `terraform`) — NOT in resource block |
| `feature` | input #5 |
| `wi_id` | input #6 |

Resource blocks emit only the 4 per-resource tags (`product`, `owner`, `feature`, `wi_id`); the other 3 inherit via provider `default_tags` in the root config. The skill does NOT duplicate them per resource — this matches `terraform-conventions-linter` check #4's expected union. Resource types AWS does not permit tagging on are exempt per `EXEMPTIONS.md` next to the linter.

## Idempotency contract — one WARN per logical artifact

Before emitting each of the 7 artifacts, the skill checks Terraform state (`terraform state list` in the affected root) + filesystem (the affected file paths). If **any** sub-resource of an artifact already exists, that entire artifact is skipped and emits exactly one WARN line:

```
WARN: artifact <N> (<short-description>) already exists — skipping
```

Decision matrix per artifact (state-check probe in parens):

| # | Pre-existence probe | Skip when |
|---|---------------------|-----------|
| 1 | `module.ecs_<service>` in compute state OR `services_<service>.tf` on disk | either present |
| 2 | `aws_lb_target_group.<service>` OR `aws_lb_listener_rule.api_<service>` in compute state | either present |
| 3 | `aws_security_group_rule.fargate_ingress_<service>` in network state OR blanket sg_alb→sg_fargate already covers; AND `aws_security_group_rule.sc_self_<service>` in network state | all present |
| 4 | `module.ecr_<service>` in compute state OR `ecr_<service>.tf` on disk | either present |
| 5 | `aws_iam_role.task_<service>` OR `aws_iam_role.gha_deploy_<service>` in compute state | either present |
| 6 | `<service-repo>/.github/workflows/deploy.yml` on disk | present |
| 7 | `<service-repo>/Dockerfile` on disk | present |

Outcomes:

| Pre-existence pattern | Output | AUDIT verb |
|------------------------|--------|------------|
| No artifacts present | Full commit plan (all 7 sections) | `add-new-service` |
| Some artifacts present | Only the missing artifacts emitted; WARN line per skipped artifact | `add-new-service-partial` |
| All 7 artifacts present | Exactly 7 WARN lines, zero diffs | `add-new-service-noop` |

The skill **never** deletes, modifies, or re-emits an existing resource. Modifications route through `modify-service-tf`.

## Pre-flight validations (halt on any fail; aggregate errors)

1. **Port allocation** — `port` is not assigned to a different service in `portfolio-conventions.md §Port Allocation` AND falls within the product's band. Collision → halt; emit "port X used by `<service>`; suggest next free: Y".
2. **Listener-rule priority free** — scan existing `listener_rules.tf` for the next-free integer in the product's band; band exhausted → halt with rationale.
3. **Naming** — resolved `ampy-<product-prefix>-prod-<resource>-<service>` ≤ 63 AWS char limit; product-prefix per `cloud-iac-conventions-aws.md` §9 (`oppor`→`oppor`, `platform-services`→`plt`, `ai-platform`→`aip`, `portfolio-shared`→`shared`).
4. **7-tag completeness** — caller supplied `feature` + `wi_id`; `owner` resolves to default if omitted; provider `default_tags` reachable. Missing → halt with list.
5. **cpu/memory ceiling** — `cpu ≤ 1024` AND `memory ≤ 1024` (intake constraint #2; module validation duplicates this but the skill catches earlier). Exceeding → halt and emit ADR-requirement boilerplate.
6. **Language template available** — `language ∈ {go, python, react}`. Anything else → halt with the enum list.
7. **Repo identity** — `repo-owner` + `repo-name` non-empty; resolved OIDC `sub` well-formed (`repo:<owner>/<name>:ref:refs/heads/main`).
8. **Health-check path coherence** — for React with default `/`, warn-only; for Go/Python with non-`/healthz` default, warn-only (caller may have a valid override).

Pre-flight failures emit a single error block listing every failed check.

## Output format

```markdown
## add-new-service — <service-name> (<product>)

Pre-flight: PASS (port <N> free; listener priority <P> free; naming OK; tags complete)

### Artifact 1 — roots/compute/services_<service>.tf (NEW)
<HCL block>

### Artifact 2 — roots/compute/target_groups.tf (APPEND) + listener_rules.tf (APPEND)
<HCL blocks>

### Artifact 3 — roots/network/security_groups.tf (APPEND)
<HCL block>

### Artifact 4 — roots/compute/ecr_<service>.tf (NEW)
<HCL block>

### Artifact 5 — roots/compute/iam.tf (APPEND)
<HCL blocks: task role + GHA deploy role>

### Artifact 6 — <service-repo>/.github/workflows/deploy.yml (NEW)
<YAML block>

### Artifact 7 — <service-repo>/Dockerfile (NEW)
<Dockerfile block>

### Suggested commit plan
- Commit 1 (portfolio/infrastructure): Artifacts 1–5 — Terraform substrate for <service>
- Commit 2 (<service-repo>): Artifacts 6–7 — deploy workflow + Dockerfile

### Required follow-up
- Apply Commit 1: `cd portfolio/infrastructure/roots/network && terraform apply` then `roots/compute && terraform apply`.
- Push Commit 2 to `<repo-owner>/<repo-name>:main`; the new deploy.yml triggers automatically.
- Verify: `aws ecs describe-services --cluster ampy-shared-prod-ecs --services ampy-<product-prefix>-prod-ecs-<service>`.
```

The skill **does not** run `terraform plan`/`apply`, push to GitHub, or invoke AWS APIs. Output is markdown only; caller applies via standard plan + GHA OIDC flow.

## Reads

- `.claude/rules/cloud-iac-conventions-aws.md` (binding rule — all conventions)
- `.claude/rules/portfolio-conventions.md §Port Allocation` + writable-paths table
- `.claude/skills/add-new-service/templates/` (the 7-artifact template set)
- `portfolio/system-refs/overview/INFRA.md` (when WI-14.DOC-4 lands — substrate at-a-glance)
- `portfolio/infrastructure/modules/{ecs-fargate-service,ecr-repo,alb}/variables.tf` (module input contracts)

## Test fixture

`.claude/skills/add-new-service/test/foo-service.expected.md` — golden output for the synthetic service:
- `service-name = foo`, `product = oppor`, `language = go`, `port = 8084` (free in oppor band 8080–8089), `feature = FX.PE`, `wi-id = WI-14.SKILL-1`, `repo-owner = ai-mpathyminds`, `repo-name = foo`, defaults for the rest.

**Acceptance** (per WI-14.SKILL-1 NFR):
- `terraform validate` on Artifacts 1–5 HCL: exit 0 (offline syntax-only).
- `docker buildx build --check --platform linux/arm64 services/foo` (BuildKit syntax check): exit 0.
- Idempotent re-run emits exactly **7 WARN lines** + zero diffs; AUDIT verb = `add-new-service-noop`.
- `description:` value in this file's frontmatter ≤200 bytes (per `.claude/rules/token-budget.md` rule #3).

## What this skill is NOT

- NOT a service-code generator beyond Dockerfile scaffold — no `cmd/server/main.go`, no `app/__init__.py`, no React `App.tsx`. Language-level scaffolding is the `{lang}-developer` lane.
- NOT a multi-service batch creator — one service per invocation; chain manually for multi-service onboarding (e.g. agrisetu PI15).
- NOT a Terraform state-modifying tool — emits commit-plan markdown only.
- NOT a service decommission tool — removal is a separate one-time runbook entry.
- NOT a port-allocation editor — when the product's band needs expansion, the caller updates `portfolio-conventions.md` manually first; the skill reads, never writes.

## Companion skills

- `modify-service-tf` — correlated-edit guard for live services (the edit complement to this create-only skill).
- `terraform-conventions-linter` — validates emitted HCL against `cloud-iac-conventions-aws.md` rules pre-commit.
- `secret-scanner` — final pass before commit; catches accidental inlined credentials.
- `compose-linter` — for the optional local-dev `docker-compose.yml` block (not auto-emitted by this skill).
