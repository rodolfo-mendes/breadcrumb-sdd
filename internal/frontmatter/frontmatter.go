// Package frontmatter reads a breadcrumb from the YAML front matter of
// a file (ADR-0002, ADR-0010). It checks the rules of the format and
// hands the core, internal/crumb, the breadcrumb's properties as
// written, with the lines of the file (ADR-0013).
package frontmatter

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
	"go.yaml.in/yaml/v3"
)

// Key is the key of the front matter that holds the breadcrumb.
const Key = "breadcrumb"

// Breadcrumb is a breadcrumb as the front matter writes it, before the
// core checks it.
type Breadcrumb struct {
	Line       int // the line of the breadcrumb key, counted from 1
	Properties map[string]crumb.Property
}

// Read reads the breadcrumb in the front matter of src, the text of a
// file. found is false when src has none. When the front matter breaks
// a rule of the format, Read returns every problem, in order of line,
// and no breadcrumb.
func Read(src []byte) (b Breadcrumb, found bool, problems []crumb.Problem) {
	text, lines, ok := split(src)
	if !ok {
		return Breadcrumb{}, false, nil
	}
	if lines == nil {
		return Breadcrumb{}, true, []crumb.Problem{{Line: 1, Message: "front matter does not close: no line after the first is ---"}}
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(text, &doc); err != nil {
		return Breadcrumb{}, true, []crumb.Problem{syntaxProblem(lines, err)}
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return Breadcrumb{}, false, nil // empty, or not a map
	}

	r := reader{lines: lines}
	var key, value *yaml.Node
	root := doc.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i]
		if k.Kind != yaml.ScalarNode || k.Value != Key {
			continue // keys outside breadcrumb are not read (ADR-0002)
		}
		if key != nil {
			r.report(k.Line, "%s is written twice; a file holds one breadcrumb", Key)
			continue
		}
		key, value = k, root.Content[i+1]
	}
	if key == nil {
		return Breadcrumb{}, false, nil
	}
	r.plain(key)
	props := r.properties(value)
	if len(r.problems) > 0 {
		sort.SliceStable(r.problems, func(i, j int) bool { return r.problems[i].Line < r.problems[j].Line })
		return Breadcrumb{}, true, r.problems
	}
	return Breadcrumb{Line: fileLine(key.Line), Properties: props}, true, nil
}

// split finds the front matter of src: its first line is ---, and it
// ends at the next line that is ---. A line ends at \n or \r\n, and a
// byte order mark before the first line is skipped. ok is false when
// src has no front matter. When it does not close, lines is nil.
// Otherwise text is the front matter between the two lines, with \n
// line ends, and lines holds each of its lines.
func split(src []byte) (text []byte, lines []string, ok bool) {
	src = bytes.TrimPrefix(src, []byte("\xEF\xBB\xBF"))
	all := strings.Split(string(src), "\n")
	for i := range all {
		all[i] = strings.TrimSuffix(all[i], "\r")
	}
	if all[0] != "---" {
		return nil, nil, false
	}
	for i := 1; i < len(all); i++ {
		if all[i] == "---" {
			lines = all[1:i]
			return []byte(strings.Join(lines, "\n")), lines, true
		}
	}
	return nil, nil, true
}

// fileLine turns a line of the front matter, counted from 1, into a
// line of the file: the front matter starts after the --- line.
func fileLine(line int) int { return line + 1 }

// syntaxLine is the start of yaml.v3's message for a syntax error.
var syntaxLine = regexp.MustCompile(`^yaml: (line \d+: )?`)

// syntaxProblem is the problem of front matter, made of lines, that is
// not YAML, with err the parser's error. yaml.v3 gives the line of an
// error only in its message, sometimes not at all, and counts it from 0
// or from 1 depending on the error. So the problem is put at the first
// line the front matter cannot be read up to: the parser's message is
// kept, and its line dropped.
func syntaxProblem(lines []string, err error) crumb.Problem {
	line := len(lines)
	for ; line > 0; line-- {
		var n yaml.Node
		if yaml.Unmarshal([]byte(strings.Join(lines[:line-1], "\n")), &n) == nil {
			break
		}
	}
	msg := syntaxLine.ReplaceAllString(err.Error(), "")
	return crumb.Problem{Line: fileLine(line), Message: "front matter is not YAML: " + msg}
}

