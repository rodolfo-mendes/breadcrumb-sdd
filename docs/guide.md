# Breadcrumb SDD Guide

Companion to the Breadcrumb SDD Specification, version 0.1.0.

## About this guide

This guide explains how the method works and why. It holds no rules. Where it and the specification disagree, the specification is right and this guide is out of date.

The specification contains only the rules needed for its current goal. This guide describes the rest: the ideas behind those rules, the mechanics planned for later versions, and advice for writing breadcrumbs. Anything described here as planned becomes binding only when a version of the specification includes it.

This guide is not installed into repositories. It lives beside the specification in the Breadcrumb SDD repository.

## Principles

### The method detects; it does not prevent

The specification is not a compiler. A compiler refuses to build a program with a syntax error. The method builds nothing and refuses nothing: breadcrumbs that break the rules stay in the repository, and the specification makes the break visible.

The specification has two purposes: to give shape to the record of changes, and to make the errors that break that record visible.

### The specification judges only what it names

If the specification says nothing about something, it is permitted. A Task file may carry extra front matter keys, pictures, notes, or anything else the specification does not mention. Only a violation of a MUST or MUST NOT is an error.

A consequence: every silence in the specification is a decision. When a version says nothing about a behaviour, that behaviour is allowed until a later version says otherwise.

### Each version holds only what its goal needs

Each version of the specification is written backwards from a goal: it contains the rules the goal needs, plus the conditions under which each of those rules fails. A rule that only says when something passes cannot detect anything, so its failure half is part of the minimum.

The goal of version 0.1.0 is to verify the two Tasks that record the install of the specification and the kit. Everything else in this guide waits for a goal that needs it.

### Only auditable breadcrumbs are verified

Work in progress needs room to move. A Task may be drafted, approved, sent back for changes, and approved again as often as developers and agents need, and a new specification may be installed in between. None of that is checked. Verification applies only to breadcrumbs that record finished work.

### The method, not the tools

The specification defines the method and prescribes no tool. The tools shipped with the kit exist so a user can check the breadcrumbs quickly. Every check they perform can also be done by hand, with standard commands such as `grep`, `sha256sum` and `git show`. If you removed every tool, the method would still work.

### The installed copy governs; the source is a working copy

A repository is governed by the specification installed at `.breadcrumb-kit/breadcrumb-sdd.md`, not by any other copy. In the Breadcrumb SDD repository itself, the specification at the root is the working copy: editing it changes nothing until it is released and installed. The rules change only when a new specification is installed, and that install is itself recorded.

## What rules can detect

Rules differ in the evidence needed to catch a violation:

| Class | Evidence | Examples |
| --- | --- | --- |
| **State rules** | One repository state | A verified Task's format; the value of its claims. |
| **History rules** | Two or more states | An approved Task was not edited; a Task was not deleted. |
| **Unverifiable rules** | None in the repository | A Task describes one change; an approval reflects real agreement. |

History rules are weaker than they look. Rewriting history (squash, rebase, force push) erases their evidence, and a commit records only its end result, not the steps taken between commits.

So the promise runs one way. If the rules are followed, the record is well-formed. A record with no visible error does not prove the rules were followed.

A **seal** turns a history rule into a state rule: it moves evidence of a past event into the present state. A hash claim is a seal. So is the install header that packaging writes at the top of the installed specification.

## Task lifecycle (planned)

Version 0.1.0 defines one status, `IMPLEMENTED`. The lifecycle below is planned for a later version.

A Task moves through three stages:

- **`PROPOSED`**: a draft. It may be edited freely.
- **`APPROVED`**: its content is agreed. It is not edited.
- **`IMPLEMENTED`**: the change is done. It is never edited again, and its status never changes.

A Task moves from `PROPOSED` to `APPROVED`, from `APPROVED` back to `PROPOSED`, and from `APPROVED` to `IMPLEMENTED`. It may move between `PROPOSED` and `APPROVED` any number of times.

An approved Task that needs changing goes back to `PROPOSED`, is edited, and is approved again before it is implemented.

`PROPOSED` and `APPROVED` Tasks are **staged**. `IMPLEMENTED` Tasks are **auditable**. Only auditable Tasks are verified.

