---
breadcrumb:
  id: ADR-0002
  type: ADR
  links: []
---
# ADR-0002: Record breadcrumbs in the artifacts' front matter

Status: Proposed
Date: 2026-09-30

## Context

A breadcrumb is metadata attached to an artifact: an id, a type and
links to other breadcrumbs. Breadcrumbs form a graph, and that graph
is derived from the artifacts, so the artifacts are the record and
anything a tool writes from them is a copy.

Front matter is where Markdown tools already look for metadata,
and other tools read it too. For example: Docusaurus reads `id`
and Hugo reads `type`. Thus, breadcrumb properties are namespaced
to avoid clashing with other tools.

## Decision

An artifact carries its breadcrumb in its Markdown front matter,
under one `breadcrumb` key:

```yaml
---
breadcrumb:
  id: PBI-00001
  type: PBI
  links:
    - implements ADR-0001
    - changes asdlc
---
```

- `id` names the breadcrumb. `type` names the kind of artifact.
  Each is a string with no whitespace; this repository chooses their
  formats.
- `links` lists links to other breadcrumbs, one per line, each a verb
  and the id of the breadcrumb it points to, separated by a space.
  An artifact with no links writes `links: []`.
- A file whose front matter has no `breadcrumb` key do not define a
  breadcrumb. Keys outside `breadcrumb` are not read.
- A `breadcrumb` key without `id`, `type` or `links`, or with a bare
  `links:` and no list, is a violation.
- A file holds at most one breadcrumb.
- Only Markdown files carry breadcrumbs for now.
- The front matter uses a small part of YAML: the `breadcrumb` map,
  plain one-line values, a list of plain one-line values, and `[]`.
- Links may be added or removed at any time. Changing an artifact's
  breadcrumb is not changing the artifact.

## Consequences

- ADR-0001, PBI-00001, `specs/asdlc/spec.md` and this ADR carry a
  breadcrumb from here on, and so does every new ADR, PBI and spec.
- An Accepted ADR may change its breadcrumb as well as its Status.
  `specs/asdlc/spec.md` says so, and lists the front matter among the
  places this repository goes beyond ASDLC.
- Until `bcr` reads front matter, breadcrumbs are checked by eye.
- A change of id or type is not a violation: nothing that reads one
  state of the repository can see it. Links to the old id stop
  resolving, and that is what shows.
- The verbs and the types this repository allows are not decided
  here.

## Alternatives Considered

- `id` and `type` as top-level keys: a file with both, written for
  another tool, would become a breadcrumb, and a file missing one
  would be silently skipped.
- A `bcr` key: it names the record after the tool that reads it.
- One key per property, such as `breadcrumb-id`: three keys to keep
  together, where one map holds them.
- A sidecar file next to each artifact: the metadata could drift from
  its file, and a rename would have to move two files.
- A section at the end of the body: it mixes metadata with the text
  the artifact exists for, and has no convention tools already know.
