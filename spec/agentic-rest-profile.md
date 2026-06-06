# Agentic REST Response Profile

**Profile version:** v0.3
**Profile header literal:** `X-YAAgents-Profile: v0.3`
**Authoritative source:** `spec/` (ADR PI1-yaa-0002 §1 — sole table source; no other
component may redefine or paraphrase the normative table below)
**Version file:** `spec/VERSION` = `0.3`
**Date adopted:** 2026-06-02 (v0.3 profile bump; sdk-go server SDK component added — see §Changelog)
**Prior version:** v0.2 adopted 2026-06-01; frozen schemas at `schemas/v0.2/`; v0.1 adopted 2026-05-17; frozen schemas at `schemas/v0.1/`

---

## 1. Purpose

This document is the normative prose definition of the YAAgents Agentic REST Response
Profile. It governs every component in the yaagents repository: the JSON schemas
(`schemas/`), the OpenAPI reusable components (`openapi/`), the Go gateway (`gateway/`),
the Python FastAPI SDK (`sdk-fastapi/`), both clients (`client-python/`, `client-ts/`),
the CLI validator (`cli/`), and the reference example (`examples/campaign-api/`).

**Consumers reference this file; they do not restate it.**

---

## 2. Profile Version Declaration

Every published package MUST carry the following declaration in its metadata:

```text
Supports YAAgents Profile v0.3
```

The HTTP request / response header that identifies profile compliance is:

```text
X-YAAgents-Profile: v0.3
```

This header MUST be emitted on every agentic response served through the yaagents
gateway, including SSE streaming responses (see §11). Clients MAY inspect it for
version negotiation.

---

## 3. Governing Contrast

YAAgents makes agentic behaviour look and behave like normal REST:

```text
POST /campaigns/{campaignId}/optimizations   ← YAAgents pattern
POST /agents/campaign-agent/invoke           ← NOT this
```

The agent remains an implementation detail behind a controlled resource operation.

---

## 4. Agentic REST Response Profile (NORMATIVE)

The following table is the normative Agentic REST Response Profile. All components MUST
implement and honour it. Paraphrasing or diverging from these media types is prohibited.

| Response Type | HTTP Status | Content-Type |
|---|---:|---|
| `success` | `200` | `application/json` |
| `created` | `201` | `application/json` |
| `accepted` | `202` | `application/vnd.yaagents.operation+json` |
| `clarification_required` | `400` | `application/vnd.yaagents.clarification+json` |
| `validation_failed` | `422` | `application/vnd.yaagents.validation-error+json` |
| `approval_required` | `412` | `application/vnd.yaagents.approval-required+json` |
| `forbidden` | `403` | `application/vnd.yaagents.error+json` |
| `conflict` | `409` | `application/vnd.yaagents.conflict+json` |
| `failed_dependency` | `424` | `application/vnd.yaagents.error+json` |
| `error` | `500` | `application/vnd.yaagents.error+json` |

**Profile versioning:** Every published package MUST declare the profile version it supports,
e.g., `Supports YAAgents Profile v0.3`.

### 4.1 Clarification Required — canonical body shape

```json
{
  "type": "clarification_required",
  "code": "CLARIFICATION_REQUIRED",
  "message": "Additional information is required.",
  "requiredInputs": [
    {
      "name": "successMetric",
      "location": "body",
      "type": "string",
      "required": true,
      "question": "Which success metric should be optimized?",
      "allowedValues": ["ctr", "cpl", "conversion_rate", "lead_quality"]
    }
  ],
  "trace": {
    "correlationId": "corr-123",
    "requestId": "req-456"
  }
}
```

`correlationId` and `requestId` are mandatory in every agentic response body (`trace` block).

---

## 5. Mandatory `trace` Block Contract

**Every agentic vendor-typed response body MUST include a `trace` object** at the
top level with the following two fields:

| Field | Type | Description |
|-------|------|-------------|
| `correlationId` | string | Cross-service correlation ID; propagated from `X-Correlation-ID`. |
| `requestId` | string | Per-request unique ID; propagated from `X-Request-ID` or gateway-generated. |

