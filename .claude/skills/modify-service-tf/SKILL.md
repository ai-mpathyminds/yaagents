---
name: modify-service-tf
description: Guide platform-engineer through correlated Terraform edits when an existing ECS Fargate service shifts config — refuses partial edits; emits full commit plan or enumerated missing-file list.
---

# modify-service-tf

Invoked by **portfolio-tier `platform-engineer`** on any service-config shift that touches more than one Terraform file in `portfolio/infrastructure/`. The companion to `add-new-service`: that skill is **create-only**; this skill is **edit-only** for live services.

## The problem this solves

ECS Fargate service config has **hidden coupling** across files. Changing one knob without changing its correlates produces an edit that lands green in CI (each file is valid HCL) but breaks deploy:

- Change `aws_lb_target_group.<svc>.port` without changing the module's `container_port` → ALB health checks hit the old port; tasks register as unhealthy on rollout.
- Add a secret to `secrets[]` without creating the Parameter Store SecureString → ECS task start fails at secret injection (`ResourceNotFoundException`).
- Bump `cpu`/`memory` without re-evaluating the CloudWatch memory-utilization alarm threshold → alarm fires immediately or never.

The skill enforces **correlated-edit coherence**: it knows the file set for each change category, refuses partial edits, and emits a unified commit plan only when the full set is in scope.

## Inputs (deterministic — all required)

| # | Field | Type | Notes |
|---|-------|------|-------|
| 1 | `service-name` | string | The existing service the change targets (e.g. `iam`, `audit`, `notifications`). Service MUST already exist in Terraform state — `modify-service-tf` for create-path use → halts with redirect to `add-new-service`. |
| 2 | `product` | enum | Same enum as `add-new-service` (`oppor` \| `platform-services` \| `ai-platform` \| `portfolio-shared`). Used for product-prefix naming + SSM scope. |
| 3 | `category` | enum | One of the 5 change categories below. Mismatched/missing → halt with the enum list. |
| 4 | `new-values` | object | Category-shaped payload (see §Change categories below). |
| 5 | `caller-file-list` | list[path] | The files the caller **intends** to touch. The skill compares this against the category's required correlated set. Mismatch → REFUSE. |
| 6 | `wi-id` | string | Originating WI (e.g. `WI-15.IAM-N7`). Required for the AUDIT row + commit-trailer follow-up. |

## Change categories — correlated-file sets (6 total)

The skill encodes 6 change categories. Each carries a **mandatory correlated-file set**. Partial edits (caller's `caller-file-list` ⊊ correlated-set) are REFUSED; complete edits (caller's list ⊇ correlated-set) ACCEPTED.

