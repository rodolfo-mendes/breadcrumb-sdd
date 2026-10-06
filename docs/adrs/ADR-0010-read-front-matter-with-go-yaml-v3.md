---
breadcrumb:
  id: ADR-0010
  type: ADR
  links: []
---
# ADR-0010: Front matter is read with go.yaml.in/yaml/v3

Status: Accepted, amended by ADR-0013
Date: 2026-10-03

## Context

Breadcrumbs live in YAML front matter (ADR-0002), so `bcr` has to
read YAML. Reading files is infrastructure, which may use a library
(ADR-0007), and each library `bcr` requires directly enters through an
ADR (ADR-0005).

The YAML library most Go code imports, `gopkg.in/yaml.v3`, was marked
unmaintained by its author in April 2025. The YAML organization took
over its maintenance under the module path `go.yaml.in/yaml`. There,
v3 is kept as it is and receives security fixes only, while new work
happens in v4, which has not reached a stable release yet.

`bcr` needs little from YAML: one map, plain values, a list of plain
values and `[]` (ADR-0002), read with the line each one is on.

## Decision

`bcr` reads front matter with `go.yaml.in/yaml/v3`, at a stable
release pinned in `go.mod`.

## Consequences

- `bcr` decodes front matter into the library's node tree, never into
  Go values, so the library's reading of a value as a number, a date
  or a boolean decides nothing. Each value is taken as the text
  written in the file, with its line.
- Before the core sees it, infrastructure turns the node tree into
  `bcr`'s own types, which the core defines (ADR-0007). The core
  imports no YAML library, and checks ADR-0002's part of YAML itself:
  anchors, aliases, tags, multi-line values and flow collections other
  than `[]` are violations.
- Splitting the front matter from the rest of the file is `bcr`'s own
  code, not the library's.
- The library is pure Go, so `bcr` stays one static binary
  (ADR-0004).
- Moving to `go.yaml.in/yaml/v4` once it has a stable release
  replaces one module with another, which needs an ADR (ADR-0005).

## Alternatives Considered

- `gopkg.in/yaml.v3`: the same code, under a path its author marked
  unmaintained.
- `go.yaml.in/yaml/v4`: where new work happens, but only release
  candidates exist, and its API is still changing between them.
- `github.com/goccy/go-yaml`: maintained and stable, but a different
  API and a different reading of YAML from the line most Go code
  uses, for no gain in a format this small.
- A parser of our own for ADR-0002's part of YAML: no dependency, but
  YAML's rules for quoting, indentation and comments are easy to get
  wrong, even in a small part of it.