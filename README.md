# Breadcrumb SDD

Breadcrumb SDD makes drift between a repository's intent and its code
visible.

## The problem

Code says what a repository does, not why. The why lives in tickets,
chats, design documents and people's heads, and it drifts away from
the code without anyone noticing. A document that once described the
code keeps reading as true after the code has changed. Later, nobody
can tell whether the code still does what was asked of it, or which
decision a piece of code was meant to carry out.

## The method

Code stays the source of truth for what the repository does. The
documents that hold its intent, such as decisions, specs and tasks,
each carry a breadcrumb: a few lines of YAML front matter that say
what the document is, which documents it links to, and what it claims
about the repository's files.

```yaml
---
breadcrumb:
  id: verify
  type: spec
  links:
    - constrained_by ADR-0016
  claims:
    - 'docs/bcr.md has-line ### verify'
---
```

- `id` names the breadcrumb, and `type` says what kind of document
  carries it.
- Each entry of `links` is a verb and the id of another breadcrumb.
- Each entry of `claims` is a file, a kind of claim and its argument.
  The only kind is `has-line`: the file has a line equal to the text,
  once the spaces and tabs at the line's start and end are removed.

The toolkit, `bcr`, reads the breadcrumbs, checks that their ids are
unique and their links point to a breadcrumb, and checks each claim
against the files. A repository may also declare its shape in
`breadcrumb.rules`: the types its breadcrumbs may have, the links
between them, and which types carry claims. `bcr` then reports any
breadcrumb that goes beyond it. A breadcrumb whose claims all hold is Confirmed,
and one with a claim that fails is Refuted. One with no claims, or
with a problem, is Undecided: nothing about it was checked. A Refuted
claim is drift: the document says one thing, and the file another.

The method makes drift visible after the fact; it does not prevent
it. A repository records as much or as little as its owners choose,
and a claim can be checked by hand, against the files, without `bcr`.

`bcr`'s behavior defines the method. Every command, record, output
and exit code is documented in [`docs/bcr.md`](docs/bcr.md), also
shipped as the man page `bcr.1`: `man ./bcr.1` in an unpacked release,
or `man ./docs/bcr.1` in a clone.

This repository is developed with Breadcrumb SDD itself: its
decisions in [`docs/adrs/`](docs/adrs/), its specs in
[`specs/`](specs/) and its tasks in [`tasks/`](tasks/) each carry a
breadcrumb, and every push audits them.

## Running bcr

`bcr` is a single Go binary. Install it with Go 1.22 or later:

```
go install github.com/rodolfo-mendes/breadcrumb-sdd/cmd/bcr@latest
```

In the root of a repository, set it up:

```
bcr init
```

It writes a `.breadcrumbs` file, a section for agents in `AGENTS.md`
(or the file `-a` names), and a GitHub Actions workflow that runs the
pipe on each change with the same release of `bcr`, so it needs a
released `bcr`, not one run with `go run`. It replaces nothing, and
writes no breadcrumb that `.breadcrumbs` names. Next, list the files
that carry breadcrumbs in `.breadcrumbs`, one pattern per line, such
as:

```
docs/adrs/*.md
specs/*/spec.md
tasks/*.md
```

A repository that follows
[ASDLC](https://asdlc.io) can be set up in one step instead:

```
bcr init --layout asdlc
```

As well as the pieces above, it writes a `.breadcrumbs` that names
ASDLC's ADRs, specs and PBIs, a `breadcrumb.rules` with ASDLC's
shape, a template for each of the three, and a section in the agents
file that says where each goes. It refuses when `.breadcrumbs` or
`breadcrumb.rules` is already there.

Then run four of the commands of `bcr` as a pipe:

```
bcr extract | bcr verify | bcr audit | bcr report > report.md
```

- `bcr extract` reads each breadcrumb and prints it as records.
- `bcr verify` checks the whole set: unique ids, links that point to
  a breadcrumb, and, when there is a `breadcrumb.rules`, the shape it
  declares.
- `bcr audit` checks each claim against the files, and gives each
  claim and each breadcrumb a verdict.
- `bcr report` writes one Markdown page: every breadcrumb by verdict,
  every problem, and a Mermaid diagram for each spec. GitHub draws it
  wherever it shows Markdown.

Each stage passes every record on, so a stage's output can also be
read with `grep`, `cut` or `awk`. `bcr filter` selects records by id,
type or kind, or links by the id they point to, to look up one
breadcrumb, what links to it, or its verdict:

```
bcr extract | bcr filter --id ADR-0019
bcr extract | bcr filter --object ADR-0019
bcr extract | bcr verify | bcr audit | bcr filter --kind verdict --id ADR-0019
```

To check a repository in a script or
in CI, keep only the problems and the Refuted claims, and let a
Refuted claim fail the pipe:

```
set -o pipefail
bcr extract | bcr verify | bcr audit > /dev/null
```

To run it from a clone of this repository instead, replace `bcr` with
`go run ./cmd/bcr`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE).
