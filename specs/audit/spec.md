---
breadcrumb:
  id: audit
  type: spec
  links:
    - follows ADR-0006
    - follows ADR-0007
    - follows ADR-0008
    - follows ADR-0011
    - follows ADR-0013
    - follows ADR-0016
    - follows ADR-0017
    - follows ADR-0019
    - follows ADR-0020
    - follows ADR-0021
    - follows ADR-0022
    - follows ADR-0023
  claims:
    - 'docs/bcr.md has-line ### audit'
---
# Feature: audit

`bcr audit` reads the records `bcr extract` and `bcr verify` print,
checks each claim against the file it is about, and adds a verdict for
each claim and each breadcrumb.

## Blueprint

### Context

`bcr extract` reads claims and checks their form; `bcr verify` checks
the set of breadcrumbs. Neither opens the file a claim is about.
`bcr audit` does: it is the stage where a breadcrumb turns Refuted
because the code moved away from what the breadcrumb says.

It sits third in the pipe, and passes on what it reads, so the stage
after it gets the breadcrumbs, the links, the problems and the
verdicts in one stream (ADR-0016, ADR-0023):

```
bcr extract | bcr verify | bcr audit
```

A verdict is `Confirmed`, `Refuted` or `Undecided` (ADR-0022). A
breadcrumb's verdict comes from its own claims and problems; no link
carries a verdict to another breadcrumb.

Claims are checked against the working tree, uncommitted changes
included (ADR-0023). `bcr audit` runs no command and follows no link
out of the repository: it only reads the files the claims name.

The old `bcr verdict` and `bcr audit-report` stay as they are, for
`breadcrumbs/`.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command | Reads standard input, copies it, prints the verdicts and the Refuted claims, sets the exit status | Infrastructure |
| Record reader, `internal/records` | Turns each line of input into a breadcrumb, a claim or a problem, with its path and line; reports a line that is not a record | Infrastructure |
| Target reader, `internal/target` | Reads a target from the current directory; tells a regular file from what is not one, and from a file that cannot be read | Infrastructure |
| Core, `internal/crumb` | Checks a claim's parts; says whether a claim holds in the content of its target (ADR-0019); gives a breadcrumb its verdict (ADR-0022) | Core domain |

- The core imports only the standard library, and knows no format,
  not even the format of records (ADR-0007, ADR-0013). It is given a
  target's content, and opens no file.
- The rule of a kind of claim is in one place in the core: the check
  of its argument, which `bcr extract` uses, and whether it holds,
  which `bcr audit` uses (ADR-0019).
- The command copies its input to its output itself, as `bcr verify`
  does (ADR-0016).
- Each target is read once, however many claims name it, so every
  claim about a file is judged on the same content.

### Interface

```
bcr audit
```

`bcr audit` takes no operand. It is run from the root of the
repository (ADR-0023). It reads records from standard input until it
ends, as `bcr extract` and `bcr verify` print them
(`specs/extract/spec.md`, `specs/verify/spec.md`), and reads three
kinds:

```
breadcrumb	ID	TYPE	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
problem	PATH	LINE	MESSAGE
```

- A record of another kind, a `link` record included, is not checked
  (ADR-0011), and is copied like any other.
- Fields after the ones above are not checked, since new fields may be
  added at the end of a record.
- A line of input may end in `\n` or `\r\n`. The last line need not
  end.
- A `breadcrumb` record has at least five fields, a `claim` record at
  least seven and a `problem` record at least four. None of the
  fields above is empty, and `LINE` is a whole number from 1, written
  in digits only.
- A `claim` record's `TARGET`, `KIND` and `ARGUMENT` follow the rules
  `bcr extract` checks (`specs/extract/spec.md`): a `TARGET` that
  starts with `/`, has a part that is `..` or holds white space, a
  `KIND` `bcr` does not know, or an `ARGUMENT` its kind does not
  allow, is not a record `bcr extract` could print.
