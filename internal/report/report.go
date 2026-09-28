// Package report draws the breadcrumbs of a repository as an HTML page:
// one rectangle per breadcrumb, an arrow from each breadcrumb to each of
// its parents, and a color per verdict.
package report

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/breadcrumb"
)

const (
	nodeWidth  = 260
	headerH    = 26 // the compartment with the id
	lineH      = 17
	padY       = 8
	gapX       = 30
	gapY       = 70
	margin     = 20
	wrapAt     = 36 // characters per title line
	titleLines = 3
	maxClaims  = 6 // claims listed in a Task's rectangle
)

// node is a rectangle in the diagram.
type node struct {
	key     string // the path it stands for
	parents []string
	X, Y    int
	W, H    int
	Header  string
	Lines   []string
	Class   string // confirmed, refuted, undecided or missing
	sortKey string

	// Positions for the template, set once the node is placed.
	TextX, HeaderY, DividerY, DividerX2 int
	Texts                               []text
}

// text is one line of a node, at its baseline.
type text struct {
	Text string
	Y    int
}

// edge is an arrow from a child's top to a parent's bottom.
type edge struct{ X1, Y1, X2, Y2 int }

// problem is a breadcrumb that could not be read as the breadcrumbs define.
type problem struct{ Path, Text string }

type page struct {
	Width, Height int
	Nodes         []*node
	Edges         []edge
	Problems      []problem
	Others        []string
	Count         int
}

//go:embed report.html
var pageHTML string

var pageTemplate = template.Must(template.New("report").Parse(pageHTML))

// Write writes the HTML report of g to w.
func Write(w io.Writer, g breadcrumb.Graph) error {
	return pageTemplate.Execute(w, build(g))
}

func build(g breadcrumb.Graph) page {
	nodes := map[string]*node{}
	var order []string
	add := func(n *node) {
		nodes[n.key] = n
		order = append(order, n.key)
	}

	var problems []problem
	for _, b := range g.Breadcrumbs {
		n := &node{
			key:     b.Path,
			Header:  b.ID,
			Lines:   append(append(wrap(b.Title), "type: "+b.Type.String(), "verdict: "+b.Verdict.String()), claimLines(b.Claims)...),
			Class:   strings.ToLower(b.Verdict.String()),
			sortKey: "0" + b.ID,
		}
		for _, p := range b.Parents {
			n.parents = append(n.parents, p.Target)
		}
		add(n)
		for _, text := range b.Problems {
			problems = append(problems, problem{Path: b.Path, Text: text})
		}
	}
	// A link to a file that is not a breadcrumb draws a missing node.
	for _, b := range g.Breadcrumbs {
		for _, p := range b.Parents {
			if _, ok := nodes[p.Target]; !ok {
				add(&node{
					key:     p.Target,
					Header:  p.Text,
					Lines:   []string{"missing", p.Target},
					Class:   "missing",
					sortKey: "1" + p.Target,
				})
			}
		}
	}
	for _, n := range nodes {
		n.H = headerH + padY*2 + lineH*len(n.Lines)
	}

	rows := layers(nodes, order)
	pg := page{Problems: problems, Others: g.Others, Count: len(g.Breadcrumbs)}
	x := map[string]int{} // center of each placed node
	y := margin
	widest := 0
	for _, row := range rows {
		if w := len(row)*(nodeWidth+gapX) - gapX; w > widest {
			widest = w
		}
	}
	for r, row := range rows {
		if r > 0 {
			// Children sit under the middle of their parents.
			for _, n := range row {
				sum := 0
				for _, p := range n.parents {
					sum += x[p]
				}
				n.sortKey = pad(sum/len(n.parents)) + n.sortKey
			}
		}
		sort.SliceStable(row, func(i, j int) bool { return row[i].sortKey < row[j].sortKey })

		rowW := len(row)*(nodeWidth+gapX) - gapX
		left := margin + (widest-rowW)/2
		rowH := 0
		for i, n := range row {
			n.X, n.Y, n.W = left+i*(nodeWidth+gapX), y, nodeWidth
			x[n.key] = n.X + n.W/2
			n.TextX, n.HeaderY, n.DividerY, n.DividerX2 = n.X+10, n.Y+headerH-8, n.Y+headerH, n.X+n.W
			for i, l := range n.Lines {
				n.Texts = append(n.Texts, text{Text: l, Y: n.Y + headerH + padY + lineH*(i+1) - 4})
			}
			if n.H > rowH {
				rowH = n.H
			}
			pg.Nodes = append(pg.Nodes, n)
		}
		y += rowH + gapY
	}
	for _, n := range pg.Nodes {
		for _, p := range n.parents {
			parent := nodes[p]
			pg.Edges = append(pg.Edges, edge{
				X1: n.X + n.W/2, Y1: n.Y,
				X2: parent.X + parent.W/2, Y2: parent.Y + parent.H,
			})
		}
	}
	pg.Width = widest + 2*margin
	pg.Height = y - gapY + margin
	return pg
}

// layers puts each node one row below its lowest parent. A node with no
// parent sits in the first row.
func layers(nodes map[string]*node, order []string) [][]*node {
	depth := map[string]int{}
	visiting := map[string]bool{}
	var depthOf func(k string) int
	depthOf = func(k string) int {
		if d, ok := depth[k]; ok {
			return d
		}
		if visiting[k] { // a cycle: draw it rather than loop
			return 0
		}
		visiting[k] = true
		d := 0
		for _, p := range nodes[k].parents {
			if pd := depthOf(p) + 1; pd > d {
				d = pd
			}
		}
		visiting[k] = false
		depth[k] = d
		return d
	}
	var rows [][]*node
	for _, k := range order {
		d := depthOf(k)
		for len(rows) <= d {
			rows = append(rows, nil)
		}
		rows[d] = append(rows[d], nodes[k])
	}
	return rows
}

// claimLines lists a Task's claims, each marked with whether it holds.
func claimLines(claims []breadcrumb.Claim) []string {
	var lines []string
	for i, c := range claims {
		if i == maxClaims {
			lines = append(lines, fmt.Sprintf("… %d more claims", len(claims)-maxClaims))
			break
		}
		mark := "✗"
		if c.Holds {
			mark = "✓"
		}
		lines = append(lines, truncate(mark+" "+c.Path+": "+c.Text))
	}
	return lines
}

// truncate cuts s to wrapAt characters.
func truncate(s string) string {
	r := []rune(s)
	if len(r) <= wrapAt {
		return s
	}
	return string(r[:wrapAt-1]) + "…"
}

// wrap breaks a title into lines of at most wrapAt characters, and at
// most titleLines lines.
func wrap(s string) []string {
	var lines []string
	line := ""
	for _, w := range strings.Fields(s) {
		if line != "" && len(line)+1+len(w) > wrapAt {
			lines = append(lines, line)
			line = w
			continue
		}
		if line != "" {
			line += " "
		}
		line += w
	}
	if line != "" {
		lines = append(lines, line)
	}
	if len(lines) > titleLines {
		lines = lines[:titleLines]
		lines[titleLines-1] += "…"
	}
	return lines
}

// pad writes n as a fixed-width string, so it sorts as a number.
func pad(n int) string { return fmt.Sprintf("%010d", n) }
