# Contributing to YAAgents

---

> **Legal disclaimer (verbatim — §3, GTM README §Appendix):**
> This includes a draft licensing strategy for product planning. It is not
> legal advice. Before publishing the license publicly or accepting external
> contributions, consult a qualified software licensing lawyer.

---

## Developer Certificate of Origin (DCO)

YAAgents uses the **Developer Certificate of Origin (DCO)** instead of a
Contributor License Agreement (CLA). Every commit you submit must carry a
`Signed-off-by:` trailer certifying that you have the right to submit the
contribution under the Apache 2.0 license.

Add it with:

```bash
git commit --signoff -m "feat(gateway): your change"
```

or manually append to every commit message:

```
Signed-off-by: Your Name <your@email.com>
```

By signing off you confirm that your contribution complies with the
[Developer Certificate of Origin v1.1](https://developercertificate.org/).
The DCO does **not** transfer copyright — you retain ownership of your work.

Pull requests without a `Signed-off-by:` on every commit will fail the DCO
check and cannot be merged.

---

## Pull request checklist

Before opening a PR, verify:

- [ ] Every commit carries `Signed-off-by: Your Name <your@email.com>` (DCO).
- [ ] The PR is linked to an issue (bug, feature, or plugin proposal).
- [ ] `CHANGELOG.md` has an entry under `[Unreleased]` with the change type
      (`Added` / `Changed` / `Fixed` / `Removed`) and a one-line description.
- [ ] Tests pass locally (`go test ./...` for Go; `hatch run test` for Python;
      `pnpm test` for TypeScript).
- [ ] Lint gates pass (`golangci-lint`, `ruff`, `eslint`).
- [ ] New public API surfaces are documented (godoc / docstring / JSDoc).
- [ ] OpenAPI components in `openapi/` are updated if the HTTP contract changed.
- [ ] For plugin WIs: the `Plugin` interface contract is implemented and the
      plugin is registered via `init()` (see **Plugin contribution path** below).

---

## Plugin contribution path

Community plugins extend the YAAgents Gateway at the `Plugin` interface level.
A plugin is a Go package that satisfies:

```go
// plugin/plugin.go (simplified)
type Plugin interface {
    Name() string
    ProcessRequest(ctx context.Context, req *PluginRequest) (*PluginRequest, error)
    ProcessResponse(ctx context.Context, resp *PluginResponse) (*PluginResponse, error)
}
```

### How to contribute a community plugin

1. **Open a Plugin Proposal issue** using the `plugin-proposal` template.
   Include the plugin name, the hook it targets (`request` / `response` /
   `both`), and a minimal API sketch. Wait for a maintainer `go-ahead` comment
   before investing in implementation.

2. **Implement the plugin** in your fork at
   `gateway/plugins/<your-plugin-name>/plugin.go`. Register it via `init()`:

   ```go
   func init() {
       gateway.RegisterPlugin(&MyPlugin{})
   }
   ```

3. **Module path convention** for community plugins published as standalone
   modules: `github.com/<you>/yaagents-plugin-<name>`. The gateway's `plugins/`
   directory hosts first-party plugins; community plugins are imported by users
   who choose to enable them.

4. **Tests required.** Every plugin must ship unit tests. Integration tests
   (using the `examples/` Compose stack) are strongly encouraged.

5. **YAML config.** Plugins that accept configuration must read from
   `gateway.yaml` under a `plugins.<name>:` key. Document the config schema
   in the plugin's README.

6. **Sign off and open the PR** per the DCO checklist above.

---

## How to open a good issue

### Bug reports

Use the **Bug Report** issue template. Include:

- YAAgents component affected (gateway / sdk-fastapi / sdk-go / client-python /
  client-ts / client-go / cli / spec)
- Profile version (see `spec/`)
- Minimal reproduction steps (curl, Python/TS/Go snippet, or compose config)
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
- Whether you would be willing to maintain the adapter
- Approximate user base / adoption signal

### Plugin proposals

Use the **Plugin Proposal** issue template. See the **Plugin contribution
path** section above for what to include.

---

## Development setup

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/ai-mpathyminds/yaagents.git
cd yaagents

# Gateway (Go 1.22+)
cd gateway && go build ./... && go test ./...

# Go server SDK (Go 1.22+)
cd sdk-go && go build ./... && go test ./...

# Python SDK / client / CLI (Python 3.11+, Hatch)
cd sdk-fastapi && hatch run test
cd client-python && hatch run test
cd cli && hatch run test

# TypeScript client (Node 20+, pnpm)
cd client-ts && pnpm install && pnpm test

# Full demo (Python reference)
cd examples/campaign-api && docker compose up

# Full demo (Go reference)
cd examples/campaign-api-go && docker compose up
```

Lint + security gates run in CI (see `.github/workflows/`). All PRs must pass
the full matrix before merge.

---

## Commit conventions

Contributors follow Conventional Commits with component scope:

```
feat(gateway): add RBAC policy reload endpoint
fix(client-python): handle 412 with empty clarification_fields
docs(spec): clarify 206 partial-content semantics
feat(sdk-go): add AgenticResponse.WithCorrelationID helper
```

Every commit must carry a DCO `Signed-off-by:` trailer (see above).

---

## Code of Conduct

All participants — in issues, discussions, and pull requests — are expected
to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
