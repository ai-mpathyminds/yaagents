# AI-Tool Skills — Extend the store example with a real LLM

These skill files contain starter prompts for extending the mock store recommender
into a real LLM-driven agent, one for each popular AI coding tool.

Pick the file that matches your tool:

| File | Tool |
|------|------|
| [`claude-code-agent.md`](claude-code-agent.md) | **Claude Code** (CLI agent) |
| [`claude-web.md`](claude-web.md) | **Claude.ai** (web or desktop) |
| [`github-copilot.md`](github-copilot.md) | **GitHub Copilot** (VS Code / JetBrains) |
| [`openai-codex.md`](openai-codex.md) | **OpenAI / ChatGPT** (web or API) |
| [`cursor.md`](cursor.md) | **Cursor** IDE |
| [`generic-prompt.md`](generic-prompt.md) | Any other AI coding tool |

---

## What the skill does

Each skill gives your AI tool enough context to:

1. Understand the YAAgents Profile v0.3 contract (what response shapes are required)
2. Understand the existing mock implementation in `src/store/recommender.py`
3. Replace the mock with a real LLM call — either via the Anthropic API, OpenAI API,
   or any LLM you have access to
4. Keep the Profile contract intact (correct response shape, headers, status codes)

Estimated time: **10–20 minutes** per tool.

---

## What the mock does today

`src/store/recommender.py → mock_recommend()` returns up to `limit` products
from the same category as the seed product. No LLM is involved.

The Profile shape is already correct — the skill helps you replace the *logic*
while keeping the *contract*.
