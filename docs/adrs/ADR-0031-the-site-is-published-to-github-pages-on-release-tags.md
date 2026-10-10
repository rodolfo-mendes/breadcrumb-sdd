---
breadcrumb:
  id: ADR-0031
  type: ADR
  links:
    - constrained_by ADR-0030
---
# ADR-0031: The site is published to GitHub Pages by a workflow on release tags

Status: Accepted
Date: 2026-10-10

## Context

ADR-0030 settles how the site is generated. This ADR settles where it
is published and when.

The repository is public, so GitHub Pages costs nothing. When the old
site was deleted, GitHub Pages was turned off (ADR-0024); it is off
now. Pages can serve a branch's root or its `docs/` folder, and
`docs/` holds the ADRs and the user documentation, not a site. A
workflow can deploy any folder it builds.

The documentation page is `docs/bcr.md`, which on the main branch
describes the `bcr` of the next release, not necessarily the last one.
The repository keeps its name, so the site lives under the path
`/breadcrumb-sdd/`.

## Decision

- The site is published to GitHub Pages at
  `https://rodolfo-mendes.github.io/breadcrumb-sdd/`, with the
  repository's Pages source set to GitHub Actions.
- `.github/workflows/pages.yml` builds `site/` and deploys it with
  `actions/upload-pages-artifact` and `actions/deploy-pages`. Like the
  other workflows, it pins every action to a commit.
- It runs when a tag `v*` is pushed, and on `workflow_dispatch`, so a
  fix to the site's text does not wait for a release.
- It deploys only from `main` or from a tag `v*`. The `github-pages`
  environment allows no other ref, so a manual run cannot publish
  another branch.
- `site/hugo.toml` sets `baseURL` to the address above, and every link
  the site writes to itself is relative.
- The built site shows the tag or commit it was built from.
- A custom domain is not part of this decision.

## Consequences

- The documentation online is that of the last release, unless
  someone runs the workflow by hand; a manual run is a deliberate
  publication.
- The Pages source and the environment's rules are settings, not
  files. No claim checks them; they are checked by hand with `gh`.
- Renaming the repository (issue #34) changes the address, `baseURL`
  and the claim that holds it. This ADR does not rely on GitHub
  redirecting the old address.
- A custom domain comes through a new ADR that amends this one, when
  one is wanted.
- GitHub Pages is a service the repository does not control. The
  built site is plain files, and opens from a folder with no server.

## Alternatives Considered

- Branch `main`, folder `/docs`: no workflow, but it publishes the
  ADRs and the documentation as they are, with no menu and no landing
  page.
- A `gh-pages` branch: Pages serves it, but a generated branch has to
  be kept and its history grows with every build.
- Deploying on every push to `main`: the site is always current, but
  its documentation would describe commands no release has.
- Another host, such as Netlify: a second account and service, for a
  site that GitHub already serves next to the repository.
