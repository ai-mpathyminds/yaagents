import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import react from '@astrojs/react';

export default defineConfig({
  site: 'https://ai-mpathyminds.github.io',
  base: '/yaagents',
  integrations: [
    starlight({
      title: 'yaagents',
      description: 'A source-available REST profile + gateway + SDKs that exposes agent operations as typed resource endpoints.',
      social: {
        github: 'https://github.com/ai-mpathyminds/yaagents',
      },
      customCss: ['./src/styles/aimpathy-tokens.css'],
      sidebar: [
        {
          label: '\u{1F393} Tutorials',
          items: [
            { label: 'Quick Start (10 min)', link: '/tutorials/quick-start/' },
            { label: 'Your first Python agent', link: '/tutorials/python-first-agent/' },
            { label: 'Your first Go agent', link: '/tutorials/go-first-agent/' },
            { label: 'Calling agents from TypeScript', link: '/tutorials/ts-client/' },
          ],
        },
        {
          label: '\u{1F3AF} How-to Guides',
          items: [
            { label: 'Configure tenant injection', link: '/how-to/tenant-injection/' },
            { label: 'Write a custom JWT validator plugin', link: '/how-to/custom-jwt-validator/' },
            { label: 'Wire OTEL audit emission', link: '/how-to/otel-audit/' },
            { label: 'Add a community plugin', link: '/how-to/add-community-plugin/' },
            { label: 'Deploy on Kubernetes', link: '/how-to/deploy-k8s/' },
            { label: 'Replace a first-party plugin', link: '/how-to/replace-plugin/' },
            { label: 'Run the LLM Gateway example', link: '/how-to/llm-gateway-example/' },
          ],
        },
        {
          label: '\u{1F4DA} Reference',
          items: [
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
          label: '\u{1F9E0} Explanation',
          items: [
            { label: 'Why yaagents?', link: '/explanation/why-yaagents/' },
            { label: 'Resource-oriented vs operation-oriented', link: '/explanation/resource-vs-operation/' },
            { label: 'The plugin model', link: '/explanation/plugin-model/' },
            { label: 'Default-plugin design principle', link: '/explanation/default-plugin-principle/' },
            { label: 'Versioning & profile evolution', link: '/explanation/versioning/' },
            { label: 'Comparison: vs LangChain / Anthropic / OpenAI', link: '/explanation/comparison/' },
          ],
        },
        {
          label: 'Community',
          items: [
            { label: 'GitHub Discussions', link: 'https://github.com/ai-mpathyminds/yaagents/discussions' },
            { label: 'Contributing', link: '/community/contributing/' },
            { label: 'Code of Conduct', link: '/community/code-of-conduct/' },
            { label: 'Security', link: '/community/security/' },
          ],
        },
        {
          label: 'About',
          items: [
            { label: 'Roadmap', link: '/about/roadmap/' },
            { label: 'Releases', link: '/about/releases/' },
            { label: 'Changelog', link: '/about/changelog/' },
          ],
        },
      ],
    }),
    react(),
  ],
});
