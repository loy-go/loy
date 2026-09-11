# Phase 3: Generator Engine Implementation Plan

**Phase:** 3 of 10  
**Status:** Completed  
**Estimated Scope:** Generator contract, typed models, template engine, atomic plan builder, conflict detection  
**Primary Specifications:** [04-Generator-Specification.md](../04-Generator-Specification.md), [12-Code-Generation-Templates.md](../12-Code-Generation-Templates.md), [ADR-007](../adrs/ADR-007-plan-based-generation.md), [ADR-013](../adrs/ADR-013-codegen-engine-selection.md), [ADR-014](../adrs/ADR-014-managed-code-splicing-via-comment-regions.md)

---

## 1. Goal & Objectives
Construct the core deterministic generation engine:
- Contract: `Generator.Generate(ctx, Input) (*Plan, error)`.
- Pure memory execution: no disk mutation until full plan validation passes.
- Typed view models passed into standard Go `text/template` + `gofmt`.
- Explicit ownership rules (`DeveloperOwned`, `GeneratedOwned`, `MixedOwned`).
- Comment-based region splicing (`// loy:region:...`) for managed code updates.

---

## 2. Package Architecture & Internal Seams

```text
internal/
└── generator/
    ├── contract.go         # Generator, Input, Options interfaces
    ├── naming/
    │   ├── casing.go       # PascalCase, camelCase, snake_case, kebab-case, pluralization
    │   └── naming_test.go
    ├── model/
    │   ├── artifact.go     # Artifact (Path, Content, Ownership, Permissions)
    │   ├── ownership.go    # DeveloperOwned, GeneratedOwned, MixedOwned
    │   └── region.go       # ManagedRegion (RegionName, StartMarker, EndMarker, Content)
    ├── renderer/
    │   ├── renderer.go     # Renderer interface (Render(ctx, templateName, model) ([]byte, error))
    │   ├── text_tmpl.go    # text/template implementation + custom funcMap
    │   └── format.go       # go/format (gofmt) runner with syntax validation
    ├── plan/
    │   ├── plan.go         # Plan struct (Operations: CreateFile, Overwrite, SpliceRegion, Skip)
    │   ├── builder.go      # PlanBuilder assembling artifacts
    │   ├── conflict.go     # ConflictDetector (checks existing disk state vs ownership)
    │   └── executor.go     # PlanExecutor (applies plan atomically to FileSystem)
    └── splicer/
        ├── splicer.go      # Comment region splicer
        └── parser.go       # Finds // loy:region:<name> and // loy:endregion
```

---

## 3. Concrete Implementation Steps

### Step 3.1: Naming Abstraction (`internal/generator/naming`)
1. Implement casing converter:
   - `ToPascalCase("user_account")` -> `"UserAccount"`
   - `ToCamelCase("user_account")` -> `"userAccount"`
   - `ToSnakeCase("UserAccount")` -> `"user_account"`
   - `ToKebabCase("UserAccount")` -> `"user-account"`
   - `ToPackageName("UserAccounts")` -> `"useraccount"`
2. Implement pluralization helpers (`users` <-> `user`).

### Step 3.2: Template Renderer (`internal/generator/renderer`)
1. Define `Renderer` interface:
   ```go
   type Renderer interface {
       Render(ctx context.Context, templateName string, data any) ([]byte, error)
   }
   ```
2. Implement `TextTemplateRenderer` using `text/template`:
   - Bind common template functions (`pascal`, `camel`, `snake`, `lower`, `upper`).
   - Run `format.Source(renderedBytes)` (`gofmt`) on all rendered `.go` source files.
   - If `gofmt` fails, return descriptive diagnostic showing failing source lines and syntax error.

### Step 3.3: Region Splicer (`internal/generator/splicer`)
1. Per ADR-014, parse lines searching for markers:
   ```text
   // loy:region:services
   ...
   // loy:endregion
   ```
2. Implement `SpliceRegion(existingContent []byte, regionName string, newEntry string) ([]byte, error)`:
   - Check if entry already exists in region (idempotency).
   - Insert new line before `// loy:endregion`.
   - Run `gofmt` on complete spliced file.
   - If region tag is missing or corrupted, emit diagnostic `LOY-GEN-002`.

### Step 3.4: Plan Builder & Conflict Detector (`internal/generator/plan`)
1. Define Plan operations:
   - `OpCreate`: Create new file (fails if file exists and is `DeveloperOwned`).
   - `OpOverwrite`: Overwrite existing file (`GeneratedOwned` only, or with `--force`).
   - `OpSplice`: Update managed region in `MixedOwned` file.
   - `OpSkip`: File exists and content is identical.
2. Conflict detection logic:
   - Compare proposed artifacts against current filesystem.
   - Flag any `DeveloperOwned` file collisions.
   - Return clean summary of conflicts before execution.
3. Plan Executor:
   - Apply operations against `FileSystem`.
   - If any step fails, roll back temporary writes.

---

## 4. Test Strategy & Acceptance Criteria

### Golden Tests
- Create test template `service.go.tmpl`.
- Render with sample view model.
- Compare byte-for-byte against golden fixture in `testdata/golden/service.go`.

### Unit Tests
- `splicer_test.go`:
  - Splice into valid region -> success + properly indented.
  - Splicing duplicate entry -> no duplicate added (idempotent).
  - Missing end marker -> returns structured diagnostic error.
- `conflict_test.go`:
  - `DeveloperOwned` existing file without `--force` -> returns Conflict error.
  - `GeneratedOwned` existing file -> allows overwrite.
- `plan_test.go`:
  - Dry run returns plan without altering `memFS`.
  - Applying plan mutates `memFS` correctly.

---

## 5. Definition of Done
- [x] Centralized naming rules covered by 100% test cases.
- [x] `text/template` + `gofmt` renderer operational with clear syntax error diagnostics.
- [x] Region splicing verified across multiple consecutive invocations.
- [x] Plan generation, dry-run diffing, and atomic execution working.

---

[← Previous: Phase 2 Plan](./02-Phase-Project-System.md) | [Back to Plans Index](./README.md) | [Next: Phase 4 Plan →](./04-Phase-Architecture-Engine.md)
