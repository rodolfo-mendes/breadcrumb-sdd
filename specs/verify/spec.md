---
breadcrumb:
  id: verify
  type: spec
  links:
    - follows ADR-0003
    - follows ADR-0006
    - follows ADR-0007
    - follows ADR-0008
    - follows ADR-0011
    - follows ADR-0013
    - follows ADR-0015
---
# Feature: verify

`bcr verify` reads the records `bcr extract` prints, and checks the
rules about the whole set of breadcrumbs they describe.

## Blueprint

### Context

`bcr extract` checks what one file can show: the form of its
breadcrumb. Some rules need every breadcrumb at once. `bcr verify`
checks two of them:

- No two breadcrumbs have the same id, and no two have ids that differ
  only in letter case (ADR-0003).
- Every link points to the id of a breadcrumb in the set.

It reads records, not files (ADR-0015):

```
bcr extract | bcr verify
```

Which files make up the set is decided before it, by `bcr extract`'s
operands or by `.breadcrumbs` (ADR-0014). For now `bcr verify` checks
only the integrity of the set; whether it will later check claims and
compute verdicts is not decided.

The old `bcr check` stays as it is, for `breadcrumbs/`.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command | Reads standard input, prints problems, sets the exit status | Infrastructure |
| Record reader | Turns each line of input into a breadcrumb or a link of the core, with its path and line; reports a line that is not a record | Infrastructure |
| Core, `internal/crumb` | Checks the set: duplicate ids and links that point nowhere | Core domain |

- The core imports only the standard library, and knows no format,
  not even the format of records (ADR-0007, ADR-0013).
- `bcr verify` reads no file that carries a breadcrumb (ADR-0015).

### Interface

```
bcr verify
```

`bcr verify` takes no operand. It reads records from standard input
until it ends, as `bcr extract` prints them (`specs/extract/spec.md`):

```
breadcrumb	ID	TYPE	PATH	LINE
link	ID	VERB	OBJECT	PATH	LINE
```

- A record of a kind it does not know is ignored (ADR-0011).
- Fields after the ones above are ignored, since new fields may be
  added at the end of a record.
- A line of input may end in `\n` or `\r\n` (ADR-0015). The last
  line need not end.
- A `breadcrumb` record has at least five fields and a `link` record
  at least six. None of the fields above is empty, and `LINE` is a
  whole number from 1, written in digits only.
- A line whose first field is empty has no kind, and is not a record.
- A line that is not a record `bcr extract` could print, an empty line
  included, stops `bcr verify`: it writes a message starting `bcr: `
  that gives the line's number in the input, prints no problem, and
  exits 2.

Rules:

- Two ids are the same when their bytes are equal. Two ids differ only
  in case when they are equal under Unicode simple case folding
  (ADR-0003).
- When two or more breadcrumbs have the same id, or ids that differ
  only in case, each of them is a problem, at the `PATH` and `LINE` of
  its `breadcrumb` record. Its message names where the others are.
- A link whose `OBJECT` is not the id of any breadcrumb in the set is
  a problem, at the `PATH` and `LINE` of its `link` record. When
  `OBJECT` differs only in case from an id in the set, the message
  names that id.
- A link to an id that two breadcrumbs share is not a problem of its
  own; the duplicate is.

`bcr verify` prints nothing to standard output. A problem is printed
to standard error as `PATH:LINE: MESSAGE` (ADR-0006), in order of
`PATH` compared as bytes, then of `LINE`.

Exit status:

- 0: no problem was found.
- 1: at least one problem was found.
- 2: `bcr verify` was given an operand, its input could not be read,
  or a line of its input was not a record.

### Constraints

- A breadcrumb whose file `bcr extract` could not read prints no
  record, so `bcr verify` does not see it: links to its id are
  reported as pointing nowhere (ADR-0015).
- In `bcr extract | bcr verify`, the shell returns `bcr verify`'s exit
  status; a problem `bcr extract` found shows only with
  `set -o pipefail`, or in `bcr extract`'s own standard error
  (ADR-0015).

## Contract

### Definition of Done

- [x] Each Scenario below has a test.
- [x] `docs/bcr.md` describes `bcr verify` under `### verify`;
      `docs/bcr.1` is generated again.
