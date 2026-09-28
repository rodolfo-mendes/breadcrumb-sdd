package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildWritesThePagesOfTheTags builds the site of a repository with
// a tag before the specification and a tag after it.
func TestBuildWritesThePagesOfTheTags(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	root := t.TempDir()
	sh := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "tag.gpgSign=false", "-c", "commit.gpgSign=false"}, args...)...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	write := func(p, s string) {
		os.MkdirAll(filepath.Join(root, filepath.Dir(p)), 0o755)
		os.WriteFile(filepath.Join(root, p), []byte(s), 0o644)
	}
	sh("init", "-q")
	write("go.mod", "module github.com/example/repo\n\ngo 1.22\n")
	sh("add", ".")
	sh("commit", "-qm", "first")
	sh("tag", "v0.1.0")
	write("docs/breadcrumb-sdd.md", "# Spec\n\nVersion 0.2.0\n\nSee [TD-0001](../breadcrumbs/TD-0001.md).\n")
	sh("add", ".")
	sh("commit", "-qm", "spec")
	sh("tag", "v0.2.0")

	old, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(old)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"-o", "out"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	page, err := os.ReadFile(filepath.Join(root, "out", "spec", "0.2.0", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), `https://github.com/example/repo/blob/v0.2.0/breadcrumbs/TD-0001.md`) {
		t.Error("the page's link does not lead to the file at its tag")
	}
	for _, p := range []string{"index.html", "spec/latest/index.html"} {
		if _, err := os.Stat(filepath.Join(root, "out", p)); err != nil {
			t.Error(err)
		}
	}
	if n := strings.Count(stdout.String(), "\n"); n != 3 {
		t.Errorf("printed %d paths, want 3: %s", n, stdout.String())
	}
}

func TestRunRefusesOperands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"extra"}, &stdout, &stderr); code != 2 || !strings.HasSuffix(stderr.String(), usage) {
		t.Errorf("exit %d: %s", code, stderr.String())
	}
}
