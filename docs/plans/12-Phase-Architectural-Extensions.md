# Phase 12: Architectural Extensions & Fullstack Evolution Implementation Plan

**Phase:** 12  
**Status:** Completed  
**Estimated Scope:** Template overrides, manifest-driven DAG validator, advanced DB topologies, CQRS scaffolding, schema ingestion, fullstack typegen, semantic MCP AST mutations, test gating  
**Primary Specifications:** [`docs/00-INDEX.md`](../00-INDEX.md), [`ADR-003`](../adrs/ADR-003-explicit-wiring.md), [`ADR-013`](../adrs/ADR-013-codegen-engine-selection.md), [`ADR-015`](../adrs/ADR-015-database-migration-and-sqlc-pipeline.md), [`ADR-016`](../adrs/ADR-016-two-phase-architecture-enforcement-engine.md)

---

## 1. Goal & Objectives

Extend the Loy platform across 7 key architectural extensions while enforcing strict documentation and verification gates:
1. **Template Override Engine**: Enable user-defined templates in `.loy/templates/` with cascade resolution and `loy make template eject`.
2. **Manifest-Driven Architectural DAG**: Configurable layer matrix in `loy.yaml` (`architecture.layers` with `allows` and `match` globs) and pattern support (`ddd`, `hexagonal`, `cqrs`, `custom`).
3. **Advanced DB Topologies & Safe DDL**: Tri-route PostgreSQL connection pooling (DataPool, SessionPool, DirectConn), dual-identifier schemas (BIGINT + UUIDv7), and zero-downtime migration recipes (`--recipe=raw|index-concurrent|shadow-column`).
4. **CQRS & Event-Driven Scaffolding**: Segregated read/write pipelines (`loy make command`, `loy make query`) and transactional deduplication (`loy make idempotency`).
5. **Schema-First Ingestion**: Scaffold complete Clean Architecture slices from OpenAPI 3.x/JSON Schema (`loy make from-spec`) or live PostgreSQL / DDL (`loy make from-db`).
6. **Fullstack Typegen**: Zero-dependency TypeScript interfaces (`types.ts`) and typed fetch client SDK (`client.ts`) generated from Go transport DTOs (`loy gen client`).
7. **Semantic MCP AST Mutations & Self-Healing**: Concrete syntax tree (`dst`) manipulations (`loy_ast_insert_field`, `loy_ast_add_route`, `loy_ast_bind_dependency`), in-memory sandbox preview (`loy_plan_preview`), and automated ARCH-005 remediation (`loy check --fix`).
8. **Documentation Synchronization Gate**: Continuous gating via `TestDocumentationSyncGate` in `cmd/docgen` and `make check-docs` ensuring code additions immediately fail CI unless documentation is synced.

---

## 2. Implemented Architecture

```text
internal/
├── architecture/
│   ├── classifier.go             # Pattern and custom glob matching
│   ├── layer.go                  # Topology DAG and allowed edge matrix
│   └── rules/arch007_010_imports.go # Configurable topology evaluation
├── astmod/
│   ├── astmod.go                 # DST-based struct field, route, and dependency insertions
│   ├── fix.go                    # Auto-remediation of ARCH-005 violations
│   └── astmod_test.go            # Fuzz tests and mutation verifications
├── cli/
│   ├── check.go                  # loy check --fix and configurable topology
│   ├── gen_client.go             # loy gen client TypeScript generator
│   └── make.go                   # make template, command, query, idempotency, from-spec, from-db
├── ingest/
│   ├── db.go                     # PostgreSQL information_schema & DDL reverse-engineering
│   ├── spec.go                   # OpenAPI 3.x and JSON Schema parsing
│   └── ingest_test.go            # Unit and contract tests (>= 80% coverage)
├── mcp/
│   └── default_tools.go          # loy_ast_*, loy_plan_preview tools
├── process/
│   └── noop_runner.go            # NoopRunner for ultra-fast, zero-overhead test execution
└── typegen/typescript/
    ├── parser.go                 # AST parser for Go request/resource DTOs
    ├── generator.go              # types.ts and client.ts code generator
    └── generator_test.go         # Type mapping and compiler verification tests
```

---

## 3. Definition of Done Checklist

- [x] **Milestone 1**: Template override cascade (`.loy/templates/`) and `loy make template eject`.
- [x] **Milestone 2**: Configurable architectural DAG in `loy check` via `loy.yaml`.
- [x] **Milestone 3**: Tri-route DB pools, dual identifier support (`--dual-id`), and zero-downtime migration recipes (`--recipe`).
- [x] **Milestone 4**: CQRS command, query, and idempotency generators.
- [x] **Milestone 5**: Schema ingestion from OpenAPI specs and PostgreSQL tables (`internal/ingest`).
- [x] **Milestone 6**: Fullstack TypeScript client generator (`internal/typegen/typescript`).
- [x] **Milestone 7**: Semantic AST mutations using `dst` (`internal/astmod`) and MCP tools.
- [x] **Documentation Sync Gate**: Automated verification gate in `cmd/docgen/main_test.go` and `make check-docs`.
- [x] **Verification**: All release gate checks pass (`make check`, `make check-docs`, `make test-all`).
