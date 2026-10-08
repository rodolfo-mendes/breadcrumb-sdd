---
breadcrumb:
  id: verify
  type: spec
  links:
    - constrained_by ADR-0003
    - constrained_by ADR-0006
    - constrained_by ADR-0007
    - constrained_by ADR-0008
    - constrained_by ADR-0011
    - constrained_by ADR-0013
    - constrained_by ADR-0015
    - constrained_by ADR-0016
    - constrained_by ADR-0021
    - constrained_by ADR-0025
  claims:
    - 'docs/bcr.md has-line ### verify'
    - 'docs/bcr.md has-line Output: standard input, copied to standard output byte for byte.'
---
# Feature: verify

`bcr verify` reads the records `bcr extract` prints, checks the rules
about the whole set of breadcrumbs they describe, and passes every
record on to the next stage of the pipe, with a record for each
problem it found.

## Blueprint

### Context

`bcr extract` checks what one file can show: the form of its
breadcrumb. Some rules need every breadcrumb at once. `bcr verify`
checks two of them:

- No two breadcrumbs have the same id, and no two have ids that differ
  only in letter case (ADR-0003).
- Every link points to the id of a breadcrumb in the set.

A repository may also declare its shape in `breadcrumb.rules`: the
types its breadcrumbs may have, the links allowed between them, and
the types that may carry claims (ADR-0025). When the file is there,
`bcr verify` checks the set against it.

It reads records, not files (ADR-0015), and copies them to standard
output unchanged, so it can sit in a pipe before the stages that need
them (ADR-0016):

```
bcr extract | bcr verify | bcr audit
```

A problem it finds is printed to standard error, and after the copy
as a `problem` record, so the stages after it can see it (ADR-0021).

Which files make up the set is decided before it, by `bcr extract`'s
operands or by `.breadcrumbs` (ADR-0014). `bcr verify` checks only the
integrity of the set; it checks no claim and computes no verdict.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command | Reads standard input, copies it, prints problems, each also as a record, sets the exit status | Infrastructure |
| Record reader | Turns each line of input into a breadcrumb, a link or a claim of the core, with its path and line; reports a line that is not a record | Infrastructure |
| Rules reader, `internal/rules` | Turns the text of `breadcrumb.rules` into the rules of the core; reports a line the format does not allow | Infrastructure |
| Core, `internal/crumb` | Checks the set: duplicate ids and links that point nowhere. Checks the rules as a whole, and the set against them | Core domain |

- The core imports only the standard library, and knows no format,
  not even the format of records (ADR-0007, ADR-0013).
- `bcr verify` reads no file that carries a breadcrumb (ADR-0015).
  The only file it reads is `breadcrumb.rules` (ADR-0025).
- The format of `breadcrumb.rules` is known only to its reader; what
  the rules mean is in the core (ADR-0013).
- The command copies its input to its output itself; the record
  reader and the core never see the copy (ADR-0016).

### Interface

```
bcr verify
```

`bcr verify` takes no operand. It reads records from standard input
until it ends, as `bcr extract` prints them (`specs/extract/spec.md`):

```
breadcrumb	ID	TYPE	PATH	LINE
link	ID	VERB	OBJECT	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
```

- A `claim` record is read for its `ID`, `PATH` and `LINE`. Its
  `KIND`, `TARGET` and `ARGUMENT` are not checked: they are
  `bcr audit`'s (`specs/audit/spec.md`).
- A record of a kind it does not know is not checked (ADR-0011), and
  is copied like any other (ADR-0016). A `problem` record
  `bcr extract` printed is one of them.
- Fields after the ones above are not checked, since new fields may be
  added at the end of a record.
- A line of input may end in `\n` or `\r\n` (ADR-0015). The last
  line need not end.
- A `breadcrumb` record has at least five fields, a `link` record at
  least six, and a `claim` record at least seven. None of the fields
  above is empty, and `LINE` is a whole number from 1, written in
  digits only.
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

`bcr verify` reads `breadcrumb.rules` from the current directory,
which is the root of the repository (ADR-0025). With no such file, no
shape is checked. The file is a list of lines:

```
type	NAME
link	NAME	VERB	NAME
claims	NAME
```

- A line ends at `\n` or `\r\n`. The last line need not end. A byte
  order mark before the first line is ignored.
