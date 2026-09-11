# Loy — Architecture Rule Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 04 Generator Spec](./04-Generator-Specification.md) | [Index](./00-INDEX.md) | [06 Test Strategy →](./06-Test-Strategy-Acceptance.md)

---

## Purpose

`loy check` answers: "Does the project respect the architectural boundaries it claims?" It operates via a two-phase analysis engine (see [ADR-016](./adrs/ADR-016-two-phase-architecture-enforcement-engine.md)).

## Rule Contract

```go
type Rule interface {
    ID() string
    Description() string
    Check(context.Context, Analysis) []Violation
}
```

## Initial Rules

- ARCH-001 dependency cycles
- ARCH-002 domain → infrastructure
- ARCH-003 domain → transport
- ARCH-004 application → transport
- ARCH-005 application → concrete infrastructure
- ARCH-006 infrastructure → transport
- ARCH-007 transport business logic is controlled through package ownership/configuration, not brittle heuristics
- ARCH-008 forbidden imports
- ARCH-009 forbidden dependency categories
- ARCH-010 layer direction
- ARCH-011 service locator where objectively detectable/configured
- ARCH-012 global mutable state where reliably detectable
- ARCH-013 package/feature boundary
- ARCH-014 generated artifact location/ownership

## Layer Matrix

| From ↓ / To → | Transport | Application | Domain | Infrastructure |
|---|---:|---:|---:|---:|
| Transport | ✓ | ✓ | ✓ | ✗ |
| Application | ✗ | ✓ | ✓ | ✗ |
| Domain | ✗ | ✗ | ✓ | ✗ |
| Infrastructure | ✗ | ✓ | ✓ | ✓ |

## Workspace Rules

Application → Package allowed. Application → Application disallowed by default. Package → Application disallowed. Cycles disallowed.

## Suppression

Foundational rules are not suppressible. Other rules may use explicit comments with required reasons.

Example:

```go
// loy:ignore ARCH-005 reason="temporary migration adapter"
```

## Non-Goals

The architecture engine does not attempt to score style, algorithmic quality, runtime performance or the number of design patterns used.

---

**Related ADRs:**
- [ADR-011: Architecture Validation Engine](./adrs/ADR-011-architecture-validation.md)
- [ADR-016: Two-Phase Architecture Enforcement Engine](./adrs/ADR-016-two-phase-architecture-enforcement-engine.md)

**Next:** [06-Test-Strategy-Acceptance.md — Test Strategy & Acceptance Specification](./06-Test-Strategy-Acceptance.md)
