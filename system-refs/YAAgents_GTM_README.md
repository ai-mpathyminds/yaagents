# YAAgents GTM Task README

## Purpose

This README defines the go-to-market execution plan for **YAAgents**, a source-available developer project for building governed, resource-oriented agentic APIs.

The product positioning:

> **Build agentic APIs, not chatbots.**

The operating thesis:

> **Keep your domain resources. Make selected operations agentic.**

YAAgents is not positioned as another agent framework. It is the interface, gateway, response-contract, and client-consumption layer for agentic capabilities built using any internal framework.

---

## 1. GTM Objective

The first GTM objective is to establish a new developer category:

> **Agentic APIs**

YAAgents should become associated with the following idea:

> Generic chat is not the right interface for production business workflows. Agentic capabilities should be exposed through normal domain resources, typed HTTP responses, gateway-controlled access, OpenAPI contracts, and native clients.

Example:

```http
POST /campaigns/{campaignId}/optimizations
```

instead of:

```http
POST /agents/campaign-agent/invoke
```

---

## 2. Target Audience

### Primary Audience

- Platform engineers
- API architects
- Backend engineers
- AI platform teams
- Developer platform teams
- Enterprise architects
- SaaS product engineering teams

### Secondary Audience

- LangGraph users
- Semantic Kernel users
- Pydantic AI users
- FastAPI developers
- OpenAPI-first teams
- API gateway teams
- Kubernetes platform teams
- Internal Developer Platform teams

---

## 3. Core Message

Use this message consistently:

> **Build the agent however you want. Expose it like a governed API.**

Supporting messages:

- YAAgents standardizes the external interface, not the internal intelligence.
- Developers can use LangGraph, Semantic Kernel, Pydantic AI, LangChain, direct LLM SDKs, rules, or custom code.
- YAAgents focuses on interface, access control, response contracts, OpenAPI, gateway integration, and native clients.
- Kubernetes remains the runtime.
- The gateway becomes the control point.
- Native clients make agentic responses feel natural in application code.

---

## 4. Product Wedge

The wedge is **not** “a better agent framework.”

The wedge is:

> **The missing API layer for production agentic systems.**

### Contrast

| Current Agent Pattern | YAAgents Pattern |
|---|---|
| Generic chatbot | Domain resource API |
| `/agents/invoke` | `/campaigns/{id}/optimizations` |
| Free-form text | Typed responses |
| Prompt parsing | Status/content-type driven handling |
| Weak access control | Gateway RBAC |
| Framework-specific integration | Framework-neutral contract |
| Ad hoc clarification | `clarification_required` response |
| Client-specific hacks | Native client SDK behavior |

---

## 5. MVP Release Assets

Before public launch, the following assets must be ready.

### Repository

- [ ] Public GitHub repository
- [ ] Clear README
- [ ] Architecture diagram
- [ ] PRD README
- [ ] GTM README
- [ ] License file
- [ ] Contributing guide
- [ ] Code of conduct
- [ ] Security policy
- [ ] Roadmap
- [ ] GitHub Discussions enabled
- [ ] Good-first-issue labels
- [ ] Project board

### Product Artifacts

- [ ] Go gateway container
- [ ] Python FastAPI SDK
- [ ] Python client
- [ ] TypeScript client
- [ ] CLI validator
- [ ] JSON schemas
- [ ] OpenAPI reusable components
- [ ] Campaign API reference example
- [ ] Docker Compose demo
- [ ] Kubernetes manifests
- [ ] Basic Helm chart or Helm chart placeholder

### Documentation

- [ ] Getting started guide
- [ ] Build your first agentic API
- [ ] Response contract guide
- [ ] Gateway configuration guide
- [ ] Bring-your-own-agent guide
- [ ] Client SDK guide
- [ ] OpenAPI generation guide
- [ ] Kubernetes deployment guide
- [ ] FAQ: Why not just OpenAPI?
- [ ] FAQ: Why not LangGraph/Semantic Kernel?
- [ ] FAQ: Is this open source?

