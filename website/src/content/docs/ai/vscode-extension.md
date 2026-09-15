---
title: "Official VS Code Extension"
description: "Real-time Clean Architecture linting, inline problem diagnostics, and automated code actions with the Loy VS Code extension."
---

The official **Loy VS Code Extension** integrates Loy's architectural validation engine directly into your editor, giving you and your AI coding assistants live, real-time feedback as code is written.

---

## Key Features

- **Live Architectural Linting**: Runs `loy check --format json` on file save or on demand, mapping violations (`ARCH-001` through `ARCH-015`) directly onto editor lines.
- **VS Code Problems Panel Integration**: Displays severity, rule ID, violation rationale, and remediation hints natively in your Problems tab.
- **Quick Fix Code Actions**: Offers one-click suggestions to invert illegal dependencies, extract interfaces, or clean up prohibited imports.
- **Visual Architecture Graphs**: Generates and previews active Mermaid architecture diagrams right inside the editor.
- **Workspace Trust Enforced**: Honors VS Code workspace trust settings to guarantee subprocess executions are strictly sandboxed.

---

## Installation & Setup

### From Repository Source

```bash
cd extensions/vscode
npm install
npm run compile
```

You can run the extension directly in development mode by pressing `F5` in VS Code or package it into a `.vsix` bundle using `vsce package`.

### Configuration Settings

In your `settings.json`:

```json
{
  "loy.path": "loy",
  "loy.checkOnSave": true,
  "loy.deepAnalysis": false
}
```

| Setting | Default | Description |
|---|---|---|
| `loy.path` | `"loy"` | Path to the `loy` CLI binary on your system `$PATH`. |
| `loy.checkOnSave` | `true` | Automatically run architectural checks whenever a Go file is saved. |
| `loy.deepAnalysis` | `false` | Enable Phase 2 deep type checking via `go/packages` (slower, but inspects concrete type assignments). |

---

## Editor Commands

Access these commands from the Command Palette (`Cmd+Shift+P` or `Ctrl+Shift+P`):

- **`Loy: Validate Architecture Boundaries`**: Manually runs `loy check` across the entire workspace.
- **`Loy: Run Environment Doctor`**: Executes `loy doctor` to verify Go compiler, Git, Docker, and toolchain readiness.
- **`Loy: Show Architecture Dependency Graph (Mermaid)`**: Renders your active package dependency hierarchy and flags boundary violations visually.
