---
breadcrumb:
  id: ADR-0014
  type: ADR
  title: A repository names the files that carry its breadcrumbs in .breadcrumbs
  links: []
---
# ADR-0014: A repository names the files that carry its breadcrumbs in .breadcrumbs

Status: Accepted
Date: 2026-10-04

## Context

`bcr extract` reads the files it is given as operands, and leaves the
choice of files to the caller. That is enough for a rule about one
file. A rule about the whole repository, such as unique ids
(ADR-0003), judges a set of files: with the set left to each caller,
two runs over the same repository can disagree, and a breadcrumb in a
file nobody passed is never seen.

The set should belong to the repository, written where anyone can
read it. Most people already know one way of naming files by pattern:
`.gitignore`. In this repository, three patterns cover every file that
carries a breadcrumb.

## Decision

- A repository names the files that carry its breadcrumbs in a file
  called `.breadcrumbs`, at its root.
- Each line of `.breadcrumbs` holds one pattern. Blank lines, and
  lines that start with `#`, are ignored.
- A pattern is matched against the whole path of a file from the root,
  with `/` between its parts. `*` matches any characters except `/`,
  `?` matches one character except `/`, and `[...]` matches one
  character from a set.
- A pattern that starts with `!` excludes the files it matches. When
  several lines match a file, the last one decides; a file that no
  line matches is not read.
- A line that uses anything else, such as `**` or a `/` at its end, is
  a problem at its line in `.breadcrumbs`.
- `bcr extract`, given no operand, reads every file `.breadcrumbs`
  names, in order of path compared as bytes. Given operands, it reads
  exactly those files, as before.
- The files are found from the current directory, which is the root of
  the repository. `.git/` is not entered, and symbolic links are not
  followed.
- With no operand and no `.breadcrumbs`, `bcr extract` writes a
  message starting `bcr: ` to standard error and exits 2 (ADR-0006).

## Consequences

- This repository's `.breadcrumbs` has three lines:

  ```
  docs/adrs/*.md
  specs/*/spec.md
  tasks/*.md
  ```

- A rule about the whole repository judges the files `.breadcrumbs`
  names. A breadcrumb in any other file is not seen, so a file put in
  a folder no pattern covers drops out without a sign.
- A file `.breadcrumbs` names that has no `breadcrumb` key is still
  ignored (ADR-0002).
- Unlike `.gitignore`, a pattern without `/` matches only at the root:
  `*.md` names the Markdown files at the root, not in every folder.
- Reading `.breadcrumbs` and finding the files is infrastructure
  (ADR-0013): it decides which files are read, not what a breadcrumb
  means.
- `specs/extract/spec.md`, where `bcr extract` with no operand is a
  usage error, changes once this ADR is accepted.
- More of `.gitignore`'s syntax, such as `**`, can come later through
  an ADR, and through a library if it needs one (ADR-0005).

## Alternatives Considered

- The caller always chooses the files: one pattern for every command,
  but whether ids are unique would depend on who asked.
- A rule built into `bcr`, such as every Markdown file except under
  `.git/`: nothing to write, but it reads every vendored or generated
  file, and a repository cannot change it.
- The files git tracks: exactly what is committed, but it needs git,
  and an agent's new PBI is not seen until it is added.
- Patterns that name the files to skip, as `.gitignore` does: everything
  else is read, so a new folder of documents is read without anyone
  deciding it should be.
- The whole syntax of `.gitignore`: familiar, but a library for rules
  this repository does not need yet.
- The patterns in `breadcrumb.rules`, with the types and link rules:
  one file for all of a repository's settings, but that file does not
  exist, and each file would then have two jobs.
