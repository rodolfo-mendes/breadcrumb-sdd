package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// The tests below follow the Scenarios of specs/extract/spec.md, one
// test each, in the same order.

// files writes each file, named by a path relative to a temporary
// directory, and runs from that directory.
func files(t *testing.T, contents map[string]string) {
	t.Helper()
	root := t.TempDir()
	for name, text := range contents {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, root)
}

// extractAll runs bcr extract with operands, and returns what it
// writes to standard output and standard error, and its exit status.
func extractAll(operands ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"extract"}, operands...), nil, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// problemLine reads a PATH:LINE: MESSAGE of standard error.
var problemLine = regexp.MustCompile(`^(.*?):([0-9]+): (.*)$`)

// extractRun is extractAll without the problem records of standard
// output, so that a test can read the other records alone. Every
// problem of standard error must also be a problem record, in the same
// order (ADR-0021): extractRun panics when they differ.
func extractRun(operands ...string) (string, string, int) {
	stdout, stderr, code := extractAll(operands...)
	var records, got, want []string
	for _, l := range strings.SplitAfter(stdout, "\n") {
		if strings.HasPrefix(l, "problem\t") {
			got = append(got, l)
		} else {
			records = append(records, l)
		}
	}
	for _, l := range strings.Split(stderr, "\n") {
		if m := problemLine.FindStringSubmatch(l); m != nil && !strings.HasPrefix(l, "bcr: ") {
			want = append(want, "problem\t"+m[1]+"\t"+m[2]+"\t"+m[3]+"\n")
		}
	}
	if !reflect.DeepEqual(got, want) {
		panic(fmt.Sprintf("problem records %q, want %q, as standard error has", got, want))
	}
	return strings.Join(records, ""), stderr, code
}

// fm returns a file whose front matter holds lines.
func fm(lines ...string) string {
	return "---\n" + strings.Join(lines, "\n") + "\n---\n# A title\n"
}

// crumbOf returns a file with a valid breadcrumb, whose id is id.
func crumbOf(id string, links ...string) string {
	lines := []string{"breadcrumb:", "  id: " + id, "  type: PBI"}
	if len(links) == 0 {
		return fm(append(lines, "  links: []")...)
	}
	lines = append(lines, "  links:")
	for _, l := range links {
		lines = append(lines, "    - "+l)
	}
	return fm(lines...)
}

// problemLines returns the line of each PATH:LINE: MESSAGE in stderr,
// and fails unless each is about path.
func problemLines(t *testing.T, path, stderr string) []string {
	t.Helper()
	var lines []string
	for _, l := range strings.Split(strings.TrimSuffix(stderr, "\n"), "\n") {
		rest, ok := strings.CutPrefix(l, path+":")
		if !ok {
			t.Errorf("problem %q is not about %s", l, path)
			continue
		}
		line, _, _ := strings.Cut(rest, ":")
		lines = append(lines, line)
	}
	return lines
}

// problems runs bcr extract on path, and fails unless it prints no
// records, problems at want, and exits 1.
func problems(t *testing.T, path string, want ...string) {
	t.Helper()
	stdout, stderr, code := extractRun(path)
	if stdout != "" {
		t.Errorf("printed records %q", stdout)
	}
	if got := problemLines(t, path, stderr); !reflect.DeepEqual(got, want) {
		t.Errorf("problems at lines %q, want %q:\n%s", got, want, stderr)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestExtractABreadcrumbWithLinks(t *testing.T) {
	files(t, map[string]string{"tasks/PBI-00001.md": crumbOf("PBI-00001", "implements ADR-0001", "changes asdlc")})
	stdout, stderr, code := extractRun("tasks/PBI-00001.md")
	want := "breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3\n" +
		"link\tPBI-00001\timplements\tADR-0001\ttasks/PBI-00001.md\t6\n" +
		"link\tPBI-00001\tchanges\tasdlc\ttasks/PBI-00001.md\t7\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractABreadcrumbWithNoLinks(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A")})
	stdout, stderr, code := extractRun("a.md")
	if stdout != "breadcrumb\tA\tPBI\ta.md\t3\n" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d", stdout, stderr, code)
	}
}