- A line of only spaces and tabs, and a line that starts with `#`,
  are ignored. No comment follows a field.
- Every other line is a `type`, `link` or `claims` line, its fields
  separated by one tab: 2 fields for `type` and `claims`, 4 for
  `link`, the kind included.
- A `NAME` is a type, and a `VERB` a verb: one or more characters,
  none of them white space. Both are compared as exact bytes.
- In a `link` line, the first `NAME` is the type of the subject, the
  breadcrumb a link is written in, and the second the type of the
  object, the breadcrumb it points to.
- Lines may come in any order.

A problem in the file is reported at `breadcrumb.rules` and the
number of its line, counted from 1:

| Problem | Message |
|---|---|
| A kind other than `type`, `link` and `claims` | `unknown kind of line "lnik"; a line is type, link or claims` |
| The wrong number of fields | `link has 3 fields, needs 4` |
| The same, on a line that has the right number when split at white space | `link has 1 field, needs 4; its fields are separated by spaces, use a tab` |
| An empty field | `link has an empty field 3` |
| A field that holds white space | `link has white space in field 3` |
| A `link` or `claims` line names a type no `type` line declares | `link names type "Spec", which no type line declares` |
| A `type` line repeats an earlier one | `type "spec" is already declared at line 3` |
| A `link` or `claims` line repeats an earlier one | `link "PBI changes spec" is already written at line 7` |

- A line has at most one of the first five problems, the first in the
  order above. The kind of a line is the first word of its first
  field.
- The last three are looked for only when no line has one of the
  first five. A line that repeats an earlier one has no other
  problem. A `link` line that names two undeclared types has a
  problem for each.
- When the file has a problem, the shape is not known: none of the
  three checks below runs. The rules about ids and links above still
  do.
- A file with no `type` line, an empty one included, has no problem:
  it declares a shape with no type.

With a `breadcrumb.rules` that has no problem:

- A breadcrumb whose `TYPE` no `type` line declares is a problem, at
  its `breadcrumb` record:
  `type "Spec" is not declared in breadcrumb.rules`.
- A link is a problem, at its `link` record, when no `link` line has
  the type of its subject, its `VERB` and the type of its object:
  `link "PBI implements ADR" matches no rule in breadcrumb.rules`. Its
  subject is the breadcrumb whose id is the record's `ID`, and its
  object the one whose id is `OBJECT`.
- A claim is a problem, at its `claim` record, when the type of the
  breadcrumb whose id is the record's `ID` has no `claims` line:
  `a breadcrumb of type "PBI" may not carry claims under breadcrumb.rules`.
- A link is not judged when its subject or its object is not the id
  of exactly one breadcrumb, or is a breadcrumb whose type is not
  declared. A claim is not judged when its `ID` is such an id. The
  problem is the missing breadcrumb, the duplicate or the type.

`bcr verify` copies its input to standard output, byte for byte: every
line in the order read, with its line ending as read, records of kinds
it does not know included (ADR-0016). It changes no record. It reads
its whole input before it writes.

- When it finds a problem, it still copies its whole input.
- When a line of its input is not a record, it prints nothing to
  standard output.

A problem is printed to standard error as `PATH:LINE: MESSAGE`
(ADR-0006), in order of `PATH` compared as bytes, then of `LINE`. A
problem in `breadcrumb.rules` has that name as its `PATH`, and takes
its place in the same order. Of two problems at one place, the one
about ids and links comes first.

After the copy, each problem is also printed to standard output, as a
record (ADR-0021), in the same order:

```
problem	PATH	LINE	MESSAGE
```

- `PATH`, `LINE` and `MESSAGE` are those of the line written to
  standard error, as in `bcr extract`'s `problem` record
  (`specs/extract/spec.md`).
- Each added record ends in `\n`. When the last line of the input has
  no line ending and a record is added, a `\n` is written before the
  first one, so that no two records share a line.
- When no problem is found, nothing is added: the output is equal to
  the input.

Exit status:

- 0: no problem was found.
- 1: at least one problem was found.
- 2: `bcr verify` was given an operand, its input could not be read,
  a line of its input was not a record, `breadcrumb.rules` exists and
  could not be read, or its output could not be written.

### Constraints

- `bcr verify` must be run from the root of the repository. Run from
  elsewhere, it finds no `breadcrumb.rules` and checks no shape
  (ADR-0025).
