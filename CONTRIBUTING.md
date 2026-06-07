# Contributing to YAAgents

---

> **Legal disclaimer (verbatim — ADR PI2-yaa-0003 §3, GTM README §Appendix):**
> This includes a draft licensing strategy for product planning. It is not
> legal advice. Before publishing the license publicly or accepting external
> contributions, consult a qualified software licensing lawyer.

---

## How to file issues

Use [GitHub Issues](https://github.com/ai-mpathyminds/yaagents/issues) for all
bug reports, feature requests, plugin proposals, and adapter requests.

**Issue labels:**

| Label | When to use |
|---|---|
| `good first issue` | Small, well-scoped issue suitable for first-time contributors |
| `help wanted` | Open ask for community contribution — maintainers want help here |
| `adapter` | Spring Boot, ASP.NET Core, NestJS, Express, LangGraph, Semantic Kernel, Pydantic AI adapters |
| `documentation` | Docs improvements (README, Pages site, per-component READMEs) |
| `enhancement` | Non-MVP features — Helm chart, LLM-based prompt sanitization, etc. |
| `plugin-proposal` | A new first-party or community gateway plugin |
| `bug` | Confirmed defect with a reproduction case |

**Good bug report: include these.**
- Component affected (gateway / sdk-fastapi / sdk-go / client-python / client-ts / client-go / cli / spec)
- Profile version (check `spec/VERSION`)
- Minimal reproduction steps — curl command, Python/TypeScript/Go snippet, or compose config
- Observed response (full HTTP status + body) vs expected
- Environment (OS, Docker version, language runtime)

**Good feature request: frame it as a use-case problem.**
- The agentic API pattern you are trying to build
- What the current profile/gateway/SDK forces you to do today
- What you wish you could do instead

---

## Dev setup

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/ai-mpathyminds/yaagents.git
cd yaagents

# Gateway (Go 1.22+)
cd gateway && go build ./... && go test ./...

# Go server SDK (Go 1.22+)
cd sdk-go && go build ./... && go test ./...

# Python server SDK (Python 3.11+, Hatch)
cd sdk-fastapi && pip install -e ".[dev]" && hatch run test

# Python client + CLI
cd client-python && pip install -e ".[dev]" && hatch run test
cd cli && pip install -e ".[dev]" && hatch run test

# TypeScript client (Node 20+, pnpm)
cd client-ts && pnpm install && pnpm test

# Docs site (Node 20+, pnpm)
cd docs && pnpm install && pnpm run dev
# Opens http://localhost:4321/yaagents/

# Full demo (Python reference)
cd examples/campaign-api && docker compose up

# Full demo (Go reference)
cd examples/campaign-api-go && docker compose up
```

---

## PR process

**Branch naming:** `feature/<description>`, `fix/<description>`, `docs/<description>`.

**Commit message style:** [Conventional Commits](https://www.conventionalcommits.org/)
with component scope. Every commit must carry a DCO `Signed-off-by:` trailer.

```
feat(gateway): add RBAC policy reload endpoint
fix(client-python): handle 412 with empty clarification_fields
docs(spec): clarify 206 partial-content semantics
feat(sdk-go): add AgenticResponse.WithCorrelationID helper
```

Add the DCO sign-off:
```bash
git commit --signoff -m "feat(gateway): your change"
```

**PR description template** — link these in every PR:
- Issue number: `Closes #<N>` or `Refs #<N>`
- What changed and why (one paragraph)
- Tests run: paste the relevant test command and exit code
- Checklist:
  - [ ] DCO `Signed-off-by:` on every commit
  - [ ] Tests pass (`go test ./...` / `hatch run test` / `pnpm test`)
  - [ ] Lint gates pass (see §Code style)
  - [ ] `CHANGELOG.md` entry under `[Unreleased]`
  - [ ] OpenAPI components updated if HTTP contract changed

Pull requests without `Signed-off-by:` on every commit will fail the DCO check and cannot be merged. The DCO does **not** transfer copyright — you retain ownership of your work.

---

## Code style

| Language | Formatter | Linter / type-checker |
|---|---|---|
| Go | `gofmt` (run automatically by `go build`) | `golangci-lint run ./...` |
| Python | `ruff format` | `ruff check` + `mypy --strict` |
| TypeScript | `prettier` | `eslint` |
| MDX (docs) | — | `pnpm run build` clean (Astro Starlight build) |

CI runs all gates on every pull request. Run them locally before pushing:

```bash
# Go
cd gateway && golangci-lint run ./...

# Python
cd sdk-fastapi && ruff check . && ruff format --check . && mypy src/

# TypeScript
cd client-ts && pnpm lint

# Docs
cd docs && pnpm run build
```

---

## Release process

Releases are **tag-driven**. Pushing a semver tag triggers the publish workflow:

```bash
git tag v0.X.Y && git push origin v0.X.Y
```

The GitHub Actions release workflow (`.github/workflows/release.yml`) publishes:
- Python packages to PyPI via [OIDC Trusted Publisher](https://docs.pypi.org/trusted-publishers/) (no long-lived API keys)
- TypeScript client to npm with provenance attestation
- Gateway container image to GHCR with Cosign signing + Syft SBOM
- Go modules via the Go module proxy (tag push triggers indexing)

**Cross-component release wave** — a major version bump (e.g. v0.4.0) must tag all
eight components at the same version in the same release cycle. The PI4-yaa launch
readiness runbook (LA-* WI series) established the cross-component release-wave pattern:
tag order is `gateway` → `sdk-*` → `client-*` → `cli` → meta-repo; each publish must
succeed before tagging the next component.

Only maintainers with write access to the `ai-mpathyminds` organization can publish releases.

---

## Security disclosure

For security issues, email **security@aimpathyminds.com**. Do not file public
issues for security vulnerabilities. See [SECURITY.md](SECURITY.md) for our
responsible-disclosure policy.

---

## Plugin contribution path

Community plugins extend the YAAgents Gateway at the `Plugin` interface level.
A plugin is a Go package that satisfies:

```go
// gateway/plugins/plugin.go (simplified)
type Plugin interface {
    Name() string
    ProcessRequest(ctx context.Context, req *PluginRequest) (*PluginRequest, error)
    ProcessResponse(ctx context.Context, resp *PluginResponse) (*PluginResponse, error)
}
```

1. Open a `plugin-proposal` issue with the plugin name, hook target, and API sketch. Wait for maintainer `go-ahead` before implementing.
2. Implement at `gateway/plugins/<your-plugin-name>/plugin.go`; register via `init()`.
3. Publish as a standalone module at `github.com/<you>/yaagents-plugin-<name>`.
4. Ship unit tests; integration tests using the examples stack are strongly encouraged.
5. Document the YAML config schema under `plugins.<name>:` in a plugin README.

---

## Code of Conduct

All participants — in issues, discussions, and pull requests — are expected
to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
