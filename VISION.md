# The vision of Breadcrumb

Who Breadcrumb is for, what it believes, and which way to lean when no
rule decides. It follows ASDLC's Product Vision pattern. It decides
nothing itself: a rule in force lives in an ADR, a spec or
`AGENTS.md`. Read it when a spec leaves a choice open.

## Who it is for

People who build software with AI agents writing most of the code,
first a solo builder or a small team. Months later they need to know
whether the code still does what was asked, and which decision a piece
of code was meant to carry out. They will not keep up a heavy process
to find out.

## Point of view

Intent travels from whoever asks for the software to the code that
runs through a chain of people, documents and agents, and like any
channel it gathers noise the longer it runs. Trying to make that
chain driftless is the failure Breadcrumb starts from: a specification
narrow enough to pin the code down starts to become code, and one
broad enough to stay readable drifts like any other document.
Breadcrumb lets drift happen and makes it visible, after the fact.

Six principles follow. Each is a direction Breadcrumb keeps moving in,
never a line it has reached; when two collide, the case decides, and
the decision says which one lost.

1. **Code is the source of truth.** Code cannot lie about what a
   repository does; a breadcrumb can. Code is already the most
   detailed specification there is, the only one precise enough to
   drive a computer, so a breadcrumb describes it and never replaces
   it. What a breadcrumb holds is what code cannot: why.
2. **Simplicity.** Fewer concepts. An artifact, a field or a rule is
   added only when it is needed.
3. **Flexibility.** A small, linked breadcrumb recombines with what a
   repository already has, without rebuilding what is there. A
   repository traces as much or as little as its owners choose, and
   when; an untraced file sits outside the record, like legacy code,
   not a violation.
4. **Self-containment.** Breadcrumb, and a repository that uses it,
   depend on nothing they do not control. What they do not control can
   change under them, a commit hash under a rebase or a host framework
   under a release. Not tying Breadcrumb to one stack, operating system
   or framework also widens who can use it.
5. **Verifiability.** A claim nobody can check against the repository
   reads as evidence but is decoration. The more one state of the
   repository can check, the less is left to trust.
6. **Detection, not prevention.** A block hides the judgment call
   where it is easiest to abuse, while a visible break can be audited
   later. The goal is not zero discretion, but discretion that shows.

Adopting Breadcrumb should cost as little as possible: one binary and
a few lines of front matter, with no service to run. A command meant
for agents should answer their question more cheaply than searching
the files as plain text would; that is how such a command is judged.

## Decision heuristics

When two options are both valid, lean this way. Each names a case
where it decided.

- **Earn the distinction.** Add a concept, a field or a command only
  if it changes what a person or an agent does. For a command meant
  for agents, count what agents will not compose from shell tools, not
  only what shell tools cannot do. Applied: no check command, since the
  pipe already makes one; `bcr filter`, which agents did not build
  from `grep` and `awk` on their own.
- **Prefer the primitive the other can be derived from.** Of two
  designs, choose the one whose output can produce the other's; the
  other direction usually loses something no script restores. Applied:
  tagged records rather than a command per question
  ([ADR-0011](docs/adrs/ADR-0011-breadcrumbs-printed-as-tagged-records.md)),
  and `bcr filter` rather than a `show` command.
- **Name the expiry.** When something is deferred, name the condition
  that brings it back, so crossing it shows. Something the record can
  rebuild at any time can wait for that need. Applied: every row of
  "Deferred, and what brings it back" in
  [ARCHITECTURE.md](ARCHITECTURE.md).
- **Prefer the false alarm.** When choosing how a check fails, take
  the false positive over the missed break. When a check stays silent
  on purpose, say so. Applied: a claim guards this repository's
  `breadcrumb.rules`, so deleting it turns the pipe red
  ([`specs/asdlc`](specs/asdlc/spec.md)). Bent: a file no
  `.breadcrumbs` pattern names drops out without a sign
  ([ADR-0014](docs/adrs/ADR-0014-breadcrumbs-file-names-the-files-to-read.md)).
- **Settle with a case.** When a decision keeps being amended, or
  questions open faster than they close, write a real instance instead
  of another decision. Applied: this repository's shape was declared
  from the links already written
  ([ADR-0026](docs/adrs/ADR-0026-this-repository-uses-the-asdlc-layout.md)).
