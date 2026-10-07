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
| Verify and audit breadcrumbs | `set -o pipefail; go run ./cmd/bcr extract \| go run ./cmd/bcr verify \| go run ./cmd/bcr audit > /dev/null` | `specs/extract/spec.md`, `specs/verify/spec.md`, `specs/audit/spec.md` |
| Report the breadcrumbs | `go run ./cmd/bcr extract \| go run ./cmd/bcr verify \| go run ./cmd/bcr audit \| go run ./cmd/bcr report` | `specs/report/spec.md` |
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
- Before adding a Go module outside the standard library. Each
  module `bcr` requires directly is introduced by its own ADR
  ([ADR-0005](docs/adrs/ADR-0005-dependencies-enter-through-adr.md)).

**ALWAYS**
- Read the feature's spec and the PBI before changing its code.
- Change a feature's spec in the same commit as the behavior it
  describes.
- Give a new ADR, PBI or spec a breadcrumb in its front matter, as
  [ADR-0002](docs/adrs/ADR-0002-breadcrumb-metadata.md) describes.

## Context Map

```yaml
ARCHITECTURE.md: describes the high-level design of Breadcrumb and the `bcr` command
docs/bcr.md: the user documentation of bcr; docs/bcr.1 is generated from it
internal/crumb: the core of the front matter breadcrumbs; knows no file format (ADR-0013)
internal/frontmatter: reads a breadcrumb from a file's YAML front matter for internal/crumb
internal/fileset: reads .breadcrumbs and finds the files it names; infrastructure (ADR-0014)
internal/records: writes and reads the records the commands of the pipe print; the only package that knows their format (ADR-0011)
internal/target: reads the file a claim is about from the working tree; infrastructure (ADR-0023)
internal/page: writes the page bcr report prints; the only package that knows its Markdown and Mermaid; infrastructure
.breadcrumbs: names by pattern the files bcr extract reads when given none
specs/: feature specs, one directory per feature; each command of bcr is a feature
tasks/: PBIs, closed after merge and kept
docs/adrs/: decisions, from ADR-0001 on
```
