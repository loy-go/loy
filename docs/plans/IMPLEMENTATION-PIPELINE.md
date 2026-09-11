# Loy Engineering Implementation Pipeline & Verification Loop

Standard protocol governing implementation of Phases 2 through 10. Derived from Phase 1 learnings.

---

## 1. Core Lessons Learned Across Phases

1. **No Global State**: Package-level globals (`cli.Globals`) violate architectural invariants and leak across tests. Context-scoped state or explicit struct passing is mandatory.
2. **Contract Parity**: Test doubles (`memFS`) must match OS behavior exactly (directory collisions, `SkipDir` pruning, file mode flags). Contract test suites must run against both.
3. **Test Isolation**: Tests mutating process state (`os.Chdir`, env vars) must register cleanup via `t.Cleanup` or use isolated temp roots.
4. **Binary-Level Smoke Tests & Compiler Acceptance**: Unit tests alone miss CLI execution wiring or missing imports in generated code. Every codegen and CLI phase must include compiled binary integration tests verifying exit codes (0, 1, 2, 3), stdout/stderr routing, `--json` stream purity, and real `go build` compilation of generated projects.
5. **Path Security**: All path operations must pass boundary validation before disk touches.
6. **Stream Purity**: Never call `fmt.Println` or write uncoordinated bytes to stdout/stderr in CLI subcommands or background checks. All CLI writes must use `cmd.OutOrStdout()` or check `--json` and `--quiet`.
7. **Pre-Warmed Engines**: Re-instantiating heavy template renderers or AST parsers on each generation pass creates unnecessary GC churn; cache shared engines as thread-safe singletons.
8. **Explicit Target Directory Subprocesses**: Tools executed by the CLI (e.g. `sqlc generate`, `go mod`) must execute explicitly within the resolved target module directory (`targetDir`), never defaulting blindly to `.` (the CLI invocation directory).

---

## 2. Standard 6-Step Implementation Loop

Every subsequent phase must execute these 6 stages in strict order:

```text
[1. Contract First]
       ↓
[2. Contract & Unit Tests]
       ↓
[3. Core Implementation]
       ↓
[4. Double / Mock Compliance]
       ↓
[5. Binary E2E & CLI Smoke]
       ↓
[6. Verification Release Gate]
```

### Stage 1: Contract & Interface Definition
- Define consumer-owned interfaces first (small, explicit).
- Define structured diagnostic error codes for the domain (`LOY-<DOMAIN>-*`).
- Define typed request/response models. No `any` or untyped maps in core pipelines.

### Stage 2: Contract & Unit Tests
- Write test matrix before or alongside implementation.
- Include failure paths: timeouts, missing inputs, corrupt formats, boundary violations.
- Ensure test isolation: no leaked file handles, goroutines, or working directories.

### Stage 3: Core Implementation
- Implement logic adhering to ADRs.
- Explicit constructor injection only. No singletons, no init-time magic.
- Validate paths via `filesystem.CleanAndValidatePath`.

### Stage 4: Test Double Compliance
- When in-memory doubles are used (`memFS`, mock runners), run identical contract test suites against both real and mock implementations.

### Stage 5: Binary E2E & CLI Verification
- Compile `bin/loy`.
- Test real command invocation via subprocess.
- Verify:
  - Exit code `0` on success.
  - Exit code `1` on command/business failure.
  - Exit code `2` on argument/usage error.
  - Formatter output: colored text on TTY, plain text on `--no-color`, structured JSON on `--json`.

### Stage 6: Verification Release Gate
Run non-negotiable checklist:
1. `go vet ./...` passes clean.
2. `go test -v -race ./...` passes with zero race conditions.
3. Coverage check: core packages meet target coverage (>= 80%).
4. All relative links and docs updated.

---

[Back to Plans Index](./README.md)