- A line whose first field is empty has no kind, and is not a record.
- A line that is not a record `bcr extract` could print, an empty line
  included, stops `bcr audit`: it writes a message starting `bcr: `
  that gives the line's number in the input, prints nothing to
  standard output, and exits 2.

The verdict of a claim (ADR-0019, ADR-0023):

- `TARGET` is read from the current directory, with `/` between its
  parts.
- A `has-line` claim is Confirmed when its target is a regular file
  and at least one of its lines, without its line ending (`\n` or
  `\r\n`) and without the spaces and tabs at its start and end, is
  equal to `ARGUMENT`, byte for byte. Any line counts, the front
  matter's included.
- Otherwise it is Refuted. A target that does not exist, or is
  anything other than a regular file, such as a directory or a
  symbolic link, makes its claims Refuted.
- A target that is a regular file and cannot be read stops
  `bcr audit`: it writes a message starting `bcr: `, prints nothing to
  standard output, and exits 2.

The verdict of a breadcrumb (ADR-0022):

- A claim belongs to the `breadcrumb` record that has its `PATH`, and
  so does a problem (ADR-0021). Each problem counts as one Undecided
  claim.
- A breadcrumb is Refuted when at least one of its claims is Refuted;
  otherwise Undecided when it has at least one problem; otherwise
  Confirmed when it has at least one claim.
- A breadcrumb with no claim and no problem is Undecided, with the
  reason `no claims`.
- A claim or a problem whose `PATH` no `breadcrumb` record has belongs
  to no breadcrumb. The claim still gets its verdict.

`bcr audit` copies its input to standard output, byte for byte, as
`bcr verify` does (ADR-0016), and reads its whole input and every
target before it writes. After the copy it prints, with fields
separated by a tab (ADR-0006):

```
claim-verdict	ID	VERDICT	PATH	LINE
verdict	ID	VERDICT	PATH	LINE
verdict	ID	Undecided	PATH	LINE	no claims
```

- One `claim-verdict` record for each `claim` record, in the order
  read. `ID`, `PATH` and `LINE` are those of the `claim` record.
- Then one `verdict` record for each `breadcrumb` record, in the order
  read. `ID`, `PATH` and `LINE` are those of the `breadcrumb` record.
- `VERDICT` is `Confirmed`, `Refuted` or `Undecided`.
- A `verdict` record has a sixth field, `no claims`, only when its
  breadcrumb has no claim and no problem.
- A verdict belongs to a record, not to an id: two breadcrumbs with
  the same id each get their own.
- Each added record ends in `\n`. When the last line of the input has
  no line ending and a record is added, a `\n` is written before the
  first one.

A Refuted claim is printed to standard error as `PATH:LINE: MESSAGE`
(ADR-0006), at the `PATH` and `LINE` of its `claim` record, in the
order read. The message names the target, and says what it lacks. It
is not a problem: no `problem` record is printed for it.

Exit status (ADR-0023):

- 0: no claim is Refuted.
- 1: at least one claim is Refuted.
- 2: `bcr audit` was given an operand, its input could not be read, a
  line of its input was not a record, a target could not be read, or
  its output could not be written.

### Constraints

- An Undecided breadcrumb does not change the exit status. A problem
  fails the pipe through the stage that found it, under
  `set -o pipefail`.
- Run from a directory other than the root of the repository,
  `bcr audit` finds no target, and every claim is Refuted (ADR-0023).
- A verdict does not name the state it was computed on; the caller
  knows which commit it ran on (ADR-0023).
- A breadcrumb whose file `bcr extract` could not read has no record,
  and gets no verdict. Its problems are passed on.
- A symbolic link is not followed when it is the target itself. A
  directory on the way to the target may be one.
- A stage after `bcr audit` starts only once `bcr audit`'s input has
  ended.

## Contract

### Definition of Done

- [x] Each Scenario below has a test.
- [x] `docs/bcr.md` describes `bcr audit` under `### audit`;
      `docs/bcr.1` is generated again.