- Deleting `breadcrumb.rules` turns the shape checks off without a
  sign.
- The rules say which links are allowed, not which are required: a
  breadcrumb with no link is not a problem.

- A breadcrumb whose file `bcr extract` could not read prints no
  record, so `bcr verify` does not see it: links to its id are
  reported as pointing nowhere (ADR-0015).
- In `bcr extract | bcr verify`, the shell returns `bcr verify`'s exit
  status; a problem `bcr extract` found shows only with
  `set -o pipefail`, or in `bcr extract`'s own standard error
  (ADR-0015).
- A breadcrumb with a problem reaches the next stage as a valid
  record; what says it has a problem is the `problem` record with its
  `PATH` (ADR-0021).
- When `bcr verify` finds a problem, its output is no longer equal to
  its input: the input is the start of the output.
- A stage after `bcr verify` starts only once `bcr verify`'s input has
  ended (ADR-0016).

## Contract

### Definition of Done

- [x] Each Scenario below has a test.
- [x] `docs/bcr.md` describes `bcr verify` under `### verify`;
      `docs/bcr.1` is generated again.
- [x] `go list -f '{{.Imports}}'` on the core's packages lists only
      the standard library and other core packages.
- [x] In this repository, `bcr extract | bcr verify > /dev/null`
      prints nothing and exits 0.
- [x] In this repository, `bcr extract > a.tsv; bcr verify < a.tsv | cmp - a.tsv`
      exits 0.
- [x] `docs/bcr.md` describes the `problem` records `bcr verify` adds;
      `docs/bcr.1` is generated again.
- [x] Each Scenario about `breadcrumb.rules` has a test.
- [x] `docs/bcr.md` describes `breadcrumb.rules` and the checks
      against it; `docs/bcr.1` is generated again.
- [x] With no `breadcrumb.rules`, `bcr verify` prints for this
      repository's records what it printed before it read the file.

### Regression Guardrails

- `docs/bcr.md` documents `bcr verify`.

- `docs/bcr.md` says that `bcr verify` passes its input on.

### Scenarios

```gherkin
Scenario: A set with no problem
  Given records whose ids are all different, and whose links all point
    to one of those ids
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, unchanged
  And nothing is printed to standard error
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
  And the two records are printed to standard output, unchanged
  And after them a problem record for each of the two problems, in the
    order of standard error
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
  Given a valid set of records, and a record "note	..."
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, unchanged, the
    "note" record in its place among them
  And nothing is printed to standard error
  And the exit status is 0

Scenario: Fields added at the end
  Given a valid set of records with one more field at the end of each
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, unchanged
  And nothing is printed to standard error
  And the exit status is 0

Scenario: Windows line endings
  Given a valid set of records whose lines end in "\r\n"
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, their lines still
    ending in "\r\n"
  And nothing is printed to standard error
  And the exit status is 0

Scenario: A last line with no line ending
  Given a valid set of records whose last line does not end in "\n"
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, the last line still
    with no line ending
  And the exit status is 0

Scenario: A line that is not a record
  Given an input whose line 4 is "breadcrumb	ADR-0001	ADR"
  When I pipe it to "bcr verify"
  Then a message starting "bcr: " that gives line 4 is printed to
    standard error
  And nothing is printed to standard output
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
  And bcr verify prints the problem record of the first file and the
    records of the second, unchanged
  And after them a problem record for the link

Scenario: Output that cannot be written
  Given a valid set of records, and a standard output that cannot be
    written
  When I pipe them to "bcr verify"
  Then a message starting "bcr: " is printed to standard error
  And the exit status is 2

Scenario: A problem is also a record
  Given the records
    """
    breadcrumb	PBI-00001	PBI	tasks/PBI-00001.md	3
    link	PBI-00001	implements	ADR-0099	tasks/PBI-00001.md	6
    """
  When I pipe them to "bcr verify"
  Then standard error is
    """
    tasks/PBI-00001.md:6: link points to "ADR-0099", which is no breadcrumb's id
    """
  And standard output is the two records, then
    """
    problem	tasks/PBI-00001.md	6	link points to "ADR-0099", which is no breadcrumb's id
    """
  And the exit status is 1

Scenario: A problem after a last line with no line ending
  Given records with a link that points nowhere, whose last line does
    not end in "\n"
  When I pipe them to "bcr verify"
  Then standard output is the input, then "\n", then the problem
    record

Scenario: A problem record in the input
  Given a valid set of records, and a record "problem	a.md	3	id has no value"
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, unchanged
  And nothing is printed to standard error
  And the exit status is 0

Scenario: A claim record that is not a record
  Given an input whose line 2 is "claim	verify	has-line	docs/bcr.md"
  When I pipe it to "bcr verify"
  Then a message starting "bcr: " that gives line 2 is printed to
    standard error
  And nothing is printed to standard output
  And the exit status is 2
```

