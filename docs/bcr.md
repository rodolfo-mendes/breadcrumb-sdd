# bcr(1)

## NAME

bcr - audit the breadcrumbs of a repository

## SYNOPSIS

```
bcr audit-report --html [-o FILE]
bcr check
bcr extract [FILE...]
bcr verify
bcr audit
bcr verdict [ID|PATH|-]...
bcr links [-r] [ID|PATH|-]...
bcr new TYPE TITLE [PARENT|-]...
bcr spec
bcr init [-a FILE]
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
- A command that takes breadcrumbs as operands takes an id, such as
  `TK-0001`, a path relative to the repository root, such as
  `breadcrumbs/TK-0001.md`, or `-`, which reads one breadcrumb per line
  of standard input, named by the line's first tab-separated field.
  With no operands, it takes every breadcrumb, in order of path. An
  operand that names no breadcrumb is written to standard error, the
  others are still taken, and the command exits with 2
  ([TD-0026](../breadcrumbs/TD-0026.md),
  [RQ-0012](../breadcrumbs/RQ-0012.md),
  [RQ-0025](../breadcrumbs/RQ-0025.md)).

Which files are breadcrumbs:
[TD-0013](../breadcrumbs/TD-0013.md). The kinds of breadcrumb:
[TD-0001](../breadcrumbs/TD-0001.md),
[TD-0003](../breadcrumbs/TD-0003.md),
[TD-0004](../breadcrumbs/TD-0004.md) and
[TD-0011](../breadcrumbs/TD-0011.md). How a breadcrumb names its
parents: [TD-0005](../breadcrumbs/TD-0005.md).

The rules of the method are gathered in its specification,
[breadcrumb-sdd.md](breadcrumb-sdd.md), which `bcr spec` prints
([TD-0028](../breadcrumbs/TD-0028.md)).

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
link	ID	VERB	OBJECT	PATH	LINE
claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
```

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
    - follows ADR-0016
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

Operands: none. `bcr verify` reads no file: which files make up the
set is decided by `bcr extract`, from its operands or from
`.breadcrumbs`.

Standard input: the records, as `bcr extract` prints them, read to the
end. A line may end in `\n` or `\r\n`. A record of a kind other than
`breadcrumb` and `link` is not checked, a `problem` record included,
and neither are fields after
the ones `bcr extract` prints today; both are still copied. With no input, there is no problem.

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
  a line of its input was not a record, or its output could not be
  written.

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

### verdict

```
bcr verdict [ID|PATH|-]...
```

Prints the verdict of each breadcrumb
([RQ-0024](../breadcrumbs/RQ-0024.md)), or of the breadcrumbs it is
given ([RQ-0025](../breadcrumbs/RQ-0025.md)).

Flags: none.

Operands: breadcrumbs, as DESCRIPTION defines, in the order given
([RQ-0025](../breadcrumbs/RQ-0025.md)).

Standard input: read when an operand is `-`. The output of any
command here can feed it.

Output: one line on standard output for each breadcrumb: its id, a
tab, and its verdict, `Confirmed`, `Refuted` or `Undecided`:

```
TK-0001	Confirmed
```

How a verdict is reached: see audit-report.

Exit status ([RQ-0026](../breadcrumbs/RQ-0026.md)):

- 0: no breadcrumb printed is Refuted. Undecided ones do not change it.
- 1: a breadcrumb printed is Refuted.
- 2: `bcr` was used wrongly, an operand names no breadcrumb, or there
  is no `breadcrumbs/` directory in the current directory.

Examples:

```
bcr verdict | grep -w Refuted | cut -f1
bcr verdict breadcrumbs/TK-*.md
```

### links

```
bcr links [-r] [ID|PATH|-]...
```

Prints each link from a breadcrumb to its parent
([RQ-0028](../breadcrumbs/RQ-0028.md)).

Flags:

- `-r`, `--recursive`: also print the links of each parent reached,
  and of their parents, up to the breadcrumbs with none. Each link is
  printed once, the first time it is reached
  ([RQ-0030](../breadcrumbs/RQ-0030.md)).

Operands: breadcrumbs, as DESCRIPTION defines. Only the links they
have are printed, in the order given
([RQ-0029](../breadcrumbs/RQ-0029.md)).

Standard input: read when an operand is `-`.

Output: one line on standard output for each `Parent:` link: the id
of the breadcrumb that has it, a tab, and the id of the parent, in
order of the breadcrumb's `Parent:` lines. A link to no breadcrumb
prints the path it points to in place of the parent's id; `bcr check`
reports it as a problem. How a breadcrumb names its parents:
[TD-0005](../breadcrumbs/TD-0005.md).

Exit status ([RQ-0029](../breadcrumbs/RQ-0029.md)):

- 0: the links were printed.
- 2: `bcr` was used wrongly, an operand names no breadcrumb, or there
  is no `breadcrumbs/` directory in the current directory.

Examples: what a Refuted Task breaks:

```
bcr verdict | grep -w Refuted | bcr links -r - | cut -f2 | sort -u
```

### new

```
bcr new TYPE TITLE [PARENT|-]...
```

