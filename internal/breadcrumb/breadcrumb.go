// Package breadcrumb is the model of Breadcrumb SDD: which files are
// breadcrumbs, what each one holds, how they link to each other, and
// the verdict of each one against the repository's files.
package breadcrumb

import (
	"bytes"
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
	Task              Type = "TK" // TD-0011
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
	case Task:
		return "Task"
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
	Line   int    // the line it is on, counted from 1
}

// Problem is a line of a breadcrumb that cannot be read as the
// breadcrumbs define (RQ-0020, TD-0024).
type Problem struct {
	Line    int // counted from 1
	Message string
}

// Claim is a contains claim of a Task (TD-0012): the file at Path holds Text.
type Claim struct {
	Path  string // relative to the repository root
	Text  string
	Holds bool // whether the claim holds against the repository's files
}

// Breadcrumb is one breadcrumb file.
type Breadcrumb struct {
	ID       string // the file name without .md, such as TD-0001
	Type     Type
	Path     string // relative to the repository root
	Title    string // the # title, without a leading "<ID>: "
	Parents  []Link
	Claims   []Claim   // a Task's claims that could be read
	Problems []Problem // what could not be read as the breadcrumbs define, in order of line
	Verdict  Verdict

	unreadable bool // a Task with a claim that could not be read
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
// four-digit number (TD-0001, TD-0003, TD-0004, TD-0011).
var fileName = regexp.MustCompile(`^(IN|RQ|TD|TK)-[0-9]{4}\.md$`)

// parentLine is a Parent: line with one Markdown link (TD-0005).
var parentLine = regexp.MustCompile(`^Parent: \[([^\]]*)\]\(([^)\s]+)\)$`)

// claimLine is a contains claim (TD-0012).
var claimLine = regexp.MustCompile("^- `([^`]+)` contains `([^`]+)`$")

// Load reads the breadcrumbs of the repository whose root is fsys and
// audits them against its files. A repository with no Dir has no
// breadcrumbs.
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
		// TD-0013: breadcrumbs sit directly in Dir.
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
	audit(fsys, g.Breadcrumbs)
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

	// TD-0013: the first line is the title.
	if !strings.HasPrefix(lines[0], "# ") {
		b.Problems = append(b.Problems, Problem{1, "the first line is not a # title"})
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
			b.Problems = append(b.Problems, Problem{i + 1, "not a Parent: link: " + lines[i]})
			continue
		}
		b.Parents = append(b.Parents, Link{Text: m[1], Target: path.Join(path.Dir(p), m[2]), Line: i + 1})
	}

	if b.Type == Task {
		parseClaims(&b, lines, i)
	}
	return b
}

// parseClaims reads the list items under a Task's ## Claims heading,
// up to the next heading (TD-0011), from lines[start:].
func parseClaims(b *Breadcrumb, lines []string, start int) {
	in := false
	for n := start; n < len(lines); n++ {
		l := lines[n]
		if strings.HasPrefix(l, "#") {
			in = l == "## Claims"
			continue
		}
		if !in || !isListItem(l) {
			continue
		}
		m := claimLine.FindStringSubmatch(l)
		if m == nil || !fs.ValidPath(m[1]) || m[1] == "." {
			b.Problems = append(b.Problems, Problem{n + 1, "not a claim: " + l})
			b.unreadable = true
			continue
		}
		b.Claims = append(b.Claims, Claim{Path: m[1], Text: m[2]})
	}
}

// isListItem reports whether l is a Markdown bullet item, at any
// indentation and with any marker, so that a claim written slightly
// wrong is refused rather than skipped.
func isListItem(l string) bool {
	t := strings.TrimLeft(l, " \t")
	return strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ")
}

// audit checks each Task's claims against fsys and gives every
// breadcrumb its verdict.
func audit(fsys fs.FS, crumbs []Breadcrumb) {
	for i := range crumbs {
		for j := range crumbs[i].Claims {
			c := &crumbs[i].Claims[j]
			content, err := fs.ReadFile(fsys, c.Path)
			c.Holds = err == nil && bytes.Contains(content, []byte(c.Text))
		}
	}

	byPath := map[string]int{}
	for i, b := range crumbs {
		byPath[b.Path] = i
	}
	children := map[string][]int{}
	for i, b := range crumbs {
		for _, p := range b.Parents {
			if _, ok := byPath[p.Target]; ok {
				children[p.Target] = append(children[p.Target], i)
				continue
			}
			// RQ-0022: a link to no breadcrumb is a problem.
			crumbs[i].Problems = append(crumbs[i].Problems, Problem{p.Line, "Parent: links to no breadcrumb: " + p.Target})
		}
		sort.SliceStable(crumbs[i].Problems, func(a, c int) bool { return crumbs[i].Problems[a].Line < crumbs[i].Problems[c].Line })
	}

	done := map[int]bool{}
	visiting := map[int]bool{}
	var verdict func(i int) Verdict
	verdict = func(i int) Verdict {
		if done[i] {
			return crumbs[i].Verdict
		}
		if visiting[i] { // a cycle decides nothing
			return Undecided
		}
		visiting[i] = true
		var below []Verdict
		if crumbs[i].Type == Task {
			below = append(below, claimsVerdict(crumbs[i]))
		}
		for _, c := range children[crumbs[i].Path] {
			below = append(below, verdict(c))
		}
		visiting[i] = false
		crumbs[i].Verdict = combine(below)
		done[i] = true
		return crumbs[i].Verdict
	}
	for i := range crumbs {
		verdict(i)
	}
}

// claimsVerdict is a Task's verdict from its own claims (RQ-0007).
func claimsVerdict(b Breadcrumb) Verdict {
	if b.unreadable {
		return Refuted
	}
	if len(b.Claims) == 0 {
		return Undecided
	}
	for _, c := range b.Claims {
		if !c.Holds {
			return Refuted
		}
	}
	return Confirmed
}

// combine is the verdict of a breadcrumb from the verdicts below it:
// Refuted if any is (RQ-0006), Confirmed if there are some and all are
// (RQ-0008), Undecided otherwise.
func combine(vs []Verdict) Verdict {
	if len(vs) == 0 {
		return Undecided
	}
	all := true
	for _, v := range vs {
		if v == Refuted {
			return Refuted
		}
		if v != Confirmed {
			all = false
		}
	}
	if all {
		return Confirmed
	}
	return Undecided
}
