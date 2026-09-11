# Loy — Generator Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 03 TIP](./03-TIP.md) | [Index](./00-INDEX.md) | [05 Architecture Rule Spec →](./05-Architecture-Rule-Specification.md)

---

## Generator Contract

```go
type Generator interface {
    Name() string
    Generate(context.Context, Input) (*Plan, error)
}
```

Generators produce plans and do not directly mutate the filesystem.

## Pipeline

```text
CLI → Input → Generator → Model → Renderer → Artifacts → Plan → Validation → Conflicts → Apply
```

## Input

```go
type Input struct {
    Name    string
    Args    map[string]string
    Options Options
}
```

## Context

The generation context contains normalized project, workspace, configuration and capability information.

## Artifact Ownership

- `DeveloperOwned`
- `GeneratedOwned`
- `MixedOwned`

## Conflict Rules

Developer-owned existing file: error by default.  
Generated-owned existing file: regenerate.  
Mixed ownership: only explicit managed regions may be reconciled.

## Naming

All generators use a centralized naming abstraction for PascalCase, camelCase, snake_case, kebab-case and Go package names.

## Rendering

`text/template` with strict typed models and `gofmt` is the primary Go source code renderer. The renderer is abstracted behind a Go `Renderer` interface so generation logic remains renderer-agnostic. Templ is dedicated to HTML/SSR UI templates.

## Determinism

No uncontrolled timestamps, randomness, machine paths or unordered output.

## Composition

Feature and CRUD generators compose generation plans; they do not invoke the Loy CLI recursively.

## Testing

Golden tests, deterministic plan tests, conflict tests, renderer tests and generated-project acceptance tests are required.

---

**Related ADRs:**
- [ADR-007: Plan-Based Generation](./adrs/ADR-007-plan-based-generation.md)
- [ADR-010: Ordinary Go Output](./adrs/ADR-010-ordinary-go-output.md)
- [ADR-013: Codegen Engine Selection](./adrs/ADR-013-codegen-engine-selection.md)
- [ADR-014: Managed Code Splicing](./adrs/ADR-014-managed-code-splicing-via-comment-regions.md)

**Next:** [05-Architecture-Rule-Specification.md — Architecture Rule Specification](./05-Architecture-Rule-Specification.md)
