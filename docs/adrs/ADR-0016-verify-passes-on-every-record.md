---
breadcrumb:
  id: ADR-0016
  type: ADR
  title: bcr verify passes on every record it reads
  links:
    - amends ADR-0015
    - constrained_by ADR-0011
---
# ADR-0016: bcr verify passes on every record it reads

Status: Accepted, amended by ADR-0021
Date: 2026-10-07

## Context

`bcr`'s whole process is meant to run as one pipe:

```
bcr extract | bcr verify | bcr audit | bcr report
```

Today `bcr verify` reads the records `bcr extract` prints (ADR-0015)
and prints nothing to standard output, so a stage after it receives
nothing. The next stages need every record: `bcr audit` needs the
breadcrumbs, and the claims that `bcr extract` will print as records
of a new kind (ADR-0011).

ADR-0011 leaves each command that prints records to decide its own
output.

## Decision

- `bcr verify` copies its standard input to its standard output, byte
  for byte: every record, in the order read, with its line ending as
  read, records of kinds it does not know included.
- It adds no record and changes none. Its problems stay on standard
  error, as `PATH:LINE: MESSAGE` (ADR-0006).
- When it finds a problem and exits 1, it still copies every record.
- When a line of its input is not a record and it exits 2, it prints
  nothing to standard output.

## Consequences

- A stage after `bcr verify` receives what `bcr extract` printed,
  so `bcr verify` can sit anywhere before `bcr audit` without hiding a
  record.
- Whether `bcr verify` changed its input can be checked by hand:

  ```
  bcr extract > a.tsv
  bcr verify < a.tsv | cmp - a.tsv
  ```

- A breadcrumb with a problem reaches the next stage as a valid
  record; the problem is only on standard error. Whether problems also
  travel down the pipe as records is decided with `bcr audit`, its
  first reader.
- `bcr extract | bcr verify` on a terminal now prints the records. A
  caller who wants only the problems sends standard output to
  `/dev/null`.
- `bcr verify` reads its whole input before it writes, since it cannot
  know whether a later line is not a record. A stage after it starts
  only once its input has ended.

## Alternatives Considered

- Copying the records `bcr verify` understood, rewritten from what it
  read: `\r\n` would become `\n` and unknown kinds would need their own
  rule, and the output would no longer be checkable with `cmp`.
- Copying every line as it is read: a stage after it starts sooner,
  but when a later line is not a record, it has already received part
  of a set it should not judge.
- A flag that turns the copy on: two outputs to document and test, and
  a pipe that breaks when the flag is forgotten.
- `bcr audit` reading `bcr extract`'s records and `bcr verify`
  running beside it: the pipe would fork, which a shell pipe cannot
  do without saving the records first.
