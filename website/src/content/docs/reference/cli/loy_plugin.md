---
title: "loy plugin"
description: "Manage and execute sandboxed external community plugins and generators"
slug: reference/cli/loy-plugin
sidebar:
  order: 47
---


Manage and execute sandboxed external community plugins and generators

### Synopsis

The plugin command tree allows installing, inspecting, and running sandboxed community generators.
Plugins execute out-of-process and cannot write to disk directly; all emitted artifacts pass path jail validation and atomic plan execution.

### Options

```
  -h, --help   help for plugin
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
* [loy plugin install](/loy/reference/cli/loy-plugin-install/)	 - Install a community plugin from a local directory or Git repository
* [loy plugin list](/loy/reference/cli/loy-plugin-list/)	 - List installed community plugins in the project
* [loy plugin run](/loy/reference/cli/loy-plugin-run/)	 - Execute an installed plugin generator command within the sandbox

