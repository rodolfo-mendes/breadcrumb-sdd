# AGENTS.md

> **Project:** Breadcrumb, a layer over a repository's artifacts that
> records how they link and what they claim about the code, and checks
> the claims against the files. `bcr` is its toolkit.
> **Core constraints:** Code is the source of truth. Breadcrumb makes
> drift between intent and code visible after the fact; it does not
> prevent it. `bcr`'s behavior defines Breadcrumb.

This repository is developed with ASDLC
([ADR-0001](docs/adrs/ADR-0001-adopt-asdlc.md)).

## Toolchain

| Action | Command | Authority |
|---|---|---|
| Test | `go test ./...` | Go version in `go.mod` |
| Check breadcrumbs | `go run ./cmd/bcr check` | `docs/bcr.md` |
| Audit claims | `go run ./cmd/bcr verdict` | `docs/bcr.md` |
| Report | `go run ./cmd/bcr audit-report --html` | `docs/bcr.md` |
| Man page | `go test ./internal/manpage -update` | Run after any change to `docs/bcr.md` |

## Judgment Boundaries

**NEVER**
- Edit a claim to turn a Refuted verdict green. Fix the code, or
  propose the change to the spec.
- Edit the decision of an Accepted ADR. Write a new ADR that
  supersedes it.
- Delete a PBI; it stays in `tasks/`.

**ASK**
- Before setting an ADR to Accepted.
- Before changing a spec's Contract beyond what the PBI asks for.
- Before changing the method specification.
- Before adding a dependency outside Go's standard library.
- Before deleting a file in `breadcrumbs/`.

**ALWAYS**
- Read the feature's spec and the PBI before changing its code.
- Change a feature's spec in the same commit as the behavior it
  describes.
- Give a new ADR, PBI or spec a breadcrumb in its front matter, as
  [ADR-0002](docs/adrs/ADR-0002-breadcrumb-metadata.md) describes.
- When `docs/breadcrumb-sdd.md` and `bcr` disagree, report it as
  drift; `bcr`'s behavior holds until one of them is fixed.

## Context Map

```yaml
docs/breadcrumb-sdd.md: describes Breadcrumb as bcr carries it out; not a feature spec
docs/bcr.md: the contract of bcr; docs/bcr.1 is generated from it
internal/breadcrumb: the model of Breadcrumb in bcr
specs/: feature specs, one directory per feature
tasks/: PBIs, closed after merge and kept
docs/adrs/: decisions, from ADR-0001 on
breadcrumbs/: intakes, requirements, decisions and tasks from before ADR-0001; still audited by bcr
```
