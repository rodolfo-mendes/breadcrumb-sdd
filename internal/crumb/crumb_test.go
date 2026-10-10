package crumb

import (
	"reflect"
	"strings"
	"testing"
)

func text(s string, line int) Property { return Property{Kind: TextValue, Text: s, Line: line} }

func list(line int, entries ...Entry) Property {
	return Property{Kind: ListValue, Entries: entries, Line: line}
}

// valid returns the properties of a breadcrumb that follows every rule,
// written at line 2, with its links on lines 6 and on.
func valid(links ...string) map[string]Property {
	var entries []Entry
	for i, l := range links {
		entries = append(entries, Entry{Text: l, Line: 6 + i})
	}
	return map[string]Property{
		"id":    text("PBI-00001", 3),
		"type":  text("PBI", 4),
		"links": list(5, entries...),
	}
}

// lines returns the line of each problem.
func lines(ps []Problem) []int {
	var got []int
	for _, p := range ps {
		got = append(got, p.Line)
	}
	return got
}

// broken reads props and fails unless the problems are at want, and no
// breadcrumb is returned.
func broken(t *testing.T, props map[string]Property, want ...int) []Problem {
	t.Helper()
	b, ok, ps := Read(2, props)
	if ok {
		t.Errorf("got ok with problems %v, want no breadcrumb", ps)
	}
	if !reflect.DeepEqual(lines(ps), want) {
		t.Errorf("problems at lines %v, want %v: %v", lines(ps), want, ps)
	}
	if !reflect.DeepEqual(b, Breadcrumb{}) {
		t.Errorf("got breadcrumb %+v with problems, want none", b)
	}
	return ps
}

func TestABreadcrumbWithLinks(t *testing.T) {
	b, _, ps := Read(2, valid("implements ADR-0001", "changes asdlc"))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	want := Breadcrumb{
		ID:   "PBI-00001",
		Type: "PBI",
		Line: 3,
		Links: []Link{
			{Verb: "implements", Object: "ADR-0001", Line: 6},
			{Verb: "changes", Object: "asdlc", Line: 7},
		},
	}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("got %+v, want %+v", b, want)
	}
}

func TestABreadcrumbWithNoLinks(t *testing.T) {
	b, _, ps := Read(2, valid())
	if ps != nil || b.ID != "PBI-00001" || len(b.Links) != 0 {
		t.Errorf("got %+v and problems %v, want PBI-00001 with no links", b, ps)
	}
}

func TestABreadcrumbWithATitle(t *testing.T) {
	props := valid()
	props["title"] = text("Adopt ASDLC to develop Breadcrumb", 5)
	b, ok, ps := Read(2, props)
	if !ok || ps != nil || b.Title != "Adopt ASDLC to develop Breadcrumb" {
		t.Errorf("got %+v, %v and problems %v, want the title as written", b, ok, ps)
	}
}

func TestATitleIsOptional(t *testing.T) {
	b, ok, ps := Read(2, valid())
	if !ok || ps != nil || b.Title != "" {
		t.Errorf("got %+v, %v and problems %v, want a breadcrumb with no title", b, ok, ps)
	}
}

func TestATitleThatIsNotOneLineOfText(t *testing.T) {
	for _, p := range []Property{
		{Kind: NoValue, Line: 5},
		text("", 5),
		list(5),
		list(5, Entry{Text: "Adopt ASDLC", Line: 6}),
		text("Adopt\tASDLC", 5),
	} {
		props := valid()
		props["title"] = p
		broken(t, props, 5)
	}
}

func TestOtherPropertiesAreNotRead(t *testing.T) {
	props := valid()
	props["status"] = Property{Kind: NoValue, Line: 9}
	if _, _, ps := Read(2, props); ps != nil {
		t.Errorf("got problems %v", ps)
	}
}

func TestAValueIsKeptAsWritten(t *testing.T) {
	props := valid()
	props["id"] = text("0001", 3)
	if b, _, _ := Read(2, props); b.ID != "0001" {
		t.Errorf("id %q, want 0001", b.ID)
	}
}

