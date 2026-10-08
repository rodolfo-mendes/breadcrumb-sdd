---
breadcrumb:
  id: filter
  type: spec
  links:
    - follows ADR-0006
    - follows ADR-0007
    - follows ADR-0008
    - follows ADR-0011
  claims:
    - 'docs/bcr.md has-line ### filter'
---
# Feature: filter

`bcr filter` reads the records `bcr extract` prints, and prints only
those that match the flags it is given, unchanged.

## Blueprint

### Context

An agent or a person often wants one breadcrumb, or one slice of the
set, not all of it: ADR-0019 and its links, or every PBI. Today that
means reading whole files and searching their text, though the
records already hold the ids, types and links. `bcr filter` selects
among the records, as a stage of the pipe:

```
bcr extract | bcr filter --id ADR-0019
```

It reads the records the stages of the pipe print and prints records
of the same format, so it can stand wherever records flow. Right after
`bcr extract`, it looks a breadcrumb up. After `bcr audit`, it shows
one breadcrumb's verdict, judged with the whole set; a check of one
breadcrumb needs no command of its own:

```
bcr extract | bcr verify | bcr audit | bcr filter --id ADR-0019
```

`bcr filter` judges nothing. It finds no problem, computes no verdict,
and checks no rule about the set: two records with the same id both
match, and no match is not an error. Those rules belong to
`bcr verify`, and they need every breadcrumb, so `bcr filter` comes
after the stages that judge, not before them.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command, `cmd/bcr` | Reads the flags and standard input, prints the records that match, sets the exit status | Infrastructure |
| Record reader, `internal/records` | Turns each line of input into a record, with its kind and fields; reports a line that is not a record | Infrastructure |

- Matching a flag against a field is no rule about breadcrumbs, so
  nothing of `bcr filter` is in the core, as with `bcr init`.
- `bcr filter` opens no file: it reads records only.

### Interface

```
bcr filter [-i ID] [-t TYPE] [-k KIND]
```

`bcr filter` takes no operand. It reads records from standard input
until it ends, as `bcr extract`, `bcr verify` and `bcr audit` print
them (`specs/extract/spec.md`, `specs/verify/spec.md`,
`specs/audit/spec.md`):

```
breadcrumb	ID	TYPE	PATH	LINE
link	ID	VERB	OBJECT	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
problem	PATH	LINE	MESSAGE
claim-verdict	ID	VERDICT	PATH	LINE
verdict	ID	VERDICT	PATH	LINE
```

Flags (ADR-0006):

- `-i ID`, `--id ID`: a `breadcrumb`, `link`, `claim`,
  `claim-verdict` or `verdict` record matches when its `ID` field
  equals `ID`. A `problem` record has no `ID` field, and never
  matches.
- `-t TYPE`, `--type TYPE`: a `breadcrumb` record matches when its
  `TYPE` field equals `TYPE`. No other kind of record has a `TYPE`
  field, so no other kind matches.
- `-k KIND`, `--kind KIND`: a record matches when its first field
  equals `KIND`. `KIND` is not checked against the kinds `bcr` knows:
  one that names no kind matches nothing (ADR-0011).

Rules:

- At least one flag is given.
- Each flag may be given more than once. A record matches a flag when
  it matches at least one of its values. A record is printed when it
  matches every flag given. A flag not given sets no condition.
- Fields are compared byte for byte, letter case included.
- A record of a kind `bcr` does not know matches only `--kind`, by its
  first field (ADR-0011). Fields after the ones above are not read,
  since new fields may be added at the end of a record.
- A line of input may end in `\n` or `\r\n`. The last line need not
  end.
- A `breadcrumb`, a `claim-verdict` and a `verdict` record have at
  least five fields, a `link` record at least six, a `claim` record at
  least seven and a `problem` record at least four. None of the fields
  above is empty, and `LINE` is a whole number from 1, written in
  digits only.
- Nothing else about a record is checked: not a `claim` record's
  `TARGET`, `KIND` and `ARGUMENT`, and not a `VERDICT`. Those rules
  belong to `bcr audit` and `bcr report`, and `bcr filter` judges
  nothing.
