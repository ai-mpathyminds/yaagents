# yaagents-cli

**Supports-YAAgents-Profile: v0.1**

CLI validator for the [YAAgents Agentic REST Profile](https://github.com/ai-mpathyminds/yaagents).

## Install

```bash
pip install yaagents-cli
```

## Usage

### `validate-response`

Validate a response body JSON file against the v0.1 schema. The media type is
inferred from the body's `type` discriminator field.

```bash
# Human-readable
yaagents validate-response response.json

# Machine-readable JSON
yaagents validate-response response.json --json
```

**Exit codes:** `0` = PASS, `1` = FAIL or error, `2` = bad usage.

**Supported `type` values:**

| `type` | Schema |
|--------|--------|
| `operation_accepted` | `operation-accepted.schema.json` |
| `clarification_required` | `clarification-required.schema.json` |
| `validation_failed` | `validation-failed.schema.json` |
| `approval_required` | `approval-required.schema.json` |
| `conflict` | `conflict.schema.json` |
| `forbidden` / `failed_dependency` / `error` | `agentic-error.schema.json` |

## License

Source-available. See [LICENSE](../LICENSE).
