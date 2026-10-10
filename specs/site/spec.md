---
breadcrumb:
  id: site
  type: spec
  links:
    - constrained_by ADR-0008
    - constrained_by ADR-0030
    - constrained_by ADR-0031
  claims:
    - 'site/content/_index.md has-line # Breadcrumb {#top}'
    - 'site/content/_index.md has-line ## The problem {#problem}'
    - 'site/content/_index.md has-line ## How it works {#how-it-works}'
    - 'site/content/_index.md has-line ## The report {#report}'
    - 'site/content/_index.md has-line ## Honest scope {#scope}'
    - 'site/content/_index.md has-line ## Try it {#try-it}'
    - 'site/content/_index.md has-line ## Who it''s for {#audience}'
    - 'site/content/_index.md has-line ## Status {#status}'
    - 'site/hugo.toml has-line baseURL = "https://rodolfo-mendes.github.io/breadcrumb-sdd/"'
    - '.github/workflows/pages.yml has-line HUGO_VERSION: 0.167.0'
    - '.gitignore has-line /site/content/docs/_index.md'
---
# Feature: site

The site is Breadcrumb's entry point on the web: a landing page that
pitches the method and `bcr` to possible adopters, and a documentation
page that hosts `docs/bcr.md`.

## Blueprint

### Context

Someone who has not cloned the repository has nowhere to learn what
Breadcrumb is, or to read how `bcr` is called. The site gives them
both, and it must be found by a search. Its primary goal is to pitch
the method and the tools; its secondary goal is to host the
documentation.

The site says what `README.md` and `docs/bcr.md` say, and no more. A
claim on it can be checked against the repository, by hand, without
`bcr`. It does not mention how the maintainer works, only Breadcrumb.

How the site is generated is ADR-0030; where and when it is published
is ADR-0031. This spec says what the site is, and what must stay true
of it.

### Architecture

```
site/
  hugo.toml              baseURL, the menus, the language
  content/
    _index.md            the landing page: the hero and seven sections
    docs/_index.md       docs/bcr.md, copied when the site is built
  layouts/               the templates, with no theme
  assets/                one stylesheet
  static/                the GitHub mark, as an SVG, and the pictures
```

`site/content/docs/_index.md` is not committed (`.gitignore`). The
workflow copies `docs/bcr.md` to it before the build, and nothing else
enters `site/content/`.

**Pages.** The landing page is `/`. The documentation page is
`/docs/`, `docs/bcr.md` as written: its headings keep the ids Hugo
derives from their text, such as `#extract` for `### extract`.

**The landing page.** The hero is the first heading, `# Breadcrumb
{#top}`, with the name and one sentence that says what Breadcrumb
does. Under it, a link to the repository, drawn as the GitHub mark.
Seven sections follow, in this order; each id is written in the
Markdown:

| Menu label | Id | Says |
|---|---|---|
| The problem | `problem` | Code says what a repository does, not why; the why drifts |
| How it works | `how-it-works` | One breadcrumb, its claims, and the three verdicts |
| The report | `report` | What `bcr report` shows, a picture of one of its diagrams, and a link to its documentation |
| Honest scope | `scope` | Drift is made visible after the fact, not prevented |
| Try it | `try-it` | The pipe on one file, with links into the documentation page |
| Who it's for | `audience` | People who build with agents writing most of the code |
| Status | `status` | The released version and what is deferred |

**The menu.** A menu lists the seven sections, in page order, as entries
of `site/hugo.toml` whose `url` is `#` and the section's id. Above
them are links to other pages: Docs, Repository (the repository's
page on GitHub) and Releases. On a narrow screen the menu is collapsed
behind one button, made of a `<details>` element, so it needs no
script. There is no highlighting of the section being read.

**Building.** The workflow (ADR-0031), with `PUBLIC` the folder it
builds into:

1. Copy `docs/bcr.md` to `site/content/docs/_index.md`.
2. Build with `hugo --source site --destination "$PUBLIC"
   --panicOnWarning`, with `HUGO_SITE_VERSION` set to the tag and the
   short commit. Only variables that start with `HUGO_` reach a
   template.
