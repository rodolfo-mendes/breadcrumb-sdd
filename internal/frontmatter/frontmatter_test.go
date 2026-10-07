package frontmatter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// file joins lines into the text of a file, each ended by \n.
func file(lines ...string) []byte { return []byte(strings.Join(lines, "\n") + "\n") }

// read reads src and fails on any problem, or when it has no breadcrumb.
func read(t *testing.T, src []byte) Breadcrumb {
	t.Helper()
	b, found, ps := Read(src)
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	if !found {
		t.Fatal("found no breadcrumb")
	}
	return b
}

// broken reads src and fails unless it has problems at want, and no
// breadcrumb.
func broken(t *testing.T, src []byte, want ...int) []crumb.Problem {
	t.Helper()
	b, found, ps := Read(src)
	var got []int
	for _, p := range ps {
		got = append(got, p.Line)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("problems at lines %v, want %v: %v\n%s", got, want, ps, src)
	}
	if !found || !reflect.DeepEqual(b, Breadcrumb{}) {
		t.Errorf("found %v, breadcrumb %+v; want found and no breadcrumb", found, b)
	}
	return ps
}

// pbi is the front matter of tasks/PBI-00001.md.
var pbi = []string{
	"---",
	"title: Adopt ASDLC",
	"breadcrumb:",
	"  id: PBI-00001",
	"  type: PBI",
	"  links:",
	"    - implements ADR-0001",
	"    - changes asdlc",
	"---",
	"# PBI-00001: Adopt ASDLC",
}

func TestABreadcrumbIsReadWithTheLinesOfTheFile(t *testing.T) {
	b := read(t, file(pbi...))
	want := Breadcrumb{
		Line: 3,
		Properties: map[string]crumb.Property{
			"id":   {Kind: crumb.TextValue, Text: "PBI-00001", Line: 4},
			"type": {Kind: crumb.TextValue, Text: "PBI", Line: 5},
			"links": {Kind: crumb.ListValue, Line: 6, Entries: []crumb.Entry{
				{Text: "implements ADR-0001", Line: 7},
				{Text: "changes asdlc", Line: 8},
			}},
		},
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("got %+v\nwant %+v", b, want)
	}
}

func TestWindowsLineEndingsAndAByteOrderMark(t *testing.T) {
	src := "\xEF\xBB\xBF" + strings.Join(pbi, "\r\n") + "\r\n"
	want := read(t, file(pbi...))
	if b := read(t, []byte(src)); !reflect.DeepEqual(b, want) {
		t.Errorf("got %+v\nwant %+v", b, want)
	}
}

func TestAnEmptyListAndOtherShapes(t *testing.T) {
	b := read(t, file("---", "breadcrumb:", "  id: 0001", "  type: true", "  links: []", "  status: ~", "  owner:", "---"))
	want := map[string]crumb.Property{
		"id":     {Kind: crumb.TextValue, Text: "0001", Line: 3},
		"type":   {Kind: crumb.TextValue, Text: "true", Line: 4},
		"links":  {Kind: crumb.ListValue, Line: 5},
		"status": {Kind: crumb.TextValue, Text: "~", Line: 6},
		"owner":  {Kind: crumb.NoValue, Line: 7},
	}
	if !reflect.DeepEqual(b.Properties, want) {
		t.Errorf("got %+v\nwant %+v", b.Properties, want)
	}
}

func TestABareBreadcrumbKeyHasNoProperties(t *testing.T) {
	b := read(t, file("---", "breadcrumb:", "---"))
	if b.Line != 2 || len(b.Properties) != 0 {
		t.Errorf("got %+v, want line 2 and no properties", b)
	}
}

func TestCommentsAreAllowed(t *testing.T) {
	b := read(t, file("---", "# about this file", "breadcrumb: # the breadcrumb", "  id: X # an id", "  links:", "    - implements Y # a link", "---"))
	if b.Properties["id"].Text != "X" || b.Properties["links"].Entries[0].Text != "implements Y" {
		t.Errorf("got %+v", b.Properties)
	}
}

