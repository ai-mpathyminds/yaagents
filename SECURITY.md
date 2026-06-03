# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `0.3.x` (latest) | Yes |
| `0.2.x` | Yes — security fixes backported |
| `0.1.x` and earlier | No — upgrade to `0.3.x` |

---

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

You have two private channels:

- **E-mail:** security@aimpathyminds.com
- **GitHub Security Advisories:** use the
  [Report a vulnerability](https://github.com/ai-mpathyminds/yaagents/security/advisories/new)
  button in the Security tab of this repository (preferred — keeps tracking in one place).

When reporting, include:

1. **Component** — which part of YAAgents is affected:
   - Gateway (`ghcr.io/ai-mpathyminds/yaagents-gateway`)
   - `yaagents-fastapi` (Python FastAPI server SDK)
   - `yaagents-sdk-go` (Go server SDK)
   - `yaagents-client` (Python client)
   - `@aimpathyminds/yaagents-client` (TypeScript client)
   - `yaagents-client-go` (Go client SDK)
   - `yaagents-cli` (CLI validator)
   - Profile spec / JSON schemas
2. **Description** — a clear summary of the vulnerability and its potential impact.
3. **Reproduction steps** — minimal curl / code snippet that demonstrates the
   issue. If a proof-of-concept exploit exists, include it.
4. **Environment** — OS, runtime versions, gateway version or image digest.
5. **Suggested severity** — CVSS score or Low / Medium / High / Critical if you
   have assessed it.

We will acknowledge receipt within **3 business days** and aim to provide an
initial assessment within **7 business days**. Critical vulnerabilities (RCE,
auth bypass, credential exposure) are treated as P0 and triaged immediately.

---

## Disclosure policy

We follow **coordinated disclosure** with a **90-day disclosure SLA**:

1. Reporter submits vulnerability privately (e-mail or GitHub Security Advisory).
2. AimpathyMinds investigates and develops a fix.
3. A patched release is prepared within **90 days** of the initial report.
   We aim to release faster for High/Critical issues, but the SLA is universal —
   90 days applies to all severity levels.
4. Reporter is notified before public disclosure and credited in the advisory
   unless they request anonymity.
5. A GitHub Security Advisory is published alongside the patch release.

If 90 days passes without a patch, we will coordinate with the reporter on
a disclosure timeline that protects users as much as possible.

We do not offer a bug bounty program at this time.

---

## Scope

In-scope:

- Authentication and authorization bypass in the gateway
- Tenant isolation failures (cross-tenant data leakage)
- RBAC policy bypass
- Secret exposure via logs, error responses, or image layers
- Supply-chain issues (dependency with active CVE pulled into a release)
- Input-validation gaps enabling injection or denial-of-service

Out-of-scope:

- Issues in the user's agent implementation (bring-your-own-agent runtime)
- Theoretical vulnerabilities without a realistic attack path
- Issues requiring physical access to the host
- Social engineering of AimpathyMinds team members
- Denial-of-service via resource exhaustion against a user-controlled deployment
  without gateway misconfiguration

---

## Security defaults

The YAAgents Gateway ships with the following security defaults:

- Non-root container user
- Read-only filesystem (where applicable)
- No secrets baked into the image or default configuration
- JWT validation required on all upstream routes (disabled only in explicit dev mode)
- Structured JSON logs — no credential values logged by default
- Graceful shutdown to prevent in-flight request drops during deploys

If you deploy the gateway, review the configuration guide before exposing it to
untrusted traffic.

---

## Dependency scanning

Every release runs:

- `govulncheck` on the Go gateway and Go SDK
- `trivy` on the gateway container image
- `pip-audit` on the Python packages
- `npm audit` on the TypeScript package

SBOM (`sbom.spdx.json`) is attached to each GitHub Release.
