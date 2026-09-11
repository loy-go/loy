# ADR-007: Plan-Based Generation

## Status
Accepted

## Context
Generators that mutate files directly on disk can leave the filesystem in a corrupted, half-written state if an error occurs mid-generation or if conflicts arise.

## Decision
All Loy generators compute an in-memory `Plan` containing all intended file modifications, additions, and region updates. The plan is validated, conflicts are detected, and changes are applied atomically only if all checks pass.

## Consequences
- Safe file operations with deterministic rollback on error.
- `--dry-run` and visual diff capabilities come naturally without touching disk.
- Safe integration with developer-owned code regions.
