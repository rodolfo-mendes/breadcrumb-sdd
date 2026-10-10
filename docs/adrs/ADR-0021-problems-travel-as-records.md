---
breadcrumb:
  id: ADR-0021
  type: ADR
  title: Problems travel down the pipe as records
  links:
    - amends ADR-0016
    - constrained_by ADR-0011
    - constrained_by ADR-0017
---
# ADR-0021: Problems travel down the pipe as records

Status: Accepted
Date: 2026-10-07

## Context

`bcr extract` and `bcr verify` print each problem they find to
standard error, as `PATH:LINE: MESSAGE` (ADR-0006). Standard error
does not travel down a pipe, so the stages after them cannot see a
problem:

```
bcr extract | bcr verify | bcr audit | bcr report
```

Two stages need to. A claim with a problem is dropped on its own, and
its breadcrumb is still printed (ADR-0017): `bcr audit` would judge
that breadcrumb by the claims that are left, and could call it
Confirmed while one of its claims was never read. And `bcr report`
is meant to show problems and verdicts on one page.

ADR-0016 says `bcr verify` adds no record, and leaves this question
to `bcr audit`, the first stage that reads problems.

## Decision

- A problem `bcr extract` or `bcr verify` finds is printed as a
  record on standard output, as well as on standard error:

  ```
  problem	PATH	LINE	MESSAGE
  ```

  `PATH`, `LINE` and `MESSAGE` are the three parts of the line
  written to standard error.
- A message that starts `bcr: `, such as for a file that cannot be
  read, is not a problem and prints no record.
- `bcr verify` still copies its input to standard output, byte for
  byte (ADR-0016). It then adds a record for each problem it found
  itself. This replaces ADR-0016's rule that it adds no record; it
  still changes none.
- A problem belongs to the breadcrumb whose record has the same
  `PATH`. A problem whose `PATH` no breadcrumb record has belongs to
  none.
- A stage passes on the problem records it reads, like any other
  record.

## Consequences

- A stage after `bcr extract` can tell that a claim was dropped, and
  from which file.
- A problem in a file whose breadcrumb could not be read has no
  breadcrumb record beside it: it is passed on, for `bcr report` to
  show.
- `bcr verify`'s output is no longer equal to its input when it finds
  a problem. Its input is still the start of its output, and the two
  are equal when it finds none, so the check with `cmp` holds for a
  set with no problem.
- Each problem is now reported twice, on two streams. A caller who
  reads standard error sees what they saw before.
- A problem record has no id: a problem can be found before an id is
  read. A consumer finds its breadcrumb by `PATH`.
- `MESSAGE` is one field on one line, so a tab or a line ending
  inside a message cannot be printed as it is.

## Alternatives Considered

- Problems on standard error only: no change to the records, but a
  breadcrumb with a dropped claim could read Confirmed, and
  `bcr report` could show no problem.
- `bcr audit` checking the rules again: it would see the problems,
  but it reads records, not front matter, and the rules would live
  in two commands.
- A problem record that carries the id of its breadcrumb: easier to
  join, but empty for a breadcrumb whose id could not be read, and a
  record's fields are never empty.
- Problem records from `bcr extract` only: ADR-0016 would stand as
  written, but a duplicated id or a link that points nowhere would
  reach `bcr report` by another way than every other problem.
