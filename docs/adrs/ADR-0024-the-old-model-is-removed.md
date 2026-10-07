---
breadcrumb:
  id: ADR-0024
  type: ADR
  links:
    - follows ADR-0001
    - follows ADR-0006
    - follows ADR-0008
---
# ADR-0024: The old model is removed

Status: Accept
Date: 2026-10-07

## Context

Before ADR-0001, this repository recorded its intent in `breadcrumbs/`:
Intakes, Requirements, Technical Decisions and Tasks, defined by the
method specification `docs/breadcrumb-sdd.md`, and read by `bcr check`,
`verdict`, `links`, `new`, `spec`, `init` and `audit-report`. A site
built from the specification's tags, with the audit report of the
latest release, is published on GitHub Pages.

Since ADR-0001, the repository records its intent in ADRs, specs and
PBIs, read by `bcr extract | bcr verify | bcr audit | bcr report`. The
pipe does for them what the old commands did for `breadcrumbs/`.

The old graph is no longer kept up. On 2026-10-07, `bcr verdict` finds
35 of its 124 breadcrumbs Refuted, and no workflow runs it. No one
outside this repository uses the old model.

Two models in one repository and one binary make every reader, person
or agent, first decide which model a file belongs to.

## Decision

- `breadcrumbs/` is deleted, with the commands that read it (`check`,
  `verdict`, `links`, `new`, `spec`, `init` and `audit-report`) and
  the packages only they use, `internal/breadcrumb` and
  `internal/report`.
- `docs/breadcrumb-sdd.md`, the method specification, is deleted, with
  `docs/docs.go`, which embeds it. Breadcrumb is defined by `bcr`'s
  behavior, its specs and its ADRs, as AGENTS.md already says.
- The site is deleted: `cmd/specsite`, `internal/specsite` and
  `.github/workflows/pages.yml`. GitHub Pages is turned off.
- `bcr init` is written again for the new model, in its own spec,
  before the next release.
- No Intake, Requirement, Technical Decision or Task stays in force.
  The rules of the old model that still hold live here:
  - command-line conventions: ADR-0006;
  - the user documentation and its man page: ADR-0008 and ADR-0009;
  - dependencies: ADR-0005;
  - the license: `LICENSE`;
  - the commit format: `CONTRIBUTING.md`;
  - the release: `.github/workflows/release.yml`.
- A file that stays cites these, not an old id. An Accepted ADR or a
  closed PBI that mentions `breadcrumbs/` or an old id keeps the
  mention, as history.

## Consequences

- The pipe, the commands and the documents describe one model.
- The old model stays in Git. The commit that deletes `breadcrumbs/`
  is found with
  `git log --diff-filter=D --format=%h -1 -- breadcrumbs/`, and any old
  breadcrumb is read at its parent, as in
  `git show HASH^:breadcrumbs/TD-0017.md`. Tag v0.7.0, the last release
  with the old model, does not hold IN-0018, RQ-0047 to RQ-0049,
  TD-0037, TD-0038 or TK-0019.
- The 35 Refuted breadcrumbs stop showing. Drift recorded before
  ADR-0001 is not carried into the new model; `bcr verdict` at the
  parent commit still shows it.
- "Spec" names one thing again, a feature spec. ADR-0001's
  Consequences expected the method specification to move on and its
  two vocabularies to end; this ADR ends them. ADR-0001's Decision,
  to build this repository with ASDLC, does not change, so ADR-0001 is
  not amended and its Status stays "Accepted".
- AGENTS.md loses its old Toolchain rows, its ASK rules about
  `breadcrumbs/` and the method specification, and its layout line
  for `breadcrumbs/`.
- Reasons written only in old breadcrumbs, such as why releases are
  built without GoReleaser (TD-0017), are kept only in Git.
- The release after this ADR has no `bcr spec`, and no `bcr init`
  until its new spec is implemented.

## Alternatives Considered

- Freeze `breadcrumbs/` as history, unedited and not audited (Backlog
  item 95): the record stays readable in the tree, but every reader
  still meets two models, and 35 Refuted verdicts that nothing will
  fix.
- Keep auditing `breadcrumbs/` and label it history: its drift stays
  in sight, but the binary carries two models and two sets of
  commands.
- Convert the old breadcrumbs into ADRs and specs: one model, with its
  history in the tree, but it rewrites records as what they never
  were.
- One ADR per rule still in force: each rule gets its reasons back in
  the tree, but ADR-0005, ADR-0006, ADR-0008 and ADR-0009 already
  restate the command-line, documentation and dependency rules, and
  the license, commit format and release are each written where they
  apply.
