---
breadcrumb:
  id: extract
  type: spec
  links:
    - follows ADR-0002
    - follows ADR-0006
    - follows ADR-0007
    - follows ADR-0008
    - follows ADR-0010
    - follows ADR-0011
    - follows ADR-0012
---
# Feature: extract

`bcr extract` reads the breadcrumb of each file it is given, and
prints it as records other programs can use.

## Blueprint

### Context

`bcr extract` is the first command of the new model. It produces data:
everything later, from the check of unique ids to the cache in
`.bcr/`, needs breadcrumbs as records, not as front matter. Feedback
for an author, printing only problems, comes later and is not this
command's job.

The caller chooses the files, for example with `git ls-files` or
`find`; `bcr extract` does not walk directories. A rule about one file
can be checked on any set of files, so this command checks the form of
each breadcrumb (ADR-0002, ADR-0010, ADR-0012). A rule about the whole
repository, such as unique ids (ADR-0003) or links that point to an
existing breadcrumb, needs every file and is not checked here.

`bcr extract` replaces the old `bcr list`, which read `breadcrumbs/`.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command | Reads the operands, opens the files, prints records and problems, sets the exit status | Infrastructure |
| Front matter reader | Splits a file's front matter from the rest; parses it with `go.yaml.in/yaml/v3` into a node tree; turns the tree into the core's types, each value as the text written, with its line (ADR-0010) | Infrastructure |
| Core | Holds the breadcrumb, the link and the problem; checks ADR-0002, ADR-0010's part of YAML and ADR-0012; builds the breadcrumb or reports what is wrong | Core domain |

- The core imports only the standard library, and reads files through
  `io/fs` (ADR-0007). Its package is named by the PBI that creates it.
- A file's front matter starts on its first line, which is exactly
  `---`, and ends at the next line that is exactly `---`. A file whose
  first line is not `---` has no front matter.
- Line numbers count from 1, from the first line of the file.

### Interface

```
bcr extract FILE...
```

Each operand is a file, handled on its own, in the order given.

For a file with a valid breadcrumb, `bcr extract` prints to standard
output one `breadcrumb` record, then one `link` record for each entry
of `links`, in the order they are written. Fields are separated by a
tab (ADR-0006):

```
breadcrumb	ID	TYPE	PATH
link	ID	VERB	OBJECT
```

- `PATH` is the operand as given.
- `ID` in a `link` record is the id of the breadcrumb the link belongs
  to; `VERB` and `OBJECT` are the two words of its entry (ADR-0012).
- A consumer selects records by their first field. New kinds of record
  may be added (ADR-0011).

A file with no front matter, or with front matter and no `breadcrumb`
key, prints nothing.

A problem is printed to standard error as `PATH:LINE: MESSAGE`
(ADR-0006). A file with a problem prints no records; the other files
are still read.

Exit status:

- 0: every file was read, and none had a problem.
- 1: at least one file had a problem in its breadcrumb.
- 2: `bcr extract` was given no operand, or a file could not be read.
  2 is returned even when another file had a problem.

### Constraints

- Keys under `breadcrumb` other than `id`, `type` and `links` are not
  covered by this spec yet.

## Contract

### Definition of Done

- [ ] Each Scenario below has a test.
- [ ] `docs/bcr.md` describes `bcr extract` under `### extract` and no
      longer has `### list`; `docs/bcr.1` is generated again.
- [ ] The old `bcr list` and its tests are removed.
- [ ] `go list -f '{{.Imports}}'` on the core's packages lists only
      the standard library and other core packages.
- [ ] In this repository, `git ls-files '*.md' | xargs bcr extract`
      prints one `breadcrumb` record for each file with a breadcrumb,
      writes nothing to standard error, and exits 0.

### Regression Guardrails

- `docs/bcr.md` documents `bcr extract`.

  Claims:
  - `docs/bcr.md` contains `### extract`

- `go.yaml.in/yaml/v3` is the only module `bcr` requires directly
  (ADR-0005, ADR-0010). Checked by hand: the `require` block of
  `go.mod`.

### Scenarios

```gherkin
Scenario: A breadcrumb with links
  Given tasks/PBI-00001.md whose front matter is
    """
    breadcrumb:
      id: PBI-00001
      type: PBI
      links:
        - implements ADR-0001
        - changes asdlc
    """
  When I run "bcr extract tasks/PBI-00001.md"
  Then standard output is
    """
    breadcrumb	PBI-00001	PBI	tasks/PBI-00001.md
    link	PBI-00001	implements	ADR-0001
    link	PBI-00001	changes	asdlc
    """
  And the exit status is 0

Scenario: A breadcrumb with no links
  Given a file whose breadcrumb has "links: []"
  When I extract it
  Then standard output has its breadcrumb record and no link record
  And the exit status is 0

Scenario: A file with no breadcrumb
  Given a file with no front matter, and a file whose front matter has
    no breadcrumb key
  When I extract both
  Then nothing is printed
  And the exit status is 0

Scenario: Several files
  Given files a.md and b.md, each with a breadcrumb
  When I run "bcr extract b.md a.md"
  Then the records of b.md come before the records of a.md

Scenario: A missing property
  Given a file whose breadcrumb key has no id, no type or no links
  When I extract it
  Then a problem is printed at the line of the breadcrumb key
  And the file prints no records
  And the exit status is 1

Scenario: Links with no list
  Given a file whose breadcrumb has a bare "links:"
  When I extract it
  Then a problem is printed at the line of "links:"
  And the exit status is 1

Scenario: White space in a link entry
  Given a link entry "  implements   ADR-0001 " with tabs or
    non-breaking spaces among the spaces
  When I extract it
  Then the link record is "link	ID	implements	ADR-0001"
  And no problem is printed

Scenario: A link entry that is not two words
  Given link entries "ADR-0001" and "depends on ADR-0003"
  When I extract them
  Then a problem is printed at the line of each
  And the exit status is 1

Scenario: The same link twice
  Given entries "implements ADR-0001" and "implements  ADR-0001" in one
    breadcrumb
  When I extract it
  Then a problem is printed at the line of the second
  And the exit status is 1

Scenario: A link to itself
  Given a breadcrumb with id PBI-00001 and a link "implements PBI-00001"
  When I extract it
  Then a problem is printed at the line of that link
  And the exit status is 1

Scenario: YAML outside the part ADR-0002 allows
  Given a breadcrumb that uses an anchor, an alias, a tag, a quoted
    value, a value over several lines, or a flow collection other
    than "[]"
  When I extract it
  Then a problem is printed at the line of each
  And the exit status is 1

Scenario: A value that looks like a number
  Given a breadcrumb whose id is written "0001"
  When I extract it
  Then its record holds the id "0001"

Scenario: One bad file among good ones
  Given a.md with a valid breadcrumb and b.md with a problem
  When I run "bcr extract a.md b.md"
  Then the records of a.md are printed
  And the problem of b.md is printed to standard error
  And the exit status is 1

Scenario: A file that cannot be read
  Given an operand that names no readable file
  When I extract it with a valid file
  Then a message starting "bcr: " is printed to standard error
  And the valid file's records are printed
  And the exit status is 2

Scenario: No operand
  When I run "bcr extract" with no operand
  Then its usage line is printed to standard error
  And the exit status is 2

Scenario: Front matter that does not close
  Given a file whose first line is "---" and that has no other line
    that is exactly "---"
  When I extract it
  Then a problem is printed at line 1
  And the exit status is 1
```
