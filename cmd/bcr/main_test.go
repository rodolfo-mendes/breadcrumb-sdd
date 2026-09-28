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
	if code := run([]string{"audit-report", "--html"}, nil, &stdout, &stderr); code != 0 {
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
	if code := run([]string{"audit-report", "--html"}, nil, &stdout, &stderr); code != 2 {
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
	if code := run([]string{"audit-report", "--html"}, nil, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1: %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, reportFile)); err != nil {
		t.Errorf("the report was not written: %v", err)
	}
}

func TestAnythingElseIsAUsageError(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		usage string
	}{
		{nil, usage},
		{[]string{"help"}, usage},
		{[]string{"audit-report"}, auditReportUsage},
		{[]string{"audit-report", "--pdf"}, auditReportUsage},
		{[]string{"audit-report", "--html=yes"}, auditReportUsage},
		{[]string{"audit-report", "-x"}, auditReportUsage},
		{[]string{"audit-report", "--html", "extra"}, auditReportUsage},
		{[]string{"check", "-x"}, checkUsage},
		{[]string{"check", "breadcrumbs/TD-0001.md"}, checkUsage},
		{[]string{"verdict", "-x"}, verdictUsage},
		{[]string{"verdict", "--refuted"}, verdictUsage},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", tc.args, code)
		}
		if stdout.Len() > 0 {
			t.Errorf("%q: wrote %q to standard output", tc.args, stdout.String())
		}
		if !strings.HasPrefix(stderr.String(), "bcr: ") || !strings.HasSuffix(stderr.String(), "\n"+tc.usage) {
			t.Errorf("%q: wrote %q to standard error", tc.args, stderr.String())
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
		if code := run(args, nil, &stdout, &stderr); code != 0 {
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
	if code := run([]string{"audit-report", "--html", "-o", "-"}, nil, &stdout, &stderr); code != 0 {
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
	if code := run([]string{"audit-report", "--html", "-o", "missing/r.html"}, nil, &stdout, &stderr); code != 2 {
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
		if code := run(args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
	}
}

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckPrintsEachProblemWithItsFileAndLine(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/RQ-0002.md": "# RQ-0002: A requirement\n\nParent: [TD-0099](TD-0099.md)\nParent: TD-0001\n",
		"breadcrumbs/IN-0001.md": "An ask with no title\n",
		"breadcrumbs/TK-0001.md": "# TK-0001: A task\n\nParent: [RQ-0002](RQ-0002.md)\n\n## Claims\n\n- `go.mod` contains `module`\n- go.mod contains module\n",
	})

	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1: %s", code, stderr.String())
	}
	want := "breadcrumbs/IN-0001.md:1: the first line is not a # title\n" +
		"breadcrumbs/RQ-0002.md:3: Parent: links to no breadcrumb: breadcrumbs/TD-0099.md\n" +
		"breadcrumbs/RQ-0002.md:4: not a Parent: link: Parent: TD-0001\n" +
		"breadcrumbs/TK-0001.md:8: not a claim: - go.mod contains module\n"
	if stdout.String() != want {
		t.Errorf("printed:\n%s\nwant:\n%s", stdout.String(), want)
	}
	if stderr.Len() > 0 {
		t.Errorf("wrote %q to standard error", stderr.String())
	}
}

// A claim that does not hold is a verdict, not a problem (RQ-0020).
func TestCheckFindsNothingWrongInAClaimThatDoesNotHold(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/TK-0001.md": "# TK-0001: A task\n\nParent: [TD-0001](TD-0001.md)\n\n## Claims\n\n- `go.mod` contains `module`\n",
	})
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0: %s", code, stdout.String())
	}
	if stdout.Len() > 0 {
		t.Errorf("printed %q", stdout.String())
	}
}

