---
title: "Production Recipes & Cookbooks"
description: "End-to-end practical recipes for building real-world enterprise systems with Loy."
---

The Loy Recipes collection provides production-ready, step-by-step blueprints for assembling enterprise architectures. Each recipe combines Loy's code generators, architectural boundaries, and standard Go libraries.

## Available Recipes

| Recipe | Architecture Pattern | Key Technologies |
|---|---|---|
| **[Multi-Tenant SaaS with PostgreSQL RLS](./multi-tenant-rls/)** | Database-level tenant isolation | PostgreSQL RLS, sqlc, JWT Context |
| **[Background Worker Fleet](./background-workers/)** | Dual-daemon background task execution | Asynq, River, Valkey, Supervisor |
| **[Real-Time WebSocket Streaming](./websocket-streaming/)** | Low-latency token streaming & bidirectional chat | Fiber WebSocket, Connection Hub, Typed Frames |
| **[gRPC & REST Dual-Listener](./grpc-microservices/)** | High-performance RPC + Public JSON API | Protocol Buffers, buf, gRPC Server Adapter |
| **[Transactional Outbox Pattern](./transactional-outbox/)** | Reliable at-least-once event delivery | PostgreSQL `SKIP LOCKED`, Asynq, Outbox Dispatcher |
| **[JWT Authentication & RBAC](./auth-rbac/)** | Secure user authentication & permissions | bcrypt, Argon2id, golang-jwt/jwt/v5, RBAC Middleware |
| **[Zero-Docker Apps with SQLite](./zero-docker-sqlite/)** | Ultra-lightweight local-first prototyping | modernc.org/sqlite, Chi router, in-memory queue |
