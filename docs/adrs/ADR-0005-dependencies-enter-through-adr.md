---
breadcrumb:
  id: ADR-0005
  type: ADR
  title: Each dependency enters through an ADR
  links: []
---
# ADR-0005: Each dependency enters through an ADR

Status: Accepted
Date: 2026-10-02

## Context

ADR-0002 gave breadcrumbs front matter, so reading it now means either
writing a YAML parser or taking one.

Parsing files and command-line flags is infrastructure around what
`bcr` checks, not part of it, and a library there does not make the
method depend on code it does not control. Every dependency is still
code this repository must trust.

## Decision

`bcr` may depend on Go modules outside the standard library. Each
module `bcr` requires directly, including one used only by tests, is
introduced by an ADR that names it and gives the reason.

## Consequences

- Every module in the `require` block of `go.mod` not marked
  `// indirect` is named in an ADR. That can be checked from one state
  of the repository.
- Upgrading a module to a new version needs no ADR; the change to
  `go.sum` shows it. Replacing a module with another does.
- Modules required only through another module come with it; the ADR
  of the module that requires them covers them.
- Whether the core of `bcr`, the part that defines Breadcrumb, may use
  a dependency is left to the ADR that draws its boundary.
- `CONTRIBUTING.md` and `AGENTS.md` still describe the standard
  library as the only dependency, and change once this ADR is
  accepted.

## Alternatives Considered

- Only the standard library: every parser written here,
  including one for YAML, where a maintained library exists.
- Dependencies with no record: `go.mod` shows what `bcr` depends on,
  not why, and each library looks harmless on its own.