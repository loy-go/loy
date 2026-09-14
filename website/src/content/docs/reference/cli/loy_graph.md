---
title: "loy graph"
description: "Visualize package dependency hierarchy and architectural layers"
slug: reference/cli/loy-graph
sidebar:
  order: 5
---


Visualize package dependency hierarchy and architectural layers

### Synopsis

loy graph renders project package dependencies, architectural layers, and boundary violations in ascii, dot, or mermaid format.

```
loy graph [path] [flags]
```

### Options

```
  -f, --format string     Graph output format (ascii, dot, mermaid) (default "ascii")
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

* [loy](/reference/cli/loy/)	 - Loy — Go developer platform with Laravel-like DX