func TestExtractAFileWithNoBreadcrumb(t *testing.T) {
	files(t, map[string]string{
		"none.md":   "# No front matter\n",
		"other.md":  fm("title: no breadcrumb"),
		"sound.mp3": "ID3\x04\x00\x00\x00\x00\x00\x00\xff\xfb",
	})
	stdout, stderr, code := extractRun("none.md", "other.md", "sound.mp3")
	if stdout != "" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want nothing, exit 0", stdout, stderr, code)
	}
}

func TestExtractSeveralFiles(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A", "follows B"), "b.md": crumbOf("B")})
	stdout, _, _ := extractRun("b.md", "a.md")
	want := "breadcrumb\tB\tPBI\tb.md\t3\n" +
		"breadcrumb\tA\tPBI\ta.md\t3\n" +
		"link\tA\tfollows\tB\ta.md\t6\n"
	if stdout != want {
		t.Errorf("got %q, want %q", stdout, want)
	}
}

func TestExtractAMissingProperty(t *testing.T) {
	for _, missing := range []string{"id", "type", "links"} {
		lines := []string{"title: T", "breadcrumb:"}
		for _, p := range []string{"  id: A", "  type: PBI", "  links: []"} {
			if !strings.HasPrefix(p, "  "+missing+":") {
				lines = append(lines, p)
			}
		}
		files(t, map[string]string{"a.md": fm(lines...)})
		problems(t, "a.md", "3")
	}
}

func TestExtractAnIdOrATypeThatIsNotOneWord(t *testing.T) {
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  id:", "  type: P B I", "  links: []")})
	problems(t, "a.md", "3", "4")
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  id:", "    - A", "  type: PBI", "  links: []")})
	problems(t, "a.md", "3")
}

func TestExtractLinksWithNoList(t *testing.T) {
	for _, links := range []string{"  links:", "  links: implements ADR-0001"} {
		files(t, map[string]string{"a.md": fm("breadcrumb:", "  id: A", "  type: PBI", links)})
		problems(t, "a.md", "5")
	}
}

func TestExtractWhiteSpaceInALinkEntry(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A", "implements \t\u00a0 ADR-0001\u00a0")})
	stdout, stderr, code := extractRun("a.md")
	if !strings.HasSuffix(stdout, "\nlink\tA\timplements\tADR-0001\ta.md\t6\n") || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d", stdout, stderr, code)
	}
}

func TestExtractALinkEntryThatIsNotTwoWords(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A", "ADR-0001", "implements ADR-0002", "depends on ADR-0003")})
	problems(t, "a.md", "6", "8")
}

func TestExtractTheSameLinkTwice(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A", "implements ADR-0001", "implements  ADR-0001")})
	problems(t, "a.md", "7")
}

func TestExtractALinkToItself(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("PBI-00001", "implements PBI-00001")})
	problems(t, "a.md", "6")
}

func TestExtractYAMLOutsideThePartADR0002Allows(t *testing.T) {
	files(t, map[string]string{"a.md": fm(
		"breadcrumb:",
		"  id: &a A",
		"  type: !!str PBI",
		"  links:",
		`    - "implements ADR-0001"`,
		"    - >",
		"      implements ADR-0002",
		"  owner: [x]",
	)})
	problems(t, "a.md", "3", "4", "6", "7", "9")
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  id: 'A'", "  type: PBI", "  links:", "    - 'implements ADR-0001'")})
	problems(t, "a.md", "3", "6")
	files(t, map[string]string{"a.md": fm("base: &b X", "breadcrumb:", "  id: *b", "  type: PBI", "  links: []")})
	problems(t, "a.md", "4")
}

func TestExtractAValueThatLooksLikeANumber(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("0001")})
	if stdout, _, _ := extractRun("a.md"); stdout != "breadcrumb\t0001\tPBI\ta.md\t3\n" {
		t.Errorf("got %q", stdout)
	}
}

func TestExtractOneBadFileAmongGoodOnes(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A"), "b.md": fm("breadcrumb:", "  id: B")})
	stdout, stderr, code := extractRun("a.md", "b.md")
	if stdout != "breadcrumb\tA\tPBI\ta.md\t3\n" {
		t.Errorf("got %q", stdout)
	}
	if !strings.HasPrefix(stderr, "b.md:2: ") || code != 1 {
		t.Errorf("got %q, exit %d; want b.md's problem, exit 1", stderr, code)
	}
}

