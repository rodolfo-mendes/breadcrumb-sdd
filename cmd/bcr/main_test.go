package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestAuditReportWritesTheReportInTheRepositoryRoot(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "breadcrumbs"), 0o755)
	os.WriteFile(filepath.Join(root, "breadcrumbs", "TD-0001.md"), []byte("# TD-0001: A decision\n"), 0o644)
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	out := filepath.Join(root, reportFile)
	if strings.TrimSpace(stdout.String()) != out {
		t.Errorf("printed %q, want %q", stdout.String(), out)
	}
	html, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "TD-0001") {
		t.Error("the report does not show TD-0001")
	}
}

func TestAuditReportNeedsABreadcrumbsDirectory(t *testing.T) {
	chdir(t, t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if stdout.Len() > 0 {
		t.Errorf("wrote %q to standard output", stdout.String())
	}
	if !strings.HasPrefix(stderr.String(), "bcr: no breadcrumbs/ directory here") {
		t.Errorf("wrote %q to standard error", stderr.String())
	}
}

func TestAuditReportExitsOneWhenABreadcrumbIsRefuted(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "breadcrumbs"), 0o755)
	os.WriteFile(filepath.Join(root, "breadcrumbs", "TK-0001.md"),
		[]byte("# TK-0001: A task\n\n## Claims\n\n- `go.mod` contains `module`\n"), 0o644)
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html"}, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1: %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, reportFile)); err != nil {
		t.Errorf("the report was not written: %v", err)
	}
}

func TestAnythingElseIsAUsageError(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"help"},
		{"audit-report"},
		{"audit-report", "--pdf"},
		{"audit-report", "--html=yes"},
		{"audit-report", "-x"},
		{"audit-report", "--html", "extra"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
		if stdout.Len() > 0 {
			t.Errorf("%q: wrote %q to standard output", args, stdout.String())
		}
		if !strings.HasPrefix(stderr.String(), "bcr: ") || !strings.HasSuffix(stderr.String(), usage) {
			t.Errorf("%q: wrote %q to standard error", args, stderr.String())
		}
	}
}

// repo makes a repository with one breadcrumb and runs from its root.
func repo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "breadcrumbs"), 0o755)
	os.WriteFile(filepath.Join(root, "breadcrumbs", "TD-0001.md"), []byte("# TD-0001: A decision\n"), 0o644)
	chdir(t, root)
	return root
}

func TestAuditReportWritesTheReportToTheFileOfOutput(t *testing.T) {
	for _, args := range [][]string{
		{"audit-report", "--html", "-o", "out/r.html"},
		{"audit-report", "--html", "-oout/r.html"},
		{"audit-report", "--html", "--output=out/r.html"},
		{"audit-report", "--html", "--output", "out/r.html"},
		{"audit-report", "-o", "out/r.html", "--html"},
	} {
		root := repo(t)
		os.Mkdir(filepath.Join(root, "out"), 0o755)

		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("%q: exit %d: %s", args, code, stderr.String())
		}
		out := filepath.Join(root, "out", "r.html")
		if strings.TrimSpace(stdout.String()) != out {
			t.Errorf("%q: printed %q, want %q", args, stdout.String(), out)
		}
		if html, err := os.ReadFile(out); err != nil || !strings.Contains(string(html), "TD-0001") {
			t.Errorf("%q: the report was not written to %s: %v", args, out, err)
		}
		if _, err := os.Stat(filepath.Join(root, reportFile)); err == nil {
			t.Errorf("%q: %s was written too", args, reportFile)
		}
	}
}

func TestAuditReportWritesTheReportToStandardOutputForADash(t *testing.T) {
	root := repo(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html", "-o", "-"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "<!doctype html>") || !strings.Contains(stdout.String(), "TD-0001") {
		t.Errorf("standard output is not the report: %.60q", stdout.String())
	}
	if stderr.Len() > 0 {
		t.Errorf("wrote %q to standard error", stderr.String())
	}
	if entries, _ := os.ReadDir(root); len(entries) != 1 {
		t.Errorf("wrote a file: %v", entries)
	}
}

func TestAuditReportCannotWriteToAMissingDirectory(t *testing.T) {
	repo(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html", "-o", "missing/r.html"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if stdout.Len() > 0 || !strings.HasPrefix(stderr.String(), "bcr: ") {
		t.Errorf("wrote %q to standard output and %q to standard error", stdout.String(), stderr.String())
	}
}

func TestAuditReportNeedsAFileNameForOutput(t *testing.T) {
	repo(t)
	for _, args := range [][]string{
		{"audit-report", "--html", "-o"},
		{"audit-report", "--html", "--output="},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
	}
}
