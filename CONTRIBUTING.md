# Contributing

These conventions govern contributions to the `breadcrumb-sdd` repository.
They are not part of the Breadcrumb SDD method, which prescribes nothing
about how a repository uses Git.

## Commit messages

Commits on `main` follow [Conventional Commits](https://www.conventionalcommits.org/).

Breadcrumbs already record the context of a change, so messages don't
repeat it. They don't list changed files, and they don't link to
breadcrumbs.

### Type

Allowed types: `feat`, `fix`, `docs`.

A new type is added to this document before it is used.

### Scope

Optional, but encouraged. Defined scopes: `spec`, `tools`.

### Description

Mandatory. One phrase in the imperative mood that completes the sentence
"If applied, this commit will…". Write `add the claim verifier`, not
`added the claim verifier` or `adding the claim verifier`.

### Body

Mandatory. A single paragraph after a blank line. It records the judgment
behind the commit: what was decided, and what was deliberately left out.
It doesn't restate the need, and it doesn't list files.

### Example

```
feat(tools): report only violations by default

A clean run should read as silence, so successes appear only with -v.
Grouping violations by breadcrumb type was left out until the report
has more than one type to show.
```

## Feature branches

Commit messages on feature branches are free form.

## Pull requests

- **Merge:** into `main` by squash only.
- **Title:** follows the description format, including type and scope.
- **Description:** follows the body rule and holds nothing else, because
  it becomes the commit message on `main`.

For maintainers: squash merging is the only merge method enabled, and its
default commit message is set to "Pull request title and description".