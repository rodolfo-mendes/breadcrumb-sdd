package breadcrumb

import (
	"reflect"
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
	want := []string{"IN-0001 Intake", "RQ-0001 Requirement", "TD-0001 Technical Decision"}
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
		{Text: "IN-0001", Target: "breadcrumbs/IN-0001.md"},
		{Text: "TD-0002", Target: "breadcrumbs/TD-0002.md"},
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
	b := only(t, load(t, fstest.MapFS{"breadcrumbs/RQ-0001.md": file(
		"# RQ-0001: A requirement\r\n\r\nParent: [IN-0001](IN-0001.md)\r\n")}))
	if b.Title != "A requirement" || len(b.Parents) != 1 || len(b.Problems) != 0 {
		t.Errorf("got %+v", b)
	}
}

func TestEveryBreadcrumbIsUndecided(t *testing.T) {
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
