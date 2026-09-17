---
title: "loy self-update"
description: "Update the Loy CLI binary to the latest or specified release"
slug: reference/cli/loy-self-update
sidebar:
  order: 66
---


Update the Loy CLI binary to the latest or specified release

### Synopsis

loy self-update checks GitHub Releases for the latest verified binary matching your operating system and architecture,
validates the SHA256 cryptographic checksum against release signatures, and atomically replaces the running executable.

```
loy self-update [flags]
```

### Options

```
      --check-only       Check if an update is available without downloading or applying
  -h, --help             help for self-update
      --version string   Target specific release version tag (e.g. v0.3.0)
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

