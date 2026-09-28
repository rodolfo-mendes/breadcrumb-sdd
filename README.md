# Breadcrumb SDD

Breadcrumb SDD is a method for recording why a repository changed.

## The problem

Code says what a repository does, not why. The why lives in tickets,
chats, design documents and people's heads, and it drifts away from
the code without anyone noticing. A document that once described the
code keeps reading as true after the code has changed. Later, nobody
can tell whether the code still does what was asked of it, or which
decision a piece of code was meant to carry out.

## The method

Code stays the source of truth for what the repository does. Next to
it, in `breadcrumbs/`, small Markdown files called breadcrumbs record
intent, each linked to the breadcrumbs it derives from:

- an **Intake** (`IN-NNNN.md`) records an ask, in the words of whoever
  made it;
- a **Requirement** (`RQ-NNNN.md`) states one thing the software must
  do;
- a **Technical Decision** (`TD-NNNN.md`) records one decision about
  how the repository is built;
- a **Task** (`TK-NNNN.md`) records one change, as claims about the
  files it left, such as ``- `cmd/bcr/main.go` contains `func run(` ``.

The toolkit, `bcr`, checks each Task's claims against the files and
passes the verdict up the links. A Task whose claims all hold is
Confirmed; one with a claim that fails is Refuted, and so is every
breadcrumb above it. Drift between intent and code shows up as a
broken chain, which a reader can follow from the ask down to the
line of code that no longer matches.

The method makes drift visible after the fact; it does not prevent
it.

This repository is developed with Breadcrumb SDD itself: its own
breadcrumbs are in [`breadcrumbs/`](breadcrumbs/).

The rules of the method are in its specification,
[`docs/breadcrumb-sdd.md`](docs/breadcrumb-sdd.md), also printed by
`bcr spec` and shipped in each release.

## Running bcr

`bcr` is a single Go binary with no dependencies. Install it with Go
1.22 or later:

```
go install github.com/rodolfo-mendes/breadcrumb-sdd/cmd/bcr@latest
```

Then, from the root of a repository that has a `breadcrumbs/`
directory:

```
bcr audit-report --html
```

It writes `audit-report.html` in the repository root: a graph of the
breadcrumbs, each drawn as a box with its properties and an arrow to
each of its parents. Open it in a browser. Boxes are green when
Confirmed, red when Refuted and gray when Undecided.

Every command, flag, output and exit code of `bcr` is documented in
[`docs/bcr.md`](docs/bcr.md), also shipped as the man page `bcr.1`:
`man ./bcr.1` in an unpacked release, or `man ./docs/bcr.1` in a
clone.

To run it from a clone of this repository instead:

```
go run ./cmd/bcr audit-report --html
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE).
