# bcr(1)

## NAME

bcr - audit the breadcrumbs of a repository

## SYNOPSIS

```
bcr extract [FILE...]
bcr verify
bcr audit
bcr report
bcr filter [-i ID] [-o ID] [-t TYPE] [-k KIND]
bcr init [-a FILE] [-l LAYOUT]
```

## DESCRIPTION

`bcr` is the toolkit of Breadcrumb. Four of its commands are the
stages of a pipe:

```
bcr extract | bcr verify | bcr audit | bcr report
```

`bcr extract` reads the breadcrumbs of a repository from the front
matter of its files and prints them as records. `bcr verify` checks
the set of breadcrumbs, and its shape when the repository declares
one in `breadcrumb.rules`, `bcr audit` checks their claims against
the repository's files, and `bcr report` writes a page of what the
pipe found. Each stage reads the records of the one before it from
standard input. `bcr filter` prints only the records that match its
flags, wherever records flow. `bcr init` sets a repository up for the
pipe. Run `bcr` from the root of the repository.

This document is the user documentation of `bcr`: its commands,
flags, inputs, outputs and exit codes. Where it and a command's spec
disagree, the spec is right (ADR-0008).

Every command follows the same conventions (ADR-0006):

- It is run as `bcr COMMAND [FLAGS] [OPERANDS]`. Flags come before
  operands; the first operand, or `--`, ends them.
- Flags are in getopt style: a short flag is one letter, such as
  `-a`, and short flags can be combined, as in `-ab`; a long flag is a
  name, such as `--agents-file`. A flag's value follows it:
  `-x VALUE`, `-xVALUE`, `--name=VALUE` or `--name VALUE`.
- Results go to standard output, and nothing else does. Messages go to
  standard error and start with `bcr: `. A usage error is followed by
  the usage line.
- The exit status says whether the command found something wrong (see
  EXIT STATUS).

## COMMANDS

### extract

```
bcr extract [FILE...]
```

Reads the breadcrumb of each file it is given, or of each file
`.breadcrumbs` names, from the file's front matter, and prints it as
records other programs can use.

Flags: none.

Operands: the files to read, in the order given. Which files to read
is the caller's choice, for example with `git ls-files` or `find`;
`bcr extract` does not check the name or the path of a file. Given
operands, `.breadcrumbs` is not read.

With no operand, `bcr extract` reads every file that `.breadcrumbs`,
in the current directory, names, in order of path compared as bytes.
Run it from the root of the repository. The files are found from the
current directory: a directory named `.git` is not entered, symbolic
links are not followed, and only regular files are read.

`.breadcrumbs` holds one pattern on each line. Blank lines, and lines
that start with `#`, are ignored:

```
# the decisions, the specs and the tasks
docs/adrs/*.md
specs/*/spec.md
tasks/*.md
!tasks/README.md
```

A pattern is matched against the whole path of a file from the
current directory, with `/` between its parts. `*` matches any
characters except `/`, `?` one character except `/`, and `[...]` one
character from a set. Unlike in `.gitignore`, a pattern with no `/`
names only files at the root: `*.md` does not name `docs/bcr.md`. A
pattern that starts with `!` excludes the files it matches. When
several lines match a file, the last one decides; a file no line
matches is not read.

A pattern may not use `**`, start or end with `/`, be empty after `!`,
or be malformed, such as with a `[` that does not close. Such a line
is a problem, written to standard error as `.breadcrumbs:LINE:
MESSAGE` and to standard output as a `problem` record, and then no
file is read.

Standard input: not read.

Output: for a file with a valid breadcrumb, one `breadcrumb` record on
standard output, then one `link` record for each entry of its
`links`, then one `claim` record for each valid entry of its `claims`,
each in the order written. Fields are separated by a tab:

```
breadcrumb	ID	TYPE	PATH	LINE
breadcrumb	ID	TYPE	PATH	LINE	TITLE
link	ID	VERB	OBJECT	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
```

