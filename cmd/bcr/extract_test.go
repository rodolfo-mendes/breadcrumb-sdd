package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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

// extractRun runs bcr extract with operands, and returns what it
// writes to standard output and standard error, and its exit status.
func extractRun(operands ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"extract"}, operands...), nil, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
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
	want := "breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\n" +
		"link\tPBI-00001\timplements\tADR-0001\n" +
		"link\tPBI-00001\tchanges\tasdlc\n"
	if stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q, nothing, exit 0", stdout, stderr, code, want)
	}
}

func TestExtractABreadcrumbWithNoLinks(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A")})
	stdout, stderr, code := extractRun("a.md")
	if stdout != "breadcrumb\tA\tPBI\ta.md\n" || stderr != "" || code != 0 {
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
	want := "breadcrumb\tB\tPBI\tb.md\n" +
		"breadcrumb\tA\tPBI\ta.md\n" +
		"link\tA\tfollows\tB\n"
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
	if !strings.HasSuffix(stdout, "\nlink\tA\timplements\tADR-0001\n") || stderr != "" || code != 0 {
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
	files(t, map[string]string{"a.md": fm("base: &b X", "breadcrumb:", "  id: *b", "  type: PBI", "  links: []")})
	problems(t, "a.md", "4")
}

func TestExtractAValueThatLooksLikeANumber(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("0001")})
	if stdout, _, _ := extractRun("a.md"); stdout != "breadcrumb\t0001\tPBI\ta.md\n" {
		t.Errorf("got %q", stdout)
	}
}

func TestExtractOneBadFileAmongGoodOnes(t *testing.T) {
	files(t, map[string]string{"a.md": crumbOf("A"), "b.md": fm("breadcrumb:", "  id: B")})
	stdout, stderr, code := extractRun("a.md", "b.md")
	if stdout != "breadcrumb\tA\tPBI\ta.md\n" {
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
		if stdout != "breadcrumb\tA\tPBI\ta.md\n" {
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

func TestExtractNoOperand(t *testing.T) {
	_, stderr, code := extractRun()
	if !strings.HasSuffix(stderr, "\n"+extractUsage) || code != 2 {
		t.Errorf("got %q, exit %d", stderr, code)
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