func TestAFileWithNoBreadcrumb(t *testing.T) {
	for _, src := range [][]byte{
		file("# A title", "---", "breadcrumb:", "  id: X", "---"),
		file("---", "title: no breadcrumb", "---"),
		file("---", "---"),
		file("---", "- a list", "---"),
		file("---", "just text", "---"),
		file(" ---", "breadcrumb:", "---"),
		[]byte(""),
	} {
		if b, found, ps := Read(src); found || ps != nil || !reflect.DeepEqual(b, Breadcrumb{}) {
			t.Errorf("%q: got %+v, %v, %v; want no breadcrumb", src, b, found, ps)
		}
	}
}

func TestKeysOutsideTheBreadcrumbAreNotChecked(t *testing.T) {
	read(t, file("---", `title: "quoted"`, "tags: [a, b]", "base: &b", "  x: 1", "other: *b", "breadcrumb:", "  id: X", "---"))
}

func TestFrontMatterThatDoesNotClose(t *testing.T) {
	broken(t, file("---", "breadcrumb:", "  id: X"), 1)
	broken(t, []byte("---"), 1)
	broken(t, file("---", "breadcrumb:", "--- "), 1)
}

func TestFrontMatterThatIsNotYAML(t *testing.T) {
	ps := broken(t, file("---", "breadcrumb:", "  id: X", " type: Y", "---"), 4)
	if !strings.HasPrefix(ps[0].Message, "front matter is not YAML: ") {
		t.Errorf("message %q", ps[0].Message)
	}
	broken(t, file("---", "breadcrumb:", "  links: [a", "---"), 3)
	broken(t, file("---", "breadcrumb:", "  id: X", "  type: *nope", "---"), 4)
	broken(t, file("---", "breadcrumb:", "  id: X", "  type: @Y", "---"), 4)
}

func TestTheBreadcrumbKeyTwice(t *testing.T) {
	broken(t, file("---", "breadcrumb:", "  id: X", "breadcrumb:", "  id: Y", "---"), 4)
}

func TestAPropertyWrittenTwice(t *testing.T) {
	broken(t, file("---", "breadcrumb:", "  id: X", "  type: T", "  id: Y", "---"), 5)
}

func TestABreadcrumbThatIsNotAMap(t *testing.T) {
	broken(t, file("---", "breadcrumb: X", "---"), 2)
	broken(t, file("---", "breadcrumb:", "  - X", "---"), 3)
}

func TestYAMLOutsideThePartADR0002Allows(t *testing.T) {
	for _, c := range []struct {
		name  string
		lines []string
		want  []int
	}{
		{"anchor", []string{"  id: &a X"}, []int{3}},
		{"alias", []string{"  id: X", "  type: *a"}, []int{4}},
		{"tag", []string{"  id: !!str X"}, []int{3}},
		{"double quotes", []string{`  id: "X"`}, []int{3}},
		{"single quotes", []string{"  id: 'X'"}, []int{3}},
		{"quoted key", []string{`  "id": X`}, []int{3}},
		{"literal", []string{"  id: |", "    X"}, []int{3}},
		{"folded", []string{"  id: >", "    X"}, []int{3}},
		{"plain over two lines", []string{"  id: X", "    Y", "  type: T"}, []int{3}},
		{"flow list", []string{"  links: [implements X]"}, []int{3}},
		{"flow map", []string{"  links: {}"}, []int{3}},
		{"map as a value", []string{"  id:", "    x: 1"}, []int{3}},
		{"list in a list", []string{"  links:", "    - - implements X"}, []int{4}},
		{"quoted entry", []string{"  links:", `    - "implements X"`}, []int{4}},
		{"entry over two lines", []string{"  links:", "    - implements", "      X"}, []int{4}},
		{"two at once", []string{`  id: "X"`, "  links:", "    - &e implements Y"}, []int{3, 5}},
	} {
		t.Run(c.name, func(t *testing.T) {
			lines := append([]string{"---", "breadcrumb:"}, c.lines...)
			broken(t, file(append(lines, "---")...), c.want...)
		})
	}
	broken(t, file("---", "breadcrumb: &b", "  id: X", "---"), 2)
}

// claims returns a file whose breadcrumb has entries as its claims,
// the first on line 6.
func claims(entries ...string) []byte {
	lines := []string{"---", "breadcrumb:", "  id: X", "  type: T", "  claims:"}
	for _, e := range entries {
		lines = append(lines, "    - "+e)
	}
	return file(append(lines, "---")...)
}