3. Run the checks of the Regression Guardrails below.
4. Upload `PUBLIC` and deploy it.

Hugo is the release named by `HUGO_VERSION` in the workflow, the
standard edition, downloaded from Hugo's releases and checked against
the checksums published with it.

### Constraints

- **Copy.** The landing page never says Breadcrumb prevents drift, or
  keeps documents in sync with code. A `has-line` claim is a text
  match: the page says that Confirmed means a line is there, not that
  the document is right. Its text comes from `README.md`, `VISION.md`
  and `docs/bcr.md`, not from memory.
- **Found by a search.** Each page has a `<title>`, a description,
  one `<h1>` and a canonical address. The site has a `sitemap.xml`,
  and a `robots.txt` that names it.
- **Self-contained.** Every link the site writes to itself is
  relative. The only absolute addresses are the repository's, its
  releases and the site's own. No page loads a script, a font, a style
  or an image from elsewhere.
- **Languages.** The text is English. It lives in `site/content/`,
  with the language set in `site/hugo.toml`, so another language can
  be added as content without moving the layouts.
- **What the claims cannot hold.** The copied documentation and the
  built site are not in the repository, so no claim is about them; the
  Guardrails below check them. A `has-line` claim cannot say that an id
  is unique or that the menu is in page order, so a Guardrail does.
- **Pictures.** A picture illustrates; it does not have to show the
  state of the repository or the example of "Try it". It is real
  output of `bcr report`, never drawn by hand, and its caption says
  where it comes from. It is in `site/static/`, with alt text that
  says what it shows.
- **Text is not claimed.** The hero and the sections' text change
  often and carry no claim.

## Contract

### Definition of Done

- [ ] `site/` holds the files of the Architecture, and `hugo` builds it
      with `--panicOnWarning` and exits 0.
- [ ] The landing page has the hero and the seven sections, each with
      the id of the table, and the menu lists them in the same order.
- [ ] The menu opens and closes on a narrow screen with no script.
- [ ] `/docs/` is `docs/bcr.md`, and the landing page's links into it
      resolve.
- [ ] `.github/workflows/pages.yml` runs on a tag `v*` and on
      `workflow_dispatch`, runs the checks before it uploads, and pins
      Hugo and every action.
- [ ] The built pages show the tag or commit they were built from.
- [ ] The repository's Pages source is GitHub Actions, and the
      `github-pages` environment allows only `main` and tags `v*`.
- [ ] `AGENTS.md` names `site/` in its Context Map, and its Toolchain
      says how to preview the site.
- [ ] Each claim in the front matter holds.

### Regression Guardrails

Each is a command whose output is empty when the guardrail holds,
unless it says what it prints. The workflow runs them before it uploads; a person can run the same ones by
hand, with `PUBLIC` the folder of a build.

- `site/content/` holds the landing page and the copy of the
  documentation, and nothing else. Checked by hand:
  `find site/content -type f | sort` prints `site/content/_index.md`
  and `site/content/docs/_index.md`.

- The copy is the documentation. Checked by hand:
  `cmp docs/bcr.md site/content/docs/_index.md`.

- The menu lists the sections' ids, in page order. Checked by hand:

  ```
  diff <(grep -o '{#[a-z-]*}' site/content/_index.md | grep -v '{#top}' | tr -d '{}#') \
       <(grep -o 'url = "#[a-z-]*"' site/hugo.toml | sed 's/^url = "#//; s/"$//')
  ```

- No id is written twice, and each is on the built page once. Checked
  by hand:

  ```
  grep -o '{#[a-z-]*}' site/content/_index.md | sort | uniq -d
  for id in $(grep -o '{#[a-z-]*}' site/content/_index.md | tr -d '{}#'); do
    test "$(grep -o "id=\"$id\"" "$PUBLIC/index.html" | wc -l)" -eq 1 || echo "$id"
  done
  ```

