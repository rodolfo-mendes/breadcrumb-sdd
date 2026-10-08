package crumb

import (
	"fmt"
	"sort"
)

// Rules is the shape a repository declares: the types its breadcrumbs
// may have, the links allowed between them, and the types that may
// carry claims (ADR-0025). Whatever reads them gives each with the
// line it was written at. Name is what a message calls the rules, such
// as the name of their file; the core only repeats it.
type Rules struct {
	Name   string
	Types  []RuleType
	Links  []RuleLink
	Claims []RuleClaims
}

// RuleType declares a type a breadcrumb may have.
type RuleType struct {
	Name string
	Line int // counted from 1
}

// RuleLink allows a link with Verb from a breadcrumb of type Subject
// to a breadcrumb of type Object.
type RuleLink struct {
	Subject, Verb, Object string
	Line                  int
}

// RuleClaims allows a breadcrumb of Type to carry claims.
type RuleClaims struct {
	Type string
	Line int
}

// Check returns every problem of the rules as a whole, in order of
// line: a rule written twice, at the later line, and a link or claims
// rule that names a type no rule declares. A rule written twice has no
// other problem.
func (r Rules) Check() []Problem {
	var problems []Problem
	report := func(line int, format string, args ...any) {
		problems = append(problems, Problem{Line: line, Message: fmt.Sprintf(format, args...)})
	}
	declared := map[string]bool{}
	for _, t := range r.Types {
		declared[t.Name] = true
	}
	// undeclared reports each of names, the types the rule of kind at
	// line names, that no rule declares.
	undeclared := func(line int, kind string, names ...string) {
		said := map[string]bool{}
		for _, n := range names {
			if !declared[n] && !said[n] {
				said[n] = true
				report(line, "%s names type %q, which no type line declares", kind, n)
			}
		}
	}

	types := append([]RuleType(nil), r.Types...)
	sort.SliceStable(types, func(i, j int) bool { return types[i].Line < types[j].Line })
	seenType := map[string]int{}
	for _, t := range types {
		if first, ok := seenType[t.Name]; ok {
			report(t.Line, "type %q is already declared at line %d", t.Name, first)
			continue
		}
		seenType[t.Name] = t.Line
	}

	links := append([]RuleLink(nil), r.Links...)
	sort.SliceStable(links, func(i, j int) bool { return links[i].Line < links[j].Line })
	seenLink := map[[3]string]int{}
	for _, l := range links {
		k := [3]string{l.Subject, l.Verb, l.Object}
		if first, ok := seenLink[k]; ok {
			report(l.Line, "link \"%s %s %s\" is already written at line %d", l.Subject, l.Verb, l.Object, first)
			continue
		}
		seenLink[k] = l.Line
		undeclared(l.Line, "link", l.Subject, l.Object)
	}

	claims := append([]RuleClaims(nil), r.Claims...)
	sort.SliceStable(claims, func(i, j int) bool { return claims[i].Line < claims[j].Line })
	seenClaims := map[string]int{}
	for _, c := range claims {
		if first, ok := seenClaims[c.Type]; ok {
			report(c.Line, "claims %q is already written at line %d", c.Type, first)
			continue
		}
		seenClaims[c.Type] = c.Line
		undeclared(c.Line, "claims", c.Type)
	}

	sort.SliceStable(problems, func(i, j int) bool { return problems[i].Line < problems[j].Line })
	return problems
}

// CheckShape returns every breadcrumb, link and claim of the set that
// breaks r, in order of place (ADR-0025). The shape is closed: a
// breadcrumb has a type r declares, a link the type of its subject,
// its verb and the type of its object as a link rule has them, and a
// claim a breadcrumb whose type may carry claims.
//
// The subject of a link is the breadcrumb that has its ID, its object
// the one that has its Object, and the breadcrumb of a claim the one
// that has its ID. One cause is one problem: a link or a claim is not
// judged when such an id is not that of exactly one breadcrumb, or is
// that of a breadcrumb whose type is not declared.
func (s Set) CheckShape(r Rules) []SetProblem {
	declared := map[string]bool{}
	for _, t := range r.Types {
		declared[t.Name] = true
	}
	allowed := map[[3]string]bool{}
	for _, l := range r.Links {
		allowed[[3]string{l.Subject, l.Verb, l.Object}] = true
	}
	carries := map[string]bool{}
	for _, c := range r.Claims {
		carries[c.Type] = true
	}
	count := map[string]int{}  // the breadcrumbs that have each id
	typ := map[string]string{} // the type of a breadcrumb that has each id
	for _, b := range s.Breadcrumbs {
		count[b.ID]++
		typ[b.ID] = b.Type
	}
	// known returns the type of the one breadcrumb that has id, when
	// there is one and its type is declared.
	known := func(id string) (string, bool) {
		return typ[id], count[id] == 1 && declared[typ[id]]
	}

	var problems []SetProblem
	report := func(at Place, format string, args ...any) {
		problems = append(problems, SetProblem{At: at, Message: fmt.Sprintf(format, args...)})
	}
	for _, b := range s.Breadcrumbs {
		if !declared[b.Type] {
			report(b.At, "type %q is not declared in %s", b.Type, r.Name)
		}
	}
	for _, l := range s.Links {
		subject, ok := known(l.ID)
		if !ok {
			continue
		}
		object, ok := known(l.Object)
		if !ok {
			continue
		}
		if !allowed[[3]string{subject, l.Verb, object}] {
			report(l.At, "link \"%s %s %s\" matches no rule in %s", subject, l.Verb, object, r.Name)
		}
	}
	for _, c := range s.Claims {
		if t, ok := known(c.ID); ok && !carries[t] {
			report(c.At, "a breadcrumb of type %q may not carry claims under %s", t, r.Name)
		}
	}
	SortProblems(problems)
	return problems
}
