package records

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// pbi is a breadcrumb with two links, as tasks/PBI-00001.md writes it.
var pbi = crumb.Breadcrumb{
	ID:   "PBI-00001",
	Type: "PBI",
	Line: 3,
	Links: []crumb.Link{
		{Verb: "implements", Object: "ADR-0001", Line: 6},
		{Verb: "changes", Object: "asdlc", Line: 7},
	},
}

// The records of pbi, and the set they describe.
const pbiRecords = "breadcrumb\tPBI-00001\tPBI\ttasks/PBI-00001.md\t3\n" +
	"link\tPBI-00001\timplements\tADR-0001\ttasks/PBI-00001.md\t6\n" +
	"link\tPBI-00001\tchanges\tasdlc\ttasks/PBI-00001.md\t7\n"

var pbiSet = crumb.Set{
	Breadcrumbs: []crumb.SetBreadcrumb{{ID: "PBI-00001", At: crumb.Place{Text: "tasks/PBI-00001.md", Line: 3}}},
	Links: []crumb.SetLink{
		{Object: "ADR-0001", At: crumb.Place{Text: "tasks/PBI-00001.md", Line: 6}},
		{Object: "asdlc", At: crumb.Place{Text: "tasks/PBI-00001.md", Line: 7}},
	},
}

// read reads input, and fails on an error.
func read(t *testing.T, input string) crumb.Set {
	t.Helper()
	set, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	return set
}

// refused fails unless reading input is an error that gives line, and
// no set is returned.
func refused(t *testing.T, input string, line string) {
	t.Helper()
	set, err := Read(strings.NewReader(input))
	if err == nil {
		t.Errorf("%q: got no error", input)
		return
	}
	if !strings.Contains(err.Error(), "line "+line+" ") {
		t.Errorf("%q: error %q does not give line %s", input, err, line)
	}
	if !reflect.DeepEqual(set, crumb.Set{}) {
		t.Errorf("%q: got set %+v with an error, want none", input, set)
	}
}

func TestWriteABreadcrumbWithLinks(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "tasks/PBI-00001.md", pbi); err != nil {
		t.Fatal(err)
	}
	if out.String() != pbiRecords {
		t.Errorf("got %q, want %q", out.String(), pbiRecords)
	}
}

