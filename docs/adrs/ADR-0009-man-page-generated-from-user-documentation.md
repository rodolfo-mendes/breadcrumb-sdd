---
breadcrumb:
  id: ADR-0009
  type: ADR
  links: []
---
# ADR-0009: The man page is generated from the user documentation

Status: Accepted
Date: 2026-10-03

## Context

The user documentation of `bcr` is `docs/bcr.md`, laid out as a man
page (ADR-0008). Users of a command-line tool also expect `man bcr`,
and a man page is written in roff, not Markdown.

## Decision

- `docs/bcr.1` is generated from `docs/bcr.md` by `internal/manpage`,
  which turns the part of Markdown ADR-0008 allows into roff.
- Both files are committed.
- A test fails when `docs/bcr.1` differs from what `docs/bcr.md`
  generates. `go test ./internal/manpage -update` writes it again.

## Consequences

- The man page cannot say anything `docs/bcr.md` does not, and a
  change to `docs/bcr.md` whose man page was not written again shows
  as a failing test.
- The man page can be read in a clone with `man ./docs/bcr.1`, and the
  release workflow packs it into each archive without running the
  generator.
- The generator is infrastructure (ADR-0007), so it may use libraries;
  it uses none today.

## Alternatives Considered

- Two documents written by hand: they would drift apart.
- A converter such as pandoc or go-md2man: pandoc is a program to
  install, and go-md2man would be a dependency needing an ADR
  (ADR-0005), for the few elements ADR-0008 allows and that a short
  generator already handles.
- Generating the man page only in the release workflow: nothing in
  the repository would show it, and a broken conversion would show
  only at release.
- Writing the Markdown from the man page: roff is harder to write and
  to review than Markdown.