- [x] The old `bcr check`, `bcr verdict` and `bcr audit-report`, and
      their tests, are unchanged.
- [x] `go list -f '{{.Imports}}'` on the core's packages lists only
      the standard library and other core packages.
- [x] In this repository,
      `set -o pipefail; bcr extract | bcr verify | bcr audit > /dev/null`
      prints nothing and exits 0.
- [x] In this repository, the pipe prints a `claim-verdict` record
      with `Confirmed` for each claim written in `specs/`.
- [x] In this repository, with one claimed line changed by hand, the
      pipe exits 1 and prints its claim and its breadcrumb as
      `Refuted`.
- [x] CI and `AGENTS.md`'s Toolchain run the pipe with `bcr audit`.

### Regression Guardrails

- `docs/bcr.md` documents `bcr audit`.

- `bcr audit` runs no command: it only reads files.

### Scenarios

```gherkin
Scenario: A claim that holds
  Given docs/bcr.md with the line "### verify", and the records
    """
    breadcrumb	verify	spec	specs/verify/spec.md	3
    claim	verify	has-line	docs/bcr.md	### verify	specs/verify/spec.md	8
    """
  When I pipe them to "bcr audit"
  Then standard output is the two records, then
    """
    claim-verdict	verify	Confirmed	specs/verify/spec.md	8
    verdict	verify	Confirmed	specs/verify/spec.md	3
    """
  And nothing is printed to standard error
  And the exit status is 0

Scenario: A claim that does not hold
  Given docs/bcr.md with no line "### verify", and the records above
  When I pipe them to "bcr audit"
  Then standard output is the two records, then
    """
    claim-verdict	verify	Refuted	specs/verify/spec.md	8
    verdict	verify	Refuted	specs/verify/spec.md	3
    """
  And a message that names "docs/bcr.md" is printed to standard error
    at "specs/verify/spec.md:8"
  And the exit status is 1

Scenario: No input
  Given an empty input
  When I pipe it to "bcr audit"
  Then nothing is printed
  And the exit status is 0

Scenario: A line with spaces around it
  Given a target with the line "	  return nil  " and Windows line
    endings
  And a has-line claim whose text is "return nil"
  When I audit it
  Then the claim is Confirmed

Scenario: A line that only holds the text
  Given a target with the line "Status: Accepted, amended by ADR-0016"
  And a has-line claim whose text is "Status: Accepted"
  When I audit it
  Then the claim is Refuted

Scenario: Letter case
  Given a target with the line "### Verify"
  And a has-line claim whose text is "### verify"
  When I audit it
  Then the claim is Refuted

Scenario: A last line with no line ending
  Given a target whose last line is "### verify", with no line ending
  And a has-line claim whose text is "### verify"
  When I audit it
  Then the claim is Confirmed

Scenario: A target that does not exist
  Given a has-line claim whose target is no file in the repository
  When I audit it
  Then the claim is Refuted
  And the exit status is 1

Scenario: A target that is not a regular file
  Given has-line claims whose targets are a directory and a symbolic
    link to a file that has the claimed line
  When I audit them
  Then each claim is Refuted

Scenario: A target that cannot be read
  Given a has-line claim whose target is a regular file that cannot be
    read
  When I audit it
  Then a message starting "bcr: " is printed to standard error
  And nothing is printed to standard output
  And the exit status is 2

Scenario: Several claims of one breadcrumb
  Given a breadcrumb with three claims, the second of which does not
    hold
  When I audit it
  Then its claims are Confirmed, Refuted and Confirmed, in the order
    read
  And the breadcrumb is Refuted

Scenario: A breadcrumb with no claims
  Given the record
    """
    breadcrumb	ADR-0001	ADR	docs/adrs/a.md	3
    """
  When I pipe it to "bcr audit"
  Then standard output is the record, then
    """
    verdict	ADR-0001	Undecided	docs/adrs/a.md	3	no claims
    """
  And the exit status is 0

Scenario: A breadcrumb with a dropped claim
  Given a breadcrumb at a.md with one claim that holds, and a record
    "problem	a.md	9	claim must be written in single quotes"
  When I audit them
  Then the claim is Confirmed
  And the breadcrumb is Undecided, with no sixth field
  And the exit status is 0

Scenario: A Refuted claim beside a problem
  Given a breadcrumb at a.md with one claim that does not hold, and a
    problem record whose PATH is "a.md"
  When I audit them
  Then the breadcrumb is Refuted
  And the exit status is 1

Scenario: A breadcrumb whose only trouble is a problem
  Given a breadcrumb at a.md with no claim, and a problem record whose
    PATH is "a.md"
  When I audit them
  Then the breadcrumb is Undecided, with no sixth field

Scenario: A problem with no breadcrumb
  Given a breadcrumb at b.md with one claim that holds, and a problem
    record whose PATH is "a.md"
  When I audit them
  Then the problem record is printed to standard output, unchanged
  And the breadcrumb at b.md is Confirmed
  And no verdict is printed for a.md

Scenario: Two breadcrumbs with the same id
  Given breadcrumb records with the id ADR-0003 at a.md and b.md, and
    a claim that holds whose PATH is "a.md"
  When I audit them
  Then the breadcrumb at a.md is Confirmed
  And the breadcrumb at b.md is Undecided, with "no claims"

Scenario: Two claims about one target
  Given two breadcrumbs, each with a claim about docs/bcr.md, one that
    holds and one that does not
  When I audit them
  Then the first breadcrumb is Confirmed and the second Refuted

Scenario: The order of the output
  Given breadcrumbs at b.md and a.md, in that order, each with claims
  When I audit them
  Then the input is printed first, unchanged
  And then every claim-verdict record, in the order of the claim
    records
  And then the verdict record of b.md before that of a.md

Scenario: Records of a kind it does not read
  Given a valid set of records with link records, and a record
    "note	..."
  When I pipe them to "bcr audit"
  Then they are printed to standard output, unchanged, in their place

Scenario: Fields added at the end
  Given breadcrumb and claim records with one more field at the end of
    each
  When I pipe them to "bcr audit"
  Then they are printed unchanged
  And their verdicts are the same as without the field

Scenario: Windows line endings in the input
  Given a breadcrumb record and a claim record whose lines end in
    "\r\n"
  When I pipe them to "bcr audit"
  Then the two records are printed with their lines still ending in
    "\r\n"
  And the claim's verdict is the same as with lines that end in "\n"

Scenario: A last line of input with no line ending
  Given records whose last line does not end in "\n"
  When I pipe them to "bcr audit"
  Then standard output is the input, then "\n", then the records
    added

Scenario: A line that is not a record
  Given an input whose line 2 is "claim	verify	has-line	docs/bcr.md"
  When I pipe it to "bcr audit"
  Then a message starting "bcr: " that gives line 2 is printed to
    standard error
  And nothing is printed to standard output
  And the exit status is 2

Scenario: A claim record bcr extract could not print
  Given claim records whose targets are "/etc/passwd" and "../x.md", a
    claim record whose kind is "contains", and a has-line claim record
    whose text ends with a space
  When I pipe each to "bcr audit"
  Then a message starting "bcr: " is printed to standard error
  And no file is read
  And the exit status is 2

Scenario: An operand
  When I run "bcr audit records.tsv"
  Then its usage line is printed to standard error
  And the exit status is 2

Scenario: Output that cannot be written
  Given a valid set of records, and a standard output that cannot be
    written
  When I pipe them to "bcr audit"
  Then a message starting "bcr: " is printed to standard error
  And the exit status is 2

Scenario: The whole pipe
  Given a file whose breadcrumb has one claim that holds and one claim
    entry in double quotes
  When I run "bcr extract a.md | bcr verify | bcr audit"
  Then bcr extract prints the problem of the second claim
  And bcr audit prints the first claim as Confirmed
  And bcr audit prints the breadcrumb as Undecided
```
