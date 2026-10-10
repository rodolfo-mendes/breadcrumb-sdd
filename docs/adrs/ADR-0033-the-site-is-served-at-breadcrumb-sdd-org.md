---
breadcrumb:
  id: ADR-0033
  type: ADR
  title: The site is served at breadcrumb-sdd.org
  links:
    - amends ADR-0031
---
# ADR-0033: The site is served at breadcrumb-sdd.org

Status: Accepted
Date: 2026-10-10

## Context

ADR-0031 publishes the site to GitHub Pages at
`https://rodolfo-mendes.github.io/breadcrumb-sdd/`, and leaves a
custom domain to a new ADR that amends it. That address depends on
two things the repository does not control: GitHub's own domain, and
the repository's name, so a rename or a move to another host changes
every link to the site.

The product is Breadcrumbs, and the repository keeps the name
`breadcrumb-sdd`, a handle that is unique where the bare word is not.
The maintainer registered `breadcrumb-sdd.org`: `.org` for an open
source project that sells nothing, and the handle so that the domain
matches the repository exactly.

GitHub's documentation (Managing a custom domain for your GitHub
Pages site) says that a site published by a custom workflow needs no
`CNAME` file and ignores one, that the domain should be added in the
repository's settings before DNS points to GitHub, and that a domain
verified in the account's settings cannot be taken over by another
account. It does not say whether the old `github.io` address
redirects to the custom domain.

## Decision

- The site is served at `https://breadcrumb-sdd.org/`, the apex
  domain, with HTTPS enforced. `https://www.breadcrumb-sdd.org/`
  redirects to it.
- The domain is set in the repository's Pages settings. No `CNAME`
  file is committed.
- The domain is verified in the maintainer's GitHub account settings,
  and the verification's TXT record stays in DNS.
- DNS has the four `A` and four `AAAA` records GitHub publishes for
  an apex domain, a `CNAME` from `www` to `rodolfo-mendes.github.io`,
  and no wildcard record.
- `site/hugo.toml` sets `baseURL` to the new address; every other
  link the site writes to itself stays relative (ADR-0031).
- The repository does not rely on the old address redirecting.
  README.md and every link the repository writes name the new one.

## Consequences

- The address no longer changes with the repository's name or its
  host: renaming the repository (issue #34) or leaving GitHub Pages
  changes DNS, not the links.
- The domain is a renewal the maintainer must keep up. If it lapses,
  anyone can buy it and every link to the site points to them; the
  registration renews automatically.
- The domain setting, its verification, DNS and "Enforce HTTPS" are
  settings, not files. No claim checks them; they are checked by hand
  with `gh` and `dig`, as specs/site says.
- Whether the old address redirects is checked once, with `curl -I`,
  after the switch. A link elsewhere to the old address works only if
  it does.
- Google Search Console needs a property for the new domain, verified
  by a DNS record; the property of the old address does not carry
  over.

## Alternatives Considered

- `breadcrumb-sdd.com`: available, but it signals a company and a
  product for sale.
- `breadcrumbsdd.org`, without the hyphen: easier to say, but it also
  reads as "breadcrumbs dd", and it differs from the repository's
  name.
- `www.breadcrumb-sdd.org` as the address, with the apex redirecting
  to it: GitHub supports both; the apex is shorter, and the name has
  no other subdomain to separate from.
- Keeping the `github.io` address: no cost, but the address stays tied
  to the repository's name and to GitHub.
