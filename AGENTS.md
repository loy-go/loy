# AGENTS.md — Loy Developer Platform Engineering Guide

> **Authoritative Specification:**
> The documentation pack in [`docs/`](docs/00-INDEX.md) is the locked single source of truth for all requirements, architecture rules, and design choices. When resolving ambiguity, always adhere to the locked specs and accepted ADRs in `docs/adrs/`. Do not guess or deviate.

---

## 1. Core Architectural Invariants (Non-Negotiable)

- **Seams over components**: Loy standardizes the architectural wiring between mature Go libraries; it does not build proprietary ORMs or web frameworks ([ADR-001](docs/adrs/ADR-001-seams-over-components.md)).
- **Zero runtime dependency**: Generated apps run as ordinary compiled Go binaries without requiring the Loy binary or framework packages at runtime ([ADR-002](docs/adrs/ADR-002-minimal-runtime.md)).
- **Explicit wiring**: No reflection DI or runtime service locators. Dependencies are injected via standard Go constructors in `internal/app/wiring.go` ([ADR-003](docs/adrs/ADR-003-explicit-wiring.md)).
- **Strict layer direction**: `Transport -> Application -> Domain <- Infrastructure`. Domain code never imports infrastructure or transport ([docs/05-Architecture-Rule-Specification.md](docs/05-Architecture-Rule-Specification.md)).
- **No global mutable state**: Package-level mutable variables are strictly prohibited. State and options must pass via explicit structs or `context.Context`.
- **Atomic plan generation**: Code generators produce an in-memory `Plan` before touching the filesystem. No half-written or corrupt disk state ([ADR-007](docs/adrs/ADR-007-plan-based-generation.md)).
- **Path sandbox**: All filesystem operations must pass path jail validation via `filesystem.CleanAndValidatePath` ([docs/16-Security.md](docs/16-Security.md)).
- **Go codegen engine**: `text/template` + `gofmt` for Go source generation; `Templ` is reserved exclusively for HTML SSR UI templates ([ADR-013](docs/adrs/ADR-013-codegen-engine-selection.md)).

---

## 2. Senior Engineering Standards & Code Hygiene

Every contributor and AI agent must uphold standard Go senior engineering practices:

### 2.1 Interface & API Design
- **Accept interfaces, return structs**: Functions and constructors accept narrow interfaces defined where they are consumed, and return concrete structs.
- **Consumer-owned interfaces**: Interfaces belong to the consuming package (e.g. `service` package defines the repository interface it needs), never the implementation package.
- **No package stutter**: Avoid `filesystem.Filesystem` or `generator.GeneratorEngine`. Use `filesystem.FileSystem` or `generator.Engine`.
- **Explicit constructor injection**: Use `New<Type>(deps...) (*<Type>, error)`. Validate all required dependencies at construction time; fail fast on `nil`.

### 2.2 Error Handling & Diagnostics
- **Structured diagnostics**: User-facing or CLI errors must wrap or construct a `diagnostics.Diagnostic` with severity, code (`LOY-*`), message, and remediation hint ([docs/14-Error-Diagnostics.md](docs/14-Error-Diagnostics.md)).
- **Wrap internal errors**: Always wrap lower-level errors with context using `%w`: `fmt.Errorf("reading manifest %s: %w", path, err)`.
- **Sentinel error matching**: Use `errors.Is` and `errors.As`. Never inspect error strings directly (`strings.Contains(err.Error(), ...)`).
- **Stream purity & JSON isolation**: Never call `fmt.Println` or write directly to standard streams (`os.Stdout`/`os.Stderr`) in subcommands or helpers without checking `--json` and `--quiet`. Use `cmd.OutOrStdout()` or structured diagnostics to avoid corrupting machine-readable streams.

### 2.3 Concurrency & Resource Safety
- **No unowned goroutines**: Every spawned goroutine must have an explicit owner, a bound context cancellation path, and a `sync.WaitGroup` or errgroup coordinating its exit.
- **Leak prevention**: Always register resource cleanups immediately via `defer`: files, HTTP bodies (`defer resp.Body.Close()`), and mutexes.
- **No magic sleeps**: Never use `time.Sleep` to synchronize concurrent operations in tests or production. Use channels, sync primitives, or polled wait conditions with timeout bounds.
- **Pre-warmed engines & caching**: Heavy template renderers, AST parsers, or reflection models must be pre-warmed or cached as singletons; avoid repeated re-instantiations and redundant allocations in hot loops.

