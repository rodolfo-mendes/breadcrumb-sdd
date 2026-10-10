---
breadcrumb:
  id: ADR-0001
  type: ADR
  title: Adopt ASDLC to develop Breadcrumb
  links: []
---
# ADR-0001: Adopt ASDLC to develop Breadcrumb

Status: Accepted
Date: 2026-09-29

## Context

This repository has been developed with its own workflow and
artifacts. Designing that workflow, and a vocabulary for
it, has taken as much work as the software it records.

What is new in Breadcrumb is the record: how a repository's artifacts
link, what they claim about its code, and a check of those claims
against the files. The workflow around the record is not new. ASDLC
(asdlc.io) describes one built from artifacts most teams already
know.

## Decision

This repository is developed with ASDLC, following its pattern and
practice pages as of its release v0.27.0. It uses four of its
patterns, as those pages describe them: `AGENTS.md`, The Spec, The
PBI and The ADR.

Where ASDLC's pages disagree with each other, the pattern page wins.

Where this repository departs from ASDLC or goes beyond it, its spec
of how the repository is built says so.

ASDLC is how this repository is built. It is not part of Breadcrumb.

## Consequences

- Decisions about how this repository is built are ADRs from here on.
- "Spec" now names two things: the method specification, which is
  the product, and ASDLC's feature specs.
- Until the method specification moves past Intakes, Requirements,
  Technical Decisions and Tasks, the repository speaks two
  vocabularies: ASDLC's for how it is built, and the method's for
  what it builds.
- The repository follows a knowledge base it does not control, and
  that keeps changing. Moving to a later release of ASDLC supersedes
  this ADR.

## Alternatives Considered

- Keep this repository's own workflow: the work on its vocabulary
  keeps competing with the software, and keeps the workflow far from
  what teams already use.
- Make ASDLC part of Breadcrumb: ties the layer to one framework,
  when it is meant to sit over any.
- Adopt all of ASDLC at once: most of it has no use here yet.
- A framework with its own toolkit, such as Spec Kit: a fixed
  workflow and tools to learn, where ASDLC offers patterns to follow
  with the tools already here.
