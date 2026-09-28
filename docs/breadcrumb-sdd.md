# Breadcrumb SDD Specification

Version 0.2.0

## Introduction

Breadcrumb SDD is a method for recording why a repository changed.

Code is the source of truth for what a repository does. Next to it,
small Markdown files called breadcrumbs record intent: the asks made
of the repository, what its software must do, the decisions about how
it is built, and the changes that carried them out. Each breadcrumb
links to the breadcrumbs it derives from, and each change states, as
claims a tool can check, what it left in the files. An audit checks
the claims and passes each verdict up the links, so drift between
intent and code shows as a broken chain that a reader can follow from
the ask down to the line that no longer matches.

The method makes drift visible after the fact; it does not prevent
it.

This document defines the method for any repository that adopts it.
The interface of `bcr`, the toolkit that carries it out, is defined
by its contract, [bcr(1)](bcr.md). Each section names the breadcrumbs
of the Breadcrumb SDD repository that its rules come from; they give
the reasons behind the rules, and are not needed to follow them.

Sections marked as a practice are how the method is meant to be
worked, and no tool checks them. Everything else is a rule, and a tool
that audits breadcrumbs follows it.

## Terms

- The repository root is the directory at the top of the
  repository. `bcr` is run from it.
- A path names a file relative to the repository root. It uses `/`
  as its separator, does not start or end with `/`, and has no empty,
  `.` or `..` segment.
- A line is the text between two line feeds (U+000A), or before
  the first or after the last. A carriage return (U+000D) at the end
  of a line is not part of it. Lines are counted from 1.
- Text is compared as UTF-8 bytes, exactly: no change of case, of
  Unicode normalization or of white space.

## Breadcrumbs

A breadcrumb is a Markdown file directly in the `breadcrumbs/`
directory at the repository root, named by a type prefix, a hyphen,
four digits and `.md`, such as `TD-0001.md`. A file named otherwise,
in a subdirectory of `breadcrumbs/`, or with a lower-case prefix is
not a breadcrumb.

A breadcrumb's id is its file name without `.md`, such as
`TD-0001`. Its type is given by its prefix:

- `IN`, an Intake, records an ask, in the words of whoever made it;
- `RQ`, a Requirement, states one thing the software must do;
- `TD`, a Technical Decision, records one decision about how the
  repository is built;
- `TK`, a Task, records one change, as claims about the files it left.

A repository with no `breadcrumbs/` directory has no breadcrumbs.

Sources: [TD-0001](../breadcrumbs/TD-0001.md),
[TD-0003](../breadcrumbs/TD-0003.md),
[TD-0004](../breadcrumbs/TD-0004.md),
[TD-0011](../breadcrumbs/TD-0011.md),
[TD-0013](../breadcrumbs/TD-0013.md).

### The title

A breadcrumb's first line is its title, and starts with `# `. The
title is the rest of the line, without a leading `ID: `:

```markdown
# TD-0005: Link a breadcrumb to its parents
```

A breadcrumb whose first line does not start with `# ` has a problem
on line 1, and nothing more is read from it: it has no title, no
parents and no claims.

Sources: [TD-0013](../breadcrumbs/TD-0013.md).

### Parents

A breadcrumb names each breadcrumb it derives from, its parent,
on a `Parent:` line:

```markdown
Parent: [IN-0001](IN-0001.md)
```

The `Parent:` lines come right below the title, after any blank
lines, one after the other. They end at the first line that does not
start with `Parent:`. A line further down that starts with `Parent:`,
such as one in an example, is not a link.

A `Parent:` line is exactly `Parent: `, then one Markdown link: its
text in `[` and `]`, holding no `]`, and its target in `(` and `)`,
holding no `)` and no white space. The target is taken relative to the
breadcrumb's own file. The link's text is not read; by convention it
is the parent's id.

A line that starts with `Parent:` but is not written this way is a
problem on that line. A link whose target is not a breadcrumb is a
problem on the line of the link, and links nothing.

A breadcrumb with no `Parent:` line has no parent. A breadcrumb may
have any number of parents, of any type. Links point from a breadcrumb
up to what it derives from; a parent does not list its children.

Sources: [TD-0005](../breadcrumbs/TD-0005.md),
[RQ-0022](../breadcrumbs/RQ-0022.md).

### The body

Below its title and `Parent:` lines, a breadcrumb is free Markdown.
The method reads nothing from the body of an Intake, a Requirement or
a Technical Decision. It reads the claims of a Task.

## The kinds of breadcrumb

### Intake

An Intake records an ask made of the repository, in the words of
whoever made it. It may be edited after it is committed, to fix its
wording or to follow a changed ask; the history of the repository
keeps its earlier words.

An Intake may name another Intake as its parent when it asks for one
part of what the parent asks for. It keeps the parent's words where it
can, and adds only what the asker decided since.

Sources: [TD-0003](../breadcrumbs/TD-0003.md),
[TD-0019](../breadcrumbs/TD-0019.md).

### Requirement

A Requirement states one thing the software must do, precisely enough
that a reader can check the code against it. It names the Intake it
closes a gap of as its parent, and may quote the words of the Intake
it comes from.

Sources: [TD-0004](../breadcrumbs/TD-0004.md).