- A line whose first field is empty has no kind, and is not a record.
- A line that is not a record `bcr extract` could print, an empty line
  included, stops `bcr filter`: it writes a message starting `bcr: `
  that gives the line's number in the input, prints nothing to
  standard output, and exits 2.

`bcr filter` reads its whole input before it writes. It prints each
record that matches as it was read, byte for byte, with its line
ending, in the order read. It adds no record and changes none.

Exit status:

- 0: the input was read to its end, whether or not a record matched.
- 2: `bcr filter` was given an operand or no flag, its input could not
  be read, a line of its input was not a record, or its output could
  not be written.

`bcr filter` never exits 1: it finds nothing wrong, it only selects.

### Constraints

- A `problem` record is printed only when `--kind problem` is given,
  alone or with other kinds. The problems of a selected breadcrumb do
  not follow it, since a `problem` record has no id (ADR-0021).
- Before `bcr verify`, `bcr filter` leaves `bcr verify` a part of the
  set: a link to a breadcrumb filtered out is reported as pointing
  nowhere. Before `bcr audit`, it drops the `problem` records that
  would make a breadcrumb Undecided (ADR-0022), so a breadcrumb with a
  dropped claim can read Confirmed. A verdict is read after
  `bcr audit`, never computed after `bcr filter`.
- `--type` alone prints `breadcrumb` records only. The links and
  claims of the breadcrumbs it selects are found by their ids, in a
  second `bcr filter`.
- Two ids that differ only in case are two ids here. Telling them
  apart, and reporting them, is `bcr verify`'s rule (ADR-0003).
- With no index, `bcr extract` still reads every file `.breadcrumbs`
  names: `bcr filter` narrows what is printed, not what is read.

## Contract

### Definition of Done

- [x] Each Scenario below has a test.
- [x] `docs/bcr.md` describes `bcr filter` under `### filter`, and
      lists it in its SYNOPSIS; `docs/bcr.1` is generated again.
- [x] `AGENTS.md`'s Toolchain shows how to look up a breadcrumb with
      `bcr filter`.
- [x] In this repository, `bcr extract | bcr filter --type ADR | wc -l`
      prints the number of files in `docs/adrs/`.
- [x] In this repository, `bcr extract | bcr filter --kind breadcrumb`
      prints the same lines as
      `bcr extract | grep '^breadcrumb'`.
- [x] In this repository,
      `bcr extract | bcr verify | bcr audit | bcr filter --kind verdict --id filter`
      prints one line: this spec's verdict.

### Regression Guardrails

- `docs/bcr.md` documents `bcr filter`.

- `bcr filter` reads no file but its standard input.

### Scenarios

