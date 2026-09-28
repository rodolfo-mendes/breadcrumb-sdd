package breadcrumb

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func file(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

func load(t *testing.T, fsys fstest.MapFS) Graph {
	t.Helper()
	g, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func only(t *testing.T, g Graph) Breadcrumb {
	t.Helper()
	if len(g.Breadcrumbs) != 1 {
		t.Fatalf("got %d breadcrumbs, want 1", len(g.Breadcrumbs))
	}
	return g.Breadcrumbs[0]
}

func TestBreadcrumbFilesAreNamedByTypeAndNumber(t *testing.T) {
	g := load(t, fstest.MapFS{
		"breadcrumbs/IN-0001.md":     file("# IN-0001: An ask\n"),
		"breadcrumbs/RQ-0001.md":     file("# RQ-0001: A requirement\n"),
		"breadcrumbs/TD-0001.md":     file("# TD-0001: A decision\n"),
		"breadcrumbs/TK-0001.md":     file("# TK-0001: A task\n"),
		"breadcrumbs/td-0002.md":     file("# lower case\n"),
		"breadcrumbs/TD-2.md":        file("# too few digits\n"),
		"breadcrumbs/XX-0001.md":     file("# unknown type\n"),
		"breadcrumbs/notes.md":       file("# notes\n"),
		"breadcrumbs/sub/TD-0003.md": file("# in a subdirectory\n"),
		"TD-0004.md":                 file("# outside breadcrumbs/\n"),
	})

	var got []string
	for _, b := range g.Breadcrumbs {
		got = append(got, b.ID+" "+b.Type.String())
	}
	want := []string{"IN-0001 Intake", "RQ-0001 Requirement", "TD-0001 Technical Decision", "TK-0001 Task"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("breadcrumbs: got %q, want %q", got, want)
	}

	wantOthers := []string{
		"breadcrumbs/TD-2.md",
		"breadcrumbs/XX-0001.md",
		"breadcrumbs/notes.md",
		"breadcrumbs/sub/TD-0003.md",
		"breadcrumbs/td-0002.md",
	}
	if !reflect.DeepEqual(g.Others, wantOthers) {
		t.Errorf("others: got %q, want %q", g.Others, wantOthers)
	}
}

func TestARepositoryWithoutBreadcrumbsHasNone(t *testing.T) {
	g := load(t, fstest.MapFS{"README.md": file("# readme\n")})
	if len(g.Breadcrumbs) != 0 || len(g.Others) != 0 {
		t.Errorf("got %+v, want an empty graph", g)
	}
}

func TestTheTitleDropsItsIDPrefix(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("# TD-0001: Adopt the method\n")}))
	if b.Title != "Adopt the method" {
		t.Errorf("got %q", b.Title)
	}
}

func TestATitleWithoutTheIDIsKept(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("# Adopt the method\n")}))
	if b.Title != "Adopt the method" {
		t.Errorf("got %q", b.Title)
	}
}

func TestAFileNotStartingWithATitleIsAProblem(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("Adopt the method\n")}))
	if len(b.Problems) != 1 {
		t.Errorf("got problems %q, want one", b.Problems)
	}
}

func TestParentsAreTheLinesRightBelowTheTitle(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/RQ-0001.md": file(
		"# RQ-0001: A requirement\n\nParent: [IN-0001](IN-0001.md)\nParent: [TD-0002](TD-0002.md)\n\nBody.\n")}))
	want := []Link{
		{Text: "IN-0001", Target: "breadcrumbs/IN-0001.md", Line: 3},
		{Text: "TD-0002", Target: "breadcrumbs/TD-0002.md", Line: 4},
	}
	if !reflect.DeepEqual(b.Parents, want) {
		t.Errorf("got %+v, want %+v", b.Parents, want)
	}
}

// TD-0005 shows a Parent: line as an example in its body.
func TestAParentLineFurtherDownIsNotALink(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0005.md": file(
		"# TD-0005: Link a breadcrumb\n\n## Decision\n\n```markdown\nParent: [IN-0001](IN-0001.md)\n```\n")}))
	if len(b.Parents) != 0 {
		t.Errorf("got parents %+v, want none", b.Parents)
	}
}

func TestAMalformedParentLineIsAProblem(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/RQ-0001.md": file(
		"# RQ-0001: A requirement\n\nParent: IN-0001\n")}))
	if len(b.Parents) != 0 || len(b.Problems) != 1 {
		t.Errorf("got parents %+v and problems %q, want no parent and one problem", b.Parents, b.Problems)
	}
}

