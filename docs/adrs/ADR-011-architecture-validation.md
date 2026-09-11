# ADR-011: Architecture Validation Engine

## Status
Accepted

## Context
Architectural guidelines in team wikis or documents degrade over time as developers bypass boundaries under delivery pressure or through oversight.

## Decision
Provide automated, compile-time architecture boundary enforcement via `loy check`. Rules enforce strict layer directions, cycle prevention, forbidden dependencies, and workspace isolation.

## Consequences
- Continuous, objective architectural governance in CI.
- Immediate developer feedback on forbidden imports or layer violations.
- Transparent escape hatches via explicit comments (`// loy:ignore ARCH-xxx reason="..."`).
