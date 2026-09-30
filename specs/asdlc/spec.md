# Feature: ASDLC

How this repository is built with ASDLC: which of its patterns it
uses, where their files live, and where it goes beyond ASDLC.

## Blueprint

### Context

[ADR-0001](../../docs/adrs/ADR-0001-adopt-asdlc.md) gives the reasons
for adopting ASDLC.

This spec covers how the repository is built, not what Breadcrumb
does. The method specification, `docs/breadcrumb-sdd.md`, defines
Breadcrumb. It is the product, not a feature spec.

### Architecture

| Pattern | Files | Follows |
|---|---|---|
| AGENTS.md | `AGENTS.md` | ASDLC's AGENTS.md Specification |
| The Spec | `specs/<feature>/spec.md`, the directory in kebab-case | Living Specs |
| The PBI | `tasks/PBI-NNNNN.md`, one file per PBI | PBI Authoring |
| The ADR | `docs/adrs/ADR-NNNN-<slug>.md` | The ADR |

- `NNNN` and `NNNNN` are the next free number of their kind, padded
  with zeros.
- A PBI whose change is about how the repository is built names this
  spec in its Context.
- `breadcrumbs/` holds the Intakes, Requirements, Technical Decisions
  and Tasks written before ADR-0001. `bcr` audits them.

### Constraints

Where this repository goes beyond ASDLC, it is listed here.

- **This spec.** It describes how the repository is built, not a
  feature of Breadcrumb; ASDLC's specs describe features.
- **Claims.** An item of a spec's Contract may carry claims: statements
  about the repository's files that a tool can check and a person can
  check by hand. They come from Breadcrumb, not from ASDLC. Each is
  written below the item it checks, under `Claims:`, in the form
  ``- `path` contains `text` ``. A claim may quote a line of code:
  unlike a copied sample, it turns Refuted when the code moves on.
  Until `bcr` reads claims from specs, they are checked by hand.
- **PBIs reach the main branch with their change.** A PBI reaches the
  main branch in the same merge as the change it directs, so every PBI
  on the main branch is closed. A PBI has no status.
- **Numbers.** ADRs have four digits and PBIs five, where ASDLC shows
  three, so that neither runs out and file names keep sorting in
  order.

## Contract

### Definition of Done
- [ ] Each spec the change touches describes the new state, in the
      same commit.
- [ ] Each claim in `specs/` holds.

### Regression Guardrails

- `AGENTS.md` tells agents the repository is developed with ASDLC.

  Claims:
  - `AGENTS.md` contains `This repository is developed with ASDLC`

- ADR-0001 is in force.

  Claims:
  - `docs/adrs/ADR-0001-adopt-asdlc.md` contains `Status: Accepted`

- An Accepted ADR keeps its Context, Decision, Consequences and
  Alternatives Considered; only its Status changes. Checked by hand:
  `git log -p -- docs/adrs/`.

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
