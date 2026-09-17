---
title: "loy mcp"
description: "Run the Model Context Protocol (MCP) server over standard I/O"
slug: reference/cli/loy-mcp
sidebar:
  order: 50
---


Run the Model Context Protocol (MCP) server over standard I/O

### Synopsis

loy mcp starts a Model Context Protocol JSON-RPC 2.0 server over stdin/stdout.
External AI agents (Claude Desktop, Cursor, Kilo, Windsurf) can connect to this server
to discover architectural rules, run AST validations, and scaffold Clean Architecture slices.

```
loy mcp [path] [flags]
```

### Options

```
  -h, --help   help for mcp
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

