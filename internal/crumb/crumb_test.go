package crumb

import (
	"reflect"
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
	b, ps := Read(2, props)
	if !reflect.DeepEqual(lines(ps), want) {
		t.Errorf("problems at lines %v, want %v: %v", lines(ps), want, ps)
	}
	if !reflect.DeepEqual(b, Breadcrumb{}) {
		t.Errorf("got breadcrumb %+v with problems, want none", b)
	}
	return ps
}

func TestABreadcrumbWithLinks(t *testing.T) {
	b, ps := Read(2, valid("implements ADR-0001", "changes asdlc"))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	want := Breadcrumb{
		ID:   "PBI-00001",
		Type: "PBI",
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
	b, ps := Read(2, valid())
	if ps != nil || b.ID != "PBI-00001" || len(b.Links) != 0 {
		t.Errorf("got %+v and problems %v, want PBI-00001 with no links", b, ps)
	}
}

func TestOtherPropertiesAreNotRead(t *testing.T) {
	props := valid()
	props["status"] = Property{Kind: NoValue, Line: 9}
	if _, ps := Read(2, props); ps != nil {
		t.Errorf("got problems %v", ps)
	}
}

func TestAValueIsKeptAsWritten(t *testing.T) {
	props := valid()
	props["id"] = text("0001", 3)
	if b, _ := Read(2, props); b.ID != "0001" {
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
	b, ps := Read(2, valid("  implements \t  ADR-0001 "))
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
