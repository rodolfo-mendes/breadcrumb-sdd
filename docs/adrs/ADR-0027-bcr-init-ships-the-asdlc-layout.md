---
breadcrumb:
  id: ADR-0027
  type: ADR
  title: bcr init ships the ASDLC layout
  links:
    - constrained_by ADR-0002
    - constrained_by ADR-0014
    - constrained_by ADR-0025
---
# ADR-0027: bcr init ships the ASDLC layout

Status: Accepted
Date: 2026-10-08

## Context

A repository that adopts Breadcrumb with ASDLC has to write its shape
(ADR-0025), name its files in `.breadcrumbs` (ADR-0014), and know
where each artifact goes, all before its first breadcrumb. `bcr init`
sets up only what every repository needs: a `.breadcrumbs` of
comments, a section for agents, and a workflow (`specs/init`).

Breadcrumb is a layer over whatever method a repository uses, and
`bcr` knows no method: with no `breadcrumb.rules`, nothing is checked
against a shape. Writing ASDLC's names into a repository, its types,
directories and templates, puts that method's vocabulary inside
`bcr`. That crosses the line drawn when Breadcrumb was split from any
one method, so it is opt-in, and it is recorded here.

This repository departs from ASDLC: its ids have four and five
digits, and one ADR `amends` another (ADR-0026, `specs/asdlc`). A
layout copied from it would hand those departures to every adopter.

## Decision

- `bcr init --layout NAME` sets up a layout as well as the pieces
  `bcr init` always sets up. `asdlc` is the only layout. A `NAME`
  `bcr` does not know is a usage error: nothing is written, the
  message names the layouts `bcr` knows, and the exit status is 2.
  Without `--layout`, `bcr init` declares no shape, as before.
- The ASDLC layout follows ASDLC's conventions, not this
  repository's: ADRs in `docs/adrs/ADR-NNN-slug.md`, specs in
  `specs/feature-name/spec.md` with the directory in kebab-case, and
  PBIs in `tasks/PBI-NNN.md`.
- It writes `breadcrumb.rules`:

  ```
  # bcr init --layout asdlc, bcr VERSION
  type	ADR
  type	spec
  type	PBI
  link	ADR	constrained_by	ADR
  link	ADR	supersedes	ADR
  link	spec	constrained_by	ADR
  link	PBI	changes	spec
  claims	spec
  ```

  A new ADR that replaces another links to it with `supersedes`, in
  its own front matter, as ASDLC supersedes ADRs and never amends
  them.
- It writes `.breadcrumbs`:

  ```
  # bcr init --layout asdlc, bcr VERSION
  docs/adrs/*.md
  !docs/adrs/TEMPLATE.md
  specs/*/spec.md
  tasks/*.md
  !tasks/TEMPLATE.md
  ```

- It writes three templates, with ASDLC's sections:
  `docs/adrs/TEMPLATE.md`, `specs/TEMPLATE.md` and
  `tasks/TEMPLATE.md`. Each starts with front matter that is a valid
  breadcrumb of its type, with `links: []`, and the links the layout
  allows written as comments below it. `.breadcrumbs` excludes the
  ADR and PBI templates; the spec template matches no pattern.
- It adds a `## ASDLC` section to the agents file: where each artifact
  lives, how it is named, where its template is, and that only specs
  carry claims. For the links allowed, it points to
  `breadcrumb.rules` and does not copy them. The `## Breadcrumbs`
  section stays for breadcrumbs and `bcr`. Each section is found by
  its heading.
- The first line of each `.breadcrumbs` and `breadcrumb.rules` that
  `bcr init` writes is a comment naming the command and the release:
  `# bcr init --layout asdlc, bcr 0.9.0`, or `# bcr init, bcr 0.9.0`
  without a layout. `specs/init` states its exact form.
- When `.breadcrumbs` or `breadcrumb.rules` exists, `--layout` writes
  nothing and exits 2. To reset a layout, delete both and run it
  again. A template, the agents sections and the workflow are left
  as they are when present, as `bcr init` leaves them.
- The text of the layout is held in `bcr`.

## Consequences

- `bcr` carries ASDLC's names, behind `--layout` only. The core still
  knows no layout: `breadcrumb.rules` is read like any repository's.
- The names are practices: `bcr` checks no file's name, so they live
  in the agents section and the templates.
- A repository keeps the files of the release it adopted when the
  layout changes; their first line tells which. A way to move to a
  later layout is left for later.
- `bcr init` followed by `bcr init --layout asdlc` refuses: the first
  wrote a `.breadcrumbs`. Delete it first.
- `specs/init`'s guardrail "bcr init writes no breadcrumb" becomes
  "bcr init writes no breadcrumb that `.breadcrumbs` names": the
  templates carry front matter.
- Three-digit ids run out at 999. A repository departs as this one
  does, and says so.
- This repository's shape stays ADR-0026, which goes beyond the
  layout with `amends`.
- Another layout comes as another name, through its own ADR.

## Alternatives Considered

- This repository's conventions: already in use here, but its
  departures from ASDLC would become every adopter's default.
- `amends` in the layout's rules: this repository uses it, but ASDLC
  does not.
- Templates without front matter, their breadcrumb in a code block:
  no exclusions needed, but the block an agent most needs to copy
  would have to be moved by hand.
- Leaving an existing `.breadcrumbs`, as `bcr init` does: a
  `.breadcrumbs` of comments would stay, and the pipe would read
  nothing with no sign. A layout's pieces only work together.
- Layouts as files outside `bcr`: no vocabulary in the binary, but
  machinery for one layout. It can be drawn out when a second one
  comes.
- Skills and agents for a harness: a harness this method does not
  assume, and none of them has earned its place yet.