```gherkin
Scenario: One id
  Given the records
    """
    breadcrumb	ADR-0001	ADR	docs/adrs/a.md	3
    link	ADR-0001	amends	ADR-0002	docs/adrs/a.md	6
    breadcrumb	ADR-0002	ADR	docs/adrs/b.md	3
    """
  When I pipe them to "bcr filter --id ADR-0001"
  Then standard output is
    """
    breadcrumb	ADR-0001	ADR	docs/adrs/a.md	3
    link	ADR-0001	amends	ADR-0002	docs/adrs/a.md	6
    """
  And nothing is printed to standard error
  And the exit status is 0

Scenario: A breadcrumb's claims follow its id
  Given a breadcrumb record, a link record and a claim record whose ID
    is "verify", and the records of another breadcrumb
  When I pipe them to "bcr filter --id verify"
  Then the breadcrumb, link and claim records of "verify" are printed
  And the records of the other breadcrumb are not

Scenario: Several values of one flag
  Given breadcrumb records with the ids ADR-0001, ADR-0002 and
    ADR-0003
  When I pipe them to "bcr filter -i ADR-0001 -i ADR-0003"
  Then the records of ADR-0001 and ADR-0003 are printed
  And the record of ADR-0002 is not

Scenario: Two flags
  Given a breadcrumb record with the id ADR-0001 and the type ADR, and
    a breadcrumb record with the id PBI-00001 and the type PBI
  When I pipe them to "bcr filter --id ADR-0001 --type PBI"
  Then nothing is printed
  And the exit status is 0

Scenario: One type
  Given breadcrumb records of the types ADR and PBI, each with a link
    record
  When I pipe them to "bcr filter --type ADR"
  Then only the breadcrumb records of the type ADR are printed

Scenario: One kind
  Given a breadcrumb record, a link record and a claim record of one id
  When I pipe them to "bcr filter --kind link"
  Then only the link record is printed

Scenario: A kind and an id
  Given the link records of ADR-0001 and of ADR-0002
  When I pipe them to "bcr filter --kind link --id ADR-0001"
  Then only the link records of ADR-0001 are printed

Scenario: Problem records
  Given a breadcrumb record with the id ADR-0001, and the record
    "problem	docs/adrs/a.md	9	claim must be written in single quotes"
  When I pipe them to "bcr filter --id ADR-0001"
  Then only the breadcrumb record is printed
  When I pipe them to "bcr filter --kind problem"
  Then only the problem record is printed

Scenario: A kind bcr does not know
  Given a valid set of records, and a record "note	x"
  When I pipe them to "bcr filter --kind note"
  Then only the record "note	x" is printed
  When I pipe them to "bcr filter --kind nonsense"
  Then nothing is printed
  And the exit status is 0

Scenario: No match
  Given a valid set of records, none with the id ADR-9999
  When I pipe them to "bcr filter --id ADR-9999"
  Then nothing is printed
  And the exit status is 0

Scenario: Two breadcrumbs with the same id
  Given breadcrumb records with the id ADR-0003 at a.md and at b.md
  When I pipe them to "bcr filter --id ADR-0003"
  Then both records are printed
  And nothing is printed to standard error

Scenario: Letter case
  Given a breadcrumb record with the id ADR-0001
  When I pipe it to "bcr filter --id adr-0001"
  Then nothing is printed

Scenario: The order of the output
  Given breadcrumb records with the ids ADR-0002 and ADR-0001, in that
    order, both of the type ADR
  When I pipe them to "bcr filter --type ADR"
  Then the record of ADR-0002 is printed before that of ADR-0001

Scenario: No flag
  Given a valid set of records
  When I pipe them to "bcr filter"
  Then its usage line is printed to standard error
  And nothing is printed to standard output
  And the exit status is 2

Scenario: An operand
  When I run "bcr filter --id ADR-0001 records.tsv"
  Then its usage line is printed to standard error
  And the exit status is 2

Scenario: A line that is not a record
  Given an input whose line 3 is "breadcrumb	ADR-0001	ADR"
  When I pipe it to "bcr filter --id ADR-0001"
  Then a message starting "bcr: " that gives line 3 is printed to
    standard error
  And nothing is printed to standard output
  And the exit status is 2

Scenario: Fields added at the end
  Given a breadcrumb record with the id ADR-0001 and one more field at
    its end
  When I pipe it to "bcr filter --id ADR-0001"
  Then the record is printed unchanged, its last field included

Scenario: Windows line endings
  Given a breadcrumb record with the id ADR-0001 whose line ends in
    "\r\n"
  When I pipe it to "bcr filter --id ADR-0001"
  Then the record is printed with its line still ending in "\r\n"

Scenario: Output that cannot be written
  Given a valid set of records, and a standard output that cannot be
    written
  When I pipe them to "bcr filter --kind breadcrumb"
  Then a message starting "bcr: " is printed to standard error
  And the exit status is 2

Scenario: One breadcrumb's verdict
  Given the records bcr audit prints for the breadcrumbs verify and
    audit, with their claim-verdict and verdict records
  When I pipe them to "bcr filter --id verify"
  Then the breadcrumb, link and claim records of verify are printed,
    then its claim-verdict and verdict records
  And no record of audit is printed

Scenario: Before bcr verify
  Given breadcrumbs ADR-0001 and ADR-0002, and a link from ADR-0001 to
    ADR-0002
  When I run "bcr extract | bcr filter --id ADR-0001 | bcr verify"
  Then bcr verify reports the link to ADR-0002 as pointing nowhere
```
