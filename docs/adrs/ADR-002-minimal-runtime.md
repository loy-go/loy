# ADR-002: Minimal Runtime

## Status
Accepted

## Context
Many developer frameworks require a heavy runtime library imported into production binaries, adding startup latency, memory overhead, and security attack surface.

## Decision
Generated applications require zero Loy runtime dependencies to run. Loy is strictly a build-time and developer tooling CLI.

## Consequences
- Production binaries are pure Go compiled applications.
- Runtime deployment does not need the Loy binary or CLI installed.
- Zero runtime overhead or framework lock-in.
