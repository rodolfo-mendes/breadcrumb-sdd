---
breadcrumb:
  id: ADR-0017
  type: ADR
  links:
    - amends ADR-0002
    - amends ADR-0010
    - follows ADR-0012
---
# ADR-0017: Claims are written under the breadcrumb key

Status: Proposed
Date: 2026-10-07

## Context

A claim is a statement about the repository's files that a tool can
check: it is what lets a breadcrumb turn red when the code moves away
from what an artifact says. Today the specs write their claims in the
body, under `Claims:` lists, and nothing reads them.

A claim is part of the breadcrumb that makes it, so it belongs with
the breadcrumb's other metadata, under the `breadcrumb` key
(ADR-0002), not spread over the document.

ADR-0002 allows only plain values on one line. Most claims cannot be
written that way. Of the five claims this repository holds, four
break when put through the YAML parser as plain values:

- A value cannot start with a backtick: the parser stops.
- ` #` starts a comment, so `docs/bcr.md has-line ### verify` is read
  as `docs/bcr.md has-line`, with no problem reported.
- `: ` makes a map, so `Status: Accepted` is no longer a value.

## Decision

- A breadcrumb's claims are a list under the key `claims`, beside
  `id`, `type` and `links`:

  ```yaml
  breadcrumb:
    id: extract
    type: spec
    links:
      - follows ADR-0002
    claims:
      - 'docs/bcr.md has-line ### extract'
  ```

- Each entry is written on one line, in single quotes, even when a
  plain value would parse. A `'` inside it is written `''`. An entry
  in double quotes, or not quoted, is a problem.
- An entry with a comment after it, on the same line, is a problem.
- Inside the quotes, an entry is three parts: `TARGET KIND ARGUMENT`.
  `TARGET` and `KIND` are each one word, followed by one space;
  `ARGUMENT` is the rest of the entry, kept byte for byte.
- `TARGET` is the file the claim is about: a path from the root of the
  repository, with `/` between its parts. It may not start with `/`,
  have a part that is `..`, or hold white space.
- `KIND` names a kind of claim `bcr` knows, and `ARGUMENT` is what
  that kind allows. Each kind is defined by its own ADR. A kind `bcr`
  does not know, or an argument its kind does not allow, is a
  problem.
- A claim with a problem is reported, and dropped. It does not stop
  the breadcrumb: its id, type, links and other claims are still read.
- Single quotes are allowed only in the entries of `claims`. Every
  other value keeps ADR-0002's rule.

## Consequences

- Claims live in the front matter of the artifact that makes them.
  The `Claims:` lists in the specs' bodies move there, and the specs
  say so.
- `bcr extract` reads claims and prints them as records of a new kind
  (ADR-0011); `specs/extract/spec.md` defines the record.
- `bcr verify` ignores the new records and passes them on (ADR-0016).
  Checking a claim against its target is the job of `bcr audit`.
- ADR-0010's list of what the core checks gains one exception: a
  single-quoted value, in a claim entry only.
- A breadcrumb can be read while one of its claims is not. Until
  problems travel down the pipe, a later stage cannot see the dropped
  claim; the problem shows on `bcr extract`'s standard error and in
  its exit status.
- Each new kind of claim changes `bcr extract`, which checks it, as
  well as the stage that checks it against the files.

## Alternatives Considered

- Claims in the body, under `Claims:` lists, as the specs write them
  now: no YAML involved, and each claim next to the item it checks,
  but the breadcrumb would be spread over the document, and its
  metadata found in two places.
- Plain values, as ADR-0002 allows: four of five real claims break,
  one of them silently.
- Double quotes: they turn `\t` and `\n` into a tab and a newline,
  which would break the one-line rule and the records' fields.
- A map for each claim, such as `target:` and `has-line:`: ADR-0002
  forbids maps inside lists, the same quoting problem stays inside
  each value, and every claim takes two or three lines.
- Quotes only when YAML needs them: two ways to write one claim, and
  an author who does not know YAML's rules writes a claim that is cut
  short.
