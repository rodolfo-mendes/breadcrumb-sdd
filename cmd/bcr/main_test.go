package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/docs"
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
		{[]string{"list", "-r"}, listUsage},
		{[]string{"links", "-x"}, linksUsage},
		{[]string{"links", "--recursive=yes"}, linksUsage},
		{[]string{"new", "TK"}, newUsage},
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

// chain makes a repository where TK-0001 carries out RQ-0001, which
// derives from IN-0001 and TD-0002; TD-0002 derives from IN-0001 too,
// and TK-0001 also links to a missing TD-0099.
func chain(t *testing.T) {
	t.Helper()
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/IN-0001.md": "# IN-0001: An ask\n",
		"breadcrumbs/TD-0002.md": "# TD-0002: A decision\n\nParent: [IN-0001](IN-0001.md)\n",
		"breadcrumbs/RQ-0001.md": "# RQ-0001: A requirement\n\nParent: [IN-0001](IN-0001.md)\nParent: [TD-0002](TD-0002.md)\n",
		"breadcrumbs/TK-0001.md": "# TK-0001: A task\n\nParent: [RQ-0001](RQ-0001.md)\nParent: [TD-0099](TD-0099.md)\n",
	})
}

func TestListPrintsEachBreadcrumbWithItsTypeAndTitle(t *testing.T) {
	chain(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"list"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0: %s", code, stderr.String())
	}
	want := "IN-0001\tIntake\tAn ask\n" +
		"RQ-0001\tRequirement\tA requirement\n" +
		"TD-0001\tTechnical Decision\tA decision\n" +
		"TD-0002\tTechnical Decision\tA decision\n" +
		"TK-0001\tTask\tA task\n"
	if stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func TestListPrintsTheBreadcrumbsItIsGiven(t *testing.T) {
	chain(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"list", "TK-0001", "-"}, strings.NewReader("IN-0001\tUndecided\nTK-0042\n"), &stdout, &stderr)
	if code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if want := "TK-0001\tTask\tA task\nIN-0001\tIntake\tAn ask\n"; stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
	if want := "bcr: TK-0042 names no breadcrumb\n"; stderr.String() != want {
		t.Errorf("wrote %q to standard error, want %q", stderr.String(), want)
	}
}

func TestLinksPrintsEachLinkToAParent(t *testing.T) {
	chain(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"links"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0: %s", code, stderr.String())
	}
	want := "RQ-0001\tIN-0001\n" +
		"RQ-0001\tTD-0002\n" +
		"TD-0002\tIN-0001\n" +
		"TK-0001\tRQ-0001\n" +
		"TK-0001\tbreadcrumbs/TD-0099.md\n"
	if stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func TestLinksOfTheBreadcrumbsItIsGiven(t *testing.T) {
	chain(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"links", "breadcrumbs/TD-0002.md", "TK-0001"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0: %s", code, stderr.String())
	}
	if want := "TD-0002\tIN-0001\nTK-0001\tRQ-0001\nTK-0001\tbreadcrumbs/TD-0099.md\n"; stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func TestLinksRecursiveFollowsTheChainUpOnce(t *testing.T) {
	chain(t)
	for _, flag := range []string{"-r", "--recursive"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{"links", flag, "TK-0001"}, nil, &stdout, &stderr); code != 0 {
			t.Errorf("%s: exit %d, want 0: %s", flag, code, stderr.String())
		}
		want := "TK-0001\tRQ-0001\n" +
			"TK-0001\tbreadcrumbs/TD-0099.md\n" +
			"RQ-0001\tIN-0001\n" +
			"RQ-0001\tTD-0002\n" +
			"TD-0002\tIN-0001\n"
		if stdout.String() != want {
			t.Errorf("%s: printed %q, want %q", flag, stdout.String(), want)
		}
	}
}

func TestLinksRecursiveEndsOnACycle(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/RQ-0001.md": "# RQ-0001: One\n\nParent: [RQ-0002](RQ-0002.md)\n",
		"breadcrumbs/RQ-0002.md": "# RQ-0002: Two\n\nParent: [RQ-0001](RQ-0001.md)\n",
	})
	var stdout, stderr bytes.Buffer
	if code := run([]string{"links", "-r", "RQ-0001"}, nil, &stdout, &stderr); code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
	if want := "RQ-0001\tRQ-0002\nRQ-0002\tRQ-0001\n"; stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
}

