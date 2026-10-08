---
breadcrumb:
  id: extract
  type: spec
  links:
    - constrained_by ADR-0002
    - constrained_by ADR-0006
    - constrained_by ADR-0007
    - constrained_by ADR-0008
    - constrained_by ADR-0010
    - constrained_by ADR-0011
    - constrained_by ADR-0012
    - constrained_by ADR-0013
    - constrained_by ADR-0014
    - constrained_by ADR-0015
    - constrained_by ADR-0017
    - constrained_by ADR-0018
    - constrained_by ADR-0019
    - constrained_by ADR-0021
  claims:
    - 'docs/bcr.md has-line ### extract'
---
# Feature: extract

`bcr extract` reads the breadcrumb of each file it is given, or of
each file `.breadcrumbs` names, and prints it as records other
programs can use.

## Blueprint

### Context

`bcr extract` is the first command of the new model. It produces data:
everything later, from the check of unique ids to the cache in
`.bcr/`, needs breadcrumbs as records, not as front matter. Feedback
for an author, printing only problems, comes later and is not this
command's job.

The files come from the caller, as operands, for example from
`git ls-files` or `find`. With no operand, they come from the
repository: `.breadcrumbs` names them by pattern (ADR-0014), and
`bcr extract` walks the repository only to find the files those
patterns name.

A rule about one file can be checked on any set of files, so this
command checks the form of each breadcrumb (ADR-0002, ADR-0010,
ADR-0012), its claims included (ADR-0017). It reads claims, but does
not check them against their targets: that needs the targets' files,
and is the job of a later stage. A rule about the whole set, such as unique ids (ADR-0003)
or links that point to an existing breadcrumb, is checked by
`bcr verify`, which reads the records this command prints (ADR-0015).

Standard error does not travel down a pipe, so each problem this
command finds is printed as a record too, for the stages after it
(ADR-0021).

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command | Reads the operands, opens the files, prints records and problems, each problem also as a record (ADR-0021), sets the exit status | Infrastructure |
| File set reader | With no operand: reads `.breadcrumbs`, checks its patterns, and finds the files they name from the root of the repository (ADR-0014) | Infrastructure |
| Front matter reader, `internal/frontmatter` | Splits a file's front matter from the rest; parses it with `go.yaml.in/yaml/v3` into a node tree; checks ADR-0002's part of YAML (ADR-0010); allows single quotes only in the entries of `claims`, and reports an entry of `claims` that is not single-quoted or has a comment after it (ADR-0017); hands the core the properties under the `breadcrumb` key, each value as the text written or a list of such texts, with its line (ADR-0013) | Infrastructure |
| Core, `internal/crumb` | Holds the breadcrumb, the link, the claim and the problem; checks the breadcrumb's properties (ADR-0002, ADR-0018), its link entries (ADR-0012) and its claim entries, with the kinds it knows (ADR-0017, ADR-0019); builds the breadcrumb or reports what is wrong | Core domain |

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
  value left empty has none. The text of a single-quoted entry is what
  is inside the quotes, with each `''` read as `'`.
- A second `breadcrumb` key, or a key written twice under it, is a
  problem at the second.
- Front matter that is not YAML is a problem at the first line it
  cannot be read up to, with the parser's message.
- Line numbers count from 1, from the first line of the file.
- A line of `.breadcrumbs` ends at `\n` or `\r\n`, and a byte order
  mark before the first is ignored. A line of only white space is
  blank.
- The walk reads only regular files: a directory or a symbolic link a
  pattern matches is not read. A directory named `.git` is not
  entered, wherever it is. A directory that cannot be read is treated
  as a file that cannot be read.

### Interface

```
bcr extract [FILE...]
```

Given operands, each is a file, handled on its own, in the order
given. Given none, `bcr extract` reads every file `.breadcrumbs`
names, in order of path compared as bytes (ADR-0014).

For a file with a valid breadcrumb, `bcr extract` prints to standard
output one `breadcrumb` record, then one `link` record for each entry
of `links`, then one `claim` record for each valid entry of `claims`,
each in the order they are written. Fields are separated by a tab
(ADR-0006):

```
breadcrumb	ID	TYPE	PATH	LINE
link	ID	VERB	OBJECT	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
```

- `PATH` is the operand as given, or, with no operand, the file's path
  from the root of the repository, with `/` between its parts.
