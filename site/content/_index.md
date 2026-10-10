---
title: Breadcrumb
description: Breadcrumb makes drift between a repository's intent and its code visible. Its toolkit, bcr, checks what a repository's documents claim about its files.
---
# Breadcrumb {#top}

Breadcrumb makes drift between a repository's intent and its code
visible.

[![The Breadcrumb repository on GitHub](github-mark.svg)](https://github.com/rodolfo-mendes/breadcrumb-sdd)

## The problem {#problem}

Code says what a repository does, not why. The why lives in tickets,
chats, design documents and people's heads, and it drifts away from
the code without anyone noticing. A document that once described the
code keeps reading as true after the code has changed.

Later, nobody can tell whether the code still does what was asked of
it, or which decision a piece of code was meant to carry out.

## How it works {#how-it-works}

Code stays the source of truth for what the repository does. The
documents that hold its intent, such as decisions, specs and tasks,
each carry a breadcrumb: a few lines of YAML front matter that say
what the document is, which documents it links to, and what it claims
about the repository's files.

```yaml
---
breadcrumb:
  id: verify
  type: spec
  links:
    - constrained_by ADR-0016
  claims:
    - 'docs/bcr.md has-line ### verify'
---
```

- `id` names the breadcrumb, and `type` says what kind of document
  carries it.
- Each entry of `links` is a verb and the id of another breadcrumb.
- Each entry of `claims` is a file, a kind of claim and its argument.

The toolkit, `bcr`, reads the breadcrumbs, checks that their ids are
unique and their links point to a breadcrumb, and checks each claim
against the files. Each breadcrumb then gets one of three verdicts:

- **Confirmed**: all of its claims hold.
- **Refuted**: one of its claims fails. The document and the file
  disagree. Either the code drifted from the intent, or the claim went
  stale because the intent changed or its target moved; a person
  decides which.
- **Undecided**: it has no claims, or it has a problem. Nothing about
  it was checked.

The only kind of claim is `has-line`: the file has a line equal to
the text, once the spaces and tabs at the line's start and end are
removed. It is a text match. Confirmed means the line is there, not
that the document is right.

## Honest scope {#scope}

Breadcrumb makes drift visible after the fact; it does not prevent
it. It lets drift happen, and shows where a document and a file
disagree.

A breadcrumb describes the code and never replaces it. What it holds
is what code cannot: why.

A repository records as much or as little as its owners choose, and a
claim can be checked by hand, against the files, without `bcr`.

## Try it {#try-it}

`bcr` is a single Go binary. Install it with Go 1.22 or later:

```
go install github.com/rodolfo-mendes/breadcrumb-sdd/cmd/bcr@latest
```

Put a breadcrumb at the top of one file, here `notes.md`:

```yaml
---
breadcrumb:
  id: notes
  type: note
  links: []
  claims:
    - 'notes.md has-line # Notes'
---
# Notes
```

Then run four of the commands of `bcr` as a pipe, on that file:

```
bcr extract notes.md | bcr verify | bcr audit | bcr report
```

- [`bcr extract`](docs/#extract) reads each breadcrumb and prints it
  as records.
- [`bcr verify`](docs/#verify) checks the whole set: unique ids, and
  links that point to a breadcrumb.
- [`bcr audit`](docs/#audit) checks each claim against the files, and
  gives each claim and each breadcrumb a verdict.
- [`bcr report`](docs/#report) writes one Markdown page: every
  breadcrumb by verdict, and every problem.

The page lists `notes` as Confirmed. Change the heading of `notes.md`
and run the pipe again: it is Refuted.

To set a whole repository up, run [`bcr init`](docs/#init) in its
root. Every command, record, output and exit code is in the
[documentation](docs/).

## Who it's for {#audience}

People who build software with AI agents writing most of the code,
first a solo builder or a small team. Months later they need to know
whether the code still does what was asked, and which decision a piece
of code was meant to carry out. They will not keep up a heavy process
to find out.

Adopting Breadcrumb should cost as little as possible: one binary and
a few lines of front matter, with no service to run.

## Status {#status}

The released version is v0.9.0; each release is on the
[releases page](https://github.com/rodolfo-mendes/breadcrumb-sdd/releases).

Before 1.0, a command, a flag or an output may still change between
releases; a breadcrumb's keys and a kind of claim keep their meaning.

Deferred: `has-line` is the only kind of claim there is.

The repository of Breadcrumb is developed with Breadcrumb itself: its
decisions, specs and tasks each carry a breadcrumb, and every push
audits them.
