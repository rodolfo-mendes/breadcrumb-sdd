package crumb

import (
	"reflect"
	"testing"
)

// The tests below follow the rules of specs/report/spec.md about a
// view. Each builds its set as values.

func typed(id, typ, text string) SetBreadcrumb {
	return SetBreadcrumb{ID: id, Type: typ, At: Place{Text: text, Line: 3}}
}

func verbAt(verb, object, text string, line int) SetLink {
	return SetLink{Verb: verb, Object: object, At: Place{Text: text, Line: line}}
}

// idsOf returns the id of each breadcrumb of s that indexes names.
func idsOf(s Set, indexes []int) []string {
	var ids []string
	for _, i := range indexes {
		ids = append(ids, s.Breadcrumbs[i].ID)
	}
	return ids
}

// drawn returns each link of v as "FROM VERB TO".
func drawn(s Set, v View) []string {
	var links []string
	for _, l := range v.Links {
		links = append(links, s.Breadcrumbs[l.From].ID+" "+l.Verb+" "+s.Breadcrumbs[l.To].ID)
	}
	return links
}

// A spec that follows two ADRs, a PBI that changes it and implements
// one of them, an ADR that amends the other, and a PBI apart.
func around() Set {
	return Set{
		Breadcrumbs: []SetBreadcrumb{
			typed("PBI-2", "PBI", "tasks/2.md"),
			typed("verify", "spec", "specs/verify/spec.md"),
			typed("ADR-2", "ADR", "adrs/2.md"),
			typed("ADR-1", "ADR", "adrs/1.md"),
			typed("PBI-1", "PBI", "tasks/1.md"),
			typed("ADR-3", "ADR", "adrs/3.md"),
		},
		Links: []SetLink{
			verbAt("implements", "ADR-1", "tasks/1.md", 7),
			verbAt("changes", "verify", "tasks/1.md", 6),
			verbAt("follows", "ADR-2", "specs/verify/spec.md", 7),
			verbAt("follows", "ADR-1", "specs/verify/spec.md", 6),
			verbAt("amends", "ADR-1", "adrs/2.md", 6),
			verbAt("amends", "ADR-2", "adrs/3.md", 6),
			verbAt("implements", "ADR-3", "tasks/2.md", 6),
		},
	}
}

func TestAViewHoldsTheSpecAndWhatItTouches(t *testing.T) {
	s := around()
	views := s.Views()
	if len(views) != 1 {
		t.Fatalf("got %d views, want 1", len(views))
	}
	v := views[0]
	if s.Breadcrumbs[v.Spec].ID != "verify" {
		t.Errorf("the spec is %q, want verify", s.Breadcrumbs[v.Spec].ID)
	}
	if got, want := idsOf(s, v.Breadcrumbs), []string{"ADR-1", "ADR-2", "PBI-1", "verify"}; !reflect.DeepEqual(got, want) {
		t.Errorf("breadcrumbs %q, want %q: ADR-3 and PBI-2 are two links away", got, want)
	}
	want := []string{"ADR-2 amends ADR-1", "PBI-1 implements ADR-1", "PBI-1 changes verify", "verify follows ADR-1", "verify follows ADR-2"}
	if got := drawn(s, v); !reflect.DeepEqual(got, want) {
		t.Errorf("links %q, want %q", got, want)
	}
}

func TestViewsAreInOrderOfTheSpecsID(t *testing.T) {
	s := Set{Breadcrumbs: []SetBreadcrumb{
		typed("verify", "spec", "v.md"), typed("ADR-1", "ADR", "a.md"), typed("audit", "spec", "u.md"), typed("Zed", "spec", "z.md"),
	}}
	var got []string
	for _, v := range s.Views() {
		got = append(got, s.Breadcrumbs[v.Spec].ID)
		if len(v.Breadcrumbs) != 1 || len(v.Links) != 0 {
			t.Errorf("the view of %s holds %v and %v, want the spec alone", got[len(got)-1], v.Breadcrumbs, v.Links)
		}
	}
	if want := []string{"Zed", "audit", "verify"}; !reflect.DeepEqual(got, want) {
		t.Errorf("views of %q, want %q: ids are compared as bytes", got, want)
	}
}

func TestNoSpecNoView(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-1", "ADR", "a.md"), typed("PBI-1", "PBI", "p.md"), typed("x", "Spec", "x.md")},
		Links:       []SetLink{verbAt("implements", "ADR-1", "p.md", 6)},
	}
	if views := s.Views(); len(views) != 0 {
		t.Errorf("got %v, want no view: the type is compared byte for byte", views)
	}
}

func TestALinkToNoBreadcrumbIsInNoView(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("verify", "spec", "v.md")},
		Links:       []SetLink{verbAt("follows", "ADR-9", "v.md", 6), verbAt("changes", "verify", "gone.md", 6)},
	}
	v := s.Views()[0]
	if len(v.Breadcrumbs) != 1 || len(v.Links) != 0 {
		t.Errorf("got %v and %v, want the spec alone", v.Breadcrumbs, v.Links)
	}
}

