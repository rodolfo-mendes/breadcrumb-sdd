package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// The tests below follow the Scenarios of specs/init/spec.md, one test
// each, in the same order.

// initRun runs bcr init in dir as a bcr released as release, and
// returns what it writes to standard output and standard error, and
// its exit status.
func initRun(dir, release string, flags ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := initIn(dir, release, append([]string{"init"}, flags...), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// read returns the content of name under dir, or "" when it is not
// there.
func read(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// write writes text to name under dir.
func write(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// filesIn lists every file under dir, as paths from dir.
func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	var list []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			list = append(list, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return list
}

func TestInitWritesEveryPiece(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := initRun(dir, "0.8.0")
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, standard error %q; want 0 and nothing", code, stderr)
	}
	want := ".breadcrumbs\nAGENTS.md\n.github/workflows/breadcrumbs.yml\n"
	if stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
	if read(t, dir, ".breadcrumbs") != "# bcr init, bcr 0.8.0\n"+patterns {
		t.Error(".breadcrumbs is not the patterns of bcr init, after its first line")
	}
	if read(t, dir, "AGENTS.md") != agentsSection {
		t.Error("AGENTS.md is not the agents section alone")
	}
	if !strings.Contains(read(t, dir, workflowFile), "bcr 0.8.0") {
		t.Error("the workflow does not name bcr 0.8.0")
	}
}

func TestInitPatternsNameNoFile(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.8.0")
	text := read(t, dir, ".breadcrumbs")
	if first, _, _ := strings.Cut(text, "\n"); first != "# bcr init, bcr 0.8.0" {
		t.Errorf("the first line of .breadcrumbs is %q", first)
	}
	for i, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			t.Errorf(".breadcrumbs line %d is a pattern: %q", i+1, line)
		}
	}
}

func TestInitWorkflowRunsTheBcrThatWroteIt(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.8.0")
	text := read(t, dir, workflowFile)
	for _, s := range []string{
		releaseURL + "v0.8.0/bcr_0.8.0_linux_amd64.tar.gz",
		releaseURL + "v0.8.0/bcr_0.8.0_checksums.txt",
		`sha256sum --check --ignore-missing "bcr_0.8.0_checksums.txt"`,
		"set -o pipefail",
		"bcr extract | bcr verify | bcr audit | bcr report",
	} {
		if !strings.Contains(text, s) {
			t.Errorf("the workflow lacks %q", s)
		}
	}
	if strings.Contains(text, "{{bcr}}") || strings.Contains(text, "{{release}}") {
		t.Error("the workflow has a placeholder left")
	}

	var wf struct {
		Jobs map[string]struct {
			Steps []struct {
				Uses string `yaml:"uses"`
				Run  string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(text), &wf); err != nil {
		t.Fatalf("the workflow is not valid YAML: %v", err)
	}
	pinned := regexp.MustCompile(`^[a-z0-9-]+/[a-z0-9-]+@[0-9a-f]{40}$`)
	uses := 0
	for _, job := range wf.Jobs {
		for _, step := range job.Steps {
			if step.Uses == "" {
				continue
			}
			uses++
			if !pinned.MatchString(step.Uses) {
				t.Errorf("action %q is not pinned to a commit", step.Uses)
			}
		}
	}
	if uses != 2 {
		t.Errorf("the workflow uses %d actions, want 2", uses)
	}
}

// lookup reads a command the agents section shows: a text in
// backticks that starts with bcr extract and ends with bcr filter.
var lookup = regexp.MustCompile("`(bcr extract \\|[^`]*bcr filter [^`]*)`")

func TestInitSectionShowsHowToLookBreadcrumbsUp(t *testing.T) {
	var commands []string
	for _, m := range lookup.FindAllStringSubmatch(agentsSection, -1) {
		commands = append(commands, m[1])
	}
	for _, flags := range []string{"--id ID`", "--type TYPE`", "--object ID`", "--kind verdict --id ID`"} {
		if !strings.Contains(agentsSection, "bcr filter "+flags) {
			t.Errorf("the section shows no bcr filter %s", strings.TrimSuffix(flags, "`"))
		}
	}
	// Each command runs, in a repository where it finds something.
	files(t, map[string]string{
		".breadcrumbs": "*.md\n",
		"a.md":         crumbOf("PBI-00001", "implements PBI-00002"),
		"b.md":         crumbOf("PBI-00002"),
	})
	for _, command := range commands {
		command = strings.NewReplacer("ID", "PBI-00002", "TYPE", "PBI").Replace(command)
		input := ""
		for _, stage := range strings.Split(command, " | ") {
			args := strings.Fields(stage)
			var stdout, stderr bytes.Buffer
			if code := run(args[1:], strings.NewReader(input), &stdout, &stderr); args[0] != "bcr" || code != 0 || stderr.Len() > 0 {
				t.Fatalf("%q: %q exits %d, writes %q", command, stage, code, stderr.String())
			}
			input = stdout.String()
		}
		if input == "" {
			t.Errorf("%q prints nothing", command)
		}
	}
	if len(commands) != 4 {
		t.Errorf("got %d commands, want 4: %q", len(commands), commands)
	}
}

func TestInitAddsTheSectionToAnAgentsFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "AGENTS.md", "# Agents")
	stdout, _, code := initRun(dir, "0.8.0")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if got, want := read(t, dir, "AGENTS.md"), "# Agents\n\n"+agentsSection; got != want {
		t.Errorf("AGENTS.md is %q, want %q", got, want)
	}
	if !strings.Contains(stdout, "AGENTS.md\n") {
		t.Errorf("standard output %q does not name AGENTS.md", stdout)
	}
}

func TestInitLeavesAnAgentsFileWithTheSection(t *testing.T) {
	dir := t.TempDir()
	old := "# Agents\n\n## Breadcrumbs\n\nOurs.\n"
	write(t, dir, "AGENTS.md", old)
	stdout, stderr, code := initRun(dir, "0.8.0")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if read(t, dir, "AGENTS.md") != old {
		t.Error("AGENTS.md changed")
	}
	if !strings.Contains(stderr, "AGENTS.md already has a ## Breadcrumbs section; left as it is") {
		t.Errorf("standard error %q does not say AGENTS.md was left", stderr)
	}
	if want := ".breadcrumbs\n.github/workflows/breadcrumbs.yml\n"; stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
}

func TestInitWritesAnotherAgentsFile(t *testing.T) {
	dir := t.TempDir()
	if _, _, code := initRun(dir, "0.8.0", "-a", "CLAUDE.md"); code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if read(t, dir, "CLAUDE.md") != agentsSection {
		t.Error("CLAUDE.md is not the agents section alone")
	}
	if read(t, dir, "AGENTS.md") != "" {
		t.Error("AGENTS.md was written")
	}
}

func TestInitAgainChangesNothing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "AGENTS.md", "# Agents\n")
	initRun(dir, "0.8.0")
	before := map[string]string{}
	for _, f := range filesIn(t, dir) {
		before[f] = read(t, dir, f)
	}
	stdout, stderr, code := initRun(dir, "0.8.0")
	if code != 0 || stdout != "" {
		t.Fatalf("exit %d, standard output %q; want 0 and nothing", code, stdout)
	}
	if n := strings.Count(stderr, "left as it is\n"); n != 3 {
		t.Errorf("standard error has %d messages, want 3: %q", n, stderr)
	}
	after := filesIn(t, dir)
	if len(after) != len(before) {
		t.Fatalf("files %q, want %d", after, len(before))
	}
	for _, f := range after {
		if read(t, dir, f) != before[f] {
			t.Errorf("%s changed", f)
		}
	}
}

