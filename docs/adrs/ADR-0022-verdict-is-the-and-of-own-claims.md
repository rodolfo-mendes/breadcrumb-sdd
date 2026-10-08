---
breadcrumb:
  id: ADR-0022
  type: ADR
  links:
    - constrained_by ADR-0017
    - constrained_by ADR-0019
    - constrained_by ADR-0021
---
# ADR-0022: A breadcrumb's verdict is the AND of its own claims

Status: Accepted
Date: 2026-10-07

## Context

A claim is a statement about the repository's files that a tool can
check (ADR-0017), and each kind says when a claim of that kind holds
(ADR-0019). A breadcrumb can make several claims, or none, and one of
its claims can have been dropped for a problem (ADR-0021).

A reader wants one answer for each breadcrumb: do the files still
support what it says? Until now nothing checked a claim, and every
breadcrumb of the new model counted as Confirmed.

Breadcrumbs also link to each other, and a break in one could be
meant to show in the breadcrumbs that link to it. Which verbs should
carry a break is not known yet.

## Decision

- A verdict is one of three values: `Confirmed`, `Refuted` or
  `Undecided`.
- A claim is Confirmed when it holds, as its kind defines, and
  Refuted when it does not. No claim of the kind `has-line` is
  Undecided.
- A problem that belongs to a breadcrumb (ADR-0021) counts as one
  Undecided claim of that breadcrumb.
- A breadcrumb's verdict comes from its own claims and problems:
  - Refuted, when at least one of its claims is Refuted;
  - otherwise Undecided, when at least one is Undecided;
  - otherwise Confirmed, when it has at least one claim.
- A breadcrumb with no claim and no problem is Undecided, with the
  reason `no claims`.
- No link carries a verdict from one breadcrumb to another.

## Consequences

- A breadcrumb is Confirmed only when something was checked, and
  everything checked holds.
- A breadcrumb with a dropped claim is never Confirmed. One Refuted
  claim still makes it Refuted, whatever else is wrong with it.
- Most breadcrumbs are Undecided today: no ADR and no PBI of this
  repository makes a claim. Undecided therefore cannot mean failure,
  and a tool that fails on it would always fail.
- A spec whose claims hold stays Confirmed when an ADR it follows is
  Refuted. A reader who wants the whole picture reads the verdicts of
  the breadcrumbs it links to.
- A later kind of claim may have a claim that is Undecided, such as
  one about a report that was not written; the rule above already
  gives its breadcrumb a verdict.

## Alternatives Considered

- The plain AND, where a breadcrumb with no claims is Confirmed: one
  rule with no special case, but a breadcrumb reads as checked when
  nothing was.
- No verdict for a breadcrumb with no claims: nothing misleading is
  printed, but every reader of verdicts has to handle a breadcrumb
  that has none.
- A problem that makes its breadcrumb Refuted: stricter, but a typo
  in one claim would read as drift in the code, which it is not.
- Verdicts that travel along links now: closer to how a break
  spreads, but the verbs that carry it would be chosen with no real
  case to test them.
