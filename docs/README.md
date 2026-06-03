# YAAgents Docs

Static documentation site for [YAAgents](https://github.com/ai-mpathyminds/yaagents) — the
source-available Agentic REST Profile. Built with [Astro Starlight](https://starlight.astro.build/).

## Local development

```bash
pnpm install
pnpm dev
# → http://localhost:4321/yaagents/
```

## Build

```bash
pnpm build
# output in dist/ (mirrors production at https://ai-mpathyminds.github.io/yaagents/)
```

## Preview production build

```bash
pnpm preview
# → http://localhost:4321/yaagents/
```

## Hosting

GitHub Pages at `https://ai-mpathyminds.github.io/yaagents/` via
`.github/workflows/pages.yml` (ADR PI3-yaa-0004).

Custom domain (`yaagents.dev`) is deferred post-traction — see ADR PI3-yaa-0004 §Migration.
