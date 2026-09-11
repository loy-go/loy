# Loy — Application Architecture & Design Pattern Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 12 Code Generation Templates Spec](./12-Code-Generation-Templates.md) | [Index](./00-INDEX.md) | [14 Error Diagnostics Spec →](./14-Error-Diagnostics.md)

---

## Core Decision

**Loy enforces architecture, not pattern ceremony.**

## Default Layers

Transport → Application → Domain. Infrastructure implements dependencies required by inner layers.

## Domain Rules

Domain must remain independent of Fiber, HTTP transport, GORM, sqlc, PostgreSQL, Valkey/Redis, Asynq, gRPC transport and external SDKs by default.

## Patterns

### Repository

Use when persistence abstraction or a meaningful boundary is needed. Not mandatory for trivial read-only tooling.

### Service / Use Case

Use for meaningful application orchestration. Do not create empty services only to satisfy layering.

### DTO

Appropriate at external boundaries. Not required internally.

### Mapper

Introduce only for complex or repeated transformation semantics.

### Factory

Use for non-trivial construction or multiple implementation selection.

### Strategy

Use for meaningful runtime-varying behavior.

### Adapter

Strongly preferred at external/vendor boundaries.

### Event / Observer

Use for decoupling, asynchronous work, multiple consumers or auditability; do not replace simple synchronous calls without reason.

## Dependency Injection

Constructor injection is required. No DI container or reflection-based discovery.

## Progressive Complexity

Small applications can stay small. Medium and complex applications may introduce stronger application/domain abstractions only where justified.

---

**Related ADRs:**
- [ADR-009: Progressive Complexity](./adrs/ADR-009-progressive-complexity.md)
- [ADR-012: Contextual Patterns Over Dogmatic Boilerplate](./adrs/ADR-012-contextual-patterns.md)

**Next:** [14-Error-Diagnostics.md — Error & Diagnostics Specification](./14-Error-Diagnostics.md)