A `breadcrumb` record has `TITLE`, the breadcrumb's title as written,
only when the breadcrumb has one; otherwise it has five fields.

`PATH` is the file as given, or, with no operand, its path from the
current directory, with `/` between its parts. `LINE` is the line of the file where the
record was written, counted from 1: for a `breadcrumb` record, the
line of its `id`; for a `link` record, the line of its entry. In a
`link` record, `ID` is the id of the breadcrumb the link belongs to,
and `VERB` and `OBJECT` are the two words of its entry.

In a `claim` record, `ID` is the id of the breadcrumb the claim
belongs to, and `LINE` the line of its entry. `TARGET`, `KIND` and
`ARGUMENT` are the three parts of the entry, without its quotes:
`TARGET` is the file the claim is about, and `PATH` the file the claim
is written in. `ARGUMENT` may hold spaces, but no tab. `bcr extract`
does not open `TARGET`: a claim about a file that does not exist is
printed like any other.

Select records by their first field, and read their fields by
position: new kinds of record may be added, and new fields may be
added at the end of a record.

A file whose first line is not `---` has no front matter, and prints
nothing; so does front matter with no `breadcrumb` key. A line may end
in `\n` or `\r\n`.

A breadcrumb is written under the `breadcrumb` key of the front
matter, in a small part of YAML: plain one-line values, lists of them,
and `[]`. It has an `id` and a `type`, each one word, and `links`, a
list of entries of a verb and an id. A file whose breadcrumb breaks a
rule prints no `breadcrumb`, `link` or `claim` record; each problem is
written to standard error, in order of line:

```
PATH:LINE: MESSAGE
```

Each problem is also written to standard output, as a record, so that
the commands after `bcr extract` in a pipe can see it:

```
problem	PATH	LINE	MESSAGE
```

`PATH`, `LINE` and `MESSAGE` are those written to standard error, with
a space in place of any tab or line ending in `MESSAGE`. A file's
`problem` records come after its other records, in order of line. A
`problem` record has no id: it belongs to the breadcrumb whose record
has the same `PATH`, when there is one.

A breadcrumb may also have a `title`, one line of text that says what
its artifact is about. `bcr report` shows it beside the id:

```
breadcrumb:
  id: ADR-0019
  type: ADR
  title: A has-line claim matches a trimmed whole line
  links: []
```

`title` may be left out. It is written plain, without quotes, like
`id` and `type`, and may hold spaces. A `title` with no value, a
`title` that is a list, and a title that holds a tab are problems like
any other, and the file prints no records. `bcr` reads no heading of
the file: a title comes only from this key. Two breadcrumbs may have
the same title.

A breadcrumb may also have `claims`, a list of statements about the
repository's files. `claims` may be left out, and `claims: []` means
the same; a `claims` with no list is a problem like any other, and the
file prints no records. Each entry is written on one line, in single
quotes, as `TARGET KIND ARGUMENT`, with a `'` inside it written `''`:

```
breadcrumb:
  id: verify
  type: spec
  links:
    - constrained_by ADR-0016
  claims:
    - 'docs/bcr.md has-line ### verify'
```

`TARGET` is a path from the root of the repository, with `/` between
its parts. The only `KIND` is `has-line`: its `ARGUMENT` is the text a
line of `TARGET` must equal, once the spaces and tabs at the line's
start and end are removed.

A problem in an entry of `claims` drops only that claim: it is written
to standard error, the breadcrumb's record, its links and its other
claims are still printed, and the exit status is 1. The dropped claim
prints no `claim` record, only its `problem` record. An entry has a
problem when:

- it is not in single quotes, has a comment after it on its line, or
  is written in any other way the breadcrumb's part of YAML does not
  allow, such as over several lines;
- it has fewer than three parts, or a part is separated from the next
  by anything other than one space;
