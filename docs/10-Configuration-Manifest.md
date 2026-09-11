# Loy — Configuration & Manifest Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 09 CLI DX Spec](./09-CLI-Developer-Experience.md) | [Index](./00-INDEX.md) | [11 Project Workspace Spec →](./11-Project-Workspace.md)

---

## Primary File

`loy.yaml`.

Minimal example:

```yaml
version: 1
project:
  name: acme
defaults:
  http: fiber
  database: postgres
  cache: valkey
  queue: asynq
  template: templ
  assets: vite
```

## Configuration Principles

Typed, explicit, deterministic, validated and free of secrets.

## Sections

```text
version
project
defaults
integrations
architecture
generation
workspace
```

Only meaningful sections should be emitted.

## Authority

`go.mod` defines Go module identity. `go.work` defines Go workspace semantics. `loy.yaml` defines Loy project intent.

## Precedence

```text
built-in defaults
→ preset
→ workspace configuration
→ project configuration
→ environment values where explicitly supported
→ CLI flags
```

## Validation

Unknown fields, invalid types, unsupported values, semantic conflicts and missing required environment values fail early.

## Secrets

Never store credentials directly in `loy.yaml`. Explicit references such as environment variables may be supported.

## Presets

Minimal, API, web, fullstack and monorepo are the initial preset set. Presets are starting points and are materialized so future Loy upgrades do not silently alter existing projects.

## Normalization

All downstream systems consume normalized immutable configuration.

---

**Next:** [11-Project-Workspace.md — Project & Workspace System Specification](./11-Project-Workspace.md)
