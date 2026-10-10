---
breadcrumb:
  id: ADR-0008
  type: ADR
  title: docs/bcr.md is bcr's user documentation, laid out as a man page
  links: []
---
# ADR-0008: docs/bcr.md is bcr's user documentation, laid out as a man page

Status: Accepted
Date: 2026-10-03

## Context

Each command of `bcr` is specified in `specs/`, one spec per command,
as `specs/asdlc/spec.md` lays out a spec for each feature. A spec is
written for the people and agents who change `bcr`, and its Contract
is what the code is checked against. The person who runs `bcr` needs
something else: what each command does and how to call it.

`docs/bcr.md` was the contract of `bcr`, laid out as a man page. This
ADR moves its layout into the ADRs, and makes the document what it has
always been read as: documentation for the user.

## Decision

- `docs/bcr.md` is the user documentation of `bcr`. It explains every
  command to the person who runs it, and is not a spec.
- It has the sections of a man page: NAME, SYNOPSIS, DESCRIPTION,
  COMMANDS, EXIT STATUS and SEE ALSO. Each command has its own
  subsection under COMMANDS, giving its synopsis, flags, operands,
  standard input, standard output and exit status.
- It uses a small part of Markdown: headings, paragraphs, lists, code
  spans, fenced code blocks and links. It has no tables and no HTML.
- A change to a command's spec that changes what its user sees
  changes `docs/bcr.md` in the same commit.
- Each command's subsection under COMMANDS has the name of its spec's
  directory: `### check` for `specs/check/`. `docs/bcr.md` does not
  link to the specs.

## Consequences

- A reader who knows man pages finds each part where they expect it,
  and the document can become a man page without being rewritten.
- Where `docs/bcr.md` and a command's spec disagree, the spec is
  right, and `docs/bcr.md` is fixed.
- A command's spec and its documentation are found from each other by
  name. Renaming a command renames both.
- The links in `docs/bcr.md` to Technical Decisions and Requirements
  in `breadcrumbs/` go as each command gets its spec: they do not
  ship with a release, and point at history.
- `AGENTS.md` still calls `docs/bcr.md` the contract of `bcr`, and
  changes once this ADR is accepted.

## Alternatives Considered

- `docs/bcr.md` as the contract: the user and the code would read the
  same document, and a change for one would be written for the other.
- The documentation in `README.md`: the README introduces the method
  to newcomers; the documentation is reference.
- Links from `docs/bcr.md` to each command's spec: they do not resolve
  in a release or in `man`, and lead users to text written for
  maintainers.
- `docs/bcr.md` generated from the specs: it cannot drift, but needs a
  generator for specs whose format is not settled.
- Free layout: every command would be documented differently, and a
  man page would need text of its own.
