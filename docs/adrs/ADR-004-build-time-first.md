# ADR-004: Build-Time First

## Status
Accepted

## Context
Validating architectural boundaries, discovering handlers, and checking configuration at runtime leads to slow boot times, hidden errors, and fragile production deployments.

## Decision
All validation, generation, template compilation, and architecture boundary enforcement occur strictly at build time via `loy check`, `loy build`, and `loy make`.

## Consequences
- Fast, predictable application boot sequence.
- Failures surface during development and CI, never in production.
- Lower operational risk in container environments.