This applies to all ten response types in the normative table (§4), including:
`accepted`, `clarification_required`, `validation_failed`, `approval_required`,
`forbidden`, `conflict`, `failed_dependency`, and `error`.

The `success` (200) and `created` (201) response types use `application/json` — their
body shape is domain-specific and owned by the implementing service; the `trace` block
is **RECOMMENDED** but not mandated for these two types because they carry no
`application/vnd.yaagents.*` content-type.

### 5.1 Gateway propagation rule

The yaagents gateway MUST:

1. Extract `X-Correlation-ID` from the inbound request (generate a UUID v4 if absent).
2. Extract `X-Request-ID` from the inbound request (generate a UUID v4 if absent).
3. Forward both headers to the upstream handler.
4. Verify that the upstream response body (for all `application/vnd.yaagents.*` types)
   contains a `trace` object matching the two propagated values.
5. Reject (500) any upstream agentic response that is missing the `trace` block.

---

## 6. `requiredInputs` Array Shape (Clarification Required)

When the response type is `clarification_required`, the body MUST include a
`requiredInputs` array. Each element MUST conform to:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Field or parameter identifier the caller must supply. |
| `location` | string | yes | One of `body`, `query`, `path`, `header`. |
| `type` | string | yes | JSON type hint: `string`, `integer`, `boolean`, `array`, `object`. |
| `required` | boolean | yes | Whether the input is mandatory to proceed. |
| `question` | string | yes | Human-readable prompt surfaced to the end user. |
| `allowedValues` | array | no | Enumeration of acceptable values; omit if open-ended. |

The array MUST contain at least one element; an empty `requiredInputs` array on a
`clarification_required` response is a schema violation.

---

## 7. Response Type Reference

### 7.1 `success` — 200 `application/json`

Normal synchronous success. Body shape is domain-defined. `trace` RECOMMENDED.

### 7.2 `created` — 201 `application/json`

Resource created synchronously. Body shape is domain-defined. `trace` RECOMMENDED.

### 7.3 `accepted` — 202 `application/vnd.yaagents.operation+json`

The operation was accepted for asynchronous processing. The body MUST include:

- `type`: `"operation_accepted"`
- `operationId`: string (stable identifier for status polling)
- `statusUrl`: string (URL to poll; relative or absolute)
- `trace`: object (mandatory)

> **PI1-yaa scope note (ADR PI1-yaa-0002 §4):** The schema and SDK factory for this
> type ship in PI1-yaa. The async polling runtime (`GET …/{opId}/status`) is v0.3 scope
> and is not built in PI1-yaa. The reference example does not exercise a 202 flow.

### 7.4 `clarification_required` — 400 `application/vnd.yaagents.clarification+json`

The agent requires additional inputs before it can proceed. Body shape: §4.1 + §6.

### 7.5 `validation_failed` — 422 `application/vnd.yaagents.validation-error+json`

The request inputs failed validation. The body MUST include:

- `type`: `"validation_failed"`
- `code`: `"VALIDATION_FAILED"`
- `message`: human-readable description
- `errors`: array of `{ field, message }` objects
- `trace`: object (mandatory)

### 7.6 `approval_required` — 412 `application/vnd.yaagents.approval-required+json`

The operation requires explicit human approval before execution. The body MUST include:

- `type`: `"approval_required"`
- `code`: `"APPROVAL_REQUIRED"`
- `message`: human-readable description
- `approvalToken`: string (opaque token the approver supplies to proceed)
- `trace`: object (mandatory)

### 7.7 `forbidden` — 403 `application/vnd.yaagents.error+json`

The caller lacks permission. Body: generic agentic error shape (§7.10).

### 7.8 `conflict` — 409 `application/vnd.yaagents.conflict+json`

A conflicting resource state prevents the operation. The body MUST include:

- `type`: `"conflict"`
- `code`: string (service-specific conflict code)
- `message`: human-readable description
- `conflictingResourceId`: string (optional but RECOMMENDED)
- `trace`: object (mandatory)

