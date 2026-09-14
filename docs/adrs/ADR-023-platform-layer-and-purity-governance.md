# ADR-023: Platform Layer & Purity Governance

## Status
Accepted

## Context
Shared project utilities (`pkg/*` or `internal/platform/*`) often contain a mix of pure helpers (string formatting, math, time) and impure components (HTTP clients, mailers, database loggers). Allowing unrestricted imports from `pkg/*` compromises domain purity, while banning `pkg/*` completely creates excessive boilerplate.

## Decision
1. Explicitly classify `pkg/*` and `platform/*` as `LayerPlatform`.
2. Update `AllowedMatrix` to permit Transport, Application, Domain, and Infrastructure to import `LayerPlatform`.
3. Introduce rule `ARCH-015` (Platform Purity): Domain and Application layers are strictly forbidden from importing impure platform packages that contain direct IO, network, or database dependencies (`net`, `net/http`, `database/sql`, `os`, `syscall`).
4. Pure utility packages remain accessible to all layers.

## Consequences
- Eliminates architecture verification bypasses on non-standard directory structures.
- Strict preservation of domain purity while enabling safe utility reuse.
- Clear diagnostics guiding developers to move impure utilities to Infrastructure.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
