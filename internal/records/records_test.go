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