func TestCheckWarnsOfFilesThatAreNotBreadcrumbs(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/TD-001.md":      "# TD-001: A typo\n",
		"breadcrumbs/old/TD-0002.md": "# TD-0002: Moved\n",
	})
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	if stdout.Len() > 0 {
		t.Errorf("printed %q", stdout.String())
	}
	want := "bcr: warning: breadcrumbs/TD-001.md is not a breadcrumb (TD-0013)\n" +
		"bcr: warning: breadcrumbs/old/TD-0002.md is not a breadcrumb (TD-0013)\n"
	if stderr.String() != want {
		t.Errorf("wrote %q to standard error, want %q", stderr.String(), want)
	}
}

func TestCheckNeedsABreadcrumbsDirectory(t *testing.T) {
	chdir(t, t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
}

// RQ-0022: audit-report counts a link to no breadcrumb as a problem.
func TestAuditReportExitsOneForALinkToNoBreadcrumb(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{"breadcrumbs/RQ-0001.md": "# RQ-0001: A requirement\n\nParent: [IN-0009](IN-0009.md)\n"})
	var stdout, stderr bytes.Buffer
	if code := run([]string{"audit-report", "--html", "-o", "-"}, nil, &stdout, &stderr); code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "breadcrumbs/RQ-0001.md:3</code>: Parent: links to no breadcrumb") {
		t.Error("the report does not list the problem")
	}
}

// verdicts makes a repository whose TK-0001 is Confirmed, TK-0002 is
// Refuted and IN-0001 is Undecided.
func verdicts(t *testing.T) {
	t.Helper()
	root := repo(t)
	write(t, root, map[string]string{
		"go.mod":                 "module example\n",
		"breadcrumbs/IN-0001.md": "# IN-0001: An ask\n",
		"breadcrumbs/TK-0001.md": "# TK-0001: A task\n\nParent: [TD-0001](TD-0001.md)\n\n## Claims\n\n- `go.mod` contains `module`\n",
		"breadcrumbs/TK-0002.md": "# TK-0002: Another\n\n## Claims\n\n- `go.mod` contains `package`\n",
	})
}

func TestVerdictPrintsEveryBreadcrumbInOrderOfPath(t *testing.T) {
	verdicts(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"verdict"}, nil, &stdout, &stderr); code != 1 {
		t.Errorf("exit %d, want 1: %s", code, stderr.String())
	}
	want := "IN-0001\tUndecided\nTD-0001\tConfirmed\nTK-0001\tConfirmed\nTK-0002\tRefuted\n"
	if stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func TestVerdictPrintsTheBreadcrumbsItIsGivenInOrder(t *testing.T) {
	verdicts(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"verdict", "TK-0001", "breadcrumbs/IN-0001.md", "-", "./breadcrumbs/TD-0001.md"},
		strings.NewReader("TK-0001\tConfirmed\n\nTD-0001\n"), &stdout, &stderr)
	if code != 0 {
		t.Errorf("exit %d, want 0: %s", code, stderr.String())
	}
	want := "TK-0001\tConfirmed\nIN-0001\tUndecided\nTK-0001\tConfirmed\nTD-0001\tConfirmed\nTD-0001\tConfirmed\n"
	if stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func TestVerdictExitsOneOnlyForARefutedBreadcrumbItPrints(t *testing.T) {
	verdicts(t)
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"verdict", "TK-0002"}, 1},
		{[]string{"verdict", "TK-0001", "TK-0002"}, 1},
		{[]string{"verdict", "TK-0001", "IN-0001"}, 0},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != tc.code {
			t.Errorf("%q: exit %d, want %d", tc.args, code, tc.code)
		}
	}
}

func TestVerdictReportsANameOfNoBreadcrumb(t *testing.T) {
	verdicts(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"verdict", "TK-0002", "TK-0099", "TK-0001"}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if want := "TK-0002\tRefuted\nTK-0001\tConfirmed\n"; stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
	if want := "bcr: TK-0099 names no breadcrumb\n"; stderr.String() != want {
		t.Errorf("wrote %q to standard error, want %q", stderr.String(), want)
	}
}

func TestVerdictOfAnEmptyStandardInputPrintsNothing(t *testing.T) {
	verdicts(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"verdict", "-"}, strings.NewReader(""), &stdout, &stderr); code != 0 || stdout.Len() > 0 {
		t.Errorf("exit %d, printed %q", code, stdout.String())
	}
}
