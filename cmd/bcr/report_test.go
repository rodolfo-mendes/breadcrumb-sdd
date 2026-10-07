package main

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"testing"
)

// The tests below follow the Scenarios of specs/report/spec.md, one
// test each, in the same order.

// reportRun runs bcr report with input on standard input, and returns
// what it writes to standard output and standard error, and its exit
// status.
func reportRun(input string, operands ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"report"}, operands...), strings.NewReader(input), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// reported returns the page of records, and fails unless bcr report
// prints nothing to standard error and exits 0.
func reported(t *testing.T, records ...string) string {
	t.Helper()
	input := ""
	if len(records) > 0 {
		input = recordsOf(records...)
	}
	page, stderr, code := reportRun(input)
	if stderr != "" || code != 0 {
		t.Fatalf("got %q on standard error, exit %d; want nothing, exit 0", stderr, code)
	}
	return page
}

// The records of the kinds bcr report reads.
func crumbRec(id, typ, path string) string {
	return "breadcrumb\t" + id + "\t" + typ + "\t" + path + "\t3"
}

func linkRec(verb, object, path, line string) string {
	return "link\tID\t" + verb + "\t" + object + "\t" + path + "\t" + line
}

func verdictRec(verdict, path string) string {
	return "verdict\tID\t" + verdict + "\t" + path + "\t3"
}

func claimVerdictRec(verdict, path, line string) string {
	return "claim-verdict\tID\t" + verdict + "\t" + path + "\t" + line
}

// A spec that follows an ADR, with one claim, after bcr audit.
var auditedRecords = []string{
	"breadcrumb\tADR-0015\tADR\tdocs/adrs/a.md\t3",
	"breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3",
	"link\tverify\tfollows\tADR-0015\tspecs/verify/spec.md\t6",
	"claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8",
	"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8",
	"verdict\tADR-0015\tUndecided\tdocs/adrs/a.md\t3\tno claims",
	"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3",
}

