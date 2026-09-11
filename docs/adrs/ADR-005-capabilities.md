# ADR-005: Capabilities-Oriented Integrations

## Status
Accepted

## Context
Tying application generators directly to concrete vendor technologies limits adaptability and complicates testing and multi-cloud configurations.

## Decision
Generators and core systems express requirements as high-level capabilities (e.g. `HTTP`, `Database`, `Cache`, `Queue`, `Telemetry`) rather than specific vendors. The integration registry maps requested capabilities to concrete implementations based on `loy.yaml`.

## Consequences
- Application architecture is loosely coupled to concrete dependencies.
- Swapping infrastructure implementations does not break domain or application contracts.
- Clear contract boundaries for future integration plugins.