### Demo Assets

- [ ] 2-minute product demo video
- [ ] 5-minute technical walkthrough
- [ ] GIF showing `clarification_required`
- [ ] Example OpenAPI snippet
- [ ] Example gateway route config
- [ ] Example TypeScript client handling clarification
- [ ] Example Python client handling clarification

---

## 6. Launch Narrative

### Primary Campaign Title

> **Stop Exposing Agents as Chatbots. Build Agentic APIs.**

### Article Outline

1. Production systems are resource-oriented.
2. Chat is a poor integration surface for business workflows.
3. Agentic behavior should live behind domain resource operations.
4. Responses should be typed, not free-form.
5. Clarification should be a machine-readable API response.
6. Gateway control is essential for access, tenant context, and audit.
7. Developers should bring their own agent framework.
8. YAAgents provides the interface layer.

### Short Social Post

```text
Most production systems do not need more chatbots.

They need agentic APIs.

Instead of exposing /agents/invoke, expose domain operations:

POST /campaigns/{id}/optimizations
POST /tickets/{id}:triage
POST /claims/{id}/reviews

YAAgents standardizes typed responses, clarification flows, gateway access control, OpenAPI contracts, and native clients for these agentic APIs.

Build the agent however you want. Expose it like a governed API.
```

---

## 7. Launch Channels

### Developer Channels

- GitHub
- Hacker News: Show HN
- Product Hunt
- Dev.to
- Medium
- daily.dev
- Lobsters
- Reddit:
  - r/programming
  - r/kubernetes
  - r/FastAPI
  - r/MachineLearning
  - r/LocalLLaMA
  - r/devops
- LinkedIn engineering audience

### Ecosystem Channels

- OpenAPI communities
- CNCF Slack
- Kubernetes Slack
- Platform Engineering Slack
- LangChain / LangGraph forums
- Semantic Kernel GitHub discussions
- FastAPI community
- Backstage community
- API gateway communities:
  - Kong
  - Tyk
  - Apache APISIX
  - Envoy Gateway
  - KrakenD

### Local / Regional Channels

- Pune developer meetups
- Bangalore platform engineering meetups
- Indian AI developer communities
- LinkedIn architecture network
- University/open-source clubs for academic adoption

---

## 8. 90-Day GTM Plan

## Phase 1: Preparation — Days 1 to 15

Goal: Make the project launch-ready.

### Tasks

- [ ] Finalize product positioning
- [ ] Finalize repository structure
- [ ] Add gateway Docker build
- [ ] Add FastAPI SDK package
- [ ] Add Python client package
- [ ] Add TypeScript client package
- [ ] Add CLI validator
- [ ] Add schemas and OpenAPI components
- [ ] Add campaign API reference example
- [ ] Add Docker Compose demo
- [ ] Add K8s manifests
- [ ] Create architecture diagram
- [ ] Write launch manifesto
- [ ] Record demo video
- [ ] Prepare launch posts
- [ ] Create issue templates
- [ ] Create good-first issues
- [ ] Add license and commercial license notice

### Success Target

- Local demo works with one command.
- README explains the concept in less than 60 seconds.
- First-time user can run example in less than 10 minutes.

---

## Phase 2: Soft Launch — Days 16 to 30

Goal: Validate positioning before broad launch.

### Tasks

- [ ] Share architecture article on LinkedIn
- [ ] Share design post with API/platform engineers
- [ ] Ask for feedback from 5-10 trusted architects
- [ ] Open GitHub Discussions
- [ ] Invite feedback on response contract
- [ ] Invite feedback on license model
- [ ] Share in smaller developer communities
- [ ] Collect objections and update FAQ

### Success Target

- 50-100 GitHub stars
- 5 meaningful technical comments
- 2 framework adapter requests
- 1 external person runs the example
- 1 external issue or PR

---

## Phase 3: Public Developer Launch — Days 31 to 45

Goal: Drive awareness and stars.

### Tasks