- its `TARGET` starts with `/`, has a part that is `..`, or holds
  white space;
- its `KIND` is not `has-line`;
- its `has-line` text is empty, starts or ends with a space or a tab,
  or holds a tab.

One problem is written for each such entry.

A file or a directory that cannot be read is written to standard
error as a message starting `bcr: `, and prints no record. The other
files are still read.

Exit status:

- 0: every file was read, and none had a problem.
- 1: a file had a problem in its breadcrumb, or `.breadcrumbs` had a
  problem.
- 2: `bcr` was used wrongly, there is no operand and no
  `.breadcrumbs`, or a file could not be read, even when another file
  had a problem.

Examples:

```
bcr extract
git ls-files '*.md' | xargs bcr extract
bcr extract tasks/*.md | grep '^link' | cut -f2,4
bcr extract | grep '^claim' | cut -f4,5
bcr extract 2>/dev/null | grep '^problem' | cut -f2-
```

### verify

```
bcr verify
```

Reads the records `bcr extract` prints, checks the rules about the
whole set of breadcrumbs they describe, and passes every record on,
with a record for each problem it found:

```
bcr extract | bcr verify
```

Flags: none.

Operands: none. `bcr verify` reads no file that carries a breadcrumb:
which files make up the set is decided by `bcr extract`, from its
operands or from `.breadcrumbs`. The only file it reads is
`breadcrumb.rules`, in the current directory, when there is one (see
The shape of a repository, below).

Standard input: the records, as `bcr extract` prints them, read to the
end. A line may end in `\n` or `\r\n`. Of a `claim` record, only its
`ID`, `PATH` and `LINE` are read. A record of a kind other than
`breadcrumb`, `link` and `claim` is not checked, a `problem` record
included, and neither are fields after the ones `bcr extract` prints
today; both are still copied. With no input, there is no problem.

Output: standard input, copied to standard output byte for byte.
Every line comes out in the order read, with its line ending as read,
records of any kind included, so a later command in the pipe gets what
`bcr extract` printed. `bcr verify` reads its whole input before it
writes. It copies the input even when it finds a problem, and prints
nothing to standard output when a line is not a record. To see only
the problems, send standard output to `/dev/null`.

Each problem is written to standard error, in order of `PATH` compared
as bytes, and then of `LINE`:

```
PATH:LINE: MESSAGE
```

After the copy, each problem is also written to standard output, in
the same order, as a record like the one `bcr extract` prints for its
own problems:

```
problem	PATH	LINE	MESSAGE
```

So the output is equal to the input when no problem is found, and
starts with the input when one is. When the last line of the input has
no line ending, a `\n` is written before the first record added.

`PATH` and `LINE` are the ones of the record the problem is about.
What is a problem:

- Two or more breadcrumbs with the same id, or with ids that differ
  only in letter case, such as `ADR-0001` and `adr-0001`. Each of them
  is a problem, at its `breadcrumb` record, and its message names
  where the others are. Case is compared with Unicode simple case
  folding.
- A link whose `OBJECT` is not the id of any breadcrumb in the set, at
  its `link` record. When `OBJECT` differs only in case from an id,
  the message names that id.

A link to an id that two breadcrumbs share is not a problem of its
own; the duplicate is.

The shape of a repository: a repository may declare which types its
breadcrumbs have, which links are allowed between them, and which
types carry claims, in a file called `breadcrumb.rules` at its root.
`bcr verify` reads it from the current directory. With no such file,
no shape is checked. Each line is of one of three kinds, with one tab
between its fields:

```
type	NAME
link	NAME	VERB	NAME
claims	NAME
```

- `type` declares a type a breadcrumb may have.
- `link` allows a link with `VERB` from a breadcrumb of the first type
  to a breadcrumb of the second.
- `claims` allows a breadcrumb of the type to carry claims.

