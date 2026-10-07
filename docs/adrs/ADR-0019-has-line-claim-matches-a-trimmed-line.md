---
breadcrumb:
  id: ADR-0019
  type: ADR
  links:
    - follows ADR-0017
---
# ADR-0019: A has-line claim matches a trimmed whole line

Status: Accepted
Date: 2026-10-07

## Context

ADR-0017 says how a claim is written, and leaves each kind of claim to
its own ADR. The first kind has to express the claims this repository
already writes in its specs: that a file holds a given line, such as
a heading in `docs/bcr.md` or an ADR's status.

All five of those claims match a whole line of their file. A match on
part of a line would confirm claims the files no longer support:

- `Status: Accepted` would still match ADR-0015, whose line is now
  `Status: Accepted, amended by ADR-0016`.
- `### verify` would still match after the heading moved to
  `#### verify`.
- A claim about the file it is written in would always match its own
  entry in the front matter.

A match on the exact line cannot claim indented code, since a claim's
text holds no tab.

## Decision

- The first kind of claim is `has-line`:

  ```yaml
  claims:
    - 'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'
  ```

- Its argument is a text that is not empty and does not start or end
  with a space or a tab. Any other argument is a problem.
- A `has-line` claim holds when at least one line of its target,
  without its line ending (`\n` or `\r\n`) and without the spaces and
  tabs at its start and end, is equal to the text, byte for byte,
  letter case included. Any line counts, the front matter's included.
- A claim whose target does not exist does not hold.

## Consequences

- A claim can quote a line of code at any indentation, and does not
  break when a file is saved with Windows line endings.
- Any edit to the claimed line breaks the claim, words added at its
  end included. That is the point of a claim, and its cost.
- A claim can be checked by hand with the standard tools:

  ```
  sed 's/\r$//; s/^[[:blank:]]*//; s/[[:blank:]]*$//' TARGET | grep -xF -- 'TEXT' > /dev/null
  ```

  `grep -q` is not used: it stops at the first match, so under
  `set -o pipefail` a long file would read as not holding.

- `has-line` cannot claim that a text appears inside a line, or how
  many times a line appears. Those would be other kinds.
- The old model's `contains` claim, which `bcr check` and
  `bcr verdict` read in `breadcrumbs/`, is a different thing in a
  different model; the new model has no kind of that name.

## Alternatives Considered

- `contains`, a match on part of a line: forgiving when words are
  added to a line, but it confirms the three cases above, and its
  name would read as a substring match forever (ADR-0020).
- A match on the exact line, spaces included: no rule about trimming,
  but no claim about indented code.
- Ignoring letter case, or comparing Unicode text by its normal form:
  closer to what a reader sees, but the tool and a reviewer using
  `grep` could disagree.
