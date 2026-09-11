# Loy — Code Generation & Template Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 11 Project Workspace Spec](./11-Project-Workspace.md) | [Index](./00-INDEX.md) | [13 Architecture Design Patterns Spec →](./13-Architecture-Design-Patterns.md)

---

## Rendering Architecture

```text
Generator → Typed Generation Model → Go Template Renderer (text/template) → Go Source → gofmt → Validation
```

## Renderer Interface

```go
type Renderer interface {
    Render(context.Context, Template, any) ([]byte, error)
}
```

## Template Engine Decision

`text/template` combined with strict typed generation models and `gofmt` is the primary engine for Go source code generation. Templ is reserved specifically for fullstack SSR/HTML component generation where its HTML-escaping and component model provide clear value.

## Template Responsibilities

Templates decide how code is rendered. Generators decide what is generated, where it lives, ownership, required capabilities and dependencies.

## Mixed Ownership

Allowed only through explicit managed regions. Arbitrary textual merge is prohibited.

## Import Handling

Generated source must contain valid imports. `gofmt` is the initial formatter. `goimports` is optional later.

## Determinism

Template execution must not depend on map iteration order, wall-clock time, random IDs or machine-local paths.

---

**Related ADRs:**
- [ADR-006: Typed Source Generation with Templ](./adrs/ADR-006-typed-source-generation-with-templ.md)
- [ADR-010: Ordinary Go Output](./adrs/ADR-010-ordinary-go-output.md)
- [ADR-013: Go Source Code Generation Engine Selection](./adrs/ADR-013-codegen-engine-selection.md)

**Next:** [13-Architecture-Design-Patterns.md — Application Architecture & Design Pattern Specification](./13-Architecture-Design-Patterns.md)