- [x] The old `bcr check` and its tests are unchanged.
- [x] `go list -f '{{.Imports}}'` on the core's packages lists only
      the standard library and other core packages.
- [x] In this repository, `bcr extract | bcr verify` prints nothing
      and exits 0.

### Regression Guardrails

- `docs/bcr.md` documents `bcr verify`.

  Claims:
  - `docs/bcr.md` contains `### verify`

### Scenarios

```gherkin
Scenario: A set with no problem
  Given records whose ids are all different, and whose links all point
    to one of those ids
  When I pipe them to "bcr verify"
  Then nothing is printed
  And the exit status is 0

Scenario: No input
  Given an empty input
  When I pipe it to "bcr verify"
  Then nothing is printed
  And the exit status is 0

Scenario: Two breadcrumbs with the same id
  Given the records
    """
    breadcrumb	ADR-0003	ADR	docs/adrs/a.md	3
    breadcrumb	ADR-0003	ADR	docs/adrs/b.md	3
    """
  When I pipe them to "bcr verify"
  Then a problem is printed at "docs/adrs/a.md:3" that names
    "docs/adrs/b.md:3"
  And a problem is printed at "docs/adrs/b.md:3" that names
    "docs/adrs/a.md:3"
  And the exit status is 1

Scenario: Three breadcrumbs with the same id
  Given three breadcrumb records with the id PBI-00007
  When I pipe them to "bcr verify"
  Then each of the three is a problem, whose message names the other
    two

Scenario: Ids that differ only in case
  Given breadcrumb records with the ids "ADR-0001" and "adr-0001"
  When I pipe them to "bcr verify"
  Then each is a problem that names the other
  And the exit status is 1

Scenario: Case beyond ASCII
  Given breadcrumb records with the ids "ΣΙΓΜΑ" and "σιγμα"
  When I pipe them to "bcr verify"
  Then each is a problem that names the other

Scenario: A link that points nowhere
  Given a breadcrumb record with the id PBI-00001, and the record
    """
    link	PBI-00001	implements	ADR-0099	tasks/PBI-00001.md	6
    """
  When I pipe them to "bcr verify"
  Then a problem is printed at "tasks/PBI-00001.md:6"
  And the exit status is 1

Scenario: A link whose object differs only in case
  Given a breadcrumb record with the id ADR-0001, and a link record
    whose object is "adr-0001"
  When I pipe them to "bcr verify"
  Then a problem is printed at the link's path and line, whose message
    names "ADR-0001"
  And the exit status is 1

Scenario: A link to a duplicated id
  Given two breadcrumb records with the id ADR-0003, and a link record
    whose object is ADR-0003
  When I pipe them to "bcr verify"
  Then the two breadcrumb records are problems
  And the link is not

Scenario: Records of a kind it does not know
  Given a valid set of records, and a record "claim	..."
  When I pipe them to "bcr verify"
  Then nothing is printed
  And the exit status is 0

Scenario: Fields added at the end
  Given a valid set of records with one more field at the end of each
  When I pipe them to "bcr verify"
  Then nothing is printed
  And the exit status is 0

Scenario: Windows line endings
  Given a valid set of records whose lines end in "\r\n"
  When I pipe them to "bcr verify"
  Then nothing is printed
  And the exit status is 0

Scenario: A line that is not a record
  Given an input whose line 4 is "breadcrumb	ADR-0001	ADR"
  When I pipe it to "bcr verify"
  Then a message starting "bcr: " that gives line 4 is printed to
    standard error
  And the exit status is 2

Scenario: An operand
  When I run "bcr verify records.tsv"
  Then its usage line is printed to standard error
  And the exit status is 2

Scenario: The order of problems
  Given problems in tasks/b.md at line 3, in docs/a.md at line 9, and
    in docs/a.md at line 2
  When I pipe the records to "bcr verify"
  Then they are printed for docs/a.md:2, docs/a.md:9, then tasks/b.md:3

Scenario: A breadcrumb extract could not read
  Given a file whose breadcrumb has a problem, with the id ADR-0005,
    and another file with a link to ADR-0005
  When I run "bcr extract a.md b.md | bcr verify"
  Then bcr extract prints the problem of the first file
  And bcr verify reports the link to ADR-0005 as pointing nowhere
```