A name or a verb is one word, compared as exact bytes: `spec` and
`Spec` are two types. Lines may come in any order. Blank lines, and
lines that start with `#`, are ignored; no comment follows a field.
For example:

```
# decisions above features above changes
type	ADR
type	spec
type	PBI
link	spec	constrained_by	ADR
link	PBI	changes	spec
claims	spec
```

The shape is closed: what no line allows is a problem, at the record
that breaks it.

- A breadcrumb whose type no `type` line declares, at its `breadcrumb`
  record: `type "Spec" is not declared in breadcrumb.rules`.
- A link when no `link` line has the type of the breadcrumb it is
  written in, its verb, and the type of the breadcrumb it points to,
  at its `link` record:
  `link "PBI implements ADR" matches no rule in breadcrumb.rules`.
- A claim of a breadcrumb whose type has no `claims` line, at its
  `claim` record:
  `a breadcrumb of type "PBI" may not carry claims under breadcrumb.rules`.

One cause is one problem. A link or a claim of a breadcrumb whose
type is not declared is not judged, and neither is a link to such a
breadcrumb, a link that points nowhere, or a link from or to an id
that two breadcrumbs share. The rules say which links are allowed,
not which are required: a breadcrumb with no link is not a problem.

A line `breadcrumb.rules` does not allow is a problem too, written as
`breadcrumb.rules:LINE: MESSAGE` and as a `problem` record whose
`PATH` is `breadcrumb.rules`:

- a line of a kind other than `type`, `link` and `claims`;
- a line with the wrong number of fields: 2 for `type` and `claims`,
  4 for `link`. A line typed with spaces between its fields is told
  to use a tab;
- a field that is empty or holds white space;
- a `link` or `claims` line that names a type no `type` line
  declares;
- a line that repeats an earlier one.

While `breadcrumb.rules` has a problem, the shape is not known, and
nothing is judged against it; ids and links are still checked as
above. A file with no `type` line, an empty one included, is a shape
with no type: every breadcrumb is then a problem. Deleting the file
turns the shape checks off.

A line of input that is not a record `bcr extract` could print, an
empty line included, stops `bcr verify`: it writes a message starting
`bcr: `, which gives the line's number in the input, and prints no
problem.

A breadcrumb `bcr extract` could not read has no record, so links to
its id are reported as pointing to no breadcrumb. Fix the problem
`bcr extract` printed first.

Exit status:

- 0: no problem was found.
- 1: at least one problem was found.
- 2: `bcr verify` was given an operand, its input could not be read,
  a line of its input was not a record, `breadcrumb.rules` exists and
  could not be read, or its output could not be written.

In a pipe, the shell returns the exit status of `bcr verify`, so a
problem `bcr extract` found is seen only on standard error. A script
that needs both runs with `set -o pipefail`, or saves the records
first.

Examples:

```
bcr extract | bcr verify > /dev/null
set -o pipefail; bcr extract | bcr verify > /dev/null
bcr extract > records.tsv && bcr verify < records.tsv | cmp - records.tsv
```

### audit

```
bcr audit
```

Reads the records `bcr extract` and `bcr verify` print, checks each
claim against the file it is about, and passes every record on, with
a verdict for each claim and each breadcrumb:

```
bcr extract | bcr verify | bcr audit
```

Flags: none.

Operands: none. Run `bcr audit` from the root of the repository: the
file a claim is about is found from the current directory.

Standard input: the records, as `bcr extract` and `bcr verify` print
them, read to the end. A line may end in `\n` or `\r\n`. `bcr audit`
reads the `breadcrumb`, `claim` and `problem` records. A record of
another kind, a `link` record included, is not checked, and neither
are fields after the ones printed today; both are still copied. With
no input, nothing is printed.

A verdict is `Confirmed`, `Refuted` or `Undecided`.

