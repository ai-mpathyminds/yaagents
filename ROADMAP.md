# yaagents Roadmap

*Last updated: 2026-06-04. Public roadmap; subject to change.*

## v0.3.x — stabilization (current)

- Stabilize Agentic REST Profile based on community feedback
- Improve gateway plugin examples + per-plugin reference docs
- Build out reusable OpenAPI components in `openapi/`
- More client SDK examples and recipes
- Deployment hardening docs (env vars, secrets, healthcheck conventions)

## v0.4 — production hardening (next)

- **prompt-sanitize**: graduate from v0.3 stub to real implementation (regex + LLM-based guard option)
- **otel-audit**: graduate from v0.3 stub to real OTLP exporter
- **Kubernetes Helm chart** for the gateway + reference deployment
- One additional shipped example (likely **ticket-triage** as a service, after the v0.3 tutorial proves the pattern)
- **GitHub Action** that validates Profile-conformance on every PR
- Custom domain migration to `yaagents.dev` (gated on community traction)

## v0.5+ — community expansion (aspirational)

- Spring Boot SDK
- .NET / ASP.NET Core SDK
- Envoy / Kong / NGINX integration guides
- Rust client SDK
- Ruby client SDK
- A2A / MCP internal-integration reference example

## How to influence the roadmap

- File a feature request via the issue template
- Join GitHub Discussions
- PR a `GOOD_FIRST_ISSUES.md` candidate

Roadmap items are commitments to **direction**, not delivery dates.
