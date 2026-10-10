---
breadcrumb:
  id: ADR-0032
  type: ADR
  links:
    - amends ADR-0002
    - constrained_by ADR-0018
---
# ADR-0032: A breadcrumb may have a title

Status: Accepted
Date: 2026-10-10

## Context

A breadcrumb names its artifact by an id, such as `ADR-0019`, and the
diagrams `bcr report` draws show that id in each box. An id says which
artifact a box is, not what the artifact is about: to learn that, a
reader opens the file.

The artifact's own title is in its body, as its first heading. `bcr`
reads only the front matter of a file (ADR-0002), and a command after
`bcr extract` reads no file at all, only records (ADR-0015). The
heading of every ADR and PBI here also holds `: `, as in
`# ADR-0019: A has-line claim matches a trimmed whole line`, which
ADR-0002's part of YAML cannot hold as a plain value.

## Decision

- A breadcrumb may have a `title`, written under the `breadcrumb` key:

  ```yaml
  breadcrumb:
    id: ADR-0019
    type: ADR
    title: A has-line claim matches a trimmed whole line
    links: []
  ```

- `title` is optional (ADR-0018). A breadcrumb without it means what
  it meant before the key existed.
- A title is one line of text, written plain, as ADR-0002 writes every
  value. It is not empty and holds no tab.
- A `title` that breaks one of these, such as a bare `title:` or a
  list, is a problem, and its file then has no breadcrumb, as with an
  `id` that is not one word.
- A title is shown to a reader. No rule about a breadcrumb or about
  the set reads it: two breadcrumbs may have the same title, and a
  title changes no link, no claim and no verdict.

## Consequences

- An artifact's breadcrumb can say what the artifact is about, and
  `bcr report` can show it beside the id.
- A breadcrumb written before this ADR stays valid, and is shown by
  its id alone.
- A title that needs `: `, ` #` or a character a plain value cannot
  start with cannot be written. The id is not repeated in a title, so
  `ADR-0019: ` is left out.
- The title and the first heading of the file are two copies of one
  text. Nothing checks that they agree, so one can change without the
  other.
- Which record carries a title, and where it is shown, is defined in
  the specs of the commands.

## Alternatives Considered

- Reading the first heading of the body: no second copy to keep, but
  `bcr extract` would read past the front matter, into Markdown, with
  rules for fenced code and for the two ways a heading is written; and
  the heading of an ADR or a PBI repeats its id.
- A top-level `title` key: Hugo and other tools already read it, and
  ADR-0002 reads no key outside `breadcrumb`.
- Allowing quotes, so that a title may hold `: `: a second exception
  to ADR-0002's plain values, after the entries of `claims`
  (ADR-0017), for a text that can be worded without it.
- A broken title dropped on its own, as a claim is (ADR-0017): the
  breadcrumb would still be printed, but under a label its author did
  not write, and the problem would be easier to miss.
