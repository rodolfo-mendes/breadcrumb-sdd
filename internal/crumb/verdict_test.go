package crumb

import "testing"

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