Whether a Task was edited after approval is a history rule. A planned **approval seal**, a hash of the Task's content recorded when it is approved, would make it a state rule: an implemented Task whose content no longer matches its seal was edited after approval.

## `spec_version`

Every Task declares the version of the specification whose rules it follows. This is not the version installed when the Task was written: a Task drafted before an upgrade may already follow the next version.

The field exists because implemented Tasks never change while the specification does. Without it, every future version would have to read every Task ever written in exactly the same way. With it, a later version can tell which rules each Task follows, and read old Tasks by their own rules.

It had to be required from the first version. Added later, a missing field would be ambiguous: an old Task written before the field existed, or a new Task whose author forgot it.

## Upgrading the installed specification

Suppose version 0.1.0 is installed and a Task records that install with a hash of the installed file. Version 0.2.0 is released.

1. A Task that upgrades the installed specification is proposed. It declares `spec_version: "0.2.0"`, although 0.1.0 is still installed.
2. The Task is approved. Staged Tasks are not verified, so declaring a version that is not installed yet is not a problem.
3. Version 0.2.0 is installed and the Task becomes `IMPLEMENTED`.
4. The Task that recorded the 0.1.0 install is now Refuted: the installed file changed, so its hash no longer matches.

That last Refuted Task is correct. It records a change that no longer holds. A planned **supersession** mechanism will let the upgrade Task declare that it replaces the earlier one, so the record shows the replacement instead of a failure.

## The audit graph

The auditable breadcrumbs of a repository state, and the links between them, form the **audit graph**: the verified record that an auditor reads.

In version 0.1.0 its vertices are the verified Tasks, and it has no edges. Planned breadcrumb types (Intake, Requirement, Technical Decision) and a `parents` field on Tasks will add the links.

The graph holds no identifier that the method does not control, such as a commit hash. Rebases and squashes can change commit hashes; they cannot change the graph.

## Writing Tasks

### Choosing ids

The specification lets each repository choose its own ids. These conventions avoid trouble:

- Use only ASCII letters, digits, `-` and `_`, starting with a letter or digit. Many shell commands read a leading `-` as an option.
- Keep ids unique regardless of case. On macOS and Windows, `Fix.task.md` and `fix.task.md` are the same file.
- Do not rename Task files. The file name is the id, so a rename creates a new Task and orphans the old id.

### Writing claims

- **Claim whole lines.** A `contains` claim matches anywhere in the file. To claim a version line, claim the full line (`"Version 0.1.0 (2026-09-26)"`), not a fragment that could appear elsewhere.
- **Give every Task at least one claim.** In version 0.1.0 a Task with no claims is Confirmed, because every one of its zero claims is.
- **Check `status` and `spec_version` spelling.** In version 0.1.0 a Task with a misspelled status or version is not verified at all.
- **Hash claims on text files.** Claims read files as stored in the commit, not as checked out. If your repository converts line endings on checkout, verify against the committed bytes (`git show <commit>:<path> | sha256sum`) or disable the conversion in `.gitattributes`.

## Known limitations of version 0.1.0

- A Task with a misspelled `status` or `spec_version` escapes verification.
- A verified Task with no claims is Confirmed.
- Deleting an implemented Task is permitted: the specification is silent on it.
- Rules about how Tasks change over time are not in the specification yet.
- The install header is written by packaging, not defined by the specification. Whether the specification should define it is pending a requirement and a decision.

## Glossary

Terms used in this guide but not defined by the specification.

- **Approval seal**: a planned hash of a Task's content, recorded at approval.
- **Audit graph**: the auditable breadcrumbs of a repository state and the links between them.
- **Auditable breadcrumb**: a breadcrumb that belongs to the audit graph; an `IMPLEMENTED` Task.
- **Breadcrumb**: a file that records part of a repository's history of intent. Version 0.1.0 defines one type, the Task.
- **Install header**: the do-not-edit notice that packaging writes at the top of the installed specification.
- **Seal**: a hash that lets one repository state show that something has not changed since the hash was taken.
- **Staged breadcrumb**: a breadcrumb that is not auditable; a `PROPOSED` or `APPROVED` Task.
- **Supersession**: a planned way for a Task to declare that it replaces an earlier one.
- **Working copy**: in the Breadcrumb SDD repository, the specification at the root, as opposed to the installed copy that governs the repository.