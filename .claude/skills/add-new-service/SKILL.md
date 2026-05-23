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
| 13 | `health-check-path` | string (optional) | **ALB target-group** health-check path. Default `/healthz` (Go/Python). React default `/`. Keycloak-like services override |
| 14 | `health-check-command` | string\|null (optional) | **Container-level** `HEALTHCHECK`/ECS `healthCheck` command. Default `null` — distroless runtime images have no shell, so a `CMD-SHELL wget` probe is impossible; the ALB target-group HTTP check (#13) is the health source. Set a value ONLY for shell-bearing runtimes (alpine) that genuinely need a container probe. See §Distroless-aware health checks (G6). |
| 15 | `aws-deps` | list (optional) | Service-specific AWS dependencies beyond the Fargate baseline — any of `ses` \| `sqs` \| `sns` \| `s3`. Each emits its resource(s) + the task-role IAM policy (+ Cloudflare DNS / sandbox-exit notes for SES). Default `[]`. See §Service AWS dependencies (G1). |

> **ECS-exec knob removed (G9 / B-58 ruling).** The former `enable-execute-command` input is gone. ECS-exec (`enable-execute-command = true`) was rejected at B-58 for realm/admin bootstrap after 4 failed builds — use the service's admin REST API (e.g. Keycloak admin REST) for bootstrap, never ECS-exec. The skill no longer offers the knob; a service that thinks it needs it should escalate to chief-architect.

Inputs that fail pre-flight halt with an error block listing **all** failed checks (not just the first); caller fixes all at once. Partial commit plans are never emitted.

## Generated artifacts — the 7-artifact set

Each numbered artifact below corresponds to one template under `.claude/skills/add-new-service/templates/`. The skill composes templates with input substitution and emits a markdown commit plan with one section per artifact.

### Artifact 1 — `roots/compute/services_<service>.tf` (NEW; Terraform — ecs-fargate-service module call)
Source: `templates/01_ecs_service.tf.tmpl`. Module path `../../modules/ecs-fargate-service`. Inputs filled per the module's `variables.tf` contract: `service_name`, `product`, `cluster_name = aws_ecs_cluster.shared.name`, `container_image = "<account-id>.dkr.ecr.<region>.amazonaws.com/ampy-<product-prefix>-prod-ecr-<service>:${var.<service>_image_tag}"`, `container_port`, `cpu`/`memory`, `task_role_arn`, `execution_role_arn = aws_iam_role.fargate_execution.arn`, `target_group_arn`, `private_subnet_ids` + `security_group_ids` from network remote state, `service_connect_enabled = true` (Service Connect alias = service name), `health_check_path`, `health_check_command` (input #14 — passed through; `null` for distroless), and the 4 caller-supplied tags (`owner`, `feature`, `wi_id`, plus `product` already passed).

**Image tag is a TF variable — never a literal (G3 / PC-5-05 / §20a).** The `:bootstrap` / `:latest` literal placeholder is an anti-pattern: a sibling `terraform apply` reverts whatever GHA last pushed because TF does not own the tag (B-59 evidence: notifications kept reverting to `:bootstrap`). The skill therefore ALSO emits, into `roots/compute/variables.tf` (APPEND), a `variable "<service>_image_tag" {}` block with **no default** (forcing the value to come from CI). The GHA deploy workflow (Artifact 6) passes the freshly-pushed deterministic tag via `-var "<service>_image_tag=<sha-timestamp>"` so TF state always matches the live image. `terraform-conventions-linter` check #13 fails any literal image tag in a compute root.

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
Source: `templates/06_deploy_workflow.yml.tmpl`. Triggers: `push` on `main` with paths filter scoped to the service directory. Permissions: `id-token: write`, `contents: read` (OIDC requirement). Steps: checkout → compute a **deterministic tag** `IMAGE_TAG=<short-sha>-$(date -u +%Y%m%dT%H%MZ)` (§20b — never a reused RC suffix) → `aws-actions/configure-aws-credentials@v4` (assume GHA deploy role via OIDC) → ECR login → **Assert tag not already in ECR** (`aws ecr describe-images … imageTag=$IMAGE_TAG`; if it exists, `exit 1` — §20c fail-loud, NO silent "skipping push") → docker buildx (arm64) → push `$IMAGE_TAG` → set the TF variable so state matches live (`terraform apply -var "<service>_image_tag=$IMAGE_TAG"` OR `aws ecs update-service --force-new-deployment` when TF apply is decoupled) → `aws ecs wait services-stable`. **No `${{ secrets.AWS_ACCESS_KEY_ID }}`** — only `AWS_ACCOUNT_ID` (non-secret) needed; OIDC role-to-assume is the auth surface. The template carries the §20c collision-assert step verbatim; `terraform-conventions-linter` check #13 greps for the prohibited `"skipping push"` string.

### Artifact 7 — `<service-repo>/Dockerfile` (NEW; multi-stage arm64)
Source: language-keyed template (`templates/07_dockerfile_go.tmpl` | `_python.tmpl` | `_react.tmpl`).
- **Go**: `golang:1.22-alpine` build → `alpine:3.19` runtime, `CGO_ENABLED=0 GOARCH=arm64`, non-root user, `HEALTHCHECK` via `wget` against `health-check-path`, `ENTRYPOINT ["/app/server"]`. **Distroless variant** (smaller, no CVE-bearing shell): `gcr.io/distroless/static-debian12:nonroot` runtime — has **no shell**, so the Dockerfile emits **NO `HEALTHCHECK`** and the caller MUST pass `health-check-command = null` (input #14); the ALB target-group HTTP check (#13) is the sole health source. See §Distroless-aware health checks.
- **Python**: `python:3.12-slim-bookworm` build (uv-installed wheels) → same image runtime, non-root, `curl` HEALTHCHECK, `ENTRYPOINT ["python", "-m", "app"]`. Distroless (`gcr.io/distroless/python3`) follows the same no-shell / `health-check-command = null` rule.
- **React**: `node:20-alpine` build (pnpm default; detect lockfile to swap) → `nginx:alpine` runtime, non-root, read-only fs prep, `nginx.conf` bundled, HEALTHCHECK via `wget`.

## Service AWS dependencies (add-on dimension) (G1)

The 7-artifact set above is the **Fargate baseline**. A service that needs email, queueing, pub/sub, or object storage has AWS dependencies the baseline does not cover — PI15 evidence: `notifications` needed an SES domain identity + DKIM + email identity + sandbox-exit + Cloudflare DNS and NONE of it was in this skill, so B-59 became an unplanned escalation (ADR-0005). When input #15 `aws-deps` is non-empty, the skill emits **one add-on artifact per dependency** plus the matching task-role IAM statement (appended to Artifact 5) and any DNS / ALB follow-ups. ALB rules are emitted **only when the dependency is itself an HTTP ingress** — `ses`/`sqs`/`sns`/`s3` are not, so they emit no listener rule (the "+ ALB rules when applicable" applies to future ingress-bearing add-ons).

| `aws-deps` value | Resources emitted (file) | Task-role IAM (appended to Artifact 5) | DNS / manual follow-ups |
|---|---|---|---|
| `ses` | `aws_ses_domain_identity` + `aws_ses_domain_dkim` + `aws_ses_domain_identity_verification` (+ optional `aws_ses_email_identity` for a pinned From-address) → `roots/data/ses_<service>.tf`. **All SES identity resources are UNTAGGABLE** (no Tags API in provider ~>5.50 — see §Untaggable AWS resources; do NOT emit a `tags` block or `terraform apply` fails, B-59 commit `894ae9f`). | `ses:SendEmail`, `ses:SendRawEmail` scoped to the domain-identity ARN | Emit **Cloudflare** DNS records (3 DKIM CNAMEs + 1 verification TXT) as a follow-up note — DNS is in Cloudflare, not Route53. **Sandbox exit** (move out of SES sandbox to send to arbitrary recipients) is a manual AWS Support request — flag it; it is NOT Terraform. |
| `sqs` | `aws_sqs_queue` (+ optional DLQ + `redrive_policy`) → `roots/data/sqs_<service>.tf`. Taggable (4 per-resource tags). | `sqs:SendMessage`, `sqs:ReceiveMessage`, `sqs:DeleteMessage`, `sqs:GetQueueAttributes` scoped to the queue ARN | — |
| `sns` | `aws_sns_topic` (+ optional `aws_sns_topic_policy` for cross-service publish — note: linter treats the policy **data source** as an advisory WARN, §16) → `roots/observability/sns_<service>.tf`. Topic taggable. | `sns:Publish` scoped to the topic ARN | — |
| `s3` | `aws_s3_bucket` (name needs the 4-char random suffix per §9) + `aws_s3_bucket_public_access_block` + versioning + SSE → `roots/data/s3_<service>.tf`. Bucket taggable. | `s3:GetObject`, `s3:PutObject`, `s3:ListBucket` scoped to the bucket + `/*` object ARNs | — |

Each add-on is idempotency-checked exactly like the baseline artifacts (state probe `aws_<dep>_<...>` / file on disk → one WARN per pre-existing add-on). Adding an AWS dep to an **existing** service is the `modify-service-tf` path (it cross-references this table).

## Distroless-aware health checks (G6)

The `ecs-fargate-service` module historically hardcoded a container `healthCheck` of `CMD-SHELL … wget …`. **Distroless runtime images have no shell**, so that probe can never pass — PI15 evidence: the real 7 MB distroless `notifications` image failed its container health check and needed a manual task-def revision (M5). Rule:

- **Container-level health check is OPTIONAL.** Input #14 `health-check-command` defaults to **`null`**. When `null`, Artifact 1 sets the module's container `healthCheck` to null and the Dockerfile (Artifact 7) emits **no `HEALTHCHECK`** instruction.
- **The ALB target-group HTTP check (input #13 `health-check-path`) is the health source** for distroless services — it probes from outside the container and needs no shell.
- Set a non-null `health-check-command` ONLY for shell-bearing runtimes (alpine) that genuinely need a container-internal probe (e.g. a non-HTTP worker with a local liveness script).
- The module README `§Preconditions` documents the no-shell constraint; the skill cites it in the commit plan whenever `health-check-command = null`.

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

Resource blocks emit only the 4 per-resource tags (`product`, `owner`, `feature`, `wi_id`); the other 3 inherit via provider `default_tags` in the root config. The skill does NOT duplicate them per resource — this matches `terraform-conventions-linter` check #4's expected union.

### Untaggable AWS resources (G2 — enumerated inline; do NOT emit a `tags` block on these)

Some AWS resource types have **no Tags API** in provider `~>5.50`. Emitting a `tags`/`tags_all` block on them makes `terraform apply` **fail** (not warn) — PI15 evidence: `aws_ses_domain_identity` broke the notifications apply; hand-revert + EXEMPTIONS rows in commit `894ae9f`. The skill emits these resources **without any tag block**, and `terraform-conventions-linter` check #4 skips them. Enumerated (relevant to this skill's emissions):

| Untaggable resource | Tag instead on… |
|---|---|
| `aws_ses_domain_identity`, `aws_ses_domain_dkim`, `aws_ses_domain_identity_verification`, `aws_ses_email_identity` | (no parent — SES identities are simply untaggable) |
| `aws_security_group_rule` (the standalone rule resource) | the parent `aws_security_group` |
| `aws_iam_role_policy` (inline policy) | the parent `aws_iam_role` |
| `aws_route53_record` | the parent `aws_route53_zone` |
| `aws_ecs_cluster_capacity_providers` | the parent `aws_ecs_cluster` |
| `aws_lb_listener_rule` condition/action sub-blocks | the `aws_lb_listener_rule` itself IS taggable — tag it |

This inline list covers what `add-new-service` + `modify-service-tf` actually emit. The **authoritative full list** remains `EXEMPTIONS.md` next to the linter; when a new untaggable type is hit, add the row there AND here.

## Idempotency contract — one WARN per logical artifact

Before emitting each of the 7 artifacts, the skill checks Terraform state (`terraform state list` in the affected root) + filesystem (the affected file paths). **TF state is REQUIRED — hard STOP if unavailable (G8).** If `terraform state list` fails for any reason (no applied state, lock held, expired SSO requiring relogin), the skill emits the stop block below and authors NOTHING — it does **NOT** silently degrade to a filesystem-only grep. A filesystem-only run cannot see resources that exist in AWS but not on disk, so it would emit a "create" plan that **duplicates live resources** — exactly the failure we cannot afford. Restore state access and re-invoke.

```
STOP: Terraform state unavailable for root <root> (<reason>). add-new-service requires live state to verify idempotency; refusing to author a plan that may duplicate existing resources. Restore state access (terraform init / re-auth SSO / release lock) and re-invoke.
```

If **any** sub-resource of an artifact already exists, that entire artifact is skipped and emits exactly one WARN line:

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
9. **TF state reachable (G8)** — `terraform state list` succeeds in `roots/compute` AND `roots/network`. Failure (no state / SSO relogin / lock) → **hard STOP** (see §Idempotency contract); idempotency is unverifiable without state and a blind create-plan risks duplicating live resources.

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
- Apply Commit 1: `cd portfolio/infrastructure/roots/network && terraform apply` then `roots/compute && terraform apply -var "<service>_image_tag=<tag>"`.
- Push Commit 2 to `<repo-owner>/<repo-name>:main`; the new deploy.yml triggers automatically.
- **Populate SSM params** the task role reads (`/ampy/<product>/prod/<service>/*`) — **always strip CRLF (G7)**: `printf %s "$VALUE" | tr -d '\r\n' | aws ssm put-parameter --name /ampy/<product>/prod/<service>/<key> --type SecureString --value "$(cat -)"` (or pipe the value through `tr -d '\r\n'` before any put). A trailing CRLF from a Windows dev host silently corrupts the value — PI15 M6: a CRLF in the pg-url SSM value broke Go `net/url` parsing. Same on reads if you re-`put` a fetched value.
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
- NOT an ECS-exec enabler — the `enable-execute-command` knob was removed (G9 / B-58 ruling). Admin / realm bootstrap uses the service's admin REST API (e.g. Keycloak admin REST), never ECS-exec.

## Companion skills

- `modify-service-tf` — correlated-edit guard for live services (the edit complement to this create-only skill).
- `terraform-conventions-linter` — validates emitted HCL against `cloud-iac-conventions-aws.md` rules pre-commit.
- `secret-scanner` — final pass before commit; catches accidental inlined credentials.
- `compose-linter` — for the optional local-dev `docker-compose.yml` block (not auto-emitted by this skill).
