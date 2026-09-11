# ADR-001: Seams Over Components

## Status
Accepted

## Context
Go has a rich ecosystem of mature, battle-tested libraries (e.g., Fiber, sqlc, Asynq, Valkey). Frameworks often attempt to reinvent these components or wrap them in heavy proprietary abstractions, creating maintenance burdens and vendor lock-in.

## Decision
Loy standardizes the architectural *seams* (wiring, conventions, interfaces, dependency flow, lifecycle hooks) between mature Go libraries rather than writing custom framework components.

## Consequences
- Applications remain idiomatic Go.
- Upgrades to underlying libraries follow upstream semantics.
- No bespoke ORM, web server, or runtime container to maintain.
