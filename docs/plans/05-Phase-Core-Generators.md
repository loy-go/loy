# Phase 5: Core Generators Implementation Plan

**Phase:** 5 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** Individual artifact generators, CRUD composer, wiring splicing  
**Primary Specifications:** [04-Generator-Specification.md](../04-Generator-Specification.md), [12-Code-Generation-Templates.md](../12-Code-Generation-Templates.md), [13-Architecture-Design-Patterns.md](../13-Architecture-Design-Patterns.md), [ADR-014](../adrs/ADR-014-managed-code-splicing-via-comment-regions.md)

---

## 1. Goal & Objectives
Implement the complete developer scaffolding toolkit (`loy make ...`):
- Atomic artifact generators: `model`, `repository`, `service`, `handler`, `request`, `resource`, `job`, `event`, `listener`, `policy`, `test`.
- Vertical slice composer: `loy make feature <name>` and `loy make crud <name>`.
- Auto-wiring into `internal/app/wiring.go` via managed comment regions (`// loy:region:...`).
- 100% deterministic, gofmt-compliant Go source output.

---

## 2. Package Architecture & Internal Seams

```text
generators/
├── base.go                 # Common generator utilities & context
├── model/
│   ├── model.go            # loy make model <name> [fields...]
│   └── template.go.tmpl    # Domain entity struct & validation methods
├── repository/
│   ├── repository.go       # loy make repository <name>
│   ├── interface.go.tmpl   # Domain interface
│   └── pg_adapter.go.tmpl  # PostgreSQL/sqlc infrastructure adapter
├── service/
│   ├── service.go          # loy make service <name>
│   └── template.go.tmpl    # Application service orchestrating repo/domain
├── handler/
│   ├── handler.go          # loy make handler <name>
│   └── fiber_tmpl.go.tmpl  # Fiber HTTP handler
├── request/
│   ├── request.go          # loy make request <name> (DTO + validation)
│   └── template.go.tmpl
├── resource/
│   ├── resource.go         # loy make resource <name> (Response transformation)
│   └── template.go.tmpl
├── job/
│   ├── job.go              # loy make job <name> (Asynq background task payload & handler)
│   └── template.go.tmpl
├── event/
│   ├── event.go            # loy make event <name>
│   ├── listener.go         # loy make listener <name>
│   └── template.go.tmpl
├── policy/
│   ├── policy.go           # loy make policy <name> (Authorization rule check)
│   └── template.go.tmpl
├── test/
│   ├── test.go             # loy make test <target> (Unit/Integration scaffold)
│   └── template.go.tmpl
├── feature/
│   └── feature.go          # loy make feature (bundles domain, service, handler)
└── crud/
    └── crud.go             # loy make crud (full vertical slice: migration + queries + repo + service + handler)
cmd/loy/
└── make.go                 # Wire `loy make` command tree
```

---

## 3. Concrete Implementation Steps

### Step 5.1: Atomic Artifact Generators
1. **Model Generator**:
   - Parses field arguments: `name:string email:string:unique status:int`.
   - Generates domain entity struct in `internal/<feature>/domain/model.go`.
2. **Repository Generator**:
   - Emits domain interface: `internal/<feature>/domain/repository.go`.
   - Emits infrastructure adapter: `internal/<feature>/repository/pg_adapter.go`.
3. **Service Generator**:
   - Emits application use case struct in `internal/<feature>/service/service.go`.
   - Injects repository interface via explicit constructor: `NewService(repo domain.Repository)`.
4. **Handler Generator**:
   - Emits HTTP transport handler in `internal/<feature>/transport/http/handler.go`.
   - Exposes `RegisterRoutes(router fiber.Router)` method.
5. **Request & Resource Generators**:
   - DTOs for request parsing/validation and response serialization.

### Step 5.2: Managed Region Wiring Splicing
1. When generating handlers or services, update `internal/app/wiring.go`:
   - Locate `// loy:region:repositories` -> append `repo := <feature>Repo.New(db)`.
   - Locate `// loy:region:services` -> append `svc := <feature>Service.New(repo)`.
   - Locate `// loy:region:handlers` -> append `h := <feature>Handler.New(svc)`.
   - Locate `// loy:region:routes` -> append `h.RegisterRoutes(v1Group)`.
2. Re-format `wiring.go` via `gofmt`.

### Step 5.3: Feature & CRUD Composers
1. `loy make feature <name>`:
   - Orchestrates Model, Service, Handler, and Test plans into a single consolidated `Plan`.
   - Slices wiring into `wiring.go`.
2. `loy make crud <name> [fields]`:
   - Extends feature by also scaffolding:
     - Goose SQL migration: `migrations/YYYYMMDDHHMMSS_create_<name>_table.sql`.
     - sqlc SQL queries: `queries/<name>.sql`.
     - Invokes `sqlc generate` check.

---

## 4. Test Strategy & Acceptance Criteria

### Golden Tests
- Golden test for every individual generator in `testdata/golden/generators/<name>`.
- Full vertical slice test: verify `loy make crud product` generates 7 files matching golden outputs.

### Unit Tests
- `field_parser_test.go`: Verify complex type mappings (`status:enum(active,inactive)`, `tags:[]string`).
- Splicing integration: Test running `loy make feature users` on an existing `wiring.go` preserves developer code and cleanly adds routes.

### Acceptance Test
- Generate a new project -> run `loy make crud users` -> compile with `go build ./...` -> verify zero compilation errors.

---

## 5. Definition of Done
- [ ] All 13 `make` subcommands implemented and tested.
- [ ] Plan composition works atomically without partial disk writes on error.
- [ ] `wiring.go` cleanly updated via region comment splicing.
- [ ] Generated code compiles under standard Go compiler.

---

[← Previous: Phase 4 Plan](./04-Phase-Architecture-Engine.md) | [Back to Plans Index](./README.md) | [Next: Phase 6 Plan →](./06-Phase-Runtime.md)
