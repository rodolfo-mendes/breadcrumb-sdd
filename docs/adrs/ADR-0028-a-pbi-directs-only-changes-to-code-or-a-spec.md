---
breadcrumb:
  id: ADR-0028
  type: ADR
  title: A PBI directs only changes to code or a spec
  links:
    - constrained_by ADR-0001
    - constrained_by ADR-0026
---
# ADR-0028: A PBI directs only changes to code or a spec

Status: Accepted
Date: 2026-10-09

## Context

`CONTRIBUTING.md` asks for a PBI on every change. A PBI `changes` a
spec (ADR-0026), but many changes touch no spec: an ADR written before
its implementation, a fix to `README.md`, a rule added to `AGENTS.md`
or to `CONTRIBUTING.md` itself. For those, the PBI has no spec to name
and records nothing the commit body does not already say. The rule is
ceremony there, and the repository is meant to carry as little process
as it can.

## Decision

- A change that touches code or a spec starts with a PBI that
  `changes` each spec it touches.
- Any other change needs no PBI: an ADR on its own, or a change to
  prose outside `specs/`, such as `AGENTS.md`, `CONTRIBUTING.md`,
  `README.md` or `ARCHITECTURE.md`. Its commit type is `docs`, and its
  commit body records the judgment, as for any commit.

## Consequences

- A prose change leaves no breadcrumb. `git log` records when and why
  it was made.
- An ADR's own gate is its acceptance by a maintainer; a PBI adds
  nothing to it.
- A change to how the repository is built still touches
  `specs/asdlc/spec.md`, so it still comes with a PBI.

## Alternatives Considered

- A PBI for every change, as before: a PBI with no spec to name, for
  every edit of a guide.
- No PBIs at all: a change to code would lose its link to the feature
  it changes.
