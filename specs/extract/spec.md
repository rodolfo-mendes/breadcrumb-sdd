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
    - follows ADR-0013
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
| Front matter reader, `internal/frontmatter` | Splits a file's front matter from the rest; parses it with `go.yaml.in/yaml/v3` into a node tree; checks ADR-0002's part of YAML (ADR-0010); hands the core the properties under the `breadcrumb` key, each value as the text written or a list of such texts, with its line (ADR-0013) | Infrastructure |
| Core, `internal/crumb` | Holds the breadcrumb, the link and the problem; checks the breadcrumb's properties (ADR-0002) and its link entries (ADR-0012); builds the breadcrumb or reports what is wrong | Core domain |

- The core imports only the standard library, and knows no file format
  (ADR-0007, ADR-0013).
- The front matter reader finds the problems of the format, and the
  core the others. The command prints both alike (ADR-0013).
- A file's front matter starts on its first line, which is exactly
  `---`, and ends at the next line that is exactly `---`. A file whose
  first line is not `---` has no front matter. A line ends at `\n` or
  `\r\n`, and a byte order mark before the first line is ignored.
- Front matter that is empty, is not a map, or has no `breadcrumb` key
  has no breadcrumb. Keys outside `breadcrumb` are not read, and
  comments are allowed (ADR-0002).
- A `breadcrumb` key with no value has no properties: the core reports
  each one missing.
- A value is the text written: `0001`, `true` and `~` are text. Only a
  value left empty has none.
- A second `breadcrumb` key, or a key written twice under it, is a
  problem at the second.
- Front matter that is not YAML is a problem at the first line it
  cannot be read up to, with the parser's message.
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
(ADR-0006). A file's problems are printed in order of line. A file
with a problem prints no records; the other files are still read.

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

Scenario: An id or a type that is not one word
  Given a file whose breadcrumb has an empty id, an id written as a
    list, or a type with white space in it
  When I extract it
  Then a problem is printed at the line of each
  And the exit status is 1

Scenario: Links with no list
  Given a file whose breadcrumb has a bare "links:", or "links:"
    followed by text on the same line
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
    value, a value over several lines, a flow collection other
    than "[]", a map as a value, or a list inside a list
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

Scenario: Front matter that is not YAML
  Given a file whose breadcrumb has "links: [a" on line 3
  When I extract it
  Then a problem is printed at line 3
  And the exit status is 1

Scenario: A key written twice
  Given a breadcrumb whose id is written on lines 3 and 5
  When I extract it
  Then a problem is printed at line 5
  And the exit status is 1

Scenario: Windows line endings
  Given a file with a valid breadcrumb, a byte order mark and lines
    that end in "\r\n"
  When I extract it
  Then its records are the same as with lines that end in "\n"

Scenario: Front matter that does not close
  Given a file whose first line is "---" and that has no other line
    that is exactly "---"
  When I extract it
  Then a problem is printed at line 1
  And the exit status is 1
```
