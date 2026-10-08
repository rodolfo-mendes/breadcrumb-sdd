---
breadcrumb:
  id: ADR-0026
  type: ADR
  links:
    - constrained_by ADR-0001
    - constrained_by ADR-0025
---
# ADR-0026: This repository uses the ASDLC layout

Status: Accepted
Date: 2026-10-08

## Context

This repository is built with ASDLC (ADR-0001): ADRs record
decisions, specs describe features, and PBIs carry changes. Its
breadcrumbs grew before any shape was declared, so the same relation
is written in more than one way. A spec `follows` an ADR, a PBI
`implements` an ADR and also `changes` the spec that the ADR already
constrains, and one ADR `changes` a spec.

ADR-0025 lets a repository declare its shape. This one says which
shape this repository declares.

## Decision

- The repository has three types of breadcrumb: `ADR`, `spec` and
  `PBI`.
- A link never points down: an ADR links only to ADRs, a spec only to
  ADRs, and a PBI only to specs.
- The links allowed are:

  | From | Verb | To | Says |
  |---|---|---|---|
  | ADR | `constrained_by` | ADR | the decision is taken within the other |
  | ADR | `amends` | ADR | the decision changes part of the other |
  | spec | `constrained_by` | ADR | the feature is built within the decision |
  | PBI | `changes` | spec | the change is to that feature |

- Only a spec carries claims.
- A verb of more than one word is written with `_`, as in
  `constrained_by`.
- The repository's `breadcrumb.rules` is:

  ```
  type	ADR
  type	spec
  type	PBI
  link	ADR	constrained_by	ADR
  link	ADR	amends	ADR
  link	spec	constrained_by	ADR
  link	PBI	changes	spec
  claims	spec
  ```

## Consequences

- `follows` becomes `constrained_by`, in specs and in ADRs.
- A PBI no longer links to an ADR. The decisions behind a change are
  reached through the spec it changes; a PBI still names them in its
  Context.
- An ADR no longer links to a spec. ADR-0014 loses `changes extract`.
- Every PBI changes at least one spec, and every spec is constrained
  by at least one ADR, by convention: `bcr verify` does not require a
  link (ADR-0025).
- `amends` is not part of ASDLC; `specs/asdlc/spec.md` lists it among
  the places where the repository goes beyond it.
- The front matter of Accepted ADRs changes with the verbs. Their
  decisions do not.
- A link from one spec to another, and a link for an ADR that
  supersedes another, are not allowed yet. Each comes with its own
  line in `breadcrumb.rules`, through an ADR.

## Alternatives Considered

- Keeping `follows` and `implements`: no file to change, but two
  verbs say one thing, and `implements` repeats what the spec's own
  links already say.
- Letting a PBI link to an ADR as well: more ways to reach a
  decision, and more links to keep true by hand.
- Claims on any type: a PBI is closed after merge and an ADR is not
  edited once accepted, so a claim on either could turn Refuted with
  no document left to change.