func TestExtractAFileThatCannotBeRead(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A"), "b.md": fm("breadcrumb:"), "dir/c.md": crumbOf("C")})
	for _, missing := range []string{"missing.md", "dir"} {
		stdout, stderr, code := extractRun(missing, "b.md", "a.md")
		if stdout != "breadcrumb\tA\tPBI\ta.md\t3\n" {
			t.Errorf("%s: got %q", missing, stdout)
		}
		if !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "\nb.md:2: ") {
			t.Errorf("%s: wrote %q to standard error", missing, stderr)
		}
		if code != 2 {
			t.Errorf("%s: exit %d, want 2", missing, code)
		}
	}
}

func TestExtractNoOperandAndNoBreadcrumbsFile(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A")})
	for _, operands := range [][]string{nil, {"--"}} {
		stdout, stderr, code := extractRun(operands...)
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || code != 2 {
			t.Errorf("got %q, %q, exit %d; want a message, exit 2", stdout, stderr, code)
		}
	}
}

func TestExtractTheFilesBreadcrumbsNames(t *testing.T) {
	files(t, map[string]string{
		".breadcrumbs":          "# the decisions and the tasks\ntasks/*.md\n\ndocs/adrs/*.md\n",
		"docs/adrs/ADR-0002.md": crumbOf("ADR-0002"),
		"tasks/PBI-00001.md":    crumbOf("PBI-00001", "implements ADR-0002"),
		"specs/asdlc/spec.md":   crumbOf("asdlc"),
	})
	stdout, stderr, code := extractRun()
	want := "breadcrumb\tADR-0002\tPBI\tdocs/adrs/ADR-0002.md\t3\n" +
		"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3\n" +
		"link\tPBI-00001\timplements\tADR-0002\ttasks/PBI-00001.md\t6\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractAPatternThatExcludes(t *testing.T) {
	files(t, map[string]string{
		".breadcrumbs":       "tasks/*.md\n!tasks/PBI-00009.md\n",
		"tasks/PBI-00001.md": crumbOf("PBI-00001"),
		"tasks/PBI-00009.md": crumbOf("PBI-00009"),
		"tasks/PBI-00010.md": crumbOf("PBI-00010"),
	})
	stdout, stderr, code := extractRun()
	want := "breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3\n" +
		"breadcrumb\tPBI-00010\tPBI\ttasks/PBI-00010.md\t3\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractAPatternWithoutASlash(t *testing.T) {
	files(t, map[string]string{
		".breadcrumbs":          "*.md\n",
		"README.md":             crumbOf("README"),
		"docs/adrs/ADR-0002.md": crumbOf("ADR-0002"),
	})
	stdout, stderr, code := extractRun()
	if stdout != "breadcrumb\tREADME\tPBI\tREADME.md\t3\n" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d", stdout, stderr, code)
	}
}

func TestExtractAPatternBreadcrumbsDoesNotAllow(t *testing.T) {
	for _, pattern := range []string{"docs/**/*.md", "docs/", "/docs/*.md", "!", "docs/[a.md"} {
		files(t, map[string]string{
			".breadcrumbs": "tasks/*.md\n" + pattern + "\n",
			"tasks/a.md":   crumbOf("A"),
			"tasks/b.md":   fm("breadcrumb:"),
		})
		stdout, stderr, code := extractRun()
		if stdout != "" {
			t.Errorf("%s: printed records %q", pattern, stdout)
		}
		if got := problemLines(t, ".breadcrumbs", stderr); !reflect.DeepEqual(got, []string{"2"}) {
			t.Errorf("%s: problems at lines %q, want line 2:\n%s", pattern, got, stderr)
		}
		if code != 1 {
			t.Errorf("%s: exit %d, want 1", pattern, code)
		}
	}
}

func TestExtractWhatTheWalkDoesNotEnter(t *testing.T) {
	files(t, map[string]string{
		".breadcrumbs":  "*/*.md\n",
		".git/x.md":     crumbOf("X"),
		"docs/a.md":     crumbOf("A"),
		"other/b.notes": crumbOf("B"),
	})
	if err := os.Symlink(filepath.Join("..", "other", "b.notes"), filepath.Join("docs", "link.md")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	if err := os.Symlink("other", "linked"); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := extractRun()
	if stdout != "breadcrumb\tA\tPBI\tdocs/a.md\t3\n" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d", stdout, stderr, code)
	}
}