- [ ] Publish “Stop Exposing Agents as Chatbots” article
- [ ] Launch Show HN
- [ ] Launch Product Hunt
- [ ] Post on Dev.to
- [ ] Post short demo on LinkedIn
- [ ] Share GIF of clarification handling
- [ ] Share technical walkthrough video
- [ ] Submit to daily.dev
- [ ] Share with OpenAPI/API gateway communities
- [ ] Share with LangGraph/Semantic Kernel communities

### Success Target

- 300-500 GitHub stars
- 20 issues/discussions
- 3 external contributors/watchers
- 1 company/team interested in pilot
- Clear top 3 requested adapters

---

## Phase 4: Ecosystem Build — Days 46 to 90

Goal: Convert attention into community and ecosystem.

### Tasks

- [ ] Add LangGraph example
- [ ] Add Semantic Kernel example
- [ ] Add OpenTelemetry design issue
- [ ] Add OPA integration design issue
- [ ] Add Spring Boot adapter roadmap
- [ ] Add ASP.NET Core adapter roadmap
- [ ] Add NestJS adapter roadmap
- [ ] Publish technical blog on OpenAPI response contracts
- [ ] Publish technical blog on gateway-controlled agents
- [ ] Present at one meetup or virtual session
- [ ] Create contributor onboarding call/demo
- [ ] Create public roadmap board

### Success Target

- 1,000 GitHub stars
- 3-5 external contributors
- 2 real pilot users
- 1 community-maintained adapter in progress
- 1 ecosystem integration discussion

---

## 9. Content Backlog

### Category Articles

- [ ] Stop Exposing Agents as Chatbots
- [ ] What Are Agentic APIs?
- [ ] Why Resource-Oriented APIs Matter for Agents
- [ ] Bring Your Own Agent Framework, But Standardize the Interface

### Technical Articles

- [ ] Designing Agentic APIs with REST Resources
- [ ] Why Clarification Should Be a Typed API Response
- [ ] 400 vs 422 vs 412 for Agentic APIs
- [ ] OpenAPI Response Contracts for Agentic APIs
- [ ] Gateway-Controlled Agentic APIs
- [ ] Kubernetes as Runtime for Agentic APIs
- [ ] How to Build a Campaign Optimization API with FastAPI
- [ ] How TypeScript Clients Should Handle `clarification_required`

### Demo Videos

- [ ] What is YAAgents? — 2 minutes
- [ ] Build an Agentic API with FastAPI — 5 minutes
- [ ] Gateway RBAC Demo — 5 minutes
- [ ] Client Handles Clarification — 5 minutes
- [ ] Campaign Optimization End-to-End — 8 minutes

---

## 10. Community Tasks

### GitHub Labels

- [ ] `good first issue`
- [ ] `help wanted`
- [ ] `adapter`
- [ ] `client-sdk`
- [ ] `gateway`
- [ ] `openapi`
- [ ] `schema`
- [ ] `docs`
- [ ] `example`
- [ ] `conformance`
- [ ] `license`
- [ ] `discussion`

### Good First Issues

- [ ] Add more clarification examples
- [ ] Add validation error examples
- [ ] Improve README quickstart
- [ ] Add curl examples
- [ ] Add Postman collection
- [ ] Add OpenAPI example
- [ ] Add Docker Compose troubleshooting
- [ ] Add client SDK usage examples
- [ ] Add gateway config examples
- [ ] Add error schema tests

### Adapter Issues

- [ ] Spring Boot adapter
- [ ] ASP.NET Core adapter
- [ ] NestJS adapter
- [ ] Express adapter
- [ ] Go Gin adapter
- [ ] Go Fiber adapter

---

## 11. Publishing Model

### Gateway

Artifact:

```text
ghcr.io/yaagents/gateway:<version>
```

Release tasks:

- [ ] Build multi-arch Docker image
- [ ] Publish to GHCR
- [ ] Add Docker Hub mirror later
- [ ] Generate SBOM
- [ ] Sign image later with Cosign

### Python FastAPI SDK

Artifact:

```text
yaagents-fastapi
```

Registry:

```text
PyPI
```

Release tasks:

- [ ] Package with Poetry or Hatch
- [ ] Publish to TestPyPI
- [ ] Publish to PyPI
- [ ] Add typed package metadata
- [ ] Add examples

### Python Client

Artifact:

```text
yaagents-client
```

Registry:

```text
PyPI
```

Release tasks:

- [ ] Package with Poetry or Hatch
- [ ] Add sync and async client if feasible
- [ ] Add typed exceptions/results
- [ ] Publish to PyPI

### TypeScript Client

Artifact:

```text
@yaagents/client
```

Registry:

```text
npm
```

Release tasks:

- [ ] Build ESM package
- [ ] Add TypeScript types
- [ ] Add browser and Node support
- [ ] Add result-style response handling
- [ ] Add exception-style helper
- [ ] Publish to npm

### CLI

Artifact:

```text
yaagents-cli
```

Registry:

```text
PyPI or GitHub Releases
```

Release tasks:

- [ ] Add `validate-openapi`
- [ ] Add `validate-response`
- [ ] Add `conformance-test`
- [ ] Add `init fastapi`
- [ ] Add release binaries later

### Schemas and OpenAPI Components

Artifacts:

```text
schemas/*.schema.json
openapi/*.yaml
```

Release tasks:

- [ ] Publish inside repo
- [ ] Attach to GitHub releases
- [ ] Add versioned folder
- [ ] Later publish stable URLs

---

## 12. License Strategy

## Important Naming Note

If YAAgents restricts commercial use by larger organizations, it should **not** be marketed as OSI-approved open source.

The better wording is:

> **source-available**
> **fair-code**
> **free for academics, individuals, and small developers**
> **commercial license required for larger organizations**

This avoids confusion with OSI-style open source.

## Recommended License Model

Use a **dual-license model**:

### Free Community License

For:

- Individual developers
- Academic research
- Teaching
- Student projects
- Non-commercial use
- Small developers and small companies below a defined threshold

### Commercial License

Required for:

- Larger organizations
- Funded companies above the threshold
- Commercial SaaS usage
- Internal enterprise production usage
- Embedding YAAgents into a commercial platform
- Offering hosted/managed services using YAAgents

---

## 13. Suggested Thresholds

You can choose one of these models.

### Option A: Revenue Threshold

Free if the organization has:

```text
Annual gross revenue < USD 1 million
```

Commercial license required if:

```text
Annual gross revenue >= USD 1 million
```

### Option B: Employee Threshold

Free if the organization has:

```text
Fewer than 10 employees
```

Commercial license required if:

```text
10 or more employees
```

### Option C: Funding Threshold

Free if the organization has:

```text
Raised less than USD 500,000
```

Commercial license required if:

```text
Raised USD 500,000 or more
```

### Recommended Initial Threshold

Use a combined threshold:

```text
Free for individuals, academic use, non-commercial use, and organizations with fewer than 10 employees and less than USD 1 million annual revenue.

Commercial license required for organizations with 10 or more employees, USD 1 million or more annual revenue, or production commercial use at scale.
```

This is simple enough to understand.

---

## 14. Suggested License Text

> This is a business and product draft, not legal advice. A qualified software licensing lawyer should review before publication.

Create a file:

```text
LICENSE
```

Suggested title:

```text
YAAgents Community License v0.1
```

Draft:

```text
YAAgents Community License v0.1

Copyright (c) [YEAR] [OWNER]

Permission is granted to use, copy, modify, and distribute this software and its documentation for the following permitted uses:

1. Personal use by individual developers.
2. Academic research, teaching, and student projects.
3. Non-commercial use.
4. Evaluation, testing, and proof-of-concept use.
5. Use by small organizations with fewer than 10 employees and less than USD 1,000,000 in annual gross revenue.

The following uses require a separate commercial license from the copyright holder:

1. Production use by organizations with 10 or more employees.
2. Production use by organizations with USD 1,000,000 or more in annual gross revenue.
3. Use as part of a commercial SaaS, hosted, or managed service offering.
4. Embedding, bundling, or redistributing the software as part of a commercial product.
5. Use by consulting, system integration, or platform companies to deliver paid services to third parties.
6. Use that competes with a paid YAAgents offering.

You may not remove copyright notices, license notices, or attribution notices.

You may not use the YAAgents name, logo, or trademarks to imply endorsement without written permission.

The software is provided "as is", without warranty of any kind, express or implied, including but not limited to warranties of merchantability, fitness for a particular purpose, and non-infringement.

In no event shall the authors or copyright holders be liable for any claim, damages, or other liability arising from the software or its use.

For commercial licensing, contact: [commercial-contact-email]
```

