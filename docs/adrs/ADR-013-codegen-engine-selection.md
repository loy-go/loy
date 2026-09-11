# ADR-013: Go Source Code Generation Engine Selection

## Status
Accepted

## Context
Initial design documents (04, 07, 12) specified `Templ` as the preferred renderer for both Go source files and frontend HTML templates. However, Templ is designed for HTML/markup generation with streaming contexts and HTML-escaping rules. Using it to emit Go structs, interfaces, and function bodies introduces high friction, escaping workarounds, and potential compilation mismatch.

## Decision
1. Use Go standard library `text/template` combined with strictly typed view models and `gofmt` (and optionally `goimports`) as the primary Go source code generator.
2. Abstract rendering behind the `Renderer` interface (`Render(ctx context.Context, tmpl Template, data any) ([]byte, error)`).
3. Reserve `Templ` strictly for fullstack HTML/SSR view template scaffolding.

## Consequences
- Clean, reliable generation of Go source code without HTML-escaping baggage.
- Standard Go tooling compatibility with zero extra third-party compiler requirements during core CLI codegen.
- Clear separation between backend source code generation and frontend UI component generation.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