### 2.4 Security & Subprocess Execution
- **Zero shell interpolation**: External processes must be invoked via `exec.CommandContext(ctx, name, args...)` with separate argument slices. Never execute `sh -c` or concatenate user inputs into shell strings.
- **Sandboxed filesystem access**: No file read, write, or stat may occur without passing through `filesystem.CleanAndValidatePath` or package-level target resolver (`ResolveTargetPath`) to guarantee the target stays within the designated root directory.
- **Atomic mutations & journal ordering**: In atomic execution pipelines, state journal records must be pre-registered *before* file mutations occur to guarantee clean rollbacks on partial write failures.
- **Robust multi-line pattern matching**: Block search and code splicing must match multi-line subsequences rather than single-line exact equality to preserve idempotency on multi-line statements.
- **Explicit target directory execution**: Subprocess runners (e.g. `sqlc generate`, `go mod`) must execute explicitly within the resolved target application module directory (`targetDir`), never defaulting blindly to `.` (the CLI invocation directory).

### 2.5 Security, Input Validation & Boundary Hygiene
- **Strict identifier & DDL validation**: Any user input interpolated into database identifiers, table names, or raw SQL must pass strict alphanumeric regex validation (`^[a-zA-Z_][a-zA-Z0-9_]*$`). Never trust CLI flags or configs directly in DDL.
- **Path jail on all target creations**: Every file or directory path derived from user input or arguments must verify `filepath.Base(name) == name` and pass `filesystem.CleanAndValidatePath` before filesystem access.
- **Flag semantics & zero-value distinction**: When evaluating optional numerical CLI flags, always check `cmd.Flags().Changed("flag")`. Never assume `> 0` because zero (`0`) is often a valid semantic argument (e.g. `--to 0` for base migration rollback).
- **Resolution cascade hygiene**: Keep intermediate configuration structures zero-valued during cascade resolution (CLI flags > Env > Config file). Apply hardcoded defaults only at the final resolution step to avoid shadowing lower-precedence config files.
- **Template compile & dead code audit**: All Go code templates must be verified for unused imports (`gofmt`/`go/parser`) and zero unreferenced dead artifacts. Never silence template rendering errors with `if err == nil`.


---

## 3. Standard 6-Step Implementation Pipeline

When implementing any phase or feature from [`docs/plans/`](docs/plans/README.md), execute this exact 6-step loop:

```text
1. Contract First       -> Define consumer interfaces & typed models in internal/<pkg>
2. Contract & Unit Tests-> Test happy paths, timeouts, errors, and test isolation (t.Cleanup)
3. Core Implementation  -> Explicit constructor injection, no globals, path jail validation
4. Test Double Parity   -> Ensure memFS/mock doubles pass identical contract suites as real OS
5. Binary E2E Smoke     -> Compile bin/loy and verify real subprocess exit codes (0, 1, 2) and --json
6. Release Gate Checks  -> Run go vet ./... and go test -v -race ./... (target coverage >= 80%)
```

Reference: [docs/plans/IMPLEMENTATION-PIPELINE.md](docs/plans/IMPLEMENTATION-PIPELINE.md).

---

## 4. Testing Discipline & Isolation Rules

- **Isolated environments**: Any test modifying global process state (e.g. `os.Chdir`, environment variables) must restore original state via `t.Cleanup(func() { ... })`.
- **Contract parity**: When introducing or modifying test doubles (e.g. `MemFileSystem`), run the double through the exact same contract test suite that exercises the OS implementation (`TestFileSystemContract`).
- **Table-driven tests**: Complex logic (parsers, validators, path sandboxing) must use table-driven tests with descriptive subtest names (`t.Run(tc.name, ...)`).
- **Golden file updates**: Code generation tests must compare output byte-for-byte against committed golden files in `testdata/golden/`.

---

## 5. CLI Conventions & Exit Codes

Every command must honor standard exit codes (see [Doc 09](docs/09-CLI-Developer-Experience.md)):
- `0`: Success
- `1`: Command or business logic failure
- `2`: Argument or usage error (e.g. unknown flag, missing required argument)
- `3`: Internal system error / panic