### 7.9 `failed_dependency` — 424 `application/vnd.yaagents.error+json`

An upstream dependency failed. Body: generic agentic error shape (§7.10).

### 7.10 `error` — 500 `application/vnd.yaagents.error+json`

Internal server error. The body MUST include:

- `type`: one of `"forbidden"`, `"failed_dependency"`, `"error"`
- `code`: string (service-specific error code)
- `message`: human-readable description
- `trace`: object (mandatory)

---

## 8. Conformance Acceptance Criteria

A component is **conformant** with YAAgents Profile v0.3 if and only if:

1. **Table fidelity** — the component maps every one of the 10 response types to the
   exact HTTP status code and Content-Type string listed in §4. No aliases, no
   alternate spellings.

2. **Mandatory `trace` block** — every response body emitted with a
   `application/vnd.yaagents.*` Content-Type carries a `trace` object containing
   both `correlationId` (non-empty string) and `requestId` (non-empty string).

3. **`clarification_required` body shape** — when the `clarification_required` type is
   returned, the body matches §4.1 exactly: `type`, `code`, `message`, `requiredInputs`
   (array, ≥1 element, each element matching the shape in §6), and `trace`.

4. **Profile header** — the gateway emits `X-YAAgents-Profile: v0.3` on every agentic
   response, including SSE streaming responses (§11).

5. **Package declaration** — every published package carries `Supports YAAgents Profile v0.3`
   in its metadata (`pyproject.toml`, `package.json`, or equivalent).

6. **Schema conformance** — every vendor-typed response body validates against its
   corresponding JSON schema in `schemas/v0.3/` (ADR PI1-yaa-0002 §2). The frozen
   backward-compat schemas at `schemas/v0.2/` remain valid for v0.2.x consumers;
   `schemas/v0.1/` remain valid for v0.1.x consumers.

7. **No table redefinition** — no component file contains a copy or paraphrase of the
   10-row table in §4. The grep command below is the EX-4 gate check:

   ```bash
   grep -rn "vnd\.yaagents\." \
     gateway/ sdk-fastapi/ client-python/ client-ts/ cli/ examples/ openapi/ schemas/ \
     | grep -v "^Binary" \
     | grep -v "spec/"
   ```

   Any hit that **defines** (not references) a media-type string outside `spec/`
   or `openapi/yaagents-components.yaml` is a conformance violation. References
   (import paths, `$ref` strings, `Content-Type` header values) are permitted.

8. **Golden corpus** — every valid fixture in the versioned example corpus
   (`spec/examples/{version}/*.valid.json`) passes its corresponding schema in
   `schemas/{version}/`; every invalid fixture fails. This corpus is the
   cross-component contract test (ADR PI1-yaa-0002 §5). The frozen backward-compat
   corpus validates against `schemas/v0.1/`.

---

## 9. Versioning Policy

- This document is versioned at `spec/VERSION` = `0.2`.
- Profile bumps produce a sibling `schemas/vX.Y/` path containing updated schemas with
  bumped `$id` values; the prior version's schemas are frozen and never deleted.
- This document is updated in-place to the current canonical version; prior version
  history is preserved in git. Frozen schemas: v0.2 at `schemas/v0.2/`; v0.1 at
  `schemas/v0.1/`.
- A profile bump requires an ADR under `docs/adr/` recording the delta, migration
  path, and affected components.

---

## 10. Authoritative Source Rule (ADR PI1-yaa-0002 §1)

`spec/` is the **single authoritative source** of the Response Profile prose and
normative table. Every other component references it; **none redefine or paraphrase**
the §4 table. This rule is enforced at:

- PR review (manual) and the EX-4 grep gate (§8 criterion 7).
- `yaagents-cli validate --conformance` output references this file on failure.
- The golden corpus under `spec/examples/` is owned by this component (WI-1yaa.SPEC-5);
  fixtures are organised by profile version subdirectory.

---

## 11. SSE Streaming Contract (normative since v0.2)

