---
title: "loy check"
description: "Validate architectural rules and boundaries"
slug: reference/cli/loy-check
sidebar:
  order: 1
---


Validate architectural rules and boundaries

### Synopsis

Validates that Go code strictly adheres to Loy architectural invariants, layer boundaries, and workspace topology.

```
loy check [path] [flags]
```

### Options

```
      --deep            Run deep type analysis via go/packages
      --format string   Output format (text, json, github, agent) (default "text")
  -h, --help            help for check
      --strict          Treat all architectural warnings as fatal errors
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

