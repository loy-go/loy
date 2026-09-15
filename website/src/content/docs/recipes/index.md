---
title: "Production Architecture Recipes"
description: "End-to-end practical recipes and architectural blueprints for building real-world enterprise systems with Loy."
---

While the [Best Practices Guides](/loy/best-practices/) explain individual components in isolation, real-world systems require composing multiple components together: authentication, background workers, databases, multi-tenancy, and streaming protocols.

The **Loy Recipes** provide end-to-end, battle-tested blueprints that demonstrate how to build complete enterprise systems with Loy.

---

## Recipe Catalog

```text
┌───────────────────────────────┬──────────────────────────────────┬─────────────────────────────┐
│ Recipe                        │ Problem Solved                   │ Key Technologies            │
├───────────────────────────────┼──────────────────────────────────┼─────────────────────────────┤
│ 1. Multi-Tenant SaaS with RLS │ Data isolation between tenants   │ PostgreSQL RLS, sqlc, JWT   │
│ 2. Background Worker Fleet    │ Heavy async tasks & cron jobs    │ Asynq, Redis/Valkey, TUI    │
│ 3. Real-Time WebSockets       │ Live bidirectional chat & feeds  │ Fiber WS, Hub, Read/Write   │
│ 4. gRPC & REST Dual-Listener  │ Internal RPC + Public JSON API   │ Protobuf, buf, gRPC Server  │
│ 5. Transactional Outbox       │ Guaranteed event publishing      │ PostgreSQL SKIP LOCKED      │
│ 6. JWT Auth & RBAC            │ Secure login & permission checks │ bcrypt, Argon2id, JWT v5    │
│ 7. Zero-Docker with SQLite    │ Ultra-fast local prototyping     │ modernc.org/sqlite, Chi     │
└───────────────────────────────┴──────────────────────────────────┴─────────────────────────────┘
```

---

## Quick Recipe Summaries

### 1. [Multi-Tenant SaaS with PostgreSQL RLS](./multi-tenant-rls/)
- **Scaffold**: `loy make tenant`
- **When to use**: Building B2B SaaS where multiple organizations share a database, but data MUST be strictly isolated so Tenant A can never see Tenant B's records.
- **How it works**: Uses native PostgreSQL Row-Level Security (`SET LOCAL app.current_tenant_id = $1`) enforced inside database transactions rather than relying on error-prone application-level `WHERE tenant_id = ?` clauses.

### 2. [Background Worker Fleet](./background-workers/)
- **Scaffold**: `loy make runtime` + `loy make job <name>`
- **When to use**: Offloading heavy computations, PDF generation, webhook dispatches, or sending emails asynchronously without blocking HTTP requests.
- **How it works**: Implements a dual-daemon architecture with a dedicated `cmd/worker/main.go` binary powered by Asynq and Valkey/Redis, supervised simultaneously during development via `loy dev --tui`.

### 3. [Real-Time WebSocket Streaming](./websocket-streaming/)
- **Scaffold**: `loy make ws <name>`
- **When to use**: Real-time collaborative apps, live chat, trading dashboards, or streaming LLM tokens to browser clients.
- **How it works**: Thread-safe central Hub with dedicated Read and Write pumps per client, ping-pong heartbeat timeouts, and non-blocking channel dispatching.

### 4. [gRPC & REST Dual-Listener](./grpc-microservices/)
- **Scaffold**: `loy make runtime --grpc` + `loy make grpc <name>`
- **When to use**: Microservices that need ultra-high throughput binary RPC internally between backend services, but also expose public JSON REST endpoints for mobile and web clients.
- **How it works**: Runs an HTTP server (port 8080) and gRPC server (port 9090) simultaneously under unified phased LIFO lifecycle supervision.

### 5. [Transactional Outbox Pattern](./transactional-outbox/)
- **Scaffold**: `loy make outbox`
- **When to use**: Publishing domain events to Kafka, RabbitMQ, or webhooks reliably when an entity is saved, preventing the "Dual-Write Problem".
- **How it works**: Writes the event record into an `outbox` table inside the *same atomic database transaction* as the entity write, then dispatches events asynchronously using `SELECT ... FOR UPDATE SKIP LOCKED`.

### 6. [JWT Authentication & Role-Based Access Control (RBAC)](./auth-rbac/)
- **Scaffold**: `loy make auth` + `loy make policy <name>`
- **When to use**: User registration, password hashing, JWT token issuance, and granular permission checking.
- **How it works**: Constant-time password hashing via bcrypt, signed JWT access tokens with standard claims, and declarative authorization policies keeping permission checks out of transport handlers.

### 7. [Zero-Docker Prototyping with SQLite](./zero-docker-sqlite/)
- **Scaffold**: `loy new myapp --db sqlite --http chi`
- **When to use**: Rapid prototyping, local-first CLI tools, offline desktop apps, or lightweight edge microservices.
- **How it works**: Pure Go SQLite (`modernc.org/sqlite`) requiring zero CGO, zero Docker containers, and zero external database daemons.