// reader gathers the problems of one breadcrumb's front matter.
type reader struct {
	lines    []string // the lines of the front matter
	problems []crumb.Problem
}

// report adds a problem at line, a line of the front matter.
func (r *reader) report(line int, format string, args ...any) {
	r.problems = append(r.problems, crumb.Problem{Line: fileLine(line), Message: fmt.Sprintf(format, args...)})
}

// properties reads the value of the breadcrumb key. With no value, the
// breadcrumb has no properties, and the core says which are missing.
func (r *reader) properties(n *yaml.Node) map[string]crumb.Property {
	props := map[string]crumb.Property{}
	switch {
	case !r.plain(n):
		return props
	case n.Kind == yaml.ScalarNode && n.Value == "":
		return props
	case n.Kind != yaml.MappingNode:
		r.report(n.Line, "%s must be a map of id, type and links", Key)
		return props
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		if !r.plain(k) {
			continue
		}
		if k.Kind != yaml.ScalarNode {
			r.report(k.Line, "a key under %s must be a word", Key)
			continue
		}
		if _, ok := props[k.Value]; ok {
			r.report(k.Line, "%s is written twice", k.Value)
			continue
		}
		props[k.Value] = r.property(k, v)
	}
	return props
}

// property reads the value v of the key k.
func (r *reader) property(k, v *yaml.Node) crumb.Property {
	p := crumb.Property{Line: fileLine(k.Line)}
	if !r.plain(v) {
		return p
	}
	switch v.Kind {
	case yaml.ScalarNode:
		if v.Value != "" {
			p.Kind, p.Text = crumb.TextValue, v.Value
		}
	case yaml.SequenceNode:
		p.Kind = crumb.ListValue
		for _, e := range v.Content {
			if !r.plain(e) {
				continue
			}
			if e.Kind != yaml.ScalarNode {
				r.report(e.Line, "a list inside a list is not allowed in a breadcrumb")
				continue
			}
			p.Entries = append(p.Entries, crumb.Entry{Text: e.Value, Line: fileLine(e.Line)})
		}
	case yaml.MappingNode:
		r.report(k.Line, "a map as the value of %s is not allowed in a breadcrumb", k.Value)
	}
	return p
}

// plain reports whether n is written in ADR-0002's part of YAML, and
// reports a problem when it is not: no anchor, alias or tag, no quoted
// value or value over several lines, and no flow collection but [].
func (r *reader) plain(n *yaml.Node) bool {
	switch {
	case n.Kind == yaml.AliasNode:
		r.report(n.Line, "alias *%s is not allowed in a breadcrumb", n.Value)
	case n.Anchor != "":
		r.report(n.Line, "anchor &%s is not allowed in a breadcrumb", n.Anchor)
	case n.Style&yaml.TaggedStyle != 0:
		r.report(n.Line, "tag %s is not allowed in a breadcrumb", n.Tag)
	case n.Style&(yaml.SingleQuotedStyle|yaml.DoubleQuotedStyle) != 0:
		r.report(n.Line, "quoted value is not allowed in a breadcrumb; write it plain")
	case n.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0,
		n.Kind == yaml.ScalarNode && !r.oneLine(n):
		r.report(n.Line, "value over several lines is not allowed in a breadcrumb")
	case n.Style&yaml.FlowStyle != 0 && (n.Kind != yaml.SequenceNode || len(n.Content) > 0):
		r.report(n.Line, "flow collection is not allowed in a breadcrumb; only [] is")
	default:
		return true
	}
	return false
}

// oneLine reports whether the plain scalar n is written on one line:
// its value is the text that starts at its column. A plain value over
// several lines is folded into one by the parser, so its value is
// longer than what its first line holds.
func (r *reader) oneLine(n *yaml.Node) bool {
	if n.Line < 1 || n.Line > len(r.lines) {
		return true
	}
	line := []rune(r.lines[n.Line-1])
	if n.Column < 1 || n.Column-1 > len(line) {
		return true
	}
	return strings.HasPrefix(string(line[n.Column-1:]), n.Value)
}
