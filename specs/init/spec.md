---
breadcrumb:
  id: init
  type: spec
  links:
    - follows ADR-0006
    - follows ADR-0007
    - follows ADR-0008
    - follows ADR-0014
    - follows ADR-0020
    - follows ADR-0024
  claims:
    - 'docs/bcr.md has-line ### init'
---
# Feature: init

`bcr init` sets a repository up for the pipe: a `.breadcrumbs` file, a
section for agents, and a GitHub Actions workflow that runs the pipe
on each change.

## Blueprint

### Context

A repository adopts Breadcrumb by naming the files that carry
breadcrumbs, writing breadcrumbs in them, and running
`bcr extract | bcr verify | bcr audit | bcr report` where changes are
checked. `bcr init` sets up what can be set up before the first
breadcrumb, and nothing more:

```
bcr init
```

It writes no breadcrumb. Which documents a repository keeps, and which
of them carry breadcrumbs, are its owners' choice; a first ADR or spec
written by `bcr init` would make that choice for them.

It replaces nothing. A piece that is already there is left as it is,
so `bcr init` can be run again, or in a repository that set up part of
it by hand.

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command, `cmd/bcr` | Reads the flag, finds which pieces are missing, writes them, sets the exit status | Infrastructure |

- `bcr init` reads no breadcrumb and no record, and runs no command.
  It has no rule about breadcrumbs, so nothing of it is in the core.
- The text of each piece is held in `bcr` itself.

### Interface

```
bcr init [-a FILE]
```

Run it from the root of the repository. It sets up three pieces, each
found by its name:

- **`.breadcrumbs`.** Written when there is no `.breadcrumbs`. It
  holds only comments: what a pattern is, and the patterns of this
  repository as examples, commented out. Until its owner writes a
  pattern, `bcr extract` reads no file, and the pipe exits 0 with a
  page that says `No breadcrumbs.`.
- **The agents section.** A `## Breadcrumbs` section for agents, in
  the agents file: `AGENTS.md`, or the file `-a` names. When the file
  is not there, it is written with the section alone. When it is
  there and has no line equal to `## Breadcrumbs`, the section is
  added at its end, after an empty line; nothing else in the file
  changes. The section says that breadcrumbs live in front matter,
  under the `breadcrumb` key, of the files `.breadcrumbs` names; to
  run the pipe before committing; and never to edit a claim to turn a
  Refuted verdict green. It then shows, as commands, how to look
  breadcrumbs up with `bcr filter` rather than search the files: one
  breadcrumb by its id, every breadcrumb of a type, what links to a
  breadcrumb, and a breadcrumb's verdict, read after `bcr audit`
  (`specs/filter/spec.md`).
- **The workflow.** `.github/workflows/breadcrumbs.yml`, written when
  there is no file of that name. On each push and pull request, it
  installs the release of `bcr` that wrote it, from its Linux x86-64
  archive, checked against the release's checksums; runs the pipe
  under `set -o pipefail`; puts the page in the job's summary; and
  keeps it as an artifact of the run, whether the job fails or not.
  Each action it uses is pinned to a commit.

Flags:

- `-a FILE`, `--agents-file FILE`: the agents file, a path from the
  root of the repository. Without it, `AGENTS.md`. A path that is
  empty, starts with `/`, or has a part that is `..` is a usage error.

Operands: none.

Standard input: not read.

Output: on standard output, the path of each file written or added
to, one per line, in the order `.breadcrumbs`, the agents file, the
workflow. On standard error, a message starting `bcr: ` for each
piece that was already there.

Exit status:

- 0: every piece is in place, whether set up now or before.
- 2: `bcr init` was used wrongly, this `bcr` has no version, the
  agents file or a piece's path is not a regular file where one is
  needed, or a file could not be read or written. Its message starts
  `bcr: ` (ADR-0006).

### Constraints

- **The workflow runs the `bcr` that wrote it.** It installs the
  release whose version this `bcr` was built with, so a later release
  cannot change a repository's verdicts with nothing in its files to
  show why. ADR-0020 rules that out for a kind of claim, not for the
  rest of `bcr`. A `bcr` built with no release version, such as with
  `go run`, cannot write the workflow, so `bcr init` exits 2 and
  writes nothing.
- **Nothing is written when a problem is found before writing.** What
  is there is read, and what is missing is decided, before the first
  file is written. A file that cannot be written then stops
  `bcr init`; the files written before it stay, and are printed.
