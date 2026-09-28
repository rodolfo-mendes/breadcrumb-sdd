// Package specsite builds the web site of the specification of
// Breadcrumb SDD (IN-0016): a page for each of its versions, one for
// the latest, and a list of them all at the root.
//
// It turns the specification into HTML itself (TD-0032), reading the
// part of Markdown TD-0022 allows: a first heading `# title`, `##` and
// `###` headings, paragraphs, `- ` lists, code spans, fenced code
// blocks and links.
package specsite

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Release is the specification as it is at a tag.
type Release struct {
	Tag  string // such as v0.5.0
	Spec string // docs/breadcrumb-sdd.md at the tag
}

// SpecPath is where the specification is in the repository.
const SpecPath = "docs/breadcrumb-sdd.md"

// tagName is a release tag, vMAJOR.MINOR.PATCH (TD-0017).
var tagName = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)

// versionLine is the specification's third line (RQ-0036).
var versionLine = regexp.MustCompile(`^Version ([0-9]+)\.([0-9]+)\.([0-9]+)$`)

// inline is a code span or a link.
var inline = regexp.MustCompile("`([^`]+)`" + `|\[([^\]]+)\]\(([^)\s]+)\)`)

// page is the version of the specification a page shows, and the
// release it is published from.
type page struct {
	Version string
	Release Release
	version [3]int
	tag     [3]int
}

// Build returns the files of the site, by their path in it, for the
// releases of the repository at repo, such as
// https://github.com/rodolfo-mendes/breadcrumb-sdd. A release whose
// tag is not vMAJOR.MINOR.PATCH is left out.
func Build(repo string, releases []Release) (map[string][]byte, error) {
	// TD-0031: the latest tag of each version.
	byVersion := map[string]page{}
	for _, r := range releases {
		m := tagName.FindStringSubmatch(r.Tag)
		if m == nil {
			continue
		}
		lines := strings.Split(r.Spec, "\n")
		if len(lines) < 3 {
			return nil, fmt.Errorf("%s: %s has no version on line 3", r.Tag, SpecPath)
		}
		v := versionLine.FindStringSubmatch(strings.TrimSuffix(lines[2], "\r"))
		if v == nil {
			return nil, fmt.Errorf("%s: %s has no version on line 3", r.Tag, SpecPath)
		}
		p := page{Version: strings.TrimPrefix(v[0], "Version "), Release: r, version: numbers(v), tag: numbers(m)}
		if old, ok := byVersion[p.Version]; !ok || less(old.tag, p.tag) {
			byVersion[p.Version] = p
		}
	}
	var pages []page
	for _, p := range byVersion {
		pages = append(pages, p)
	}
	sort.Slice(pages, func(i, j int) bool { return less(pages[j].version, pages[i].version) })

	files := map[string][]byte{}
	for i, p := range pages {
		b, err := render(repo, p, i == 0)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", p.Release.Tag, err)
		}
		files["spec/"+p.Version+"/index.html"] = b
		if i == 0 {
			files["spec/latest/index.html"] = b // RQ-0038
		}
	}
	var index bytes.Buffer
	if err := indexPage.Execute(&index, struct {
		Repo  string
		Pages []page
	}{repo, pages}); err != nil {
		return nil, err
	}
	files["index.html"] = index.Bytes()
	return files, nil
}

// render is the page of p. Its links to the root are relative, so the
// same page serves spec/VERSION/ and spec/latest/.
func render(repo string, p page, latest bool) ([]byte, error) {
	title, body, err := Convert(p.Release.Spec, func(target string) string {
		return link(repo, p.Release.Tag, target)
	})
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = specPage.Execute(&b, struct {
		Title, Version, Tag, TagURL string
		Latest                      bool
		Body                        template.HTML
	}{title, p.Version, p.Release.Tag, repo + "/tree/" + p.Release.Tag, latest, template.HTML(body)})
	return b.Bytes(), err
}

// link is where a link in the specification leads on its page: a
// relative one to the file on GitHub at the tag (RQ-0039), anything
// else as it is.
func link(repo, tag, target string) string {
	if strings.HasPrefix(target, "#") || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
		return target
	}
	return repo + "/blob/" + tag + "/" + path.Join(path.Dir(SpecPath), target)
}

