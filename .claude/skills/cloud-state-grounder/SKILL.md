---
name: cloud-state-grounder
description: "A-3 cloud-readiness alignment matrix at portfolio/REPORTS/cloud-state/: 10 live checks DNS/ALB/ECS/image/SG/SSM/KC/DB/in-VPC/SC vs TF+code. chief-architect A-1, *-architect A-3. Reports only."
---

# cloud-state-grounder

Reads the **live cloud** (TF state + AWS API + Keycloak admin + SSM) and the **design source** (TF source, service code, realm-export bundle, migrations) and emits the **A-3 cloud-readiness alignment matrix** — one row per cell of the cross-cutting state the design depends on. This is the **machine implementation of the [PROPOSED] PC-5-01 PROCESS delta** ("A-3 cloud-readiness alignment matrix mandate", `portfolio/PROCESS.md` 2026-05-21) and the executable answer to the blunder-report Meta-1 finding ("No A-3 cloud-readiness alignment matrix", `portfolio/REPORTS/2026-05-21-pi15-plt-aip-blunder-accountability.md §"Quick what would have caught all of this at A-3"`).

**Reports only. Never mutates cloud, TF, code, or roadmaps.** Cost to produce: ~4h of architect time at A-3 (or seconds when this skill runs it). Cost to NOT produce: PI15 (1.5wk slip + 2 vetoes + 15 defects).

## Who calls it

- **chief-architect at A-1** (seed authoring) — runs against the seed's `target_services` to ground the seed in live state before it flips `[DRAFT]→[READY]`. Surfaces what already exists vs. what the PI must build.
- **`{product}-architect` at A-3** (WI breakdown) — runs against the `[READY]` roadmap so every DRIFT/ABSENT cell becomes an explicit WI, not a Phase-B surprise.

Becomes a **REQUIRED** A-1/A-3 step for any PI whose `target_services` include cloud on PC-5-01 ratification (until then: advisory, strongly recommended).

## Inputs (deterministic)

