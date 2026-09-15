---
title: "Model Context Protocol (MCP) Server"
description: "Turnkey MCP JSON-RPC 2.0 server for Claude Desktop, Cursor, Kilo, and Windsurf via loy mcp."
---

The **Model Context Protocol (MCP)** is the open standard that allows external AI assistants to securely inspect files, run specialized tools, and consume structured project context.

Loy includes a native, built-in MCP server over standard I/O (`stdio`) via:
```bash
loy mcp [path]
```

It communicates over standard JSON-RPC 2.0 framing, ensuring stream purity by routing all internal logs to `stderr` while keeping `stdout` reserved for protocol messages.

---

## Client Configuration

### Claude Desktop (`claude_desktop_config.json`)

On macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`  
On Linux: `~/.config/Claude/claude_desktop_config.json`  
On Windows: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "loy": {
      "command": "loy",
      "args": ["mcp", "/path/to/your/project"]
    }
  }
}
```

### Cursor IDE (`.cursor/mcp.json`)

Create `.cursor/mcp.json` in your workspace root:

```json
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

## Exposed MCP Tools

External agents can invoke any of the following strongly-typed tools:

### 1. `loy_check_architecture` (or `loy_check`)
Performs Phase 1 AST and optional Phase 2 deep type analysis across all 4 architectural layers.

- **Inputs**:
  - `path` *(string, optional)*: Target directory or workspace module.
  - `deep` *(boolean, optional)*: Enable deep type analysis using `go/packages`.
- **Output**:
  - If clean: `✅ All architecture rules passed (0 violations across 4 layers).`
  - If violations exist: Structured JSON array of `AgentDiagnostic` payloads with exact file, line, violation rationale, and self-healing remediation instructions.

### 2. `loy_make_crud`
Atomically scaffolds a full 4-layer Clean Architecture CRUD slice and wires it into the composition root (`internal/app/wiring.go`).

- **Inputs**:
  - `name` *(string, required)*: Entity name (e.g. `order`, `product`, `invoice`).
  - `fields` *(string or string[], optional)*: Entity fields with types and constraints (e.g. `"title:string:required price:float status:string"` or `["title:string:required", "price:float"]`).
  - `modular` *(boolean, optional)*: When `true`, creates an isolated `wire_<domain>.go` composition root instead of flat `wiring.go`.
- **Output**:
  - Summary of planned and atomically executed disk operations (`domain/`, `repository/`, `service/`, `transport/http/`, `test/`).

### 3. `loy_inspect_routes`
Statically inspects source code AST to discover all registered HTTP and WebSocket endpoints.

- **Inputs**: none
- **Output**: JSON array of discovered routes:
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

### 4. `loy_get_graph_diff`
Computes architectural drift between your working tree and a Git base reference.

- **Inputs**:
  - `base_ref` *(string, optional, default "main")*: Target Git commit, branch, or tag (e.g. `main`, `origin/main`, `HEAD~1`).
- **Output**: Markdown architectural drift report highlighting added/removed packages and newly introduced or resolved violations.

### 5. `loy_make_migration`
Creates a timestamped SQL migration file powered by the embedded Goose migration engine.

- **Inputs**:
  - `name` *(string, required)*: Migration name (e.g. `create_orders_table`).
- **Output**: Created file path (e.g. `migrations/20260916120000_create_orders_table.sql`).

---

## Exposed MCP Resources

Agents can read live architectural specifications and runtime topography via URI:

| Resource URI | MIME Type | Description |
|---|---|---|
| `loy://rules/catalog` | `text/markdown` | Complete specification of rules `ARCH-001` through `ARCH-015` with permitted vs. prohibited code snippets. |
| `loy://project/manifest` | `application/x-yaml` | Active `loy.yaml` configuration, capabilities, and workspace topology. |
| `loy://architecture/graph` | `text/vnd.mermaid` | Live Mermaid dependency diagram mapping all packages and cross-layer dependencies. |

---

## Security & Path Sandboxing

1. **Path Jail**: Every file access inside MCP tool executions must pass through `filesystem.CleanAndValidatePath`, strictly prohibiting directory traversal (`../`).
2. **Stream Isolation**: Internal diagnostic messages and subprocess logs are isolated to `stderr`, guaranteeing zero JSON-RPC framing corruption.
3. **No Shell Interpolation**: All tool invocations map directly to typed Go execution pipelines without passing through `sh -c`.
