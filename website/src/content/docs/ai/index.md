---
title: "The Agent-Native Go Platform"
description: "Why AI coding agents hallucinate or drift in Go codebases, and how Loy acts as the deterministic substrate for 10x more reliable agentic workflows."
---

The software engineering landscape has fundamentally shifted: code is increasingly authored by autonomous AI agents—such as **Claude Code**, **Cursor**, **Copilot**, **Kilo**, and **Windsurf**—rather than manual keystrokes.

However, in large Go projects, external AI agents routinely run into severe friction:
1. **Architectural Drift**: Agents eagerly import persistence drivers (`database/sql`, `gorm`) directly into HTTP transport handlers or pure domain entities, destroying separation of concerns.
2. **Wiring Hallucinations**: Agents struggle to coordinate constructor dependencies across large codebases, inventing global variables or dynamic reflection containers (`uber/dig`) that fail at runtime.
3. **Destructive Splicing**: When updating existing code, agents accidentally overwrite developer logic, delete managed comment tags, or leave invalid Go syntax.

Loy is purposefully engineered to be the **standard agent-native Go platform**. Rather than attempting to become an LLM wrapper, Loy provides the **deterministic substrate** that makes external AI agents 10x more reliable.

```text
┌─────────────────────────────────────────────────────────────────────────┐
│              External AI Agent (Claude Code / Cursor / Windsurf)        │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │ JSON-RPC 2.0 (stdio)
                                     ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Loy Agent-Native Substrate                        │
│  ┌───────────────────────┬──────────────────────┬────────────────────┐  │
│  │ Model Context Protocol│ Self-Healing Loop    │ Universal Rules    │  │
│  │ (MCP) Server          │ Diagnostic Prompts   │ Context Generator  │  │
│  │ `loy mcp`             │ `loy check --format` │ `loy make rules`   │  │
│  └───────────────────────┴──────────────────────┴────────────────────┘  │
└────────────────────────────────────┬────────────────────────────────────┘
                                     ▼
      Deterministic Go 1.22+ Codebase (Clean Architecture Invariants)
```

---

## The 4 Core Architectural Invariants for Agents

External agents operating inside a Loy project are governed by non-negotiable architectural invariants:

### 1. Zero LLM Dependency in Binary
Loy will **never** require an OpenAI or Anthropic API key to function. Intelligence belongs to the developer's agent of choice; Loy provides the deterministic tooling, AST parser, and validation substrate.

### 2. Strict 4-Layer Dependency Direction
```text
Transport -> Application -> Domain <- Infrastructure
```
- **Domain**: Pure Go business entities and repository interfaces. Prohibited from importing persistence, web frameworks, or infrastructure.
- **Application**: Use cases and service orchestration. Consumes domain interfaces.
- **Infrastructure**: Concrete database adapters (`sqlc`, `pgx`, Redis). Implements domain interfaces.
- **Transport**: HTTP, gRPC, and WebSocket handlers. Translates DTOs and invokes application services.

### 3. Atomic Plan Generation
Every generative action produces an in-memory `plan.Plan` that passes path jail validation and conflict detection before touching disk ([ADR-007](/loy/adrs/)). No corrupt, half-written disk states.

### 4. Guarded Comment Regions
Agents may only inject code into managed comment regions (`// loy:region:...` and `// loy:endregion` / [ADR-014](/loy/adrs/)). Corrupted or missing markers trigger rule `ARCH-014` with auto-repair diagnostics.

---

## The 3 Pillars of Loy's Agentic Architecture

| Pillar | Capability | Developer Advantage |
|---|---|---|
| **[MCP Server](/loy/ai/mcp-server/)** | Native JSON-RPC 2.0 protocol server over `stdio` | Connect Claude Desktop, Cursor, and Kilo directly to run AST checks, inspect routes, and scaffold vertical slices. |
| **[Self-Healing Loop](/loy/ai/self-healing/)** | Structured diagnostics via `loy check --format agent` | Emits machine-parsable remediation prompts (`action`, `prompt`, `rationale`) enabling agents to self-correct in one shot. |
| **[Universal Rules](/loy/ai/rules-generator/)** | Turnkey generator `loy make agent-rules` | Automatically outputs `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, `.github/copilot-instructions.md`, and `.windsurfrules`. |

---

## Next Steps

- [Configure the MCP Server with Claude Desktop & Cursor](/loy/ai/mcp-server/)
- [Automate Agent Self-Healing with `--format agent`](/loy/ai/self-healing/)
- [Generate Universal AI Rules for Your Team](/loy/ai/rules-generator/)
- [Use the Official VS Code Extension for Live Linting](/loy/ai/vscode-extension/)