The verdict of a claim: a `has-line` claim is `Confirmed` when its
`TARGET` is a regular file and one of its lines, without its line
ending and without the spaces and tabs at its start and end, is equal
to the claim's text, byte for byte. Otherwise it is `Refuted`: so is a
claim whose `TARGET` does not exist, or is a directory, a symbolic
link, or anything else that is not a regular file. A symbolic link is
not followed. The files are read as they are on disk, uncommitted
changes included.

The verdict of a breadcrumb comes from its own claims, and from its
problems, the `problem` records with its `PATH`:

- `Refuted`, when at least one of its claims is `Refuted`;
- otherwise `Undecided`, when it has a problem, such as a claim
  `bcr extract` dropped;
- otherwise `Confirmed`, when it has at least one claim;
- `Undecided`, with the reason `no claims`, when it has no claim and
  no problem: nothing was checked, so nothing is confirmed.

No link carries a verdict from one breadcrumb to another.

Output: standard input, copied to standard output byte for byte, as
`bcr verify` does. After it, one `claim-verdict` record for each
`claim` record, in the order read, then one `verdict` record for each
`breadcrumb` record, in the order read. Fields are separated by a tab:

```
claim-verdict	ID	VERDICT	PATH	LINE
verdict	ID	VERDICT	PATH	LINE
verdict	ID	Undecided	PATH	LINE	no claims
```

`ID`, `PATH` and `LINE` are those of the `claim` or `breadcrumb`
record the verdict is about, so two breadcrumbs with the same id each
get their own. A `verdict` record has a sixth field only for a
breadcrumb with no claims. `bcr audit` reads its whole input, and
every file the claims are about, before it writes. When the last line
of the input has no line ending, a `\n` is written before the first
record added.

Each `Refuted` claim is written to standard error, in the order read,
at the `PATH` and `LINE` of its `claim` record. The message names the
file and what it lacks:

```
PATH:LINE: MESSAGE
```

A line of input that is not a record `bcr extract` could print stops
`bcr audit`: it writes a message starting `bcr: `, which gives the
line's number in the input, and prints nothing to standard output. So
does a `claim` record whose `TARGET` starts with `/` or has a part
that is `..`, whose `KIND` is not `has-line`, or whose text no line
can equal. A regular file that cannot be read stops it in the same
way.

Exit status:

- 0: no claim is `Refuted`.
- 1: at least one claim is `Refuted`.
- 2: `bcr audit` was given an operand, its input could not be read, a
  line of its input was not a record, a file a claim is about could
  not be read, or its output could not be written.

An `Undecided` breadcrumb does not change the exit status. A problem
`bcr extract` or `bcr verify` found fails the pipe only with
`set -o pipefail`.

Examples:

```
set -o pipefail; bcr extract | bcr verify | bcr audit > /dev/null
bcr extract | bcr verify | bcr audit 2>/dev/null | grep '^verdict' | cut -f2,3
bcr extract | bcr audit | grep -w Refuted
```

### report

```
bcr report
```

Reads the records `bcr extract`, `bcr verify` and `bcr audit` print,
and writes one page that shows them: every breadcrumb with its
verdict, every problem, every claim, and one diagram for each spec. It
is the last command of the pipe, and does not pass the records on:

```
bcr extract | bcr verify | bcr audit | bcr report > report.md
```

Flags: none.

Operands: none. `bcr report` reads no file of the repository.

Standard input: the records, read to the end. A line may end in `\n`
or `\r\n`. `bcr report` reads the `breadcrumb`, `link`, `claim`,
`problem`, `claim-verdict` and `verdict` records. A record of another
kind is not read, and neither are fields after the ones printed today.

Output: one page, in Markdown, on standard output. Its diagrams are
Mermaid, in fenced blocks marked `mermaid`; GitHub draws them wherever
it shows Markdown. The page has four sections, in this order:

- `## Breadcrumbs`: a line of counts, then a table with one row for
  each breadcrumb, with its verdict, its id, its type and where it is
  written, as `PATH:LINE`. Rows are in order of verdict, `Refuted`
  first, then `Undecided`, `Confirmed` and `not audited`, and then of
  id.