func TestExtractOperandsAndBreadcrumbs(t *testing.T) {
	files(t, map[string]string{
		".breadcrumbs":   "tasks/*.md\ndocs/**\n",
		"tasks/a.md":     crumbOf("A"),
		"notes/draft.md": crumbOf("DRAFT"),
	})
	stdout, stderr, code := extractRun("notes/draft.md")
	if stdout != "breadcrumb\tDRAFT\tPBI\tnotes/draft.md\t3\n" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d", stdout, stderr, code)
	}
}

func TestExtractFrontMatterThatIsNotYAML(t *testing.T) {
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  links: [a")})
	problems(t, "a.md", "3")
}

func TestExtractAKeyWrittenTwice(t *testing.T) {
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  id: A", "  type: PBI", "  id: B", "  links: []")})
	problems(t, "a.md", "5")
}

func TestExtractWindowsLineEndings(t *testing.T) {
	text := crumbOf("A", "implements ADR-0001")
	files(t, map[string]string{"a.md": text, "b.md": "\xEF\xBB\xBF" + strings.ReplaceAll(text, "\n", "\r\n")})
	a, _, _ := extractRun("a.md")
	b, stderr, code := extractRun("b.md")
	if strings.ReplaceAll(a, "a.md", "b.md") != b || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q", b, stderr, code, a)
	}
}

func TestExtractFrontMatterThatDoesNotClose(t *testing.T) {
	files(t, map[string]string{"a.md": "---\nbreadcrumb:\n  id: A\n"})
	problems(t, "a.md", "1")
}

// claimsOf returns a file with a valid breadcrumb, whose id is A, with
// one link on line 6 and claims as the entries of its claims, the
// first on line 8.
func claimsOf(claims ...string) string {
	lines := []string{"breadcrumb:", "  id: A", "  type: spec", "  links:", "    - follows ADR-0016", "  claims:"}
	for _, c := range claims {
		lines = append(lines, "    - "+c)
	}
	return fm(lines...)
}

// The records of the breadcrumb and the link of claimsOf, in a.md.
const claimsOfRecords = "breadcrumb\tA\tspec\ta.md\t3\n" + "link\tA\tfollows\tADR-0016\ta.md\t6\n"

