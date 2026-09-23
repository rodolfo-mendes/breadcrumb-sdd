# Breadcrumb SDD Specification

Version 0.1.0 (2026-09-26)

Editor: Rodolfo Mendes

## Introduction

This is the specification of Breadcrumb SDD, a method for keeping a verifiable record of the changes made to a software repository.

The code in a repository is the source of truth for what the repository does. Neither this specification nor any breadcrumb replaces it. A breadcrumb records claims about the repository that anyone can check against its files, so a change that drifts from what was agreed becomes visible.

This version defines one type of breadcrumb, the Task, and only the rules needed to verify a Task.

This specification defines the method, not the tools that implement it. Where a tool disagrees with this specification, the tool is wrong.

## Notation

The key words MUST, MUST NOT, SHOULD, SHOULD NOT and MAY are to be interpreted as described in RFC 2119 and RFC 8174 when, and only when, they appear in capitals.

This specification judges only what it names. Anything it does not mention is permitted, and only a violation of a MUST or MUST NOT is an error.

A path names a file relative to the repository root. It uses `/` as its separator, does not begin with `/`, and contains no `.` or `..` segment.

A line ends with a line feed (U+000A). A carriage return (U+000D) immediately before the line feed belongs to the line ending, not to the line.

Text is encoded in UTF-8.

### Front matter

Front matter is a block of YAML at the start of a file. It begins with a first line consisting of `---` and ends at the next line consisting of `---`. Everything after that line is the file's body.

Front matter MUST be valid YAML whose top level is a block mapping. The keys this specification defines are written in a subset of YAML:

- Only block mappings, block sequences and scalars are used. Flow collections (`{…}`, `[…]`), block scalars (`|`, `>`), anchors, aliases and tags are not.
- A double-quoted string holds no line break. Its only escape sequences are `\"` for a quotation mark and `\\` for a backslash.

Keys this specification does not define are ignored.

## Terminology

- **Claim**: a predicate about a repository state. See [Claims](#claims).
- **Error**: a violation of a MUST or MUST NOT of this specification.
- **Installed specification**: the copy of this specification that governs a repository. See [Installed specification](#installed-specification).
- **Repository state**: the files of a repository at one commit. See [Repository state](#repository-state).
- **Task**: a record of one change as a set of claims. See [Tasks](#tasks).
- **Task file**: the file of a Task. See [Task files](#task-files).
- **Verdict**: the value of a verified Task. See [Verdicts](#verdicts).

## Repository state

A repository state is the set of files of a repository at one commit. The content of a file in a repository state is its bytes as stored in that commit, before any conversion applied when the file is checked out.

## Installed specification

A repository adopts the method by installing this specification at `.breadcrumb-kit/breadcrumb-sdd.md`. That file is the installed specification.

The installed specification in a repository state governs that state. A repository state with no installed specification is not governed by the method.

## Tasks

A Task records one change to a repository, of any size, as a set of claims about the repository state the change leaves behind.

### Task files

A Task file is a file in the directory `breadcrumbs/` at the repository root, at any depth, whose name ends in `.task.md`. The part of the name before `.task.md` is the Task's id. How ids are chosen is up to the repository.

```
breadcrumbs/TASK-0001.task.md          Task file: id TASK-0001
breadcrumbs/cli/add-save.task.md       Task file: id add-save
breadcrumbs/notes.md                   not a Task file: name does not end in .task.md
src/fix.task.md                        not a Task file: outside breadcrumbs/
```

### Format

A Task file consists of front matter and a body. The front matter defines:

- `status`: the Task's stage, written as a plain scalar. This version defines one value, `IMPLEMENTED`: the change is done.
- `spec_version`: the version of the specification whose rules the Task follows, written as a double-quoted string.
- `claims`: a block sequence of claims, as defined in [Claims](#claims).

The body is Markdown and is free text.

A Task (fictional):

```markdown
---
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "src/cli.py"
    text: "def save(path):"
  - type: "sha256"
    path: "config/defaults.toml"
    hash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
---

# Add a save command to the CLI

The CLI can save the current session to a file.
```

### Verified Tasks

This version verifies a Task file if, and only if, its `status` is `IMPLEMENTED` and its `spec_version` is `"0.1.0"`. It does not judge any other Task file.

A verified Task file MUST follow [Format](#format), and every claim in it MUST follow [Claims](#claims). A verified Task file that does not is an error, and it has no verdict. Otherwise its verdict is determined as in [Verdicts](#verdicts).

## Claims

A claim is a predicate about a repository state. Evaluated against a repository state, it is either **Confirmed**, if the predicate holds, or **Refuted**, if it does not.

A claim states a fact about a state, not a change between states: "`src/cli.py` contains `def save(`", not "`src/cli.py` was modified". The value of a claim depends only on the claim and the repository state.

A claim is a block mapping with the keys `type` and `path`, plus the key its type requires. Every value in a claim is a double-quoted string. `type` is one of the claim types defined below, and `path` is a path.

### Contains claims

Type `contains`. Required key: `text`.

A contains claim is Confirmed if a file exists at `path` and its content holds the UTF-8 encoding of `text` as a contiguous sequence of bytes, compared exactly. Otherwise it is Refuted.

Since `text` holds no line break, a contains claim matches within a single line.

```
type: "contains", path: "src/cli.py", text: "def save(path):"

def save(path):              Confirmed
def Save(path):              Refuted: comparison is exact, case included
(no file at src/cli.py)      Refuted
```

### Hash claims

Type `sha256`. Required key: `hash`, written as 64 lowercase hexadecimal digits.

A hash claim is Confirmed if a file exists at `path` and the SHA-256 of its content equals `hash`. Otherwise it is Refuted.

## Verdicts

The verdict of a verified Task against a repository state is Confirmed if every one of its claims is Confirmed, and Refuted otherwise.