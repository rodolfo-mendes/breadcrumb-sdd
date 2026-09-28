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
	if code := run([]string{"audit-report", "--html"}, &stdout, &stderr); code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestAnythingElseIsAUsageError(t *testing.T) {
	for _, args := range [][]string{nil, {"audit-report"}, {"audit-report", "--pdf"}, {"help"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
	}
}
