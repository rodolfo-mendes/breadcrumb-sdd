# Breadcrumb report

## Breadcrumbs

Breadcrumbs: 5. Refuted: 1. Undecided: 2. Confirmed: 1. Not audited: 1.

| Verdict | Breadcrumb | Type | File |
|---|---|---|---|
| Refuted | verify | spec | specs/verify/spec.md:3 |
| Undecided, no claims | ADR-0015 | ADR | docs/adrs/ADR-0015.md:3 |
| Undecided, no claims | PBI-00006 | PBI | tasks/PBI-00006.md:3 |
| Confirmed | ADR-0004 | ADR | docs/adrs/ADR-0004.md:3 |
| not audited | audit | spec | specs/audit/spec.md:3 |

## Problems

- docs/adrs/ADR-0099.md:2: breadcrumb has no id
- specs/verify/spec.md:11: claim "docs/bcr.md" must be a target, a kind and an argument

## Claims

| Verdict | Breadcrumb | Claim | File |
|---|---|---|---|
| Refuted | verify | docs/bcr.md has-line \#\#\# check | specs/verify/spec.md:9 |
| Confirmed | verify | docs/bcr.md has-line \#\#\# verify | specs/verify/spec.md:8 |
| not audited | verify | docs/bcr.md has-line \#\# NAME | specs/verify/spec.md:10 |

## Views

### audit

```mermaid
flowchart LR
  n1["audit<br/>spec · not audited"]
  classDef refuted stroke:#cf222e,stroke-width:3px
  classDef undecided stroke:#9a6700,stroke-width:2px
  classDef confirmed stroke:#1a7f37,stroke-width:2px
  classDef unaudited stroke:#6e7781,stroke-width:2px,stroke-dasharray:4
  class n1 unaudited
```

### verify

```mermaid
flowchart LR
  n1["ADR-0015<br/>bcr verify reads the records of bcr extract<br/>ADR · Undecided, no claims"]
  n2["PBI-00006<br/>Add bcr verify<br/>PBI · Undecided, no claims"]
  n3["verify<br/>How bcr verify judges the set<br/>spec · Refuted<br/>Confirmed: docs/bcr.md has-line #35;#35;#35; verify<br/>Refuted: docs/bcr.md has-line #35;#35;#35; check<br/>not audited: docs/bcr.md has-line #35;#35; NAME"]
  n2 -- "implements" --> n1
  n2 -- "changes" --> n3
  n3 -- "follows" --> n1
  classDef refuted stroke:#cf222e,stroke-width:3px
  classDef undecided stroke:#9a6700,stroke-width:2px
  classDef confirmed stroke:#1a7f37,stroke-width:2px
  classDef unaudited stroke:#6e7781,stroke-width:2px,stroke-dasharray:4
  class n1 undecided
  class n2 undecided
  class n3 refuted
```
