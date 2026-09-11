# Phase 2: Project & Workspace System Implementation Plan

**Phase:** 2 of 10  
**Status:** Ready for Implementation  
**Estimated Scope:** Discovery engine, `loy.yaml` parser, Go workspace resolver, preset initializer  
**Primary Specifications:** [10-Configuration-Manifest.md](../10-Configuration-Manifest.md), [11-Project-Workspace.md](../11-Project-Workspace.md), [09-CLI-Developer-Experience.md](../09-CLI-Developer-Experience.md)

---

## 1. Goal & Objectives
Enable Loy to detect, inspect, and model projects and workspaces:
- Discover project root by locating `loy.yaml`, `go.mod`, and `go.work`.
- Parse, validate, and normalize `loy.yaml` manifests.
- Model multi-app and multi-package workspaces.
- Provide `loy init` to onboard existing Go repos and initial presets for `loy new`.

---

## 2. Package Architecture & Internal Seams

```text
internal/
├── manifest/
│   ├── manifest.go         # Manifest schema struct (v1)
│   ├── parser.go           # YAML parser with strict field validation
│   ├── validator.go        # Semantic validator (conflicts, unknown capabilities)
│   └── defaults.go         # Built-in defaults provider
├── discovery/
│   ├── discoverer.go       # Project/workspace discovery engine
│   ├── locator.go          # Root walk up parent directories
│   └── result.go           # DiscoveredContext (RootDir, HasGoMod, HasGoWork, ManifestPath)
├── workspace/
│   ├── workspace.go        # Workspace model (Apps, Packages, Targets)
│   ├── module.go           # GoModule representation (Name, Path, DirectDependencies)
│   ├── resolver.go         # go.work and go.mod reader via `go env GOWORK` / `go list`
│   └── target.go           # Target selector (e.g. loy --target=api)
├── preset/
│   ├── preset.go           # Preset interface and registry
│   ├── api.go              # Standard API preset (Fiber + PG/sqlc + Valkey + Asynq)
│   ├── minimal.go          # Minimal preset
│   ├── fullstack.go        # Fullstack preset (Fiber + Templ + HTMX + Vite)
│   └── monorepo.go         # Monorepo preset (go.work + apps/ + packages/)
└── cli/
    ├── init.go             # loy init command
    └── new.go              # loy new command (project creation scaffold)
```

---

## 3. Concrete Implementation Steps

### Step 2.1: Manifest Parsing & Normalization (`internal/manifest`)
1. Define typed `Manifest` Go struct matching Doc 10:
   ```go
   type Manifest struct {
       Version      int                     `yaml:"version"`
       Project      ProjectConfig           `yaml:"project"`
       Defaults     DefaultsConfig          `yaml:"defaults"`
       Integrations map[string]Integration  `yaml:"integrations,omitempty"`
       Architecture ArchitectureConfig      `yaml:"architecture,omitempty"`
       Workspace    WorkspaceConfig         `yaml:"workspace,omitempty"`
   }
   ```
2. Parse YAML using strict decoding (`yaml.UnmarshalStrict` or `go-yaml/v3`) to reject unknown fields immediately.
3. Validate semantics:
   - `version` must be 1.
   - `project.name` must be a valid Go package identifier / directory name.
   - Check that capabilities (HTTP, DB, Cache, Queue) match known options.
4. Normalization: produce an immutable `NormalizedConfig` where defaults are merged.

### Step 2.2: Discovery Engine (`internal/discovery`)
1. Implement upward traversal from current working directory:
   - Check `--directory` flag first if provided.
   - Search for `loy.yaml`. If not found, check parents up to filesystem root or git boundary.
   - Inspect presence of `go.mod` and `go.work`.
2. Wrap `go env GOMOD` and `go env GOWORK` calls to ensure Go's authoritative resolution is honored.

### Step 2.3: Workspace Topology Resolver (`internal/workspace`)
1. Parse `go.work` to identify active member modules.
2. Classify member modules:
   - **Apps** (`apps/<name>` or modules containing `cmd/` or target binaries).
   - **Packages** (`packages/<name>` or shared library modules).
3. Validate initial workspace rules:
   - No conflicting module names.
   - Target matching: `--target <name>` resolves to the specific app directory.

### Step 2.4: Preset Definitions & `loy init` (`internal/preset`, `internal/cli`)
1. Implement presets: `minimal`, `api`, `web`, `fullstack`, `monorepo`.
2. Implement `loy init`:
   - Detect existing `go.mod`.
   - Prompt or infer project name and defaults.
   - Generate `loy.yaml` with chosen preset materialized (no hidden inheritance).
3. Implement `loy new <project_name>`:
   - Create destination directory.
   - Initialize `go.mod` via `go mod init`.
   - Materialize preset `loy.yaml`.

---

## 4. Test Strategy & Acceptance Criteria

### Unit Tests
- `parser_test.go`: Unknown YAML keys trigger diagnostic `LOY-CFG-001`.
- `discovery_test.go`: Detect `loy.yaml` located 3 levels above working directory.
- `workspace_test.go`: Parse sample `go.work` fixture with 2 apps and 2 packages.
- `preset_test.go`: Materializing preset generates byte-for-byte expected YAML.

### CLI Integration Tests
- `loy init` in empty folder with `go.mod` -> creates valid `loy.yaml`.
- `loy init` with existing `loy.yaml` without `--force` -> returns error code 1.
- `loy new demoapp --preset api` -> creates directory `demoapp/`, `go.mod`, `loy.yaml`, exit 0.

---

## 5. Definition of Done
- [ ] Manifest loading, validation, and normalization fully tested.
- [ ] Multi-module workspace discovery functioning on test fixtures.
- [ ] `loy new` and `loy init` working smoothly via CLI.

---

[← Previous: Phase 1 Plan](./01-Phase-Foundation.md) | [Back to Plans Index](./README.md) | [Next: Phase 3 Plan →](./03-Phase-Generator-Engine.md)