// Convert returns the title of the Markdown document md, its first
// heading, and the rest of it as HTML. Each link's target goes through
// link. Each heading gets an id made from its text (RQ-0040).
func Convert(md string, link func(string) string) (title, body string, err error) {
	lines := strings.Split(strings.TrimRight(md, "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	if !strings.HasPrefix(lines[0], "# ") {
		return "", "", fmt.Errorf("line 1: the first line is not a # heading")
	}
	title = lines[0][2:]

	var b strings.Builder
	ids := map[string]int{}
	var para []string // the lines of the paragraph or list item being read
	var item bool     // whether para is a list item
	var list bool     // whether a <ul> is open
	flush := func() {
		if len(para) > 0 {
			tag := "p"
			if item {
				tag = "li"
			}
			fmt.Fprintf(&b, "<%s>%s</%s>\n", tag, text(strings.Join(para, " "), link), tag)
		}
		para, item = nil, false
	}
	endList := func() {
		flush()
		if list {
			b.WriteString("</ul>\n")
			list = false
		}
	}
	heading := func(level int, s string) {
		endList()
		id := anchor(s)
		if ids[id]++; ids[id] > 1 {
			id += "-" + strconv.Itoa(ids[id])
		}
		fmt.Fprintf(&b, "<h%d id=\"%s\">%s</h%d>\n", level, id, text(s, link), level)
	}

	for i := 1; i < len(lines); i++ {
		l := lines[i]
		switch {
		case l == "":
			if item {
				flush()
			} else {
				endList()
			}
		case strings.HasPrefix(l, "```"):
			endList()
			b.WriteString("<pre><code>")
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				b.WriteString(html.EscapeString(lines[i]) + "\n")
			}
			if i == len(lines) {
				return "", "", fmt.Errorf("line %d: a code block is not closed", len(lines))
			}
			b.WriteString("</code></pre>\n")
		case strings.HasPrefix(l, "## "):
			heading(2, l[3:])
		case strings.HasPrefix(l, "### "):
			heading(3, l[4:])
		case strings.HasPrefix(l, "#"):
			return "", "", fmt.Errorf("line %d: only ## and ### headings may follow the first", i+1)
		case strings.HasPrefix(l, "- "):
			flush()
			if !list {
				b.WriteString("<ul>\n")
				list = true
			}
			para, item = []string{l[2:]}, true
		case item && strings.HasPrefix(l, "  "):
			para = append(para, strings.TrimSpace(l))
		case strings.HasPrefix(l, "|") || strings.HasPrefix(l, "<"):
			return "", "", fmt.Errorf("line %d: tables and HTML cannot be converted", i+1)
		default:
			if item || list {
				endList()
			}
			para = append(para, strings.TrimSpace(l))
		}
	}
	endList()
	return title, b.String(), nil
}

// text turns a line of Markdown into HTML: code spans and links.
func text(s string, link func(string) string) string {
	var b strings.Builder
	last := 0
	for _, m := range inline.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(html.EscapeString(s[last:m[0]]))
		if m[2] >= 0 {
			b.WriteString("<code>" + html.EscapeString(s[m[2]:m[3]]) + "</code>")
		} else {
			fmt.Fprintf(&b, `<a href="%s">%s</a>`, html.EscapeString(link(s[m[6]:m[7]])), html.EscapeString(s[m[4]:m[5]]))
		}
		last = m[1]
	}
	b.WriteString(html.EscapeString(s[last:]))
	return b.String()
}

// anchor is the id of a heading: its letters and digits in lower case,
// with a hyphen for each run of anything else, such as the-title.
func anchor(s string) string {
	var b strings.Builder
	gap := false
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if gap && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			gap = false
		} else {
			gap = true
		}
	}
	if b.Len() == 0 {
		return "section"
	}
	return b.String()
}

// numbers is the three numbers matched by m, after the whole match.
func numbers(m []string) [3]int {
	var n [3]int
	for i := range n {
		n[i], _ = strconv.Atoi(m[i+1])
	}
	return n
}

// less reports whether version a comes before version b.
func less(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

//go:embed site.html
var siteHTML string

var (
	site      = template.Must(template.New("site").Parse(siteHTML))
	specPage  = site.Lookup("page")
	indexPage = site.Lookup("index")
)
