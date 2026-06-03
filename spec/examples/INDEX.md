# YAAgents Conformance Corpus — Index

**Current profile version:** v0.3
**Frozen backward-compat corpus:** `spec/examples/v0.2/` — validates against `schemas/v0.2/`; `spec/examples/v0.1/` — validates against `schemas/v0.1/`
**Authority:** §5 — single shared conformance oracle.
SDK, both clients, and CLI all test against this corpus; prevents per-component drift.

Valid fixtures must **pass** their schema. Invalid fixtures must **fail** their schema.

---

## `application/vnd.yaagents.clarification+json` (HTTP 400)

Schema: `schemas/v0.3/clarification-required.schema.json` (frozen: `schemas/v0.2/clarification-required.schema.json`)

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `clarification-required.valid.json` | PASS | Minimal valid; §4.1 body shape verbatim; single requiredInput with allowedValues |
| `clarification-required.valid.multi-input.json` | PASS | Three requiredInputs including optional fields (no allowedValues) |
| `clarification-required.invalid.missing-trace.json` | FAIL | `trace` object absent — violates mandatory trace contract |
| `clarification-required.invalid.wrong-type.json` | FAIL | `type` is `"validation_failed"` — violates `const: "clarification_required"` |
| `clarification-required.invalid.empty-inputs.json` | FAIL | `requiredInputs: []` — violates `minItems: 1` |

---

## `application/vnd.yaagents.validation-error+json` (HTTP 422)

Schema: `schemas/v0.3/validation-failed.schema.json` (frozen: `schemas/v0.2/validation-failed.schema.json`)

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `validation-failed.valid.json` | PASS | Single field error in `errors[]` |
| `validation-failed.valid.multi-error.json` | PASS | Multiple field errors; `errors` is optional so absence is also valid |
| `validation-failed.invalid.missing-trace.json` | FAIL | `trace` object absent |
| `validation-failed.invalid.wrong-type.json` | FAIL | `type` is `"clarification_required"` — violates `const: "validation_failed"` |
| `validation-failed.invalid.missing-message.json` | FAIL | Required field `message` absent |

---

## `application/vnd.yaagents.approval-required+json` (HTTP 412)

Schema: `schemas/v0.3/approval-required.schema.json` (frozen: `schemas/v0.2/approval-required.schema.json`)

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `approval-required.valid.json` | PASS | Standard approval with short token |
| `approval-required.valid.long-token.json` | PASS | JWT-style approval token |
| `approval-required.invalid.missing-trace.json` | FAIL | `trace` object absent |
| `approval-required.invalid.wrong-type.json` | FAIL | `type` is `"conflict"` — violates `const: "approval_required"` |
| `approval-required.invalid.missing-approval-token.json` | FAIL | Required field `approvalToken` absent |

---

## `application/vnd.yaagents.conflict+json` (HTTP 409)

Schema: `schemas/v0.3/conflict.schema.json` (frozen: `schemas/v0.2/conflict.schema.json`)

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `conflict.valid.json` | PASS | With optional `conflictingResourceId` |
| `conflict.valid.no-resource-id.json` | PASS | Without optional `conflictingResourceId` |
| `conflict.invalid.missing-trace.json` | FAIL | `trace` object absent |
| `conflict.invalid.wrong-type.json` | FAIL | `type` is `"error"` — violates `const: "conflict"` |
| `conflict.invalid.missing-code.json` | FAIL | Required field `code` absent |

---

## `application/vnd.yaagents.error+json` (HTTP 403 / 424 / 500)

Schema: `schemas/v0.3/agentic-error.schema.json` (frozen: `schemas/v0.2/agentic-error.schema.json`)

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `agentic-error.valid.forbidden.json` | PASS | `type: "forbidden"` (HTTP 403) |
| `agentic-error.valid.failed-dependency.json` | PASS | `type: "failed_dependency"` (HTTP 424) |
| `agentic-error.valid.error.json` | PASS | `type: "error"` (HTTP 500) |
| `agentic-error.invalid.missing-trace.json` | FAIL | `trace` object absent |
| `agentic-error.invalid.wrong-type.json` | FAIL | `type` is `"not_an_error"` — not in enum |
| `agentic-error.invalid.empty-code.json` | FAIL | `code: ""` — violates `minLength: 1` |

---

## `application/vnd.yaagents.operation+json` (HTTP 202)

Schema: `schemas/v0.3/operation-accepted.schema.json` (frozen: `schemas/v0.2/operation-accepted.schema.json`)

> **Scope note:** Schema and fixtures ship in v0.1;
> async polling runtime is v0.2 scope.

| Fixture | Verdict | Notes |
|---------|---------|-------|
| `operation-accepted.valid.json` | PASS | Relative `statusUrl` |
| `operation-accepted.valid.absolute-url.json` | PASS | Absolute `statusUrl` |
| `operation-accepted.invalid.missing-trace.json` | FAIL | `trace` object absent |
| `operation-accepted.invalid.wrong-type.json` | FAIL | `type` is `"accepted"` — violates `const: "operation_accepted"` |
| `operation-accepted.invalid.missing-operation-id.json` | FAIL | Required field `operationId` absent |

---

## Totals

| Category | Count |
|----------|-------|
| Valid fixtures | 13 |
| Invalid fixtures | 18 |
| **Total** | **31** |

---

## How consumers use this corpus

```bash
# CLI validator (*) — corpus path is the version subdirectory
yaagents-cli validate-corpus spec/examples/v0.1/   # frozen backward-compat corpus

# Python (pytest, *)
pytest tests/conformance/ --corpus spec/examples/v0.1/

# TypeScript (jest, *)
npx jest --testPathPattern conformance
```

Each test runner loads the INDEX via path convention (`*.valid.json` must pass,
`*.invalid.json` must fail) and the schema map above. No per-component fixture copies.
The corpus directory name matches the schema version it validates against.
