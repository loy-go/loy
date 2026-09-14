# ADR-021: Transport Seam Expansion (gRPC & WebSockets)

## Status
Accepted

## Context
Enterprise systems require diverse transport protocols beyond standard REST HTTP: high-performance RPC for internal microservice communication (gRPC) and bidirectional streaming for real-time AI and messaging (WebSockets).

## Decision
1. Add `loy make grpc <name>` scaffolding Proto contracts (`proto/<domain>/v1/`) and gRPC server adapters in `internal/<domain>/transport/grpc/`.
2. Add `loy make ws <name>` scaffolding connection hubs, client read/write pumps, and typed frame protocols in `internal/<domain>/transport/ws/`.
3. Strict clean architecture enforcement: Proto and WebSocket frame structs are transport DTOs and must never leak into Domain or Application layers.
4. `cmd/api/main.go` supports dual-listener coordination (HTTP on `:8080`, gRPC on `:9090`).

## Consequences
- Full protocol flexibility without sacrificing clean architecture purity.
- Application use cases remain transport-agnostic.
- Native support for real-time streaming and high-throughput RPC.

---

[Back to ADR Index](./README.md) | [Back to Documentation Index](../00-INDEX.md)