func TestInitLeavesAPieceSetUpByHand(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".breadcrumbs", "docs/*.md\n")
	stdout, stderr, code := initRun(dir, "0.8.0")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if read(t, dir, ".breadcrumbs") != "docs/*.md\n" {
		t.Error(".breadcrumbs changed")
	}
	if want := "AGENTS.md\n.github/workflows/breadcrumbs.yml\n"; stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
	if !strings.Contains(stderr, ".breadcrumbs is already there; left as it is") {
		t.Errorf("standard error %q does not say .breadcrumbs was left", stderr)
	}
}

func TestInitNeedsAReleasedBcr(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := initRun(dir, "")
	if code != 2 || stdout != "" {
		t.Errorf("exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.Contains(stderr, "no release version") {
		t.Errorf("standard error %q does not say bcr has no release version", stderr)
	}
	if f := filesIn(t, dir); len(f) > 0 {
		t.Errorf("wrote %q", f)
	}
}

func TestInitAgentsFileThatIsNotAFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := initRun(dir, "0.8.0")
	if code != 2 || stdout != "" {
		t.Errorf("exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.HasPrefix(stderr, "bcr: ") {
		t.Errorf("standard error %q does not start with bcr: ", stderr)
	}
	if f := filesIn(t, dir); len(f) > 0 {
		t.Errorf("wrote %q", f)
	}
}

func TestInitPathOutOfTheRepository(t *testing.T) {
	for _, agents := range []string{"../AGENTS.md", "/AGENTS.md", "docs/../../AGENTS.md", ""} {
		dir := t.TempDir()
		stdout, stderr, code := initRun(dir, "0.8.0", "-a", agents)
		if code != 2 || stdout != "" {
			t.Errorf("-a %q: exit %d, standard output %q; want 2 and nothing", agents, code, stdout)
		}
		if !strings.HasSuffix(stderr, "\n"+initUsage) {
			t.Errorf("-a %q: standard error %q does not end with the usage line", agents, stderr)
		}
		if f := filesIn(t, dir); len(f) > 0 {
			t.Errorf("-a %q: wrote %q", agents, f)
		}
	}
}

func TestInitOperand(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := initRun(dir, "0.8.0", "x")
	if code != 2 || stdout != "" {
		t.Errorf("exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.HasSuffix(stderr, "\n"+initUsage) {
		t.Errorf("standard error %q does not end with the usage line", stderr)
	}
	if f := filesIn(t, dir); len(f) > 0 {
		t.Errorf("wrote %q", f)
	}
}

// The templates of the ASDLC layout, in the order bcr init writes them.
var asdlcTemplates = []string{"docs/adrs/TEMPLATE.md", "specs/TEMPLATE.md", "tasks/TEMPLATE.md"}

// command runs bcr with args and stdin, and returns what it writes to
// standard output and standard error, and its exit status.
func command(stdin string, args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(args, strings.NewReader(stdin), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// unchanged fails unless the files under dir are those of before.
func unchanged(t *testing.T, dir string, before map[string]string) {
	t.Helper()
	after := filesIn(t, dir)
	if len(after) != len(before) {
		t.Errorf("files %q, want %d", after, len(before))
	}
	for _, f := range after {
		if read(t, dir, f) != before[f] {
			t.Errorf("%s changed", f)
		}
	}
}

// contents returns the content of each file under dir, by path.
func contents(t *testing.T, dir string) map[string]string {
	t.Helper()
	m := map[string]string{}
	for _, f := range filesIn(t, dir) {
		m[f] = read(t, dir, f)
	}
	return m
}

func TestInitASDLCLayout(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := initRun(dir, "0.9.0", "--layout", "asdlc")
	if code != 0 || stderr != "" {
		t.Fatalf("exit %d, standard error %q; want 0 and nothing", code, stderr)
	}
	want := ".breadcrumbs\nbreadcrumb.rules\ndocs/adrs/TEMPLATE.md\nspecs/TEMPLATE.md\ntasks/TEMPLATE.md\nAGENTS.md\n.github/workflows/breadcrumbs.yml\n"
	if stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
}

// specBlock returns the text of the code block of specs/init/spec.md
// that follows after, without its fences and its indent of two spaces.
func specBlock(t *testing.T, after string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "specs", "init", "spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(b), after)
	if !ok {
		t.Fatalf("the spec has no %q", after)
	}
	_, rest, _ = strings.Cut(rest, "  ```\n")
	block, _, ok := strings.Cut(rest, "  ```\n")
	if !ok {
		t.Fatalf("no code block follows %q", after)
	}
	var text strings.Builder
	for _, line := range strings.SplitAfter(block, "\n") {
		text.WriteString(strings.TrimPrefix(line, "  "))
	}
	return text.String()
}

func TestInitLayoutRulesAndPatterns(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.9.0", "-l", "asdlc")
	for _, f := range []string{"breadcrumb.rules", ".breadcrumbs"} {
		got := read(t, dir, f)
		if first, _, _ := strings.Cut(got, "\n"); first != "# bcr init --layout asdlc, bcr 0.9.0" {
			t.Errorf("the first line of %s is %q", f, first)
		}
		want := strings.Replace(specBlock(t, "**`"+f+"`.** Written as:"), "VERSION", "0.9.0", 1)
		if got != want {
			t.Errorf("%s is\n%s\nwant, as the spec shows it,\n%s", f, got, want)
		}
	}
}

func TestInitTemplatesAreNotBreadcrumbs(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.9.0", "--layout", "asdlc")
	chdir(t, dir)
	if stdout, stderr, code := command("", "extract"); stdout != "" || stderr != "" || code != 0 {
		t.Errorf("bcr extract prints %q and %q, exits %d; want nothing and 0", stdout, stderr, code)
	}
	for i, f := range asdlcTemplates {
		stdout, stderr, code := command("", "extract", f)
		if code != 0 || stderr != "" {
			t.Errorf("bcr extract %s: exit %d, standard error %q", f, code, stderr)
			continue
		}
		lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
		fields := strings.Split(lines[0], "\t")
		if len(lines) != 1 || len(fields) < 3 || fields[0] != "breadcrumb" || fields[2] != []string{"ADR", "spec", "PBI"}[i] {
			t.Errorf("bcr extract %s prints %q; want one breadcrumb record of its type", f, stdout)
		}
	}
}

// fromTemplate returns the template at path under dir with each old
// replaced by its new, and fails when the template lacks an old.
func fromTemplate(t *testing.T, dir, path string, oldNew ...string) string {
	t.Helper()
	text := read(t, dir, path)
	for i := 0; i < len(oldNew); i += 2 {
		if !strings.Contains(text, oldNew[i]) {
			t.Fatalf("%s has no %q", path, oldNew[i])
		}
		text = strings.Replace(text, oldNew[i], oldNew[i+1], 1)
	}
	return text
}

func TestInitDocumentsFromTheTemplates(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.9.0", "--layout", "asdlc")
	write(t, dir, "docs/adrs/ADR-001-first.md", fromTemplate(t, dir, "docs/adrs/TEMPLATE.md",
		"id: ADR-NNN", "id: ADR-001"))
	write(t, dir, "specs/first/spec.md", fromTemplate(t, dir, "specs/TEMPLATE.md",
		"id: feature-name", "id: first",
		"  links: []\n  # links:\n  #   - constrained_by ADR-NNN\n", "  links:\n    - constrained_by ADR-001\n"))
	write(t, dir, "tasks/PBI-001.md", fromTemplate(t, dir, "tasks/TEMPLATE.md",
		"id: PBI-NNN", "id: PBI-001",
		"  links: []\n  # links:\n  #   - changes feature-name\n", "  links:\n    - changes first\n"))
	chdir(t, dir)

	input := ""
	for _, stage := range []string{"extract", "verify", "audit"} {
		stdout, stderr, code := command(input, stage)
		if code != 0 || stderr != "" {
			t.Fatalf("bcr %s: exit %d, standard error %q; want 0 and nothing", stage, code, stderr)
		}
		input = stdout
	}
	for _, want := range []string{"\tADR-001\t", "\tfirst\t", "\tPBI-001\t"} {
		if !strings.Contains(input, want) {
			t.Errorf("the pipe has no record of%s", strings.TrimSuffix(want, "\t"))
		}
	}
}

func TestInitAgentsFileWithTheLayout(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.9.0", "--layout", "asdlc")
	if got, want := read(t, dir, "AGENTS.md"), agentsSection+"\n"+asdlcSection; got != want {
		t.Errorf("AGENTS.md is %q, want the Breadcrumbs section, then the ASDLC section", got)
	}
	if !strings.HasPrefix(asdlcSection, "## ASDLC\n") {
		t.Error("the ASDLC section does not start with ## ASDLC")
	}
	for _, name := range append([]string{"breadcrumb.rules"}, asdlcTemplates...) {
		if !strings.Contains(asdlcSection, "`"+name+"`") {
			t.Errorf("the ASDLC section does not name %s", name)
		}
	}
}

func TestInitLayoutOverTheBreadcrumbsSection(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.8.0")
	if err := os.Remove(filepath.Join(dir, ".breadcrumbs")); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := initRun(dir, "0.9.0", "--layout", "asdlc")
	if code != 0 {
		t.Fatalf("exit %d, standard error %q; want 0", code, stderr)
	}
	if got, want := read(t, dir, "AGENTS.md"), agentsSection+"\n"+asdlcSection; got != want {
		t.Errorf("AGENTS.md is %q, want the ASDLC section after the Breadcrumbs section", got)
	}
	if !strings.Contains(stderr, "AGENTS.md already has a ## Breadcrumbs section; left as it is") {
		t.Errorf("standard error %q does not say the Breadcrumbs section was left", stderr)
	}
	if !strings.Contains(stdout, "AGENTS.md\n") {
		t.Errorf("standard output %q does not name AGENTS.md", stdout)
	}
}

func TestInitLayoutRefusesAnyShapeFile(t *testing.T) {
	dir := t.TempDir()
	initRun(dir, "0.8.0")
	before := contents(t, dir)
	stdout, stderr, code := initRun(dir, "0.9.0", "--layout", "asdlc")
	if code != 2 || stdout != "" {
		t.Errorf("exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.Contains(stderr, ".breadcrumbs is already there") || !strings.Contains(stderr, "delete both") || !strings.Contains(stderr, "breadcrumb.rules") {
		t.Errorf("standard error %q does not name .breadcrumbs and say to delete both", stderr)
	}
	unchanged(t, dir, before)

	dir = t.TempDir()
	write(t, dir, "breadcrumb.rules", "type\tPBI\n")
	before = contents(t, dir)
	stdout, stderr, code = initRun(dir, "0.9.0", "--layout", "asdlc")
	if code != 2 || stdout != "" {
		t.Errorf("with a breadcrumb.rules: exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.Contains(stderr, "breadcrumb.rules is already there") {
		t.Errorf("standard error %q does not name breadcrumb.rules", stderr)
	}
	unchanged(t, dir, before)
}

func TestInitLeavesATemplate(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "specs/TEMPLATE.md", "Ours.\n")
	stdout, stderr, code := initRun(dir, "0.9.0", "--layout", "asdlc")
	if code != 0 {
		t.Fatalf("exit %d, want 0", code)
	}
	if read(t, dir, "specs/TEMPLATE.md") != "Ours.\n" {
		t.Error("specs/TEMPLATE.md changed")
	}
	if !strings.Contains(stderr, "specs/TEMPLATE.md is already there; left as it is") {
		t.Errorf("standard error %q does not say specs/TEMPLATE.md was left", stderr)
	}
	want := ".breadcrumbs\nbreadcrumb.rules\ndocs/adrs/TEMPLATE.md\ntasks/TEMPLATE.md\nAGENTS.md\n.github/workflows/breadcrumbs.yml\n"
	if stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
}

func TestInitUnknownLayout(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := initRun(dir, "0.9.0", "--layout", "spec-kit")
	if code != 2 || stdout != "" {
		t.Errorf("exit %d, standard output %q; want 2 and nothing", code, stdout)
	}
	if !strings.Contains(stderr, "asdlc") || !strings.HasSuffix(stderr, "\n"+initUsage) {
		t.Errorf("standard error %q does not name asdlc, then the usage line", stderr)
	}
	if f := filesIn(t, dir); len(f) > 0 {
		t.Errorf("wrote %q", f)
	}
}

func TestModuleVersion(t *testing.T) {
	for v, want := range map[string]string{
		"v0.8.0":                             "0.8.0",
		"0.8.0":                              "0.8.0",
		"(devel)":                            "",
		"":                                   "",
		"v0.8.1-0.20261007120000-abcdef0123": "",
	} {
		if got := moduleVersion(v); got != want {
			t.Errorf("moduleVersion(%q) = %q, want %q", v, got, want)
		}
	}
}
