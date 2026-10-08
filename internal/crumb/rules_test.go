package crumb

import (
	"reflect"
	"testing"
)

// The tests below follow the rules of specs/verify/spec.md about
// breadcrumb.rules. Each builds its rules and its set as values.

// layout returns the rules the tests share, each at the line it has
// in the Scenarios of the spec.
func layout() Rules {
	return Rules{
		Name:  "breadcrumb.rules",
		Types: []RuleType{{"ADR", 1}, {"spec", 2}, {"PBI", 3}},
		Links: []RuleLink{
			{"spec", "constrained_by", "ADR", 4},
			{"PBI", "changes", "spec", 5},
		},
		Claims: []RuleClaims{{"spec", 6}},
	}
}

// from returns a link written in the breadcrumb whose id is id.
func from(id, verb, object, text string, line int) SetLink {
	return SetLink{ID: id, Verb: verb, Object: object, At: Place{Text: text, Line: line}}
}

// claimOf returns a claim of the breadcrumb whose id is id.
func claimOf(id, text string, line int) SetClaim {
	return SetClaim{ID: id, At: Place{Text: text, Line: line}}
}

// shaped returns a set that has the shape of layout.
func shaped() Set {
	return Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-7", "ADR", "a.md"), typed("verify", "spec", "v.md"), typed("PBI-30", "PBI", "p.md")},
		Links:       []SetLink{from("verify", "constrained_by", "ADR-7", "v.md", 6), from("PBI-30", "changes", "verify", "p.md", 6)},
		Claims:      []SetClaim{claimOf("verify", "v.md", 8)},
	}
}

// shapeChecked checks s against r and fails unless its problems are
// at want, in that order.
func shapeChecked(t *testing.T, s Set, r Rules, want ...string) []SetProblem {
	t.Helper()
	ps := s.CheckShape(r)
	if got := places(ps); !reflect.DeepEqual(got, want) {
		t.Errorf("problems at %q, want %q: %v", got, want, ps)
	}
	return ps
}

// rulesChecked checks r and fails unless its problems are want.
func rulesChecked(t *testing.T, r Rules, want ...Problem) {
	t.Helper()
	if got := r.Check(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestASetThatHasTheDeclaredShape(t *testing.T) {
	shapeChecked(t, shaped(), layout())
}

func TestATypeThatIsNotDeclared(t *testing.T) {
	s := shaped()
	s.Breadcrumbs = append(s.Breadcrumbs, typed("ADR-8", "adr", "b.md"))
	ps := shapeChecked(t, s, layout(), "b.md:3")
	if want := `type "adr" is not declared in breadcrumb.rules`; len(ps) == 1 && ps[0].Message != want {
		t.Errorf("got %q, want %q", ps[0].Message, want)
	}
}

func TestALinkNoRuleAllows(t *testing.T) {
	s := shaped()
	s.Links = append(s.Links, from("PBI-30", "implements", "ADR-7", "p.md", 7))
	ps := shapeChecked(t, s, layout(), "p.md:7")
	if want := `link "PBI implements ADR" matches no rule in breadcrumb.rules`; len(ps) == 1 && ps[0].Message != want {
		t.Errorf("got %q, want %q", ps[0].Message, want)
	}
}

func TestALinkRuleHasADirection(t *testing.T) {
	s := shaped()
	s.Links = append(s.Links, from("verify", "changes", "PBI-30", "v.md", 7))
	shapeChecked(t, s, layout(), "v.md:7")
}

func TestAClaimOnATypeThatMayNotCarryClaims(t *testing.T) {
	s := shaped()
	s.Claims = append(s.Claims, claimOf("PBI-30", "p.md", 9))
	ps := shapeChecked(t, s, layout(), "p.md:9")
	if want := `a breadcrumb of type "PBI" may not carry claims under breadcrumb.rules`; len(ps) == 1 && ps[0].Message != want {
		t.Errorf("got %q, want %q", ps[0].Message, want)
	}
}

func TestWithNoClaimsRuleNoTypeCarriesClaims(t *testing.T) {
	r := layout()
	r.Claims = nil
	shapeChecked(t, shaped(), r, "v.md:8")
}

func TestRulesThatDeclareNothing(t *testing.T) {
	shapeChecked(t, shaped(), Rules{}, "a.md:3", "p.md:3", "v.md:3")
}

func TestALinkOrAClaimOfAnUndeclaredTypeIsNotJudged(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-7", "ADR", "a.md"), typed("verify", "Spec", "v.md")},
		Links:       []SetLink{from("verify", "follows", "ADR-7", "v.md", 6)},
		Claims:      []SetClaim{claimOf("verify", "v.md", 8)},
	}
	shapeChecked(t, s, layout(), "v.md:3")
}

func TestALinkToAnUndeclaredTypeIsNotJudged(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-7", "adr", "a.md"), typed("verify", "spec", "v.md")},
		Links:       []SetLink{from("verify", "constrained_by", "ADR-7", "v.md", 6)},
	}
	shapeChecked(t, s, layout(), "a.md:3")
}

