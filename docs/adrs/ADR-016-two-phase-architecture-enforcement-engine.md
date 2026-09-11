# ADR-016: Two-Phase Architecture Enforcement Engine

## Status
Accepted

## Context
`loy check` verifies 14 architecture rules (ARCH-001 through ARCH-014), including dependency cycles, layer boundaries, and workspace topology. Full type checking using `golang.org/x/tools/go/packages` is computationally expensive and fails on syntax or partial build errors. Conversely, pure regex or naive parsing cannot inspect deeper type constraints.

## Decision
1. Implement a two-phase architecture analysis engine:
   - **Phase 1 (Fast Syntax & Import AST)**: Uses `go/parser` and `go/token` to build package and module import graphs. Validates layer direction (`Transport -> Application -> Domain <- Infrastructure`), import boundaries, and cycle detection instantly. Works even if types are incomplete.
   - **Phase 2 (Type & Declaration Analysis)**: Uses `go/packages` with type information when running `loy check --deep` or in CI environments to catch service locators, package-level mutable state, and interface compliance rules.

## Consequences
- Sub-second feedback during everyday local CLI usage (`loy make`, `loy dev`).
- Deep, sound architectural verification when preparing PRs or running release gates.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
