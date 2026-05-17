---
name: compose-linter
description: Validate docker-compose files across the portfolio for production-readiness — pinned images, health checks, resource limits, named volumes, port allocation per portfolio-conventions.md.
---

# compose-linter

Static lint of every `docker-compose.yml` under the portfolio.

## Files scanned

- `docker-compose.yml` (root, chief-architect's include-only file)
- `{product}/docker-compose.yml` (per product)
- `{service}/docker-compose.yml` where a service runs its own (e.g. `platform-services/iam/docker-compose.yml`)

Excluded: anything under `node_modules/`, `.venv/`, `dist/`, `build/`.

## Checks (each either PASS or FAIL — no warnings)

### 1. Image pinning
Every `image:` value must include a tag that is not `latest` and not empty. Digest pins (`@sha256:...`) are PASS.

### 2. Health checks
Every service must have a `healthcheck:` block. The root include-only compose is exempt.

### 3. Resource limits
Every service must set both `deploy.resources.limits.cpus` and `deploy.resources.limits.memory`. Sidecar-only or one-shot services (`command: ["...init..."]` with `restart: "no"`) are exempt; note the exemption with a `# compose-linter: skip-resources — <reason>` comment on the service block.

### 4. Named volumes
Any `volumes:` mount that persists state must use a named volume, not a bind mount to the host (except read-only config mounts, which are PASS).

### 5. Port allocation
Every `ports:` host-side port must appear in the authoritative port table in `.claude/rules/portfolio-conventions.md`. Mismatches are FAIL.

### 6. Network isolation
Services within a product share a named network. The root include-only compose must not declare a shared cross-product network; cross-product comms go via host ports.

## Output format

```markdown
## compose-linter — <scope>

| File | Service | Check | Result | Note |
|------|---------|-------|--------|------|

**Summary**: <n> checks, <n> FAIL
```

## Auto-fix policy

**None.** This skill reports only. Fixes are the `platform-engineer` or `chief-architect` (root compose) lane. The linter never modifies files.

## What this skill does NOT do

- Does not run `docker compose config` — that's a syntax check, a separate concern (do both: linter for policy, `config` for syntax).
- Does not lint Dockerfiles — a separate linter.
- Does not inspect image contents.
