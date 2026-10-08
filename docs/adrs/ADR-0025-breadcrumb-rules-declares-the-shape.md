---
breadcrumb:
  id: ADR-0025
  type: ADR
  links:
    - constrained_by ADR-0013
    - constrained_by ADR-0014
    - constrained_by ADR-0021
    - amends ADR-0015
---
# ADR-0025: A repository declares its shape in breadcrumb.rules

Status: Accepted
Date: 2026-10-08

## Context

A breadcrumb has a type, and a link has a verb (ADR-0002, ADR-0012).
`bcr` reads both as words and gives them no meaning: `type: adr` next
to `type: ADR`, a PBI that links to a decision with a verb nobody
else uses, or a claim on a document that should carry none, all pass.
A repository that follows a layout, such as ADR above spec above PBI,
has no place to say so, and nothing tells it when a new file leaves
the layout.

The shape belongs to the repository, as the set of files does
(ADR-0014): written where anyone can read it, and checked by the
stage that already sees every breadcrumb at once, `bcr verify`
(ADR-0015).

## Decision

- A repository declares its shape in a file called `breadcrumb.rules`,
  at its root. `bcr verify` reads it from the current directory.
- The file is a list of lines. A line ends at `\n` or `\r\n`; the last
  line need not end; a byte order mark before the first line is
  ignored. A line of only spaces and tabs, and a line that starts with
  `#`, are ignored. Every other line is of one of three kinds, its
  fields separated by one tab:

  ```
  type	NAME
  link	NAME	VERB	NAME
  claims	NAME
  ```

  - `type` declares a type a breadcrumb may have.
  - `link` allows a link with `VERB` from a breadcrumb of the first
    type to a breadcrumb of the second.
  - `claims` allows a breadcrumb of the type to carry claims.
- A name or a verb is one or more characters, none of them white
  space. Names and verbs are compared as exact bytes: `spec` and
  `Spec` are two types.
- Lines may come in any order. No comment follows a field.
- Each of these is a problem at its line in `breadcrumb.rules`:
  - a line of a kind other than `type`, `link` and `claims`;
  - a line with a number of fields other than 2 for `type` and
    `claims`, and 4 for `link`;
  - a field that is empty or holds white space;
  - a `link` or `claims` line that names a type no `type` line
    declares;
  - a line that repeats an earlier one.
- With no `breadcrumb.rules`, no shape is checked, and `bcr verify`
  does what it did before. A file that declares nothing, an empty one
  included, declares an empty shape: every breadcrumb's type is then
  undeclared.
- A shape is closed. With a valid `breadcrumb.rules`, each of these is
  a problem, at the record that breaks it:
  - a breadcrumb whose type no `type` line declares;
  - a link whose subject's type, verb and object's type match no
    `link` line, where the subject is the breadcrumb the link is
    written in, and the object the one it points to;
  - a claim of a breadcrumb whose type has no `claims` line.
- One cause is one problem. A link or a claim of a breadcrumb whose
  type is undeclared is not judged, and neither is a link to such a
  breadcrumb, a link that points nowhere, or a link whose subject or
  object has an id two breadcrumbs share.
- A problem in `breadcrumb.rules` is printed as
  `breadcrumb.rules:LINE: MESSAGE` and as a `problem` record
  (ADR-0021), and exits 1. The shape is then unknown, so nothing is
  judged against it; the checks that need no shape still run.
- A `breadcrumb.rules` that exists and cannot be read stops
  `bcr verify`: it writes a message starting `bcr: ` and exits 2
  (ADR-0006).
- Reading the file is infrastructure; what the rules mean, and every
  check against them, is in the core (ADR-0013).

## Consequences

- `bcr verify` now reads one file. It still reads no file that
  carries a breadcrumb, and which files make up the set is still
  decided before it (ADR-0015).
- `bcr verify` must be run from the root of the repository, as
  `bcr extract` is. Run from elsewhere, it finds no
  `breadcrumb.rules` and checks no shape; in the pipe, `bcr extract`
  has then already failed for want of `.breadcrumbs`.
- `bcr verify` reads `claim` records, for the id, path and line of
  each. Their kind and argument stay `bcr audit`'s.
- A typo in the file, such as `lnik`, is seen at once, and so is a
  file written for a later `bcr` that has a kind of line this one
  does not know.
- A layout that lets every type carry claims writes one `claims` line
  for each type.
- Deleting `breadcrumb.rules` turns every shape check off without a
  sign. A repository that wants the file kept makes a claim about it.
- A `bcr` older than this ADR ignores the file.
- A missing link is not a problem: the rules say which links are
  allowed, not which are required.
- New needs come as new kinds of line, each through an ADR.

## Alternatives Considered

- Rules built into `bcr` for one layout: nothing to write, but every
  repository would have to follow ASDLC as this one does.
- The shape in `.breadcrumbs`: one file, but `bcr extract` would read
  rules it does not check, and a pattern and a rule would share a
  syntax.
- YAML or TOML: richer, but three kinds of line need no nesting, and
  the file reads and diffs line by line like the records do.
- Spaces between fields: easier to type, but ids, types and verbs are
  words in a format that already separates fields with a tab
  (ADR-0011). A line typed with spaces is reported with that advice.
- An open shape, where only what a line forbids is a problem: a new
  type or verb would pass until someone wrote a line against it.
- `bcr extract` checks the shape: it sees one file at a time, and a
  link's object is in another file.