Creates a breadcrumb in `breadcrumbs/` and prints its path
([RQ-0031](../breadcrumbs/RQ-0031.md)).

Flags: none.

Operands:

- `TYPE`: `IN`, `RQ`, `TD` or `TK`, the prefix of the breadcrumb's
  kind ([TD-0027](../breadcrumbs/TD-0027.md)).
- `TITLE`: its title, on one line. Quote it in the shell.
- `PARENT`: each breadcrumb it derives from, as DESCRIPTION defines;
  with none, it has no parent ([RQ-0033](../breadcrumbs/RQ-0033.md)).

Standard input: read when a `PARENT` is `-`.

Output: the path of the new breadcrumb, relative to the repository
root, on standard output. Its number is one more than the highest of
its type ([TD-0027](../breadcrumbs/TD-0027.md)). It holds its title
line, a `Parent:` line for each parent, and, for a Technical Decision
or a Task, the empty sections to fill in
([RQ-0032](../breadcrumbs/RQ-0032.md)).

A breadcrumb made by `bcr new` is a draft. Who approves it before it
is committed: [TD-0002](../breadcrumbs/TD-0002.md) and
[TD-0006](../breadcrumbs/TD-0006.md).

Exit status ([RQ-0033](../breadcrumbs/RQ-0033.md)):

- 0: the breadcrumb was written.
- 2: `bcr` was used wrongly, the type or title is not valid, a parent
  names no breadcrumb, no number is left, the file could not be
  written, or there is no `breadcrumbs/` directory in the current
  directory. Nothing is written, and no file is ever replaced.

Examples:

```
bcr new TD "Use a cache" IN-0003
bcr verdict | grep -w Undecided | grep '^RQ' | bcr new TK "Carry them out" -
```

### spec

```
bcr spec
```

Prints the specification of Breadcrumb SDD that this `bcr` carries
out ([RQ-0035](../breadcrumbs/RQ-0035.md)). The specification is
built into the binary, so `bcr spec` reads no file and runs anywhere
([TD-0029](../breadcrumbs/TD-0029.md)).

Flags: none.

Operands: none.

Standard input: not read.

Output: the specification, as Markdown, on standard output. Its third
line is its version, `Version MAJOR.MINOR.PATCH`
([RQ-0036](../breadcrumbs/RQ-0036.md)); how the version changes:
[TD-0030](../breadcrumbs/TD-0030.md).

Exit status:

- 0: the specification was printed.
- 2: `bcr` was used wrongly, or standard output could not be written.

Example, the version of the specification this `bcr` carries out:

```
bcr spec | sed -n 3p
```

### init

```
bcr init [-a FILE]
```

Sets up Breadcrumb SDD in the repository whose root is the current
directory ([RQ-0042](../breadcrumbs/RQ-0042.md)). Each piece is set up
only when it is not there yet, found by its name
([TD-0034](../breadcrumbs/TD-0034.md)):

- `breadcrumbs/TD-0001.md`, a Technical Decision to adopt the version
  of the specification `bcr spec` prints, written only when
  `breadcrumbs/` holds no breadcrumb;
- a `## Breadcrumb SDD` section in the agents file, pointing agents to
  the breadcrumbs and to that version's page on the site
  ([RQ-0043](../breadcrumbs/RQ-0043.md));
- `.github/workflows/breadcrumbs.yml`, a GitHub Actions workflow that
  runs `bcr check` and `bcr verdict`, with this release of `bcr`, on
  each push and pull request ([RQ-0044](../breadcrumbs/RQ-0044.md),
  [TD-0035](../breadcrumbs/TD-0035.md)).

Flags:

- `-a FILE`, `--agents-file FILE`: the agents file, `AGENTS.md` or
  `CLAUDE.md`. Without it, `AGENTS.md`
  ([RQ-0046](../breadcrumbs/RQ-0046.md)).

Operands: none.

Standard input: not read.

Output: on standard output, the path of each file created or added
to, relative to the repository root, one per line. On standard
error, a note for each piece that was already there, and left as it
is.

`bcr init` replaces no file, and changes none but the agents file,
by adding its section at the end
([RQ-0045](../breadcrumbs/RQ-0045.md)). It needs a released `bcr`,
since the workflow installs the same release
([TD-0036](../breadcrumbs/TD-0036.md)).

On a CI other than GitHub Actions, the same check is:

```
bcr check && bcr verdict
```

Exit status ([RQ-0045](../breadcrumbs/RQ-0045.md)):

- 0: everything is set up, including when it already was.
- 2: `bcr` was used wrongly, has no version, or a piece cannot be set
  up. Nothing is written when this is found before writing.

Examples:

```
bcr init
bcr init -a CLAUDE.md
```

## EXIT STATUS

Every command exits with ([RQ-0010](../breadcrumbs/RQ-0010.md)):

- 0: it succeeded and found nothing wrong.
- 1: it ran and found something wrong in the breadcrumbs.
- 2: it was used wrongly, or could not run.

## SEE ALSO

The method: `README.md`. Its rules: the specification, printed by
`bcr spec`, and the breadcrumbs in `breadcrumbs/`. How the repository
is developed: `CONTRIBUTING.md`.
