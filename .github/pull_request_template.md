## Description

Provide a clear and concise summary of the changes in this pull request and the problem they solve.

Fixes #(issue)

---

## Type of Change

- [ ] `feat`: A new feature, generator, or command
- [ ] `fix`: A bug fix or error diagnostic correction
- [ ] `perf`: Performance improvement
- [ ] `refactor`: Internal code restructuring without behavioral change
- [ ] `docs`: Documentation or ADR updates
- [ ] `test`: New or updated tests / golden fixtures
- [ ] `chore`: Toolchain, dependency, or CI maintenance

---

## Architectural Invariants Checklist

Every pull request must uphold Loy's non-negotiable architectural invariants:

- [ ] **Seams over components**: No proprietary web framework, ORM, or reflection DI engine introduced ([ADR-001](docs/adrs/ADR-001-seams-over-components.md)).
- [ ] **Zero runtime dependency**: Generated applications remain pure Go with zero runtime framework dependencies ([ADR-002](docs/adrs/ADR-002-minimal-runtime.md)).
- [ ] **Explicit wiring**: Wiring occurs via explicit constructor injection in `internal/app/wiring.go` ([ADR-003](docs/adrs/ADR-003-explicit-wiring.md)).
- [ ] **Path sandboxing**: All file read/write operations pass `filesystem.CleanAndValidatePath` ([Doc 16](docs/16-Security.md)).
- [ ] **Zero shell interpolation**: Subprocesses are invoked with separate arguments via `exec.CommandContext`. Never `sh -c`.
- [ ] **Stream purity**: No raw `fmt.Print*` in subcommands or helpers; respects `--json` and `--quiet`.

---

## Contributor Quality Verification

- [ ] `make vet` runs cleanly without warnings
- [ ] `make lint` passes with 0 issues
- [ ] `make test` passes fast unit tests with race detector (`-race`)
- [ ] `make test-all` passes all tests including CLI integration
- [ ] `make build` compiles clean binary and `./bin/loy doctor` passes
- [ ] Golden fixtures updated if templates modified (`make golden`)
- [ ] Documentation updated in `docs/` or `README.md` if applicable
