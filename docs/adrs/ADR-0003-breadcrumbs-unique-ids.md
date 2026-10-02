---
breadcrumb:
  id: ADR-0003
  type: ADR
  links: []
---
# ADR-0003: Breadcrumb ids are unique

Status: Proposed
Date: 2026-10-02

## Context

ADR-0002 gives each breadcrumb an `id` and lets links point to it,
but does not say that two breadcrumbs cannot share one. A link to an
id that two breadcrumbs carry points to neither of them for certain.

Ids are easy to duplicate by accident: two branches that each add
the next PBI pick the same number, and both arrive on the main
branch. Ids also tend to become file names, such as
`ADR-0001-adopt-asdlc.md`, and on the case-insensitive file systems
of macOS and Windows, `ADR-0001` and `adr-0001` name the same file.

## Decision

- No two breadcrumbs in one state of the repository have the same
  `id`.
- Two ids are the same when their bytes are equal.
- Two ids that differ only in letter case, compared with Unicode
  simple case folding, are a violation as well.
- When two breadcrumbs share an id, or differ only in case, each of
  them is reported, in its own file. Neither is taken as the right
  one.

## Consequences

- A duplicate id is fixed by changing the id of one of the
  breadcrumbs, and the links that meant it.
- Which breadcrumb a link to a duplicated id means is not decided;
  the repository is broken there until one id changes.
- An id used again after its breadcrumb was deleted is not a
  violation: nothing that reads one state of the repository can see
  it.

## Alternatives Considered

- Taking the first breadcrumb found with an id: a duplicate would
  disappear from view, and which one is first would depend on the
  order files are read.
- Treating ids that differ only in case as the same id: a link could
  then name a breadcrumb in either spelling, which makes two
  spellings of one id valid instead of reporting the second.
- Ids unique within each type, so that a spec and an ADR could share
  one: a link names only an id, so it could not tell them apart.
- Ids unique across the repository's history: it cannot be checked
  from one state of the repository.