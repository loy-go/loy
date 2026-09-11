# Phase 1: Foundation Implementation Plan

**Phase:** 1 of 10  
**Status:** Completed  
**Estimated Scope:** Core CLI plumbing, OS abstractions, structured error reporting  
**Primary Specifications:** [02-TDS.md](../02-TDS.md), [09-CLI-Developer-Experience.md](../09-CLI-Developer-Experience.md), [14-Error-Diagnostics.md](../14-Error-Diagnostics.md), [16-Security.md](../16-Security.md)

---

## 1. Goal & Objectives
Establish the rock-solid, zero-dependency foundation for the Loy CLI toolchain:
- Memory-safe, sandboxed filesystem abstraction.
- Subprocess execution runner with stream capture and timeouts.
- Standardized, structured diagnostic error system (`LOY*`).
- Base CLI runner using `spf13/cobra` with version command and flag conventions.

---

## 2. Package Architecture & Internal Seams

```text
internal/
├── filesystem/
│   ├── filesystem.go       # FS interface (Read, Write, Exists, MkdirAll, Stat, Walk)
│   ├── os_fs.go            # Production OS filesystem implementation
│   ├── mem_fs.go           # In-memory filesystem for isolated unit tests
│   └── path.go             # Path sanitization, jail checks (prevent directory traversal)
├── process/
│   ├── runner.go           # ProcessRunner interface (Run, RunWithContext, Stream)
│   ├── exec_runner.go      # os/exec implementation with signal propagation
│   └── result.go           # Execution result (ExitCode, Stdout, Stderr, Duration)
├── diagnostics/
│   ├── diagnostic.go       # Diagnostic struct (Severity, Code, Message, Hint, File, Line)
│   ├── severity.go         # Enum: Info, Warning, Error
│   ├── codes.go            # Stable error code definitions (LOY-CLI-*, LOY-FS-*, etc.)
│   └── formatter.go        # Human-readable & JSON renderers
├── version/
│   ├── version.go          # Build-time injected version, commit, date, go-version
│   └── info.go             # Structured version payload
└── cli/
    ├── root.go             # Root cobra.Command, global flags (--json, --verbose, --quiet, --no-color)
    └── version.go          # loy version subcommand
cmd/
└── loy/
    └── main.go             # Binary entrypoint with os.Exit code mapping (0, 1, 2, 3)
```

---

## 3. Concrete Implementation Steps

### Step 1.1: Filesystem Abstraction (`internal/filesystem`)
1. Define `FileSystem` interface:
   - `ReadFile(path string) ([]byte, error)`
   - `WriteFile(path string, data []byte, perm os.FileMode) error`
   - `Exists(path string) (bool, error)`
   - `MkdirAll(path string, perm os.FileMode) error`
   - `Remove(path string) error`
   - `Walk(root string, fn filepath.WalkFunc) error`
2. Implement `osFS` wrapping standard `os` package.
3. Implement `memFS` using `map[string][]byte` for deterministic fast tests.
4. Implement path jail utility: `EnsureWithinRoot(root string, target string) (string, error)` preventing path traversal attacks (`../`).

### Step 1.2: Subprocess Runner (`internal/process`)
1. Define `Runner` interface:
   - `Run(ctx context.Context, dir string, name string, args ...string) (*Result, error)`
   - `RunWithInput(ctx context.Context, dir string, in io.Reader, name string, args ...string) (*Result, error)`
2. Ensure no direct string shell interpolation (`sh -c`) to prevent command injection.
3. Enforce context cancellation: context timeout cleanly kills child process with `SIGKILL` fallback.

### Step 1.3: Diagnostics & Structured Errors (`internal/diagnostics`)
1. Implement `Diagnostic` model per Doc 14:
   ```go
   type Diagnostic struct {
       Severity Severity `json:"severity"`
       Code     string   `json:"code"`
       Message  string   `json:"message"`
       Detail   string   `json:"detail,omitempty"`
       Hint     string   `json:"hint,omitempty"`
       File     string   `json:"file,omitempty"`
       Line     int      `json:"line,omitempty"`
       Column   int      `json:"column,omitempty"`
   }
   ```
2. Implement Formatters:
   - `HumanFormatter`: Colored ANSI output when TTY attached; plain text when redirected or `--no-color` is set.
   - `JSONFormatter`: Valid JSON output streaming array of diagnostics.
3. Define Phase 1 error codes (`LOY-CLI-001` argument error, `LOY-FS-001` file not found, `LOY-PROC-001` process timeout).

### Step 1.4: CLI Core & Entrypoint (`cmd/loy`, `internal/cli`)
1. Setup Cobra root command with global flags:
   - `--json`: machine-readable output.
   - `--verbose`, `-v`: detailed debug output.
   - `--quiet`, `-q`: suppress non-essential output.
   - `--no-color`: strip ANSI color codes.
   - `--directory`, `-C`: execute as if launched in specified directory.
2. Implement `loy version` supporting `--json`.
3. Wire entrypoint in `cmd/loy/main.go` with deterministic exit codes:
   - `0`: Success
   - `1`: Project/Command failure
   - `2`: Usage/Argument error
   - `3`: Internal system panic/error

---

## 4. Test Strategy & Acceptance Criteria

### Unit Tests
- `filesystem_test.go`: Verify `memFS` and `osFS` behave identically.
- `path_test.go`: Test traversal vectors (`/app/../../etc/passwd`, symlinks outside root) return error.
- `runner_test.go`: Test process timeout cancellation, non-zero exit codes, stderr capture.
- `diagnostics_test.go`: Verify JSON serialization matches JSON schema even on empty fields.

### CLI Integration Tests
- Execute binary with `--version` -> returns version string, exit 0.
- Execute binary with `--version --json` -> returns valid JSON payload, exit 0.
- Execute invalid command -> returns code `2`.
- Verify `--no-color` emits zero ANSI escape sequences.

---

## 5. Definition of Done
- [x] `go vet ./...` and `golangci-lint` pass with zero warnings.
- [x] Test coverage for `internal/filesystem`, `internal/process`, and `internal/diagnostics` >= 90%.
- [x] `cmd/loy` builds cleanly into a single self-contained binary.

---

[Back to Plans Index](./README.md) | [Next: Phase 2 Plan →](./02-Phase-Project-System.md)
