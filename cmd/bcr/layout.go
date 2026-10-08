package main

import "sort"

// A layout is a shape bcr init can set up in a repository, with the
// files that go with it (ADR-0027, specs/init/spec.md).
type layout struct {
	patterns  string        // the .breadcrumbs, after its first line
	rules     string        // the breadcrumb.rules, after its first line
	templates []initWrite   // written when they are not there
	sections  []initSection // added to the agents file after ## Breadcrumbs
}

// layouts are the layouts bcr knows, by name.
var layouts = map[string]layout{
	"asdlc": {
		patterns: asdlcPatterns,
		rules:    asdlcRules,
		templates: []initWrite{
			{path: "docs/adrs/TEMPLATE.md", text: asdlcADRTemplate},
			{path: "specs/TEMPLATE.md", text: asdlcSpecTemplate},
			{path: "tasks/TEMPLATE.md", text: asdlcPBITemplate},
		},
		sections: []initSection{{head: asdlcHead, text: asdlcSection}},
	},
}

// layoutNames are the names of the layouts bcr knows, in order.
func layoutNames() []string {
	var names []string
	for n := range layouts {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// asdlcPatterns names ASDLC's files, but the templates (ADR-0014).
const asdlcPatterns = `docs/adrs/*.md
!docs/adrs/TEMPLATE.md
specs/*/spec.md
tasks/*.md
!tasks/TEMPLATE.md
`

// asdlcRules is ASDLC's shape: ADR above spec above PBI, an ADR
// superseded rather than amended, and claims on specs only (ADR-0027).
const asdlcRules = "type\tADR\n" +
	"type\tspec\n" +
	"type\tPBI\n" +
	"link\tADR\tconstrained_by\tADR\n" +
	"link\tADR\tsupersedes\tADR\n" +
	"link\tspec\tconstrained_by\tADR\n" +
	"link\tPBI\tchanges\tspec\n" +
	"claims\tspec\n"

// asdlcADRTemplate is ASDLC's ADR: its title, its status, and its four
// sections.
const asdlcADRTemplate = `---
breadcrumb:
  id: ADR-NNN
  type: ADR
  links: []
  # links:
  #   - constrained_by ADR-NNN
  #   - supersedes ADR-NNN
---
# ADR-NNN: Decision Summary

Status: Proposed

## Context

What makes a decision necessary, and what is known.

## Decision

What is decided.

## Consequences

What follows from the decision, good and bad.

## Alternatives Considered

What else was weighed, and why it was not chosen.
`

// asdlcSpecTemplate is ASDLC's living spec: its Blueprint and its
// Contract.
const asdlcSpecTemplate = "---\n" +
	"breadcrumb:\n" +
	"  id: feature-name\n" +
	"  type: spec\n" +
	"  links: []\n" +
	"  # links:\n" +
	"  #   - constrained_by ADR-NNN\n" +
	"  claims: []\n" +
	"  # claims:\n" +
	"  #   - 'path/to/file has-line a whole line of that file'\n" +
	"---\n" +
	"# Feature: Feature Name\n" +
	"\n" +
	"## Blueprint\n" +
	"\n" +
	"### Context\n" +
	"\n" +
	"Why the feature exists, and what it does.\n" +
	"\n" +
	"### Architecture\n" +
	"\n" +
	"Its parts: their contracts, their data, what they depend on, and\n" +
	"their constraints.\n" +
	"\n" +
	"## Contract\n" +
	"\n" +
	"### Definition of Done\n" +
	"\n" +
	"- [ ] What must hold for the feature to be done.\n" +
	"\n" +
	"### Regression Guardrails\n" +
	"\n" +
	"- What must never break.\n" +
	"\n" +
	"### Scenarios\n" +
	"\n" +
	"```gherkin\n" +
	"Scenario: A behavior of the feature\n" +
	"  Given a state\n" +
	"  When something happens\n" +
	"  Then a result\n" +
	"```\n"

// asdlcPBITemplate is ASDLC's PBI: its Directive and the four sections
// after it.
const asdlcPBITemplate = `---
breadcrumb:
  id: PBI-NNN
  type: PBI
  links: []
  # links:
  #   - changes feature-name
---
# PBI-NNN: Brief Imperative Title

## Directive

What to do, in the imperative.

**Scope:**
- What the change includes, and what it leaves out.

## Dependencies
- Blocked by: None
- Must merge before: None

## Context
Read: ` + "`specs/feature-name/spec.md`" + `.

## Verification
- [ ] How to check the change is done.

## Refinement Protocol
What to do when the work needs more than the Directive says.
`

// asdlcHead is the heading of the ASDLC section of the agents file.
const asdlcHead = "## ASDLC"

// asdlcSection tells agents where ASDLC's files go, and points to
// breadcrumb.rules for the links rather than copy them (ADR-0027).
const asdlcSection = asdlcHead + `

This repository follows ASDLC's layout, set up by
` + "`bcr init --layout asdlc`" + `.

- ADRs are ` + "`docs/adrs/ADR-NNN-slug.md`" + `, written from ` + "`docs/adrs/TEMPLATE.md`" + `.
- Specs are ` + "`specs/feature-name/spec.md`" + `, the directory in kebab-case,
  written from ` + "`specs/TEMPLATE.md`" + `. A spec's id is the name of its
  directory.
- PBIs are ` + "`tasks/PBI-NNN.md`" + `, written from ` + "`tasks/TEMPLATE.md`" + `.
- NNN is the next free number of its kind, in three digits.
- Only specs carry claims.
- The links each kind of breadcrumb may have are the ` + "`link`" + ` lines of
  ` + "`breadcrumb.rules`" + `; ` + "`bcr verify`" + ` reports any other.
`
