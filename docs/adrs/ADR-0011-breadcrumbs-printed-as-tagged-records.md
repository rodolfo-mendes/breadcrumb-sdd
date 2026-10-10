---
breadcrumb:
  id: ADR-0011
  type: ADR
  title: Breadcrumbs read from files are printed as tagged records
  links: []
---
# ADR-0011: Breadcrumbs read from files are printed as tagged records

Status: Accepted
Date: 2026-10-03

## Context

Every later part of `bcr`, from the check of unique ids to the cache
in `.bcr/`, needs breadcrumbs as data, not as front matter. The first
command of the new model therefore reads breadcrumbs from the files it
is given and prints them. Which files to read is left to the caller,
through operands.

A breadcrumb has two shapes of data: its own properties, `id`, `type`
and `path`, and its links, each a verb and the id it points to. Under
ADR-0006, a result on standard output is one record per line, with
tab-separated fields, so the two shapes cannot share one table.

This ADR covers only that command. A later command that prints more
than one kind of record decides its own output.

## Decision

- The command that reads breadcrumbs from files prints them as tagged
  records on standard output: one record per line, whose first field
  names the record's kind.
- A breadcrumb's properties and its links are records of different
  kinds, in one stream.
- Each file is read once, and every record about its breadcrumb comes
  from that read.
- Consumers select records by their first field. New kinds of record
  may be added; a consumer that does not select by kind relies on
  something this ADR does not promise.

## Consequences

- A file's whole breadcrumb comes from one call, so a person or an
  agent gets one answer in one output.
- The records about one breadcrumb always describe the same state of
  its file.
- A malformed breadcrumb is reported once, on standard error
  (ADR-0006).
- The table of any one kind is a filter away: select the lines of that
  kind and drop the first field.
- The names of the kinds, the fields of each, and the order of the
  records are defined in the command's spec.

## Alternatives Considered

- Two commands, one printing breadcrumbs and one printing links: each
  prints one uniform table, but a file edited between the two calls
  gives two states; a malformed breadcrumb is reported by both, or by
  one only; and a whole breadcrumb takes two calls. Their output can
  be derived from tagged records with a filter, but tagged records
  cannot be derived from theirs: no script restores one read or one
  report.
- One command with a flag choosing the kind of record: one kind per
  run, with the same two reads and two reports.
- Links folded into the breadcrumb's line: one line per file, but a
  varying number of fields, which `cut` and `awk` handle badly.