func TestALinkThatPointsNowhereIsNotJudged(t *testing.T) {
	s := shaped()
	s.Links = append(s.Links, from("PBI-30", "implements", "ADR-99", "p.md", 7))
	shapeChecked(t, s, layout())
}

func TestALinkOrAClaimOfNoBreadcrumbIsNotJudged(t *testing.T) {
	s := shaped()
	s.Links = append(s.Links, from("PBI-99", "implements", "ADR-7", "x.md", 6))
	s.Claims = append(s.Claims, claimOf("PBI-99", "x.md", 8))
	shapeChecked(t, s, layout())
}

func TestADuplicatedIDIsNotJudged(t *testing.T) {
	s := shaped()
	s.Breadcrumbs = append(s.Breadcrumbs, typed("ADR-7", "ADR", "b.md"), typed("verify", "spec", "w.md"))
	s.Links = append(s.Links, from("PBI-30", "implements", "ADR-7", "p.md", 7))
	s.Claims = append(s.Claims, claimOf("verify", "w.md", 8))
	// Neither the links from and to the two ids, nor their claims.
	shapeChecked(t, s, layout())
}

func TestShapeProblemsAreInOrderOfPlace(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("PBI-30", "PBI", "tasks/p.md"), typed("ADR-7", "ADR", "docs/a.md"), typed("x", "note", "docs/a.md")},
		Links:       []SetLink{from("PBI-30", "implements", "ADR-7", "tasks/p.md", 7)},
		Claims:      []SetClaim{claimOf("ADR-7", "docs/a.md", 9), claimOf("PBI-30", "tasks/p.md", 5)},
	}
	shapeChecked(t, s, layout(), "docs/a.md:3", "docs/a.md:9", "tasks/p.md:5", "tasks/p.md:7")
}

func TestTheNameOfTheRulesIsRepeatedAsGiven(t *testing.T) {
	r := layout()
	r.Name = "the layout"
	s := Set{Breadcrumbs: []SetBreadcrumb{typed("x", "note", "x.md")}}
	if ps := s.CheckShape(r); len(ps) != 1 || ps[0].Message != `type "note" is not declared in the layout` {
		t.Errorf("got %v", ps)
	}
}

func TestRulesWithNoProblem(t *testing.T) {
	rulesChecked(t, layout())
	rulesChecked(t, Rules{})
}

func TestARuleThatNamesATypeNoRuleDeclares(t *testing.T) {
	r := layout()
	r.Links = append(r.Links, RuleLink{"PBI", "changes", "Spec", 7}, RuleLink{"Note", "cites", "Page", 9}, RuleLink{"Note", "cites", "Note", 10})
	r.Claims = append(r.Claims, RuleClaims{"pbi", 8})
	rulesChecked(t, r,
		Problem{7, `link names type "Spec", which no type line declares`},
		Problem{8, `claims names type "pbi", which no type line declares`},
		Problem{9, `link names type "Note", which no type line declares`},
		Problem{9, `link names type "Page", which no type line declares`},
		Problem{10, `link names type "Note", which no type line declares`},
	)
}

func TestATypeMayBeDeclaredAfterTheRuleThatNamesIt(t *testing.T) {
	rulesChecked(t, Rules{
		Links:  []RuleLink{{"PBI", "changes", "spec", 1}},
		Claims: []RuleClaims{{"spec", 2}},
		Types:  []RuleType{{"PBI", 3}, {"spec", 4}},
	})
}

func TestARuleWrittenTwice(t *testing.T) {
	r := layout()
	r.Types = append(r.Types, RuleType{"spec", 7}, RuleType{"spec", 10})
	r.Links = append(r.Links, RuleLink{"PBI", "changes", "spec", 8})
	r.Claims = append(r.Claims, RuleClaims{"spec", 9})
	rulesChecked(t, r,
		Problem{7, `type "spec" is already declared at line 2`},
		Problem{8, `link "PBI changes spec" is already written at line 5`},
		Problem{9, `claims "spec" is already written at line 6`},
		Problem{10, `type "spec" is already declared at line 2`},
	)
}

func TestARuleWrittenTwiceHasNoOtherProblem(t *testing.T) {
	rulesChecked(t, Rules{Links: []RuleLink{{"a", "b", "c", 1}, {"a", "b", "c", 2}}},
		Problem{1, `link names type "a", which no type line declares`},
		Problem{1, `link names type "c", which no type line declares`},
		Problem{2, `link "a b c" is already written at line 1`},
	)
}

func TestTheFirstOfTwoRulesIsTheOneWithTheLowerLine(t *testing.T) {
	rulesChecked(t, Rules{Types: []RuleType{{"spec", 5}, {"spec", 2}}},
		Problem{5, `type "spec" is already declared at line 2`},
	)
}
