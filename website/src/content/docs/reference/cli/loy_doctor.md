---
title: "loy doctor"
description: "Validate development environment, toolchain prerequisites, and project state"
slug: reference/cli/loy-doctor
sidebar:
  order: 4
---


Validate development environment, toolchain prerequisites, and project state

### Synopsis

loy doctor inspects host tools (Go, Git, Docker, sqlc) and project validity without modifying the system.

```
loy doctor [path] [flags]
```

### Options

```
  -h, --help     help for doctor
      --strict   Treat warnings as errors and exit with code 1
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --json               Output results in JSON format
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy](/loy/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX

