---
title: "loy dev"
description: "Start multi-process live development server with hot reload"
slug: reference/cli/loy-dev
sidebar:
  order: 3
---


Start multi-process live development server with hot reload

### Synopsis

loy dev supervises application processes (API server, background workers) and reloads on file changes using native fsnotify watching.

```
loy dev [path] [flags]
```

### Options

```
      --api-only                Run only the API server process
      --debounce duration       File change debounce window (default 200ms)
      --grace-period duration   Grace period before sending SIGKILL (default 3s)
  -h, --help                    help for dev
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

