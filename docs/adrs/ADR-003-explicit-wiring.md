# ADR-003: Explicit Wiring

## Status
Accepted

## Context
Reflection-based dependency injection containers or runtime service locators make dependency graphs invisible to static analysis, slow down startup, and cause runtime panic failures for missing dependencies.

## Decision
All dependencies in generated applications are explicitly constructed via standard Go constructor functions and wired together in an explicit composition root (`internal/app/wiring.go`).

## Consequences
- Compile-time safety: missing dependencies fail at build time.
- Static analysis tools and `loy check` can trace the exact dependency graph.
- No reflection overhead or magic globals.
