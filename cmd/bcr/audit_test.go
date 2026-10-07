package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// The tests below follow the Scenarios of specs/audit/spec.md, one
// test each, in the same order.

// auditRun runs bcr audit with input on standard input, and returns
// what it writes to standard output and standard error, and its exit
// status.
func auditRun(input string, operands ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"audit"}, operands...), strings.NewReader(input), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// The records of a spec with one claim, about docs/bcr.md.
const (
	verifyCrumb = "breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3"
	verifyClaim = "claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8"
)

// crumbAt returns the record of a breadcrumb written in path.
func crumbAt(id, path string) string {
	return "breadcrumb\t" + id + "\tspec\t" + path + "\t3"
}

// claimAt returns the record of a has-line claim written at line of
// path, about target.
func claimAt(id, target, text, path, line string) string {
	return "claim\t" + id + "\thas-line\t" + target + "\t" + text + "\t" + path + "\t" + line
}

// added runs bcr audit on input, fails unless standard output starts
// with input, and returns the records added after it, one each, with
// standard error and the exit status.
func added(t *testing.T, input string) ([]string, string, int) {
	t.Helper()
	stdout, stderr, code := auditRun(input)
	rest, ok := strings.CutPrefix(stdout, input)
	if !ok {
		t.Fatalf("standard output %q does not start with the input %q", stdout, input)
	}
	if rest == "" {
		return nil, stderr, code
	}
	return strings.Split(strings.TrimSuffix(rest, "\n"), "\n"), stderr, code
}

// verdictOf audits one has-line claim with text, about docs/bcr.md,
// and returns the verdict of the claim.
func verdictOf(t *testing.T, text string) string {
	t.Helper()
	recs, _, _ := added(t, recordsOf(verifyCrumb, claimAt("verify", "docs/bcr.md", text, "specs/verify/spec.md", "8")))
	if len(recs) != 2 {
		t.Fatalf("added %q, want a claim-verdict and a verdict", recs)
	}
	return strings.Split(recs[0], "\t")[2]
}

// wantRecords fails unless got holds the records of want, in order.
func wantRecords(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("added\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestAuditAClaimThatHolds(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "# bcr\n\n### verify\n"})
	recs, stderr, code := added(t, recordsOf(verifyCrumb, verifyClaim))
	wantRecords(t, recs,
		"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8",
		"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3")
	if stderr != "" || code != 0 {
		t.Errorf("got %q, exit %d; want nothing, exit 0", stderr, code)
	}
}

func TestAuditAClaimThatDoesNotHold(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "# bcr\n\n#### verify\n"})
	recs, stderr, code := added(t, recordsOf(verifyCrumb, verifyClaim))
	wantRecords(t, recs,
		"claim-verdict\tverify\tRefuted\tspecs/verify/spec.md\t8",
		"verdict\tverify\tRefuted\tspecs/verify/spec.md\t3")
	if !strings.HasPrefix(stderr, "specs/verify/spec.md:8: ") || !strings.Contains(stderr, "docs/bcr.md") || strings.Count(stderr, "\n") != 1 {
		t.Errorf("standard error %q, want one message at specs/verify/spec.md:8 that names docs/bcr.md", stderr)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestAuditNoInput(t *testing.T) {
	if stdout, stderr, code := auditRun(""); stdout != "" || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want nothing, exit 0", stdout, stderr, code)
	}
}

func TestAuditALineWithSpacesAroundIt(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "func f() error {\r\n\t  return nil  \r\n}\r\n"})
	if got := verdictOf(t, "return nil"); got != "Confirmed" {
		t.Errorf("got %s, want Confirmed", got)
	}
}

func TestAuditALineThatOnlyHoldsTheText(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "Status: Accepted, amended by ADR-0016\n"})
	if got := verdictOf(t, "Status: Accepted"); got != "Refuted" {
		t.Errorf("got %s, want Refuted", got)
	}
}

func TestAuditLetterCase(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### Verify\n"})
	if got := verdictOf(t, "### verify"); got != "Refuted" {
		t.Errorf("got %s, want Refuted", got)
	}
}

