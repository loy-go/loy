// @ts-check
import { defineConfig, passthroughImageService } from 'astro/config';
import starlight from '@astrojs/starlight';

const isGitHubPages = process.env.GITHUB_PAGES === 'true' || process.env.GITHUB_ACTIONS === 'true';

export default defineConfig({
  site: isGitHubPages ? 'https://loy-go.github.io' : 'https://loy.dev',
  base: isGitHubPages ? '/loy' : '/',
  image: {
    service: passthroughImageService(),
  },
  integrations: [
    starlight({
      title: 'Loy',
      description: 'Build-time Go developer platform with zero runtime framework lock-in',
      tagline: 'Typed scaffolding, explicit wiring, and architecture enforcement for ordinary Go',
      logo: {
        src: './src/assets/hero.svg',
      },
      social: {
        github: 'https://github.com/loy-go/loy',
      },
      customCss: [],
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Installation', slug: 'start/installation' },
            { label: 'Quickstart (60s)', slug: 'start/quickstart' },
            { label: 'Prerequisites & Doctor', slug: 'start/prerequisites' },
          ],
        },
        {
          label: 'Core Concepts',
          items: [
            { label: 'Layered Architecture', slug: 'concepts/architecture' },
            { label: 'Explicit Constructor Wiring', slug: 'concepts/explicit-wiring' },
            { label: 'Zero Runtime Invariant', slug: 'concepts/zero-runtime' },
            { label: 'Plan-Based Code Generation', slug: 'concepts/plan-based-generation' },
          ],
        },
        {
          label: 'Guides & Tutorials',
          items: [
            { label: 'Building Vertical CRUD Slices', slug: 'guides/vertical-slice' },
            { label: 'Database Migrations & SQLC', slug: 'guides/database-migrations' },
            { label: 'Live Hot Reload Supervisor', slug: 'guides/live-reload' },
            { label: 'Docker, K8s & Deployment', slug: 'guides/deployment' },
          ],
        },
        {
          label: 'Architecture Rules',
          items: [
            { label: 'Overview & loy check', slug: 'rules' },
            { label: 'ARCH-001 Dependency Cycles', slug: 'rules/arch001-cycles' },
            { label: 'ARCH-002–006 Layer Boundaries', slug: 'rules/arch002-006-layers' },
            { label: 'ARCH-007–010 Import Restrictions', slug: 'rules/arch007-010-imports' },
            { label: 'ARCH-011–014 Governance & State', slug: 'rules/arch011-014-governance' },
          ],
        },
        {
          label: 'CLI Reference',
          autogenerate: {
            directory: 'reference/cli',
          },
        },
        {
          label: 'Configuration',
          items: [
            { label: 'loy.yaml Manifest Reference', slug: 'reference/manifest' },
          ],
        },
        {
          label: 'Architecture Decisions (ADRs)',
          items: [
            { label: 'ADR Index & Rationale', slug: 'adrs' },
          ],
        },
      ],
    }),
  ],
});
