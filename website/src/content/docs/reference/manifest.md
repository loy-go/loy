---
title: "loy.yaml Manifest Reference"
description: "Authoritative specification, annotated schema, resolution cascade, and configuration options for the Loy project manifest."
---

The `loy.yaml` file is the declarative manifest for Loy projects. It defines project identity, integration adapter defaults, multi-tenancy rules, topological architectural layers, and workspace structure.

---

## 1. Schema Overview & Full Example

```yaml title="loy.yaml"
version: 1

project:
  name: bookstore
  description: "High-performance enterprise bookstore API"
  module: github.com/example/bookstore

defaults:
  http: fiber          # fiber | chi | gin | nethttp | echo
  database: postgres   # postgres | sqlite | mysql | none
  cache: valkey        # valkey | redis | memory | none
  queue: asynq         # asynq | river | none
  template: templ      # templ | none
  assets: vite         # vite | tailwind | none

multi_tenancy:
  enabled: true
  strategy: rls        # rls (PostgreSQL Row Level Security) | column
  tenant_key: org_id   # database column identifier for tenant separation

architecture:
  pattern: custom      # ddd | hexagonal | cqrs | custom
  strict: true         # promote architectural warnings to fatal errors
  excluded:
    - "testdata/**"
    - "vendor/**"
  layers:
    transport:
      allows: [application, platform]
      match:
        - "internal/transport/**"
        - "cmd/**"
    application:
      allows: [domain, platform]
      match:
        - "internal/application/**"
        - "internal/*/service/**"
        - "internal/*/command/**"
        - "internal/*/query/**"
    domain:
      allows: [platform]
      match:
        - "internal/domain/**"
        - "internal/*/domain/**"
        - "internal/*/model/**"
    infrastructure:
      allows: [domain, platform]
      match:
        - "internal/infrastructure/**"
        - "internal/*/repository/**"
        - "internal/platform/database/**"

integrations:
  postgres:
    driver: pgx
    enabled: true
    config:
      max_conns: 50
      min_conns: 5
      sslmode: disable

workspace:
  default_target: api
  apps:
    - apps/api
    - apps/worker
  packages:
    - packages/domain
    - packages/telemetry
```

---

## 2. Configuration Resolution Cascade

Loy resolves configuration values across four levels of precedence. Higher levels unconditionally override lower levels:

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 1. CLI Flags (Highest Precedence)                                      │
│    e.g. --http=chi --db=sqlite --multi-tenant=rls                      │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ overrides
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 2. Environment Variables                                               │
│    e.g. LOY_DEFAULTS_HTTP=chi LOY_DEFAULTS_DATABASE=sqlite            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ overrides
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 3. Manifest File (`loy.yaml`)                                          │
│    Active project configuration parsed strictly from workspace root    │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ overrides
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ 4. Hardcoded Preset Defaults (Lowest Precedence)                       │
│    Preset defaults (e.g. api preset defaults to fiber + postgres)      │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Annotated Specification

### Root Fields

#### `version` (`int`, required)
The schema specification version. Must be `1`.

#### `project` (`object`, required)
Defines project identity metadata.
- **`name`** (`string`, required): Name of the project or root application. Must match alphanumeric naming rules `^[a-zA-Z0-9_-]+$`.
- **`description`** (`string`, optional): Human-readable summary of the application.
- **`module`** (`string`, optional): Authoritative Go module path (e.g. `github.com/myorg/myapp`). If omitted, Loy discovers the module path directly from `go.mod`.

---

### `defaults` (`object`, optional)

Specifies default integration drivers applied when running scaffolding commands (`loy make ...`):

| Field | Type | Allowed Values | Default (API Preset) | Description |
|---|---|---|---|---|
| **`http`** | `string` | `fiber`, `chi`, `gin`, `nethttp`, `echo` | `fiber` | HTTP transport framework and router engine. |
| **`database`** | `string` | `postgres`, `sqlite`, `mysql`, `none` | `postgres` | Persistence database engine and driver. |
| **`cache`** | `string` | `valkey`, `redis`, `memory`, `none` | `valkey` | Caching engine and client adapter. |
| **`queue`** | `string` | `asynq`, `river`, `none` | `asynq` | Distributed background task queue client. |
| **`template`** | `string` | `templ`, `none` | `none` | Server-side UI component templating engine. |
| **`assets`** | `string` | `vite`, `tailwind`, `none` | `none` | Frontend bundle toolchain. |

---

### `multi_tenancy` (`object`, optional)

Configures data isolation strategies for multi-tenant applications:

- **`enabled`** (`bool`, required): Activates multi-tenancy filters across code generators.
- **`strategy`** (`string`, required): Isolation mechanism:
  - `rls`: PostgreSQL Row Level Security. Queries use automatic session variables (`SET LOCAL app.current_org_id = ...`) to enforce isolation at the database kernel level without manual `WHERE` clauses.
  - `column`: Explicit column tenancy. Automatically appends tenant filter parameters (`WHERE org_id = $1`) to all SQL queries and Go repository interfaces.
