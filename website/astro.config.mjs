// @ts-check
import { defineConfig, passthroughImageService } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://loy-go.github.io',
  base: '/loy',
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
            { label: 'Why Loy? (Comparison)', slug: 'start/comparison' },
            { label: 'Prerequisites & Doctor', slug: 'start/prerequisites' },
          ],
        },
        {
          label: 'Core Concepts',
          items: [
            { label: 'Layered Architecture', slug: 'concepts/architecture' },
            { label: 'The Layer Matrix', slug: 'concepts/layer-matrix' },
            { label: 'Explicit Constructor Wiring', slug: 'concepts/explicit-wiring' },
            { label: 'Zero Runtime Invariant', slug: 'concepts/zero-runtime' },
            { label: 'Plan-Based Code Generation', slug: 'concepts/plan-based-generation' },
          ],
        },
        {
          label: 'Agentic & AI Platform',
          items: [
            { label: 'The Agent-Native Platform', slug: 'ai' },
            { label: 'Model Context Protocol (MCP)', slug: 'ai/mcp-server' },
            { label: 'Self-Healing Agent Loops', slug: 'ai/self-healing' },
            { label: 'Universal Agent Rules', slug: 'ai/rules-generator' },
            { label: 'Official VS Code Extension', slug: 'ai/vscode-extension' },
          ],
        },
        {
          label: 'Component Best Practices',
          items: [
            { label: 'Best Practices Overview', slug: 'best-practices' },
            { label: 'Domain Models & Entities', slug: 'best-practices/domain-and-models' },
            { label: 'Repositories & Persistence', slug: 'best-practices/repositories-and-persistence' },
            { label: 'Services & Use Cases', slug: 'best-practices/services-and-usecases' },
            { label: 'Handlers & Transports', slug: 'best-practices/handlers-and-transports' },
            { label: 'Request & Resource DTOs', slug: 'best-practices/dtos-and-validation' },
            { label: 'Background Jobs & Workers', slug: 'best-practices/background-jobs-and-workers' },
            { label: 'Domain Events & Listeners', slug: 'best-practices/events-and-listeners' },
            { label: 'Authorization Policies & RBAC', slug: 'best-practices/policies-and-security' },
            { label: 'Database Migrations & Seeders', slug: 'best-practices/migrations-and-seeding' },
            { label: 'WebSockets & Streaming', slug: 'best-practices/websockets-and-streaming' },
            { label: 'gRPC & Protocol Buffers', slug: 'best-practices/grpc-and-protobuf' },
            { label: 'Transactional Outbox', slug: 'best-practices/transactional-outbox' },
            { label: 'Composition Root Wiring', slug: 'best-practices/composition-root-wiring' },
            { label: 'Architecture Enforcement in CI', slug: 'best-practices/architecture-enforcement' },
          ],
        },
        {
          label: 'Recipes & Cookbooks',
          items: [
            { label: 'Overview & Catalog', slug: 'recipes' },
            { label: 'Multi-Tenant PostgreSQL RLS', slug: 'recipes/multi-tenant-rls' },
            { label: 'Background Worker Fleet', slug: 'recipes/background-workers' },
            { label: 'Real-Time WebSocket Streaming', slug: 'recipes/websocket-streaming' },
            { label: 'gRPC & REST Dual-Listener', slug: 'recipes/grpc-microservices' },
            { label: 'Transactional Outbox Pattern', slug: 'recipes/transactional-outbox' },
            { label: 'JWT Auth & Role-Based Access', slug: 'recipes/auth-rbac' },
            { label: 'Zero-Docker Apps with SQLite', slug: 'recipes/zero-docker-sqlite' },
          ],
        },
        {
          label: 'Guides & Tutorials',
          items: [
            { label: 'Building Vertical CRUD Slices', slug: 'guides/vertical-slice' },
            { label: 'Database Migrations & SQLC', slug: 'guides/database-migrations' },
            { label: 'CQRS & Commands', slug: 'guides/cqrs-and-commands' },
            { label: 'Fullstack Typegen (TypeScript)', slug: 'guides/typegen-client' },
            { label: 'Schema-First Ingestion', slug: 'guides/schema-ingestion' },
            { label: 'Template Overrides', slug: 'guides/template-overrides' },
            { label: 'Live Hot Reload Supervisor', slug: 'guides/live-reload' },
            { label: 'Interactive Live TUI', slug: 'guides/interactive-tui' },
            { label: 'Architecture Drift Diffing', slug: 'guides/architecture-drift' },
            { label: 'Prometheus & Grafana Metrics', slug: 'guides/observability-metrics' },
            { label: 'Community Plugins', slug: 'guides/plugin-system' },
            { label: 'Production Readiness Guide', slug: 'guides/production-readiness' },
            { label: 'Docker, K8s & Deployment', slug: 'guides/deployment' },
            { label: 'HTTP Engine Benchmarks', slug: 'guides/http-benchmarks' },
            { label: 'Migrating from Gin', slug: 'guides/migrating-from-gin' },
            { label: 'Migrating from Laravel', slug: 'guides/migrating-from-laravel' },
          ],
        },
        {
          label: 'Case Studies',
          items: [
            { label: 'Building Intivai Enterprise SaaS', slug: 'case-studies/enterprise-saas' },
          ],
        },
        {
          label: 'Architecture Rules',
          items: [
            { label: 'Overview & loy check', slug: 'rules' },
            { label: 'ARCH-001 Dependency Cycles', slug: 'rules/arch001-cycles' },
            { label: 'ARCH-002–006 Layer Boundaries', slug: 'rules/arch002-006-layers' },
            { label: 'ARCH-007–010 Import Restrictions', slug: 'rules/arch007-010-imports' },
            { label: 'ARCH-011–015 Governance & Purity', slug: 'rules/arch011-014-governance' },
          ],
        },
        {
          label: 'Extending Loy',
          items: [
            { label: 'Custom Rules & Templates', slug: 'extending/custom-rules' },
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
