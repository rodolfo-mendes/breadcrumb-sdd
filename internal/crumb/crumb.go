// Package crumb is the core of Breadcrumb: what a breadcrumb holds,
// and the rules it must follow. It knows no file format (ADR-0013).
//
// Whatever reads a file gives the core a breadcrumb as it is written:
// a Property for each name, holding text or a list of Entries, each
// value as the text written, with its line. Read returns a Breadcrumb
// that follows every rule, or the Problems that stop it. A claim that
// breaks a rule is left out on its own (ADR-0017).
package crumb

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Kind is the shape of a property's value.
type Kind int

const (
	NoValue   Kind = iota // the property is written with no value
	TextValue             // one text
	ListValue             // a list of texts, possibly empty
)

// Property is a property of a breadcrumb, as written. Read is given
// one for each name, such as id or links.
type Property struct {
	Kind    Kind
	Text    string  // when Kind is TextValue
	Entries []Entry // when Kind is ListValue, in the order written
	Line    int     // the line of the property, counted from 1
}

// Entry is one text of a list, as written, such as an entry of links.
type Entry struct {
	Text string
	Line int
}

// Breadcrumb is a breadcrumb that follows every rule (ADR-0002). It is
// built only from properties that broke none.
type Breadcrumb struct {
	ID    string
	Type  string
	Line  int    // the line of its id, counted from 1
	Links []Link // in the order written
	// Claims holds the claims that broke no rule, in the order written.
	Claims []Claim
}

// Link is one entry of a breadcrumb's links: a verb and the id of the
// breadcrumb it points to (ADR-0012).
type Link struct {
	Verb   string
	Object string
	Line   int
}

// Claim is one entry of a breadcrumb's claims: a statement of a kind,
// with its argument, about the file at Target (ADR-0017). Whether the
// target supports it is not known here: Verdict says, given the
// target's content.
type Claim struct {
	Target   string // a path from the root of the repository, with / between its parts
	Kind     string
	Argument string // as written, byte for byte
	Line     int
}

// Problem is a rule a breadcrumb breaks, at the line where it does.
type Problem struct {
	Line    int // counted from 1
	Message string
}

// Read builds the breadcrumb whose properties are props, written at
// line. ok is false when id, type, links or the claims key breaks a
// rule: there is then no breadcrumb. A claim that breaks a rule is
// dropped, and the breadcrumb is built without it (ADR-0017). problems
// holds every problem of both kinds, in order of line. Properties
// other than id, type, links and claims are not read.
func Read(line int, props map[string]Property) (b Breadcrumb, ok bool, problems []Problem) {
	var r reader
	id, idOK := r.word(line, props, "id")
	typ, _ := r.word(line, props, "type")
	links := r.links(line, props, id, idOK)
	claims := r.claims(props)
	ok = len(r.problems) == 0
	problems = append(r.problems, r.dropped...)
	sort.SliceStable(problems, func(i, j int) bool { return problems[i].Line < problems[j].Line })
	if !ok {
		return Breadcrumb{}, false, problems
	}
	return Breadcrumb{ID: id, Type: typ, Line: props["id"].Line, Links: links, Claims: claims}, true, problems
}

// reader gathers the problems of one breadcrumb.
type reader struct {
	problems []Problem // those that leave no breadcrumb
	dropped  []Problem // those of the claims left out
}

func (r *reader) report(line int, format string, args ...any) {
	r.problems = append(r.problems, Problem{Line: line, Message: fmt.Sprintf(format, args...)})
}

// drop adds the problem of a claim that is left out.
func (r *reader) drop(line int, format string, args ...any) {
	r.dropped = append(r.dropped, Problem{Line: line, Message: fmt.Sprintf(format, args...)})
}

// word reads a property whose value is one word, such as id and type
// (ADR-0002), and reports whether it could.
func (r *reader) word(line int, props map[string]Property, name string) (string, bool) {
	p, ok := props[name]
	switch {
	case !ok:
		r.report(line, "breadcrumb has no %s", name)
	case p.Kind == ListValue:
		r.report(p.Line, "%s is a list; it must be one word", name)
	case p.Kind == NoValue || p.Text == "":
		r.report(p.Line, "%s has no value", name)
	case strings.IndexFunc(p.Text, unicode.IsSpace) >= 0:
		r.report(p.Line, "%s %q has white space; it must be one word", name, p.Text)
	default:
		return p.Text, true
	}
	return "", false
}

