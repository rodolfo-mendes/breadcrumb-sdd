# Contributing to Breadcrumb SDD

This repository is developed with ASDLC
([ADR-0001](docs/adrs/ADR-0001-adopt-asdlc.md)).
[`specs/asdlc/spec.md`](specs/asdlc/spec.md) says how a change is
made: which artifacts it uses, where their files live, and where the
repository goes beyond ASDLC. Where this guide and the spec disagree,
the spec wins.

## Build and test

`bcr` is written in Go, as one static binary
([ADR-0004](docs/adrs/ADR-0004-one-go-binary.md)). It may depend on
Go modules outside the standard library, each introduced by an ADR
that names it and gives the reason
([ADR-0005](docs/adrs/ADR-0005-dependencies-enter-through-adr.md)).
You need the Go version named in `go.mod`.

```
go test ./...
go run ./cmd/bcr audit-report --html
```

The second command writes `audit-report.html` at the repository root:
a graph of the breadcrumbs, each colored by its verdict. Green is
Confirmed, red is Refuted, gray is Undecided.

## Artifacts

| Artifact | Files | Records |
|---|---|---|
| Spec | `specs/<feature>/spec.md` | what a feature does, and the claims that check it |
| PBI | `tasks/PBI-NNNNN.md` | one change, and how to verify it |
| ADR | `docs/adrs/ADR-NNNN-<slug>.md` | one decision about how the repository is built |

`breadcrumbs/` holds the Intakes, Requirements, Technical Decisions
and Tasks written before ADR-0001. `bcr` still audits them.

## Making a change

1. **Write a PBI.** Name the spec it changes in its Context.
2. **Record decisions.** When the change decides how the repository is
   built, write an ADR. A maintainer accepts it. An Accepted ADR is not
   edited; a new ADR supersedes it.
3. **Change the code and its spec together.** The spec describes the
   new state in the same commit. When the change touches a command,
   flag, output or exit code of `bcr`, update its contract in
   `docs/bcr.md` and write the man page again with
   `go test ./internal/manpage -update`.
4. **Check.** Each item of the Definition of Done in
   `specs/asdlc/spec.md` holds.
5. **Commit.** The PBI reaches the main branch in the same merge as its
   change.

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

Coding agents learn how the repository is built from `AGENTS.md`.

## License

This repository is licensed under the MIT License (TD-0014). By
contributing, you agree that your contribution is licensed under it
too.
