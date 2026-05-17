---
name: secret-scanner
description: Scan recent commits across all product repos for likely secrets — API keys, tokens, private keys, embedded passwords, cloud credentials. Reports findings; does not auto-remediate.
---

# secret-scanner

Pattern-based scan. Run before any governance veto escalation touching a commit.

## Patterns (regex, case-insensitive unless noted)

| Category | Pattern (illustrative) |
|----------|------------------------|
| AWS access key | `AKIA[0-9A-Z]{16}` (case-sensitive) |
| AWS secret | `aws(.{0,20})?(secret\|access).{0,20}?[=:]\s*['"]?[0-9a-zA-Z/+]{40}['"]?` |
| Azure storage key | `DefaultEndpointsProtocol=.*AccountKey=[A-Za-z0-9+/=]{40,}` |
| Generic private key | `-----BEGIN (RSA\|OPENSSH\|EC\|DSA\|PGP) PRIVATE KEY-----` |
| JWT (long) | `eyJ[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}` |
| Slack token | `xox[abpr]-[0-9]{10,}-[0-9a-zA-Z]{24,}` |
| GitHub token | `ghp_[A-Za-z0-9]{36}\|github_pat_[A-Za-z0-9_]{50,}` |
| Generic password in URI | `[a-z]+://[^:]+:[^@]{6,}@` |
| Anthropic key | `sk-ant-[A-Za-z0-9_-]{80,}` (case-sensitive) |
| OpenAI key | `sk-[A-Za-z0-9]{48,}` |

False positives are expected — every hit must be human-reviewed before a veto is raised.

## Scan scope

- Diffs in commits since the last governance pass (per product)
- New files added during the scan window
- `.env*` files that were accidentally committed (any content triggers)
- Roadmap files and PRDs (lower-signal, but included because agents sometimes paste examples)

## Scope exclusions

- `node_modules/`, `.venv/`, `dist/`, `build/` — never scan
- Any path listed in a `.gitignore` — respect it
- Test fixtures that explicitly mark themselves `// TEST FIXTURE — SAFE` on a preceding line

## Output format

```markdown
## Secret scan — <product> — <window>
**Commits scanned**: <n>
**Findings**: <n>

| Commit | File:Line | Pattern | Snippet (redacted) | Likely? |
|--------|-----------|---------|--------------------|---------|
```

"Likely?" is one of: `YES`, `NO`, `NEEDS-REVIEW`. Default `NEEDS-REVIEW` — never auto-classify as `NO` without context.

## What this skill does NOT do

- Does not commit changes.
- Does not remediate (rotate keys, rewrite history, etc.). That is a human decision with compliance implications.
- Does not scan outside the portfolio directory.
- Does not cache results between runs — scans are idempotent.
