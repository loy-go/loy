---
title: "loy make agent-rules"
description: "Scaffold authoritative AI agent rules (Cursor, Claude, Copilot, Windsurf)"
slug: reference/cli/loy-make-agent-rules
sidebar:
  order: 13
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
      --dual-id            Scaffold dual identifier schema (BIGINT identity + UUID public)
      --force              Overwrite existing files if developer owned
      --json               Output results in JSON format
      --modular            Scaffold sub-domain modular wiring file instead of flat wiring
      --no-color           Disable colored ANSI output
      --no-tidy            Skip running go mod tidy after generation
  -q, --quiet              Suppress non-essential output
  -v, --verbose            Enable verbose/debug output
```

### SEE ALSO

* [loy make](/loy/reference/cli/loy-make/)	 - Scaffold application components, slices, and vertical features

