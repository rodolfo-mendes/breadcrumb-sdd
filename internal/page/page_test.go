package page

import (
	"errors"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

var update = flag.Bool("update", false, "write the page of the test again")

// page returns the page of s.
func page(t *testing.T, s crumb.Set) string {
	t.Helper()
	var b strings.Builder
	if err := Write(&b, s); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func at(text string, line int) crumb.Place { return crumb.Place{Text: text, Line: line} }

// small is a spec that follows one ADR, a PBI that implements it and
// changes the spec, an ADR and a spec no link touches, a claim of each
// verdict and one that was not audited, and two problems.
func small() crumb.Set {
	claim := func(text string, line int) crumb.SetClaim {
		return crumb.SetClaim{
			ID:    "verify",
			Claim: crumb.Claim{Target: "docs/bcr.md", Kind: "has-line", Argument: text, Line: line},
			At:    at("specs/verify/spec.md", line),
		}
	}
	return crumb.Set{
		Breadcrumbs: []crumb.SetBreadcrumb{
			{ID: "verify", Type: "spec", At: at("specs/verify/spec.md", 3)},
			{ID: "PBI-00006", Type: "PBI", At: at("tasks/PBI-00006.md", 3)},
			{ID: "ADR-0015", Type: "ADR", At: at("docs/adrs/ADR-0015.md", 3)},
			{ID: "ADR-0004", Type: "ADR", At: at("docs/adrs/ADR-0004.md", 3)},
			{ID: "audit", Type: "spec", At: at("specs/audit/spec.md", 3)},
		},
		Links: []crumb.SetLink{
			{Verb: "follows", Object: "ADR-0015", At: at("specs/verify/spec.md", 6)},
			{Verb: "changes", Object: "verify", At: at("tasks/PBI-00006.md", 7)},
			{Verb: "implements", Object: "ADR-0015", At: at("tasks/PBI-00006.md", 6)},
		},
		Claims: []crumb.SetClaim{claim("### verify", 8), claim("### check", 9), claim("## NAME", 10)},
		Problems: []crumb.SetProblem{
			{At: at("specs/verify/spec.md", 11), Message: `claim "docs/bcr.md" must be a target, a kind and an argument`},
			{At: at("docs/adrs/ADR-0099.md", 2), Message: "breadcrumb has no id"},
		},
		ClaimVerdicts: []crumb.SetVerdict{
			{At: at("specs/verify/spec.md", 8), Verdict: crumb.Confirmed},
			{At: at("specs/verify/spec.md", 9), Verdict: crumb.Refuted},
		},
		Verdicts: []crumb.SetVerdict{
			{At: at("specs/verify/spec.md", 3), Verdict: crumb.Refuted},
			{At: at("tasks/PBI-00006.md", 3), Verdict: crumb.Undecided, Reason: crumb.NoClaims},
			{At: at("docs/adrs/ADR-0015.md", 3), Verdict: crumb.Undecided, Reason: crumb.NoClaims},
			{At: at("docs/adrs/ADR-0004.md", 3), Verdict: crumb.Confirmed},
		},
	}
}

func TestTheWholePage(t *testing.T) {
	const file = "testdata/small.golden.md"
	got := page(t, small())
	if *update {
		if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("the page is\n%s\nwant\n%s\nrun go test ./internal/page -update when the page is meant to change", got, want)
	}
}

func TestThePageOfAnEmptySet(t *testing.T) {
	const want = "# Breadcrumb report\n\n## Breadcrumbs\n\nNo breadcrumbs.\n\n## Problems\n\nNo problems.\n\n" +
		"## Claims\n\nNo claims.\n\n## Views\n\nNo specs.\n"
	if got := page(t, crumb.Set{}); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMarkdownEscapes(t *testing.T) {
	for text, want := range map[string]string{
		"ADR-0001":                 "ADR-0001",
		"docs/bcr.md:3":            "docs/bcr.md:3",
		"### verify":               `\#\#\# verify`,
		"a|b":                      `a\|b`,
		"`x`":                      "\\`x\\`",
		"<script>":                 `\<script\>`,
		"[a](b)":                   `\[a\](b)`,
		`a\|b`:                     `a\\\|b`,
		"*a* _b_ ~c~ $d$ &amp;":    `\*a\* \_b\_ \~c\~ \$d\$ \&amp;`,
		"a\rb\x00c\x1bd e":         "a�b�c�d�e",
		"a\xffb":                   "a�b",
		"claim \"x\" is 'quoted'.": "claim \"x\" is 'quoted'.",
	} {
		if got := markdown(text); got != want {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}
}

func TestLabelEscapes(t *testing.T) {
	for text, want := range map[string]string{
		"ADR-0001":             "ADR-0001",
		"docs/bcr.md has-line": "docs/bcr.md has-line",
		"Undecided, no claims": "Undecided, no claims",
		"### verify":           "#35;#35;#35; verify",
		`a"b#c|d<e` + "`f":     "a#34;b#35;c#124;d#60;e#96;f",
		"%%{init}%%":           "#37;#37;#123;init#125;#37;#37;",
		"<br/>&amp;;":          "#60;br/#62;#38;amp#59;#59;",
		"$$x$$ $y$":            "#36;#8203;#36;x#36;#8203;#36; #36;y#36;",
		`a\nb`:                 "a#92;#8203;nb",
		"é·ü":                  "é·ü",
		"a\rb\xffc":            "a�b�c",
		"(a) [b] {c} *d* _e_":  "#40;a#41; #91;b#93; #123;c#125; #42;d#42; #95;e#95;",
	} {
		if got := label(text); got != want {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}
}

type closed struct{}

func (closed) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestWriteReturnsTheErrorOfItsWriter(t *testing.T) {
	if err := Write(closed{}, small()); err == nil {
		t.Error("got no error")
	}
}
