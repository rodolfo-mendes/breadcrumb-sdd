# bcr(1)

## NAME

bcr - audit the breadcrumbs of a repository

## SYNOPSIS

```
bcr audit-report --html [-o FILE]
bcr check
```

## DESCRIPTION

`bcr` is the toolkit of Breadcrumb SDD. It reads the breadcrumbs of a
repository, checks the claims of its Tasks against the repository's
files, and reports what it finds. Run it from the root of the
repository.

This document is the contract of the interface of `bcr`: its commands,
flags, inputs, outputs and exit codes. The rules of the method are in
the breadcrumbs of the Breadcrumb SDD repository, and this document
points to the one that holds each rule
([TD-0018](../breadcrumbs/TD-0018.md)).

Every command follows the same conventions:

- It is run as `bcr COMMAND [FLAGS] [OPERANDS]`. Flags come before
  operands; the first operand, or `--`, ends them
  ([TD-0020](../breadcrumbs/TD-0020.md)).
- Flags are in getopt style: a short flag is one letter, such as `-a`,
  and short flags can be combined, as in `-ab`; a long flag is a name,
  such as `--html`. A flag's value follows it: `-x VALUE`, `-xVALUE`,
  `--name=VALUE` or `--name VALUE`
  ([RQ-0011](../breadcrumbs/RQ-0011.md)).
- Results go to standard output, and nothing else does. Messages go to
  standard error and start with `bcr: `. A usage error is followed by
  the usage line ([RQ-0009](../breadcrumbs/RQ-0009.md),
  [TD-0020](../breadcrumbs/TD-0020.md)).
- The exit status says whether the command found something wrong (see
  EXIT STATUS).

Which files are breadcrumbs:
[TD-0013](../breadcrumbs/TD-0013.md). The kinds of breadcrumb:
[TD-0001](../breadcrumbs/TD-0001.md),
[TD-0003](../breadcrumbs/TD-0003.md),
[TD-0004](../breadcrumbs/TD-0004.md) and
[TD-0011](../breadcrumbs/TD-0011.md). How a breadcrumb names its
parents: [TD-0005](../breadcrumbs/TD-0005.md).

## COMMANDS

### audit-report

```
bcr audit-report --html [-o FILE]
```

Audits the breadcrumbs and writes the result as an HTML report
([RQ-0001](../breadcrumbs/RQ-0001.md)).

Flags:

- `--html`: write the report as HTML. It is required, and HTML is the
  only format.
- `-o FILE`, `--output FILE`: write the report to `FILE` instead of
  `audit-report.html`, replacing any file there. A relative `FILE` is
  taken from the current directory. `-o -` writes the report to
  standard output ([RQ-0017](../breadcrumbs/RQ-0017.md),
  [RQ-0018](../breadcrumbs/RQ-0018.md)).

Operands: none.

Standard input: not read.

Output: without `-o`, the report is written to `audit-report.html` in
the current directory, replacing any file there
([RQ-0019](../breadcrumbs/RQ-0019.md)). When the report is written to
a file, standard output gets one line, the absolute path of the
report. With `-o -`, standard output gets the report, and no file is
written.

The report draws the breadcrumbs as a graph
([RQ-0002](../breadcrumbs/RQ-0002.md)), each with its properties
([RQ-0003](../breadcrumbs/RQ-0003.md)) and an arrow to each of its
parents ([RQ-0004](../breadcrumbs/RQ-0004.md)), colored by its verdict
([RQ-0005](../breadcrumbs/RQ-0005.md)). How a verdict is reached:
[TD-0012](../breadcrumbs/TD-0012.md) for a claim,
[RQ-0007](../breadcrumbs/RQ-0007.md) for a Task, and
[RQ-0006](../breadcrumbs/RQ-0006.md) and
[RQ-0008](../breadcrumbs/RQ-0008.md) for the breadcrumbs above it.
The report lists the same problems as `bcr check`, each with its file
and line.

Exit status:

- 0: the report is written, and no breadcrumb is Refuted or has a
  problem.
- 1: the report is written, and a breadcrumb is Refuted or has a
  problem.
- 2: `bcr` was used wrongly, or there is no `breadcrumbs/` directory
  in the current directory, or the report could not be written, such
  as to a directory that does not exist. No report is written.

### check

```
bcr check
```

Checks the syntax of the breadcrumbs and prints each problem it finds
([RQ-0020](../breadcrumbs/RQ-0020.md)).

Flags: none.

Operands: none. Every breadcrumb is checked.

Standard input: not read.

Output: one line on standard output for each problem, in order of path
and then of line:

```
PATH:LINE: MESSAGE
```

`PATH` is relative to the repository root, and `LINE` is counted from
1 ([TD-0020](../breadcrumbs/TD-0020.md)). What is a problem: a first
line that is not a title ([TD-0013](../breadcrumbs/TD-0013.md)), a
`Parent:` line that is not a link or links to no breadcrumb
([TD-0005](../breadcrumbs/TD-0005.md),
[RQ-0022](../breadcrumbs/RQ-0022.md)), and a list item under a Task's
`## Claims` heading that is not a claim
([TD-0011](../breadcrumbs/TD-0011.md),
[TD-0012](../breadcrumbs/TD-0012.md)). A claim that does not hold is
not a problem; `bcr audit-report` shows it in the Task's verdict.

Standard error gets a warning for each file in `breadcrumbs/` that is
not a breadcrumb ([RQ-0023](../breadcrumbs/RQ-0023.md)):

```
bcr: warning: PATH is not a breadcrumb (TD-0013)
```

Exit status ([RQ-0021](../breadcrumbs/RQ-0021.md)):

- 0: no problem was found. Warnings do not change it.
- 1: a problem was found.
- 2: `bcr` was used wrongly, or there is no `breadcrumbs/` directory
  in the current directory.

## EXIT STATUS

Every command exits with ([RQ-0010](../breadcrumbs/RQ-0010.md)):

- 0: it succeeded and found nothing wrong.
- 1: it ran and found something wrong in the breadcrumbs.
- 2: it was used wrongly, or could not run.

## SEE ALSO

The method: `README.md`. How the repository is developed:
`CONTRIBUTING.md`. The rules: the breadcrumbs in `breadcrumbs/`.
