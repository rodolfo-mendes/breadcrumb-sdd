// Package manpage turns the Markdown contract of bcr into a man page in
// roff (ADR-0009). It reads the part of Markdown ADR-0008 allows: a first
// heading `# name(section)`, `##` and `###` headings, paragraphs, `- `
// lists, code spans, fenced code blocks and links.
package manpage

import (
	"fmt"
	"regexp"
	"strings"
)

// title is the first heading: the page's name and section, such as bcr(1).
var title = regexp.MustCompile(`^# ([a-z][a-z0-9-]*)\(([1-9])\)$`)

// inline is a code span or a link.
var inline = regexp.MustCompile("`([^`]+)`" + `|\[([^\]]+)\]\(([^)\s]+)\)`)

// Convert returns the man page of the Markdown document md.
func Convert(md string) (string, error) {
	lines := strings.Split(strings.TrimRight(md, "\n"), "\n")
	m := title.FindStringSubmatch(lines[0])
	if m == nil {
		return "", fmt.Errorf("line 1: the first line is not a # name(section) heading")
	}
	var b strings.Builder
	fmt.Fprintf(&b, ".TH %q %q \"\" \"\" \"Breadcrumb SDD\"\n", strings.ToUpper(m[1]), m[2])

	var para []string // the lines of the paragraph or list item being read
	var item bool     // whether para is a list item
	flush := func() {
		if len(para) == 0 {
			return
		}
		if item {
			b.WriteString(".IP \\(bu 2\n")
		} else {
			b.WriteString(".PP\n")
		}
		b.WriteString(text(strings.Join(para, " ")) + "\n")
		para, item = nil, false
	}

	for i := 1; i < len(lines); i++ {
		l := lines[i]
		switch {
		case l == "":
			flush()
		case strings.HasPrefix(l, "```"):
			flush()
			b.WriteString(".PP\n.RS 4\n.nf\n")
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				b.WriteString(lead(escape(lines[i])) + "\n")
			}
			if i == len(lines) {
				return "", fmt.Errorf("line %d: a code block is not closed", len(lines))
			}
			b.WriteString(".fi\n.RE\n")
		case strings.HasPrefix(l, "## "):
			flush()
			fmt.Fprintf(&b, ".SH %s\n", quote(strings.ToUpper(l[3:])))
		case strings.HasPrefix(l, "### "):
			flush()
			fmt.Fprintf(&b, ".SS %s\n", quote(l[4:]))
		case strings.HasPrefix(l, "#"):
			return "", fmt.Errorf("line %d: only ## and ### headings may follow the first", i+1)
		case strings.HasPrefix(l, "- "):
			flush()
			para, item = []string{l[2:]}, true
		case item && strings.HasPrefix(l, "  "):
			para = append(para, strings.TrimSpace(l))
		case strings.HasPrefix(l, "|") || strings.HasPrefix(l, "<"):
			return "", fmt.Errorf("line %d: tables and HTML cannot be converted", i+1)
		default:
			if item {
				flush()
			}
			para = append(para, strings.TrimSpace(l))
		}
	}
	flush()
	return b.String(), nil
}

// text turns a line of Markdown into roff: code spans in bold, links as
// their text, with the target after it when it is a web address.
func text(s string) string {
	var b strings.Builder
	last := 0
	for _, m := range inline.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(escape(s[last:m[0]]))
		if m[2] >= 0 {
			b.WriteString(`\fB` + escape(s[m[2]:m[3]]) + `\fR`)
		} else {
			label, target := s[m[4]:m[5]], s[m[6]:m[7]]
			b.WriteString(escape(label))
			if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") {
				b.WriteString(" <" + escape(target) + ">")
			}
		}
		last = m[1]
	}
	b.WriteString(escape(s[last:]))
	return lead(b.String())
}

// escape keeps roff from reading s as anything but text.
func escape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\e`)
	return strings.ReplaceAll(s, "-", `\-`)
}

// lead keeps a line that starts with . or ' from being read as a request.
func lead(s string) string {
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "'") {
		return `\&` + s
	}
	return s
}

// quote is s as one argument of a roff request.
func quote(s string) string {
	return `"` + strings.ReplaceAll(escape(s), `"`, `""`) + `"`
}