- `## Problems`: one item for each problem `bcr extract` and
  `bcr verify` printed, as `PATH:LINE: MESSAGE`, in order of `PATH`
  and then of `LINE`.
- `## Claims`: a table with one row for each claim, with its verdict,
  the id of its breadcrumb, the claim as it is written, and where it
  is written. Rows are in the same order as the breadcrumbs.
- `## Views`: one diagram for each breadcrumb whose type is `spec`,
  under a heading with its id, in order of id.

A section with nothing to show says so, as in `No problems.`.

A view draws the spec, every breadcrumb with a link to it, every
breadcrumb it has a link to, and the links among those. A breadcrumb
two links away is not in it, and a breadcrumb in no view is still in
the table. Each box shows a breadcrumb's id, its title when it has
one, its type and its verdict, and each of its claims after the
claim's own verdict. A breadcrumb with no title is shown by its id
alone. A title is shown in its box and nowhere else on the page. Each arrow goes
from the breadcrumb a link is written in to the one it points to, and
is labelled with the link's verb.

A verdict is shown as `bcr audit` printed it: `Confirmed`, `Refuted`,
`Undecided`, or `Undecided, no claims`. `bcr report` judges nothing. A
breadcrumb or a claim with no verdict in the input is shown as
`not audited`, so the page can be made without `bcr audit`.

The same records give the same page, byte for byte: it holds no date,
and its rows, boxes and arrows are in order of id, whatever the order
of the records.

An id, a title, a path, a verb, a claim or a message is written so
that it cannot change the page around it. In the tables and lists, a `\` comes
before each character Markdown reads as syntax, such as `|`, `` ` ``
and `<`. In a diagram, a box is named `n1`, `n2` and so on, never by
an id, and in its label every character Mermaid could read as syntax
is written as a code, such as `#35;` for `#`. `bcr report` writes no
`click`, no link and no directive into a diagram.

A line of input that is not a record the pipe could print stops
`bcr report`: it writes a message starting `bcr: `, which gives the
line's number in the input, and prints no page. So does a `verdict` or
`claim-verdict` record whose verdict is not `Confirmed`, `Refuted` or
`Undecided`.

Exit status:

- 0: the page was written, whatever the verdicts and the problems.
- 2: `bcr report` was given an operand, its input could not be read, a
  line of its input was not a record, or its output could not be
  written.

A `Refuted` claim does not change the exit status of `bcr report`. It
fails the pipe through `bcr audit`, under `set -o pipefail`.

Examples:

```
bcr extract | bcr verify | bcr audit | bcr report > report.md
set -o pipefail; bcr extract | bcr verify | bcr audit | bcr report > report.md
bcr extract | bcr report
```

### filter

```
bcr filter [-i ID] [-o ID] [-t TYPE] [-k KIND]
```

Reads the records of the pipe, and prints only those that match the
flags it is given, unchanged. It can stand wherever records flow.
Right after `bcr extract`, it looks a breadcrumb up:

```
bcr extract | bcr filter --id ADR-0019
```

After `bcr audit`, it shows one breadcrumb's verdict, judged with the
whole set:

```
bcr extract | bcr verify | bcr audit | bcr filter --id ADR-0019
```

Flags:

- `-i ID`, `--id ID`: a `breadcrumb`, `link`, `claim`,
  `claim-verdict` or `verdict` record matches when its `ID` field
  equals `ID`. A `problem` record has no `ID` field, and never
  matches.
- `-o ID`, `--object ID`: a `link` record matches when its `OBJECT`
  field, the id the link points to, equals `ID`. No other kind of
  record matches.
- `-t TYPE`, `--type TYPE`: a `breadcrumb` record matches when its
  `TYPE` field equals `TYPE`. No other kind of record matches.
