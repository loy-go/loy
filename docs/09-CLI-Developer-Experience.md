# Loy — CLI & Developer Experience Specification

**Version:** 1.0  
**Status:** Locked  
**Navigation:** [← 08 Runtime Lifecycle Spec](./08-Runtime-Application-Lifecycle.md) | [Index](./00-INDEX.md) | [10 Configuration Manifest Spec →](./10-Configuration-Manifest.md)

---

## Command Tree

```text
loy
├── new
├── init
├── make
│   ├── feature
│   ├── model
│   ├── service
│   ├── repository
│   ├── handler
│   ├── request
│   ├── resource
│   ├── job
│   ├── event
│   ├── listener
│   ├── policy
│   ├── crud
│   └── test
├── generate
├── build
├── test
├── dev
├── migrate
├── queue
├── check
├── doctor
├── graph
├── upgrade
└── version
```

## Common Flags

Where meaningful: `--json`, `--quiet`, `--verbose`, `--dry-run`, `--force`, `--directory`, `--non-interactive`, `--no-color`.

## `make` vs `generate`

`make` creates a new artifact. `generate` reconciles managed artifacts.

## Discovery

Explicit directory wins, followed by current directory and parent discovery. `loy.yaml` is the Loy project marker; `go.mod` and `go.work` remain authoritative for Go.

## Output

Human output is the default. JSON output must be valid JSON even on failure. Colors are disabled automatically when output is not a TTY.

## Exit Codes

```text
0 success
1 command/project failure
2 usage/argument error
3 internal error
```

## Safety

No hidden destructive operations. `--force` cannot bypass security invariants.

## CI

No interactive prompts. Non-zero failure codes must be reliable under shell automation.

---

**Next:** [10-Configuration-Manifest.md — Configuration & Manifest Specification](./10-Configuration-Manifest.md)