| # | Input | How supplied | Notes |
|---|-------|--------------|-------|
| 1 | **Target PI** | arg, e.g. `PI16-plt-aip` | Names the output file; lane suffix preserved |
| 2 | **Target services** | arg list, e.g. `iam-api, ui-platform` | The surface to ground; usually the seed's `target_services` |
| 3 | TF source + state | `portfolio/infrastructure/roots/{network,data,compute,ci,observability}/**` + `terraform output` / remote state | **Primary source of truth** per `.claude/rules/cloud-iac-conventions-aws.md §21** |
| 4 | Service code | `{product}/services/{svc}/**` (`config.go`, `repository.go`, KC client defaults) | read-only |
| 5 | KC realm bundle | `realm-export.json` in the service/infra tree | compared against the live realm |
| 6 | Migrations | `{product}/services/{svc}/**/migrations/**` | compared against live DB tables |
| 7 | Live AWS / KC / DB | read-only CLI: `aws {ecs,ecr,elbv2,ssm,ec2}`, `nslookup`, `curl` KC admin, `psql -lqt` | region `ap-south-1` (`§2`); creds via read-only role |

If a target service has zero cloud footprint (pure-frontend, docs), record it as `N/A — no cloud surface` and move on. If AWS/KC/DB access is unavailable on this host (16GB dev box / no creds — see MEMORY: defer heavy live-infra to CI), the affected cells are `UNKNOWN-NOACCESS`, **never silently PASS** (Meta-2: absence is fail-loud, not skip). Emit `> note: live-access degraded — N cells UNKNOWN; re-run in CI to ground.`

## The 10 checks (one matrix row each, per blunder report)

| # | Layer | Item | Source of truth | How to verify | Owner |
|---|-------|------|-----------------|---------------|-------|
| 1 | DNS | every public hostname expected to exist resolves | TF (Route53/ALB) + corporate-site `endpoints.ts` + service config | `nslookup <host>` from public DNS | architect |
| 2 | Hostname → routing | each hostname has an ALB listener rule with priority **<** the catch-all | listener-rule TF + `aws elbv2 describe-rules` | cross-check rule priority vs default action | architect |
| 3 | TF → ECR image | every TF-managed ECS service uses `var.{service}_image_tag` (NOT a literal — §20a) matching a live image | grep TF (`roots/compute`) + `aws ecr describe-images` | tag exists in ECR AND equals running task image | platform-engineer |
| 4 | TF → live ECS | service names **and** cluster name match live | grep TF + `aws ecs list-services` | every TF service present in the live cluster | platform-engineer |
| 5 | SG ports | every `container_port` has SG ingress from the ALB SG (+ intra-SG rule if Service Connect) | grep TF (SG + task-def `container_port`) | each port has an ingress path | platform-engineer |
| 6 | SSM params | every secret/param in service `config.go` has a populated, **non-PLACEHOLDER** SSM entry | grep `config.go` + `aws ssm describe-parameters` (+ existence/sentinel check, never log the value) | param exists AND value ≠ `PLACEHOLDER`/empty | platform-engineer |
| 7 | KC realm | every client referenced in code defaults OR `realm-export.json` exists in the live realm | grep code + bundle + `curl /admin/realms/{realm}/clients` | each referenced `clientId` present in live realm | platform-services-architect |
| 8 | DB schema | every table referenced by any `repository.go` INSERT/SELECT exists in migrations (and live DB) | grep repo + migrations + `psql -lqt` / `\dt` | referenced table has a migration AND is live (F4's class) | architect |
| 9 | Cross-service | every service-to-service URL resolves from **inside the VPC** | SSM RunCommand or a probe from any Fargate task | intra-VPC DNS resolves + connects | platform-engineer |
| 10 | Service Connect | every TF-declared SC service shows `enabled=true` at runtime | grep TF + `aws ecs describe-services` | declared SC == enabled in describe-services | platform-engineer |

## Cell verdict vocabulary

- **LIVE-PASS** — design cell matches live cloud. Source of truth = TF (§21); cite `# tf-ref: roots/<root>/<file>.tf:<line>` for the grounding line.
- **DRIFT-FAIL** — design cell exists but live disagrees (e.g. literal image tag, missing ALB rule, table absent, SSM PLACEHOLDER). At **execution-time re-run** this is real drift → `[BLOCKED]` on the dependent WI until resolved (§21(b) taxonomy).
- **PLANNED-ABSENT** — cell not live **yet**, expected at A-1/A-3 design-time because the PI builds it. MUST map to a roadmap WI (the architect's action item). An ABSENT cell with no owning WI is the gap this skill exists to prevent.
- **UNKNOWN-NOACCESS** — could not verify (no creds / infra offline). Never counts as PASS; blocks the matrix from asserting "grounded"; re-run in CI.
- **N/A** — service has no surface for this layer.

## Procedure

1. **Resolve inputs** — parse PI + services; load TF source + (where reachable) state outputs; record region `ap-south-1`.
2. **Run the 10 checks** per target service. For each cell pick a verdict from the vocabulary above; capture the one-line evidence (command + result excerpt; for SSM, existence + sentinel only — **never the secret value**, per `git-as-memory.md`).
3. **Ground every LIVE-PASS / DRIFT-FAIL in TF** — append the `# tf-ref:` citation (§21(a)); a cell that can't cite TF is itself a finding (resource exists in AWS but not in TF → §21(b) drift row).
4. **Map every PLANNED-ABSENT to a WI** (or flag `no WI yet` for the architect).
5. **Render** the matrix to `portfolio/REPORTS/cloud-state/PI{n}-alignment-matrix.md`.
6. **Append** `portfolio/AUDIT.md` verb `cloud-state-grounded` (or `cloud-state-drift` if any DRIFT-FAIL).
7. **Emit** the exit sentinel + Handoff.

## Output schema (`portfolio/REPORTS/cloud-state/PI{n}-alignment-matrix.md`)

```
# Cloud-State Alignment Matrix — PI{n}[-{lane}]
Target services: <list>   Region: ap-south-1   Date: <ISO8601-UTC>   By: <chief-architect|{product}-architect>
Summary: LIVE-PASS <a> · DRIFT-FAIL <b> · PLANNED-ABSENT <c> · UNKNOWN <d> · N/A <e>   Grounded: YES (no DRIFT/UNKNOWN) | NO

## Matrix
| Svc | # | Layer | Item | Verdict | Evidence (cmd→result) | tf-ref / WI | Owner |
|-----|---|-------|------|---------|-----------------------|-------------|-------|
| iam-api | 6 | SSM | /ampy/plt/prod/iam/jwt-key | DRIFT-FAIL | ssm get-parameter → value==PLACEHOLDER | WI-16.x | platform-engineer |
...

## Findings
- DRIFT-FAIL: <one bullet each — cell, discrepancy, dependent WI, owner agent>
- PLANNED-ABSENT without WI: <bullets — architect must add a WI>
- UNKNOWN-NOACCESS: <bullets — re-run in CI>
```

## Relationship to other skills/agents

- **`smoke-tester` agent (PC-5-02)** runs at **Phase B push-and-test** against *deployed endpoints*. This skill runs at **design time (A-1/A-3)** against *cloud + design state* — it grounds the plan before code; smoke-tester verifies the result after deploy. Complementary, not overlapping; both read TF as primary input (`§21(c)`).
- **`terraform-conventions-linter`** checks TF *source* conventions; this skill checks TF *vs. live AWS + code* alignment. Different planes.

## What this skill does NOT do

- ❌ Does not mutate anything — no `terraform apply`, no `aws ... create/update`, no code/roadmap edits. Reports only.
- ❌ Does not log secret values — SSM/KC checks assert existence + non-PLACEHOLDER sentinel only.
- ❌ Does not author WIs — it flags PLANNED-ABSENT cells; the architect turns them into WIs at A-3.
- ❌ Does not replace integration tests or smoke — it grounds *design* against *state*; it does not exercise endpoints.
- ❌ Does not treat UNKNOWN as PASS — degraded access is fail-loud (Meta-2).

## Exit signal

Complete when the matrix file is written, every target service × applicable layer is a row (no silent omission), every LIVE-PASS/DRIFT-FAIL carries a `# tf-ref:` citation, every PLANNED-ABSENT maps to a WI or is flagged, the AUDIT row is appended, and the run prints `CLOUD-STATE-GROUNDER: COMPLETE — services=<n> rows=<r> drift=<b> unknown=<d> grounded=YES|NO`. Absence of the sentinel = hard stop (never a silent partial matrix).
