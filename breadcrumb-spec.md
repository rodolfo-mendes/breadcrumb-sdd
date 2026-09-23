# Breadcrumb-SDD — Specification
 
This document is the specification of Breadcrumb-SDD and the source of truth
for its intent. It lives in `breadcrumb-spec.md` at the root of the repository and also
governs that repository: only rules written in this repository bind it.
 
It starts small on purpose. Every other part of the method enters later,
through the rules below.
 
## 1. Commits and validity
 
The method judges the commits on the main branch. In this document a
**commit** is one of them. A commit's **parent** is the previous commit on
main (for a merge, its first parent).
 
A **change** is one file that a commit adds, modifies or deletes, compared
with its parent. A rename is a deletion and an addition. A commit may make
any number of changes, including none.
 
The **singularity** is the first commit that contains a Vision (section 2).
It is valid by definition.
 
Every later commit is **valid** if it satisfies the rules of this document
as they stood in its parent. A rule that a commit introduces binds from the
next commit on: a commit can neither benefit from a rule it adds nor be
condemned by it. This holds for commits to this document too.
 
Validity is a property of commits, not of files. Whatever the singularity
contains is where the method starts; it is neither valid nor invalid.
 
Only what can be checked from the repository, its history and its files,
decides validity. Anything else may be recommended elsewhere, but it is not a
rule here.
 
An invalid commit stays in the history, and every commit after it inherits
the break. How a repository recovers from one is not yet defined.
 
## 2. Vision
 
The Vision says what we are building and what we are optimizing it for. It
lives in `VISION.md` at the root of the repository and is accepted by being
committed.
 
The Vision is ground, not a link: no artifact is required to point at it.
 
Checks:
 
- The singularity contains `VISION.md`.
- No commit deletes `VISION.md`.
## 3. Intake
 
An Intake records a need or an observation, in words the committer stands
behind. It is not a specification. Each Intake is one file directly in
`intakes/`, named by its id: `INT-YYYY-NNNN.md`, where YYYY and NNNN are
four digits each. An Intake is never empty: it holds at least one character
other than whitespace.
 
Every commit after the singularity enters through exactly one Intake, which
it adds. It names that Intake with a line in its commit message that reads
exactly `Intake: INT-YYYY-NNNN`, and its message contains no other text of
the form `INT-YYYY-NNNN`.
 
An Intake never lands alone: the commit also adds or modifies at least one
file outside `intakes/`. A deletion alone does not count. Once committed, an
Intake is never modified or deleted. A later Intake may reference an earlier
one, and that is the only relation between Intakes.
 
Checks, for every commit after the singularity:
 
- Inside `intakes/`, the commit makes exactly one change: it adds an Intake
  that is not empty.
- The commit message has a line that reads exactly `Intake: <id>`, where
  `<id>` is that Intake's id, and contains no other text of the form
  `INT-YYYY-NNNN`.
- The commit adds or modifies at least one file outside `intakes/`.