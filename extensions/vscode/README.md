# Loy VS Code Extension

The official Visual Studio Code extension for the **Loy Developer Platform**.

## Features

- **Live Clean Architecture Enforcement**: Runs `loy check` automatically on Go file saves and populates the **Problems** panel with rule codes (`ARCH-001` through `ARCH-015`).
- **Quick Fixes (Code Actions)**: One-click guidance to invert layer dependencies and decouple web frameworks from domain services.
- **Command Palette Integration**:
  - `Loy: Check Architecture`
  - `Loy: Inspect HTTP Routes`
  - `Loy: View Architecture Dependency Graph`
  - `Loy: Start Development Server`

## Extension Settings

| Setting | Default | Description |
|---|---|---|
| `loy.path` | `"loy"` | Path to the `loy` CLI binary. |
| `loy.checkOnSave` | `true` | Automatically run `loy check` on Go file save. |
| `loy.deepAnalysis` | `false` | Enable deep type analysis using `go/packages`. |

## Requirements

The `loy` CLI binary must be installed and accessible via your `$PATH` (or configured via `loy.path`).
