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
        {
          label: 'Start Here',
          items: [
            { label: 'Overview', link: '/start-here/overview/' },
            { label: 'Why YAAgents?', link: '/why/' },
          ],
        },
        { label: 'Quick Start', link: '/quickstart/' },
        { label: 'Profile Spec', link: '/profile/' },
        {
          label: 'Plugins',
          items: [
            { label: 'Plugin pipeline overview', link: '/plugins/' },
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
            { label: 'Agent implementation patterns', link: '/concepts/agent-implementation-patterns/' },
          ],
        },
        {
          label: 'How-to',
          items: [
            { label: 'Host in production', link: '/how-to/host-in-production/' },
          ],
        },
        { label: 'SDK Quickstarts', autogenerate: { directory: 'sdks' } },
        { label: 'Examples', autogenerate: { directory: 'examples' } },
        {
          label: 'Case Studies',
          items: [
            { label: 'E-commerce Recommendations', link: '/case-studies/ecommerce-product-recommendations/' },
          ],
        },
        {
          label: 'Architecture',
          items: [
            { label: 'Inter-Agent Calls', link: '/architecture/inter-agent-calls/' },
            { label: 'Audit and Observability', link: '/architecture/audit-and-observability/' },
          ],
        },
        { label: 'Public Roadmap', link: '/roadmap/' },
        { label: 'Contributing', link: '/contributing/' },
        { label: 'Community', link: '/community/' },
      ],
    }),
  ],
});
