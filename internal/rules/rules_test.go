package rules

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// The tests below follow the rules of specs/verify/spec.md about the
// format of breadcrumb.rules.

// text joins lines as a breadcrumb.rules.
func text(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

// parsed parses lines as a breadcrumb.rules, and fails on any problem.
func parsed(t *testing.T, lines ...string) crumb.Rules {
	t.Helper()
	r, ps := Parse(text(lines...))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	return r
}

// refused parses lines as a breadcrumb.rules, and fails unless it
// returns no rules and the problems want.
func refused(t *testing.T, lines []string, want ...crumb.Problem) {
	t.Helper()
	r, ps := Parse(text(lines...))
	if !reflect.DeepEqual(r, crumb.Rules{}) {
		t.Errorf("got rules %+v, want none", r)
	}
	if !reflect.DeepEqual(ps, want) {
		t.Errorf("got problems %v, want %v", ps, want)
	}
}

// layout is the rules the lines of TestTheThreeKindsOfLine declare.
var layout = crumb.Rules{
	Name:  "breadcrumb.rules",
	Types: []crumb.RuleType{{Name: "ADR", Line: 2}, {Name: "spec", Line: 3}, {Name: "PBI", Line: 4}},
	Links: []crumb.RuleLink{
		{Subject: "ADR", Verb: "constrained_by", Object: "ADR", Line: 5},
		{Subject: "PBI", Verb: "changes", Object: "spec", Line: 6},
	},
	Claims: []crumb.RuleClaims{{Type: "spec", Line: 7}},
}

func TestTheThreeKindsOfLine(t *testing.T) {
	got := parsed(t,
		"# The ASDLC layout",
		"type\tADR",
		"type\tspec",
		"type\tPBI",
		"link\tADR\tconstrained_by\tADR",
		"link\tPBI\tchanges\tspec",
		"claims\tspec",
	)
	if !reflect.DeepEqual(got, layout) {
		t.Errorf("got %+v, want %+v", got, layout)
	}
}

func TestLinesInAnyOrder(t *testing.T) {
	got := parsed(t, "claims\tspec", "link\tPBI\tchanges\tspec", "type\tspec", "type\tPBI")
	want := crumb.Rules{
		Name:   Name,
		Types:  []crumb.RuleType{{Name: "spec", Line: 3}, {Name: "PBI", Line: 4}},
		Links:  []crumb.RuleLink{{Subject: "PBI", Verb: "changes", Object: "spec", Line: 2}},
		Claims: []crumb.RuleClaims{{Type: "spec", Line: 1}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestBlankLinesCommentsAndLineEndings(t *testing.T) {
	r, ps := Parse([]byte("\xEF\xBB\xBF# the layout\r\ntype\tspec\r\n\r\n  \t\r\n#type\tPBI\r\nclaims\tspec"))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	want := crumb.Rules{
		Name:   Name,
		Types:  []crumb.RuleType{{Name: "spec", Line: 2}},
		Claims: []crumb.RuleClaims{{Type: "spec", Line: 6}},
	}
	if !reflect.DeepEqual(r, want) {
		t.Errorf("got %+v, want %+v", r, want)
	}
}

func TestAnEmptyFileDeclaresNothing(t *testing.T) {
	for _, src := range []string{"", "\n", "# nothing yet\n"} {
		r, ps := Parse([]byte(src))
		if ps != nil || !reflect.DeepEqual(r, crumb.Rules{Name: Name}) {
			t.Errorf("%q: got %+v, %v; want rules that declare nothing", src, r, ps)
		}
	}
}

func TestNamesAndVerbsAreExactBytes(t *testing.T) {
	got := parsed(t, "type\tspec", "type\tSpec", "type\tação")
	if len(got.Types) != 3 || got.Types[1].Name != "Spec" || got.Types[2].Name != "ação" {
		t.Errorf("got %+v", got.Types)
	}
}

func TestAnUnknownKindOfLine(t *testing.T) {
	refused(t, []string{"type\tspec", "lnik\tPBI\tchanges\tspec", "Type\tPBI", "\ttype\tPBI", "types spec"},
		crumb.Problem{Line: 2, Message: `unknown kind of line "lnik"; a line is type, link or claims`},
		crumb.Problem{Line: 3, Message: `unknown kind of line "Type"; a line is type, link or claims`},
		crumb.Problem{Line: 4, Message: `unknown kind of line ""; a line is type, link or claims`},
		crumb.Problem{Line: 5, Message: `unknown kind of line "types"; a line is type, link or claims`},
	)
}

func TestTheWrongNumberOfFields(t *testing.T) {
	refused(t, []string{"type", "type\tspec\tPBI", "link\tPBI\tchanges", "link\tPBI\tchanges\tspec\tADR", "claims", "claims\tspec\t"},
		crumb.Problem{Line: 1, Message: "type has 1 field, needs 2"},
		crumb.Problem{Line: 2, Message: "type has 3 fields, needs 2"},
		crumb.Problem{Line: 3, Message: "link has 3 fields, needs 4"},
		crumb.Problem{Line: 4, Message: "link has 5 fields, needs 4"},
		crumb.Problem{Line: 5, Message: "claims has 1 field, needs 2"},
		crumb.Problem{Line: 6, Message: "claims has 3 fields, needs 2"},
	)
}

func TestFieldsSeparatedBySpaces(t *testing.T) {
	refused(t, []string{"link PBI changes spec", "type spec", "link\tPBI changes spec"},
		crumb.Problem{Line: 1, Message: "link has 1 field, needs 4; its fields are separated by spaces, use a tab"},
		crumb.Problem{Line: 2, Message: "type has 1 field, needs 2; its fields are separated by spaces, use a tab"},
		crumb.Problem{Line: 3, Message: "link has 2 fields, needs 4; its fields are separated by spaces, use a tab"},
	)
}

func TestACommentAfterAFieldIsAField(t *testing.T) {
	refused(t, []string{"type\tspec\t# a feature", "type spec # a feature"},
		crumb.Problem{Line: 1, Message: "type has 3 fields, needs 2"},
		crumb.Problem{Line: 2, Message: "type has 1 field, needs 2"},
	)
}

func TestAnEmptyFieldAndWhiteSpaceInAField(t *testing.T) {
	refused(t, []string{"link\tspec\t\tspec", "link\tspec\tbuilt on\tspec", "type\t", "type \tspec", "claims\tspec ", "type\tthe spec"},
		crumb.Problem{Line: 1, Message: "link has an empty field 3"},
		crumb.Problem{Line: 2, Message: "link has white space in field 3"},
		crumb.Problem{Line: 3, Message: "type has an empty field 2"},
		crumb.Problem{Line: 4, Message: "type has white space in field 1"},
		crumb.Problem{Line: 5, Message: "claims has white space in field 2"},
		crumb.Problem{Line: 6, Message: "type has white space in field 2"},
	)
}

func TestEveryProblemIsReturnedAndNoRules(t *testing.T) {
	refused(t, []string{"type\tspec", "lnik\tPBI\tchanges\tspec", "", "claims"},
		crumb.Problem{Line: 2, Message: `unknown kind of line "lnik"; a line is type, link or claims`},
		crumb.Problem{Line: 4, Message: "claims has 1 field, needs 2"},
	)
}

func TestWhatTheRulesSayIsNotCheckedHere(t *testing.T) {
	// A type no line declares and a line written twice are the core's.
	parsed(t, "link\tPBI\tchanges\tspec", "type\tspec", "type\tspec")
}
