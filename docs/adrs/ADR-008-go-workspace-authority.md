# ADR-008: Go Workspace Authority

## Status
Accepted

## Context
Monorepos and multi-module Go setups often fight proprietary workspace tools when managing dependencies, build paths, and module resolutions.

## Decision
Go's native `go.mod` and `go.work` remain the sole authoritative source of truth for Go package identity, dependency resolution, and workspace semantics. `loy.yaml` defines higher-level architectural intent and application targets.

## Consequences
- Full compatibility with official Go toolchain commands (`go test`, `go build`, `go work`).
- IDEs (gopls) function without specialized Loy plugins.
- Zero impedance mismatch with existing Go CI pipelines.
