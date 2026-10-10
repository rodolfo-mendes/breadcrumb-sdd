---
breadcrumb:
  id: ADR-0013
  type: ADR
  title: The core holds the meaning of a breadcrumb, not its format
  links:
    - amends ADR-0007
    - amends ADR-0010
---
# ADR-0013: The core holds the meaning of a breadcrumb, not its format

Status: Accepted
Date: 2026-10-03

## Context

ADR-0007 puts Breadcrumb's rules in the core of `bcr`, and everything
else in infrastructure. It counts the rules of the front matter among
the core's: ADR-0002's small part of YAML is checked by the core, and
every problem is found by the core. ADR-0010 says the same: the core
checks that a breadcrumb uses no anchor, alias, tag, multi-line value
or flow collection other than `[]`.

That makes a file format part of the definition of Breadcrumb. A
breadcrumb is an id, a type and links to other breadcrumbs. YAML front
matter is where this repository writes one (ADR-0002), not what a
breadcrumb is. If breadcrumbs were written somewhere else, for example
as JSON in a file next to each artifact, every rule about YAML would
leave the core, and the specification extracted from the core would
change, though no breadcrumb would mean anything different.

Some problems can only be found while a format is read: front matter
that does not close, or YAML that does not parse. They are found
before there is a breadcrumb for the core to check, so under ADR-0007
they have no place.

## Decision

- The core holds the meaning of a breadcrumb: the properties it has,
  the values they take, and the rules about its links. It knows no
  file format.
- A format is infrastructure. Its reader finds where a breadcrumb is
  written in a file, checks the rules of the format, and hands the
  core the breadcrumb's properties, each value as the text written or
  a list of such texts, with the line it is on.
- A format's reader finds the problems of its format, with the line
  where each is. The core finds every other problem. Commands and the
  report show both alike, and find none of their own.
- The rules of a format are decided by ADRs, as ADR-0002 and ADR-0010
  decide YAML front matter. They are not part of Breadcrumb.
- This amends ADR-0007, whose core checks the rules of the front
  matter and finds every problem, and ADR-0010, whose core checks
  ADR-0002's part of YAML.

## Consequences

- The rules decided so far fall on either side:
  - The core: a breadcrumb has an `id`, a `type` and `links`; `id`
    and `type` are text with no white space; `links` is a list; a
    link entry is two words, is not written twice, and does not
    point to its own breadcrumb (ADR-0012); ids are unique
    (ADR-0003).
  - The reader of YAML front matter: the front matter opens and
    closes with `---`; it is YAML; it uses ADR-0002's part of YAML;
    it has at most one `breadcrumb` key; only Markdown files carry
    breadcrumbs.
- A link entry is read as two words by the core, since ADR-0012
  decides it on the entry's text. A format that writes the verb and
  the id apart needs an ADR that says where that rule goes.
- Changing the format changes its reader, not the core. The
  specification describes what a breadcrumb holds, not where it is
  written.
- A value reaches the core as the text written, so `id: 0001` is
  still `0001` (ADR-0007).
- A format's reader is infrastructure, so it may use a library
  (ADR-0005), and its tests read text in that format. The core's
  tests build breadcrumbs as values.
- `go list -f '{{.Imports}}'` on the core's packages still lists only
  the standard library and other core packages.
- The Architecture of `specs/extract/spec.md`, where the core checks
  ADR-0010's part of YAML, changes once this ADR is accepted.

## Alternatives Considered

- Keep the format in the core, as ADR-0007 decides: the specification
  of Breadcrumb would change with the file format, and the problems a
  reader finds before the core has a breadcrumb would have no place.
- The core splits the front matter, and the reader hands it a failed
  parse as a fact to report: every problem stays in the core, but the
  core still knows that a breadcrumb is YAML front matter.
- Format problems reported apart from the core's, with an exit status
  of their own: the user would see two kinds of problem in one file,
  when a mistake in either is fixed the same way, by editing it.
