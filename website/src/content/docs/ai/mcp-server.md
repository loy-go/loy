---
title: "Model Context Protocol (MCP) Server"
description: "Turnkey MCP JSON-RPC 2.0 server for Claude Desktop, Cursor, Kilo, and Windsurf via loy mcp with concrete syntax tree (DST) tools."
---

The **Model Context Protocol (MCP)** is the open standard that enables autonomous AI coding assistants to securely inspect project architecture, execute deterministic tools, and safely manipulate source code without syntax errors or hallucinated workarounds.

Loy provides a native, built-in MCP server communicating over standard I/O (`stdio`) via:

```bash
loy mcp [path]
```

It implements the official JSON-RPC 2.0 framing specification. To guarantee stream purity and prevent stream corruption, all internal logging is strictly isolated to `stderr`, while `stdout` is reserved exclusively for machine-readable JSON-RPC envelopes.

---

## 1. Client Configuration Setup

### 1.1 Claude Desktop
Edit `claude_desktop_config.json`:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json title="claude_desktop_config.json"
{
  "mcpServers": {
    "loy": {
      "command": "loy",
      "args": ["mcp", "/absolute/path/to/your/project"]
    }
  }
}
```

---

### 1.2 Cursor IDE
Create `.cursor/mcp.json` in your workspace root:

```json title=".cursor/mcp.json"
{
  "mcpServers": {
    "loy": {
      "command": "loy",
      "args": ["mcp", "."]
    }
  }
}
```

---

### 1.3 Windsurf & Kilo
In your toolchain settings or `.kilo/config.json`:

```json title=".kilo/config.json"
{
  "mcp": {
    "servers": {
      "loy": {
        "command": "loy",
        "args": ["mcp", "."]
      }
    }
  }
}
```

---

## 2. Complete MCP Tools Reference

External agents can invoke any of the 9 strongly-typed tools exposed by Loy:

### 1. `loy_check_architecture` (or `loy_check`)
Performs static AST and deep type analysis across all 4 architectural layers.

- **Parameters**:
  - `path` *(string, optional)*: Target directory or workspace module.
  - `deep` *(boolean, optional)*: Enables deep type checking via `go/packages`.
- **Response**:
  - If clean: `✅ All architecture rules passed (0 violations across 4 layers).`
  - If violations exist: Structured JSON array of `AgentDiagnostic` payloads detailing the exact file, line number, architectural rationale, and step-by-step remediation prompt.

---

### 2. `loy_make_crud`
Atomically scaffolds a full 4-layer Clean Architecture vertical slice and wires it into the composition root (`internal/app/wiring.go`).

- **Parameters**:
  - `name` *(string, required)*: Entity name (e.g. `order`, `product`, `invoice`).
  - `fields` *(string or string[], optional)*: Entity fields with types and constraints (e.g. `"title:string:required price:float status:string"`).
  - `modular` *(boolean, optional)*: Creates an isolated `wire_<domain>.go` sub-domain composition root instead of flat `wiring.go`.
- **Response**: Summary of planned and atomically executed disk operations (`domain/`, `repository/`, `service/`, `transport/http/`, `test/`).

---

### 3. `loy_inspect_routes`
Statically inspects source code AST to discover all registered HTTP and WebSocket endpoints without running the application.

- **Parameters**: None.
- **Response**:
```json
[
  {
    "method": "GET",
    "path": "/api/v1/orders/:id",
    "handler": "h.GetByID",
    "file": "internal/order/transport/http/handler.go",
    "line": 42
  }
]
```

---

### 4. `loy_get_graph_diff`
Computes architectural drift between your current working tree and a Git base reference.

- **Parameters**:
  - `base_ref` *(string, optional, default "main")*: Target Git commit, branch, or tag (e.g. `main`, `origin/main`, `HEAD~1`).
- **Response**: Markdown drift report highlighting added/removed packages and newly introduced or resolved violations.

---

### 5. `loy_make_migration`
Creates a timestamped SQL migration file powered by the embedded Goose migration engine.

- **Parameters**:
  - `name` *(string, required)*: Migration name (e.g. `create_orders_table`).
  - `recipe` *(string, optional)*: Zero-downtime recipe: `raw`, `index-concurrent`, or `shadow-column`.
  - `table` *(string, optional)*: Target table name for recipes.
- **Response**: Created migration file path.

---

### 6. `loy_ast_insert_field`
Safely inserts a typed struct field with comments and struct tags into Go source code using Concrete Syntax Tree (`dst`) parsing.

- **Why it matters**: Unlike regex or naive line insertion, `dst` manipulation **guarantees valid Go syntax** and **preserves all surrounding developer comments**.
- **Parameters**:
  - `file` *(string, required)*: Target Go file path relative to project root.
  - `struct_name` *(string, required)*: Target struct type name (e.g. `User`, `OrderRequest`).
  - `field_name` *(string, required)*: Field identifier (e.g. `Email`, `Roles`).
  - `field_type` *(string, required)*: Go type expression (e.g. `string`, `*int`, `[]string`).
  - `tags` *(string, optional)*: Field struct tags (e.g. `json:"email,omitempty"`).
- **Response**: Confirmation of successful AST insertion.

---

### 7. `loy_ast_add_route`
Safely adds an HTTP route registration statement to the router setup function in Go source code using DST AST mutations.

- **Parameters**:
  - `file` *(string, required)*: Target Go file containing route registration.
  - `method` *(string, required)*: HTTP method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`).
  - `path` *(string, required)*: Route path (e.g. `/users`, `/orders/:id`).
  - `handler` *(string, required)*: Handler identifier or method (e.g. `h.Create`, `h.List`).
- **Response**: Confirmation of successful route injection.

---

### 8. `loy_ast_bind_dependency`
Explicitly binds a constructor dependency into the composition root (`internal/app/wiring.go`) per ADR-003 without reflection.

- **Parameters**:
  - `file` *(string, optional, default "internal/app/wiring.go")*: Target composition root file.
  - `provider_func` *(string, required)*: Constructor invocation (e.g. `userRepo.NewPostgresRepository(a.db)`).
  - `dep_name` *(string, required)*: Variable identifier (e.g. `userRepo`).
- **Response**: Confirmation of dependency binding.

---

### 9. `loy_plan_preview`
Executes code generation in an in-memory virtual filesystem sandbox (`filesystem.MemFileSystem`) and validates architectural boundaries (`loy check`) before touching the physical disk.

- **Parameters**:
  - `generator` *(string, required)*: Generator name (`crud`, `model`, `command`, `query`, `migration`).
  - `name` *(string, required)*: Entity or feature name.
  - `fields` *(string, optional)*: Space-separated field specifications.
  - `args` *(object, optional)*: Additional generator arguments.
- **Response**:
```json
{
  "status": "ready",
  "artifacts": [
    { "path": "internal/book/domain/book.go", "action": "create", "bytes": 412 },
    { "path": "internal/book/service/service.go", "action": "create", "bytes": 854 },
    { "path": "internal/app/wiring.go", "action": "modify", "bytes": 128 }
  ],
  "violations": []
}
```

---

## 3. Exposed MCP Resources

AI agents can read live project state directly via standard MCP resource URIs:

- **`loy://manifest`**: Returns the parsed, validated `loy.yaml` project specification.
- **`loy://graph`**: Returns the complete project package dependency graph in JSON format.
- **`loy://routes`**: Returns all registered HTTP and WebSocket endpoints discovered via static AST inspection.
