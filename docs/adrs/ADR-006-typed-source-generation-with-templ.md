# ADR-006: Typed Source Generation with Templ

## Status
Superseded in part by [ADR-013](./ADR-013-codegen-engine-selection.md)

## Context
Initial design proposed using Templ across all code generation tasks for Go-native type checking during template definition.

## Decision
Adopt Templ for template-driven rendering. (Note: Subsequent evaluation in ADR-013 restricted Templ to HTML/SSR UI rendering, using `text/template` for Go source generation).

## Consequences
- See [ADR-013](./ADR-013-codegen-engine-selection.md).