// links reads the links of the breadcrumb whose id is id; idOK says
// whether id could be read (ADR-0002, ADR-0012).
func (r *reader) links(line int, props map[string]Property, id string, idOK bool) []Link {
	p, ok := props["links"]
	switch {
	case !ok:
		r.report(line, "breadcrumb has no links")
		return nil
	case p.Kind == NoValue:
		r.report(p.Line, "links has no list; a breadcrumb with no links has an empty list")
		return nil
	case p.Kind == TextValue:
		r.report(p.Line, "links is text; it must be a list")
		return nil
	}
	type key struct{ verb, object string }
	seen := map[key]bool{}
	var links []Link
	for _, e := range p.Entries {
		words := strings.FieldsFunc(e.Text, unicode.IsSpace)
		if len(words) != 2 {
			r.report(e.Line, "link %q must be a verb and an id", e.Text)
			continue
		}
		l := Link{Verb: words[0], Object: words[1], Line: e.Line}
		k := key{l.Verb, l.Object}
		switch {
		case seen[k]:
			r.report(e.Line, "link %q is written twice", e.Text)
		case idOK && l.Object == id:
			r.report(e.Line, "link %q points to its own breadcrumb", e.Text)
		}
		seen[k] = true
		links = append(links, l)
	}
	return links
}

// claimKind is the rule of one kind of claim, kept in one place for
// the reading of a claim and for its verdict (ADR-0019).
type claimKind struct {
	// argument returns what is wrong with an argument, or "".
	argument func(argument string) string
	// holds reports whether a claim with argument holds in content,
	// the content of its target.
	holds func(content []byte, argument string) bool
	// lacks says what a target lacks when the claim does not hold.
	lacks func(argument string) string
}

// claimKinds holds the kinds of claim the core knows. The only kind is
// has-line (ADR-0019).
var claimKinds = map[string]claimKind{
	"has-line": {argument: hasLineArgument, holds: hasLine, lacks: lacksLine},
}

// hasLineArgument checks the argument of a has-line claim: the text a
// line of the target must equal once the spaces and tabs at its start
// and end are removed (ADR-0019). No line can equal a text that is
// empty, or that starts or ends with a space or a tab; a tab would
// also break the record the claim is printed as.
func hasLineArgument(text string) string {
	switch {
	case text == "":
		return "has no text after has-line"
	case strings.ContainsRune(text, '\t'):
		return "has a tab in its text; no has-line text may hold one"
	case strings.HasPrefix(text, " ") || strings.HasSuffix(text, " "):
		return "has a text that starts or ends with a space; no line can equal it"
	}
	return ""
}

// claims reads the claims of the breadcrumb. claims is optional
// (ADR-0018). An entry is TARGET KIND ARGUMENT, one space apart, and
// one that breaks a rule is dropped with one problem (ADR-0017).
func (r *reader) claims(props map[string]Property) []Claim {
	p, ok := props["claims"]
	switch {
	case !ok:
		return nil
	case p.Kind == NoValue:
		r.report(p.Line, "claims has no list; leave it out, or write an empty list")
		return nil
	case p.Kind == TextValue:
		r.report(p.Line, "claims is text; it must be a list")
		return nil
	}
	var claims []Claim
	for _, e := range p.Entries {
		target, rest, _ := strings.Cut(e.Text, " ")
		kind, argument, three := strings.Cut(rest, " ")
		if !three || target == "" || kind == "" {
			r.drop(e.Line, "claim %q must be a target, a kind and an argument, with one space between them", e.Text)
			continue
		}
		c := Claim{Target: target, Kind: kind, Argument: argument, Line: e.Line}
		if why := c.Check(); why != "" {
			r.drop(e.Line, "claim %q %s", e.Text, why)
			continue
		}
		claims = append(claims, c)
	}
	return claims
}

// Check returns what is wrong with the target, the kind or the
// argument of c, or "" (ADR-0017). It is the first rule c breaks, as
// the words that follow "claim ...".
func (c Claim) Check() string {
	kind, known := claimKinds[c.Kind]
	switch {
	case strings.HasPrefix(c.Target, "/"):
		return "has a target that starts with /; it must be a path from the root of the repository"
	case hasPart(c.Target, ".."):
		return "has a target with a part that is .."
	case strings.IndexFunc(c.Target, unicode.IsSpace) >= 0:
		return "has a target with white space"
	case !known:
		return fmt.Sprintf("has kind %q, which bcr does not know; the only kind is has-line", c.Kind)
	}
	return kind.argument(c.Argument)
}

// hasPart reports whether part is one of the parts of path, which has
// / between them.
func hasPart(path, part string) bool {
	for _, p := range strings.Split(path, "/") {
		if p == part {
			return true
		}
	}
	return false
}
