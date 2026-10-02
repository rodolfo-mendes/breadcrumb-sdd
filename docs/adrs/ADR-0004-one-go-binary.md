---
breadcrumb:
  id: ADR-0004
  type: ADR
  links: []
---
# ADR-0004: bcr is one Go binary

Status: Accepted
Date: 2026-10-02

## Context

`bcr` should run on a developer's own machine, in a repository of any
stack and on any operating system, with nothing installed but `bcr`
itself. A team working in Java or .NET on Windows may have no Python
or Node, and should not need one to run it.

## Decision

`bcr` is written in Go, and is built as one static binary per
operating system.

## Consequences

- Running `bcr` needs nothing installed besides the binary.
- A Go module that needs cgo, Go's bridge to C code, would break the
  static build on some systems, so taking one means revisiting this
  ADR.

## Alternatives Considered

- Python: needs an interpreter, which many machines, Windows ones
  above all, do not have.
- Rust: one static binary too, but slower to write, and harder to read
  for people and agents.
- TypeScript on Node: needs a runtime.
- Bash: hard to read and to test as the model grows, and does not run
  on Windows without extra tools.