- `-k KIND`, `--kind KIND`: a record matches when its first field
  equals `KIND`, whether or not `bcr` knows that kind.

At least one flag is given. Each flag may be given more than once: a
record matches a flag when it matches at least one of its values, and
is printed when it matches every flag given. Fields are compared byte
for byte, letter case included, so `adr-0019` does not select
`ADR-0019`.

Operands: none. `bcr filter` reads no file.

Standard input: the records, as `bcr extract`, `bcr verify` and
`bcr audit` print them, read to the end. A line may end in `\n` or
`\r\n`. A record of a kind `bcr` does not know matches only `--kind`,
and fields after the ones printed today are not read.

Output: each record that matches, on standard output, as it was read,
byte for byte, with its line ending, in the order read. `bcr filter`
adds no record and changes none. It reads its whole input before it
writes. When no record matches, it prints nothing, and that is not an
error.

`bcr filter` judges nothing: two records with the same id both match,
and a claim or a verdict is not checked. So it comes after the
commands that judge, not before them. Before `bcr verify`, a link to a
breadcrumb filtered out is reported as pointing nowhere; before
`bcr audit`, a `problem` record that is dropped can no longer make its
breadcrumb `Undecided`.

A `problem` record is printed only with `--kind problem`: it has no
id, so the problems of a breadcrumb do not follow it. `--type` alone
prints `breadcrumb` records only; the links and claims of those
breadcrumbs are found by their ids, with a second `bcr filter`.

A link is written only in the breadcrumb it belongs to: `--id X`
selects the links X writes, and `--object X` the links that point to
X, wherever they are written. In what `bcr extract` prints, which
has no link from a breadcrumb to itself, `--id X --object X` selects
nothing, so the two directions are two runs over the same records:

```
bcr extract > records.tsv
bcr filter --kind link --id ADR-0019 < records.tsv
bcr filter --object ADR-0019 < records.tsv
```

A line of input that is not a record the pipe could print, an empty
line included, stops `bcr filter`: it writes a message starting
`bcr: `, which gives the line's number in the input, and prints no
record. A record of a kind the pipe prints needs the fields of its
kind, none of them empty, and a `LINE` that is a whole number from 1.

Exit status:

- 0: the input was read to its end, whether or not a record matched.
- 2: `bcr filter` was given an operand or no flag, its input could not
  be read, a line of its input was not a record, or its output could
  not be written.

Examples:

```
bcr extract | bcr filter --id ADR-0019
bcr extract | bcr filter --type ADR | wc -l
bcr extract | bcr filter -k link -i PBI-00001 | cut -f3,4
bcr extract | bcr filter --object ADR-0019 | cut -f2,3
bcr extract | bcr verify | bcr audit | bcr filter --kind verdict --id filter
bcr extract | bcr verify | bcr filter --kind problem | cut -f2-
```

### init

```
bcr init [-a FILE] [-l LAYOUT]
```

Sets up the repository whose root is the current directory for the
pipe. It writes no breadcrumb that `.breadcrumbs` names: which files
carry breadcrumbs, and what they are, is the repository's choice.
Each of three pieces is set up only when it is not there yet, found
by its name:

- `.breadcrumbs`, whose first line is `# bcr init, bcr VERSION`, the
  release of `bcr` that wrote it, and which holds only comments: what
  a pattern is, and example patterns, commented out. Until a pattern
  is written in it,
  `bcr extract` reads no file, and the pipe exits 0 with a page that
  says `No breadcrumbs.`.
- A `## Breadcrumbs` section for agents in the agents file, saying
  where breadcrumbs live, to run the pipe before committing, and never
  to edit a claim to turn a `Refuted` verdict green, then showing how
  to look breadcrumbs up with `bcr filter`: one by its id, every one
  of a type, what links to one, and one's verdict. It is added at the
  end of the file, after an empty line, when the file has no line
  `## Breadcrumbs`; a file that is not there is written with the
  section alone.