func TestAuditALastLineWithNoLineEnding(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "# bcr\n### verify"})
	if got := verdictOf(t, "### verify"); got != "Confirmed" {
		t.Errorf("got %s, want Confirmed", got)
	}
}

func TestAuditATargetThatDoesNotExist(t *testing.T) {
	files(t, map[string]string{"docs/other.md": "### verify\n"})
	recs, stderr, code := added(t, recordsOf(verifyCrumb, verifyClaim))
	if len(recs) != 2 || !strings.HasPrefix(recs[0], "claim-verdict\tverify\tRefuted\t") {
		t.Errorf("added %q, want the claim Refuted", recs)
	}
	if !strings.Contains(stderr, "docs/bcr.md is not a file") || code != 1 {
		t.Errorf("got %q, exit %d; want a message, exit 1", stderr, code)
	}
}

func TestAuditATargetThatIsNotARegularFile(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	if err := os.Symlink("bcr.md", "docs/link.md"); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	recs, _, code := added(t, recordsOf(verifyCrumb,
		claimAt("verify", "docs", "### verify", "specs/verify/spec.md", "8"),
		claimAt("verify", "docs/link.md", "### verify", "specs/verify/spec.md", "9")))
	wantRecords(t, recs,
		"claim-verdict\tverify\tRefuted\tspecs/verify/spec.md\t8",
		"claim-verdict\tverify\tRefuted\tspecs/verify/spec.md\t9",
		"verdict\tverify\tRefuted\tspecs/verify/spec.md\t3")
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestAuditATargetThatCannotBeRead(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	if err := os.Chmod("docs/bcr.md", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile("docs/bcr.md"); err == nil {
		t.Skip("this user can read a file with no permissions")
	}
	stdout, stderr, code := auditRun(recordsOf(verifyCrumb, verifyClaim))
	if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || code != 2 {
		t.Errorf("got %q, %q, exit %d; want a message, exit 2", stdout, stderr, code)
	}
}

func TestAuditSeveralClaimsOfOneBreadcrumb(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### extract\n### verify\n"})
	recs, _, code := added(t, recordsOf(crumbAt("A", "a.md"),
		claimAt("A", "docs/bcr.md", "### extract", "a.md", "8"),
		claimAt("A", "docs/bcr.md", "### audit", "a.md", "9"),
		claimAt("A", "docs/bcr.md", "### verify", "a.md", "10")))
	wantRecords(t, recs,
		"claim-verdict\tA\tConfirmed\ta.md\t8",
		"claim-verdict\tA\tRefuted\ta.md\t9",
		"claim-verdict\tA\tConfirmed\ta.md\t10",
		"verdict\tA\tRefuted\ta.md\t3")
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestAuditABreadcrumbWithNoClaims(t *testing.T) {
	files(t, nil)
	recs, stderr, code := added(t, recordsOf("breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3"))
	wantRecords(t, recs, "verdict\tADR-0001\tUndecided\tdocs/adrs/a.md\t3\tno claims")
	if stderr != "" || code != 0 {
		t.Errorf("got %q, exit %d; want nothing, exit 0", stderr, code)
	}
}

func TestAuditABreadcrumbWithADroppedClaim(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, stderr, code := added(t, recordsOf(crumbAt("A", "a.md"),
		claimAt("A", "docs/bcr.md", "### verify", "a.md", "8"),
		"problem\ta.md\t9\tclaim must be written in single quotes"))
	wantRecords(t, recs,
		"claim-verdict\tA\tConfirmed\ta.md\t8",
		"verdict\tA\tUndecided\ta.md\t3")
	if stderr != "" || code != 0 {
		t.Errorf("got %q, exit %d; want nothing, exit 0", stderr, code)
	}
}

func TestAuditARefutedClaimBesideAProblem(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### extract\n"})
	recs, _, code := added(t, recordsOf(crumbAt("A", "a.md"),
		claimAt("A", "docs/bcr.md", "### verify", "a.md", "8"),
		"problem\ta.md\t9\tclaim must be written in single quotes"))
	if len(recs) != 2 || recs[1] != "verdict\tA\tRefuted\ta.md\t3" || code != 1 {
		t.Errorf("added %q, exit %d; want the breadcrumb Refuted, exit 1", recs, code)
	}
}

func TestAuditABreadcrumbWhoseOnlyTroubleIsAProblem(t *testing.T) {
	files(t, nil)
	recs, _, code := added(t, recordsOf(crumbAt("A", "a.md"), "problem\ta.md\t3\tid \"A\" is also at b.md:3"))
	wantRecords(t, recs, "verdict\tA\tUndecided\ta.md\t3")
	if code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
}

func TestAuditAProblemWithNoBreadcrumb(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, code := added(t, recordsOf(
		"problem\ta.md\t2\tbreadcrumb has no type",
		crumbAt("B", "b.md"),
		claimAt("B", "docs/bcr.md", "### verify", "b.md", "8")))
	wantRecords(t, recs,
		"claim-verdict\tB\tConfirmed\tb.md\t8",
		"verdict\tB\tConfirmed\tb.md\t3")
	if code != 0 {
		t.Errorf("exit %d, want 0", code)
	}
}

func TestAuditTwoBreadcrumbsWithTheSameID(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, _ := added(t, recordsOf(
		"breadcrumb\tADR-0003\tADR\ta.md\t3",
		"breadcrumb\tADR-0003\tADR\tb.md\t3",
		claimAt("ADR-0003", "docs/bcr.md", "### verify", "a.md", "8")))
	wantRecords(t, recs,
		"claim-verdict\tADR-0003\tConfirmed\ta.md\t8",
		"verdict\tADR-0003\tConfirmed\ta.md\t3",
		"verdict\tADR-0003\tUndecided\tb.md\t3\tno claims")
}

func TestAuditTwoClaimsAboutOneTarget(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, code := added(t, recordsOf(
		crumbAt("A", "a.md"), claimAt("A", "docs/bcr.md", "### verify", "a.md", "8"),
		crumbAt("B", "b.md"), claimAt("B", "docs/bcr.md", "### audit", "b.md", "8")))
	wantRecords(t, recs,
		"claim-verdict\tA\tConfirmed\ta.md\t8",
		"claim-verdict\tB\tRefuted\tb.md\t8",
		"verdict\tA\tConfirmed\ta.md\t3",
		"verdict\tB\tRefuted\tb.md\t3")
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestAuditTheOrderOfTheOutput(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, _ := added(t, recordsOf(
		crumbAt("B", "b.md"),
		crumbAt("A", "a.md"),
		claimAt("A", "docs/bcr.md", "### verify", "a.md", "9"),
		claimAt("B", "docs/bcr.md", "### verify", "b.md", "8"),
		claimAt("A", "docs/bcr.md", "### verify", "a.md", "8")))
	wantRecords(t, recs,
		"claim-verdict\tA\tConfirmed\ta.md\t9",
		"claim-verdict\tB\tConfirmed\tb.md\t8",
		"claim-verdict\tA\tConfirmed\ta.md\t8",
		"verdict\tB\tConfirmed\tb.md\t3",
		"verdict\tA\tConfirmed\ta.md\t3")
}

func TestAuditRecordsOfAKindItDoesNotRead(t *testing.T) {
	files(t, nil)
	recs, stderr, code := added(t, recordsOf(
		"note\tanything at all",
		"link\tA\tfollows",
		crumbAt("A", "a.md"),
		"link\tA\tfollows\tADR-0001\ta.md\t6"))
	wantRecords(t, recs, "verdict\tA\tUndecided\ta.md\t3\tno claims")
	if stderr != "" || code != 0 {
		t.Errorf("got %q, exit %d; want nothing, exit 0", stderr, code)
	}
}

func TestAuditFieldsAddedAtTheEnd(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, _ := added(t, recordsOf(verifyCrumb+"\tmore", verifyClaim+"\tmore"))
	wantRecords(t, recs,
		"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8",
		"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3")
}

func TestAuditWindowsLineEndingsInTheInput(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	recs, _, _ := added(t, verifyCrumb+"\r\n"+verifyClaim+"\r\n")
	wantRecords(t, recs,
		"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8",
		"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3")
}

func TestAuditALastLineOfInputWithNoLineEnding(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	input := verifyCrumb + "\n" + verifyClaim
	stdout, _, _ := auditRun(input)
	want := input + "\n" +
		"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8\n" +
		"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3\n"
	if stdout != want {
		t.Errorf("got %q, want %q", stdout, want)
	}
}

func TestAuditALineThatIsNotARecord(t *testing.T) {
	files(t, map[string]string{"docs/bcr.md": "### verify\n"})
	stdout, stderr, code := auditRun(recordsOf(verifyCrumb, "claim\tverify\thas-line\tdocs/bcr.md", verifyClaim))
	if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "line 2 ") || code != 2 {
		t.Errorf("got %q, %q, exit %d; want a message that gives line 2, exit 2", stdout, stderr, code)
	}
}

func TestAuditAClaimRecordExtractCouldNotPrint(t *testing.T) {
	files(t, map[string]string{"x.md": "x\n", "docs/bcr.md": "x\n"})
	for _, claim := range []string{
		"claim\tA\thas-line\t/etc/passwd\tx\ta.md\t8",
		"claim\tA\thas-line\t../x.md\tx\ta.md\t8",
		"claim\tA\tcontains\tdocs/bcr.md\tx\ta.md\t8",
		"claim\tA\thas-line\tdocs/bcr.md\tx \ta.md\t8",
	} {
		stdout, stderr, code := auditRun(recordsOf(crumbAt("A", "a.md"), claim))
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "line 2 ") || code != 2 {
			t.Errorf("%q: got %q, %q, exit %d; want a message that gives line 2, exit 2", claim, stdout, stderr, code)
		}
	}
}

func TestAuditAnOperand(t *testing.T) {
	stdout, stderr, code := auditRun("", "records.tsv")
	if stdout != "" || !strings.HasSuffix(stderr, auditUsage) || code != 2 {
		t.Errorf("got %q, %q, exit %d; want the usage line, exit 2", stdout, stderr, code)
	}
}

func TestAuditOutputThatCannotBeWritten(t *testing.T) {
	files(t, nil)
	var stderr bytes.Buffer
	code := run([]string{"audit"}, strings.NewReader(recordsOf(crumbAt("A", "a.md"))), io.Writer(unwritable{}), &stderr)
	if stderr.String() != "bcr: output is closed\n" || code != 2 {
		t.Errorf("got %q, exit %d; want one message, exit 2", stderr.String(), code)
	}
}

func TestAuditTheWholePipe(t *testing.T) {
	files(t, map[string]string{
		"docs/bcr.md": "### verify\n",
		"a.md": fm("breadcrumb:", "  id: A", "  type: spec", "  links: []", "  claims:",
			"    - 'docs/bcr.md has-line ### verify'", `    - "docs/bcr.md has-line ### extract"`),
	})
	extracted, stderr, code := extractAll("a.md")
	if !strings.HasPrefix(stderr, "a.md:8: ") || code != 1 {
		t.Fatalf("bcr extract wrote %q, exit %d; want the problem of the second claim, exit 1", stderr, code)
	}
	verified, stderr, code := verifyRun(extracted)
	if verified != extracted || stderr != "" || code != 0 {
		t.Fatalf("bcr verify wrote %q, %q, exit %d; want its input, exit 0", verified, stderr, code)
	}
	recs, stderr, code := added(t, verified)
	wantRecords(t, recs,
		"claim-verdict\tA\tConfirmed\ta.md\t7",
		"verdict\tA\tUndecided\ta.md\t3")
	if stderr != "" || code != 0 {
		t.Errorf("bcr audit wrote %q, exit %d; want nothing, exit 0", stderr, code)
	}
}
