---
breadcrumb:
  id: ADR-0007
  type: ADR
  links: []
---
# ADR-0007: The core domain owns Breadcrumb's rules

Status: Accepted
Date: 2026-10-02

## Context

`bcr` is the reference implementation of Breadcrumb. The method is
not all of `bcr`, only its core domain: the rules about what a
repository must contain and what it is judged on. The specification
is extracted from that core. How `bcr` reads a repository and talks
to whoever runs it is infrastructure around the core, and another
implementation could do it differently.

## Decision

- Breadcrumb's rules live in the core domain of `bcr`: Go packages
  that hold those rules and nothing else. Everything else in `bcr` is
  infrastructure, including its commands, the report, reading files,
  and parsing flags and YAML.
- Infrastructure uses the core. The core uses no infrastructure, and
  imports only Go's standard library.
- The core reads the repository through Go's `io/fs` interfaces, and
  opens no file itself.
- Every problem is found by the core, with the number of the line
  where it is, counted from 1. Commands and the report show the
  problems the core finds, and find none of their own.
- The specification describes the core only. How `bcr` is run, its
  flags, its output and its exit status are in `docs/bcr.md`
  (ADR-0006).

## Consequences

- Whether the core holds to this can be checked from one state of the
  repository: `go list -f '{{.Imports}}'` on its packages lists only
  the standard library and other core packages.
- A library's reading of a value cannot decide a rule. Whatever
  infrastructure hands the core keeps each value as the text written
  in the file, with its line, so `id: 0001` reaches the core as
  `0001`, not as a number.
- Rules in the front matter, such as ADR-0002's small part of YAML,
  are checked by the core, even when a library does the parsing.
- The core's tests run on file systems held in memory.
- A rule that is not in the core is not part of Breadcrumb.
- The core's packages are named by the work that creates them.

## Alternatives Considered

- One package for everything: shorter, but the definition of the method
  would be spread through code that has nothing to do with it.
- Problems found by each command: two readers of one format, drifting
  apart.
- A core that may use libraries, like the rest of `bcr`: a library's
  behavior could then become part of the method, which would depend
  on code it does not control.
- A core that opens files itself: its tests would need real files,
  and reading files is infrastructure.