func read(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.FromSlash(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestNewCreatesEachTypeWithItsShape(t *testing.T) {
	chain(t)
	for _, tc := range []struct {
		args       []string
		path, want string
	}{
		{[]string{"new", "IN", "An ask"}, "breadcrumbs/IN-0002.md", "# IN-0002: An ask\n"},
		{[]string{"new", "RQ", "A need", "IN-0001"}, "breadcrumbs/RQ-0002.md",
			"# RQ-0002: A need\n\nParent: [IN-0001](IN-0001.md)\n"},
		{[]string{"new", "TD", "A choice", "breadcrumbs/IN-0001.md", "RQ-0001"}, "breadcrumbs/TD-0003.md",
			"# TD-0003: A choice\n\nParent: [IN-0001](IN-0001.md)\nParent: [RQ-0001](RQ-0001.md)\n\n## Decision\n\n## Why\n\n## What it beat\n"},
		{[]string{"new", "TK", "A change", "TD-0003"}, "breadcrumbs/TK-0002.md",
			"# TK-0002: A change\n\nParent: [TD-0003](TD-0003.md)\n\n## Claims\n"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 0 {
			t.Fatalf("%q: exit %d: %s", tc.args, code, stderr.String())
		}
		if stdout.String() != tc.path+"\n" {
			t.Errorf("%q: printed %q, want %q", tc.args, stdout.String(), tc.path)
		}
		if got := read(t, tc.path); got != tc.want {
			t.Errorf("%q: wrote %q, want %q", tc.args, got, tc.want)
		}
	}

	// RQ-0032: bcr check finds no problem in them.
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 1 {
		t.Fatalf("check: exit %d", code)
	}
	for _, l := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if !strings.HasPrefix(l, "breadcrumbs/TK-0001.md:") { // chain's broken link
			t.Errorf("check found %q", l)
		}
	}
}

func TestNewNumbersAfterTheHighestOfItsType(t *testing.T) {
	root := repo(t)
	write(t, root, map[string]string{
		"breadcrumbs/TK-0003.md":  "# TK-0003: Three\n",
		"breadcrumbs/TK-0007.md":  "# TK-0007: Seven\n",
		"breadcrumbs/RQ-0041.md":  "# RQ-0041: Other type\n",
		"breadcrumbs/TK-0099.txt": "not a breadcrumb\n",
	})
	for _, want := range []string{"breadcrumbs/TK-0008.md", "breadcrumbs/TK-0009.md"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{"new", "TK", "Next"}, nil, &stdout, &stderr); code != 0 {
			t.Fatalf("exit %d: %s", code, stderr.String())
		}
		if strings.TrimSpace(stdout.String()) != want {
			t.Errorf("printed %q, want %q", stdout.String(), want)
		}
	}
	var stdout, stderr bytes.Buffer
	run([]string{"new", "IN", "First"}, nil, &stdout, &stderr)
	if strings.TrimSpace(stdout.String()) != "breadcrumbs/IN-0001.md" {
		t.Errorf("printed %q, want breadcrumbs/IN-0001.md", stdout.String())
	}
}

func TestNewReadsParentsFromStandardInput(t *testing.T) {
	chain(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{"new", "TK", "Carry them out", "-"}, strings.NewReader("RQ-0001\tUndecided\nTD-0002\tUndecided\n"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	want := "# TK-0002: Carry them out\n\nParent: [RQ-0001](RQ-0001.md)\nParent: [TD-0002](TD-0002.md)\n\n## Claims\n"
	if got := read(t, "breadcrumbs/TK-0002.md"); got != want {
		t.Errorf("wrote %q, want %q", got, want)
	}
}

func TestNewWritesNothingItCannotWriteRight(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		files map[string]string
	}{
		{[]string{"new"}, nil},
		{[]string{"new", "TK"}, nil},
		{[]string{"new", "XX", "A title"}, nil},
		{[]string{"new", "tk", "A title"}, nil},
		{[]string{"new", "Task", "A title"}, nil},
		{[]string{"new", "TK", ""}, nil},
		{[]string{"new", "TK", "  "}, nil},
		{[]string{"new", "TK", "Two\nlines"}, nil},
		{[]string{"new", "TK", "A title", "TD-0001", "TD-0099"}, nil},
		{[]string{"new", "TK", "A title"}, map[string]string{"breadcrumbs/TK-9999.md": "# TK-9999: Last\n"}},
		{[]string{"new", "-x", "TK", "A title"}, nil},
	} {
		root := repo(t)
		write(t, root, tc.files)
		before, _ := os.ReadDir(filepath.Join(root, "breadcrumbs"))

		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", tc.args, code)
		}
		if stdout.Len() > 0 || !strings.HasPrefix(stderr.String(), "bcr: ") {
			t.Errorf("%q: wrote %q and %q", tc.args, stdout.String(), stderr.String())
		}
		if after, _ := os.ReadDir(filepath.Join(root, "breadcrumbs")); len(after) != len(before) {
			t.Errorf("%q: wrote a file", tc.args)
		}
	}
}

func TestNewNeverReplacesAFile(t *testing.T) {
	root := repo(t)
	// Not a breadcrumb, so not counted, but in the way of TK-0001.md.
	write(t, root, map[string]string{"breadcrumbs/TK-0001.md/keep": "x"})
	var stdout, stderr bytes.Buffer
	if code := run([]string{"new", "TK", "A title"}, nil, &stdout, &stderr); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if read(t, "breadcrumbs/TK-0001.md/keep") != "x" {
		t.Error("the file in the way was changed")
	}
}

func TestSpecPrintsTheSpecificationOutsideARepository(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("..", "..", "docs", "breadcrumb-sdd.md"))
	if err != nil {
		t.Fatal(err)
	}
	chdir(t, t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := run([]string{"spec"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if !bytes.Equal(stdout.Bytes(), want) || stderr.Len() > 0 {
		t.Errorf("bcr spec does not print docs/breadcrumb-sdd.md alone")
	}
}

func TestSpecTakesNoFlagsOrOperands(t *testing.T) {
	for _, args := range [][]string{{"spec", "-x"}, {"spec", "extra"}} {
		var stdout, stderr bytes.Buffer
		if code := run(args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
		if stdout.Len() > 0 || !strings.HasSuffix(stderr.String(), specUsage) {
			t.Errorf("%q: printed %q, %q", args, stdout.String(), stderr.String())
		}
	}
}

func TestTheSpecificationStatesItsVersionOnItsThirdLine(t *testing.T) {
	lines := strings.Split(docs.Spec, "\n")
	if len(lines) < 3 || !regexp.MustCompile(`^Version [0-9]+\.[0-9]+\.[0-9]+$`).MatchString(lines[2]) {
		t.Errorf("the third line is not Version MAJOR.MINOR.PATCH (RQ-0036)")
	}
}

// withVersion sets the version bcr was released as, for one test.
func withVersion(t *testing.T, v string) {
	t.Helper()
	old := version
	version = v
	t.Cleanup(func() { version = old })
}

func readFile(t *testing.T, root, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestInitSetsUpARepository(t *testing.T) {
	withVersion(t, "0.7.0")
	root := t.TempDir()
	chdir(t, root)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"init"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if want := "breadcrumbs/TD-0001.md\nAGENTS.md\n.github/workflows/breadcrumbs.yml\n"; stdout.String() != want {
		t.Errorf("printed %q, want %q", stdout.String(), want)
	}
	page := siteURL + "spec/" + specVersion() + "/"
	if td := readFile(t, root, "breadcrumbs/TD-0001.md"); !strings.HasPrefix(td, "# TD-0001: Adopt Breadcrumb SDD\n") || !strings.Contains(td, page) {
		t.Errorf("TD-0001 does not adopt the specification at %s (RQ-0042):\n%s", page, td)
	}
	if a := readFile(t, root, "AGENTS.md"); !strings.HasPrefix(a, agentsHead+"\n") || !strings.Contains(a, page) {
		t.Errorf("AGENTS.md does not point to %s (RQ-0043):\n%s", page, a)
	}
	wf := readFile(t, root, ".github/workflows/breadcrumbs.yml")
	for _, want := range []string{"pull_request:", releaseURL + "v0.7.0/bcr_0.7.0_linux_amd64.tar.gz", "sha256sum --check", "run: bcr check", "run: bcr verdict"} {
		if !strings.Contains(wf, want) {
			t.Errorf("the workflow has no %s (RQ-0044, TD-0035)", want)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"check"}, nil, &stdout, &stderr); code != 0 || stdout.Len() > 0 || stderr.Len() > 0 {
		t.Errorf("bcr check of a new repository: exit %d: %s%s", code, stdout.String(), stderr.String())
	}
}

func TestInitTwiceChangesNothing(t *testing.T) {
	withVersion(t, "0.7.0")
	root := t.TempDir()
	chdir(t, root)
	var stdout, stderr bytes.Buffer
	run([]string{"init"}, nil, &stdout, &stderr)
	before := readFile(t, root, "AGENTS.md") + readFile(t, root, "breadcrumbs/TD-0001.md") + readFile(t, root, ".github/workflows/breadcrumbs.yml")

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"init"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if stdout.Len() > 0 || strings.Count(stderr.String(), "left as it is") != 3 {
		t.Errorf("printed %q, %q", stdout.String(), stderr.String())
	}
	if after := readFile(t, root, "AGENTS.md") + readFile(t, root, "breadcrumbs/TD-0001.md") + readFile(t, root, ".github/workflows/breadcrumbs.yml"); after != before {
		t.Error("a second bcr init changed a file (RQ-0045)")
	}
}

func TestInitAddsItsSectionToTheEndOfTheAgentsFile(t *testing.T) {
	withVersion(t, "0.7.0")
	for _, tc := range []struct {
		args []string
		file string
	}{
		{[]string{"init"}, "AGENTS.md"},
		{[]string{"init", "-a", "CLAUDE.md"}, "CLAUDE.md"},
		{[]string{"init", "--agents-file=CLAUDE.md"}, "CLAUDE.md"},
	} {
		root := t.TempDir()
		chdir(t, root)
		write(t, root, map[string]string{tc.file: "# Our notes\n\nKeep them."})
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 0 {
			t.Fatalf("%q: exit %d: %s", tc.args, code, stderr.String())
		}
		if got := readFile(t, root, tc.file); !strings.HasPrefix(got, "# Our notes\n\nKeep them.\n\n"+agentsHead+"\n") {
			t.Errorf("%q: %s is\n%s", tc.args, tc.file, got)
		}
		other := "CLAUDE.md"
		if tc.file == other {
			other = "AGENTS.md"
		}
		if _, err := os.Stat(filepath.Join(root, other)); err == nil {
			t.Errorf("%q: wrote %s too (RQ-0046)", tc.args, other)
		}
	}
}

func TestInitKeepsWhatTheRepositoryHas(t *testing.T) {
	withVersion(t, "0.7.0")
	root := repo(t)
	files := map[string]string{
		"AGENTS.md":                         "# Agents\n\n## Breadcrumb SDD\n\nOur own words.\n",
		".github/workflows/breadcrumbs.yml": "name: Ours\n",
	}
	write(t, root, files)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init"}, nil, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if stdout.Len() > 0 {
		t.Errorf("printed %q", stdout.String())
	}
	for name, want := range files {
		if got := readFile(t, root, name); got != want {
			t.Errorf("%s changed to %q (RQ-0045)", name, got)
		}
	}
	if got := readFile(t, root, "breadcrumbs/TD-0001.md"); got != "# TD-0001: A decision\n" {
		t.Errorf("TD-0001 changed to %q", got)
	}
}

func TestInitWritesNothingItCannotWriteRight(t *testing.T) {
	for _, tc := range []struct {
		version string
		args    []string
		files   map[string]string
	}{
		{"0.7.0", []string{"init", "-a", "README.md"}, nil},
		{"0.7.0", []string{"init", "extra"}, nil},
		{"0.7.0", []string{"init"}, map[string]string{"breadcrumbs": "a file"}},
		{"", []string{"init"}, nil},
	} {
		withVersion(t, tc.version)
		root := t.TempDir()
		chdir(t, root)
		write(t, root, tc.files)
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", tc.args, code)
		}
		entries, _ := os.ReadDir(root)
		if stdout.Len() > 0 || len(entries) != len(tc.files) {
			t.Errorf("%q with version %q wrote something: %q", tc.args, tc.version, stdout.String())
		}
	}
}

func TestModuleVersion(t *testing.T) {
	for v, want := range map[string]string{"v0.7.0": "0.7.0", "(devel)": "", "v0.0.0-20260928-abcdef": "", "": ""} {
		if got := moduleVersion(v); got != want {
			t.Errorf("moduleVersion(%q) = %q, want %q", v, got, want)
		}
	}
}
