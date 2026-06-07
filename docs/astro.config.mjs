import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://ai-mpathyminds.github.io',
  base: '/yaagents',
  integrations: [
    starlight({
      title: 'YAAgents',
      description: 'Source-available Agentic REST Profile',
      social: {
        github: 'https://github.com/ai-mpathyminds/yaagents',
      },
      sidebar: [
        { label: 'Quick Start', link: '/quickstart/' },
        { label: 'Why YAAgents', link: '/why/' },
        { label: 'Profile Spec', link: '/profile/' },
        { label: 'SDK Quickstarts', autogenerate: { directory: 'sdks' } },
        { label: 'Examples', autogenerate: { directory: 'examples' } },
        {
          label: 'Plugins',
          items: [
            { label: 'Plugin Overview', link: '/plugins/' },
            { label: 'token-validator', link: '/plugins/token-validator/' },
            { label: 'tenant-injector', link: '/plugins/tenant-injector/' },
            { label: 'license-check', link: '/plugins/license-check/' },
            { label: 'prompt-sanitize', link: '/plugins/prompt-sanitize/' },
            { label: 'otel-audit', link: '/plugins/otel-audit/' },
          ],
        },
        {
          label: 'Concepts',
          items: [
            { label: 'Why Agentic REST?', link: '/concepts/why-agentic-rest/' },
            { label: 'Comparisons', link: '/concepts/comparisons/' },
          ],
        },
        {
          label: 'Architecture',
          items: [
            { label: 'Inter-Agent Calls', link: '/architecture/inter-agent-calls/' },
          ],
        },
        { label: 'Public Roadmap', link: '/roadmap/' },
        { label: 'Contributing', link: '/contributing/' },
        { label: 'Community', link: '/community/' },
      ],
    }),
  ],
});