---

## 15. Alternative License Options

### Option 1: PolyForm Noncommercial + Commercial License

Good if you want a standardized non-commercial license.

Pros:

- Existing known license family
- Clear non-commercial positioning
- Better than inventing everything yourself

Cons:

- Does not automatically solve small-company free usage unless customized with another license
- Not OSI open source

### Option 2: Functional Source License

Good if you want fair-source with eventual open-source conversion.

Pros:

- Known fair-source model
- Converts to Apache 2.0 or MIT after a delay
- Friendly to developer trust

Cons:

- More permissive over time than you may want
- Designed more for preventing direct free-riding than for revenue threshold licensing

### Option 3: Custom Community License

Good if your goal is exactly:

```text
free for academics + small developers
commercial for larger organizations
```

Pros:

- Matches business intent
- Simple to explain

Cons:

- Requires legal review
- Some enterprises avoid custom licenses
- Not OSI open source

### Recommended Path

For v0.1:

```text
Use a custom Community License draft + separate Commercial License.
Do not call it OSI open source.
Call it source-available or fair-code.
```

Later, if adoption grows, consider moving parts of the ecosystem to Apache 2.0 while keeping enterprise gateway features commercial.

---

## 16. Commercial Packaging Later

### Free Community Edition

- Response profile
- Schemas
- OpenAPI components
- FastAPI SDK
- Python client
- TypeScript client
- Basic gateway
- Campaign example

### Commercial Edition

- Enterprise gateway
- Advanced RBAC
- OPA policy integration
- Audit dashboards
- Multi-tenant management
- Admin UI
- SSO templates
- Enterprise support
- Compliance reports
- Private adapter support
- Hosted control plane

---

## 17. GTM Metrics

### Awareness

- GitHub stars
- Article views
- Social shares
- Newsletter mentions
- Community discussions

### Developer Activation

- Repo clones
- Docker Compose runs
- PyPI downloads
- npm downloads
- Gateway image pulls
- CLI validation runs

### Community

- Issues created
- Discussions opened
- PRs submitted
- Adapter requests
- External examples created

### Commercial Intent

- Enterprise inquiries
- Commercial license requests
- Pilot requests
- Integration calls
- Support requests from funded companies

---

## 18. Launch Checklist

### Before Soft Launch

- [ ] README complete
- [ ] License added
- [ ] PRD README added
- [ ] GTM README added
- [ ] Docker Compose works
- [ ] Gateway image builds
- [ ] FastAPI SDK installs locally
- [ ] Python client works
- [ ] TypeScript client works
- [ ] OpenAPI generation works
- [ ] Campaign example documented
- [ ] Clarification demo works
- [ ] Good-first issues created

### Before Public Launch

- [ ] PyPI packages published
- [ ] npm package published
- [ ] Gateway image published
- [ ] Demo video published
- [ ] Launch blog published
- [ ] Show HN post ready
- [ ] Product Hunt assets ready
- [ ] LinkedIn launch post ready
- [ ] Dev.to post ready
- [ ] FAQ updated
- [ ] Commercial licensing email ready

---

## 19. Recommended First Launch Post