SSE streaming through the plugin chain is a normative gateway feature first introduced in
Profile v0.2 (PRD §5.1, §7 LLM Gateway; WI-2yaa.BUMP-3); carried forward unchanged in v0.3.

### 11.1 Activation

SSE streaming is **config-driven**. A route opts in by declaring `mode: sse` in
`routes.yaml`. Routes without `mode: sse` are unaffected; the SSE code path is dormant.

```yaml
routes:
  - id: completions
    method: POST
    path: /llm/completions
    target: http://upstream:8080
    mode: sse
```

### 11.2 Request matching

The gateway activates SSE pipe-and-flush when **both** conditions hold:

1. The matched route declares `mode: sse`.
2. The inbound request carries `Accept: text/event-stream`.

When only condition 1 holds (no `Accept: text/event-stream`), the route falls back to
the standard JSON reverse-proxy path.

### 11.3 Response contract

| Property | Value |
|---|---|
| `Content-Type` | `text/event-stream` |
| `X-YAAgents-Profile` | `v0.3` (mandatory — same as all agentic responses) |
| `Cache-Control` | `no-cache` |
| Flush semantics | Per upstream chunk; gateway does not buffer |

The plugin chain (token-validator, tenant-injector, etc.) runs **before** the SSE proxy
path executes. `X-YAAgents-Profile: v0.3` is injected by the profile-header middleware
on the response before any chunk is flushed to the client.

### 11.4 Error responses on SSE routes

Error responses on SSE routes use the standard `application/vnd.yaagents.error+json`
body (§7.10). Two SSE-specific error codes are defined:

| Condition | HTTP Status | `code` field |
|---|---|---|
| Per-tenant concurrency limit exceeded | 429 | `LIMIT_EXCEEDED` |
| Execution timeout | 500 | `EXECUTION_TIMEOUT` |

The `retryAfter` field (integer seconds) SHOULD be set on 429 responses (`retryAfter: 60`
is the gateway default).

### 11.5 Per-tenant concurrency

The gateway enforces a per-tenant SSE concurrency ceiling configured via
`gateway.llm.max_sse_connections_per_tenant` (default: 10). The tenant is identified
from the `X-Tenant-ID` header injected by the `tenant-injector` plugin.

### 11.6 Execution timeout

Per-route field `executionTimeoutSeconds` (integer, default 0 = no timeout). When set:

- SSE routes: `context.WithDeadline(ctx, now + N*s + 30s)` — the 30 s surplus is the
  SSE read-budget (PRD §7.1).
- Non-SSE routes: `context.WithDeadline(ctx, now + N*s)`.

On deadline exceeded the gateway returns `500 application/vnd.yaagents.error+json` with
`code: EXECUTION_TIMEOUT`.

### 11.7 Metrics

Profile v0.3 adds two SSE-specific Prometheus metrics on `/metrics`:

| Metric | Type | Labels |
|---|---|---|
| `yaagents_gateway_sse_connections_active` | Gauge | `tenant_id`, `route_id` |
| `yaagents_gateway_sse_errors_total` | Counter | `tenant_id`, `route_id`, `error_kind` |

`error_kind` ∈ `client_disconnect` / `upstream_error` / `timeout` / `limit_exceeded`.

---

## Changelog

### v0.3 changes (2026-06-02)

`sdk-go` server SDK component added (PRD §5.10); no schema body changes; `Supports-YAAgents-Profile`
header value bumps to `v0.3`; `X-YAAgents-Profile: v0.3` response header MUST be added by every
component. Profile version file `spec/VERSION` bumped from `0.2` to `0.3`. `schemas/v0.3/` created
as directory-prefix copy of `schemas/v0.2/` with `$id` paths updated; no body changes (PRD §5.2).

### v0.2 changes (2026-06-01)

SSE streaming contract added as normative gateway feature (§11); per-tenant concurrency +
execution timeout; `schemas/v0.2/` created with 6 updated schema files; `X-YAAgents-Profile: v0.2`
header mandated on SSE responses; `go-client` SDK component added.