func TestCarriageReturnsBelongToTheLineEnding(t *testing.T) {
	g := load(t, fstest.MapFS{
		"breadcrumbs/IN-0001.md": file("# IN-0001: An ask\n"),
		"breadcrumbs/RQ-0001.md": file("# RQ-0001: A requirement\r\n\r\nParent: [IN-0001](IN-0001.md)\r\n"),
	})
	b, _ := g.Lookup("breadcrumbs/RQ-0001.md")
	if b.Title != "A requirement" || len(b.Parents) != 1 || len(b.Problems) != 0 {
		t.Errorf("got %+v", b)
	}
}

func TestABreadcrumbWithNothingBelowItIsUndecided(t *testing.T) {
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("# TD-0001: A decision\n")}))
	if b.Verdict != Undecided {
		t.Errorf("got %v", b.Verdict)
	}
}

func TestLookupFindsABreadcrumbByPath(t *testing.T) {
	g := load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("# TD-0001: A decision\n")})
	if _, ok := g.Lookup("breadcrumbs/TD-0001.md"); !ok {
		t.Error("TD-0001 not found")
	}
	if _, ok := g.Lookup("breadcrumbs/TD-0002.md"); ok {
		t.Error("TD-0002 found")
	}
}

func task(claims ...string) string {
	return "# TK-0001: A task\n\nParent: [TD-0001](TD-0001.md)\n\n## Claims\n\n" + strings.Join(claims, "\n") + "\n\n## Notes\n\n- not a claim, outside the section\n"
}

func verdicts(g Graph) map[string]string {
	m := map[string]string{}
	for _, b := range g.Breadcrumbs {
		m[b.ID] = b.Verdict.String()
	}
	return m
}

var repo = fstest.MapFS{
	"src/cli.go":             file("package cli\n\nfunc save(path string) {}\n"),
	"breadcrumbs/TD-0001.md": file("# TD-0001: The decision every task carries out\n"),
}

func with(extra fstest.MapFS) fstest.MapFS {
	fsys := fstest.MapFS{}
	for k, v := range repo {
		fsys[k] = v
	}
	for k, v := range extra {
		fsys[k] = v
	}
	return fsys
}

func TestClaimsAreTheListItemsUnderTheClaimsHeading(t *testing.T) {
	g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0001.md": file(task(
		"- `src/cli.go` contains `func save(`",
		"- `src/cli.go` contains `package cli`",
	))}))
	b, _ := g.Lookup("breadcrumbs/TK-0001.md")
	want := []Claim{{Path: "src/cli.go", Text: "func save(", Holds: true}, {Path: "src/cli.go", Text: "package cli", Holds: true}}
	if !reflect.DeepEqual(b.Claims, want) || len(b.Problems) != 0 {
		t.Errorf("got claims %+v and problems %q, want %+v", b.Claims, b.Problems, want)
	}
}

func TestATaskIsConfirmedWhenEveryClaimHolds(t *testing.T) {
	g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0001.md": file(task("- `src/cli.go` contains `func save(`"))}))
	if v := verdicts(g)["TK-0001"]; v != "Confirmed" {
		t.Errorf("got %s", v)
	}
}

func TestATaskIsRefutedWhenAClaimFails(t *testing.T) {
	for name, claim := range map[string]string{
		"text not in the file": "- `src/cli.go` contains `func Save(`",
		"no such file":         "- `src/gone.go` contains `func save(`",
		"a directory":          "- `src` contains `func save(`",
	} {
		g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0001.md": file(task("- `src/cli.go` contains `func save(`", claim))}))
		if v := verdicts(g)["TK-0001"]; v != "Refuted" {
			t.Errorf("%s: got %s", name, v)
		}
	}
}

func TestATaskIsRefutedWhenAClaimCannotBeRead(t *testing.T) {
	for _, claim := range []string{
		"- src/cli.go contains func save(",
		"- `src/cli.go` includes `func save(`",
		"* `src/cli.go` contains `func save(`",
		"  - `src/cli.go` contains `func save(`",
		"- `/src/cli.go` contains `func save(`",
		"- `src/../src/cli.go` contains `func save(`",
		"- `./src/cli.go` contains `func save(`",
	} {
		g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0001.md": file(task("- `src/cli.go` contains `func save(`", claim))}))
		b, _ := g.Lookup("breadcrumbs/TK-0001.md")
		if b.Verdict != Refuted || len(b.Problems) != 1 {
			t.Errorf("%q: got %v with problems %q, want Refuted with one problem", claim, b.Verdict, b.Problems)
		}
	}
}

func TestATaskWithNoClaimsIsUndecided(t *testing.T) {
	g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0001.md": file(task())}))
	if v := verdicts(g)["TK-0001"]; v != "Undecided" {
		t.Errorf("got %s", v)
	}
}

func TestClaimsOnlyCountInTasks(t *testing.T) {
	g := load(t, with(fstest.MapFS{"breadcrumbs/TD-0001.md": file("# TD-0001: A decision\n\n## Claims\n\n- `src/cli.go` contains `nope`\n")}))
	b := only(t, g)
	if len(b.Claims) != 0 || b.Verdict != Undecided {
		t.Errorf("got %+v", b)
	}
}

