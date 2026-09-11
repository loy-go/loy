# ADR-009: Progressive Complexity

## Status
Accepted

## Context
Frameworks that mandate enterprise-grade architectural boilerplate (e.g. 5 layers for a simple CRUD) alienate developers building lightweight services or prototypes.

## Decision
Loy supports progressive complexity: minimal applications stay minimal (single file or flat structure), while medium and complex applications scale into modular, layered architectures (`Transport -> Application -> Domain <- Infrastructure`) as required.

## Consequences
- Low barrier to entry for small services and microservices.
- Clear, standardized migration path to full Clean/Hexagonal architecture as requirements grow.
- Avoids dogmatic architecture ceremony for trivial use cases.
