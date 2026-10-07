---
breadcrumb:
  id: ADR-0018
  type: ADR
  links:
    - amends ADR-0002
---
# ADR-0018: A key added under breadcrumb is optional

Status: Accepted
Date: 2026-10-07

## Context

ADR-0002 requires three keys under `breadcrumb`: `id`, `type` and
`links`, with `links: []` for a breadcrumb that has none. ADR-0017
adds a fourth, `claims`, and more may follow.

If each new key were required, every breadcrumb written before it
would become a problem the day the key appears, though nothing in it
changed.

## Decision

- `id`, `type` and `links` stay required, as ADR-0002 says.
- Every key added under `breadcrumb` after them is optional. A
  breadcrumb without it means what it meant before the key existed.
- A key written with no value, such as a bare `claims:`, is a
  problem. A list key with nothing in it is written `[]`.

## Consequences

- A breadcrumb written before a key existed stays valid, with every
  version of `bcr` that knows the key.
- A breadcrumb with no claims omits `claims`; `claims: []` means the
  same.
- `links` and `claims` behave differently: no links is written, no
  claims may be left out. The difference follows when each key came,
  not what it means.
- This is a rule of the breadcrumb format. The commands of `bcr` may
  still change while the project is not released.

## Alternatives Considered

- Each new key required, written empty when unused, as `links` is:
  one rule for every key, but every existing breadcrumb must be
  edited when a key appears, and a breadcrumb that is not becomes a
  problem.
- Deciding for each key when it comes: no rule to keep, but every key
  reopens the question, and an author cannot tell from the format
  which keys can be left out.
