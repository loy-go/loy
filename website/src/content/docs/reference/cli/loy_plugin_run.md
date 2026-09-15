---
title: "loy plugin run"
description: "Execute an installed plugin generator command within the sandbox"
slug: reference/cli/loy-plugin-run
sidebar:
  order: 50
---


Execute an installed plugin generator command within the sandbox

```
loy plugin run <plugin-name> <command> [args...] [flags]
```

### Options

```
      --dry-run   Preview generated artifacts without writing to disk
      --force     Overwrite existing developer-owned files
  -h, --help      help for run
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

* [loy plugin](/loy/reference/cli/loy-plugin/)	 - Manage and execute sandboxed external community plugins and generators