// droppedLines returns the line of each problem in b.Dropped.
func droppedLines(b Breadcrumb) []int {
	var got []int
	for _, p := range b.Dropped {
		got = append(got, p.Line)
	}
	return got
}

func TestAClaimIsReadWithoutItsQuotes(t *testing.T) {
	b := read(t, claims(
		"'docs/bcr.md has-line ### verify'",
		"'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'  ",
		"'cmd/bcr/main.go has-line return fmt.Sprintf(''%s'', id)'",
		"''",
	))
	want := crumb.Property{Kind: crumb.ListValue, Line: 5, Entries: []crumb.Entry{
		{Text: "docs/bcr.md has-line ### verify", Line: 6},
		{Text: "docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted", Line: 7},
		{Text: "cmd/bcr/main.go has-line return fmt.Sprintf('%s', id)", Line: 8},
		{Text: "", Line: 9},
	}}
	if got := b.Properties["claims"]; !reflect.DeepEqual(got, want) || b.Dropped != nil {
		t.Errorf("got %+v, dropped %v\nwant %+v", got, b.Dropped, want)
	}
}

func TestAClaimOutsideTheFormatIsDroppedOnItsOwn(t *testing.T) {
	for _, c := range []struct{ name, entry string }{
		{"plain", "docs/bcr.md has-line verify"},
		{"plain, cut by a comment", "docs/bcr.md has-line ### verify"},
		{"plain, read as a map", "docs/bcr.md has-line Status: Accepted"},
		{"double quotes", `"docs/bcr.md has-line ### verify"`},
		{"comment", "'docs/bcr.md has-line ### verify' # why"},
		{"empty", ""},
		{"literal", "|\n      docs/bcr.md has-line x"},
		{"over two lines", "'docs/bcr.md has-line\n      x'"},
		{"anchor", "&a 'docs/bcr.md has-line x'"},
		{"tag", "!!str 'docs/bcr.md has-line x'"},
		{"list", "- 'docs/bcr.md has-line x'"},
		{"flow list", "['docs/bcr.md has-line x']"},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := read(t, claims("'a.md has-line first'", c.entry, "'a.md has-line last'"))
			if got := droppedLines(b); !reflect.DeepEqual(got, []int{7}) {
				t.Errorf("dropped at lines %v, want line 7: %v", got, b.Dropped)
			}
			var texts []string
			for _, e := range b.Properties["claims"].Entries {
				texts = append(texts, e.Text)
			}
			if want := []string{"a.md has-line first", "a.md has-line last"}; !reflect.DeepEqual(texts, want) {
				t.Errorf("got entries %q, want %q", texts, want)
			}
			if b.Properties["id"].Text != "X" {
				t.Errorf("got %+v, want the rest of the breadcrumb", b.Properties)
			}
		})
	}
}

func TestSingleQuotesAreAllowedOnlyInTheEntriesOfClaims(t *testing.T) {
	broken(t, file("---", "breadcrumb:", "  id: X", "  claims: 'a.md has-line x'", "---"), 4)
	broken(t, file("---", "breadcrumb:", "  id: X", "  'claims':", "    - 'a.md has-line x'", "---"), 4)
	broken(t, file("---", "breadcrumb:", "  id: X", "  links:", "    - 'implements Y'", "---"), 5)
}

func TestADroppedClaimIsAProblemWhenThereIsNoBreadcrumb(t *testing.T) {
	broken(t, file("---", "breadcrumb:", "  id: X", "  claims:", "    - a.md has-line x", "  type: 'T'", "---"), 5, 6)
}

func TestAnEmptyListOfClaims(t *testing.T) {
	b := read(t, file("---", "breadcrumb:", "  id: X", "  claims: []", "---"))
	if want := (crumb.Property{Kind: crumb.ListValue, Line: 4}); !reflect.DeepEqual(b.Properties["claims"], want) {
		t.Errorf("got %+v, want %+v", b.Properties["claims"], want)
	}
	b = read(t, file("---", "breadcrumb:", "  id: X", "  claims:", "---"))
	if want := (crumb.Property{Kind: crumb.NoValue, Line: 4}); !reflect.DeepEqual(b.Properties["claims"], want) {
		t.Errorf("got %+v, want %+v", b.Properties["claims"], want)
	}
}
