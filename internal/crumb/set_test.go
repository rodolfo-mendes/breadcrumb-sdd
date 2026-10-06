package crumb

import (
	"reflect"
	"strings"
	"testing"
)

// The tests below follow the rules of specs/verify/spec.md about a
// set. Each builds its set as values.

func crumbAt(id, text string, line int) SetBreadcrumb {
	return SetBreadcrumb{ID: id, At: Place{Text: text, Line: line}}
}

func linkAt(object, text string, line int) SetLink {
	return SetLink{Object: object, At: Place{Text: text, Line: line}}
}

// places returns the place of each problem, as text.
func places(ps []SetProblem) []string {
	var got []string
	for _, p := range ps {
		got = append(got, p.At.String())
	}
	return got
}

// checked checks s and fails unless its problems are at want, in that
// order.
func checked(t *testing.T, s Set, want ...string) []SetProblem {
	t.Helper()
	ps := s.Check()
	if got := places(ps); !reflect.DeepEqual(got, want) {
		t.Errorf("problems at %q, want %q: %v", got, want, ps)
	}
	return ps
}

// names fails unless the message of p has every one of want, and none
// of not.
func names(t *testing.T, p SetProblem, want []string, not ...string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(p.Message, w) {
			t.Errorf("problem at %s is %q, which does not name %q", p.At, p.Message, w)
		}
	}
	for _, n := range not {
		if strings.Contains(p.Message, n) {
			t.Errorf("problem at %s is %q, which names %q", p.At, p.Message, n)
		}
	}
}

func TestASetWithNoProblem(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{crumbAt("ADR-0001", "docs/adrs/a.md", 3), crumbAt("PBI-00001", "tasks/a.md", 3)},
		Links:       []SetLink{linkAt("ADR-0001", "tasks/a.md", 6)},
	}
	if ps := s.Check(); ps != nil {
		t.Errorf("got problems %v", ps)
	}
}

func TestAnEmptySetHasNoProblem(t *testing.T) {
	if ps := (Set{}).Check(); ps != nil {
		t.Errorf("got problems %v", ps)
	}
}

func TestTwoBreadcrumbsWithTheSameID(t *testing.T) {
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{
		crumbAt("ADR-0003", "docs/adrs/a.md", 3),
		crumbAt("ADR-0003", "docs/adrs/b.md", 3),
		crumbAt("ADR-0004", "docs/adrs/c.md", 3),
	}}, "docs/adrs/a.md:3", "docs/adrs/b.md:3")
	if len(ps) == 2 {
		names(t, ps[0], []string{`"ADR-0003"`, "docs/adrs/b.md:3"}, "docs/adrs/a.md:3", "case")
		names(t, ps[1], []string{`"ADR-0003"`, "docs/adrs/a.md:3"}, "docs/adrs/b.md:3", "case")
	}
}

func TestThreeBreadcrumbsWithTheSameID(t *testing.T) {
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{
		crumbAt("PBI-00007", "tasks/c.md", 3),
		crumbAt("PBI-00007", "tasks/a.md", 3),
		crumbAt("PBI-00007", "tasks/b.md", 4),
	}}, "tasks/a.md:3", "tasks/b.md:4", "tasks/c.md:3")
	if len(ps) == 3 {
		names(t, ps[0], []string{"tasks/b.md:4, tasks/c.md:3"}, "tasks/a.md:3")
		names(t, ps[1], []string{"tasks/a.md:3, tasks/c.md:3"}, "tasks/b.md:4")
		names(t, ps[2], []string{"tasks/a.md:3, tasks/b.md:4"}, "tasks/c.md:3")
	}
}

func TestTwoBreadcrumbsWrittenAtTheSamePlace(t *testing.T) {
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{
		crumbAt("ADR-0003", "a.md", 3),
		crumbAt("ADR-0003", "a.md", 3),
	}}, "a.md:3", "a.md:3")
	for _, p := range ps {
		names(t, p, []string{"a.md:3"})
	}
}

func TestIDsThatDifferOnlyInCase(t *testing.T) {
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{
		crumbAt("ADR-0001", "docs/adrs/a.md", 3),
		crumbAt("adr-0001", "docs/adrs/b.md", 3),
	}}, "docs/adrs/a.md:3", "docs/adrs/b.md:3")
	if len(ps) == 2 {
		names(t, ps[0], []string{`"ADR-0001"`, "case", `"adr-0001"`, "docs/adrs/b.md:3"}, "docs/adrs/a.md:3")
		names(t, ps[1], []string{`"adr-0001"`, "case", `"ADR-0001"`, "docs/adrs/a.md:3"}, "docs/adrs/b.md:3")
	}
}

