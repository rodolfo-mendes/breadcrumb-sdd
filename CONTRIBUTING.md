# Contributing to Breadcrumb SDD

This repository is developed with Breadcrumb SDD itself (TD-0001).
Every change leaves breadcrumbs in `breadcrumbs/`: small Markdown files
that record why the repository changed, in a form a reader can check
against the code. Read them before you change anything.

This guide summarizes the breadcrumbs. Where the two disagree, the
breadcrumbs win.

## Build and test

`bcr` is written in Go (TD-0008), using only the standard library
(TD-0010). You need the Go version named in `go.mod`.

```
go test ./...
go run ./cmd/bcr audit-report --html
```

The second command writes `audit-report.html` at the repository root:
a graph of the breadcrumbs, each colored by its verdict. Green is
Confirmed, red is Refuted, gray is Undecided.

## The breadcrumbs

| Prefix | Kind                | Records                                       | Decided by |
|--------|---------------------|-----------------------------------------------|------------|
| `IN`   | Intake              | an ask, in the words of whoever made it       | TD-0003    |
| `RQ`   | Requirement         | one thing the software must do                | TD-0004    |
| `TD`   | Technical Decision  | one decision about how the repository is built | TD-0001   |
| `TK`   | Task                | one change, as claims about the files it left | TD-0011    |

Each sits directly in `breadcrumbs/`, in a file named `XX-NNNN.md`
with the next free four-digit number, and starts with a `# ` title
(TD-0013). A breadcrumb names each breadcrumb it derives from on its
own line, right below its title (TD-0005):

```markdown
Parent: [IN-0001](IN-0001.md)
```

A Task lists its claims under a `## Claims` heading. The only kind of
claim says that a file contains a piece of text (TD-0012):

```markdown
- `cmd/bcr/main.go` contains `func run(`
```

## Making a change

1. **Start from an ask.** A new ask becomes an Intake, and what the
   software must do becomes Requirements. They state what the
   maintainers want, so a maintainer approves each one (TD-0006). Open
   an issue before proposing one.
2. **Record decisions.** When the change decides how the repository is
   built, write a Technical Decision with `## Decision`, `## Why` and
   `## What it beat` sections. Earlier ones show the form.
3. **Change the code.**
4. **Write a Task.** Name the Requirements and Technical Decisions it
   carries out as its parents, and state as claims what the change
   left in the files.
5. **Audit.** Run the tests and `bcr audit-report --html`. No
   breadcrumb should be red.
6. **Commit.**

## Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
(TD-0016), with three types:

- `feat`: the change adds to or changes what the software does.
- `fix`: the change corrects something that does not work as intended.
- `docs`: the change touches only prose, such as this guide, and leaves
  what the software does unchanged.

The description says what the change does, in the imperative:

```
feat: Audit Tasks and pass their verdicts up the breadcrumbs
```

The body says what changed, lists each breadcrumb the commit adds or
changes with a line on what it records, and ends with how to verify
the change:

```
- TD-0012: the contains claim, the only kind of claim.
- TK-0002: this change.

Verify with: go test ./... and go run ./cmd/bcr audit-report --html
```

## Agents

Coding agents learn about the breadcrumbs from `AGENTS.md` (TD-0002).
They follow this guide too: they may write Technical Decisions and
Tasks, and they draft Intakes and Requirements for a person to approve
(TD-0006).

## License

This repository is licensed under the MIT License (TD-0014). By
contributing, you agree that your contribution is licensed under it
too.
