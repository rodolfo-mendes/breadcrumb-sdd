# Breadcrumb-SDD — Specification

This document is the specification of Breadcrumb-SDD and the source of truth
for its intent. It lives in `breadcrumb-spec.md` at the root of the repository and also
governs that repository: only rules written in this repository bind it.

It starts small on purpose. Every other part of the method enters later,
through the rules below.

## 1. Changes and validity

A **change** is one commit on the main branch, however many files it touches.
A change's **parent** is the previous commit on main (for a merge, its first
parent).

The **singularity** is the first change that contains a Vision (section 2). It
is valid by definition.

Every later change is **valid** if it satisfies the rules of this document as
they stood in its parent. A rule that a change introduces binds from the next
change on: a change can neither benefit from a rule it adds nor be condemned
by it. This holds for changes to this document too.

Validity is a property of changes, not of files. Whatever the singularity
contains is where the method starts; it is neither valid nor invalid.

Only what can be checked from the repository, its history and its files,
decides validity. Anything else may be recommended elsewhere, but it is not a
rule here.

An invalid change stays in the history, and every change after it inherits
the break. How a repository recovers from one is not yet defined.

## 2. Vision

The Vision says what we are building and what we are optimizing it for, in one
or two paragraphs. It lives in `VISION.md` at the root of the repository and is
accepted by being committed.

The Vision is ground, not a link: no artifact is required to point at it.

Checks:

- The singularity contains `VISION.md`.
- A change that alters `VISION.md` leaves it at one or two paragraphs.

## 3. Intake

An Intake records a need or an observation, in words the committer stands
behind. It is not a specification. Each Intake is one file in `intakes/`,
named by its id: `INT-YYYY-NNNN.md`.

Every change after the singularity enters through an Intake. The change either
adds one or names one already in the repository. An Intake never lands alone:
a change that adds Intakes also changes something else. Once committed, an
Intake is never edited or removed. A later Intake may reference an earlier
one, and that is the only relation between Intakes.

A change names its Intakes in its commit message, one per line:
`Intake: INT-YYYY-NNNN`.

Checks, for every change after the singularity:

- The commit message names at least one Intake.
- Every named Intake exists in the change's files.
- Every Intake the change adds is named in its commit message.
- A change that adds an Intake also changes a file outside `intakes/`.
- No existing file in `intakes/` is modified or deleted.
