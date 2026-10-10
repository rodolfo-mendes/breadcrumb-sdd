---
breadcrumb:
  id: ADR-0020
  type: ADR
  title: A released kind of claim keeps its meaning
  links:
    - constrained_by ADR-0017
---
# ADR-0020: A released kind of claim keeps its meaning

Status: Accepted
Date: 2026-10-07

## Context

A claim's verdict comes from its kind: the kind says what the claim
checks (ADR-0017). Breadcrumbs are written once and read for a long
time, by every later version of `bcr`.

If a kind changed what it checks, every existing claim of that kind
would be judged by a rule its author never chose, and its verdict
could change with no edit to the breadcrumb or to its target.

## Decision

- Once a kind of claim is in a release of `bcr`, what it checks never
  changes.
- A different meaning is a new kind, with a new name. The old kind
  stays as it was.

## Consequences

- A kind's name acts as its version: claims carry no version field.
- A breadcrumb's verdict changes only when the breadcrumb or the files
  it claims change.
- A repository that uses a kind its `bcr` does not know gets a problem
  from `bcr extract` (ADR-0017), never a verdict computed by another
  meaning.
- A kind's name is chosen for good, so it should say what the kind
  checks: `has-line`, not `contains` (ADR-0019).
- Fixing a mistake in a released kind is a new kind too. Before a kind
  is released, it may still change.
- Whether a released kind can ever be removed is not decided here.

## Alternatives Considered

- A version field on each claim, such as `has-line@1`: a kind could
  change, and old claims keep the old meaning, but every claim carries
  a number, and `bcr` keeps every version of every kind.
- Kinds that may change at a major release of `bcr`: fewer kinds in
  the long run, but a repository's verdicts would change when it
  upgrades `bcr`, with nothing in its files to show why.
