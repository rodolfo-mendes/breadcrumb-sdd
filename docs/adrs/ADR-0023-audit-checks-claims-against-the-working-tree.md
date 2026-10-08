---
breadcrumb:
  id: ADR-0023
  type: ADR
  links:
    - constrained_by ADR-0016
    - constrained_by ADR-0019
    - constrained_by ADR-0022
---
# ADR-0023: bcr audit checks claims against the working tree

Status: Accepted
Date: 2026-10-07

## Context

`bcr audit` is the third stage of the pipe:

```
bcr extract | bcr verify | bcr audit | bcr report
```

It reads the records the stages before it print, checks each claim
against its target, and gives each claim and each breadcrumb a
verdict (ADR-0022). It is the second command, after `bcr extract`, to
read the repository's files.

Four things are left to decide: which state of a target it reads,
what it does with a target that is not a file, what it prints for
`bcr report`, and when it fails.

## Decision

- `bcr audit` reads each target as it is on disk, uncommitted changes
  included, as `bcr extract` reads breadcrumbs.
- A claim's `TARGET` is a path from the directory `bcr audit` is run
  in, which is the root of the repository (ADR-0017).
- A target that does not exist does not hold (ADR-0019). A target
  that is a directory or a symbolic link does not hold either. Each
  such claim is Refuted.
- A target that is a regular file and cannot be read is not judged:
  `bcr audit` stops with exit status 2.
- `bcr audit` copies its input to standard output, byte for byte, as
  `bcr verify` does (ADR-0016). It then prints a `claim-verdict`
  record for each claim and a `verdict` record for each breadcrumb.
  Each carries the `PATH` and `LINE` of the record it judges.
- Exit status:
  - 0: no claim is Refuted.
  - 1: at least one claim is Refuted.
  - 2: a line of its input is not a record, or a target could not be
    read.

## Consequences

- An agent in the middle of a change gets the verdict of what it has
  written. A reviewer gets a commit's verdict by running the pipe on
  that commit, as CI does.
- A verdict does not name the state it was computed on: a working
  tree has no name. The caller knows which commit it ran on.
- Under `set -o pipefail`, a Refuted claim fails the pipe, so CI
  turns red when the code moves away from a spec.
- An Undecided breadcrumb does not fail `bcr audit`. A problem still
  fails the pipe, through the stage that found it.
- `bcr audit` must be run from the root of the repository; run from
  elsewhere, its targets are not found, and their claims are Refuted.
- A claim cannot be made about a file through a symbolic link.
- `bcr audit` reads its whole input before it writes, like
  `bcr verify`, so `bcr report` starts only once it has ended.

## Alternatives Considered

- Reading targets from a commit, with `git`: a verdict that names its
  state, but `bcr` would need `git`, and work not yet committed could
  not be checked.
- A target that cannot be read is Refuted, or Undecided: the pipe
  would go on, but the verdict would speak of the machine it ran on,
  not of the repository.
- Following symbolic links: a claim could reach a file by another
  name, or one outside the repository, which a `TARGET` may not name.
- Failing on Undecided as well: stricter, but most breadcrumbs make
  no claim, so the pipe would always fail.
- Printing only the verdicts: a smaller output, but `bcr report`
  needs the breadcrumbs and links as well, and would have to read
  them from another place.