Output flags:
- `--json`: Machine-readable JSON output stream (schema: `[]Diagnostic` on error per [Doc 14](docs/14-Error-Diagnostics.md)).
- `--no-color`: Suppress ANSI color sequences.
- `-C`, `--directory`: Change working directory before command execution.

---

## 6. Documentation Index & Context Pointers

Consult these authoritative documents on demand:

- **Master Specification Index**: [docs/00-INDEX.md](docs/00-INDEX.md)
- **Technical Design Specification**: [docs/02-TDS.md](docs/02-TDS.md)
- **Technical Implementation Plan**: [docs/03-TIP.md](docs/03-TIP.md)
- **Detailed Phase Plans (1-10)**: [docs/plans/README.md](docs/plans/README.md)
- **Implementation Pipeline**: [docs/plans/IMPLEMENTATION-PIPELINE.md](docs/plans/IMPLEMENTATION-PIPELINE.md)
- **Architectural Decision Records (ADR-001 - ADR-017)**: [docs/adrs/README.md](docs/adrs/README.md)
  - [ADR-001](docs/adrs/ADR-001-seams-over-components.md): Seams over components
  - [ADR-002](docs/adrs/ADR-002-minimal-runtime.md): Minimal runtime
  - [ADR-003](docs/adrs/ADR-003-explicit-wiring.md): Explicit wiring
  - [ADR-007](docs/adrs/ADR-007-plan-based-generation.md): Plan-based generation
  - [ADR-013](docs/adrs/ADR-013-codegen-engine-selection.md): Go source codegen engine (`text/template`)
  - [ADR-014](docs/adrs/ADR-014-managed-code-splicing-via-comment-regions.md): Managed code splicing via comment regions (`// loy:region:...`)
  - [ADR-015](docs/adrs/ADR-015-database-migration-and-sqlc-pipeline.md): Database migration and sqlc pipeline (embedded `goose`)
  - [ADR-016](docs/adrs/ADR-016-two-phase-architecture-enforcement-engine.md): Two-phase architecture enforcement engine (`loy check`)
  - [ADR-017](docs/adrs/ADR-017-built-in-process-supervision-for-loy-dev.md): Built-in process supervision (`loy dev`)

---

## 7. Governance & Making Changes

Per [Doc 22 (Traceability & Roadmap)](docs/22-Traceability-Roadmap.md), architectural changes must follow the governance loop:
```text
Idea → ADR in docs/adrs/ → Impact analysis → Spec update in docs/ → Implementation → Tests
```
Never modify codebase invariants or add framework dependencies without an accepted ADR in `docs/adrs/`.

### 7.1 Plan & Documentation Synchronization
- **Real-time Phase Plan updates**: Whenever an implementation step finishes and passes verification, update the corresponding phase plan (`docs/plans/0X-Phase-*.md`) status to `Completed` and mark definition of done checkboxes immediately. Never leave completed phases in `Ready for Implementation` status.
- **Pre-Review self-audit**: Before requesting or performing code review, verify code against all Section 2 invariants (path jail, zero shell interpolation, explicit constructor injection, journal pre-registration).

---

## 8. Definition of Done & Verification Commands

Before concluding any implementation phase or submitting changes, run these verification commands:

```bash
# 1. Code standards & syntax check
go vet ./...

# 2. Concurrency & race detector test suite
go test -v -race ./...

# 3. Coverage validation (target >= 80% on core internal packages)
go test -cover ./internal/... ./cmd/...

# 4. Binary compilation smoke test
go build -o bin/loy ./cmd/loy

# 5. Link integrity check across docs
python3 -c '
import os, re
for root, _, files in os.walk("docs"):
    for file in files:
        if file.endswith(".md"):
            p = os.path.join(root, file)
            for _, link in re.findall(r"\[([^\]]+)\]\(([^)]+)\)", open(p).read()):
                if not link.startswith("http") and not link.startswith("#"):
                    tgt = os.path.normpath(os.path.join(root, link.split("#")[0]))
                    if not os.path.exists(tgt):
                        print(f"BROKEN: {p} -> {link}")
'
```