// section returns the lines of page under the heading "## name", up to
// the next one, without the empty ones.
func section(t *testing.T, page, name string) []string {
	t.Helper()
	_, rest, ok := strings.Cut("\n"+page, "\n## "+name+"\n")
	if !ok {
		t.Fatalf("the page has no heading %q:\n%s", "## "+name, page)
	}
	rest, _, _ = strings.Cut(rest, "\n## ")
	var lines []string
	for _, l := range strings.Split(rest, "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// rows returns the rows of the table in lines, without its two lines
// of heading, each as its cells.
func rows(t *testing.T, lines []string) [][]string {
	t.Helper()
	var table [][]string
	for _, l := range lines {
		if !strings.HasPrefix(l, "| ") {
			continue
		}
		table = append(table, strings.Split(strings.TrimSuffix(strings.TrimPrefix(l, "| "), " |"), " | "))
	}
	if len(table) == 0 {
		t.Fatalf("no table in %q", lines)
	}
	return table[1:]
}

// column returns cell n of each row.
func column(table [][]string, n int) []string {
	var cells []string
	for _, r := range table {
		cells = append(cells, r[n])
	}
	return cells
}

// diagram returns the lines of the view whose heading is "### id",
// without the lines that are the same in every view.
func diagram(t *testing.T, page, id string) (boxes, arrows, classes []string) {
	t.Helper()
	_, rest, ok := strings.Cut(page, "\n### "+id+"\n\n```mermaid\nflowchart LR\n")
	if !ok {
		t.Fatalf("the page has no view %q:\n%s", id, page)
	}
	rest, _, _ = strings.Cut(rest, "```\n")
	for _, l := range strings.Split(strings.TrimSuffix(rest, "\n"), "\n") {
		switch {
		case strings.HasPrefix(l, "  classDef "):
		case strings.HasPrefix(l, "  class "):
			classes = append(classes, strings.TrimPrefix(l, "  class "))
		case strings.Contains(l, " --> "):
			arrows = append(arrows, strings.TrimSpace(l))
		default:
			boxes = append(boxes, strings.TrimSpace(l))
		}
	}
	return boxes, arrows, classes
}

// same fails unless got and want hold the same texts, in order.
func same(t *testing.T, what string, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("%s:\n%s\nwant\n%s", what, strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestReportASpecAndWhatItFollows(t *testing.T) {
	page := reported(t, auditedRecords...)
	at := 0
	for _, h := range []string{"# Breadcrumb report\n", "\n## Breadcrumbs\n", "\n## Problems\n", "\n## Claims\n", "\n## Views\n", "\n### verify\n"} {
		i := strings.Index(page[at:], h)
		if i < 0 {
			t.Fatalf("no heading %q after byte %d of\n%s", h, at, page)
		}
		at += i + len(h)
	}
	boxes, arrows, classes := diagram(t, page, "verify")
	same(t, "boxes", boxes,
		`n1["ADR-0015<br/>ADR · Undecided, no claims"]`,
		`n2["verify<br/>spec · Confirmed<br/>Confirmed: docs/bcr.md has-line #35;#35;#35; verify"]`)
	same(t, "arrows", arrows, `n2 -- "follows" --> n1`)
	same(t, "classes", classes, "n1 undecided", "n2 confirmed")
}

func TestReportNoRecordsArePassedOn(t *testing.T) {
	page := reported(t, auditedRecords...)
	for _, l := range strings.Split(page, "\n") {
		if strings.Contains(l, "\t") {
			t.Errorf("the page has the line %q, which has a tab", l)
		}
	}
	for _, r := range auditedRecords {
		if strings.Contains(page, r) {
			t.Errorf("the page holds the record %q", r)
		}
	}
}

func TestReportEveryBreadcrumbByVerdict(t *testing.T) {
	page := reported(t,
		crumbRec("b", "ADR", "b.md"), crumbRec("a", "ADR", "a.md"), crumbRec("c", "ADR", "c.md"),
		crumbRec("d", "ADR", "d.md"), crumbRec("e", "ADR", "e.md"),
		verdictRec("Undecided", "b.md"), verdictRec("Undecided", "a.md")+"\tno claims",
		verdictRec("Confirmed", "c.md"), verdictRec("Refuted", "d.md"))
	lines := section(t, page, "Breadcrumbs")
	if want := "Breadcrumbs: 5. Refuted: 1. Undecided: 2. Confirmed: 1. Not audited: 1."; lines[0] != want {
		t.Errorf("the counts are %q, want %q", lines[0], want)
	}
	table := rows(t, lines)
	same(t, "breadcrumbs", column(table, 1), "d", "a", "b", "c", "e")
	same(t, "verdicts", column(table, 0), "Refuted", "Undecided, no claims", "Undecided", "Confirmed", "not audited")
	same(t, "files", column(table, 3), "d.md:3", "a.md:3", "b.md:3", "c.md:3", "e.md:3")
}

func TestReportABreadcrumbInNoView(t *testing.T) {
	page := reported(t,
		crumbRec("verify", "spec", "v.md"), crumbRec("ADR-1", "ADR", "1.md"), crumbRec("ADR-4", "ADR", "4.md"),
		linkRec("follows", "ADR-1", "v.md", "6"))
	same(t, "breadcrumbs", column(rows(t, section(t, page, "Breadcrumbs")), 1), "ADR-1", "ADR-4", "verify")
	boxes, _, _ := diagram(t, page, "verify")
	if len(boxes) != 2 || strings.Contains(strings.Join(boxes, "\n"), "ADR-4") {
		t.Errorf("boxes %q, want ADR-1 and verify", boxes)
	}
}

func TestReportABreadcrumbTwoLinksAway(t *testing.T) {
	page := reported(t,
		crumbRec("verify", "spec", "v.md"), crumbRec("ADR-1", "ADR", "1.md"), crumbRec("ADR-2", "ADR", "2.md"),
		linkRec("follows", "ADR-1", "v.md", "6"), linkRec("amends", "ADR-1", "2.md", "6"))
	boxes, arrows, _ := diagram(t, page, "verify")
	same(t, "boxes", boxes, `n1["ADR-1<br/>ADR · not audited"]`, `n2["verify<br/>spec · not audited"]`)
	same(t, "arrows", arrows, `n2 -- "follows" --> n1`)
}

func TestReportLinksAmongTheBreadcrumbsOfAView(t *testing.T) {
	page := reported(t,
		crumbRec("verify", "spec", "v.md"), crumbRec("ADR-1", "ADR", "1.md"), crumbRec("ADR-2", "ADR", "2.md"),
		crumbRec("PBI-1", "PBI", "p.md"),
		linkRec("follows", "ADR-2", "v.md", "7"), linkRec("follows", "ADR-1", "v.md", "6"),
		linkRec("amends", "ADR-1", "2.md", "6"),
		linkRec("implements", "ADR-1", "p.md", "7"), linkRec("changes", "verify", "p.md", "6"))
	_, arrows, _ := diagram(t, page, "verify")
	same(t, "arrows", arrows,
		`n2 -- "amends" --> n1`,
		`n3 -- "implements" --> n1`,
		`n3 -- "changes" --> n4`,
		`n4 -- "follows" --> n1`,
		`n4 -- "follows" --> n2`)
}

func TestReportWhatABoxShows(t *testing.T) {
	page := reported(t,
		crumbRec("verify", "spec", "v.md"),
		claimAt("verify", "docs/bcr.md", "### verify", "v.md", "9"),
		claimAt("verify", "docs/a.md", "Status: Accepted", "v.md", "8"),
		claimVerdictRec("Refuted", "v.md", "8"), claimVerdictRec("Confirmed", "v.md", "9"),
		verdictRec("Refuted", "v.md"))
	boxes, _, classes := diagram(t, page, "verify")
	same(t, "boxes", boxes, `n1["verify<br/>spec · Refuted`+
		`<br/>Confirmed: docs/bcr.md has-line #35;#35;#35; verify`+
		`<br/>Refuted: docs/a.md has-line Status: Accepted"]`)
	same(t, "classes", classes, "n1 refuted")
}

func TestReportRecordsWithoutVerdicts(t *testing.T) {
	files(t, map[string]string{
		"a.md": fm("breadcrumb:", "  id: A", "  type: spec", "  links: []", "  claims:", "    - 'docs/bcr.md has-line ### verify'"),
	})
	extracted, stderr, code := extractAll("a.md")
	if stderr != "" || code != 0 {
		t.Fatalf("bcr extract wrote %q, exit %d", stderr, code)
	}
	page := reported(t, strings.TrimSuffix(extracted, "\n"))
	same(t, "breadcrumbs", column(rows(t, section(t, page, "Breadcrumbs")), 0), "not audited")
	same(t, "claims", column(rows(t, section(t, page, "Claims")), 0), "not audited")
	boxes, _, classes := diagram(t, page, "A")
	same(t, "boxes", boxes, `n1["A<br/>spec · not audited<br/>not audited: docs/bcr.md has-line #35;#35;#35; verify"]`)
	same(t, "classes", classes, "n1 unaudited")
}

func TestReportARefutedClaim(t *testing.T) {
	page, stderr, code := reportRun(recordsOf(verifyCrumb, verifyClaim,
		claimVerdictRec("Refuted", "specs/verify/spec.md", "8"), verdictRec("Refuted", "specs/verify/spec.md")))
	if stderr != "" || code != 0 {
		t.Errorf("got %q, exit %d; want nothing, exit 0: a Refuted claim fails bcr audit, not bcr report", stderr, code)
	}
	same(t, "breadcrumbs", rows(t, section(t, page, "Breadcrumbs"))[0], "Refuted", "verify", "spec", "specs/verify/spec.md:3")
	same(t, "claims", rows(t, section(t, page, "Claims"))[0], "Refuted", "verify", `docs/bcr.md has-line \#\#\# verify`, "specs/verify/spec.md:8")
}

func TestReportProblems(t *testing.T) {
	page := reported(t,
		crumbRec("A", "ADR", "a.md"),
		"problem\tb.md\t9\tbreadcrumb has no id",
		"problem\ta.md\t12\tlink points to \"ADR-9\", which is no breadcrumb's id",
		"problem\ta.md\t3\tid \"A\" is also at c.md:3")
	same(t, "problems", section(t, page, "Problems"),
		`- a.md:3: id "A" is also at c.md:3`,
		`- a.md:12: link points to "ADR-9", which is no breadcrumb's id`,
		`- b.md:9: breadcrumb has no id`)
}

func TestReportAClaimWithNoBreadcrumb(t *testing.T) {
	page := reported(t, crumbRec("A", "ADR", "a.md"),
		claimAt("gone", "docs/bcr.md", "x", "gone.md", "8"), claimVerdictRec("Confirmed", "gone.md", "8"))
	same(t, "claims", rows(t, section(t, page, "Claims"))[0], "Confirmed", "gone", "docs/bcr.md has-line x", "gone.md:8")
}

func TestReportTwoBreadcrumbsWithTheSameID(t *testing.T) {
	page := reported(t,
		crumbRec("ADR-0003", "ADR", "b.md"), crumbRec("ADR-0003", "ADR", "a.md"), crumbRec("verify", "spec", "v.md"),
		linkRec("follows", "ADR-0003", "v.md", "6"),
		verdictRec("Confirmed", "a.md"), verdictRec("Refuted", "b.md"))
	table := rows(t, section(t, page, "Breadcrumbs"))
	same(t, "the first row", table[0], "Refuted", `ADR-0003`, "ADR", "b.md:3")
	same(t, "the second row", table[1], "Confirmed", `ADR-0003`, "ADR", "a.md:3")
	boxes, arrows, classes := diagram(t, page, "verify")
	if len(boxes) != 3 {
		t.Errorf("boxes %q, want three", boxes)
	}
	same(t, "arrows", arrows, `n3 -- "follows" --> n1`, `n3 -- "follows" --> n2`)
	same(t, "classes", classes, "n1 confirmed", "n2 refuted", "n3 unaudited")
}

func TestReportALinkToNoBreadcrumb(t *testing.T) {
	page := reported(t, crumbRec("verify", "spec", "v.md"), linkRec("follows", "ADR-0099", "v.md", "6"))
	boxes, arrows, _ := diagram(t, page, "verify")
	if len(boxes) != 1 || len(arrows) != 0 {
		t.Errorf("boxes %q and arrows %q, want one box and no arrow", boxes, arrows)
	}
}

func TestReportNoInput(t *testing.T) {
	page := reported(t)
	for name, want := range map[string]string{
		"Breadcrumbs": "No breadcrumbs.", "Problems": "No problems.", "Claims": "No claims.", "Views": "No specs.",
	} {
		same(t, name, section(t, page, name), want)
	}
}

func TestReportTheSameRecordsTheSamePage(t *testing.T) {
	if first, second := reported(t, auditedRecords...), reported(t, auditedRecords...); first != second {
		t.Errorf("the two pages differ:\n%s\n%s", first, second)
	}
}

func TestReportTheOrderOfTheRecords(t *testing.T) {
	records := []string{
		crumbRec("verify", "spec", "v.md"), crumbRec("ADR-1", "ADR", "1.md"), crumbRec("ADR-2", "ADR", "2.md"),
		crumbRec("audit", "spec", "u.md"), crumbRec("PBI-1", "PBI", "p.md"),
		linkRec("follows", "ADR-2", "v.md", "7"), linkRec("follows", "ADR-1", "v.md", "6"),
		linkRec("follows", "ADR-1", "u.md", "6"), linkRec("amends", "ADR-1", "2.md", "6"),
		linkRec("implements", "ADR-1", "p.md", "7"), linkRec("changes", "verify", "p.md", "6"),
		verdictRec("Refuted", "v.md"), verdictRec("Confirmed", "u.md"),
	}
	var opposite []string
	for i := len(records) - 1; i >= 0; i-- {
		opposite = append(opposite, records[i])
	}
	if first, second := reported(t, records...), reported(t, opposite...); first != second {
		t.Errorf("the two pages differ:\n%s\n%s", first, second)
	}
}

// What the test of values that look like syntax reads a page with,
// and the value it gives every field.
var (
	code        = regexp.MustCompile(`#[0-9]+;`)
	barePipe    = regexp.MustCompile(`(^|[^\\])\|`)
	boxLine     = regexp.MustCompile(`^n[0-9]+\["(.*)"\]$`)
	arrowLine   = regexp.MustCompile(`^n[0-9]+ -- "(.*)" --> n[0-9]+$`)
	hostile     = "a\"b#c|d<e`f"
	hostilePath = "p/" + hostile + ".md"
)

func TestReportValuesThatLookLikeSyntax(t *testing.T) {
	page := reported(t,
		crumbRec(hostile, "spec", hostilePath), crumbRec("ADR-1", hostile, "1.md"),
		linkRec(hostile, "ADR-1", hostilePath, "6"),
		claimAt(hostile, "docs/bcr.md", hostile, hostilePath, "8"),
		"problem\t"+hostilePath+"\t9\tid "+hostile+" is wrong",
		verdictRec("Undecided", hostilePath)+"\t"+hostile)
	for _, name := range []string{"Breadcrumbs", "Claims"} {
		for _, l := range section(t, page, name) {
			if strings.HasPrefix(l, "|") && len(barePipe.FindAllString(l, -1)) != 5 {
				t.Errorf("the row %q of %s has %d bare |, want 5", l, name, len(barePipe.FindAllString(l, -1)))
			}
		}
	}
	if problems := section(t, page, "Problems"); len(problems) != 1 || strings.ContainsAny(strings.ReplaceAll(problems[0], "\\`", ""), "`") {
		t.Errorf("problems %q, want one item with no bare `", problems)
	}
	if got := strings.Count(page, "```"); got != 2 {
		t.Errorf("the page has %d fences, want one that opens and one that closes", got)
	}
	if got := strings.Count(page, "\n### "); got != 1 {
		t.Errorf("the page has %d views, want 1", got)
	}
	_, rest, _ := strings.Cut(page, "\n```mermaid\nflowchart LR\n")
	body, _, _ := strings.Cut(rest, "```\n")
	var boxes, arrows int
	for _, l := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
		l = strings.TrimSpace(l)
		var text string
		switch {
		case strings.HasPrefix(l, "class"):
			continue
		case boxLine.MatchString(l):
			boxes++
			text = boxLine.FindStringSubmatch(l)[1]
		case arrowLine.MatchString(l):
			arrows++
			text = arrowLine.FindStringSubmatch(l)[1]
		default:
			t.Errorf("the line %q of the diagram is not a box, an arrow or a class", l)
			continue
		}
		text = code.ReplaceAllString(strings.ReplaceAll(text, "<br/>", ""), "")
		if strings.ContainsAny(text, "\"#|<>`") {
			t.Errorf("the label of %q holds a character of a record: %q", l, text)
		}
	}
	if boxes != 2 || arrows != 1 {
		t.Errorf("%d boxes and %d arrows, want 2 and 1", boxes, arrows)
	}
	if !strings.Contains(body, `  n2["a#34;b#35;c#124;d#60;e#96;f<br/>`) {
		t.Errorf("the spec is not the box n2, with its id as codes:\n%s", body)
	}
}

func TestReportAValueThatAsksToNavigate(t *testing.T) {
	page := reported(t,
		crumbRec("%%{init:{}}%%", "spec", "v.md"), crumbRec("click", "ADR", "c.md"),
		linkRec("click", "click", "v.md", "6"),
		claimAt("x", "docs/bcr.md", `click n1 href "javascript:alert(1)"`, "v.md", "8"))
	_, rest, ok := strings.Cut(page, "\n```mermaid\n")
	if !ok {
		t.Fatalf("no diagram in\n%s", page)
	}
	body, _, _ := strings.Cut(rest, "```\n")
	for _, l := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "click") || strings.Contains(l, "%%") || strings.Contains(l, "href \"") {
			t.Errorf("the diagram has the line %q", l)
		}
	}
	if !strings.Contains(body, `  n2["click<br/>ADR · not audited"]`) {
		t.Errorf("the breadcrumb whose id is click is not the box n2:\n%s", body)
	}
}

func TestReportRecordsOfAKindItDoesNotRead(t *testing.T) {
	with := append([]string{"note\tanything\tat all", "note"}, auditedRecords...)
	if got, want := reported(t, with...), reported(t, auditedRecords...); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestReportFieldsAddedAtTheEnd(t *testing.T) {
	var more []string
	for _, r := range auditedRecords {
		if strings.HasPrefix(r, "verdict\t") && strings.Count(r, "\t") == 4 {
			r += "\t" // a verdict with no reason has an empty one before a new field
		}
		more = append(more, r+"\tmore")
	}
	if got, want := reported(t, more...), reported(t, auditedRecords...); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestReportWindowsLineEndingsInTheInput(t *testing.T) {
	page, stderr, code := reportRun(strings.ReplaceAll(recordsOf(auditedRecords...), "\n", "\r\n"))
	if stderr != "" || code != 0 {
		t.Fatalf("got %q, exit %d", stderr, code)
	}
	if strings.Contains(page, "\r") {
		t.Errorf("the page has a \\r:\n%q", page)
	}
	if want := reported(t, auditedRecords...); page != want {
		t.Errorf("got\n%s\nwant\n%s", page, want)
	}
}

func TestReportALineThatIsNotARecord(t *testing.T) {
	stdout, stderr, code := reportRun(recordsOf(crumbAt("A", "a.md"), "verdict\tverify\tPassed\ta.md\t3", crumbAt("B", "b.md")))
	if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "line 2 ") || code != 2 {
		t.Errorf("got %q, %q, exit %d; want a message that gives line 2, exit 2", stdout, stderr, code)
	}
}

func TestReportAnOperand(t *testing.T) {
	stdout, stderr, code := reportRun("", "records.tsv")
	if stdout != "" || !strings.HasSuffix(stderr, reportUsage) || code != 2 {
		t.Errorf("got %q, %q, exit %d; want the usage line, exit 2", stdout, stderr, code)
	}
}

func TestReportOutputThatCannotBeWritten(t *testing.T) {
	var stderr bytes.Buffer
	code := run([]string{"report"}, strings.NewReader(recordsOf(auditedRecords...)), io.Writer(unwritable{}), &stderr)
	if stderr.String() != "bcr: output is closed\n" || code != 2 {
		t.Errorf("got %q, exit %d; want one message, exit 2", stderr.String(), code)
	}
}

func TestReportTheWholePipe(t *testing.T) {
	files(t, map[string]string{
		"docs/bcr.md": "#### verify\n",
		"a.md":        fm("breadcrumb:", "  id: A", "  type: spec", "  links: []", "  claims:", "    - 'docs/bcr.md has-line ### verify'"),
	})
	extracted, _, code := extractAll("a.md")
	if code != 0 {
		t.Fatalf("bcr extract exits %d", code)
	}
	verified, _, code := verifyRun(extracted)
	if code != 0 {
		t.Fatalf("bcr verify exits %d", code)
	}
	audited, _, code := auditRun(verified)
	if code != 1 {
		t.Fatalf("bcr audit exits %d, want 1", code)
	}
	page, stderr, code := reportRun(audited)
	if stderr != "" || code != 0 {
		t.Errorf("bcr report wrote %q, exit %d; want nothing, exit 0", stderr, code)
	}
	same(t, "breadcrumbs", rows(t, section(t, page, "Breadcrumbs"))[0], "Refuted", "A", "spec", "a.md:3")
	same(t, "claims", rows(t, section(t, page, "Claims"))[0], "Refuted", "A", `docs/bcr.md has-line \#\#\# verify`, "a.md:7")
	boxes, _, classes := diagram(t, page, "A")
	same(t, "boxes", boxes, `n1["A<br/>spec · Refuted<br/>Refuted: docs/bcr.md has-line #35;#35;#35; verify"]`)
	same(t, "classes", classes, "n1 refuted")
}