func TestAMissingPropertyIsAProblemAtTheBreadcrumb(t *testing.T) {
	for _, name := range []string{"id", "type", "links"} {
		props := valid()
		delete(props, name)
		broken(t, props, 2)
	}
	broken(t, map[string]Property{}, 2, 2, 2)
}

func TestAnIdOrATypeThatIsNotOneWord(t *testing.T) {
	for _, p := range []Property{
		{Kind: NoValue, Line: 3},
		text("", 3),
		list(3, Entry{Text: "PBI-00001", Line: 4}),
		text("PBI 00001", 3),
		text("PBI\t00001", 3),
		text("PBI 00001", 3),
	} {
		props := valid()
		props["id"] = p
		broken(t, props, 3)

		props = valid()
		props["type"] = p
		broken(t, props, 3)
	}
}

func TestLinksWithNoList(t *testing.T) {
	for _, p := range []Property{{Kind: NoValue, Line: 5}, text("implements ADR-0001", 5)} {
		props := valid()
		props["links"] = p
		broken(t, props, 5)
	}
}

func TestWhiteSpaceInALinkEntry(t *testing.T) {
	b, _, ps := Read(2, valid("  implements \t  ADR-0001 "))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	want := []Link{{Verb: "implements", Object: "ADR-0001", Line: 6}}
	if !reflect.DeepEqual(b.Links, want) {
		t.Errorf("got %+v, want %+v", b.Links, want)
	}
}

func TestALinkEntryThatIsNotTwoWords(t *testing.T) {
	broken(t, valid("ADR-0001", "implements ADR-0002", "depends on ADR-0003", ""), 6, 8, 9)
}

func TestTheSameLinkTwice(t *testing.T) {
	broken(t, valid("implements ADR-0001", "changes asdlc", "implements  ADR-0001", "implements ADR-0001 "), 8, 9)
}

func TestALinkToItself(t *testing.T) {
	broken(t, valid("implements ADR-0001", "implements PBI-00001"), 7)
}

func TestALinkToItselfIsNotFoundWithoutAnId(t *testing.T) {
	props := valid("implements PBI 00001")
	props["id"] = text("PBI 00001", 3)
	ps := broken(t, props, 3, 6)
	if ps[1].Message != `link "implements PBI 00001" must be a verb and an id` {
		t.Errorf("got %q", ps[1].Message)
	}
}

func TestProblemsAreInOrderOfLine(t *testing.T) {
	props := valid("one", "implements PBI-00001")
	props["type"] = text("a b", 4)
	delete(props, "id")
	broken(t, props, 2, 4, 6)
}

// claimed returns the properties of a valid breadcrumb with claims,
// written on lines 8 and on.
func claimed(claims ...string) map[string]Property {
	var entries []Entry
	for i, c := range claims {
		entries = append(entries, Entry{Text: c, Line: 8 + i})
	}
	props := valid("implements ADR-0001")
	props["claims"] = list(7, entries...)
	return props
}

// dropped reads props and fails unless the breadcrumb is built, with
// problems at want and the claims whose arguments are kept.
func dropped(t *testing.T, props map[string]Property, want []int, kept ...string) []Problem {
	t.Helper()
	b, ok, ps := Read(2, props)
	if !ok || b.ID != "PBI-00001" || len(b.Links) != 1 {
		t.Errorf("got ok %v and %+v, want the breadcrumb and its link", ok, b)
	}
	if !reflect.DeepEqual(lines(ps), want) {
		t.Errorf("problems at lines %v, want %v: %v", lines(ps), want, ps)
	}
	var got []string
	for _, c := range b.Claims {
		got = append(got, c.Argument)
	}
	if !reflect.DeepEqual(got, kept) {
		t.Errorf("kept the claims with arguments %q, want %q", got, kept)
	}
	return ps
}

