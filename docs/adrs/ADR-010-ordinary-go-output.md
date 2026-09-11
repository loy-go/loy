# ADR-010: Ordinary Go Output

## Status
Accepted

## Context
Code generators that produce obfuscated, non-idiomatic, or proprietary code structures make maintenance, debugging, and code reviews difficult.

## Decision
All Loy-generated code is formatted with `gofmt`, follows standard Go idioms (effective Go), uses standard library idioms where possible, and reads as if handwritten by an experienced Go engineer.

## Consequences
- No magic code or obfuscated generated artifacts.
- Teams can easily inspect, modify, and own generated files.
- Linter and static analysis friendly out of the box.