- Each link from the landing page into the documentation reaches a
  heading. Checked by hand:

  ```
  for a in $(grep -oE 'href="docs/#[a-z0-9-]+"' "$PUBLIC/index.html" | sed 's/.*#//; s/"$//'); do
    grep -q "id=\"$a\"" "$PUBLIC/docs/index.html" || echo "$a"
  done
  ```

- Each picture the landing page shows is in `site/static/`, and has
  alt text. Checked by hand:

  ```
  grep -oE '!\[[^]]*\]\([^)]+\)' site/content/_index.md | sed 's/.*(//; s/)$//' | while read -r f; do
    test -f "site/static/$f" || echo "$f"
  done
  grep -oE '!\[\]\(' site/content/_index.md
  ```

- No link outside the repository and the site, no root-relative link,
  and no script, import or `url(`. Checked by hand:

  ```
  grep -rhoE '(src|href)="(https?:)?//[^"]*"' "$PUBLIC" \
    | grep -vE '"https://(github\.com/rodolfo-mendes/breadcrumb-sdd|rodolfo-mendes\.github\.io/breadcrumb-sdd/)'
  grep -rhoE '(src|href)="/[^/"][^"]*"' "$PUBLIC"
  grep -rlE '<script|@import|url\(' "$PUBLIC"
  ```

- Each page has a title, a description, a canonical address and one
  `<h1>`; the sitemap lists both pages; `robots.txt` names it. Checked
  by hand:

  ```
  for p in '<title>' '<meta name="description"' 'rel="canonical"'; do
    grep -L "$p" "$PUBLIC/index.html" "$PUBLIC/docs/index.html"
  done
  grep -c '<h1' "$PUBLIC/index.html" "$PUBLIC/docs/index.html"
  grep -o '<loc>[^<]*' "$PUBLIC/sitemap.xml"
  grep -x 'Sitemap: https://rodolfo-mendes.github.io/breadcrumb-sdd/sitemap.xml' "$PUBLIC/robots.txt"
  ```

  The first loop prints nothing, the count is 1 for each page, the
  sitemap lists the two addresses, and the last line matches.

- The built landing page shows its version. Checked by hand:
  `grep -F "$HUGO_SITE_VERSION" "$PUBLIC/index.html"` matches.

- The landing page never says Breadcrumb prevents drift. This is a
  practice, not a rule: `grep -n -i -E 'prevent|in sync'
  site/content/_index.md` lists the lines to read, and a person judges
  each.

- The Pages source and the environment's rules are the ones of
  ADR-0031. Checked by hand:
  `gh api repos/rodolfo-mendes/breadcrumb-sdd/pages --jq .build_type`
  prints `workflow`, and
  `gh api repos/rodolfo-mendes/breadcrumb-sdd/environments/github-pages/deployment-branch-policies --jq '.branch_policies[].name'`
  prints `main` and `v*`.

### Scenarios

```gherkin
Scenario: A section is renamed
  Given the id "problem" in site/content/_index.md becomes "problems"
  When the pipe runs
  Then the claim on "## The problem {#problem}" is Refuted
  And the menu check prints the two ids that differ

Scenario: Hugo is bumped
  Given HUGO_VERSION in pages.yml names a newer release
  When the pipe runs
  Then the claim on the old line is Refuted until this spec names the new one
  And the change comes with a PBI that changes this spec

Scenario: A typo in the landing page
  Given a misspelled word in a section's text
  When it is fixed in site/content/_index.md
  Then no PBI is needed, and the commit type is docs
  And a manual run of the workflow publishes it

Scenario: A manual run from another branch
  Given a person runs the workflow on a branch that is not main
  When the deploy job starts
  Then the github-pages environment refuses it
  And nothing is published

Scenario: A link that points at the site's root
  Given a layout writes href="/docs/"
  When the checks run
  Then the root-relative check prints that link and the workflow fails

Scenario: A picture that is not in site/static
  Given the landing page shows report-filter.png
  And site/static has no such file
  When the checks run
  Then the picture check prints report-filter.png and the workflow fails
```
