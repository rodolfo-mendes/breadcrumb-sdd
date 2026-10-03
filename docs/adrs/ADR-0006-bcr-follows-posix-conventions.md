---
breadcrumb:
  id: ADR-0006
  type: ADR
  links: []
---
# ADR-0006: bcr follows POSIX command-line conventions

Status: Accepted
Date: 2026-10-02

## Context

People and agents run `bcr` in scripts and pipes, next to the tools
a machine already has, such as `cut`, `awk`, `grep` and `sort`. For
that, `bcr` should behave like any other POSIX utility: text on
standard output, messages on standard error, and an exit status that
says what happened.

POSIX's Utility Syntax Guidelines define short flags only. Long flags,
such as `--output`, come from GNU's `getopt_long`.

## Decision

- A command is run as `bcr COMMAND [FLAGS] [OPERANDS]`. Flags come
  before operands: the first operand ends the flags, and so does `--`.
- Flags follow getopt style:
  - a short flag is `-` and one letter, and short flags that take no
    value can be combined, so `-ab` is `-a -b`;
  - a short flag's value follows it, attached or as the next
    argument: `-oFILE` or `-o FILE`;
  - a long flag is `--` and a name, and its value follows `=` or comes
    as the next argument: `--output=FILE` or `--output FILE`.
- A result on standard output is one record per line, with its fields
  separated by a tab. There is no header line and no color.
- A message on standard error starts with `bcr: `. A usage error is
  followed by the command's usage line.
- A problem found in a file is written `PATH:LINE: MESSAGE`, with the
  path relative to the root of the repository.
- The exit status is 0 when the command found nothing wrong, 1 when
  it found something wrong, and 2 when it was used wrongly or could
  not do its work.

## Consequences

- Results can be read by `cut`, `awk`, `grep` and `sort` without a
  parser, and editors can jump to a `PATH:LINE` problem, as they do
  for compilers and `grep -n`.
- A field cannot hold a tab or a line break, so values `bcr` writes as
  fields must not contain them.
- An operand that starts with `-` has to come after `--`.
- Flags are read by a getopt-style parser of this repository's own, in
  `internal/cli`: Go's `flag` package cannot combine short flags, and
  reads `-name` and `--name` alike. The parser exists and is tested,
  so it stays; replacing it with a library would need an ADR (ADR-0005).
- Each command's flags, operands and exit status are listed in
  `docs/bcr.md`.

## Alternatives Considered

- JSON output: exact, but it needs `jq` to go down a pipe, and POSIX
  does not have `jq`.
- Aligned columns: easy to read by eye, but a value with spaces cannot
  be cut out of them.
- Flags anywhere among the operands, as GNU utilities allow:
  convenient, but not POSIX, and an operand could be read as a flag.
- Short flags only, as POSIX defines: strictly POSIX, but a long name
  says what a flag does where a letter cannot.
- A library such as `pflag`: getopt style too, but it would replace a
  parser that already works.