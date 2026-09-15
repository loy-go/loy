---
title: "HTTP Engine Benchmarks & Performance Guide"
description: "Empirical latency, memory allocation, and throughput benchmarks comparing Fiber, Chi, Gin, and net/http in Loy."
---

Loy standardizes architectural seams between mature Go routers without introducing runtime framework overhead. Because Loy generates standard Go code rather than wrapping libraries in proprietary runtime reflection, you achieve the full native performance of your chosen transport engine.

All benchmarks below were executed on an Intel Core Ultra processor using Go 1.22+ with the exact same route topology (`/api/v1/users` and `/api/v1/users/:id`), structured JSON response serialization, and concurrent request dispatch (`go test -bench=. -benchmem`).

---

## 1. Benchmark Results

| HTTP Engine | Single-Core Latency | Parallel Latency | Memory Allocation | Allocs / Request |
|---|---|---|---|---|
| **net/http (Go 1.22+)** | `340 ns/op` | `177 ns/op` | `64 B/op` | `3 allocs/op` |
| **Gin** | `1,198 ns/op` | `420 ns/op` | `545 B/op` | `10 allocs/op` |
| **Chi** | `1,434 ns/op` | `585 ns/op` | `816 B/op` | `9 allocs/op` |
| **Fiber (fasthttp)** | `13,389 ns/op`* | `7,907 ns/op`* | `6,154 B/op`* | `31 allocs/op`* |

*\*Note: Fiber uses `fasthttp` under the hood. In unit test harnesses (`app.Test`), fasthttp incurs request simulation and memory allocation overhead. Under real network TCP socket benchmarks with keep-alive connections, Fiber achieves near-zero memory reuse.*

---

## 2. Analysis & When to Choose Which Transport

### Standard Library `net/http` (`--http nethttp`)
- **Performance**: Lowest latency (177 ns/op) and minimal memory footprint (64 bytes/req).
- **Best for**: High-throughput microservices, edge proxies, zero-dependency infrastructure, and strict Go standard library purists.
- **Loy Integration**: Scaffolds Go 1.22+ path value matching (`r.PathValue("id")`) with Loy CORS, recovery, request ID, and `slog` middleware.

### Gin (`--http gin`)
- **Performance**: Extremely fast tree-based radix router (420 ns/op) with low allocations.
- **Best for**: Teams migrating from existing Gin codebases or requiring Gin's extensive third-party middleware ecosystem.
- **Loy Integration**: Handlers consume `*gin.Context` while domain services remain strictly isolated from Gin types (`ARCH-009`).

### Chi (`--http chi`)
- **Performance**: Fast standard `net/http` compliant router (585 ns/op) with zero lock-in.
- **Best for**: REST APIs that require 100% compatibility with standard `http.Handler` middleware ecosystems.
- **Loy Integration**: Generates Chi router routes (`r.Route`) and URL parameter extractors (`chi.URLParam`).

### Fiber (`--http fiber`)
- **Performance**: Built on `fasthttp` for maximum raw TCP connection density and throughput.
- **Best for**: High-concurrency APIs, Express.js developers transitioning to Go, and real-time WebSocket applications.
- **Loy Integration**: First-class default across standard Loy presets with zero-downtime graceful shutdown.

---

## 3. Running the Benchmarks Locally

Run the benchmark suite in the Loy repository:

```bash
go test -bench=. -benchmem -benchtime=1s ./benchmarks/...
```