// droppedClaims writes a.md with claims, runs bcr extract on it, and
// fails unless it prints the breadcrumb and link records, then the
// claim records wantClaims, problems at wantLines, and exits 1. It
// returns standard error.
func droppedClaims(t *testing.T, claims []string, wantLines []string, wantClaims ...string) string {
	t.Helper()
	files(t, map[string]string{"a.md": claimsOf(claims...)})
	stdout, stderr, code := extractRun("a.md")
	if want := claimsOfRecords + strings.Join(wantClaims, ""); stdout != want {
		t.Errorf("got %q, want %q", stdout, want)
	}
	if got := problemLines(t, "a.md", stderr); !reflect.DeepEqual(got, wantLines) {
		t.Errorf("problems at lines %q, want %q:\n%s", got, wantLines, stderr)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	return stderr
}

func TestExtractABreadcrumbWithClaims(t *testing.T) {
	files(t, map[string]string{"specs/verify/spec.md": fm(
		"breadcrumb:",
		"  id: verify",
		"  type: spec",
		"  links:",
		"    - follows ADR-0016",
		"  claims:",
		"    - 'docs/bcr.md has-line ### verify'",
		"    - 'docs/adrs/ADR-0001-adopt-asdlc.md has-line Status: Accepted'",
	)})
	stdout, stderr, code := extractRun("specs/verify/spec.md")
	want := "breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3\n" +
		"link\tverify\tfollows\tADR-0016\tspecs/verify/spec.md\t6\n" +
		"claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8\n" +
		"claim\tverify\thas-line\tdocs/adrs/ADR-0001-adopt-asdlc.md\tStatus: Accepted\tspecs/verify/spec.md\t9\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractNoClaims(t *testing.T) {
	files(t, map[string]string{
		"a.md": crumbOf("A", "follows ADR-0016"),
		"b.md": fm("breadcrumb:", "  id: B", "  type: PBI", "  links:", "    - follows ADR-0016", "  claims: []"),
	})
	stdout, stderr, code := extractRun("a.md", "b.md")
	want := "breadcrumb\tA\tPBI\ta.md\t3\n" + "link\tA\tfollows\tADR-0016\ta.md\t6\n" +
		"breadcrumb\tB\tPBI\tb.md\t3\n" + "link\tB\tfollows\tADR-0016\tb.md\t6\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractClaimsWithNoList(t *testing.T) {
	for _, claims := range []string{"  claims:", "  claims: docs/bcr.md has-line x"} {
		files(t, map[string]string{"a.md": fm("breadcrumb:", "  id: A", "  type: PBI", "  links: []", claims)})
		problems(t, "a.md", "6")
	}
}

func TestExtractAQuoteInsideAClaim(t *testing.T) {
	files(t, map[string]string{"a.md": claimsOf("'cmd/bcr/main.go has-line return fmt.Sprintf(''%s'', id)'")})
	stdout, stderr, code := extractRun("a.md")
	want := claimsOfRecords + "claim\tA\thas-line\tcmd/bcr/main.go\treturn fmt.Sprintf('%s', id)\ta.md\t8\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractAClaimEntryThatIsNotSingleQuoted(t *testing.T) {
	droppedClaims(t, []string{"docs/bcr.md has-line verify", `"docs/bcr.md has-line ### verify"`}, []string{"8", "9"})
}

func TestExtractAClaimWithACommentAfterIt(t *testing.T) {
	droppedClaims(t, []string{"docs/bcr.md has-line ### verify", "'docs/bcr.md has-line ### verify' # why"}, []string{"8", "9"})
}

func TestExtractAClaimEntryThatIsNotThreeParts(t *testing.T) {
	droppedClaims(t, []string{"'docs/bcr.md'", "'docs/bcr.md has-line'", "'docs/bcr.md  has-line ### verify'"},
		[]string{"8", "9", "10"})
}

func TestExtractATargetOutsideTheRules(t *testing.T) {
	droppedClaims(t, []string{"'/etc/passwd has-line x'", "'../x.md has-line x'", "'docs/../x.md has-line x'"},
		[]string{"8", "9", "10"})
}

func TestExtractAKindBcrDoesNotKnow(t *testing.T) {
	stderr := droppedClaims(t, []string{"'docs/bcr.md contains ### verify'", "'docs/bcr.md has-lines ### verify'"},
		[]string{"8", "9"})
	for _, kind := range []string{`kind "contains"`, `kind "has-lines"`} {
		if !strings.Contains(stderr, kind) {
			t.Errorf("no problem names the %s:\n%s", kind, stderr)
		}
	}
}

func TestExtractAHasLineTextThatNoLineCanEqual(t *testing.T) {
	droppedClaims(t, []string{"'docs/bcr.md has-line  ### verify'", "'docs/bcr.md has-line ### verify '", "'docs/bcr.md has-line a\tb'"},
		[]string{"8", "9", "10"})
}

func TestExtractAnInvalidClaimDropsOnlyItself(t *testing.T) {
	files(t, map[string]string{"a.md": fm(
		"breadcrumb:",
		"  id: A",
		"  type: spec",
		"  links:",
		"    - follows ADR-0016",
		"    - follows ADR-0017",
		"  claims:",
		"    - 'docs/bcr.md has-line ### extract'",
		"    - 'docs/bcr.md contains ### verify'",
		"    - 'docs/bcr.md has-line ### verify'",
	)})
	stdout, stderr, code := extractRun("a.md")
	want := "breadcrumb\tA\tspec\ta.md\t3\n" +
		"link\tA\tfollows\tADR-0016\ta.md\t6\n" +
		"link\tA\tfollows\tADR-0017\ta.md\t7\n" +
		"claim\tA\thas-line\tdocs/bcr.md\t### extract\ta.md\t9\n" +
		"claim\tA\thas-line\tdocs/bcr.md\t### verify\ta.md\t11\n"
	if stdout != want {
		t.Errorf("got %q, want %q", stdout, want)
	}
	if got := problemLines(t, "a.md", stderr); !reflect.DeepEqual(got, []string{"10"}) || code != 1 {
		t.Errorf("problems at lines %q, exit %d; want line 10, exit 1:\n%s", got, code, stderr)
	}
}

func TestExtractAClaimWhoseTargetDoesNotExist(t *testing.T) {
	files(t, map[string]string{"a.md": claimsOf("'no/such/file.md has-line x'")})
	stdout, stderr, code := extractRun("a.md")
	if want := claimsOfRecords + "claim\tA\thas-line\tno/such/file.md\tx\ta.md\t8\n"; stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractAProblemIsAlsoARecord(t *testing.T) {
	files(t, map[string]string{"tasks/PBI-00001.md": fm("breadcrumb:", "  id: PBI-00001", "  links: []")})
	stdout, stderr, code := extractAll("tasks/PBI-00001.md")
	if stderr != "tasks/PBI-00001.md:2: breadcrumb has no type\n" {
		t.Errorf("standard error %q", stderr)
	}
	if stdout != "problem\ttasks/PBI-00001.md\t2\tbreadcrumb has no type\n" {
		t.Errorf("standard output %q", stdout)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestExtractTheRecordOfADroppedClaim(t *testing.T) {
	files(t, map[string]string{"a.md": fm("breadcrumb:", "  id: A", "  type: spec", "  links:", "    - follows ADR-0016",
		"  claims:", "    - 'docs/bcr.md contains ### verify'", "    - 'docs/bcr.md has-line ### verify'")})
	stdout, _, code := extractAll("a.md")
	want := "breadcrumb\tA\tspec\ta.md\t3\n" +
		"link\tA\tfollows\tADR-0016\ta.md\t6\n" +
		"claim\tA\thas-line\tdocs/bcr.md\t### verify\ta.md\t9\n" +
		"problem\ta.md\t8\t"
	if !strings.HasPrefix(stdout, want) || strings.Count(stdout, "\n") != 4 {
		t.Errorf("got %q, want %q and a message", stdout, want)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestExtractProblemRecordsAmongSeveralFiles(t *testing.T) {
	files(t, map[string]string{
		"a.md": fm("breadcrumb:", "  id: a b", "  type: PBI", "  links: x"),
		"b.md": crumbOf("B"),
	})
	stdout, _, _ := extractAll("a.md", "b.md")
	lines := strings.Split(stdout, "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[0], "problem\ta.md\t3\t") ||
		!strings.HasPrefix(lines[1], "problem\ta.md\t5\t") || lines[2] != "breadcrumb\tB\tPBI\tb.md\t3" {
		t.Errorf("got %q", stdout)
	}
}

func TestExtractAMessageThatIsNotAProblem(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A")})
	stdout, stderr, code := extractAll("none.md", "a.md")
	if !strings.HasPrefix(stderr, "bcr: ") || strings.Contains(stdout, "problem") || code != 2 {
		t.Errorf("got %q, %q, exit %d; want a message, no problem record, exit 2", stdout, stderr, code)
	}
}

// titled returns a file with a valid breadcrumb whose id is A, with
// title, the whole line of its title key, after its type.
func titled(title ...string) string {
	return fm(append(append([]string{"breadcrumb:", "  id: A", "  type: ADR"}, title...), "  links: []")...)
}

func TestExtractABreadcrumbWithATitle(t *testing.T) {
	files(t, map[string]string{"a.md": titled("  title: Adopt ASDLC to develop Breadcrumb")})
	stdout, stderr, code := extractRun("a.md")
	if want := "breadcrumb\tA\tADR\ta.md\t3\tAdopt ASDLC to develop Breadcrumb\n"; stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractNoTitle(t *testing.T) {
	files(t, map[string]string{"a.md": titled()})
	stdout, stderr, code := extractRun("a.md")
	if want := "breadcrumb\tA\tADR\ta.md\t3\n"; stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractATitleThatIsNotOneLineOfText(t *testing.T) {
	for _, title := range [][]string{
		{"  title:"},
		{"  title: []"},
		{"  title:", "    - Adopt ASDLC"},
		{"  title: Adopt\tASDLC"},
		{"  title: 'Adopt ASDLC'"},
		{"  title: Adopt", "    ASDLC"},
	} {
		files(t, map[string]string{"a.md": titled(title...)})
		problems(t, "a.md", "5")
	}
}