- **No file is replaced.** A file is written only when it is not
  there, and the agents file is only added to.
- **A section written before keeps its words.** `bcr init` finds the
  section by its heading and replaces nothing, so a repository whose
  section an earlier `bcr` wrote does not get the lookup commands;
  they are added by hand.
- The workflow is for GitHub Actions. On another CI, the same check is
  the pipe: `set -o pipefail; bcr extract | bcr verify | bcr audit`.

## Contract

### Definition of Done

- [ ] Each Scenario below has a test.
- [x] `docs/bcr.md` describes `bcr init` under `### init`, and its
      flag; `docs/bcr.1` is generated again.
- [x] In a new directory, `bcr init` with a version, then the pipe,
      exits 0 and the page says `No breadcrumbs.`.
- [x] The workflow `bcr init` writes is valid YAML, and every action
      it uses is pinned to a commit.
- [ ] Each lookup command the agents section shows runs, and prints
      the records it names, in a repository with breadcrumbs.

### Regression Guardrails

- `docs/bcr.md` documents `bcr init`.

- `bcr init` replaces no file.

- `bcr init` writes no breadcrumb.

### Scenarios

```gherkin
Scenario: A repository with none of the pieces
  Given a directory with no .breadcrumbs, no AGENTS.md and no workflow
  When I run "bcr init" with a bcr released as 0.8.0
  Then .breadcrumbs, AGENTS.md and .github/workflows/breadcrumbs.yml
    are written
  And standard output is their three paths, in that order
  And nothing is printed to standard error
  And the exit status is 0

Scenario: The .breadcrumbs it writes names no file
  Given the .breadcrumbs "bcr init" wrote
  Then each of its lines is empty or starts with "#"

Scenario: The workflow runs the bcr that wrote it
  Given the workflow "bcr init" wrote with a bcr released as 0.8.0
  Then it downloads bcr_0.8.0_linux_amd64.tar.gz and
    bcr_0.8.0_checksums.txt from the release v0.8.0
  And it checks the archive against the checksums
  And it runs "bcr extract | bcr verify | bcr audit | bcr report"
    under "set -o pipefail"
  And every action it uses is pinned to a commit

Scenario: The agents section shows how to look breadcrumbs up
  Given the agents section "bcr init" writes
  Then it shows "bcr extract | bcr filter --id ID",
    "bcr extract | bcr filter --type TYPE",
    "bcr extract | bcr filter --object ID" and
    "bcr extract | bcr verify | bcr audit | bcr filter --kind verdict --id ID"
  And each of them, run with an id and a type of a repository's
    breadcrumbs, exits 0 and prints their records

Scenario: An agents file without the section
  Given an AGENTS.md that holds "# Agents" and no "## Breadcrumbs"
  When I run "bcr init"
  Then AGENTS.md starts with its old content, unchanged
  And the section follows it, after an empty line
  And AGENTS.md is printed to standard output

Scenario: An agents file with the section
  Given an AGENTS.md with a line "## Breadcrumbs"
  When I run "bcr init"
  Then AGENTS.md is unchanged
  And standard error says it was left as it is
  And the other pieces are written

Scenario: Another agents file
  Given no CLAUDE.md
  When I run "bcr init -a CLAUDE.md"
  Then CLAUDE.md is written with the section
  And no AGENTS.md is written

Scenario: Every piece already there
  Given a directory where "bcr init" already ran
  When I run "bcr init" again
  Then no file changes
  And nothing is printed to standard output
  And standard error has a message for each piece
  And the exit status is 0

Scenario: A piece set up by hand
  Given a .breadcrumbs that names docs/*.md
  When I run "bcr init"
  Then .breadcrumbs is unchanged
  And AGENTS.md and the workflow are written

Scenario: A bcr with no version
  Given a bcr built with no release version
  When I run "bcr init" in a directory with none of the pieces
  Then no file is written
  And standard error says the workflow needs a released bcr
  And the exit status is 2

Scenario: An agents file that is not a file
  Given AGENTS.md is a directory
  When I run "bcr init"
  Then no file is written
  And the exit status is 2

Scenario: A path out of the repository
  When I run "bcr init -a ../AGENTS.md"
  Then no file is written
  And standard error has the usage line
  And the exit status is 2

Scenario: An operand
  When I run "bcr init x"
  Then no file is written
  And the exit status is 2
```