func TestABreadcrumbWithClaims(t *testing.T) {
	b, ok, ps := Read(2, claimed(
		"docs/bcr.md has-line ### verify",
		"docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted",
		"cmd/bcr/main.go has-line return  fmt.Sprintf('%s', id)",
	))
	if !ok || ps != nil {
		t.Fatalf("got ok %v and problems %v", ok, ps)
	}
	want := []Claim{
		{Target: "docs/bcr.md", Kind: "has-line", Argument: "### verify", Line: 8},
		{Target: "docs/adrs/ADR-0001-adopt-asdlc.md", Kind: "has-line", Argument: "Status: Accepted", Line: 9},
		{Target: "cmd/bcr/main.go", Kind: "has-line", Argument: "return  fmt.Sprintf('%s', id)", Line: 10},
	}
	if !reflect.DeepEqual(b.Claims, want) {
		t.Errorf("got %+v, want %+v", b.Claims, want)
	}
}

func TestClaimsAreOptional(t *testing.T) {
	for _, props := range []map[string]Property{valid(), claimed()} {
		if b, ok, ps := Read(2, props); !ok || ps != nil || b.Claims != nil {
			t.Errorf("got ok %v, claims %+v and problems %v; want a breadcrumb with no claims", ok, b.Claims, ps)
		}
	}
}

func TestClaimsWithNoList(t *testing.T) {
	for _, p := range []Property{{Kind: NoValue, Line: 7}, text("docs/bcr.md has-line x", 7)} {
		props := valid()
		props["claims"] = p
		broken(t, props, 7)
	}
}

func TestAClaimEntryThatIsNotThreeParts(t *testing.T) {
	dropped(t, claimed(
		"docs/bcr.md",
		"docs/bcr.md has-line",
		"docs/bcr.md  has-line ### verify",
		" docs/bcr.md has-line ### verify",
		"",
		"docs/bcr.md has-line x",
	), []int{8, 9, 10, 11, 12}, "x")
}

func TestATargetOutsideTheRules(t *testing.T) {
	dropped(t, claimed(
		"/etc/passwd has-line x",
		"../x.md has-line x",
		"docs/../x.md has-line x",
		"docs/.. has-line x",
		"docs/a\u00a0b.md has-line x",
		"docs/a..b.md has-line kept",
		"docs/.../x.md has-line kept too",
	), []int{8, 9, 10, 11, 12}, "kept", "kept too")
}

func TestAKindTheCoreDoesNotKnow(t *testing.T) {
	ps := dropped(t, claimed(
		"docs/bcr.md contains ### verify",
		"docs/bcr.md has-lines ### verify",
		"docs/bcr.md Has-Line ### verify",
	), []int{8, 9, 10})
	for i, kind := range []string{`"contains"`, `"has-lines"`, `"Has-Line"`} {
		if i < len(ps) && !strings.Contains(ps[i].Message, kind) {
			t.Errorf("problem %q does not name the kind %s", ps[i].Message, kind)
		}
	}
}

func TestAHasLineTextThatNoLineCanEqual(t *testing.T) {
	dropped(t, claimed(
		"docs/bcr.md has-line ",
		"docs/bcr.md has-line  ### verify",
		"docs/bcr.md has-line ### verify ",
		"docs/bcr.md has-line a\tb",
		"docs/bcr.md has-line \t### verify",
		"docs/bcr.md has-line ### verify\t",
		"docs/bcr.md has-line a  b",
	), []int{8, 9, 10, 11, 12, 13}, "a  b")
}

func TestAnInvalidClaimDropsOnlyItself(t *testing.T) {
	dropped(t, claimed("a.md has-line one", "a.md contains two", "a.md has-line three"), []int{9}, "one", "three")
}

func TestAProblemInALinkDropsTheClaimsToo(t *testing.T) {
	props := claimed("a.md has-line one", "a.md contains two")
	props["links"] = list(5, Entry{Text: "ADR-0001", Line: 6})
	broken(t, props, 6, 9)
}