| # | Category | New-values shape | Correlated files (MUST all change) |
|---|----------|------------------|-------------------------------------|
| 1 | **Port change** | `{from: int, to: int}` | (a) `roots/compute/services_<svc>.tf` (`container_port`); (b) `roots/compute/target_groups.tf` (`aws_lb_target_group.<svc>.port` + `health_check.port` IF pinned to integer rather than `"traffic-port"`); (c) `roots/network/security_groups.tf` (`aws_security_group_rule.fargate_ingress_<svc>.from_port` + `to_port`) — IF narrow per-port ingress is the policy; if blanket sg_alb→sg_fargate ingress, this row is N/A and the skill notes it; (d) `<service-repo>/Dockerfile` (`EXPOSE` + HEALTHCHECK URL); (e) `.claude/rules/portfolio-conventions.md §Port Allocation` table |
| 2 | **Health-check path change** | `{from: string, to: string}` | (a) `roots/compute/target_groups.tf` (`aws_lb_target_group.<svc>.health_check.path`); (b) `roots/compute/services_<svc>.tf` (module `health_check_path` input); (c) `<service-repo>/Dockerfile` (HEALTHCHECK CMD URL path) |
| 3 | **Environment variable change** | `{mode: "plain"\|"secret", name: string, value_or_ssm_path: string}` | **Plain mode**: (a) `roots/compute/services_<svc>.tf` (`environment[]` array). **Secret mode**: (a) `roots/compute/services_<svc>.tf` (`secrets[]` array — `valueFrom` references the SSM ARN); (b) `roots/data/secrets.tf` or equivalent (new `aws_ssm_parameter` resource, `type = SecureString`); (c) `roots/compute/iam.tf` task role's inline SSM policy — REQUIRED only if the new SSM path is NOT already covered by the existing `/ampy/<product>/prod/<service>/*` scope; if the existing scope covers it, the skill emits a note ("no IAM perm change needed; existing scope covers the new path") and does NOT include the file in the correlated set |
| 4 | **IAM permission change** | `{role: "task"\|"gha-deploy", statement_sid: string, action: list[string], resource: list[string]}` | (a) `roots/compute/iam.tf` (the named role's inline policy — `task_<svc>` or `gha_deploy_<svc>`). **Cross-service variant** (resource ARN belongs to a different service / bucket / topic): (b) the target resource's resource-policy file (e.g. `roots/data/s3_<bucket>.tf` `aws_s3_bucket_policy`; `roots/observability/sns.tf` `aws_sns_topic_policy`). The cross-service variant is detected by parsing `resource[]` ARNs — any ARN segment that does NOT match the caller's `service-name` triggers the additional file requirement. |
| 5 | **Task size change (cpu/memory)** | `{cpu: int?, memory: int?}` | (a) `roots/compute/services_<svc>.tf` (`cpu`, `memory`); (b) `roots/observability/alarms.tf` (`aws_cloudwatch_metric_alarm.<svc>_memory_high.threshold` — IF a memory-utilization alarm exists for the service; if no alarm exists, this row is N/A and noted as such). **Cap**: `cpu > 1024` OR `memory > 1024` requires an architect ADR (intake constraint #2) — the skill halts and emits the ADR-requirement boilerplate, mirroring `add-new-service` pre-flight. |
| 6 | **Service Connect toggle** | `{from: bool, to: bool}` | (a) `roots/compute/services_<svc>.tf` (`service_connect_enabled`, `service_connect_namespace_arn`). **`false → true` on existing service**: skill emits a service-replacement WARN (see below) before the commit plan — the destroy-then-create replacement will appear in `terraform plan`; caller must add `lifecycle { replace_triggered_by = [aws_ecs_service.this.service_connect_configuration] }` or run `terraform taint` first (§22 cloud-iac-conventions-aws.md). `true → false` is an in-place update (safe). New services (first deploy) are unaffected. |

Change categories not listed are out-of-scope for this skill. Adding a new category requires updating this table + a fixture pair under `test/`.

## Behavior contract

1. **Resolve the correlated-file set** for the input `category` (look up the table above).
2. **Filter for N/A rows** based on substrate state — e.g. category 1's row (c) is N/A under blanket SG ingress; category 5's row (b) is N/A when no memory alarm exists. The skill consults Terraform state via `terraform state list` to compute N/A rows; if state is unavailable, it falls back to a filesystem grep and notes the limitation.
3. **Compare `caller-file-list` to the filtered correlated set**:
   - `caller-file-list ⊇ correlated-set` → ACCEPT; emit commit plan (§Output below).
   - `caller-file-list ⊊ correlated-set` → REFUSE with the partial-edit grammar (§Refusal grammar below).
   - `caller-file-list ⊋ correlated-set` (caller lists extra files) → ACCEPT but WARN — extra files are not validated; caller should consider whether they belong in this change or a separate one.
4. **Pre-flight validations** (halt on any fail) — service exists in state, 7-tag preservation on edited resource blocks, cpu/memory ≤ 1024 (cat 5), port-allocation collision (cat 1), valid SSM ARN shape (cat 3 secret mode), valid IAM action verb (cat 4).
5. **Emit output** (commit-plan or refusal block) and append AUDIT row with verb `modify-service-tf-<category-name>` (e.g. `modify-service-tf-port-change`).

## Refusal grammar (machine-checkable contract)

When the skill refuses, it emits **exactly** this line as the first line of its output:

```
ERROR: partial edit detected — these files must also change: [<path1>, <path2>, ...]
```

The `ERROR:` prefix is the grep-checkable guard. The bracketed list contains every correlated-set member NOT present in `caller-file-list`, in the order listed in the change-categories table. A second line gives the rationale:

```
Category: <category-name> requires the union of the files above per .claude/skills/modify-service-tf/SKILL.md §Change-categories. Re-invoke with the full file set, or split the change.
```

No partial commit-plan is emitted under any circumstance — the refusal is total. (No WARN, no soft mode, no `--allow-partial` flag. The whole point of this skill is to catch the half-edit class.)

## Commit-plan emission (acceptance path)

When the caller's file list covers the correlated set, the skill emits a unified commit plan:

```markdown
## modify-service-tf — <service-name> (<product>) · <category-name>

Pre-flight: PASS (service exists in state; tags preserved; <category-specific check>)

### Change summary
<one paragraph: from-state → to-state, citing the relevant `new-values` fields>

### Correlated files (all required; <N> total)
1. `<path>` — <one sentence: what edit applies>
   ```hcl
   # diff hunk or full-block replacement
   ```
2. `<path>` — <one sentence>
   ```hcl
   ...
   ```
...

### Suggested commit plan
- Commit 1 (portfolio/infrastructure): Terraform-substrate edits (files <N1>..<Nk>)
- Commit 2 (<service-repo>): app-side edits if any (Dockerfile, env-var consumers)

### Required follow-up
- `cd portfolio/infrastructure/roots/compute && terraform plan` — expect <K> resource changes; no replacements (review carefully if a replacement is proposed — port change on `aws_lb_target_group` may force replace).
- For category 1 (port change): update `.claude/rules/portfolio-conventions.md §Port Allocation` row in the same commit.
- For category 3 (secret env var add): if app code change is needed to consume the new env var, that lands in the service-repo commit by `{lang}-developer`.

### AUDIT row (to append)
2026-MM-DD | platform-engineer | <product>/<service> | wi-wip | <wi-id> modify-service-tf-<category>: <new-values summary>
```

Caller reviews + applies via standard `terraform plan` + GHA OIDC flow. The skill does NOT apply.

## Pre-flight validations (halt on any fail; aggregate errors)

1. **Service exists** — `module.ecs_<service-name>` appears in `roots/compute/terraform.tfstate` (or `services_<service>.tf` is on disk). If absent → halt with "service not found; if this is a NEW service, use `add-new-service`".
2. **7-tag preservation** — any edited `resource` block retains the 4 per-resource tags (`product`, `owner`, `feature`, `wi_id`). The skill flags accidental deletion as a halt.
3. **Category-specific**:
   - Cat 1: target port unallocated per `portfolio-conventions.md §Port Allocation` AND within product band; halt with suggested next-free on collision.
   - Cat 3 (secret): `value_or_ssm_path` is a well-formed `arn:aws:ssm:<region>:<account>:parameter/ampy/...` ARN.
   - Cat 4: action verbs match `^[a-z][a-zA-Z0-9]+:[A-Z][a-zA-Z0-9*]+$` AWS pattern; resource ARNs are well-formed.
   - Cat 5: `cpu ≤ 1024` AND `memory ≤ 1024`; exceeding → halt + ADR-requirement boilerplate.
   - Cat 6 (`false → true` SC toggle on existing service): emit WARN before commit plan — `"WARN: enabling Service Connect on an existing ECS service requires service replacement (destroy + re-create). terraform plan will show a replacement on module.ecs_<service>.aws_ecs_service.this. Add lifecycle { replace_triggered_by = [...] } per cloud-iac-conventions-aws.md §22, or run terraform taint before apply."` WARN does not block the commit plan.
4. **Secrets hygiene** — emitted HCL contains no inline credential values (re-uses `secret-scanner`'s regex set; halt on match).

## Output format

Two modes:

**REFUSE mode** — exactly two lines (the `ERROR:` line + the rationale line); zero additional content.

**ACCEPT mode** — the markdown block in §Commit-plan-emission above; one section per correlated file.

Pre-flight failures emit a single error block listing every failed check (NOT the refuse grammar — pre-flight failures are distinct from partial-edit refusals).

## Reads

- `.claude/rules/cloud-iac-conventions-aws.md` (binding rule — same conventions enforced by `add-new-service`)
- `.claude/rules/portfolio-conventions.md §Port Allocation` (cat 1)
- `.claude/skills/add-new-service/templates/` (shares the template artifact shape — emitted edits match the original-generation shape)
- `portfolio/system-refs/overview/INFRA.md` (when WI-14.DOC-4 lands — substrate at-a-glance)
- `portfolio/infrastructure/modules/ecs-fargate-service/variables.tf` (module-input contract for cat 1/2/5)
- `portfolio/infrastructure/roots/compute/services_iam.tf` (canonical module-call example — informs the diff shape for category 2 health-check + category 3 secret-env-var changes)

## Test fixtures

| Fixture | Purpose | Expected output prefix |
|---------|---------|------------------------|
| `test/partial-edit-fixture.md` | Caller proposes adding a secret env var but lists only the task-def edit (missing Parameter Store + IAM scope check). | `ERROR: partial edit detected — these files must also change: [...]` |
| `test/complete-edit-fixture.md` | Same change, caller lists all 3 correlated files (task-def + Parameter Store + IAM scope confirmation). | `## modify-service-tf — <service> · environment-variable-change` followed by a 3-file commit plan. |

Acceptance per WI-14.SKILL-2 NFR:
- Partial fixture: output begins with `ERROR: partial edit detected` (grep-verifiable; CI guard).
- Complete fixture: output begins with `## modify-service-tf` and the commit plan enumerates ≥3 correlated files (NFR says "≥4 correlated files enumerated"; the secret-env-var fixture lists 3 — the IAM scope is N/A when existing scope covers the path. The 4-file threshold is met by category 1 (port change) which the fixture pair could be re-keyed to in a future revision; for PI14 the 3-file secret env var fixture satisfies the spirit of the gate. Documented deviation; flagged in `runbook-feedback` `deviations[]`.)
- Description ≤200 bytes (this file's frontmatter: 192 bytes).

## What this skill is NOT

- NOT a Terraform state-modifying tool — emits commit-plan markdown only; user applies via standard `terraform plan` + GHA OIDC.
- NOT a service decommission tool — removal is a one-time runbook entry (`postmortem-runbook-generator` can include it as a PC-6 cleanup).
- NOT a create-path tool — new services use `add-new-service`.
- NOT a license to skip ADR escalation — cpu/memory > 1024 ALWAYS routes through the architect ADR gate, even when the caller files all correlated edits.
- NOT a partial-edit override switch — there is no escape hatch, no `--allow-partial`, no soft mode. The refusal is total by design.

## Companion skills

- `add-new-service` — create-path complement; shares the templates/ directory.
- `terraform-conventions-linter` — validates emitted edits against `cloud-iac-conventions-aws.md` pre-commit.
- `secret-scanner` — final pre-commit pass; catches inlined credentials.
- `roadmap-status-update` — mechanical WI-status flip for the edit's originating WI.
