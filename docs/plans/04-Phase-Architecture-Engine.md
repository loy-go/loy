# Phase 4: Architecture Engine Implementation Plan

**Phase:** 4 of 10  
**Status:** Completed  
**Estimated Scope:** Static analysis engine, two-phase analyzer, 14 architecture rules, suppression engine  
**Primary Specifications:** [05-Architecture-Rule-Specification.md](../05-Architecture-Rule-Specification.md), [13-Architecture-Design-Patterns.md](../13-Architecture-Design-Patterns.md), [ADR-011](../adrs/ADR-011-architecture-validation.md), [ADR-016](../adrs/ADR-016-two-phase-architecture-enforcement-engine.md)

---

## 1. Goal & Objectives
Deliver `loy check` to enforce compile-time architecture invariants:
- Two-phase analysis: Fast import AST graph (`go/parser`) + deep type analysis (`go/packages`).
- Enforce the 4-layer dependency matrix: `Transport -> Application -> Domain <- Infrastructure`.
- Detect cyclic package dependencies across modules and internal packages.
- Validate app-to-app boundaries in monorepo workspaces.
- Support structured suppressions via explicit code comments (`// loy:ignore ARCH-xxx`).

---

## 2. Package Architecture & Internal Seams

```text
internal/
├── graph/
│   ├── node.go             # PackageNode (ImportPath, Layer, FilePaths)
│   ├── edge.go             # DependencyEdge (From, To, File, Line)
│   ├── graph.go            # DirectedGraph representation
│   └── cycle.go            # Tarjan's strongly connected components cycle detector
├── architecture/
│   ├── layer.go            # Layer enum: Transport, Application, Domain, Infrastructure, Platform
│   ├── classifier.go       # LayerClassifier (maps package paths to architectural layers)
│   ├── analyzer.go         # Two-phase Analysis coordinator
│   ├── ast_phase.go        # Phase 1: Fast AST parser using go/parser and go/token
│   ├── type_phase.go       # Phase 2: Deep type inspection via golang.org/x/tools/go/packages
│   ├── violation.go        # RuleViolation model
│   ├── suppression.go      # AST comment parser checking `// loy:ignore ARCH-xxx reason="..."`
│   ├── rule.go             # Rule interface (ID, Description, Check(ctx, Analysis) []Violation)
│   └── rules/
│       ├── arch001_cycles.go
│       ├── arch002_domain_infra.go
│       ├── arch003_domain_transport.go
│       ├── arch004_app_transport.go
│       ├── arch005_app_concrete_infra.go
│       ├── arch006_infra_transport.go
│       ├── arch008_forbidden_imports.go
│       ├── arch010_layer_direction.go
│       └── arch013_workspace_boundaries.go
└── cli/
    └── check.go            # loy check CLI command (--deep flag, --json, exit code 1 on violations)
```

---

## 3. Concrete Implementation Steps

### Step 3.1: Graph Engine & Cycle Detection (`internal/graph`)
1. Implement `DirectedGraph`:
   - `AddNode(path string, data any)`
   - `AddEdge(from string, to string, file string, line int)`
2. Implement Tarjan's SCC algorithm to detect cycles. Return exact import chain causing cycle (e.g. `A -> B -> C -> A`).

### Step 3.2: Layer Classification (`internal/architecture/classifier.go`)
1. Classify packages based on directory conventions:
   - `internal/transport/...` or `internal/handler/...` -> `LayerTransport`
   - `internal/service/...` or `internal/usecase/...` -> `LayerApplication`
   - `internal/domain/...` or `internal/model/...` -> `LayerDomain`
   - `internal/platform/...`, `internal/repository/...`, `internal/database/...` -> `LayerInfrastructure`
2. Allow overrides via `loy.yaml` under `architecture.layers`.

### Step 3.3: Two-Phase AST Analyzer (`internal/architecture/analyzer.go`)
1. **Phase 1 (Fast AST)**:
   - Scan all `.go` files using `go/parser.ParseFile` with `parser.ImportsOnly`.
   - Extract imports, file paths, line numbers.
   - Build import graph and pass to layer rules.
2. **Phase 2 (Deep Type Inspection - activated with `--deep` or in CI)**:
   - Load packages with `packages.Load` mode `NeedTypes | NeedSyntax | NeedTypesInfo`.
   - Inspect concrete struct implementations and global mutable variables (ARCH-011, ARCH-012).

### Step 3.4: Rule Implementations
1. **ARCH-001**: Check cycles via Tarjan's algorithm.
2. **ARCH-002**: Flag any import from `Domain` into `Infrastructure`.
3. **ARCH-003**: Flag any import from `Domain` into `Transport`.
4. **ARCH-004**: Flag any import from `Application` into `Transport`.
5. **ARCH-005**: Flag `Application` importing concrete infrastructure adapters instead of domain interfaces.
6. **ARCH-006**: Flag `Infrastructure` importing `Transport`.
7. **ARCH-010**: Enforce Layer Matrix rules globally.
8. **ARCH-013**: Enforce `App -> App` import prohibition across workspace modules.

### Step 3.5: Suppression Engine (`internal/architecture/suppression.go`)
1. Parse AST comments matching pattern `// loy:ignore (ARCH-\d+) reason="(.+)"`.
2. Filter violations where suppression matches rule ID, file, and line.
3. Reject suppressions without an explicit non-empty `reason`.

---

## 4. Test Strategy & Acceptance Criteria

### Fixture-Based Architecture Tests
Create `testdata/fixtures/` containing known architectural scenarios:
1. `valid_clean_app/`: Zero violations, passes check.
2. `cycle_app/`: Package A imports B imports A -> flags `ARCH-001`.
3. `domain_leaking_infra/`: Domain imports sqlc package -> flags `ARCH-002`.
4. `app_importing_fiber/`: Application imports Fiber context -> flags `ARCH-004`.
5. `suppressed_violation/`: Valid `// loy:ignore ARCH-002 reason="..."` -> check passes.
6. `invalid_suppression/`: `// loy:ignore ARCH-002` without reason -> flags error.

### Performance Benchmark
- Fast AST phase must analyze a 50-package Go project in under 200ms.

---

## 5. Definition of Done
- [x] Tarjan cycle detector verified.
- [x] All 14 architecture rules implemented or assigned clear stubs.
- [x] `loy check` surfaces file, line, and remediation hint for every violation.
- [x] Suppression engine validated with strict reason checks.

---

[← Previous: Phase 3 Plan](./03-Phase-Generator-Engine.md) | [Back to Plans Index](./README.md) | [Next: Phase 5 Plan →](./05-Phase-Core-Generators.md)
