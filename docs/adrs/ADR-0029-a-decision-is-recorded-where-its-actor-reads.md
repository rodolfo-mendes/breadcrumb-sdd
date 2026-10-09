---
breadcrumb:
  id: ADR-0029
  type: ADR
  links:
    - constrained_by ADR-0001
---
# ADR-0029: A decision is recorded where whoever must obey it reads

Status: Accepted
Date: 2026-10-09

## Context

Until now, most of the project's decisions were kept in a log outside
the repository. Many of them were later written again as ADRs; others
were rules for contributors, statements of what the product is, or
answers to open questions, and had no home in the repository. A
decision kept where the actor who must follow it does not look is not
followed: an agent filing a change reads `AGENTS.md`,
`specs/asdlc/spec.md` and `CONTRIBUTING.md`, not an outside log or a
closed issue.

Backlog items, open questions and flags move to GitHub Issues. Issues
hold work and debate well, but a closed issue drops out of sight and
is not part of a clone.

## Decision

A decision goes to the first home that fits:

1. It changes how `bcr` behaves or how the repository is built: an
   ADR.
2. It is a rule for making a change: `specs/asdlc/spec.md` states it,
   and `CONTRIBUTING.md` explains it.
3. It is a boundary an agent keeps on every task: `AGENTS.md`.
4. It says what the product is: `README.md`, and the header of
   `AGENTS.md` when agents need it.
5. It answers an open question without imposing a rule: the question's
   issue closes with the answer.
6. It governs only the maintainer's own way of working, or nothing any
   more: it is not recorded in the repository.

When two homes fit, the decision goes to the one its actor reads
first, and the other links to it. A debate happens in an issue, which
closes by naming the decision's home, such as "Decided in ADR-0030".

## Consequences

- No decision lives only in an issue.
- The outside log stops. Its entries still in force are moved by this
  order, one at a time; the rest stay in its archive.
- Issues are labelled `backlog`, `question` or `flag`, and
  `specs/asdlc/spec.md` lists them among the places where the
  repository goes beyond ASDLC.

## Alternatives Considered

- A decision log kept in the repository beside the ADRs: two records
  of decisions, which drift apart.
- Decisions kept in issues: out of sight once closed, outside a clone,
  and with no breadcrumb for `bcr` to check.