- `LINE` is the line where the record was written: for a `breadcrumb`
  record, the line of its `id`; for a `link` record, the line of its
  entry. A `link` record carries its `PATH` and `LINE` so that a
  check of the whole set can report it where it is (ADR-0015).
- `bcr extract` checks neither the
  name nor the path of a file: which files to read, and how they are
  named, is the caller's choice. A file of any kind whose first line
  is not `---` prints nothing.
- `ID` in a `link` record is the id of the breadcrumb the link belongs
  to; `VERB` and `OBJECT` are the two words of its entry (ADR-0012).
- In a `claim` record, `ID` is the id of the breadcrumb the claim
  belongs to, and `LINE` the line of its entry. `TARGET`, `KIND` and
  `ARGUMENT` are the three parts of the entry, after its quotes are
  removed (ADR-0017): `TARGET` is the file the claim is about, and
  `PATH` the file the claim is written in.
- `ARGUMENT` holds no tab, since a claim's entry holds none. No field
  of a `claim` record is empty while `has-line` is its only kind.
- A consumer selects records by their first field, and reads their
  fields by position. New kinds of record may be added (ADR-0011), and
  new fields may be added at the end of a record.

A file with no front matter, or with front matter and no `breadcrumb`
key, prints nothing.

A problem is printed to standard error as `PATH:LINE: MESSAGE`
(ADR-0006). A file's problems are printed in order of line. A file
with a problem in its id, type or links prints no `breadcrumb`, `link`
or `claim` record; the other files are still read.

Each problem is also printed to standard output, as a record
(ADR-0021):

```
problem	PATH	LINE	MESSAGE
```

- `PATH`, `LINE` and `MESSAGE` are those of the line written to
  standard error. A tab, a `\n` or a `\r` inside `MESSAGE` is written
  as a space, so that the record stays one line of four fields.
- A file's `problem` records come after its other records, in order of
  line.
- A `problem` record has no id: a problem can be found before an id is
  read. It belongs to the breadcrumb whose record has the same `PATH`,
  when there is one.
- A message that starts `bcr: `, such as for a file that cannot be
  read, is not a problem, and prints no record.

A problem in a claim entry drops only that claim (ADR-0017). The
breadcrumb's record, its links and its other claims are still printed,
and the exit status is 1. A claim entry has a problem when:

- it is not single-quoted, or has a comment after it on its line;
- it is outside ADR-0002's part of YAML in another way: it has an
  anchor, an alias or a tag, is a list or a map, or is written over
  several lines;
- it has fewer than three parts, or a part is separated from the next
  by anything other than one space;
- its `TARGET` starts with `/`, has a part that is `..`, or holds
  white space;
- its `KIND` is not a kind `bcr` knows. The only kind is `has-line`
  (ADR-0019);
- its argument is one its kind does not allow. For `has-line`: an
  empty text, or a text that starts or ends with a space or a tab, or
  holds a tab.

One problem is printed for each such entry: the first it has, in the
order above.

`claims` is optional (ADR-0018). A bare `claims:`, or `claims`
written as text, is a problem at its line, and the breadcrumb prints
no records, as with `links`. `claims: []` is the same as no `claims`.

A pattern may not use `**`, start or end with `/`, be empty after
`!`, or be malformed, such as with a `[` that does not close.
A pattern `.breadcrumbs` does not allow is a problem at its line in
`.breadcrumbs`, printed as a `problem` record too, whose `PATH` is
`.breadcrumbs`. When `.breadcrumbs` has a problem, no file is read,
since the set it names is not known.

Exit status:

- 0: every file was read, and none had a problem.
- 1: at least one file had a problem in its breadcrumb, or
  `.breadcrumbs` had a problem.
- 2: `bcr extract` was given no operand and there is no
  `.breadcrumbs`, or a file could not be read. 2 is returned even when
  another file had a problem.

### Constraints

- Keys under `breadcrumb` other than `id`, `type`, `links` and
  `claims` are not read.
- A dropped claim prints no `claim` record. Its `problem` record says
  a problem was found at its line, not what the claim was (ADR-0017,
  ADR-0021).
- A file that cannot be read prints no record: it shows only on
  standard error and in the exit status. In a pipe, the exit status
  shows only with `set -o pipefail`.
