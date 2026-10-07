package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// The tests below follow the Scenarios of specs/verify/spec.md, one
// test each, in the same order.

// verifyRun runs bcr verify with input on standard input, and returns
// what it writes to standard output and standard error, and its exit
// status.
func verifyRun(input string, operands ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"verify"}, operands...), strings.NewReader(input), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// recordsOf joins records, each a line of input.
func recordsOf(records ...string) string {
	return strings.Join(records, "\n") + "\n"
}

// A set with no problem.
var validRecords = []string{
	"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
	"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3",
	"link\tPBI-00001\timplements\tADR-0001\ttasks/PBI-00001.md\t6",
}

// passes fails unless bcr verify, given input, copies it to standard
// output unchanged, prints nothing to standard error and exits 0
// (ADR-0016).
func passes(t *testing.T, input string) {
	t.Helper()
	if stdout, stderr, code := verifyRun(input); stdout != input || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want the input, no problem, exit 0", stdout, stderr, code)
	}
}

// verifyProblems runs bcr verify on input, and fails unless it copies
// input to standard output unchanged, then a problem record for each
// line of standard error (ADR-0021), prints a problem at each of
// want, in that order, and exits 1. It returns the problems, one line
// each.
func verifyProblems(t *testing.T, input string, want ...string) []string {
	t.Helper()
	stdout, stderr, code := verifyRun(input)
	added := ""
	for _, l := range strings.SplitAfter(stderr, "\n") {
		if m := problemLine.FindStringSubmatch(strings.TrimSuffix(l, "\n")); m != nil {
			added += "problem\t" + m[1] + "\t" + m[2] + "\t" + m[3] + "\n"
		}
	}
	if stdout != input+added {
		t.Errorf("wrote %q to standard output, want the input and its problems %q", stdout, input+added)
	}
	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	ok := len(lines) == len(want)
	for i := 0; ok && i < len(want); i++ {
		ok = strings.HasPrefix(lines[i], want[i]+": ")
	}
	if !ok {
		t.Fatalf("got problems\n%swant them at %q", stderr, want)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
	return lines
}

// naming fails unless problem names every one of want.
func naming(t *testing.T, problem string, want ...string) {
	t.Helper()
	_, message, _ := strings.Cut(problem, ": ")
	for _, w := range want {
		if !strings.Contains(message, w) {
			t.Errorf("problem %q does not name %q", problem, w)
		}
	}
}

func TestVerifyASetWithNoProblem(t *testing.T) {
	passes(t, recordsOf(validRecords...))
}

func TestVerifyNoInput(t *testing.T) {
	passes(t, "")
}

func TestVerifyTwoBreadcrumbsWithTheSameID(t *testing.T) {
	ps := verifyProblems(t, recordsOf(
		"breadcrumb\tADR-0003\tADR\tdocs/adrs/a.md\t3",
		"breadcrumb\tADR-0003\tADR\tdocs/adrs/b.md\t3",
	), "docs/adrs/a.md:3", "docs/adrs/b.md:3")
	naming(t, ps[0], "docs/adrs/b.md:3")
	naming(t, ps[1], "docs/adrs/a.md:3")
}

func TestVerifyThreeBreadcrumbsWithTheSameID(t *testing.T) {
	ps := verifyProblems(t, recordsOf(
		"breadcrumb\tPBI-00007\tPBI\ttasks/a.md\t3",
		"breadcrumb\tPBI-00007\tPBI\ttasks/b.md\t3",
		"breadcrumb\tPBI-00007\tPBI\ttasks/c.md\t3",
	), "tasks/a.md:3", "tasks/b.md:3", "tasks/c.md:3")
	naming(t, ps[0], "tasks/b.md:3", "tasks/c.md:3")
	naming(t, ps[1], "tasks/a.md:3", "tasks/c.md:3")
	naming(t, ps[2], "tasks/a.md:3", "tasks/b.md:3")
}

func TestVerifyIDsThatDifferOnlyInCase(t *testing.T) {
	ps := verifyProblems(t, recordsOf(
		"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
		"breadcrumb\tadr-0001\tADR\tdocs/adrs/b.md\t3",
	), "docs/adrs/a.md:3", "docs/adrs/b.md:3")
	naming(t, ps[0], "docs/adrs/b.md:3")
	naming(t, ps[1], "docs/adrs/a.md:3")
}

func TestVerifyCaseBeyondASCII(t *testing.T) {
	ps := verifyProblems(t, recordsOf(
		"breadcrumb\tΣΙΓΜΑ\tADR\ta.md\t3",
		"breadcrumb\tσιγμα\tADR\tb.md\t3",
	), "a.md:3", "b.md:3")
	naming(t, ps[0], "b.md:3")
	naming(t, ps[1], "a.md:3")
}

func TestVerifyALinkThatPointsNowhere(t *testing.T) {
	verifyProblems(t, recordsOf(
		"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3",
		"link\tPBI-00001\timplements\tADR-0099\ttasks/PBI-00001.md\t6",
	), "tasks/PBI-00001.md:6")
}

func TestVerifyALinkWhoseObjectDiffersOnlyInCase(t *testing.T) {
	ps := verifyProblems(t, recordsOf(
		"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
		"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3",
		"link\tPBI-00001\timplements\tadr-0001\ttasks/PBI-00001.md\t6",
	), "tasks/PBI-00001.md:6")
	naming(t, ps[0], `"ADR-0001"`)
}

func TestVerifyALinkToADuplicatedID(t *testing.T) {
	verifyProblems(t, recordsOf(
		"breadcrumb\tADR-0003\tADR\tdocs/adrs/a.md\t3",
		"breadcrumb\tADR-0003\tADR\tdocs/adrs/b.md\t3",
		"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3",
		"link\tPBI-00001\timplements\tADR-0003\ttasks/PBI-00001.md\t6",
	), "docs/adrs/a.md:3", "docs/adrs/b.md:3")
}

func TestVerifyRecordsOfAKindItDoesNotKnow(t *testing.T) {
	passes(t, recordsOf(append([]string{"claim\tPBI-00001\tcontains\tdocs/bcr.md"}, validRecords...)...))
}

func TestVerifyFieldsAddedAtTheEnd(t *testing.T) {
	passes(t, strings.ReplaceAll(recordsOf(validRecords...), "\n", "\tAccepted\t2026\n"))
}

func TestVerifyWindowsLineEndings(t *testing.T) {
	passes(t, strings.ReplaceAll(recordsOf(validRecords...), "\n", "\r\n"))
}

func TestVerifyALastLineWithNoLineEnding(t *testing.T) {
	passes(t, strings.TrimSuffix(recordsOf(validRecords...), "\n"))
}

func TestVerifyALineThatIsNotARecord(t *testing.T) {
	for _, line := range []string{"breadcrumb\tADR-0001\tADR", "", "breadcrumb\tADR-0009\tADR\ta.md\tthree"} {
		// A problem before the line is not printed: the set is not known.
		input := recordsOf(validRecords[0], validRecords[0], validRecords[2], line, validRecords[1])
		stdout, stderr, code := verifyRun(input)
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "line 4 ") || strings.Count(stderr, "\n") != 1 {
			t.Errorf("%q: got %q, %q; want one message that gives line 4", line, stdout, stderr)
		}
		if code != 2 {
			t.Errorf("%q: exit %d, want 2", line, code)
		}
	}
}

