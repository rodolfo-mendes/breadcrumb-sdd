---
breadcrumb:
  id: ADR-0012
  type: ADR
  title: A link entry is read as two words
  links:
    - amends ADR-0002
---
# ADR-0012: A link entry is read as two words

Status: Accepted
Date: 2026-10-03

## Context

ADR-0002 writes each entry of `links` as a verb and the id of the
breadcrumb it points to, separated by a space, and gives ids no white
space. It does not say what happens with two spaces, a tab, a space at
the end of the line, a verb of two words, or the same link written
twice.

Extra white space is hard for a person to see. A link rejected for a
second space, or for a space at the end of the line, sends its author
looking for a mistake that is not on the screen.

Everything decided here can be checked by reading one file. Whether
the id names a breadcrumb that exists cannot, and is not decided here.

## Decision

- A link entry is split into words at white space: any character that
  Unicode classes as white space, such as a space, a tab or a
  non-breaking space. White space at the start or the end of the entry
  is ignored, and a run of it between two words counts as one
  separator.
- The entry must split into exactly two words: the verb, then the id
  it points to. An entry with one word, or more than two, is a
  violation.
- Two entries in one breadcrumb with the same verb and the same id are
  a violation. Each entry after the first is reported, at its own
  line.
- An entry whose id is the breadcrumb's own id is a violation.
- This amends ADR-0002, whose entries are separated by "a space": any
  run of white space now separates them.

## Consequences

- `implements ADR-0001`, `implements  ADR-0001` and
  `implements ADR-0001 ` are the same link, so writing two of them in
  one breadcrumb is a duplicate.
- White space that cannot be seen never makes an entry wrong, and never
  makes two links different.
- A verb is one word, so a verb and an id can be printed as fields of
  a tab-separated record (ADR-0006) without quoting.
- A verb of two words, such as `depends on`, has to be written as one,
  such as `depends_on`.
- A link to an id that no breadcrumb has is not found by reading one
  file; it is left to a check of the whole repository.

## Alternatives Considered

- Exactly one space, nothing before or after: one form per link, but a
  second space or a trailing one is an error nobody can see.
- Verbs of several words, with the last word taken as the id: allows
  `depends on ADR-0003`, but the verb then holds white space, and two
  spellings of the same verb could differ only in their spaces.
- The same link written twice counted once: harmless to read, but it
  hides a copy mistake, and says nothing the first entry did not.
- Links from a breadcrumb to itself allowed: no verb in this
  repository needs one; it can be allowed when one does.