func TestCaseBeyondASCII(t *testing.T) {
	for _, ids := range [][2]string{
		{"ΣΙΓΜΑ", "σιγμα"},
		{"ΣΙΓΜΑΣ", "σιγμας"}, // the final sigma folds to Σ too
		{"K-1", "\u212A-1"},  // the Kelvin sign folds to k
		{"ÀÉ", "àé"},
	} {
		s := Set{Breadcrumbs: []SetBreadcrumb{crumbAt(ids[0], "a.md", 3), crumbAt(ids[1], "b.md", 3)}}
		ps := checked(t, s, "a.md:3", "b.md:3")
		if len(ps) == 2 {
			names(t, ps[0], []string{"case", "b.md:3"})
			names(t, ps[1], []string{"case", "a.md:3"})
		}
	}
}

func TestFoldingIsSimpleCaseFolding(t *testing.T) {
	for _, ids := range [][2]string{
		{"ADR-0001", "ADR-0002"},
		{"straße", "STRASSE"}, // equal only under full case folding
		{"ADR-0001", "ADR‐0001"},
		{"é", "e\u0301"},
	} {
		if strings.EqualFold(ids[0], ids[1]) {
			t.Fatalf("%q and %q are equal under simple case folding", ids[0], ids[1])
		}
		s := Set{Breadcrumbs: []SetBreadcrumb{crumbAt(ids[0], "a.md", 3), crumbAt(ids[1], "b.md", 3)}}
		if ps := s.Check(); ps != nil {
			t.Errorf("%q and %q: got problems %v", ids[0], ids[1], ps)
		}
	}
}

func TestAnIDThatIsAnothersAndDiffersInCaseFromAThird(t *testing.T) {
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{
		crumbAt("ADR-0001", "a.md", 3),
		crumbAt("ADR-0001", "b.md", 3),
		crumbAt("adr-0001", "c.md", 3),
	}}, "a.md:3", "a.md:3", "b.md:3", "b.md:3", "c.md:3")
	if len(ps) == 5 {
		names(t, ps[0], []string{"b.md:3"}, "c.md:3", "case")
		names(t, ps[1], []string{"case", `"adr-0001"`, "c.md:3"}, "b.md:3")
		names(t, ps[4], []string{"case", `"ADR-0001" at a.md:3, "ADR-0001" at b.md:3`})
	}
}

func TestALinkThatPointsNowhere(t *testing.T) {
	ps := checked(t, Set{
		Breadcrumbs: []SetBreadcrumb{crumbAt("PBI-00001", "tasks/PBI-00001.md", 3)},
		Links:       []SetLink{linkAt("ADR-0099", "tasks/PBI-00001.md", 6), linkAt("PBI-00001", "tasks/PBI-00001.md", 7)},
	}, "tasks/PBI-00001.md:6")
	if len(ps) == 1 {
		names(t, ps[0], []string{`"ADR-0099"`}, "case")
	}
}

func TestALinkInASetWithNoBreadcrumb(t *testing.T) {
	checked(t, Set{Links: []SetLink{linkAt("ADR-0001", "a.md", 6)}}, "a.md:6")
}

func TestALinkWhoseObjectDiffersOnlyInCase(t *testing.T) {
	ps := checked(t, Set{
		Breadcrumbs: []SetBreadcrumb{crumbAt("ADR-0001", "docs/adrs/a.md", 3), crumbAt("PBI-00001", "tasks/a.md", 3)},
		Links:       []SetLink{linkAt("adr-0001", "tasks/a.md", 6)},
	}, "tasks/a.md:6")
	if len(ps) == 1 {
		names(t, ps[0], []string{`"adr-0001"`, `"ADR-0001"`, "case"})
	}
}

func TestALinkToADuplicatedID(t *testing.T) {
	checked(t, Set{
		Breadcrumbs: []SetBreadcrumb{
			crumbAt("ADR-0003", "docs/adrs/a.md", 3),
			crumbAt("ADR-0003", "docs/adrs/b.md", 3),
			crumbAt("PBI-00001", "tasks/a.md", 3),
		},
		Links: []SetLink{linkAt("ADR-0003", "tasks/a.md", 6)},
	}, "docs/adrs/a.md:3", "docs/adrs/b.md:3")
}

func TestTheOrderOfProblems(t *testing.T) {
	checked(t, Set{
		Breadcrumbs: []SetBreadcrumb{
			crumbAt("B", "tasks/b.md", 3),
			crumbAt("b", "docs/a.md", 2),
			crumbAt("X", "docs/a-b.md", 30),
			crumbAt("X", "docs/a-b.md", 4),
		},
		Links: []SetLink{linkAt("NONE", "docs/a.md", 9), linkAt("NONE", "Tasks/z.md", 1)},
	}, "Tasks/z.md:1", "docs/a-b.md:4", "docs/a-b.md:30", "docs/a.md:2", "docs/a.md:9", "tasks/b.md:3")
}

func TestThePlaceIsRepeatedAsGiven(t *testing.T) {
	const text = `C:\my docs\a:b.md`
	ps := checked(t, Set{Breadcrumbs: []SetBreadcrumb{crumbAt("A", text, 3), crumbAt("A", "", 0)}}, ":0", text+":3")
	if len(ps) == 2 {
		names(t, ps[0], []string{text + ":3"})
	}
}
