---
title: "loy make agent-rules"
description: "Scaffold authoritative AI agent rules (Cursor, Claude, Copilot, Windsurf)"
slug: reference/cli/loy-make-agent-rules
sidebar:
  order: 11
---


Scaffold authoritative AI agent rules (Cursor, Claude, Copilot, Windsurf)

### Synopsis

loy make agent-rules scaffolds authoritative AI agent instruction documents
enforcing Clean Architecture layer boundaries, managed comment region safety, and verification gates.

```
loy make agent-rules [name] [flags]
```

### Options

```
      --client string   Alias for --target (default "all")
      --for string      Alias for --target (default "all")
  -h, --help            help for agent-rules
      --target string   Target AI assistant (all, cursor, claude, copilot, windsurf) (default "all")
```

### Options inherited from parent commands

```
  -C, --directory string   Change execution directory
      --dry-run            Preview generated operations without writing to disk
      --force              Overwrite existing files if developer owned
      --json               Output results in JSON format
      --modular            Scaffold sub-domain modular wiring file instead of flat wiring
      --no-color           Disable colored ANSI output
  -q, --quiet              Suppress non-essential output
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy make](/loy/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

