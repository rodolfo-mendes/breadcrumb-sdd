package crumb

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Place is where a breadcrumb, a link, a claim or a problem was
// written: a text the caller gives, such as a path, and a line. The
// core never reads the text: it repeats it, and compares it with the
// text of other places.
type Place struct {
	Text string
	Line int // counted from 1
}

func (p Place) String() string { return fmt.Sprintf("%s:%d", p.Text, p.Line) }

// SetBreadcrumb is a breadcrumb of a set: its id, and where it was
// written. Type is its type, when the set was read for Views.
type SetBreadcrumb struct {
	ID   string
	Type string
	At   Place
}

// SetLink is a link of a set: the id it points to, and where it was
// written. Verb is its verb, when the set was read for Views.
type SetLink struct {
	Verb   string
	Object string
	At     Place
}

// SetClaim is a claim of a set: the id of its breadcrumb, the claim,
// and where it was written.
type SetClaim struct {
	ID    string
	Claim Claim
	At    Place
}

// Set is the breadcrumbs of one state of a repository, with their
// links, their claims, and the problems found in them before. Its
// rules need every breadcrumb at once. Check reads the breadcrumbs
// and the links; Audit the breadcrumbs, the claims and the problems;
// Views and Given all of it, with the verdicts given before.
type Set struct {
	Breadcrumbs []SetBreadcrumb
	Links       []SetLink
	Claims      []SetClaim
	Problems    []SetProblem
	// ClaimVerdicts and Verdicts hold the verdicts given to the claims
	// and to the breadcrumbs of the set before, each at the place of
	// what it is about (ADR-0023).
	ClaimVerdicts []SetVerdict
	Verdicts      []SetVerdict
}

// SetVerdict is a verdict given before to a claim or to a breadcrumb
// of a set, written at At, with its reason when it has one.
type SetVerdict struct {
	At      Place
	Verdict Verdict
	Reason  string
}

// SetProblem is a rule a set breaks, at the place where it does.
type SetProblem struct {
	At      Place
	Message string
}

// Check returns every problem of the set, in order of place: the text
// compared as bytes, then the line. No two breadcrumbs have the same
// id, or ids that differ only in case (ADR-0003), and every link
// points to the id of a breadcrumb in the set.
func (s Set) Check() []SetProblem {
	ids := map[string]bool{}
	folded := map[string][]int{} // the breadcrumbs whose ids differ at most in case
	for i, b := range s.Breadcrumbs {
		ids[b.ID] = true
		folded[foldKey(b.ID)] = append(folded[foldKey(b.ID)], i)
	}
	for _, group := range folded {
		sort.SliceStable(group, func(i, j int) bool {
			return before(s.Breadcrumbs[group[i]].At, s.Breadcrumbs[group[j]].At)
		})
	}

	var problems []SetProblem
	report := func(at Place, format string, args ...any) {
		problems = append(problems, SetProblem{At: at, Message: fmt.Sprintf(format, args...)})
	}
	for i, b := range s.Breadcrumbs {
		var same, cased []string
		for _, j := range folded[foldKey(b.ID)] {
			switch o := s.Breadcrumbs[j]; {
			case j == i:
			case o.ID == b.ID:
				same = append(same, o.At.String())
			default:
				cased = append(cased, fmt.Sprintf("%q at %s", o.ID, o.At))
			}
		}
		if len(same) > 0 {
			report(b.At, "id %q is also at %s", b.ID, strings.Join(same, ", "))
		}
		if len(cased) > 0 {
			report(b.At, "id %q differs only in case from %s", b.ID, strings.Join(cased, ", "))
		}
	}
	for _, l := range s.Links {
		if ids[l.Object] {
			continue
		}
		var cased []string
		seen := map[string]bool{}
		for _, j := range folded[foldKey(l.Object)] {
			if id := s.Breadcrumbs[j].ID; !seen[id] {
				seen[id] = true
				cased = append(cased, fmt.Sprintf("%q", id))
			}
		}
		if len(cased) == 0 {
			report(l.At, "link points to %q, which is no breadcrumb's id", l.Object)
			continue
		}
		sort.Strings(cased)
		report(l.At, "link points to %q, which is no breadcrumb's id; %s differs only in case", l.Object, strings.Join(cased, ", "))
	}
	sort.SliceStable(problems, func(i, j int) bool { return before(problems[i].At, problems[j].At) })
	return problems
}

// before reports whether a comes before b: the text compared as bytes,
// then the line.
func before(a, b Place) bool {
	if a.Text != b.Text {
		return a.Text < b.Text
	}
	return a.Line < b.Line
}

// foldKey returns a text that is the same for two ids exactly when
// they are equal under Unicode simple case folding, as
// strings.EqualFold compares them (ADR-0003).
func foldKey(id string) string {
	return strings.Map(func(r rune) rune {
		least := r
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if f < least {
				least = f
			}
		}
		return least
	}, id)
}