```text
Most production systems do not need more chatbots.

They need agentic APIs.

YAAgents is a source-available project for exposing agentic capabilities through normal domain resources:

POST /campaigns/{id}/optimizations
POST /tickets/{id}:triage
POST /claims/{id}/reviews

Instead of free-form chat responses, YAAgents defines typed HTTP responses:

- success
- clarification_required
- validation_failed
- approval_required
- failed_dependency

It also provides:

- Go gateway
- FastAPI SDK
- Python client
- TypeScript client
- OpenAPI components
- JSON schemas
- Campaign API example

Build the agent however you want.
Expose it like a governed API.
```

---

## 20. Immediate Next Actions

1. Finalize license decision.
2. Add `LICENSE` and `COMMERCIAL.md`.
3. Publish PRD README.
4. Publish GTM README.
5. Create GitHub repo structure.
6. Build gateway Docker skeleton.
7. Build FastAPI response builder.
8. Build campaign example.
9. Build TypeScript client response handling.
10. Record clarification demo.
11. Soft launch architecture post.
12. Invite early architecture feedback.

---

## Appendix: License Disclaimer

This GTM README includes a draft licensing strategy for product planning. It is not legal advice. Before publishing the license publicly or accepting external contributions, consult a qualified software licensing lawyer.

---

## Amendment — v0.2.0 Apache 2.0 Transition (2026-05-30)

**Sections §12–§15 of this GTM README describe the v0.1.x source-available / dual-license
model. The following amendment supersedes that model for v0.2.0 and later.**

### §12 (Amended) — License Strategy for v0.2.0+

**v0.2.0 ships under Apache 2.0.** The YAAgents Community License (§14) and COMMERCIAL.md
commercial terms (§16) are retired with v0.2.0. The strategic rationale (user-direct,
2026-05-30): the community-contribution gate imposed by the source-available license outweighs
the commercial-paywall optionality; Apache 2.0 enables the community plugin flywheel that is
the core value proposition of v0.2.0+.

**Marketing / README MUST use:**
- "Apache 2.0 open source" (OSI-approved)
- No longer "source-available" or "fair-code" from v0.2.0 onward

**Non-retroactive boundary:** v0.1.x packages already published to PyPI/npm/GHCR stay under
the YAAgents Community License. Users who need Apache 2.0 must upgrade to v0.2.0.

**Legal-review-pending disclaimer (verbatim, GTM README §Appendix):**

> This GTM README includes a draft licensing strategy for product planning. It is not legal
> advice. Before publishing the license publicly or accepting external contributions, consult
> a qualified software licensing lawyer.

Legal review gates the public re-announce and the removal of the `legal-review-pending` banner
from `CONTRIBUTING.md`. It does **not** gate PI2-yaa close; the `LICENSE` file, copyright
headers, and package metadata ship with the Apache 2.0 text before legal sign-off.

### §13 (Superseded) — Thresholds

The employee / revenue / funding thresholds defined in §13 (v0.1.x) no longer apply from
v0.2.0 onward. Apache 2.0 has no usage thresholds.

### §14 (Superseded) — Community License Text

The "YAAgents Community License v0.1" draft in §14 applied to v0.1.x only. From v0.2.0, the
`LICENSE` file in the repository root contains the Apache License, Version 2.0 text verbatim
(source: `http://www.apache.org/licenses/LICENSE-2.0.txt`). See
`yaagents/system-refs/yaagents-v0.2_detailed.md §8.2` for the verbatim text.

### §15 (Superseded) — Alternative License Options

The §15 alternative license analysis (PolyForm Noncommercial, Functional Source License,
custom Community License) is superseded. No alternative license analysis is required for
v0.2.0; the decision is Apache 2.0, locked.

### §16 (Superseded) — Commercial Packaging

The Community Edition / Commercial Edition split described in §16 is retired. From v0.2.0,
all yaagents components ship under Apache 2.0 as a single unified edition. Commercial
packaging (enterprise gateway, admin UI, OPA integration, hosted control plane) may be
re-introduced in a future PI as a separate commercial product built atop the Apache 2.0 core;
that is a separate business decision requiring a fresh GTM analysis.

Full v0.2.0 PRD: `yaagents/system-refs/yaagents-v0.2_detailed.md` [READY]