- **`tenant_key`** (`string`, optional): Database column identifier for tenant isolation. Defaults to `org_id`.

---

### `architecture` (`object`, optional)

Configures architectural rules and layer boundary validation for `loy check`:

- **`strict`** (`bool`, optional, default `false`): When `true`, any architectural rule violation or warning triggers an immediate non-zero exit code (`1`), blocking CI pipelines.
- **`pattern`** (`string`, optional, default `ddd`): Pre-defined architectural style topology:
  - `ddd`: Canonical 4-layer Domain-Driven Design (`Transport -> Application -> Domain <- Infrastructure`).
  - `hexagonal`: Ports & Adapters (`Adapters -> Application Core <- Secondary Adapters`).
  - `cqrs`: Segregated command and query topologies.
  - `custom`: Explicit directed graph defined via `layers:`.
- **`excluded`** (`list[string]`, optional): Relative glob patterns excluded from static AST analysis.
- **`layers`** (`map[string]object`, optional): User-defined topological Directed Acyclic Graph (DAG):
  - **`allows`** (`list[string]`): Layer names that packages in this layer are permitted to import.
  - **`match`** (`list[string]`): Relative glob patterns defining which packages belong to this layer.

#### Architecture DAG Matching Behavior
When `loy check` analyzes a Go file:
1. The relative package path is matched against layer `match` globs.
2. For every imported package, Loy resolves its matching layer.
3. If the imported package's layer is not declared in the importing layer's `allows` slice, Loy emits an `ARCH-010` architectural violation.

---

### `integrations` (`map[string]object`, optional)

Provides fine-grained adapter configurations for external infrastructure:

```yaml title="loy.yaml"
integrations:
  postgres:
    driver: pgx
    enabled: true
    config:
      max_conns: 25
      min_conns: 5
      max_idle_time: 15m
  valkey:
    driver: valkey-go
    enabled: true
    config:
      cluster: false
      db: 0
```

---

### `workspace` (`object`, optional)

For Go multi-module monorepos managed via `go.work`:

- **`default_target`** (`string`, optional): Default sub-module targeted by CLI commands when `-C` or `--target` is omitted.
- **`apps`** (`list[string]`, optional): List of application daemon directory paths (e.g. `apps/api`, `apps/worker`).
- **`packages`** (`list[string]`, optional): List of shared internal library directory paths (e.g. `packages/domain`, `packages/telemetry`).

---

## 4. Environment Variable Overrides

Every configuration field in `loy.yaml` can be overridden via environment variables:

| Environment Variable | Overrides Manifest Key | Example |
|---|---|---|
| `LOY_PROJECT_NAME` | `project.name` | `LOY_PROJECT_NAME=inventory` |
| `LOY_PROJECT_MODULE` | `project.module` | `LOY_PROJECT_MODULE=github.com/org/app` |
| `LOY_DEFAULTS_HTTP` | `defaults.http` | `LOY_DEFAULTS_HTTP=chi` |
| `LOY_DEFAULTS_DATABASE` | `defaults.database` | `LOY_DEFAULTS_DATABASE=sqlite` |
| `LOY_DEFAULTS_CACHE` | `defaults.cache` | `LOY_DEFAULTS_CACHE=redis` |
| `LOY_DEFAULTS_QUEUE` | `defaults.queue` | `LOY_DEFAULTS_QUEUE=river` |
| `LOY_MULTI_TENANCY_ENABLED` | `multi_tenancy.enabled` | `LOY_MULTI_TENANCY_ENABLED=true` |
| `LOY_MULTI_TENANCY_STRATEGY` | `multi_tenancy.strategy` | `LOY_MULTI_TENANCY_STRATEGY=rls` |
| `LOY_ARCHITECTURE_STRICT` | `architecture.strict` | `LOY_ARCHITECTURE_STRICT=true` |
| `LOY_ARCHITECTURE_PATTERN` | `architecture.pattern` | `LOY_ARCHITECTURE_PATTERN=custom` |

---

## 5. Diagnostics & Validation Error Codes

When `loy.yaml` is parsed or validated, errors are returned as structured diagnostics:

| Code | Severity | Cause | Remediation |
|---|---|---|---|
| **`LOY-CFG-001`** | Error | Missing required field (`version` or `project.name`). | Add `version: 1` and a non-empty `project.name`. |
| **`LOY-CFG-002`** | Error | Schema version mismatch (`version != 1`). | Update schema to `version: 1`. |
| **`LOY-CFG-003`** | Error | Unknown integration driver or capability specified. | Check permitted values in table above. |
| **`LOY-CFG-004`** | Error | Invalid YAML syntax or unmarshal failure. | Verify indentation and valid YAML 1.2 syntax. |
| **`LOY-CFG-005`** | Error | Malicious path or directory traversal in workspace paths. | Ensure all workspace paths are relative and confined to project root. |
