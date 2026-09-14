---
title: "loy.yaml Manifest Reference"
description: "Annotated schema and configuration options for the Loy project manifest."
---

The `loy.yaml` file defines project metadata, defaults, integration capabilities, and workspace layout.

## Minimal Example

```yaml title="loy.yaml"
version: 1

project:
  name: bookstore

defaults:
  http: fiber
  database: postgres
  cache: valkey
  queue: asynq
```

---

## Annotated Specification

### Root Fields

- **`version`** (`int`, required): The schema version of the manifest. Must be `1`.
- **`project`** (`object`, required): Project identity settings.
  - **`name`** (`string`, required): Name of the project or root module.
- **`defaults`** (`object`, optional): Default integration adapters for scaffolding:
  - **`http`** (`string`): HTTP transport router (`fiber`, `nethttp`).
  - **`database`** (`string`): Persistence engine (`postgres`, `sqlite`, `none`).
  - **`cache`** (`string`): Caching driver (`valkey`, `redis`, `memory`, `none`).
  - **`queue`** (`string`): Background task engine (`asynq`, `none`).
  - **`template`** (`string`): UI templating engine (`templ`, `none`).
  - **`assets`** (`string`): Frontend asset pipeline (`vite`, `tailwind`, `none`).

---

### Workspace Configuration (Monorepo)

For Go multi-module workspaces:

```yaml title="loy.yaml"
version: 1

project:
  name: platform

workspace:
  default_target: api
  apps:
    - apps/api
    - apps/worker
  packages:
    - packages/domain
    - packages/telemetry
```