func TestWriteABreadcrumbWithNoLinks(t *testing.T) {
	var out bytes.Buffer
	Write(&out, "a.md", crumb.Breadcrumb{ID: "A", Type: "ADR", Line: 3})
	if want := "breadcrumb\tA\tADR\ta.md\t3\n"; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

// spec is pbi with two claims, as a spec writes them.
func spec() crumb.Breadcrumb {
	b := pbi
	b.Claims = []crumb.Claim{
		{Target: "docs/bcr.md", Kind: "has-line", Argument: "### verify", Line: 9},
		{Target: "cmd/bcr/main.go", Kind: "has-line", Argument: "return fmt.Sprintf('%s', id)", Line: 10},
	}
	return b
}

func TestWriteABreadcrumbWithClaims(t *testing.T) {
	var out bytes.Buffer
	if err := Write(&out, "tasks/PBI-00001.md", spec()); err != nil {
		t.Fatal(err)
	}
	want := pbiRecords +
		"claim\tPBI-00001\thas-line\tdocs/bcr.md\t### verify\ttasks/PBI-00001.md\t9\n" +
		"claim\tPBI-00001\thas-line\tcmd/bcr/main.go\treturn fmt.Sprintf('%s', id)\ttasks/PBI-00001.md\t10\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

func TestReadSkipsTheClaimsWriteWrites(t *testing.T) {
	var out bytes.Buffer
	Write(&out, "tasks/PBI-00001.md", spec())
	if got := read(t, out.String()); !reflect.DeepEqual(got, pbiSet) {
		t.Errorf("got %+v, want %+v", got, pbiSet)
	}
}

// failing is a writer that cannot be written to.
type failing struct{}

func (failing) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestWriteReturnsTheErrorOfItsWriter(t *testing.T) {
	if err := Write(failing{}, "a.md", pbi); err == nil {
		t.Error("got no error")
	}
}

func TestWhatIsWrittenIsReadBackTheSame(t *testing.T) {
	adr := crumb.Breadcrumb{ID: "ADR-0001", Type: "ADR", Line: 4}
	var out bytes.Buffer
	Write(&out, "tasks/PBI-00001.md", pbi)
	Write(&out, "docs/adrs/my decision.md", adr)
	want := crumb.Set{
		Breadcrumbs: append(append([]crumb.SetBreadcrumb{}, pbiSet.Breadcrumbs...),
			crumb.SetBreadcrumb{ID: "ADR-0001", At: crumb.Place{Text: "docs/adrs/my decision.md", Line: 4}}),
		Links: pbiSet.Links,
	}
	if got := read(t, out.String()); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadNoInput(t *testing.T) {
	if got := read(t, ""); !reflect.DeepEqual(got, crumb.Set{}) {
		t.Errorf("got %+v, want an empty set", got)
	}
}

func TestReadALastLineThatDoesNotEnd(t *testing.T) {
	if got := read(t, strings.TrimSuffix(pbiRecords, "\n")); !reflect.DeepEqual(got, pbiSet) {
		t.Errorf("got %+v, want %+v", got, pbiSet)
	}
}

func TestReadWindowsLineEndings(t *testing.T) {
	if got := read(t, strings.ReplaceAll(pbiRecords, "\n", "\r\n")); !reflect.DeepEqual(got, pbiSet) {
		t.Errorf("got %+v, want %+v", got, pbiSet)
	}
}

func TestReadFieldsAddedAtTheEnd(t *testing.T) {
	input := strings.ReplaceAll(pbiRecords, "\n", "\tnew\t\tmore\n")
	if got := read(t, input); !reflect.DeepEqual(got, pbiSet) {
		t.Errorf("got %+v, want %+v", got, pbiSet)
	}
}

func TestReadRecordsOfAKindItDoesNotKnow(t *testing.T) {
	input := "claim\tPBI-00001\tcontains\n" + pbiRecords + "claim\nBreadcrumb\tX\tPBI\tx.md\t3\nlinks\t\t\n"
	if got := read(t, input); !reflect.DeepEqual(got, pbiSet) {
		t.Errorf("got %+v, want %+v", got, pbiSet)
	}
}

func TestReadALineThatIsNotARecord(t *testing.T) {
	const valid = "breadcrumb\tADR-0002\tADR\ta.md\t3\n"
	for _, line := range []string{
		"breadcrumb\tADR-0001\tADR",                             // too few fields
		"breadcrumb\tADR-0001\tADR\ta.md",                       // no LINE
		"link\tPBI-00001\timplements\tADR-0001\t6",              // no PATH
		"link\tPBI-00001\timplements\tADR-0001",                 // too few fields
		"breadcrumb",                                            // only its kind
		"",                                                      // empty
		"\tADR-0001\tADR\ta.md\t3",                              // no kind
		"breadcrumb\t\tADR\ta.md\t3",                            // an empty ID
		"breadcrumb\tADR-0001\t\ta.md\t3",                       // an empty TYPE
		"breadcrumb\tADR-0001\tADR\t\t3",                        // an empty PATH
		"link\tPBI-00001\timplements\t\ta.md\t6",                // an empty OBJECT
		"breadcrumb\tADR-0001\tADR\ta.md\t",                     // an empty LINE
		"breadcrumb\tADR-0001\tADR\ta.md\t0",                    // LINE is not from 1
		"breadcrumb\tADR-0001\tADR\ta.md\t-3",                   // LINE is not from 1
		"breadcrumb\tADR-0001\tADR\ta.md\t+3",                   // LINE is not written in digits
		"breadcrumb\tADR-0001\tADR\ta.md\t3.0",                  // LINE is not whole
		"breadcrumb\tADR-0001\tADR\ta.md\t 3",                   // LINE is not written in digits
		"breadcrumb\tADR-0001\tADR\ta.md\tthree",                // LINE is not a number
		"breadcrumb\tADR-0001\tADR\ta.md\t99999999999999999999", // LINE is too large
		"breadcrumb\tADR-0001\tADR\ta.md\tx\t3",                 // LINE is the fifth field, not the last
	} {
		refused(t, valid+valid+valid+line+"\n"+valid, "4")
		refused(t, line+"\r\n", "1")
		if line != "" { // no input at all is no record, and no error
			refused(t, line, "1")
		}
	}
}

func TestReadAnEmptyLineAtTheEnd(t *testing.T) {
	refused(t, pbiRecords+"\n", "4")
	refused(t, pbiRecords+"\r\n", "4")
}

// broken is a reader that fails after its text.
type broken struct{ text io.Reader }

func (b broken) Read(p []byte) (int, error) {
	if n, _ := b.text.Read(p); n > 0 {
		return n, nil
	}
	return 0, errors.New("cannot read")
}

func TestReadReturnsTheErrorOfItsReader(t *testing.T) {
	set, err := Read(broken{strings.NewReader(pbiRecords)})
	if err == nil || err.Error() != "cannot read" || !reflect.DeepEqual(set, crumb.Set{}) {
		t.Errorf("got %+v and error %v, want no set and the reader's error", set, err)
	}
}

func TestWriteProblem(t *testing.T) {
	var out strings.Builder
	if err := WriteProblem(&out, "tasks/PBI-00001.md", 2, "breadcrumb has no type"); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "problem\ttasks/PBI-00001.md\t2\tbreadcrumb has no type\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteAProblemsMessageOnOneLine(t *testing.T) {
	var out strings.Builder
	WriteProblem(&out, "a.md", 3, "a\tb\r\nc")
	if got, want := out.String(), "problem\ta.md\t3\ta b  c\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReadSkipsTheProblemsWriteProblemWrites(t *testing.T) {
	var out strings.Builder
	WriteProblem(&out, "a.md", 3, "id has no value")
	if set := read(t, out.String()); len(set.Breadcrumbs) != 0 || len(set.Links) != 0 {
		t.Errorf("got %+v, want an empty set", set)
	}
}

// The records of a spec with two claims, one dropped, as bcr extract
// and bcr verify print them.
const auditInput = "breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3\n" +
	"link\tverify\tfollows\tADR-0016\tspecs/verify/spec.md\t6\n" +
	"claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8\n" +
	"problem\tspecs/verify/spec.md\t9\tclaim must be written in single quotes\n" +
	"note\tanything\n"

func TestReadAuditBreadcrumbsClaimsAndProblems(t *testing.T) {
	want := crumb.Set{
		Breadcrumbs: []crumb.SetBreadcrumb{{ID: "verify", At: crumb.Place{Text: "specs/verify/spec.md", Line: 3}}},
		Claims: []crumb.SetClaim{{
			ID:    "verify",
			Claim: crumb.Claim{Target: "docs/bcr.md", Kind: "has-line", Argument: "### verify", Line: 8},
			At:    crumb.Place{Text: "specs/verify/spec.md", Line: 8},
		}},
		Problems: []crumb.SetProblem{{
			At:      crumb.Place{Text: "specs/verify/spec.md", Line: 9},
			Message: "claim must be written in single quotes",
		}},
	}
	for _, input := range []string{auditInput, strings.ReplaceAll(auditInput, "\n", "\r\n"), strings.TrimSuffix(auditInput, "\n")} {
		got, err := ReadAudit(strings.NewReader(input))
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %+v, %v; want %+v", input, got, err, want)
		}
	}
}

func TestReadAuditFieldsAddedAtTheEnd(t *testing.T) {
	input := "breadcrumb\tA\tspec\ta.md\t3\tmore\n" +
		"claim\tA\thas-line\tdocs/bcr.md\t### verify\ta.md\t8\tmore\n" +
		"problem\ta.md\t9\tid has no value\tmore\n"
	got, err := ReadAudit(strings.NewReader(input))
	if err != nil || len(got.Breadcrumbs) != 1 || len(got.Claims) != 1 || len(got.Problems) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.Claims[0].At.Line != 8 || got.Problems[0].Message != "id has no value" {
		t.Errorf("got %+v", got)
	}
}

func TestReadAuditSkipsALinkThatIsNotARecord(t *testing.T) {
	if _, err := ReadAudit(strings.NewReader("link\tPBI-00001\n")); err != nil {
		t.Errorf("got error %v, want none: bcr audit does not read links", err)
	}
}

func TestReadAuditALineThatIsNotARecord(t *testing.T) {
	const valid = "breadcrumb\tA\tspec\ta.md\t3\n"
	for _, line := range []string{
		"",
		"\tA\tspec\ta.md\t3",
		"breadcrumb\tA\tspec",
		"breadcrumb\tA\tspec\ta.md\tx",
		"claim\tverify\thas-line\tdocs/bcr.md", // too few fields
		"claim\tverify\thas-line\tdocs/bcr.md\t### verify\ta.md",     // no LINE
		"claim\tverify\thas-line\tdocs/bcr.md\t\ta.md\t8",            // an empty ARGUMENT
		"claim\tverify\thas-line\tdocs/bcr.md\tx\ta.md\t0",           // LINE is not from 1
		"claim\tverify\thas-line\t/etc/passwd\tx\ta.md\t8",           // a target from the root of the machine
		"claim\tverify\thas-line\t../x.md\tx\ta.md\t8",               // a target outside the repository
		"claim\tverify\thas-line\tdocs/../../x.md\tx\ta.md\t8",       // a target outside the repository
		"claim\tverify\thas-line\tdocs/a b.md\tx\ta.md\t8",           // a target with white space
		"claim\tverify\tcontains\tdocs/bcr.md\tx\ta.md\t8",           // a kind bcr does not know
		"claim\tverify\thas-line\tdocs/bcr.md\t### verify \ta.md\t8", // a text no line can equal
		"problem\ta.md\t3",                      // no MESSAGE
		"problem\ta.md\t3\t",                    // an empty MESSAGE
		"problem\t\t3\tid has no value",         // an empty PATH
		"problem\ta.md\tthree\tid has no value", // LINE is not a number
	} {
		set, err := ReadAudit(strings.NewReader(valid + line + "\n" + valid))
		if err == nil || !strings.Contains(err.Error(), "line 2 ") {
			t.Errorf("%q: got error %v, want one that gives line 2", line, err)
		}
		if !reflect.DeepEqual(set, crumb.Set{}) {
			t.Errorf("%q: got set %+v with an error, want none", line, set)
		}
	}
}

func TestReadStillSkipsClaimsAndProblems(t *testing.T) {
	set := read(t, "claim\tverify\tcontains\t/etc/passwd\n"+"problem\ta.md\n")
	if !reflect.DeepEqual(set, crumb.Set{}) {
		t.Errorf("got %+v, want an empty set: bcr verify does not read these kinds", set)
	}
}

const reportInput = "breadcrumb\tverify\tspec\tspecs/verify/spec.md\t3\n" +
	"link\tverify\tfollows\tADR-0015\tspecs/verify/spec.md\t6\n" +
	"claim\tverify\thas-line\tdocs/bcr.md\t### verify\tspecs/verify/spec.md\t8\n" +
	"breadcrumb\tADR-0015\tADR\tdocs/adrs/a.md\t3\n" +
	"note\tanything\n" +
	"problem\tspecs/verify/spec.md\t9\tclaim must be written in single quotes\n" +
	"claim-verdict\tverify\tConfirmed\tspecs/verify/spec.md\t8\n" +
	"verdict\tverify\tUndecided\tspecs/verify/spec.md\t3\n" +
	"verdict\tADR-0015\tUndecided\tdocs/adrs/a.md\t3\tno claims\n"

func TestReadReportEveryKindOfThePipe(t *testing.T) {
	spec, adr := crumb.Place{Text: "specs/verify/spec.md", Line: 3}, crumb.Place{Text: "docs/adrs/a.md", Line: 3}
	want := crumb.Set{
		Breadcrumbs: []crumb.SetBreadcrumb{{ID: "verify", Type: "spec", At: spec}, {ID: "ADR-0015", Type: "ADR", At: adr}},
		Links:       []crumb.SetLink{{Verb: "follows", Object: "ADR-0015", At: crumb.Place{Text: "specs/verify/spec.md", Line: 6}}},
		Claims: []crumb.SetClaim{{
			ID:    "verify",
			Claim: crumb.Claim{Target: "docs/bcr.md", Kind: "has-line", Argument: "### verify", Line: 8},
			At:    crumb.Place{Text: "specs/verify/spec.md", Line: 8},
		}},
		Problems: []crumb.SetProblem{{
			At:      crumb.Place{Text: "specs/verify/spec.md", Line: 9},
			Message: "claim must be written in single quotes",
		}},
		ClaimVerdicts: []crumb.SetVerdict{{At: crumb.Place{Text: "specs/verify/spec.md", Line: 8}, Verdict: crumb.Confirmed}},
		Verdicts: []crumb.SetVerdict{
			{At: spec, Verdict: crumb.Undecided},
			{At: adr, Verdict: crumb.Undecided, Reason: crumb.NoClaims},
		},
	}
	for _, input := range []string{reportInput, strings.ReplaceAll(reportInput, "\n", "\r\n"), strings.TrimSuffix(reportInput, "\n")} {
		got, err := ReadReport(strings.NewReader(input))
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %+v, %v; want %+v", input, got, err, want)
		}
	}
}

func TestReadReportReadsWhatTheWritersWrite(t *testing.T) {
	var b strings.Builder
	at := crumb.Place{Text: "a.md", Line: 3}
	for _, err := range []error{
		WriteClaimVerdict(&b, "A", crumb.Refuted, crumb.Place{Text: "a.md", Line: 8}),
		WriteVerdict(&b, "A", crumb.BreadcrumbVerdict{Verdict: crumb.Undecided, Reason: crumb.NoClaims}, at),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := ReadReport(strings.NewReader(b.String()))
	want := crumb.Set{
		ClaimVerdicts: []crumb.SetVerdict{{At: crumb.Place{Text: "a.md", Line: 8}, Verdict: crumb.Refuted}},
		Verdicts:      []crumb.SetVerdict{{At: at, Verdict: crumb.Undecided, Reason: crumb.NoClaims}},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, %v; want %+v", got, err, want)
	}
}

func TestReadReportFieldsAddedAtTheEnd(t *testing.T) {
	input := "breadcrumb\tA\tspec\ta.md\t3\tmore\n" +
		"link\tA\tfollows\tB\ta.md\t6\tmore\n" +
		"claim-verdict\tA\tConfirmed\ta.md\t8\tmore\n" +
		"verdict\tA\tUndecided\ta.md\t3\tno claims\tmore\n"
	got, err := ReadReport(strings.NewReader(input))
	if err != nil || len(got.Breadcrumbs) != 1 || len(got.Links) != 1 || len(got.ClaimVerdicts) != 1 || len(got.Verdicts) != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.Links[0].At.Line != 6 || got.ClaimVerdicts[0].At.Line != 8 || got.Verdicts[0].Reason != "no claims" {
		t.Errorf("got %+v", got)
	}
}

func TestReadReportALineThatIsNotARecord(t *testing.T) {
	const valid = "breadcrumb\tA\tspec\ta.md\t3\n"
	for _, line := range []string{
		"",
		"\tA\tspec\ta.md\t3",
		"breadcrumb\tA\tspec\ta.md\tx",
		"link\tA\tfollows\tB\ta.md", // no LINE
		"link\tA\t\tB\ta.md\t6",     // an empty VERB
		"claim\tverify\tcontains\tdocs/bcr.md\tx\ta.md\t8", // a kind bcr does not know
		"claim\tverify\thas-line\t../x.md\tx\ta.md\t8",     // a target outside the repository
		"problem\ta.md\t3",                              // no MESSAGE
		"claim-verdict\tA\tConfirmed\ta.md",             // no LINE
		"claim-verdict\tA\tPassed\ta.md\t8",             // no such verdict
		"claim-verdict\tA\tConfirmed\ta.md\t0",          // LINE is not from 1
		"verdict\tA\tconfirmed\ta.md\t3",                // letter case
		"verdict\tA\t\ta.md\t3",                         // an empty VERDICT
		"verdict\tA\tUndecided\ta.md\tthree\tno claims", // LINE is not a number
	} {
		set, err := ReadReport(strings.NewReader(valid + line + "\n" + valid))
		if err == nil || !strings.Contains(err.Error(), "line 2 ") {
			t.Errorf("%q: got error %v, want one that gives line 2", line, err)
		}
		if !reflect.DeepEqual(set, crumb.Set{}) {
			t.Errorf("%q: got set %+v with an error, want none", line, set)
		}
	}
}

func TestReadAndReadAuditStillSkipVerdicts(t *testing.T) {
	const input = "claim-verdict\tA\tPassed\n" + "verdict\tA\n"
	if set := read(t, input); !reflect.DeepEqual(set, crumb.Set{}) {
		t.Errorf("Read got %+v, want an empty set", set)
	}
	if set, err := ReadAudit(strings.NewReader(input)); err != nil || !reflect.DeepEqual(set, crumb.Set{}) {
		t.Errorf("ReadAudit got %+v, %v; want an empty set", set, err)
	}
}
