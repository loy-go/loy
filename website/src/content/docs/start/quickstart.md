---
title: "Quickstart (Under 60 Seconds)"
description: "From installation to running an API service with live reload in 60 seconds."
---

This guide walks you through creating a complete, production-ready Go microservice with database persistence, HTTP endpoints, architectural boundary validation, and live reload.

---

## Step 1: Validate Your Environment

Run `loy doctor` to verify that your development toolchain is ready:

```bash
loy doctor
```

```text
[✓] Go Compiler     : go version go1.24.0 linux/amd64
[✓] git             : git version 2.40.0
[✓] Docker          : Docker daemon is active and running
[✓] Go Module       : Ready
```

---

## Step 2: Scaffold a New Project

Create a new application using the `api` preset:

```bash
loy new bookstore --preset api
cd bookstore
```

Loy scaffolds an ordinary Go project adhering strictly to clean architecture:

```text
bookstore/
├── cmd/api/main.go          # Application composition root
├── internal/
│   ├── app/                 # App lifecycle & wiring.go
│   ├── config/              # Typed environment configuration
│   └── platform/            # Infrastructure drivers (database, cache, telemetry)
├── loy.yaml                 # Project manifest
├── go.mod
└── go.sum
```

Verify that the fresh project compiles immediately:

```bash
go build ./...
```

---

## Step 3: Scaffold a Vertical Feature Slice

Scaffold a domain entity, repository interface, PostgreSQL adapter, service layer, HTTP transport handler, unit test, and automated wiring in a single command:

```bash
loy make crud Book title:string isbn:string:unique price:int
```

Loy outputs an atomic plan execution:
```text
Successfully generated crud Book
  + migrations/20260914_create_books_table.sql (create)
  + queries/books.sql (create)
  + internal/book/domain/book.go (create)
  + internal/book/domain/repository.go (create)
  + internal/book/repository/pg_adapter.go (create)
  + internal/book/service/service.go (create)
  + internal/book/transport/http/handler.go (create)
  + internal/book/service/service_test.go (create)
  + internal/app/wiring.go (splice)
```

Notice how `internal/app/wiring.go` was spliced cleanly via managed comment regions (`// loy:region:...`) with zero reflection or magic annotations.

---

## Step 4: Validate Architecture Boundaries

Run Loy's two-phase static architecture engine to ensure no package has introduced layer boundary violations or circular dependencies:

```bash
loy check
```

```text
All architecture rules passed.
```

---

## Step 5: Start Live Development Server

Start the application with live hot reload and process supervision:

```bash
loy dev
```

Loy watches your Go source files and re-compiles the binary on save:

```text
[app] Starting application on :8080 (framework: fiber)...
[app] HTTP server listening at 0.0.0.0:8080
```

Test your new endpoint:

```bash
curl http://localhost:8080/api/v1/books
```
