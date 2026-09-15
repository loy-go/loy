package mcp

// RulesCatalogMarkdown contains the authoritative documentation for ARCH-001 through ARCH-015.
const RulesCatalogMarkdown = `# Loy Clean Architecture Rules Catalog (ARCH-001 through ARCH-015)

Loy strictly enforces a 4-layer Clean Architecture model:
` + "```" + `
Transport -> Application -> Domain <- Infrastructure
` + "```" + `

---

### ARCH-001: Package Dependency Cycles Prohibited
- **Invariant**: The Go package dependency graph must be a strict Directed Acyclic Graph (DAG).
- **Prohibited**:
` + "```go" + `
// package a imports b, and package b imports a
package a
import "github.com/example/app/internal/b"
` + "```" + `
- **Permitted**: Extract shared interfaces or leaf types to a separate package.
` + "```go" + `
package a
import "github.com/example/app/internal/common"
` + "```" + `

---

### ARCH-002: Domain Must Not Import Infrastructure
- **Invariant**: Domain models and logic must remain pure Go, decoupled from persistence and database drivers.
- **Prohibited**:
` + "```go" + `
package domain
import "github.com/example/app/internal/order/repository"
` + "```" + `
- **Permitted**: Invert the dependency by defining an interface in the domain package.
` + "```go" + `
package domain

type OrderRepository interface {
    FindByID(ctx context.Context, id int64) (*Order, error)
}
` + "```" + `

---

### ARCH-003: Domain Must Not Import Transport
- **Invariant**: Domain entities must never depend on HTTP, gRPC, WebSocket, or request/response types.
- **Prohibited**:
` + "```go" + `
package domain
import "github.com/example/app/internal/order/transport/http"
` + "```" + `
- **Permitted**: Handlers in transport convert HTTP DTOs into domain entities and invoke application services.

---

### ARCH-004: Application Must Not Import Transport
- **Invariant**: Application services orchestrate business use cases and must not depend on incoming transport protocols.
- **Prohibited**:
` + "```go" + `
package service
import "github.com/example/app/internal/transport/http"
` + "```" + `
- **Permitted**: Transport layer imports Application services and calls service methods.

---

### ARCH-005: Application Must Not Import Concrete Infrastructure
- **Invariant**: Application use cases depend on consumer-owned interfaces, not concrete SQL/driver adapters.
- **Prohibited**:
` + "```go" + `
package service
import "github.com/example/app/internal/order/repository/pg"
` + "```" + `
- **Permitted**: Inject repository interface into service constructor in ` + "`internal/app/wiring.go`" + `.

---

### ARCH-006: Infrastructure Must Not Import Transport
- **Invariant**: Infrastructure persistence adapters (repositories, caches, queues) must not depend on web transport.
- **Prohibited**:
` + "```go" + `
package repository
import "github.com/example/app/internal/transport/http"
` + "```" + `
- **Permitted**: Keep infrastructure adapters focused strictly on persistence and data mapping.

---

### ARCH-007: Transport Must Not Contain Raw Persistence
- **Invariant**: Handlers parse requests, invoke application services, and serialize responses. No direct SQL queries.
- **Prohibited**:
` + "```go" + `
package handler
import "database/sql"
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    db.QueryRow(...)
}
` + "```" + `
- **Permitted**: Delegate queries to application services.

---

### ARCH-008: Forbidden Framework Packages in Domain
- **Invariant**: Domain cannot import ` + "`database/sql`" + `, ` + "`net/http`" + `, ` + "`gorm.io/gorm`" + `, or Redis/Valkey SDKs.
- **Prohibited**: Importing third-party frameworks into domain entities.
- **Permitted**: Pure Go structs and domain error constants.

---

### ARCH-009: Forbidden Transport Frameworks in Application
- **Invariant**: Application services must not import web frameworks (Fiber, Chi, Gin, Echo).
- **Prohibited**: Passing ` + "`*fiber.Ctx`" + ` or ` + "`*gin.Context`" + ` into service methods.
- **Permitted**: Pass standard ` + "`context.Context`" + ` and plain Go structs.

---

### ARCH-010: Strict 4-Layer Dependency Direction
- **Invariant**: Layer flow is strictly ` + "`Transport -> Application -> Domain <- Infrastructure`" + `.
- **Prohibited**: Backwards layer calls (e.g. Infrastructure calling Application directly).

---

### ARCH-011: Service Locators and Dynamic DI Containers Prohibited
- **Invariant**: Dynamic DI reflection containers (Uber Dig, Google Wire, DI) are prohibited per ADR-003.
- **Permitted**: Explicit constructor injection in ` + "`internal/app/wiring.go`" + `.

---

### ARCH-012: Package-Level Mutable State Prohibited
- **Invariant**: No mutable package-level ` + "`var`" + ` variables (e.g. ` + "`var DB *sql.DB`" + `).
- **Permitted**: Pass state explicitly via struct fields initialized in constructors.

---

### ARCH-013: Workspace App-to-App Dependencies Prohibited
- **Invariant**: In monorepos, ` + "`apps/foo`" + ` must never import ` + "`apps/bar`" + `.
- **Permitted**: Share logic via packages in ` + "`internal/`" + ` or ` + "`packages/`" + `.

---

### ARCH-014: Managed Comment Regions Preservation
- **Invariant**: Generated scaffolding must retain valid ` + "`// loy:region:<name>`" + ` and ` + "`// loy:endregion`" + ` markers.
- **Remediation**: Run ` + "`loy upgrade`" + ` or restore missing comment markers.

---

### ARCH-015: Platform Package Purity
- **Invariant**: Platform utility packages in ` + "`internal/platform/`" + ` must not perform direct IO, network, or DB calls if imported by Domain.
`