// RQ-0006 and RQ-0008.
func TestVerdictsPassUpTheChain(t *testing.T) {
	holds := "- `src/cli.go` contains `func save(`"
	fails := "- `src/cli.go` contains `func Save(`"
	crumbs := func(second string) fstest.MapFS {
		return with(fstest.MapFS{
			"breadcrumbs/IN-0001.md": file("# IN-0001: An ask\n"),
			"breadcrumbs/RQ-0001.md": file("# RQ-0001: A requirement\n\nParent: [IN-0001](IN-0001.md)\n"),
			"breadcrumbs/TD-0001.md": file("# TD-0001: A decision\n\nParent: [RQ-0001](RQ-0001.md)\n"),
			"breadcrumbs/TK-0001.md": file(task(holds)),
			"breadcrumbs/TK-0002.md": file(strings.Replace(task(second), "TK-0001", "TK-0002", 1)),
		})
	}

	tests := []struct {
		name   string
		second string
		want   string
	}{
		{"all Confirmed", holds, "Confirmed"},
		{"one Refuted", fails, "Refuted"},
		{"one Undecided", "", "Undecided"},
	}
	for _, tt := range tests {
		g := load(t, crumbs(tt.second))
		v := verdicts(g)
		for _, id := range []string{"TD-0001", "RQ-0001", "IN-0001"} {
			if v[id] != tt.want {
				t.Errorf("%s: %s is %s, want %s", tt.name, id, v[id], tt.want)
			}
		}
	}
}

func TestATaskIsRefutedByARefutedChild(t *testing.T) {
	g := load(t, with(fstest.MapFS{
		"breadcrumbs/TK-0001.md": file(task("- `src/cli.go` contains `func save(`")),
		"breadcrumbs/TK-0002.md": file("# TK-0002: A follow-up\n\nParent: [TK-0001](TK-0001.md)\n\n## Claims\n\n- `src/cli.go` contains `nope`\n"),
	}))
	if v := verdicts(g)["TK-0001"]; v != "Refuted" {
		t.Errorf("got %s", v)
	}
}

func TestACycleDecidesNothing(t *testing.T) {
	g := load(t, fstest.MapFS{
		"breadcrumbs/TD-0001.md": file("# TD-0001: A\n\nParent: [TD-0002](TD-0002.md)\n"),
		"breadcrumbs/TD-0002.md": file("# TD-0002: B\n\nParent: [TD-0001](TD-0001.md)\n"),
	})
	for id, v := range verdicts(g) {
		if v != "Undecided" {
			t.Errorf("%s is %s", id, v)
		}
	}
}

func TestProblemsCarryTheirLines(t *testing.T) {
	g := load(t, with(fstest.MapFS{"breadcrumbs/TK-0002.md": file(
		"# TK-0002: A task\n\nParent: [TD-0001](TD-0001.md)\nParent: TD-0001\nParent: [TD-0099](TD-0099.md)\n\n## Claims\n\n- `src/cli.go` contains `func save(`\n- src/cli.go contains func save(\n")}))
	b, _ := g.Lookup("breadcrumbs/TK-0002.md")
	want := []Problem{
		{4, "not a Parent: link: Parent: TD-0001"},
		{5, "Parent: links to no breadcrumb: breadcrumbs/TD-0099.md"},
		{10, "not a claim: - src/cli.go contains func save("},
	}
	if !reflect.DeepEqual(b.Problems, want) {
		t.Errorf("got %+v, want %+v", b.Problems, want)
	}

	b = only(t, load(t, fstest.MapFS{"breadcrumbs/TD-0001.md": file("Adopt the method\n")}))
	if want := []Problem{{1, "the first line is not a # title"}}; !reflect.DeepEqual(b.Problems, want) {
		t.Errorf("got %+v, want %+v", b.Problems, want)
	}
}

// RQ-0022
func TestALinkToNoBreadcrumbIsAProblem(t *testing.T) {
	g := load(t, fstest.MapFS{
		"breadcrumbs/RQ-0001.md": file("# RQ-0001: A requirement\n\nParent: [IN-0001](IN-0001.md)\n"),
		"breadcrumbs/notes.md":   file("# Notes\n"),
		"breadcrumbs/RQ-0002.md": file("# RQ-0002: Another\n\nParent: [notes](notes.md)\n"),
	})
	for _, id := range []string{"RQ-0001", "RQ-0002"} {
		b, _ := g.Lookup("breadcrumbs/" + id + ".md")
		if len(b.Problems) != 1 || b.Problems[0].Line != 3 {
			t.Errorf("%s: got problems %+v, want one on line 3", id, b.Problems)
		}
	}
}
