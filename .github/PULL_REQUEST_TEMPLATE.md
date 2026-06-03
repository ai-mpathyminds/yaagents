## Linked issue

Closes #<!-- issue number -->

<!-- If there is no linked issue, explain why this change is self-contained. -->

---

## Summary

<!-- One paragraph describing what this PR does and why. -->

---

## Checklist

### DCO sign-off (required)

- [ ] Every commit in this PR carries `Signed-off-by: Your Name <your@email.com>`
      (DCO — see [CONTRIBUTING.md](CONTRIBUTING.md#developer-certificate-of-origin-dco)).
      PRs with unsigned commits cannot be merged.

### Change scope

- [ ] Tests added or updated (unit and/or integration).
- [ ] Existing tests still pass locally.
- [ ] Lint gates pass (`golangci-lint` / `ruff` / `eslint` as appropriate).

### API contract

- [ ] If HTTP contract changed: `openapi/` components updated.
- [ ] If new public Go symbols: godoc comments added.
- [ ] If new Python public API: docstrings added.
- [ ] If new TypeScript public API: JSDoc added.

### Changelog entry

- [ ] `CHANGELOG.md` has an entry under `[Unreleased]`:

```
### Added / Changed / Fixed / Removed
- <one-line description> (#<issue-number>)
```

### For plugin PRs

- [ ] Plugin implements the `Plugin` interface in `gateway/plugins/<name>/plugin.go`.
- [ ] Plugin registered via `init()`.
- [ ] YAML config documented in plugin README under `plugins.<name>:` key.
- [ ] Maintainer `go-ahead` comment exists on the linked Plugin Proposal issue.

---

## Testing notes

<!-- How did you verify this change? Include relevant curl commands, test
output excerpts, or Compose demo steps. -->
