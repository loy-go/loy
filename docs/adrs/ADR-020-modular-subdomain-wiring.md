# ADR-020: Modular Sub-Domain Composition Roots

## Status
Accepted

## Context
As clean architecture applications grow across 10+ bounded contexts, splicing all constructor calls into a single flat `wireDependencies()` function in `internal/app/wiring.go` creates a monolithic file exceeding 400+ lines, degrading maintainability.

## Decision
1. Support modular domain composition roots via `internal/app/wire_<domain>.go`.
2. Each domain wire file declares `func (a *App) wire<Domain>() error` with explicit standard Go constructors.
3. `internal/app/wiring.go` acts solely as the orchestrator calling sub-domain wire methods (< 50 LOC).
4. Splicing supports `--modular` flag to generate modular wire files automatically.

## Consequences
- Composition root stays permanently organized regardless of project scale.
- Adheres strictly to ADR-003: explicit Go constructors only, zero reflection DI.
- Bounded contexts stay self-contained within the composition root.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
