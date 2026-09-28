package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/breadcrumb"
)

func crumb(id string, parents ...string) breadcrumb.Breadcrumb {
	b := breadcrumb.Breadcrumb{ID: id, Type: breadcrumb.Type(id[:2]), Path: "breadcrumbs/" + id + ".md", Title: "Title of " + id}
	for _, p := range parents {
		b.Parents = append(b.Parents, breadcrumb.Link{Text: p, Target: "breadcrumbs/" + p + ".md"})
	}
	return b
}

func find(t *testing.T, pg page, header string) *node {
	t.Helper()
	for _, n := range pg.Nodes {
		if n.Header == header {
			return n
		}
	}
	t.Fatalf("no node %s", header)
	return nil
}

func TestEveryBreadcrumbIsANode(t *testing.T) {
	pg := build(breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{crumb("IN-0001"), crumb("TD-0001")}})
	if len(pg.Nodes) != 2 {
		t.Fatalf("got %d nodes, want 2", len(pg.Nodes))
	}
	if n := find(t, pg, "TD-0001"); n.Class != "undecided" {
		t.Errorf("got class %q, want undecided", n.Class)
	}
}

func TestAChildSitsBelowItsParentWithAnArrowToIt(t *testing.T) {
	pg := build(breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{crumb("IN-0001"), crumb("RQ-0001", "IN-0001")}})
	in, rq := find(t, pg, "IN-0001"), find(t, pg, "RQ-0001")
	if rq.Y <= in.Y+in.H {
		t.Errorf("RQ-0001 at y=%d is not below IN-0001 ending at y=%d", rq.Y, in.Y+in.H)
	}
	want := edge{X1: rq.X + rq.W/2, Y1: rq.Y, X2: in.X + in.W/2, Y2: in.Y + in.H}
	if len(pg.Edges) != 1 || pg.Edges[0] != want {
		t.Errorf("got edges %+v, want %+v", pg.Edges, want)
	}
}

func TestALinkToNoBreadcrumbDrawsAMissingNode(t *testing.T) {
	pg := build(breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{crumb("RQ-0001", "IN-0009")}})
	n := find(t, pg, "IN-0009")
	if n.Class != "missing" {
		t.Errorf("got class %q, want missing", n.Class)
	}
	if len(pg.Edges) != 1 {
		t.Errorf("got %d edges, want 1", len(pg.Edges))
	}
}

func TestACycleIsDrawnRatherThanLooping(t *testing.T) {
	pg := build(breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{crumb("TD-0001", "TD-0002"), crumb("TD-0002", "TD-0001")}})
	if len(pg.Nodes) != 2 || len(pg.Edges) != 2 {
		t.Errorf("got %d nodes and %d edges, want 2 and 2", len(pg.Nodes), len(pg.Edges))
	}
}

func TestNodesDoNotOverlap(t *testing.T) {
	g := breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{
		crumb("IN-0001"), crumb("IN-0002"),
		crumb("RQ-0001", "IN-0001"), crumb("RQ-0002", "IN-0001"), crumb("RQ-0003", "IN-0002"),
		crumb("TD-0001"), crumb("TD-0002", "RQ-0001", "IN-0002"),
	}}
	pg := build(g)
	for i, a := range pg.Nodes {
		for _, b := range pg.Nodes[i+1:] {
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Errorf("%s and %s overlap", a.Header, b.Header)
			}
		}
		if a.X+a.W > pg.Width || a.Y+a.H > pg.Height {
			t.Errorf("%s is outside the drawing", a.Header)
		}
	}
}

func TestLongTitlesWrapToThreeLines(t *testing.T) {
	got := wrap(strings.Repeat("word ", 40))
	if len(got) != titleLines || !strings.HasSuffix(got[titleLines-1], "…") {
		t.Errorf("got %q", got)
	}
	for _, l := range got {
		if len([]rune(l)) > wrapAt+1 {
			t.Errorf("line %q is longer than %d", l, wrapAt)
		}
	}
}

func TestTheReportListsProblemsAndOtherFiles(t *testing.T) {
	b := crumb("TD-0001")
	b.Problems = []breadcrumb.Problem{{Line: 3, Message: "not a Parent: link: Parent: x"}}
	var out bytes.Buffer
	if err := Write(&out, breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{b}, Others: []string{"breadcrumbs/notes.md"}}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"breadcrumbs/TD-0001.md:3</code>: not a Parent: link: Parent: x", "breadcrumbs/notes.md", "TD-0001"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("report does not contain %q", want)
		}
	}
}

func TestATaskListsItsClaims(t *testing.T) {
	b := crumb("TK-0001")
	b.Verdict = breadcrumb.Refuted
	b.Claims = []breadcrumb.Claim{{Path: "a.go", Text: "func a(", Holds: true}, {Path: "b.go", Text: "func b(", Holds: false}}
	pg := build(breadcrumb.Graph{Breadcrumbs: []breadcrumb.Breadcrumb{b}})
	n := find(t, pg, "TK-0001")
	lines := strings.Join(n.Lines, "\n")
	for _, want := range []string{"✓ a.go: func a(", "✗ b.go: func b("} {
		if !strings.Contains(lines, want) {
			t.Errorf("lines %q do not contain %q", n.Lines, want)
		}
	}
	if n.Class != "refuted" {
		t.Errorf("got class %q", n.Class)
	}
}

func TestManyClaimsAreCut(t *testing.T) {
	var claims []breadcrumb.Claim
	for i := 0; i < maxClaims+3; i++ {
		claims = append(claims, breadcrumb.Claim{Path: "a.go", Text: strings.Repeat("x", 80), Holds: true})
	}
	lines := claimLines(claims)
	if len(lines) != maxClaims+1 || lines[maxClaims] != "… 3 more claims" {
		t.Errorf("got %q", lines)
	}
	if len([]rune(lines[0])) > wrapAt {
		t.Errorf("line %q is longer than %d", lines[0], wrapAt)
	}
}