func TestVerifyAnOperand(t *testing.T) {
	for _, operands := range [][]string{{"records.tsv"}, {"-"}, {"-x"}, {"--", "records.tsv"}} {
		stdout, stderr, code := verifyRun(recordsOf(validRecords...), operands...)
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.HasSuffix(stderr, "\n"+verifyUsage) || code != 2 {
			t.Errorf("%q: got %q, %q, exit %d; want the usage line, exit 2", operands, stdout, stderr, code)
		}
	}
}

func TestVerifyTheOrderOfProblems(t *testing.T) {
	verifyProblems(t, recordsOf(
		"breadcrumb\tB\tPBI\ttasks/b.md\t3",
		"link\tB\timplements\tNONE\tdocs/a.md\t9",
		"breadcrumb\tb\tADR\tdocs/a.md\t2",
	), "docs/a.md:2", "docs/a.md:9", "tasks/b.md:3")
}

func TestVerifyABreadcrumbExtractCouldNotRead(t *testing.T) {
	files(t, map[string]string{
		"a.md": fm("breadcrumb:", "  id: ADR-0005", "  type: ADR"),
		"b.md": crumbOf("PBI-00001", "implements ADR-0005"),
	})
	records, stderr, code := extractAll("a.md", "b.md")
	if !strings.HasPrefix(stderr, "a.md:2: ") || !strings.HasPrefix(records, "problem\ta.md\t2\t") || code != 1 {
		t.Errorf("bcr extract wrote %q, exit %d; want the problem of a.md, exit 1", stderr, code)
	}
	ps := verifyProblems(t, records, "b.md:6")
	naming(t, ps[0], `"ADR-0005"`)
}

