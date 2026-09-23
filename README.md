# Breadcrumb SDD

Code is the source of truth. Breadcrumb SDD is a method for keeping a verifiable record of the changes made to it.

A breadcrumb records claims about a repository that anyone can check against its files, so a change that drifts from what was agreed becomes visible. Version 0.1.0 defines one breadcrumb, the Task.

## What is here

- [`breadcrumb-sdd.md`](breadcrumb-sdd.md): the specification. It defines the method.
- [`docs/guide.md`](docs/guide.md): the guide. It explains the method and holds no rules; where it and the specification disagree, the specification is right.
- [`tools/`](tools): breadcrumb-kit, the tools that check Tasks. Every check they make can also be done by hand with `grep`, `sha256sum` and `git show`.
  - `verdict.sh`: the verdict of one Task, Confirmed or Refuted.
  - `task.sh`: checks the format of one Task file.
  - `tasks.sh`: checks the format of every Task file in a repository.
- [`scripts/package.sh`](scripts/package.sh): packages the specification and the tools for release.

## Status

Nothing is installed in this repository yet, so the method does not govern it. The specification here is the working copy: a repository is governed only by the copy installed at `.breadcrumb-kit/breadcrumb-sdd.md`.

## Releases

Each release publishes four kinds of files, meant to be copied into `.breadcrumb-kit/` of a repository:

- `breadcrumb-sdd.md`: the specification, under a do-not-edit header that carries the SHA-256 of the source.
- The tools, unchanged.
- `VERSION`: the package version.
- `MANIFEST`: the SHA-256 of each tool and of `VERSION`.

To check an installed kit by hand, from `.breadcrumb-kit/`:

```sh
tail -n +3 breadcrumb-sdd.md | sha256sum   # equals the hash in the header
sha256sum -c MANIFEST
```

The tools need Bash 3.2 or later.

## Checking Tasks

These commands assume the kit is installed in `.breadcrumb-kit/`. Run them with `bash`, since files downloaded from a release may not be executable.

### With the tools

Check the format of every Task file in the repository, from anywhere inside it:

```sh
bash .breadcrumb-kit/tasks.sh      # one line per violation; no output means none
bash .breadcrumb-kit/tasks.sh -v   # also lists the Task files that pass
```

Check the format of one Task file:

```sh
bash .breadcrumb-kit/task.sh breadcrumbs/TASK-0001.task.md   # "Ok", or one "Error:" line per violation
```

Get the verdict of one Task, from the repository root:

```sh
bash .breadcrumb-kit/verdict.sh breadcrumbs/TASK-0001.task.md
```

It prints each claim with its value, then `<id>: Confirmed` or `<id>: Refuted` as the last line. Its exit status is:

| Exit | Meaning |
| --- | --- |
| 0 | Confirmed |
| 1 | Refuted |
| 2 | Not verified: the Task's `status` is not `IMPLEMENTED` or its `spec_version` is not `"0.1.0"` |
| 3 | The Task is verified but breaks the format; the violations go to stderr |
| 4 | Usage or environment error |

Get the verdict of every Task:

```sh
find breadcrumbs -name '*.task.md' | sort | while IFS= read -r t; do
  bash .breadcrumb-kit/verdict.sh "$t" | tail -n 1
done
```

The tools read the files as they are in the working tree. To judge a commit, check that commit out first, for example in a separate worktree:

```sh
git worktree add --detach /tmp/judge <commit>
cd /tmp/judge && bash .breadcrumb-kit/verdict.sh breadcrumbs/TASK-0001.task.md
```

### By hand

Every check can be made without the tools, and against a commit's stored bytes. Replace `<commit>` with a commit, or with an empty string to read the staged files.

1. A Task is verified only if its front matter has `status: IMPLEMENTED` and `spec_version: "0.1.0"`. Any other Task is not judged.
2. A `contains` claim is Confirmed if the file exists and holds `text` exactly:

   ```sh
   git show <commit>:<path> | grep -q -F -e '<text>' && echo Confirmed || echo Refuted
   ```

3. A `sha256` claim is Confirmed if the file's SHA-256 equals `hash`:

   ```sh
   git show <commit>:<path> | sha256sum   # compare with the claim's hash
   ```

4. The Task is Confirmed if every claim is Confirmed, and Refuted otherwise. A Task with no claims is Confirmed.

## Temporary Files

Use the directory /temp to write temporary files that should not be versioned. 
