# Contributing to YAAgents

---

> **IMPORTANT — Contributions not currently accepted**
>
> YAAgents has not yet completed legal review of its license model.
> **External contributions (code, documentation, or other materials) are not
> accepted until legal review is complete and a Contributor License Agreement
> (CLA) process is in place.**
>
> This is a pre-launch gate, not a permanent policy. We expect to open
> contributions after PI1-yaa PC-6 (legal sign-off). Watch the repository
> or GitHub Discussions for the announcement.
>
> You are welcome to:
> - Open issues (bugs, feature requests, adapter requests)
> - Participate in GitHub Discussions
> - Report security vulnerabilities via [SECURITY.md](SECURITY.md)
>
> You are **not** able to submit pull requests that will be merged at this time.

---

## Contributor License Agreement (CLA) — Placeholder

A CLA will be required for all external contributors once legal review is
complete. The CLA will:

- Grant AimpathyMinds a perpetual, irrevocable, royalty-free license to
  use, reproduce, distribute, and sublicense your contributions.
- Confirm that you have the right to make the contribution.
- Not transfer copyright ownership — you retain your copyright.

The exact CLA text is pending counsel review. When contributions open, a
CLA bot will be wired to every pull request; you will sign electronically
before your first contribution is accepted.

---

## How to open a good issue

Even though code contributions are paused, issue quality matters. The
following guidelines apply immediately.

### Bug reports

Use the **Bug Report** issue template. Include:

- YAAgents component affected (gateway / sdk-fastapi / client-python /
  client-ts / cli / spec)
- Profile version (see `spec/`)
- Minimal reproduction steps (curl, Python/TS snippet, or compose config)
- Observed vs expected response (include the full HTTP status + body)
- Environment (OS, Docker version, language runtime version)

### Feature requests

Use the **Feature Request** issue template. Frame the request as a use-case
problem, not a solution. Include:

- The agentic API pattern you are trying to build
- What the current profile/gateway/SDK forces you to do today
- What you wish you could do instead
- Any OpenAPI or HTTP semantics references that support the change

### Adapter requests

Use the **Adapter Request** issue template if you want native YAAgents
support for a framework (Spring Boot, ASP.NET Core, Express, etc.). Include:

- The framework and language
- Whether you would be willing to maintain the adapter under a future
  contribution model
- Approximate user base / adoption signal

---

## Development setup (for core team)

```bash
# Clone
git clone https://github.com/ai-mpathyminds/yaagents.git
cd yaagents

# Gateway (Go 1.22+)
cd gateway && go build ./... && go test ./...

# Python SDK / client / CLI (Python 3.11+, Hatch)
cd sdk-fastapi && hatch run test
cd client-python && hatch run test
cd cli && hatch run test

# TypeScript client (Node 20+, pnpm)
cd client-ts && pnpm install && pnpm test

# Full demo
cd examples/campaign-api && docker compose up
```

Lint + security gates run in CI (see `.github/workflows/ci.yml`). All PRs
(internal) must pass the full matrix before merge.

---

## Commit conventions

Internal contributors follow Conventional Commits with component scope:

```
feat(gateway): add RBAC policy reload endpoint
fix(client-python): handle 412 with empty clarification_fields
docs(spec): clarify 206 partial-content semantics
```

Commit trailers carry `Agent:` and `WI:` fields per `.claude/rules/git-as-memory.md`
(internal portfolio convention; not required from external contributors once the CLA
opens).

---

## Code of Conduct

All participants — in issues, discussions, and future pull requests — are
expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