The scenarios below are about `breadcrumb.rules`. In them, "the
layout" is a `breadcrumb.rules` in the current directory with these
lines, a tab between fields:

```
type	ADR
type	spec
type	PBI
link	spec	constrained_by	ADR
link	PBI	changes	spec
claims	spec
```

and "the set" is these records:

```
breadcrumb	ADR-0007	ADR	docs/adrs/a.md	3
breadcrumb	verify	spec	specs/verify/spec.md	3
link	verify	constrained_by	ADR-0007	specs/verify/spec.md	6
claim	verify	has-line	docs/bcr.md	### verify	specs/verify/spec.md	8
breadcrumb	PBI-00030	PBI	tasks/PBI-00030.md	3
link	PBI-00030	changes	verify	tasks/PBI-00030.md	6
```

```gherkin
Scenario: A set that has the declared shape
  Given the layout
  When I pipe the set to "bcr verify"
  Then the records are printed to standard output, unchanged
  And nothing is printed to standard error
  And the exit status is 0

Scenario: No breadcrumb.rules
  Given no "breadcrumb.rules" in the current directory
  And the set, with the record
    """
    link	PBI-00030	implements	ADR-0007	tasks/PBI-00030.md	7
    """
  When I pipe them to "bcr verify"
  Then the records are printed to standard output, unchanged
  And nothing is printed to standard error
  And the exit status is 0

Scenario: A breadcrumb.rules that cannot be read
  Given a directory named "breadcrumb.rules" in the current directory
  When I pipe the set to "bcr verify"
  Then a message starting "bcr: " is printed to standard error
  And nothing is printed to standard output
  And the exit status is 2

Scenario: A type that is not declared
  Given the layout, and the set with the record
    """
    breadcrumb	ADR-0008	adr	docs/adrs/b.md	3
    """
  When I pipe them to "bcr verify"
  Then standard error is
    """
    docs/adrs/b.md:3: type "adr" is not declared in breadcrumb.rules
    """
  And standard output is the records, then a problem record for it
  And the exit status is 1

Scenario: A link no rule allows
  Given the layout, and the set with the record
    """
    link	PBI-00030	implements	ADR-0007	tasks/PBI-00030.md	7
    """
  When I pipe them to "bcr verify"
  Then standard error is
    """
    tasks/PBI-00030.md:7: link "PBI implements ADR" matches no rule in breadcrumb.rules
    """
  And the exit status is 1

Scenario: A claim on a type that may not carry claims
  Given the layout, and the set with the record
    """
    claim	PBI-00030	has-line	docs/bcr.md	### verify	tasks/PBI-00030.md	9
    """
  When I pipe them to "bcr verify"
  Then standard error is
    """
    tasks/PBI-00030.md:9: a breadcrumb of type "PBI" may not carry claims under breadcrumb.rules
    """
  And the exit status is 1

Scenario: No claims line
  Given the layout without its "claims" line
  When I pipe the set to "bcr verify"
  Then one problem is printed, at "specs/verify/spec.md:8"
  And the exit status is 1

Scenario: A link of a breadcrumb whose type is not declared
  Given the layout, and the records
    """
    breadcrumb	ADR-0007	ADR	docs/adrs/a.md	3
    breadcrumb	verify	Spec	specs/verify/spec.md	3
    link	verify	follows	ADR-0007	specs/verify/spec.md	6
    claim	verify	has-line	docs/bcr.md	### verify	specs/verify/spec.md	8
    """
  When I pipe them to "bcr verify"
  Then one problem is printed, at "specs/verify/spec.md:3"

Scenario: A link to a breadcrumb whose type is not declared
  Given the layout, and the records
    """
    breadcrumb	ADR-0007	adr	docs/adrs/a.md	3
    breadcrumb	verify	spec	specs/verify/spec.md	3
    link	verify	constrained_by	ADR-0007	specs/verify/spec.md	6
    """
  When I pipe them to "bcr verify"
  Then one problem is printed, at "docs/adrs/a.md:3"

Scenario: A link that points nowhere, with a shape
  Given the layout, and the set with the record
    """
    link	PBI-00030	implements	ADR-0099	tasks/PBI-00030.md	7
    """
  When I pipe them to "bcr verify"
  Then one problem is printed, at "tasks/PBI-00030.md:7", that says
    the link points to no breadcrumb's id

Scenario: A link to a duplicated id, with a shape
  Given the layout, and the set with the records
    """
    breadcrumb	ADR-0007	ADR	docs/adrs/b.md	3
    link	PBI-00030	implements	ADR-0007	tasks/PBI-00030.md	7
    """
  When I pipe them to "bcr verify"
  Then the two breadcrumb records with the id ADR-0007 are problems
  And no link is

Scenario: An unknown kind of line
  Given the layout, with the line "lnik	PBI	changes	spec" added as
    line 7
  When I pipe the set to "bcr verify"
  Then standard error is
    """
    breadcrumb.rules:7: unknown kind of line "lnik"; a line is type, link or claims
    """
  And standard output is the set, then
    """
    problem	breadcrumb.rules	7	unknown kind of line "lnik"; a line is type, link or claims
    """
  And the exit status is 1

Scenario: Fields separated by spaces
  Given the layout, with the line "link PBI changes ADR" added as
    line 7, a space between its fields
  When I pipe the set to "bcr verify"
  Then standard error is
    """
    breadcrumb.rules:7: link has 1 field, needs 4; its fields are separated by spaces, use a tab
    """
  And the exit status is 1

Scenario: The wrong number of fields
  Given a "breadcrumb.rules" whose lines are "type	spec	# a feature"
    and "claims"
  When I pipe the set to "bcr verify"
  Then a problem is printed at "breadcrumb.rules:1" that says
    "type has 3 fields, needs 2"
  And a problem is printed at "breadcrumb.rules:2" that says
    "claims has 1 field, needs 2"

Scenario: An empty field, and white space in a field
  Given a "breadcrumb.rules" whose lines are "type	spec",
    "link	spec		spec" and "link	spec	built on	spec"
  When I pipe the set to "bcr verify"
  Then a problem is printed at "breadcrumb.rules:2" that says
    "link has an empty field 3"
  And a problem is printed at "breadcrumb.rules:3" that says
    "link has white space in field 3"

Scenario: A type no type line declares
  Given the layout, with the line "link	PBI	changes	Spec" added as
    line 7
  When I pipe the set to "bcr verify"
  Then standard error is
    """
    breadcrumb.rules:7: link names type "Spec", which no type line declares
    """
  And the exit status is 1

Scenario: A line written twice
  Given the layout, with the lines "type	spec" and
    "link	PBI	changes	spec" added as lines 7 and 8
  When I pipe the set to "bcr verify"
  Then standard error is
    """
    breadcrumb.rules:7: type "spec" is already declared at line 2
    breadcrumb.rules:8: link "PBI changes spec" is already written at line 5
    """
  And the exit status is 1

Scenario: A breadcrumb.rules with a problem
  Given the layout, with the line "lnik	PBI	changes	spec" added as
    line 7
  And the set, with the records
    """
    breadcrumb	ADR-0008	adr	docs/adrs/b.md	3
    link	PBI-00030	implements	ADR-0099	tasks/PBI-00030.md	7
    """
  When I pipe them to "bcr verify"
  Then problems are printed at "breadcrumb.rules:7" and at
    "tasks/PBI-00030.md:7", in that order
  And none is printed at "docs/adrs/b.md:3"
  And the exit status is 1

Scenario: An empty breadcrumb.rules
  Given an empty "breadcrumb.rules"
  When I pipe the set to "bcr verify"
  Then a problem is printed at each of the three breadcrumb records,
    and at no other record
  And the exit status is 1

Scenario: Blank lines, comments and line endings in breadcrumb.rules
  Given the layout, starting with a byte order mark, its lines ending
    in "\r\n", with a line "# the layout", an empty line and a line
    of spaces among them, and no line ending after the last
  When I pipe the set to "bcr verify"
  Then nothing is printed to standard error
  And the exit status is 0
```
