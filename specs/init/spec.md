---
breadcrumb:
  id: init
  type: spec
  links:
    - constrained_by ADR-0006
    - constrained_by ADR-0007
    - constrained_by ADR-0008
    - constrained_by ADR-0014
    - constrained_by ADR-0020
    - constrained_by ADR-0024
    - constrained_by ADR-0025
    - constrained_by ADR-0027
  claims:
    - 'docs/bcr.md has-line ### init'
---
# Feature: init

`bcr init` sets a repository up for the pipe: a `.breadcrumbs` file, a
section for agents, and a GitHub Actions workflow that runs the pipe
on each change. With `--layout asdlc`, it also sets up ASDLC's shape:
its `breadcrumb.rules`, a `.breadcrumbs` that names its files, three
templates, and a section that tells agents where each file goes.

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

It writes no breadcrumb that `.breadcrumbs` names. Which documents a
repository keeps, and which of them carry breadcrumbs, are its owners'
choice; a first ADR or spec written by `bcr init` would make that
choice for them.

It replaces nothing. A piece that is already there is left as it is,
so `bcr init` can be run again, or in a repository that set up part of
it by hand.

A layout is the one choice `bcr init` makes for a repository, and only
when asked (ADR-0027). `--layout asdlc` writes ASDLC's names into the
repository: its types and links in `breadcrumb.rules` (ADR-0025), its
directories and templates, and where each file goes. It follows
ASDLC's own conventions, not this repository's. A layout's pieces work
only together, so it refuses a `.breadcrumbs` or a `breadcrumb.rules`
that is already there rather than leave it:

```
bcr init --layout asdlc
```

### Architecture

| Part | Does | Kind (ADR-0007) |
|---|---|---|
| Command, `cmd/bcr` | Reads the flag, finds which pieces are missing, writes them, sets the exit status | Infrastructure |

- `bcr init` reads no breadcrumb and no record, and runs no command.
  It has no rule about breadcrumbs, so nothing of it is in the core.
- The text of each piece is held in `bcr` itself.

### Interface

```
bcr init [-a FILE] [-l LAYOUT]
```

Run it from the root of the repository. It sets up three pieces, each
found by its name:

- **`.breadcrumbs`.** Written when there is no `.breadcrumbs`. Its
  first line is `# bcr init, bcr VERSION`, where `VERSION` is the
  release of `bcr` that wrote it. It holds only comments: what a
  pattern is, and the patterns of this repository as examples,
  commented out. Until its owner writes a
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

With `--layout asdlc`, it sets up the ASDLC layout (ADR-0027) instead
of the `.breadcrumbs` above, and five pieces more:

- **`.breadcrumbs`.** Written as:

  ```
  # bcr init --layout asdlc, bcr VERSION
  docs/adrs/*.md
  !docs/adrs/TEMPLATE.md
  specs/*/spec.md
  tasks/*.md
  !tasks/TEMPLATE.md
  ```

- **`breadcrumb.rules`.** Written as:

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

- **The templates.** `docs/adrs/TEMPLATE.md`, `specs/TEMPLATE.md` and
  `tasks/TEMPLATE.md`, each written when it is not there, with the
  sections of ASDLC's ADR, spec and PBI. Each starts with front matter
  that is a breadcrumb of its type: the id `ADR-NNN`, `feature-name`
  or `PBI-NNN`, `links: []`, and, as comments under it, the links
  `breadcrumb.rules` allows from that type. The spec template also
  has `claims: []`, with the form of a claim as a comment.
  `.breadcrumbs` excludes the ADR and PBI templates, and the spec
  template matches none of its patterns, so no template is read as a
  breadcrumb of the repository.
- **The ASDLC section.** A `## ASDLC` section in the agents file,
  found by its heading and added as the `## Breadcrumbs` section is:
  where ADRs, specs and PBIs live and how they are named, their
  templates, that a spec's id is the name of its directory, that ids
  are three digits, and that only specs carry claims. For the links
  each type may have, it points to `breadcrumb.rules` and does not
  copy them. When the agents file is written, it holds the
  `## Breadcrumbs` section, then this one.

When `.breadcrumbs` or `breadcrumb.rules` is there, `--layout` writes
nothing: it prints a message for each, which says to delete both to
set the layout up again, and exits 2.

Flags:

- `-a FILE`, `--agents-file FILE`: the agents file, a path from the
  root of the repository. Without it, `AGENTS.md`. A path that is
  empty, starts with `/`, or has a part that is `..` is a usage error.
- `-l LAYOUT`, `--layout LAYOUT`: the layout to set up. The only one
  is `asdlc`. Any other is a usage error, whose message names the
  layouts `bcr` knows.

Operands: none.

Standard input: not read.

Output: on standard output, the path of each file written or added
to, one per line, in the order `.breadcrumbs`, `breadcrumb.rules`,
`docs/adrs/TEMPLATE.md`, `specs/TEMPLATE.md`, `tasks/TEMPLATE.md`, the
agents file, the workflow. On standard error, a message starting
`bcr: ` for each piece, or section, that was already there.

Exit status:

- 0: every piece is in place, whether set up now or before.
- 2: `bcr init` was used wrongly, this `bcr` has no version, the
  agents file or a piece's path is not a regular file where one is
  needed, `--layout` found `.breadcrumbs` or `breadcrumb.rules`, or a
  file could not be read or written. Its message starts `bcr: `
  (ADR-0006).

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
- **A layout is set up whole.** Without `--layout`, each piece is
  independent of the others, so one that is there is left. With it,
  `.breadcrumbs` and `breadcrumb.rules` must agree, and a
  `.breadcrumbs` of comments left by `bcr init` would make the pipe
  read nothing with no sign. So both must be absent: a layout is set
  up again by deleting both first.
- **The first line names what wrote the file.** `# bcr init, bcr
  VERSION` or `# bcr init --layout NAME, bcr VERSION`, exactly, as
  the first line of each `.breadcrumbs` and `breadcrumb.rules`
  `bcr init` writes. `bcr` reads it as a comment; it tells a person,
  or a later tool, which release and layout set the repository up.
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
- [ ] `docs/bcr.md` describes `--layout`, and names it in SYNOPSIS;
      `docs/bcr.1` is generated again.
- [x] In a new directory, `bcr init` with a version, then the pipe,
      exits 0 and the page says `No breadcrumbs.`.
- [x] The workflow `bcr init` writes is valid YAML, and every action
      it uses is pinned to a commit.
- [ ] In a new directory, `bcr init --layout asdlc` with a version,
      then the pipe, exits 0.
- [ ] In that directory, an ADR, a spec and a PBI written from the
      templates, linked as `breadcrumb.rules` allows, pass the pipe.
- [x] Each lookup command the agents section shows runs, and prints
      the records it names, in a repository with breadcrumbs.

### Regression Guardrails

- `docs/bcr.md` documents `bcr init`.

- `bcr init` replaces no file.

- `bcr init` writes no breadcrumb that `.breadcrumbs` names.

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
  Given the .breadcrumbs "bcr init" wrote with a bcr released as 0.8.0
  Then its first line is "# bcr init, bcr 0.8.0"
  And each of its lines is empty or starts with "#"

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

Scenario: The ASDLC layout
  Given a directory with none of the pieces
  When I run "bcr init --layout asdlc" with a bcr released as 0.9.0
  Then standard output is
    """
    .breadcrumbs
    breadcrumb.rules
    docs/adrs/TEMPLATE.md
    specs/TEMPLATE.md
    tasks/TEMPLATE.md
    AGENTS.md
    .github/workflows/breadcrumbs.yml
    """
  And nothing is printed to standard error
  And the exit status is 0

Scenario: The layout's breadcrumb.rules and .breadcrumbs
  Given the files "bcr init --layout asdlc" wrote with a bcr released
    as 0.9.0
  Then breadcrumb.rules and .breadcrumbs are as this spec shows them
  And the first line of each is "# bcr init --layout asdlc, bcr 0.9.0"

Scenario: The templates are not the repository's breadcrumbs
  Given a directory where "bcr init --layout asdlc" ran
  When I run "bcr extract" with no operand
  Then nothing is printed
  When I run "bcr extract" with each template as its operand
  Then it prints one breadcrumb record of the template's type, and no
    link record
  And the exit status is 0

Scenario: Documents written from the templates
  Given a directory where "bcr init --layout asdlc" ran
  And docs/adrs/ADR-001-first.md, specs/first/spec.md and
    tasks/PBI-001.md, each a copy of its template with its id filled
    in and the links of its comments written: the spec constrained by
    ADR-001, the PBI changing first
  When I run "bcr extract | bcr verify | bcr audit"
  Then nothing is printed to standard error
  And the exit status of each is 0

Scenario: The agents file with the layout
  Given a directory with no AGENTS.md
  When I run "bcr init --layout asdlc"
  Then AGENTS.md holds the Breadcrumbs section, then the ASDLC section
  And the ASDLC section names breadcrumb.rules and each template

Scenario: An agents file with the Breadcrumbs section only
  Given a directory where "bcr init" ran, then .breadcrumbs was deleted
  When I run "bcr init --layout asdlc"
  Then the ASDLC section is added at the end of AGENTS.md
  And its Breadcrumbs section is unchanged
  And standard error says the Breadcrumbs section was left as it is

Scenario: A layout over a .breadcrumbs or a breadcrumb.rules
  Given a directory where "bcr init" ran
  When I run "bcr init --layout asdlc"
  Then no file is written
  And standard error names .breadcrumbs and says to delete
    .breadcrumbs and breadcrumb.rules
  And the exit status is 2
  Given instead a directory with a breadcrumb.rules and no .breadcrumbs
  When I run "bcr init --layout asdlc"
  Then no file is written
  And the exit status is 2

Scenario: A template already there
  Given a directory with a specs/TEMPLATE.md and none of the other
    pieces
  When I run "bcr init --layout asdlc"
  Then specs/TEMPLATE.md is unchanged
  And standard error says it was left as it is
  And the other pieces are written

Scenario: A layout bcr does not know
  When I run "bcr init --layout spec-kit"
  Then no file is written
  And standard error names the layout asdlc, then the usage line
  And the exit status is 2
```
