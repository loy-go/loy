---
title: "loy graph"
description: "Visualize package dependency hierarchy and architectural layers"
slug: reference/cli/loy-graph
sidebar:
  order: 6
---


Visualize package dependency hierarchy and architectural layers

### Synopsis

loy graph renders project package dependencies, architectural layers, and boundary violations in ascii, dot, or mermaid format.

```
loy graph [path] [flags]
```

### Options

```
      --diff string       Compare architectural dependencies against git ref (e.g. main, origin/main)
  -f, --format string     Graph output format (ascii, dot, mermaid, markdown) (default "ascii")
  -h, --help              help for graph
      --violations-only   Render only packages involved in architectural violations
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

