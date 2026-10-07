---
breadcrumb:
  id: asdlc
  type: spec
  links:
    - follows ADR-0001
    - follows ADR-0002
  claims:
    - 'AGENTS.md has-line This repository is developed with ASDLC'
    - 'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'
---
# Feature: ASDLC

How this repository is built with ASDLC: which of its patterns it
uses, where their files live, and where it goes beyond ASDLC.

## Blueprint

### Context

[ADR-0001](../../docs/adrs/ADR-0001-adopt-asdlc.md) gives the reasons
for adopting ASDLC.

This spec covers how the repository is built, not what Breadcrumb
does. `bcr`'s behavior defines Breadcrumb; it is not a feature spec.

### Architecture

| Pattern | Files | Follows |
|---|---|---|
| AGENTS.md | `AGENTS.md` | ASDLC's AGENTS.md Specification |
| The Spec | `specs/<feature>/spec.md`, the directory in kebab-case | Living Specs |
| The PBI | `tasks/PBI-NNNNN.md`, one file per PBI | PBI Authoring |
| The ADR | `docs/adrs/ADR-NNNN-<slug>.md` | The ADR |

- `NNNN` and `NNNNN` are the next free number of their kind, padded
  with zeros.
- Each command of `bcr` is a feature: its spec is
  `specs/<command>/spec.md`, and its directory has the command's name
  ([ADR-0008](../../docs/adrs/ADR-0008-user-documentation-in-docs-bcr-md.md)).
- `ARCHITECTURE.md` is at the root of the repository, beside
  `AGENTS.md`. It describes how the parts of `bcr` fit together and
  links the ADRs and specs that decide them; it decides nothing
  itself.
- A PBI whose change is about how the repository is built names this
  spec in its Context.

### Constraints

Where this repository goes beyond ASDLC, it is listed here.

- **This spec.** It describes how the repository is built, not a
  feature of Breadcrumb; ASDLC's specs describe features.
- **Claims.** An item of a spec's Contract may carry claims: statements
  about the repository's files that a tool can check. They come from
  Breadcrumb, not from ASDLC. Each is
  written in the front matter of its spec, under the `claims` key of
  the breadcrumb, in the form `'TARGET KIND ARGUMENT'`
  ([ADR-0017](../../docs/adrs/ADR-0017-claims-written-under-breadcrumb-key.md)).
  The only kind is `has-line`: a line of the target, without the
  spaces and tabs at its start and end, equals the argument
  ([ADR-0019](../../docs/adrs/ADR-0019-has-line-claim-matches-a-trimmed-line.md)).
  A claim may quote a line of code:
  unlike a copied sample, it turns Refuted when the code moves on.
  `bcr audit` checks them: a claim that does not hold is Refuted, and
  fails the pipe
  ([ADR-0023](../../docs/adrs/ADR-0023-audit-checks-claims-against-the-working-tree.md)).
- **PBIs reach the main branch with their change.** A PBI reaches the
  main branch in the same merge as the change it directs, so every PBI
  on the main branch is closed. A PBI has no status.
- **Numbers.** ADRs have four digits and PBIs five, where ASDLC shows
  three, so that neither runs out and file names keep sorting in
  order.
- **Breadcrumbs.** Each ADR, PBI and spec carries a breadcrumb in its
  front matter, under a `breadcrumb` key
  ([ADR-0002](../../docs/adrs/ADR-0002-breadcrumb-metadata.md)).
  It comes from Breadcrumb, not from ASDLC. A spec's id is the name
  of its directory, such as `asdlc`.

## Contract

### Definition of Done
- [ ] Each spec the change touches describes the new state, in the
      same commit.
- [ ] Each claim in `specs/` holds.
- [ ] `bcr extract | bcr verify | bcr audit` exits 0 under
      `set -o pipefail`.

### Regression Guardrails

- `AGENTS.md` tells agents the repository is developed with ASDLC.

- ADR-0001 is in force.

- An Accepted ADR keeps its Context, Decision, Consequences and
  Alternatives Considered; only its Status, its breadcrumb and
  editorial fixes that keep its meaning, such as a typo, change.
  Checked by hand: `git log -p -- docs/adrs/`.

- A PBI stays in `tasks/`. Checked by hand:
  `git log --diff-filter=D --name-only -- tasks/` lists no file.

### Scenarios

```gherkin
Scenario: A change to how the repository is built
  Given a PBI that changes a file or rule this spec names
  When its change is merged
  Then this spec describes the new state in the same commit

Scenario: Going beyond ASDLC
  Given a practice that ASDLC's pages do not describe
  When this repository adopts it
  Then this spec lists it under Constraints in the same commit

Scenario: Superseding ADR-0001
  Given an ADR that supersedes ADR-0001
  When it is accepted
  Then the claim on ADR-0001's status is Refuted
  And this spec is rewritten or removed in the same commit
```
