// Package breadcrumb is the model of Breadcrumb SDD: which files are
// breadcrumbs, what each one holds, and how they link to each other.
package breadcrumb

import (
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

// Dir is the directory, at the repository root, that holds the breadcrumbs.
const Dir = "breadcrumbs"

// Type is the kind of a breadcrumb, given by the prefix of its file name.
type Type string

const (
	Intake            Type = "IN" // TD-0003
	Requirement       Type = "RQ" // TD-0004
	TechnicalDecision Type = "TD" // TD-0001
)

// String returns the name of the type, as the breadcrumbs spell it.
func (t Type) String() string {
	switch t {
	case Intake:
		return "Intake"
	case Requirement:
		return "Requirement"
	case TechnicalDecision:
		return "Technical Decision"
	}
	return string(t)
}

// Verdict is the result of auditing a breadcrumb.
type Verdict int

const (
	Undecided Verdict = iota
	Confirmed
	Refuted
)

// String returns the name of the verdict.
func (v Verdict) String() string {
	switch v {
	case Confirmed:
		return "Confirmed"
	case Refuted:
		return "Refuted"
	}
	return "Undecided"
}

// Link is a Parent: line of a breadcrumb (TD-0005).
type Link struct {
	Text   string // the link's text, as written
	Target string // the path it points to, relative to the repository root
}

// Breadcrumb is one breadcrumb file.
type Breadcrumb struct {
	ID       string // the file name without .md, such as TD-0001
	Type     Type
	Path     string // relative to the repository root
	Title    string // the # title, without a leading "<ID>: "
	Parents  []Link
	Problems []string // what could not be read as the breadcrumbs define
	Verdict  Verdict
}

// Graph is every breadcrumb of a repository.
type Graph struct {
	Breadcrumbs []Breadcrumb // in order of path
	Others      []string     // files under Dir that are not breadcrumbs, in order of path
}

// Lookup returns the breadcrumb at the path p, relative to the repository root.
func (g Graph) Lookup(p string) (Breadcrumb, bool) {
	for _, b := range g.Breadcrumbs {
		if b.Path == p {
			return b, true
		}
	}
	return Breadcrumb{}, false
}

// fileName is the name of a breadcrumb file: a type prefix and a
// four-digit number (TD-0001, TD-0003, TD-0004).
var fileName = regexp.MustCompile(`^(IN|RQ|TD)-[0-9]{4}\.md$`)

// parentLine is a Parent: line with one Markdown link (TD-0005).
var parentLine = regexp.MustCompile(`^Parent: \[([^\]]*)\]\(([^)\s]+)\)$`)

// Load reads the breadcrumbs of the repository whose root is fsys.
// A repository with no Dir has no breadcrumbs.
func Load(fsys fs.FS) (Graph, error) {
	var g Graph
	err := fs.WalkDir(fsys, Dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == Dir && d == nil {
				return fs.SkipDir
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if path.Dir(p) != Dir || !fileName.MatchString(d.Name()) {
			g.Others = append(g.Others, p)
			return nil
		}
		content, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		g.Breadcrumbs = append(g.Breadcrumbs, parse(p, string(content)))
		return nil
	})
	if err != nil {
		return Graph{}, err
	}
	sort.Slice(g.Breadcrumbs, func(i, j int) bool { return g.Breadcrumbs[i].Path < g.Breadcrumbs[j].Path })
	sort.Strings(g.Others)
	return g, nil
}

// parse reads the breadcrumb at path p from its content.
func parse(p, content string) Breadcrumb {
	name := path.Base(p)
	id := strings.TrimSuffix(name, ".md")
	b := Breadcrumb{ID: id, Type: Type(name[:2]), Path: p, Verdict: Undecided}

	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}

	if !strings.HasPrefix(lines[0], "# ") {
		b.Problems = append(b.Problems, "the first line is not a # title")
		return b
	}
	b.Title = strings.TrimPrefix(strings.TrimPrefix(lines[0], "# "), id+": ")

	// Parent: lines count only right below the title, so an example
	// further down, as in TD-0005, is not a link.
	i := 1
	for i < len(lines) && lines[i] == "" {
		i++
	}
	for ; i < len(lines) && strings.HasPrefix(lines[i], "Parent:"); i++ {
		m := parentLine.FindStringSubmatch(lines[i])
		if m == nil {
			b.Problems = append(b.Problems, "not a Parent: link: "+lines[i])
			continue
		}
		b.Parents = append(b.Parents, Link{Text: m[1], Target: path.Join(path.Dir(p), m[2])})
	}
	return b
}
