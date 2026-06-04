import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import react from '@astrojs/react';
import mermaid from 'astro-mermaid';

export default defineConfig({
  site: 'https://ai-mpathyminds.github.io',
  base: '/yaagents',
  integrations: [
    starlight({
      title: 'yaagents — Governed REST APIs for production AI agents',
      description: 'Expose AI agents as typed, governed, OpenAPI-friendly REST APIs. Use any agent framework internally; keep RBAC, tenancy, audit, SDKs, and gateway discipline at the application boundary.',
      social: {
        github: 'https://github.com/ai-mpathyminds/yaagents',
      },
      customCss: ['./src/styles/aimpathy-tokens.css'],
      sidebar: [
        {
          label: 'Start Here',
          items: [
            { label: 'Overview', link: '/start-here/overview/' },
            { label: 'Why yaagents?', link: '/explanation/why-yaagents/' },
            { label: 'Where yaagents fits', link: '/start-here/where-yaagents-fits/' },
            { label: 'yaagents vs A2A / AGNTCY / MCP / frameworks', link: '/explanation/yaagents-vs-a2a-agntcy-mcp/' },
            { label: 'Production Agent API Checklist', link: '/explanation/production-agent-api-checklist/' },
          ],
        },
        {
          label: 'Tutorials',
          items: [
            { label: 'Quick Start (10 min)', link: '/tutorials/quick-start/' },
            { label: 'Build an Agentic REST API in 30 Min', link: '/tutorials/build-an-agentic-rest-api-in-30-min/' },
          ],
        },
        {
          label: 'Patterns',
          items: [
            { label: 'Agentic REST endpoint design', link: '/explanation/agentic-rest-patterns/' },
            { label: 'Resource vs operation', link: '/explanation/resource-vs-operation/' },
            { label: 'Typed outcome model', link: '/patterns/typed-outcome-model/' },
            { label: 'Typed outcome decision flow', link: '/patterns/typed-outcome-decision-flow/' },
          ],
        },
        {
          label: 'Architecture',
          items: [
            { label: 'Reference architecture', link: '/architecture/reference-architecture/' },
            { label: 'Gateway request lifecycle', link: '/architecture/gateway-request-lifecycle/' },
            { label: 'Plugin pipeline', link: '/architecture/plugin-pipeline/' },
          ],
        },
        {
          label: 'Examples',
          items: [
            { label: 'Examples overview', link: '/examples/overview/' },
            {
              label: 'store',
              items: [
                { label: 'store (Python)', link: '/examples/store/' },
                { label: 'store-go (Go)', link: '/examples/store-go/' },
              ],
            },
            {
              label: 'campaign-optimizer',
              items: [
                { label: 'campaign-api (Python)', link: '/examples/campaign-api/' },
                { label: 'campaign-api-go (Go)', link: '/examples/campaign-api-go/' },
              ],
            },
            { label: 'LLM gateway', link: '/examples/llm-gateway/' },
          ],
        },
        {
          label: 'How-to Guides',
          items: [
            { label: 'Planned guides', link: '/how-to/planned-guides/' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Reference overview', link: '/reference/overview/' },
            { label: 'Profile v0.3 (full spec)', link: '/reference/profile-v03/' },
            { label: 'Gateway config', link: '/reference/gateway-config/' },
            { label: 'Plugin interface', link: '/reference/plugin-interface/' },
            {
              label: 'First-party plugins',
              items: [
                { label: 'token-validator', link: '/reference/plugins/token-validator/' },
                { label: 'tenant-injector', link: '/reference/plugins/tenant-injector/' },
                { label: 'license-check', link: '/reference/plugins/license-check/' },
                { label: 'prompt-sanitize', link: '/reference/plugins/prompt-sanitize/' },
                { label: 'otel-audit', link: '/reference/plugins/otel-audit/' },
              ],
            },
            {
              label: 'Server SDKs',
              items: [
                { label: 'sdk-fastapi (Python)', link: '/reference/sdk-fastapi/' },
                { label: 'sdk-go (Go)', link: '/reference/sdk-go/' },
              ],
            },
            {
              label: 'Client SDKs',
              items: [
                { label: 'yaagents-client (Python)', link: '/reference/client-python/' },
                { label: '@aimpathyminds/yaagents-client (TS)', link: '/reference/client-ts/' },
                { label: 'yaagents-client-go (Go)', link: '/reference/client-go/' },
              ],
            },
            { label: 'CLI commands', link: '/reference/cli/' },
            { label: 'OpenAPI components', link: '/reference/openapi/' },
            { label: 'JSON Schemas', link: '/reference/schemas/' },
          ],
        },
        {
          label: 'Community',
          items: [
            { label: 'GitHub Discussions', link: 'https://github.com/ai-mpathyminds/yaagents/discussions' },
            { label: 'Contributing', link: '/community/contributing/' },
            { label: 'Code of Conduct', link: '/community/code-of-conduct/' },
            { label: 'Security', link: '/community/security/' },
            { label: 'Roadmap', link: '/about/roadmap/' },
          ],
        },
      ],
    }),
    react(),
    mermaid(),
  ],
});
