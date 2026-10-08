package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// The tests below follow the Scenarios of specs/filter/spec.md, one
// test each, in the same order.

// filterRun runs bcr filter with args and input on standard input,
// and returns what it writes to standard output and standard error,
// and its exit status.
func filterRun(input string, args ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	code := run(append([]string{"filter"}, args...), strings.NewReader(input), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

// filtered fails unless bcr filter, given args and the records of
// input, prints the records of want, nothing to standard error, and
// exits 0.
func filtered(t *testing.T, args []string, input []string, want ...string) {
	t.Helper()
	wanted := ""
	if len(want) > 0 {
		wanted = recordsOf(want...)
	}
	if stdout, stderr, code := filterRun(recordsOf(input...), args...); stdout != wanted || stderr != "" || code != 0 {
		t.Errorf("%q: got %q, %q, exit %d; want %q, nothing, exit 0", args, stdout, stderr, code, wanted)
	}
}

// The records bcr audit prints for two specs, verify and audit.
var (
	verifyRecords = []string{
		"breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3",
		"link\tverify\tfollows\tADR-0015\tspecs/verify/spec.md\t6",
		"claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8",
	}
	otherRecords = []string{
		"breadcrumb\taudit\tspec\tspecs/audit/spec.md\t3",
		"link\taudit\tfollows\tADR-0023\tspecs/audit/spec.md\t6",
		"claim\taudit\thas-line\tdocs/bcr.md\t### audit\tspecs/audit/spec.md\t8",
	}
	verifyVerdicts = []string{
		"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8",
		"verdict\tverify\tConfirmed\tspecs/verify/spec.md\t3",
	}
	otherVerdicts = []string{
		"claim-verdict\taudit\tRefuted\tspecs/audit/spec.md\t8",
		"verdict\taudit\tRefuted\tspecs/audit/spec.md\t3",
	}
)

// join returns the records of each of sets, in order.
func join(sets ...[]string) []string {
	var all []string
	for _, s := range sets {
		all = append(all, s...)
	}
	return all
}

func TestFilterOneID(t *testing.T) {
	filtered(t, []string{"--id", "ADR-0001"}, []string{
		"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
		"link\tADR-0001\tamends\tADR-0002\tdocs/adrs/a.md\t6",
		"breadcrumb\tADR-0002\tADR\tdocs/adrs/b.md\t3",
	},
		"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
		"link\tADR-0001\tamends\tADR-0002\tdocs/adrs/a.md\t6",
	)
}

func TestFilterABreadcrumbsClaimsFollowItsID(t *testing.T) {
	filtered(t, []string{"--id", "verify"}, join(otherRecords, verifyRecords), verifyRecords...)
	filtered(t, []string{"--id=verify"}, join(verifyRecords, otherRecords), verifyRecords...)
}

func TestFilterSeveralValuesOfOneFlag(t *testing.T) {
	filtered(t, []string{"-i", "ADR-0001", "-i", "ADR-0003"}, []string{
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"breadcrumb\tADR-0002\tADR\tb.md\t3",
		"breadcrumb\tADR-0003\tADR\tc.md\t3",
	},
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"breadcrumb\tADR-0003\tADR\tc.md\t3",
	)
}

func TestFilterTwoFlags(t *testing.T) {
	input := []string{
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"breadcrumb\tPBI-00001\tPBI\tb.md\t3",
	}
	filtered(t, []string{"--id", "ADR-0001", "--type", "PBI"}, input)
	filtered(t, []string{"--id", "ADR-0001", "--type", "ADR"}, input, input[0])
}

func TestFilterOneType(t *testing.T) {
	filtered(t, []string{"--type", "ADR"}, []string{
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"link\tADR-0001\tamends\tADR-0002\ta.md\t6",
		"breadcrumb\tPBI-00001\tPBI\tb.md\t3",
		"link\tPBI-00001\timplements\tADR\tb.md\t6",
		"breadcrumb\tADR-0002\tADR\tc.md\t3",
	},
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"breadcrumb\tADR-0002\tADR\tc.md\t3",
	)
}

func TestFilterOneKind(t *testing.T) {
	filtered(t, []string{"--kind", "link"}, verifyRecords, verifyRecords[1])
	filtered(t, []string{"-k", "link", "-k", "claim"}, verifyRecords, verifyRecords[1], verifyRecords[2])
}

func TestFilterAKindAndAnID(t *testing.T) {
	filtered(t, []string{"--kind", "link", "--id", "ADR-0001"}, []string{
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
		"link\tADR-0001\tamends\tADR-0002\ta.md\t6",
		"link\tADR-0002\tamends\tADR-0001\tb.md\t6",
		"link\tADR-0001\tfollows\tADR-0003\ta.md\t7",
	},
		"link\tADR-0001\tamends\tADR-0002\ta.md\t6",
		"link\tADR-0001\tfollows\tADR-0003\ta.md\t7",
	)
}

// The records of ADR-0019, of a breadcrumb that follows it and of a
// PBI that implements it.
var objectRecords = []string{
	"breadcrumb\tADR-0019\tADR\tdocs/adrs/c.md\t3",
	"link\tADR-0019\tfollows\tADR-0017\tdocs/adrs/c.md\t6",
	"breadcrumb\tADR-0022\tADR\tdocs/adrs/d.md\t3",
	"link\tADR-0022\tfollows\tADR-0019\tdocs/adrs/d.md\t7",
	"breadcrumb\tPBI-00016\tPBI\ttasks/PBI-00016.md\t3",
	"link\tPBI-00016\timplements\tADR-0019\ttasks/PBI-00016.md\t6",
}

func TestFilterWhatLinksToABreadcrumb(t *testing.T) {
	filtered(t, []string{"--object", "ADR-0019"}, objectRecords, objectRecords[3], objectRecords[5])
	filtered(t, []string{"-o", "ADR-0019", "-o", "ADR-0017"}, objectRecords, objectRecords[1], objectRecords[3], objectRecords[5])
	// Only a link record has an OBJECT field: another kind does not
	// match, even with the value in its fourth field.
	filtered(t, []string{"--object", "ADR-0019", "--kind", "breadcrumb"}, objectRecords)
	filtered(t, []string{"--object", "ADR-0019"}, []string{
		"claim\tA\thas-line\tADR-0019\tx\ta.md\t8",
		"claim-verdict\tA\tConfirmed\tADR-0019\t8",
		"verdict\tA\tConfirmed\tADR-0019\t3",
		"problem\ta.md\t3\tADR-0019",
		"note\tA\tx\tADR-0019",
	})
}

func TestFilterALinkFromOneBreadcrumbToAnother(t *testing.T) {
	filtered(t, []string{"--id", "PBI-00016", "--object", "ADR-0019"}, objectRecords, objectRecords[5])
	filtered(t, []string{"--id", "ADR-0019", "--object", "ADR-0019"}, objectRecords)
	// bcr filter judges nothing: a link to itself, written by hand, is
	// printed like any other record.
	filtered(t, []string{"--id", "X", "--object", "X"}, []string{"link\tX\tamends\tX\ta.md\t6"}, "link\tX\tamends\tX\ta.md\t6")
}

func TestFilterProblemRecords(t *testing.T) {
	input := []string{
		"breadcrumb\tADR-0001\tADR\tdocs/adrs/a.md\t3",
		"problem\tdocs/adrs/a.md\t9\tclaim must be written in single quotes",
	}
	filtered(t, []string{"--id", "ADR-0001"}, input, input[0])
	filtered(t, []string{"--kind", "problem"}, input, input[1])
	// A problem record has no ID field: its PATH is not one.
	filtered(t, []string{"--id", "docs/adrs/a.md"}, input)
	filtered(t, []string{"--kind", "problem", "--id", "ADR-0001"}, input)
	filtered(t, []string{"--object", "docs/adrs/a.md"}, input)
}

func TestFilterAKindBcrDoesNotKnow(t *testing.T) {
	input := append([]string{"note\tx"}, validRecords...)
	filtered(t, []string{"--kind", "note"}, input, "note\tx")
	filtered(t, []string{"--kind", "nonsense"}, input)
	// It matches only --kind: its second field is not an ID.
	filtered(t, []string{"--id", "x"}, input)
	filtered(t, []string{"--type", "x"}, input)
	filtered(t, []string{"--object", "x"}, input)
}

func TestFilterNoMatch(t *testing.T) {
	filtered(t, []string{"--id", "ADR-9999"}, validRecords)
}

func TestFilterTwoBreadcrumbsWithTheSameID(t *testing.T) {
	input := []string{
		"breadcrumb\tADR-0003\tADR\ta.md\t3",
		"breadcrumb\tADR-0003\tADR\tb.md\t3",
	}
	filtered(t, []string{"--id", "ADR-0003"}, input, input...)
}

func TestFilterLetterCase(t *testing.T) {
	input := []string{"breadcrumb\tADR-0001\tADR\ta.md\t3"}
	filtered(t, []string{"--id", "adr-0001"}, input)
	filtered(t, []string{"--type", "adr"}, input)
	filtered(t, []string{"--kind", "Breadcrumb"}, input)
	filtered(t, []string{"--object", "adr-0002"}, []string{"link\tADR-0001\tamends\tADR-0002\ta.md\t6"})
}

func TestFilterTheOrderOfTheOutput(t *testing.T) {
	input := []string{
		"breadcrumb\tADR-0002\tADR\tb.md\t3",
		"breadcrumb\tADR-0001\tADR\ta.md\t3",
	}
	filtered(t, []string{"--type", "ADR"}, input, input...)
	// The order of the flags' values does not change it.
	filtered(t, []string{"-i", "ADR-0001", "-i", "ADR-0002"}, input, input...)
}

func TestFilterNoFlag(t *testing.T) {
	stdout, stderr, code := filterRun(recordsOf(validRecords...))
	if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.HasSuffix(stderr, "\n"+filterUsage) || code != 2 {
		t.Errorf("got %q, %q, exit %d; want the usage line, exit 2", stdout, stderr, code)
	}
}

func TestFilterAnOperand(t *testing.T) {
	for _, args := range [][]string{{"--id", "ADR-0001", "records.tsv"}, {"records.tsv"}, {"--id", "ADR-0001", "--", "-k"}} {
		stdout, stderr, code := filterRun(recordsOf(validRecords...), args...)
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.HasSuffix(stderr, "\n"+filterUsage) || code != 2 {
			t.Errorf("%q: got %q, %q, exit %d; want the usage line, exit 2", args, stdout, stderr, code)
		}
	}
}

func TestFilterALineThatIsNotARecord(t *testing.T) {
	for _, line := range []string{"breadcrumb\tADR-0001\tADR", "", "breadcrumb\tADR-0001\tADR\ta.md\tthree", "\tADR-0001"} {
		input := recordsOf(validRecords[0], validRecords[1], line, validRecords[2])
		stdout, stderr, code := filterRun(input, "--id", "ADR-0001")
		if stdout != "" || !strings.HasPrefix(stderr, "bcr: ") || !strings.Contains(stderr, "line 3 ") || strings.Count(stderr, "\n") != 1 || code != 2 {
			t.Errorf("%q: got %q, %q, exit %d; want one message that gives line 3, exit 2", line, stdout, stderr, code)
		}
	}
	// A claim and a verdict are not judged: only their shape is read.
	filtered(t, []string{"--id", "A"}, []string{
		"claim\tA\tcontains\t../x.md\tx\ta.md\t8",
		"verdict\tA\tPassed\ta.md\t3",
	},
		"claim\tA\tcontains\t../x.md\tx\ta.md\t8",
		"verdict\tA\tPassed\ta.md\t3",
	)
}

func TestFilterFieldsAddedAtTheEnd(t *testing.T) {
	input := []string{"breadcrumb\tADR-0001\tADR\ta.md\t3\tmore\t", "breadcrumb\tADR-0002\tADR\tb.md\t3\tADR-0001"}
	filtered(t, []string{"--id", "ADR-0001"}, input, input[0])
}

func TestFilterWindowsLineEndings(t *testing.T) {
	const input = "breadcrumb\tADR-0001\tADR\ta.md\t3\r\nbreadcrumb\tADR-0002\tADR\tb.md\t3\r\nlink\tADR-0001\tamends\tADR-0002\ta.md\t6"
	const want = "breadcrumb\tADR-0001\tADR\ta.md\t3\r\nlink\tADR-0001\tamends\tADR-0002\ta.md\t6"
	if stdout, stderr, code := filterRun(input, "--id", "ADR-0001"); stdout != want || stderr != "" || code != 0 {
		t.Errorf("got %q, %q, exit %d; want %q", stdout, stderr, code, want)
	}
}

func TestFilterOutputThatCannotBeWritten(t *testing.T) {
	var stderr bytes.Buffer
	code := run([]string{"filter", "--kind", "breadcrumb"}, strings.NewReader(recordsOf(validRecords...)), io.Writer(unwritable{}), &stderr)
	if stderr.String() != "bcr: output is closed\n" || code != 2 {
		t.Errorf("got %q, exit %d; want one message, exit 2", stderr.String(), code)
	}
}

func TestFilterOneBreadcrumbsVerdict(t *testing.T) {
	input := join(verifyRecords, otherRecords, []string{verifyVerdicts[0], otherVerdicts[0], otherVerdicts[1], verifyVerdicts[1]})
	filtered(t, []string{"--id", "verify"}, input, join(verifyRecords, verifyVerdicts)...)
	filtered(t, []string{"--kind", "verdict", "--id", "verify"}, input, verifyVerdicts[1])
}

func TestFilterBeforeBcrVerify(t *testing.T) {
	files(t, map[string]string{
		"a.md": crumbOf("ADR-0001", "amends ADR-0002"),
		"b.md": crumbOf("ADR-0002"),
	})
	extracted, _, code := extractAll("a.md", "b.md")
	if code != 0 {
		t.Fatalf("bcr extract exits %d", code)
	}
	passes(t, extracted)
	kept, stderr, code := filterRun(extracted, "--id", "ADR-0001")
	if stderr != "" || code != 0 {
		t.Fatalf("bcr filter wrote %q, exit %d", stderr, code)
	}
	ps := verifyProblems(t, kept, "a.md:6")
	naming(t, ps[0], `"ADR-0002"`)
}

func TestFilterInputThatCannotBeRead(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"filter", "--id", "ADR-0001"}, io.Reader(unreadable{}), &stdout, &stderr)
	if stdout.String() != "" || stderr.String() != "bcr: input is closed\n" || code != 2 {
		t.Errorf("got %q, %q, exit %d", stdout.String(), stderr.String(), code)
	}
}