// unwritable is an output that cannot be written.
type unwritable struct{}

func (unwritable) Write([]byte) (int, error) { return 0, errors.New("output is closed") }

func TestVerifyOutputThatCannotBeWritten(t *testing.T) {
	var stderr bytes.Buffer
	code := run([]string{"verify"}, strings.NewReader(recordsOf(validRecords...)), io.Writer(unwritable{}), &stderr)
	if stderr.String() != "bcr: output is closed\n" || code != 2 {
		t.Errorf("got %q, exit %d; want one message, exit 2", stderr.String(), code)
	}
}

// unreadable is an input that cannot be read.
type unreadable struct{}

func (unreadable) Read([]byte) (int, error) { return 0, errors.New("input is closed") }

func TestVerifyInputThatCannotBeRead(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"verify"}, io.Reader(unreadable{}), &stdout, &stderr)
	if stdout.String() != "" || stderr.String() != "bcr: input is closed\n" || code != 2 {
		t.Errorf("got %q, %q, exit %d", stdout.String(), stderr.String(), code)
	}
}

func TestVerifyAProblemIsAlsoARecord(t *testing.T) {
	input := recordsOf(
		"breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3",
		"link\tPBI-00001\timplements\tADR-0099\ttasks/PBI-00001.md\t6",
	)
	stdout, stderr, code := verifyRun(input)
	message := `link points to "ADR-0099", which is no breadcrumb's id`
	if stderr != "tasks/PBI-00001.md:6: "+message+"\n" {
		t.Errorf("standard error %q", stderr)
	}
	if want := input + "problem\ttasks/PBI-00001.md\t6\t" + message + "\n"; stdout != want {
		t.Errorf("standard output %q, want %q", stdout, want)
	}
	if code != 1 {
		t.Errorf("exit %d, want 1", code)
	}
}

func TestVerifyAProblemAfterALastLineWithNoLineEnding(t *testing.T) {
	input := "breadcrumb\tPBI-00001\tPBI\ta.md\t3\nlink\tPBI-00001\timplements\tADR-0099\ta.md\t6"
	stdout, _, _ := verifyRun(input)
	if !strings.HasPrefix(stdout, input+"\nproblem\ta.md\t6\t") || strings.Count(stdout, "\n") != 3 {
		t.Errorf("standard output %q", stdout)
	}
}

func TestVerifyAProblemRecordInTheInput(t *testing.T) {
	passes(t, recordsOf(append(validRecords, "problem\ta.md\t3\tid has no value")...))
}
