package crumb

import (
	"reflect"
	"testing"
)

func hasLineClaim(text string) Claim {
	return Claim{Target: "docs/bcr.md", Kind: "has-line", Argument: text, Line: 8}
}

func TestAHasLineClaim(t *testing.T) {
	tests := []struct {
		name    string
		content string
		text    string
		want    Verdict
	}{
		{"a line that is equal", "# bcr\n### verify\nmore\n", "### verify", Confirmed},
		{"no such line", "# bcr\n### extract\n", "### verify", Refuted},
		{"an empty file", "", "### verify", Refuted},
		{"spaces and tabs around the line", "\t  return nil  \t\n", "return nil", Confirmed},
		{"Windows line endings", "a\r\n### verify\r\nb\r\n", "### verify", Confirmed},
		{"spaces before a Windows line ending", "### verify \r\n", "### verify", Confirmed},
		{"a last line with no line ending", "a\n### verify", "### verify", Confirmed},
		{"a line that only starts with the text", "Status: Accepted, amended by ADR-0016\n", "Status: Accepted", Refuted},
		{"a line that only ends with the text", "#### verify\n", "### verify", Refuted},
		{"letter case", "### Verify\n", "### verify", Refuted},
		{"spaces inside the line", "return  nil\n", "return nil", Refuted},
		{"a line of the front matter", "---\nbreadcrumb:\n  id: verify\n---\n", "id: verify", Confirmed},
		{"a text over two lines", "a\nb\n", "a\nb", Refuted},
	}
	for _, tt := range tests {
		if got := hasLineClaim(tt.text).Verdict(true, []byte(tt.content)); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestAClaimWhoseTargetIsNotAFile(t *testing.T) {
	c := hasLineClaim("### verify")
	if got := c.Verdict(false, []byte("### verify\n")); got != Refuted {
		t.Errorf("got %v, want Refuted", got)
	}
	if got, want := c.Lacks(false), "docs/bcr.md is not a file"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWhatATargetLacks(t *testing.T) {
	if got, want := hasLineClaim("### verify").Lacks(true), `docs/bcr.md has no line "### verify"`; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCheckAClaim(t *testing.T) {
	if why := hasLineClaim("### verify").Check(); why != "" {
		t.Errorf("a valid claim: got %q", why)
	}
	for _, c := range []Claim{
		{Target: "/etc/passwd", Kind: "has-line", Argument: "x"},
		{Target: "../x.md", Kind: "has-line", Argument: "x"},
		{Target: "docs/../x.md", Kind: "has-line", Argument: "x"},
		{Target: "docs/a b.md", Kind: "has-line", Argument: "x"},
		{Target: "docs/bcr.md", Kind: "contains", Argument: "x"},
		{Target: "docs/bcr.md", Kind: "has-line", Argument: ""},
		{Target: "docs/bcr.md", Kind: "has-line", Argument: "x "},
		{Target: "docs/bcr.md", Kind: "has-line", Argument: "a\tb"},
	} {
		if c.Check() == "" {
			t.Errorf("%+v: no problem", c)
		}
	}
}

func TestJudgeABreadcrumb(t *testing.T) {
	tests := []struct {
		name     string
		claims   []Verdict
		problems int
		want     Verdict
		reason   string
	}{
		{"every claim holds", []Verdict{Confirmed, Confirmed}, 0, Confirmed, ""},
		{"one claim does not hold", []Verdict{Confirmed, Refuted, Confirmed}, 0, Refuted, ""},
		{"an Undecided claim", []Verdict{Confirmed, Undecided}, 0, Undecided, ""},
		{"a Refuted and an Undecided claim", []Verdict{Undecided, Refuted}, 0, Refuted, ""},
		{"no claims", nil, 0, Undecided, NoClaims},
		{"a problem beside claims that hold", []Verdict{Confirmed}, 1, Undecided, ""},
		{"a problem beside a Refuted claim", []Verdict{Refuted}, 2, Refuted, ""},
		{"a problem and no claim", nil, 1, Undecided, ""},
	}
	for _, tt := range tests {
		got, reason := Judge(tt.claims, tt.problems)
		if got != tt.want || reason != tt.reason {
			t.Errorf("%s: got %v %q, want %v %q", tt.name, got, reason, tt.want, tt.reason)
		}
	}
}

func TestVerdictsAsText(t *testing.T) {
	for v, want := range map[Verdict]string{Confirmed: "Confirmed", Refuted: "Refuted", Undecided: "Undecided"} {
		if v.String() != want {
			t.Errorf("got %q, want %q", v.String(), want)
		}
	}
}

func auditSet() Set {
	claim := func(id, target, text, path string, line int) SetClaim {
		return SetClaim{ID: id, Claim: Claim{Target: target, Kind: "has-line", Argument: text, Line: line}, At: Place{Text: path, Line: line}}
	}
	return Set{
		Breadcrumbs: []SetBreadcrumb{
			{ID: "verify", At: Place{Text: "specs/verify/spec.md", Line: 3}},
			{ID: "ADR-0001", At: Place{Text: "docs/adrs/a.md", Line: 3}},
			{ID: "extract", At: Place{Text: "specs/extract/spec.md", Line: 3}},
			{ID: "asdlc", At: Place{Text: "specs/asdlc/spec.md", Line: 3}},
		},
		Claims: []SetClaim{
			claim("verify", "docs/bcr.md", "### verify", "specs/verify/spec.md", 8),
			claim("extract", "docs/bcr.md", "### extract", "specs/extract/spec.md", 8),
			claim("verify", "docs/none.md", "x", "specs/verify/spec.md", 9),
			claim("asdlc", "docs/bcr.md", "### verify", "specs/asdlc/spec.md", 8),
			claim("gone", "docs/bcr.md", "### verify", "specs/gone/spec.md", 8),
		},
		Problems: []SetProblem{
			{At: Place{Text: "specs/asdlc/spec.md", Line: 9}, Message: "claim must be written in single quotes"},
			{At: Place{Text: "tasks/broken.md", Line: 2}, Message: "breadcrumb has no type"},
		},
	}
}

func TestTheTargetsOfASet(t *testing.T) {
	if got, want := auditSet().Targets(), []string{"docs/bcr.md", "docs/none.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAuditASet(t *testing.T) {
	claims, breadcrumbs := auditSet().Audit(map[string]Target{
		"docs/bcr.md": {File: true, Content: []byte("### verify\n")},
	})
	if want := []Verdict{Confirmed, Refuted, Refuted, Confirmed, Confirmed}; !reflect.DeepEqual(claims, want) {
		t.Errorf("claims %v, want %v", claims, want)
	}
	want := []BreadcrumbVerdict{
		{Refuted, ""},         // a claim that holds, and one about no file
		{Undecided, NoClaims}, // no claim and no problem
		{Refuted, ""},         // its one claim does not hold
		{Undecided, ""},       // its claim holds, and one was dropped
	}
	if !reflect.DeepEqual(breadcrumbs, want) {
		t.Errorf("breadcrumbs %v, want %v", breadcrumbs, want)
	}
}

func TestAuditTwoBreadcrumbsWithTheSameID(t *testing.T) {
	s := Set{
		Breadcrumbs: []SetBreadcrumb{{ID: "ADR-0003", At: Place{Text: "a.md", Line: 3}}, {ID: "ADR-0003", At: Place{Text: "b.md", Line: 3}}},
		Claims:      []SetClaim{{ID: "ADR-0003", Claim: Claim{Target: "x.md", Kind: "has-line", Argument: "x"}, At: Place{Text: "a.md", Line: 8}}},
	}
	_, breadcrumbs := s.Audit(map[string]Target{"x.md": {File: true, Content: []byte("x")}})
	if want := []BreadcrumbVerdict{{Confirmed, ""}, {Undecided, NoClaims}}; !reflect.DeepEqual(breadcrumbs, want) {
		t.Errorf("got %v, want %v", breadcrumbs, want)
	}
}
