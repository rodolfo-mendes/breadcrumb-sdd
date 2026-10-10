// Package page writes the page bcr report prints: every breadcrumb of
// a set with its verdict, every problem, every claim, and one diagram
// for each spec (specs/report/spec.md). It is the only package that
// knows the page's Markdown and Mermaid; what is in a view is the
// core's to say (ADR-0013).
package page

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// The groups of the order by verdict, with the class of a box in each.
const (
	refuted = iota
	undecided
	confirmed
	unaudited
)

var classes = [...]string{refuted: "refuted", undecided: "undecided", confirmed: "confirmed", unaudited: "unaudited"}

// classDefs is how a box of each class is drawn, the same in every
// diagram.
const classDefs = "  classDef refuted stroke:#cf222e,stroke-width:3px\n" +
	"  classDef undecided stroke:#9a6700,stroke-width:2px\n" +
	"  classDef confirmed stroke:#1a7f37,stroke-width:2px\n" +
	"  classDef unaudited stroke:#6e7781,stroke-width:2px,stroke-dasharray:4\n"

// group returns the group of a verdict in the order by verdict.
func group(g crumb.GivenVerdict) int {
	switch {
	case !g.Audited:
		return unaudited
	case g.Verdict == crumb.Refuted:
		return refuted
	case g.Verdict == crumb.Confirmed:
		return confirmed
	}
	return undecided
}

// written returns a verdict as the page writes it, not yet escaped.
func written(g crumb.GivenVerdict) string {
	switch {
	case !g.Audited:
		return "not audited"
	case g.Reason != "":
		return g.Verdict.String() + ", " + g.Reason
	}
	return g.Verdict.String()
}

// Write writes the page of s to w, in one piece.
func Write(w io.Writer, s crumb.Set) error {
	claims, breadcrumbs := s.Given()
	var b strings.Builder
	b.WriteString("# Breadcrumb report\n")
	b.WriteString("\n## Breadcrumbs\n\n")
	writeBreadcrumbs(&b, s, breadcrumbs)
	b.WriteString("\n## Problems\n\n")
	writeProblems(&b, s)
	b.WriteString("\n## Claims\n\n")
	writeClaims(&b, s, claims)
	b.WriteString("\n## Views\n")
	writeViews(&b, s, claims, breadcrumbs)
	_, err := io.WriteString(w, b.String())
	return err
}

// writeBreadcrumbs writes the counts and the row of each breadcrumb,
// in the order by verdict and then by id.
func writeBreadcrumbs(b *strings.Builder, s crumb.Set, given []crumb.GivenVerdict) {
	if len(s.Breadcrumbs) == 0 {
		b.WriteString("No breadcrumbs.\n")
		return
	}
	var counts [len(classes)]int
	for _, g := range given {
		counts[group(g)]++
	}
	fmt.Fprintf(b, "Breadcrumbs: %d. Refuted: %d. Undecided: %d. Confirmed: %d. Not audited: %d.\n\n",
		len(s.Breadcrumbs), counts[refuted], counts[undecided], counts[confirmed], counts[unaudited])
	order := s.ByID()
	sort.SliceStable(order, func(i, j int) bool { return group(given[order[i]]) < group(given[order[j]]) })
	b.WriteString("| Verdict | Breadcrumb | Type | File |\n|---|---|---|---|\n")
	for _, i := range order {
		c := s.Breadcrumbs[i]
		fmt.Fprintf(b, "| %s | %s | %s | %s:%d |\n", markdown(written(given[i])), markdown(c.ID), markdown(c.Type), markdown(c.At.Text), c.At.Line)
	}
}

// writeProblems writes each problem, in order of place and then as
// read.
func writeProblems(b *strings.Builder, s crumb.Set) {
	if len(s.Problems) == 0 {
		b.WriteString("No problems.\n")
		return
	}
	problems := append([]crumb.SetProblem(nil), s.Problems...)
	sort.SliceStable(problems, func(i, j int) bool { return before(problems[i].At, problems[j].At) })
	for _, p := range problems {
		fmt.Fprintf(b, "- %s:%d: %s\n", markdown(p.At.Text), p.At.Line, markdown(p.Message))
	}
}