### Technical Decision

A Technical Decision records one decision about how the repository is
built. It has three sections: `## Decision`, what was decided; `## Why`,
the reasons; and `## What it beat`, the alternatives and why each lost.
It names as parents the breadcrumbs that asked for the decision, if
any.

Sources: [TD-0001](../breadcrumbs/TD-0001.md).

### Task

A Task records one change as claims about the files the change left
in the repository. It names as parents the Requirements and Technical
Decisions it carries out. A Task has no status: a committed Task is
audited.

Its claims are the list items under a heading that is exactly
`## Claims`, up to the next line that starts with `#`. A Task may have
more than one such section; the claims of each count.

Every list item there, marked with `-`, `*` or `+` at any indentation,
must be a claim. A list item that is not a claim is a problem on its
line, and makes the Task Refuted.

Sources: [TD-0011](../breadcrumbs/TD-0011.md).

## Claims

There is one kind of claim, the contains claim:

```markdown
- `cmd/bcr/main.go` contains `func run(`
```

It is exactly `- `, a code span holding a path, ` contains `, and a
code span holding the text, with nothing after it. Neither code span
is empty or holds a backtick. The text holds no line break, so it
matches within one line.

The claim holds when a file can be read at the path and its bytes
contain the text's UTF-8 bytes. It does not hold when there is no
file at the path, the path names a directory, or the file does not
contain the text.

A claim that can be read but does not hold is not a problem: it is
what the audit reports as a verdict.

Sources: [TD-0012](../breadcrumbs/TD-0012.md),
[RQ-0020](../breadcrumbs/RQ-0020.md).

## Verdicts

The audit gives every breadcrumb one verdict: Confirmed, Refuted
or Undecided.

A Task's verdict from its own claims is:

- Refuted, when a list item under `## Claims` is not a claim, or any
  claim does not hold;
- Confirmed, when it has claims and every one holds;
- Undecided, when it has no claims.

A breadcrumb's children are the breadcrumbs that name it as a
parent. The verdicts below a breadcrumb are the verdict of each
child, and, for a Task, its verdict from its own claims. A
breadcrumb's verdict is then:

- Refuted, when any verdict below it is Refuted;
- Confirmed, when there is at least one verdict below it and every one
  is Confirmed;
- Undecided, otherwise. A breadcrumb with nothing below it, such as a
  Requirement that no Task carries out, is Undecided.

So a Refuted Task makes every breadcrumb above it Refuted, up to the
Intake it came from, and a breadcrumb is Confirmed only when
everything below it is.

The links may form a cycle. A breadcrumb reached again while its own
verdict is being found counts as Undecided below the one that reached
it.

Problems other than a list item that is not a claim do not change a
verdict.

Sources: [RQ-0005](../breadcrumbs/RQ-0005.md),
[RQ-0006](../breadcrumbs/RQ-0006.md),
[RQ-0007](../breadcrumbs/RQ-0007.md),
[RQ-0008](../breadcrumbs/RQ-0008.md).

## Problems

A problem is a breadcrumb, or a line of one, that cannot be read
as this document defines. Each problem has the path of its breadcrumb
and the line where it is. The problems are:

- the first line does not start with `# `;
- a line starting with `Parent:` is not a `Parent:` link;
- a `Parent:` link points to no breadcrumb;
- a list item under a Task's `## Claims` is not a claim.

A file in `breadcrumbs/` that is not a breadcrumb is not a problem. A
tool may warn of it, since it is often a breadcrumb named wrong.

Sources: [TD-0024](../breadcrumbs/TD-0024.md),
[RQ-0020](../breadcrumbs/RQ-0020.md),
[RQ-0022](../breadcrumbs/RQ-0022.md),
[RQ-0023](../breadcrumbs/RQ-0023.md).

## Practices

These are how the method is meant to be worked. No tool checks them.

### Numbering

A new breadcrumb takes the number one more than the highest of its
type, or `0001` for the first. A number left free by a deleted
breadcrumb is not used again, since the history of the repository
still names it.

Sources: [TD-0027](../breadcrumbs/TD-0027.md).

### Who writes what

Intakes and Requirements state what the people behind the repository
want. An agent may draft them, and a person approves each one before
it is committed. Agents may write Technical Decisions and Tasks
themselves.

The repository points agents to its breadcrumbs, for example from
`AGENTS.md` at its root.

Sources: [TD-0002](../breadcrumbs/TD-0002.md),
[TD-0006](../breadcrumbs/TD-0006.md).

### A change

A change starts from an ask. A new ask becomes an Intake, and what the
software must do becomes Requirements. A decision about how the
repository is built becomes a Technical Decision. The change itself
becomes a Task, naming the Requirements and Technical Decisions it
carries out, and claiming what it left in the files. Before the change
is merged, the audit shows no Refuted breadcrumb.

### Claims follow the code

When a change rewrites a line that a claim of an earlier Task quotes,
and what the earlier Task recorded still holds, the change updates
that claim in place, and its own Task names each claim it updated.
When what the earlier Task recorded no longer holds, the claim is left
Refuted: that is drift, and the change fixes the code or records a new
decision instead.

Sources: [TD-0025](../breadcrumbs/TD-0025.md).
