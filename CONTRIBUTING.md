# Contributing to Loy

Welcome and thank you for your interest in contributing to Loy!

Loy is an opinionated, build-time Go developer platform designed to deliver a modern, typed developer experience without runtime framework lock-in.

---

## 1. Architectural Invariants (Non-Negotiable)

Before authoring code, review our non-negotiable architectural principles:
- **Seams over components**: Loy configures and wires mature third-party libraries (`pgx`, `fiber`, `asynq`, `goose`, `sqlc`, `otel`). We do not author proprietary web frameworks or ORMs ([ADR-001](docs/adrs/ADR-001-seams-over-components.md)).
- **Zero runtime dependency**: Generated applications compile as ordinary Go binaries without depending on the `loy` CLI or runtime packages ([ADR-002](docs/adrs/ADR-002-minimal-runtime.md)).
- **Explicit wiring**: No reflection DI or magic annotations. Dependencies are injected via explicit Go constructors in `internal/app/wiring.go` ([ADR-003](docs/adrs/ADR-003-explicit-wiring.md)).
- **Plan-based atomic generation**: Code generators produce an in-memory plan before mutating disk. Splicing into existing files occurs through managed comment regions (`// loy:region:...`) ([ADR-007](docs/adrs/ADR-007-plan-based-generation.md), [ADR-014](docs/adrs/ADR-014-managed-code-splicing-via-comment-regions.md)).
- **Path sandboxing**: All file mutations must pass validation through `filesystem.CleanAndValidatePath` to prevent directory traversal attacks.
- **Zero shell interpolation**: Subprocesses are invoked with discrete arguments via `exec.CommandContext`. Never execute shell strings or `sh -c`.

---

## 2. Local Development & Quick Setup

### Prerequisites
- Go 1.24+ installed
- Git
- Docker (optional, for database integration tests)
- `golangci-lint` (v1.64+)

### Clone & Test
```bash
git clone https://github.com/loy-go/loy.git
cd loy

# Fast test run (unit tests, < 5 seconds)
make test

# Full test run including CLI integration tests
make test-all

# Lint & vet checks
make lint
make vet

# Build local binary
make build
./bin/loy doctor
```

---

## 3. Contributing a New Generator

All generators implement the `generator.Generator` interface defined in `internal/generator/contract.go`:

```go
type Generator interface {
    Name() string
    Generate(ctx context.Context, input Input) ([]model.Artifact, error)
}
```

### Generator Implementation Steps:

1. **Implement Generator**:
   Create `internal/generator/builtin/<name>.go`:
   ```go
   package builtin

   import (
       "context"
       "fmt"
       "github.com/loy-go/loy/internal/generator"
       "github.com/loy-go/loy/internal/generator/model"
   )

   type MyGenerator struct {
       modulePath string
   }

   func NewMyGenerator(modulePath string) *MyGenerator {
       return &MyGenerator{modulePath: modulePath}
   }

   func (g *MyGenerator) Name() string { return "myartifact" }

   func (g *MyGenerator) Generate(ctx context.Context, input generator.Input) ([]model.Artifact, error) {
       data := NewTemplateData(input.Name, g.modulePath)
       content, err := RenderTemplate("my_template.tmpl", data)
       if err != nil {
           return nil, err
       }

       // Always use data.FeaturePkg (lowercase) in paths and imports
       path := fmt.Sprintf("internal/%s/subpkg/%s.go", data.FeaturePkg, data.Snake)
       return []model.Artifact{
           {
               Path:        path,
               Content:     content,
               Type:        model.ArtifactGoSource,
               Permissions: 0644,
           },
       }, nil
   }
   ```

2. **Add Source Template**:
   Create `internal/generator/builtin/templates/<name>.go.tmpl`.
   **Important:** Always import domain packages using `{{.FeaturePkg}}` (lowercase) to avoid case-collision on case-insensitive filesystems:
   ```gotemplate
   package subpkg

   import (
       "{{.ModulePath}}/internal/{{.FeaturePkg}}/domain"
   )

   type {{.Pascal}}Worker struct {
       entity *domain.{{.Pascal}}
   }
   ```

3. **Register Generator in CLI**:
   Add command hook in `internal/cli/make.go`:
   ```go
   cmd.AddCommand(newArtifactCmd("myartifact", "Scaffold my custom artifact", fs, runner, opts, func(mod string) generator.Generator {
       return builtin.NewMyGenerator(mod)
   }, []string{"myart"}))
   ```

4. **Add Unit & Golden File Tests**:
   Create test in `internal/generator/builtin/<name>_test.go`:
   ```go
   func TestMyGenerator(t *testing.T) {
       gen := builtin.NewMyGenerator("github.com/example/myapp")
       artifacts, err := gen.Generate(context.Background(), generator.Input{Name: "User"})
       require.NoError(t, err)
       assert.Equal(t, "internal/user/subpkg/user.go", artifacts[0].Path)
   }
   ```

---

## 4. Contributing a New Architecture Rule

Architecture rules enforce layer boundaries during `loy check`.

Rule contract (`internal/architecture/rule.go`):
```go
type Rule interface {
    ID() string
    Description() string
    Check(ctx context.Context, a *Analysis) []Violation
}
```

### Rule Implementation Steps:

1. **Implement Rule**:
   Create `internal/architecture/rules/arch0XX_<name>.go`:
   ```go
   package rules

   import (
       "context"
       "fmt"
       "strings"
       "github.com/loy-go/loy/internal/architecture"
   )

   type RuleArch015 struct{}

   func (r *RuleArch015) ID() string { return "ARCH-015" }
   func (r *RuleArch015) Description() string {
       return "Domain layer must never import infrastructure drivers directly"
   }

   func (r *RuleArch015) Check(ctx context.Context, a *architecture.Analysis) []architecture.Violation {
       var violations []architecture.Violation
       for _, f := range a.Files {
           if f.Layer != architecture.LayerDomain {
               continue
           }
           for _, imp := range f.AST.Imports {
               importPath := strings.Trim(imp.Path.Value, `"`)
               if strings.Contains(importPath, "/internal/platform/") {
                   violations = append(violations, architecture.Violation{
                       RuleID:   r.ID(),
                       Severity: architecture.SeverityError,
                       FilePath: f.Path,
                       Line:     f.FileSet.Position(imp.Pos()).Line,
                       Message:  fmt.Sprintf("Domain file imports infrastructure: %s", importPath),
                   })
               }
           }
       }
       return violations
   }
   ```

2. **Register Rule in Registry**:
   Add rule to `DefaultRules()` in `internal/architecture/rules/registry.go`.

3. **Add Tests**:
   Add unit test in `internal/architecture/rules/rules_test.go` exercising both compliant and violating code structures.

---

## 5. Golden Test Workflow

Generators verify byte-for-byte output against committed fixtures in `testdata/golden/`:

```bash
# Verify generator output matches golden files
go test -v ./internal/generator/...

# Update golden fixtures when templates intentionally change
go test -v ./internal/generator/... -update
```

---

## 6. Pre-PR Quality Verification Gate

Before submitting a pull request, run the complete verification gate:

```bash
# 1. Syntax and static analysis
make vet
make lint

# 2. Concurrency race check and tests
make test-all

# 3. Clean binary build
make build
./bin/loy version
```