- `bcr extract` does not open a claim's target, so a claim about a file
  that does not exist is printed like any other.

## Contract

### Definition of Done

- [x] Each Scenario below has a test.
- [x] The old `bcr list` and its tests are removed.
- [x] `docs/bcr.md` describes `bcr extract` under `### extract`,
      including the reading of `.breadcrumbs` and the `PATH` and
      `LINE` fields; `docs/bcr.1` is generated again.
- [x] `go list -f '{{.Imports}}'` on the core's packages lists only
      the standard library and other core packages.
- [x] This repository has a `.breadcrumbs` with three lines:
      `docs/adrs/*.md`, `specs/*/spec.md` and `tasks/*.md`.
- [x] In this repository, `bcr extract` with no operand prints one
      `breadcrumb` record for each file with a breadcrumb, writes
      nothing to standard error, and exits 0.
- [x] In this repository, `git ls-files '*.md' | xargs bcr extract`
      prints the same records, in the order of its operands.
- [x] `docs/bcr.md` describes the `claim` record and what makes a
      claim entry a problem; `docs/bcr.1` is generated again.
- [x] `go.yaml.in/yaml/v3` is still the only module `bcr` requires
      directly: reading claims adds no library.
- [x] `docs/bcr.md` describes the `problem` record; `docs/bcr.1` is
      generated again.
- [x] In this repository, `bcr extract` prints no `problem` record.

### Regression Guardrails

- `docs/bcr.md` documents `bcr extract`.

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
    breadcrumb	PBI-00001	PBI	tasks/PBI-00001.md	3
    link	PBI-00001	implements	ADR-0001	tasks/PBI-00001.md	6
    link	PBI-00001	changes	asdlc	tasks/PBI-00001.md	7
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
  And the file prints no breadcrumb, link or claim record
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
  Then the link record's verb is "implements" and its object
    "ADR-0001"
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
    value outside claims, a value over several lines, a flow collection other
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

Scenario: No operand and no .breadcrumbs
  Given a repository with no .breadcrumbs at its root
  When I run "bcr extract" with no operand
  Then a message starting "bcr: " is printed to standard error
  And the exit status is 2

Scenario: The files .breadcrumbs names
  Given a .breadcrumbs with the lines "docs/adrs/*.md" and "tasks/*.md"
  And breadcrumbs in docs/adrs/ADR-0002.md, tasks/PBI-00001.md and
    specs/asdlc/spec.md
  When I run "bcr extract" with no operand
  Then the records of docs/adrs/ADR-0002.md come before those of
    tasks/PBI-00001.md
  And nothing is printed for specs/asdlc/spec.md

Scenario: A pattern that excludes
  Given a .breadcrumbs with the lines "tasks/*.md" and
    "!tasks/PBI-00009.md"
  When I run "bcr extract" with no operand
  Then nothing is printed for tasks/PBI-00009.md
  And the records of the other files in tasks/ are printed

Scenario: A pattern without a slash
  Given a .breadcrumbs with the line "*.md"
  And breadcrumbs in README.md and docs/adrs/ADR-0002.md
  When I run "bcr extract" with no operand
  Then the records of README.md are printed
  And nothing is printed for docs/adrs/ADR-0002.md

Scenario: A pattern .breadcrumbs does not allow
  Given a .breadcrumbs whose line 2 is "docs/**/*.md"
  When I run "bcr extract" with no operand
  Then a problem is printed at ".breadcrumbs:2"
  And a problem record is printed whose PATH is ".breadcrumbs" and
    whose LINE is 2
  And no file is read
  And the exit status is 1

Scenario: What the walk does not enter
  Given a .breadcrumbs with the line "*/*.md"
  And a breadcrumb in .git/x.md, and a symbolic link docs/link.md to a
    file with a breadcrumb
  When I run "bcr extract" with no operand
  Then nothing is printed for .git/x.md or docs/link.md

Scenario: Operands and .breadcrumbs
  Given a .breadcrumbs that does not name notes/draft.md
  When I run "bcr extract notes/draft.md"
  Then the records of notes/draft.md are printed

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

