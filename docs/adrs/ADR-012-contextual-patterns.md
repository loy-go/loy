# ADR-012: Contextual Patterns Over Dogmatic Boilerplate

## Status
Accepted

## Context
Forcing developers to write Repositories, Services, Mappers, and DTOs for simple key-value reads or internal handlers creates fatigue and meaningless indirection.

## Decision
Design patterns are applied contextually: Repositories and Services are used when persistence abstraction or non-trivial business orchestration exists; trivial endpoints can map directly from transport to data adapters without artificial intermediate hops.

## Consequences
- Clean, maintainable code without hollow pass-through layers.
- Architecture rules encourage pragmatic structure over strict pattern ceremony.
