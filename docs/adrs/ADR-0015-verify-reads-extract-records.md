---
breadcrumb:
  id: ADR-0015
  type: ADR
  links:
    - follows ADR-0011
---
# ADR-0015: bcr verify reads the records bcr extract prints

Status: Accepted
Date: 2026-10-04

## Context

`bcr verify` checks the rules about the whole set of a repository's
breadcrumbs: no two share an id (ADR-0003), and every link points to a
breadcrumb in the set. It needs every breadcrumb, with its links, and
the path and line where each was written.

`bcr extract` already reads files, checks the rules of their format
and of each breadcrumb, and prints what it finds as tagged records
(ADR-0011). Its files come from its operands, or from `.breadcrumbs`
(ADR-0014).

## Decision

- `bcr verify` reads tagged records from standard input, as
  `bcr extract` prints them, and checks the rules about the set of
  breadcrumbs they describe:

  ```
  bcr extract | bcr verify
  ```

- `bcr verify` reads no file that carries a breadcrumb. Which files
  make up the set is decided before it, by `bcr extract`'s operands or
  by `.breadcrumbs`.
- It uses the kinds of record it knows, and ignores the others
  (ADR-0011).
- A line that is not a record `bcr extract` could print is not a
  problem in a breadcrumb: `bcr verify` writes a message starting
  `bcr: ` to standard error and exits 2 (ADR-0006).
- A line of its input may end in `\n` or `\r\n`.

## Consequences

- `bcr verify` holds no reader of any format: it is the core, fed
  records (ADR-0013). A new format changes `bcr extract`, not
  `bcr verify`.
- A problem is reported at the path and line its record carries, so
  every record of `bcr extract` carries the line where it was written.
  `specs/extract/spec.md` adds that field at the end of each record.
- A breadcrumb `bcr extract` could not read prints no record, so
  `bcr verify` never sees it: links to its id are reported as pointing
  nowhere, though the file exists. `bcr extract` has already reported
  the file's own problem, which is the one to fix first.
- In a pipe, the shell returns the exit status of its last command, so
  a problem `bcr extract` found, with exit 1, is hidden by
  `bcr verify` exiting 0. A script that needs both runs with
  `set -o pipefail`, or saves the records first.
- Any stream of records can be verified: a saved file, or the output
  of `bcr extract` on a few files. The set judged is the records
  given; leaving some out changes the answer.
- `bcr verify` with no input on a terminal waits for it, as `sort` and
  `grep` do.

## Alternatives Considered

- `bcr verify` reading the files itself, through the same reader as
  `bcr extract`: one command, one exit status and no pipe, but two
  commands reading files, and the file set chosen in two places.
- `bcr verify` doing both, reading files when given operands and
  records otherwise: two ways to get one answer, and two kinds of
  input to document and test.
- `bcr extract` checking the whole set as well: one command, but it
  would no longer just produce data (`specs/extract/spec.md`), and a
  caller who wants only records would pay for every check.