Scenario: A breadcrumb with claims
  Given specs/verify/spec.md whose front matter is
    """
    breadcrumb:
      id: verify
      type: spec
      links:
        - follows ADR-0016
      claims:
        - 'docs/bcr.md has-line ### verify'
        - 'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'
    """
  When I run "bcr extract specs/verify/spec.md"
  Then standard output is
    """
    breadcrumb	verify	spec	specs/verify/spec.md	3
    link	verify	follows	ADR-0016	specs/verify/spec.md	6
    claim	verify	has-line	docs/bcr.md	### verify	specs/verify/spec.md	8
    claim	verify	has-line	docs/adrs/ADR-0001-adopt-asdlc.md	Status: Accepted	specs/verify/spec.md	9
    """
  And the exit status is 0

Scenario: No claims
  Given a breadcrumb with no claims key, and one with "claims: []"
  When I extract them
  Then each prints its breadcrumb and link records, and no claim record
  And the exit status is 0

Scenario: Claims with no list
  Given a breadcrumb with a bare "claims:", or "claims:" followed by
    text on the same line
  When I extract it
  Then a problem is printed at the line of "claims:"
  And the file prints no breadcrumb, link or claim record
  And the exit status is 1

Scenario: A quote inside a claim
  Given a claim entry 'cmd/bcr/main.go has-line return fmt.Sprintf(''%s'', id)'
  When I extract it
  Then the claim record's argument is "return fmt.Sprintf('%s', id)"

Scenario: A claim entry that is not single-quoted
  Given a plain claim entry, and a claim entry in double quotes
  When I extract them
  Then a problem is printed at the line of each
  And no claim record is printed for either

Scenario: A claim with a comment after it
  Given the claim entries "- docs/bcr.md has-line ### verify" and
    "- 'docs/bcr.md has-line ### verify' # why"
  When I extract them
  Then a problem is printed at the line of each
  And no claim record is printed for either

Scenario: A claim entry that is not three parts
  Given the claim entries 'docs/bcr.md', 'docs/bcr.md has-line' and
    'docs/bcr.md  has-line ### verify'
  When I extract them
  Then a problem is printed at the line of each

Scenario: A target outside the rules
  Given claims whose targets are "/etc/passwd", "../x.md" and
    "docs/../x.md"
  When I extract them
  Then a problem is printed at the line of each

Scenario: A kind bcr does not know
  Given the claim entries 'docs/bcr.md contains ### verify' and
    'docs/bcr.md has-lines ### verify'
  When I extract them
  Then a problem that names the kind is printed at the line of each

Scenario: A has-line text that no line can equal
  Given has-line claims whose texts start with a space, end with a
    space, or hold a tab
  When I extract them
  Then a problem is printed at the line of each

Scenario: An invalid claim drops only itself
  Given a breadcrumb with two links and three claims, the second of
    which has a kind bcr does not know
  When I extract it
  Then its breadcrumb record, its two link records and the claim
    records of the first and third claims are printed
  And a problem is printed at the line of the second claim
  And the exit status is 1

Scenario: A claim whose target does not exist
  Given a has-line claim whose target is no file in the repository
  When I extract it
  Then its claim record is printed
  And the exit status is 0

Scenario: A problem is also a record
  Given tasks/PBI-00001.md whose breadcrumb has no type, at line 2
  When I run "bcr extract tasks/PBI-00001.md"
  Then standard error is
    """
    tasks/PBI-00001.md:2: breadcrumb has no type
    """
  And standard output is
    """
    problem	tasks/PBI-00001.md	2	breadcrumb has no type
    """
  And the exit status is 1

Scenario: The record of a dropped claim
  Given a.md whose breadcrumb has one link and two claims, the first of
    which, at line 8, has a kind bcr does not know
  When I run "bcr extract a.md"
  Then standard output has its breadcrumb record, its link record and
    the claim record of the second claim
  And after them a problem record whose PATH is "a.md" and whose LINE
    is 8
  And the exit status is 1

Scenario: Problem records among several files
  Given a.md with two problems, at lines 3 and 5, and b.md with a
    valid breadcrumb
  When I run "bcr extract a.md b.md"
  Then the problem records of a.md are printed in order of line
  And the records of b.md come after them

Scenario: A message that is not a problem
  Given an operand that names no readable file
  When I extract it
  Then a message starting "bcr: " is printed to standard error
  And no problem record is printed

Scenario: A problem's message on one line
  Given a problem whose message holds a tab or a line ending
  When its record is printed
  Then the record is one line of four fields, with a space in place of
    each
```
