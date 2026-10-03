// Package crumb is the core of Breadcrumb: what a breadcrumb holds,
// and the rules it must follow. It knows no file format (ADR-0013).
//
// Whatever reads a file gives the core a breadcrumb as it is written:
// a Property for each name, holding text or a list of Entries, each
// value as the text written, with its line. Read returns a Breadcrumb
// that follows every rule, or the Problems that stop it.
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
	Links []Link // in the order written
}

// Link is one entry of a breadcrumb's links: a verb and the id of the
// breadcrumb it points to (ADR-0012).
type Link struct {
	Verb   string
	Object string
	Line   int
}

// Problem is a rule a breadcrumb breaks, at the line where it does.
type Problem struct {
	Line    int // counted from 1
	Message string
}

// Read builds the breadcrumb whose properties are props, written at
// line. It returns the breadcrumb, or, when a rule is broken, no
// breadcrumb and every problem, in order of line. Properties other
// than id, type and links are not read.
func Read(line int, props map[string]Property) (Breadcrumb, []Problem) {
	var r reader
	id, idOK := r.word(line, props, "id")
	typ, _ := r.word(line, props, "type")
	links := r.links(line, props, id, idOK)
	if len(r.problems) > 0 {
		sort.SliceStable(r.problems, func(i, j int) bool { return r.problems[i].Line < r.problems[j].Line })
		return Breadcrumb{}, r.problems
	}
	return Breadcrumb{ID: id, Type: typ, Links: links}, nil
}

// reader gathers the problems of one breadcrumb.
type reader struct {
	problems []Problem
}

func (r *reader) report(line int, format string, args ...any) {
	r.problems = append(r.problems, Problem{Line: line, Message: fmt.Sprintf(format, args...)})
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
