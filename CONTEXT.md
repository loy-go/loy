# Project Context & Architecture Memory: Loy

## 1. Project Overview
- **Name**: Loy (`loy`)
- **Module Identity**: `github.com/loy-go/loy`
- **Go Baseline**: `go 1.24.0`
- **License**: Apache-2.0 (with 100% unrestricted user ownership of generated code)
- **Role**: Build-time Go developer platform providing typed scaffolding, vertical slicing, and automated architectural boundary enforcement without runtime framework lock-in.

---

## 2. Core Invariants
1. **Seams over components**: Standardizes architectural wiring between mature Go libraries (`pgx`, `fiber`, `asynq`, `goose`, `sqlc`, `otel`); no proprietary ORMs or web frameworks.
2. **Zero runtime dependency**: Generated applications compile as plain Go binaries without depending on the `loy` CLI or runtime packages.
3. **Explicit wiring**: No reflection DI or magic annotations; constructor injection in `internal/app/wiring.go`.
4. **Strict layer direction**: `Transport -> Application -> Domain <- Infrastructure`. Domain layer never imports infrastructure or transport.
5. **Plan-based atomic generation**: In-memory `plan.Plan` generated before disk mutations; managed comment splicing via `// loy:region:...`.
6. **Path sandboxing**: All filesystem access validated via `filesystem.CleanAndValidatePath`.
7. **Template casing safety**: Always use `{{.FeaturePkg}}` (lowercase flat/snake name) for package paths and imports rather than `{{.Feature}}` to prevent case-collision compilation bugs on case-insensitive filesystems.

---

## 3. Directory Layout & Key Packages
- `cmd/loy/`: CLI entry point (`main.go`).
- `internal/cli/`: Cobra CLI commands (`new`, `init`, `make`, `check`, `doctor`, `dev`, `migrate`, `graph`, `version`).
- `internal/generator/`:
  - `contract.go`: `Generator` interface.
  - `builtin/`: Built-in atomic and composite generators (`model`, `service`, `repository`, `handler`, `crud`, `docker`, `k8s`, `helm`, `ci`).
  - `builtin/templates/`: Go template files (`.tmpl`).
  - `plan/`: Atomic execution engine with conflict detection and transaction rollback.
  - `splicer/`: Managed comment region splicer.
- `internal/architecture/`: Two-phase architecture validation engine (`loy check`).
  - `rules/`: 14 built-in architectural verification rules (`ARCH-001` through `ARCH-014`).
- `internal/dev/`: Multi-process supervisor with live file watcher and hot reload (`loy dev`).
- `internal/integration/providers/database/`: Embedded `goose` migration engine (`loy migrate`) and `sqlc` runner.
- `internal/version/`: Build info inspector falling back to `runtime/debug.ReadBuildInfo()`.
- `docs/`: Master specification pack (`00-INDEX.md` through `22-Traceability-Roadmap.md`) and ADRs (`docs/adrs/ADR-001` through `ADR-017`).

---

## 4. Developer & Verification Workflows
- `make check`: Run `go vet`, `golangci-lint`, and fast unit tests with race detector (`go test -short -race ./...`).
- `make test-all`: Run complete test suite including CLI integration tests.
- `make build`: Compile CGO-free, stripped, trimpathed binary to `bin/loy`.
- `make golden`: Refresh golden file test fixtures in `testdata/golden/`.
- CI: `.github/workflows/ci.yml` (Linux, macOS, Windows with Go 1.24, `govulncheck`, `golangci-lint`).
- Releases: GoReleaser (`.goreleaser.yaml`) with Cosign signatures and Homebrew tap (`loy-go/homebrew-tap`).