// writeClaims writes the row of each claim, in the order by verdict
// and then by id.
func writeClaims(b *strings.Builder, s crumb.Set, given []crumb.GivenVerdict) {
	if len(s.Claims) == 0 {
		b.WriteString("No claims.\n")
		return
	}
	order := make([]int, len(s.Claims))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, c := s.Claims[order[i]], s.Claims[order[j]]
		switch ga, gc := group(given[order[i]]), group(given[order[j]]); {
		case ga != gc:
			return ga < gc
		case a.ID != c.ID:
			return a.ID < c.ID
		}
		return before(a.At, c.At)
	})
	b.WriteString("| Verdict | Breadcrumb | Claim | File |\n|---|---|---|---|\n")
	for _, i := range order {
		c := s.Claims[i]
		fmt.Fprintf(b, "| %s | %s | %s | %s:%d |\n", markdown(written(given[i])), markdown(c.ID), markdown(entry(c.Claim)), markdown(c.At.Text), c.At.Line)
	}
}

// writeViews writes the view of each spec as a Mermaid flowchart.
func writeViews(b *strings.Builder, s crumb.Set, claims, breadcrumbs []crumb.GivenVerdict) {
	views := s.Views()
	if len(views) == 0 {
		b.WriteString("\nNo specs.\n")
		return
	}
	claimsOf := s.ClaimsOf()
	for _, v := range views {
		fmt.Fprintf(b, "\n### %s\n\n```mermaid\nflowchart LR\n", markdown(s.Breadcrumbs[v.Spec].ID))
		// A box is named by its place in the view, never by an id.
		name := map[int]string{}
		for n, i := range v.Breadcrumbs {
			name[i] = fmt.Sprintf("n%d", n+1)
			c := s.Breadcrumbs[i]
			fmt.Fprintf(b, "  %s[\"%s", name[i], label(c.ID))
			if c.Title != "" {
				fmt.Fprintf(b, "<br/>%s", label(c.Title))
			}
			fmt.Fprintf(b, "<br/>%s · %s", label(c.Type), label(written(breadcrumbs[i])))
			for _, j := range claimsOf[i] {
				fmt.Fprintf(b, "<br/>%s: %s", label(written(claims[j])), label(entry(s.Claims[j].Claim)))
			}
			b.WriteString("\"]\n")
		}
		for _, l := range v.Links {
			fmt.Fprintf(b, "  %s -- \"%s\" --> %s\n", name[l.From], label(l.Verb), name[l.To])
		}
		b.WriteString(classDefs)
		for _, i := range v.Breadcrumbs {
			fmt.Fprintf(b, "  class %s %s\n", name[i], classes[group(breadcrumbs[i])])
		}
		b.WriteString("```\n")
	}
}

// entry returns a claim as it is written in a breadcrumb: its target,
// its kind and its argument.
func entry(c crumb.Claim) string {
	return c.Target + " " + c.Kind + " " + c.Argument
}

// before reports whether a comes before b: the text compared as bytes,
// then the line.
func before(a, b crumb.Place) bool {
	if a.Text != b.Text {
		return a.Text < b.Text
	}
	return a.Line < b.Line
}

// markdown returns text so that Markdown reads it as text: a \ before
// each character Markdown, a table or GitHub could read as syntax.
func markdown(text string) string {
	var b strings.Builder
	for _, r := range clean(text) {
		if strings.ContainsRune("\\`*_[]<>#|&~$", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// label returns text so that Mermaid reads it as the text of a label
// in double quotes: every ASCII character other than a letter, a
// digit, a space and - . / : , is written as its code. Mermaid reads
// two $ in a row as the start of a formula, and \n as the end of a
// line, even when they are written as codes, so U+200B, which shows
// nothing, is written between two $ and after a \.
func label(text string) string {
	var b strings.Builder
	last := rune(0)
	for _, r := range clean(text) {
		switch {
		case r >= utf8.RuneSelf, r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', strings.ContainsRune(" -./:,", r):
			b.WriteRune(r)
		case r == '$' && last == '$':
			fmt.Fprintf(&b, "#%d;#%d;", '\u200b', r)
		case r == '\\':
			fmt.Fprintf(&b, "#%d;#%d;", r, '\u200b')
		default:
			fmt.Fprintf(&b, "#%d;", r)
		}
		last = r
	}
	return b.String()
}

// clean returns text with each control character, and each byte that
// is not UTF-8, as U+FFFD, so that none ends a line of the page.
func clean(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == ' ' || r == ' ' {
			return utf8.RuneError
		}
		return r
	}, text)
}
