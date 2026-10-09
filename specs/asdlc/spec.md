---
breadcrumb:
  id: asdlc
  type: spec
  links:
    - constrained_by ADR-0001
    - constrained_by ADR-0002
    - constrained_by ADR-0004
    - constrained_by ADR-0005
    - constrained_by ADR-0009
    - constrained_by ADR-0025
    - constrained_by ADR-0026
    - constrained_by ADR-0028
    - constrained_by ADR-0029
  claims:
    - 'AGENTS.md has-line This repository is developed with ASDLC'
    - 'CONTRIBUTING.md has-line ## Where a decision goes'
    - 'VISION.md has-line ## Decision heuristics'
    - 'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'
    - 'breadcrumb.rules has-line # The shape of this repository (ADR-0026): ADR above spec above PBI, with amends'
    - '.github/workflows/check.yml has-line "$RUNNER_TEMP/bcr" extract | "$RUNNER_TEMP/bcr" verify | "$RUNNER_TEMP/bcr" audit | "$RUNNER_TEMP/bcr" report > "$RUNNER_TEMP/report.md"'
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
| Product Vision | `VISION.md` | Product Vision |
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
- `VISION.md` is at the root too. It says who Breadcrumb is for, its
  principles, and the tie-breakers for a choice no rule decides; it
  decides nothing itself.
- A change that touches code or a spec starts with a PBI that
  `changes` each spec it touches. Any other change, such as an ADR on
  its own or a change to `AGENTS.md`, `CONTRIBUTING.md` or
  `README.md`, needs no PBI; its commit type is `docs`
  ([ADR-0028](../../docs/adrs/ADR-0028-a-pbi-directs-only-changes-to-code-or-a-spec.md)).
- A PBI whose change is about how the repository is built names this
  spec in its Context.
- A decision goes where whoever must obey it reads it
  ([ADR-0029](../../docs/adrs/ADR-0029-a-decision-is-recorded-where-its-actor-reads.md)):
  how `bcr` behaves or how the repository is built, an ADR; a rule
  for making a change, this spec, explained in `CONTRIBUTING.md`; a
  boundary for agents, `AGENTS.md`; what the product is, `README.md`.
- `.github/workflows/check.yml` runs `go test ./...` and then the
  whole pipe on every push, with the `bcr` built from the commit,
  under `set -o pipefail`: a failing test, a problem or a Refuted
  claim fails it. The page `bcr report` writes goes to the job's
  summary and is kept as an artifact of the run, whether the job
  fails or not.

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
- **The layout.** `breadcrumb.rules`, at the root, declares the shape
  of the breadcrumbs: an ADR is `constrained_by` or `amends` an ADR, a
  spec is `constrained_by` an ADR, a PBI `changes` a spec, and only a
  spec carries claims
  ([ADR-0026](../../docs/adrs/ADR-0026-this-repository-uses-the-asdlc-layout.md)).
  `bcr verify` checks every breadcrumb against it
  ([ADR-0025](../../docs/adrs/ADR-0025-breadcrumb-rules-declares-the-shape.md)).
- **`amends`.** An ADR that changes part of an Accepted ADR links to
  it with `amends`, and both stay in force. ASDLC's ADR is superseded
  as a whole.
- **Issues.** Backlog items, open questions and flags are GitHub
  Issues, labelled `backlog`, `question` and `flag`. They hold work and
  debate, not decisions: an issue that settles something closes by
  naming the decision's home
  ([ADR-0029](../../docs/adrs/ADR-0029-a-decision-is-recorded-where-its-actor-reads.md)).

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

- CI runs the tests and the whole pipe on every push, `bcr report`
  included.

- `breadcrumb.rules` is at the root, and declares the ASDLC layout.

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

Scenario: A change that touches only prose
  Given a change that touches no code and no spec
  When it is merged
  Then it carries no PBI
  And its commit type is docs

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
