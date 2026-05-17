---
name: terraform-conventions-linter
description: Validate Terraform IaC under portfolio/infrastructure/ for AWS conventions per .claude/rules/cloud-iac-conventions-aws.md — provider pins, region, 7-tag set, naming, no inline secrets.
---

# terraform-conventions-linter

Static lint of every Terraform file under `portfolio/infrastructure/`. Authority: `.claude/rules/cloud-iac-conventions-aws.md` — the linter encodes that rule's checks. When the rule changes, the linter checks change in the same PR.

## Files scanned

- `portfolio/infrastructure/**/*.tf`, `**/*.tf.json`
- `portfolio/infrastructure/**/*.tfvars` (read-only — backend / region values)

Excluded: `.terraform/`, `*.tfstate*`, anything matching `.gitignore`.

## Checks (each PASS or FAIL — no warnings)

### 1. Terraform version pin
`terraform { required_version = ">= 1.7.0" }` (or stricter) in every root and module.

### 2. AWS provider version pin
`required_providers.aws.version = "~> 5.50"` (matches the rule's current pin). Floating constraints (`>= 5.0` without `~>`, missing version) FAIL.

### 3. Region pin
- Roots: `provider "aws" { region = "ap-south-1" }`.
- Variables: `region` default = `"ap-south-1"`.
- No hard-coded non-`ap-south-1` region literal anywhere; cross-region provider aliases require `# tf-linter: region-alias-justified — ADR-<id>`.

### 4. Required tag set on every taggable resource
The 7 tags from `cloud-iac-conventions-aws.md` §10 (`product`, `environment`, `owner`, `cost_center`, `managed_by`, `feature`, `wi_id`). Linter unions resource `tags` + `tags_all` + provider `default_tags`. Resource types AWS does not allow tagging on are exempt — list at `EXEMPTIONS.md` next to this skill (kept current with provider releases). Ad-hoc skip: `# tf-linter: skip-tags — <reason>` above the resource block.

### 5. Naming convention
`name`/`name_prefix` matches `ampy-{product}-{env}-{resource}[-{suffix}]` (rule §9). `product` ∈ `oppor|plt|aip|shared`; `env` ∈ `prod|stg`; `resource` from approved kind-code set. Mismatch FAILS unless `# tf-linter: skip-name — <reason>`.

### 6. Module shape
Every `portfolio/infrastructure/modules/<name>/` has `main.tf`, `variables.tf`, `outputs.tf`, `versions.tf`, `README.md`. Every `variable` has `type` + `description`; every `output` has `description`. Missing → FAIL.

### 7. State-backend declaration
Roots declare an `s3` backend with `bucket = "ampy-tf-state-prod-ap-south-1"`, `dynamodb_table = "ampy-tf-locks-prod"`, `region = "ap-south-1"`, `key = "<root>/terraform.tfstate"`, `encrypt = true`. `bootstrap/` exempt (creates the backend).

### 8. No inline secrets
Reuses `secret-scanner`'s regex set (AWS access keys `AKIA[0-9A-Z]{16}`, JWT bearer secrets, hex keys ≥32 chars, Postgres connection strings with passwords). Matches in tracked `*.tf`/`*.tfvars` FAIL.

### 9. No provider blocks inside modules
Files under `modules/**` containing `provider "aws" { ... }` (or any other instance block) FAIL. `required_providers` declarations in `versions.tf` are PASS — those are constraints, not instances.

### 10. NAT Gateway cost gate
`resource "aws_nat_gateway"` FAILS unless a sibling `# tf-linter: nat-gw-justified — ADR-<id>` is present. Default pattern is the `nat-instance` module (~$4/mo vs ~$32/mo NAT GW).

### 11. CloudWatch log group retention ≤ 7 days (NFR-FIN-2)
Any `resource "aws_cloudwatch_log_group"` block where `retention_in_days` is absent OR set to a value > 7 FAILS.
Rationale: intake constraint #2 (cost-first); CW Logs charges ~$0.03/GB/month; unretained logs accumulate indefinitely.
PASS examples: `retention_in_days = 7`, `retention_in_days = 1`.
FAIL examples: `retention_in_days = 14`, `retention_in_days = 30`, block with no `retention_in_days` attribute.
Skip marker (requires governance-auditor approval, must cite ADR): `# tf-linter: skip-retention — ADR-<id>`.
Negative test fixture: `.claude/skills/terraform-conventions-linter/test_fixtures/log_retention_14d.tf`.

### 12. SNS topic policy — no wildcard Action (PI13 PC-5-07; apply-time→lint-time up-shift)
Any `resource "aws_sns_topic_policy"` block whose `policy` JSON contains a statement with `Action: "SNS:*"` (string) OR `Action: [...]` containing the literal `"SNS:*"` element FAILS. Actions MUST be specific (e.g. `SNS:Publish`, `SNS:Subscribe`, `SNS:GetTopicAttributes`).
Rationale: AWS rejects account-level SNS wildcards in topic policies — account-level SNS actions (`CreateTopic`, `ListTopics`, `SetTopicAttributes`) are not topic-scoped, so the wildcard expansion crosses scope boundaries and `terraform apply` returns `Invalid parameter: Policy Error: PrincipalNotFound` or `action out of service scope`. PI13 WI-13.OBS-1 evidence: apply-time rejection 2026-05-11; remediated in commit `882217c` (drop `AllowAccountOwner` statement; reduce to specific `AllowBudgetsPublish` with `SNS:Publish` only); feedback.ndjson row records `"SNS:* in topic policy rejected by AWS"`. This check moves the catch up-shift to pre-commit so it never reaches apply again.
Detection: scan the body of each `aws_sns_topic_policy` block for `"SNS:*"` substring (covers both `jsonencode({...})` inline form and `<<-EOF ... EOF` heredoc form). The `data.aws_iam_policy_document` indirect form is a known blind spot — flag at advisory level (WARN, not FAIL) when an `aws_sns_topic_policy.policy` references a `data.aws_iam_policy_document.*.json` attribute; reviewer must inspect the data-source body for the same wildcard pattern.
PASS examples: `"Action": "SNS:Publish"`, `"Action": ["SNS:Publish", "SNS:Subscribe"]`, `"Action": "SNS:GetTopicAttributes"`.
FAIL examples: `"Action": "SNS:*"`, `"Action": ["SNS:*"]`, `"Action": ["SNS:Publish", "SNS:*"]`.
Skip marker (requires governance-auditor approval + ADR citing why account-scope wildcard is needed despite AWS's documented rejection): `# tf-linter: skip-sns-wildcard — ADR-<id>`.
Negative test fixture: `.claude/skills/terraform-conventions-linter/test_fixtures/sns_topic_policy_wildcard.tf` (replicates the WI-13.OBS-1 pre-fix shape — the `AllowAccountOwner` statement with `Action: "SNS:*"` that AWS rejected).

## Output format

```markdown
## terraform-conventions-linter — <scope>

| File | Resource / Module | Check | Result | Note |
|------|-------------------|-------|--------|------|

**Summary**: <n> checks, <n> FAIL
```

Exit code 0 if all PASS, 1 if any FAIL. Used in pre-commit hooks and CI.

## Auto-fix policy

**None.** Reports only. Fixes are the `platform-engineer` lane.

## What this skill does NOT do

- Does not run `terraform plan` / `apply` — the OPA/Conftest gate (post-`plan`) catches dynamic / merge-resolved tag failures.
- Does not validate Rego policies — `conftest test` runs separately in CI.
- Does not check actual AWS state — static against `.tf` files only; drift detection is `terraform plan` against current state.
- Does not vet cost broadly — `infracost` runs separately; only the NAT-GW canary lives here.
- Does not enforce ADR text — only the `# tf-linter: <type>-justified — ADR-<id>` comment marker.

## Companion skills

- `compose-linter` — same shape, docker-compose targeted (the original; this skill mirrors its layout intentionally).
- `secret-scanner` — supplies the regex set referenced in check #8.
- `external-library-vetting` — vets new third-party Terraform modules before adoption.
