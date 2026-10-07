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
	if read(t, dir, ".breadcrumbs") != patterns {
		t.Error(".breadcrumbs is not the patterns of bcr init")
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
	for i, line := range strings.Split(strings.TrimSuffix(read(t, dir, ".breadcrumbs"), "\n"), "\n") {
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
