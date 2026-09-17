---
title: "loy gen client"
description: "Generate zero-dependency TypeScript types and API client from Go transport DTOs"
slug: reference/cli/loy-gen-client
sidebar:
  order: 7
---


Generate zero-dependency TypeScript types and API client from Go transport DTOs

### Synopsis

loy gen client statically analyzes Go request/resource DTOs and registered routes to generate types.ts and client.ts.

```
loy gen client [root-dir] [flags]
```

### Options

```
  -h, --help         help for client
  -o, --out string   Output directory for generated TypeScript files (default "client")
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

* [loy gen](/loy/reference/cli/loy-gen/)	 - Generate client SDKs, types, and external schemas from application code