func TestTwoSpecsThatLinkAreInEachOthersView(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("audit", "spec", "a.md"), typed("verify", "spec", "v.md")},
		Links:       []SetLink{verbAt("extends", "verify", "a.md", 6)},
	}
	for _, v := range s.Views() {
		if got, want := idsOf(s, v.Breadcrumbs), []string{"audit", "verify"}; !reflect.DeepEqual(got, want) {
			t.Errorf("the view of %s holds %q, want %q", s.Breadcrumbs[v.Spec].ID, got, want)
		}
		if got, want := drawn(s, v), []string{"audit extends verify"}; !reflect.DeepEqual(got, want) {
			t.Errorf("the view of %s draws %q, want %q", s.Breadcrumbs[v.Spec].ID, got, want)
		}
	}
}

func TestTwoBreadcrumbsWithTheSameIDAreBothInAView(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-1", "ADR", "b.md"), typed("verify", "spec", "v.md"), typed("ADR-1", "ADR", "a.md")},
		Links:       []SetLink{verbAt("follows", "ADR-1", "v.md", 6)},
	}
	v := s.Views()[0]
	if got, want := v.Breadcrumbs, []int{2, 0, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("breadcrumbs %v, want %v: the same id is ordered by place", got, want)
	}
	if got, want := v.Links, []ViewLink{{From: 1, To: 2, Verb: "follows"}, {From: 1, To: 0, Verb: "follows"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("links %v, want %v", got, want)
	}
}

func TestTheSameLinkReadTwiceIsDrawnOnce(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-1", "ADR", "a.md"), typed("verify", "spec", "v.md")},
		Links: []SetLink{
			verbAt("follows", "ADR-1", "v.md", 6), verbAt("follows", "ADR-1", "v.md", 7), verbAt("amends", "ADR-1", "v.md", 8),
		},
	}
	if got, want := drawn(s, s.Views()[0]), []string{"verify amends ADR-1", "verify follows ADR-1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("links %q, want %q", got, want)
	}
}

func TestViewsDoNotDependOnTheOrderRead(t *testing.T) {
	s := around()
	want := s.Views()[0]
	r := Set{}
	for i := len(s.Breadcrumbs) - 1; i >= 0; i-- {
		r.Breadcrumbs = append(r.Breadcrumbs, s.Breadcrumbs[i])
	}
	for i := len(s.Links) - 1; i >= 0; i-- {
		r.Links = append(r.Links, s.Links[i])
	}
	got := r.Views()[0]
	if !reflect.DeepEqual(idsOf(r, got.Breadcrumbs), idsOf(s, want.Breadcrumbs)) || !reflect.DeepEqual(drawn(r, got), drawn(s, want)) {
		t.Errorf("got %q and %q, want %q and %q", idsOf(r, got.Breadcrumbs), drawn(r, got), idsOf(s, want.Breadcrumbs), drawn(s, want))
	}
}

func TestClaimsOfABreadcrumb(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-1", "ADR", "a.md"), typed("verify", "spec", "v.md")},
		Claims: []SetClaim{
			{ID: "verify", Claim: hasLineClaim("x"), At: Place{Text: "v.md", Line: 8}},
			{ID: "gone", Claim: hasLineClaim("y"), At: Place{Text: "gone.md", Line: 8}},
			{ID: "verify", Claim: hasLineClaim("z"), At: Place{Text: "v.md", Line: 9}},
		},
	}
	if got, want := s.ClaimsOf(), [][]int{nil, {0, 2}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestTheVerdictsGivenBefore(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{typed("ADR-1", "ADR", "a.md"), typed("verify", "spec", "v.md"), typed("audit", "spec", "u.md")},
		Claims: []SetClaim{
			{ID: "verify", Claim: hasLineClaim("x"), At: Place{Text: "v.md", Line: 8}},
			{ID: "verify", Claim: hasLineClaim("y"), At: Place{Text: "v.md", Line: 9}},
		},
		ClaimVerdicts: []SetVerdict{{At: Place{Text: "v.md", Line: 9}, Verdict: Refuted}},
		Verdicts: []SetVerdict{
			{At: Place{Text: "v.md", Line: 3}, Verdict: Confirmed},
			{At: Place{Text: "a.md", Line: 3}, Verdict: Undecided, Reason: NoClaims},
			{At: Place{Text: "v.md", Line: 3}, Verdict: Refuted},
		},
	}
	claims, breadcrumbs := s.Given()
	if want := []GivenVerdict{{}, {Audited: true, Verdict: Refuted}}; !reflect.DeepEqual(claims, want) {
		t.Errorf("claims %+v, want %+v", claims, want)
	}
	want := []GivenVerdict{{Audited: true, Verdict: Undecided, Reason: NoClaims}, {Audited: true, Verdict: Refuted}, {}}
	if !reflect.DeepEqual(breadcrumbs, want) {
		t.Errorf("breadcrumbs %+v, want %+v: the verdict given last holds, and audit has none", breadcrumbs, want)
	}
}

func TestParseVerdict(t *testing.T) {
	for _, v := range []Verdict{Undecided, Confirmed, Refuted} {
		if got, ok := ParseVerdict(v.String()); !ok || got != v {
			t.Errorf("%v: got %v, %v", v, got, ok)
		}
	}
	for _, s := range []string{"", "confirmed", "Confirmed ", "Passed"} {
		if _, ok := ParseVerdict(s); ok {
			t.Errorf("%q names a verdict, want none", s)
		}
	}
}