- `.github/workflows/breadcrumbs.yml`, a GitHub Actions workflow that,
  on each push and pull request, installs this release of `bcr` from
  its Linux archive, checked against the release's checksums, runs
  the pipe under `set -o pipefail`, and puts the page `bcr report`
  writes in the job's summary and in an artifact of the run.

With `--layout asdlc`, `bcr init` sets up the layout of ASDLC as
well, following ASDLC's conventions: ADRs in
`docs/adrs/ADR-NNN-slug.md`, specs in `specs/feature-name/spec.md`,
and PBIs in `tasks/PBI-NNN.md`. In place of the `.breadcrumbs` of
comments, it writes a `.breadcrumbs` that names those files and
excludes the templates:

```
# bcr init --layout asdlc, bcr VERSION
docs/adrs/*.md
!docs/adrs/TEMPLATE.md
specs/*/spec.md
tasks/*.md
!tasks/TEMPLATE.md
```

It writes `breadcrumb.rules`, the shape `bcr verify` checks:

```
# bcr init --layout asdlc, bcr VERSION
type	ADR
type	spec
type	PBI
link	ADR	constrained_by	ADR
link	ADR	supersedes	ADR
link	spec	constrained_by	ADR
link	PBI	changes	spec
claims	spec
```

And it sets up two pieces more:

- `docs/adrs/TEMPLATE.md`, `specs/TEMPLATE.md` and
  `tasks/TEMPLATE.md`, each with the sections of its kind of document
  and front matter that is a breadcrumb of its type, with the links it
  may have as comments. Each is written when it is not there.
- A `## ASDLC` section in the agents file, after the `## Breadcrumbs`
  section and added as it is: where each kind of document lives, how
  it is named, and where its template is. For the links, it points to
  `breadcrumb.rules`.

A layout's files work only together, so when `.breadcrumbs` or
`breadcrumb.rules` is there, `--layout` writes nothing and exits 2.
To set the layout up again, delete both and run it again.

Flags:

- `-a FILE`, `--agents-file FILE`: the agents file, a path from the
  root of the repository, such as `CLAUDE.md`. Without it,
  `AGENTS.md`. A path that starts with `/` or has a part that is `..`
  is a usage error.
- `-l LAYOUT`, `--layout LAYOUT`: a layout to set up as well. The one
  layout is `asdlc`. Any other is a usage error that names it.

Operands: none.

Standard input: not read.

Output: on standard output, the path of each file written or added
to, one per line: `.breadcrumbs`, then, with `--layout`,
`breadcrumb.rules` and the templates, then the agents file and the
workflow. On standard error, a message for
each piece that was already there, and is left as it is.

`bcr init` replaces no file, and changes no file but the agents file,
by adding its section at the end. It finds what to write before it
writes anything. The workflow installs the release of `bcr` that wrote
it, so `bcr init` needs a released `bcr`: one built with `go run`, or
with no release version, exits 2 and writes nothing.

On a CI other than GitHub Actions, the same check is:

```
set -o pipefail; bcr extract | bcr verify | bcr audit > /dev/null
```

Exit status:

- 0: every piece is in place, set up now or before.
- 2: `bcr init` was used wrongly, this `bcr` has no release version,
  `--layout` found a `.breadcrumbs` or a `breadcrumb.rules`, the
  agents file is not a regular file, or a file could not be read or
  written.

Examples:

```
bcr init
bcr init -a CLAUDE.md
bcr init --layout asdlc
```

## EXIT STATUS

Every command exits with (ADR-0006):

- 0: it succeeded and found nothing wrong.
- 1: it ran and found something wrong in the breadcrumbs.
- 2: it was used wrongly, or could not run.

`bcr report`, `bcr filter` and `bcr init` judge nothing, so they exit
only with 0 or 2.

## SEE ALSO

The method: `README.md`. How the repository is developed:
`CONTRIBUTING.md`.
