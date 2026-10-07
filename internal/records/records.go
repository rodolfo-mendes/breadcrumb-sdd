// Package records writes and reads the records the commands of the
// pipe print: one line for each breadcrumb, each link, each claim,
// each problem and each verdict, its fields separated by a tab
// (ADR-0011, ADR-0021, ADR-0023). It is the only package that knows
// this format; the core does not use it (ADR-0013).
//
//	breadcrumb	ID	TYPE	PATH	LINE
//	link	ID	VERB	OBJECT	PATH	LINE
//	claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
//	problem	PATH	LINE	MESSAGE
//	claim-verdict	ID	VERDICT	PATH	LINE
//	verdict	ID	VERDICT	PATH	LINE
//	verdict	ID	Undecided	PATH	LINE	no claims
package records

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// The kind of a record is its first field, with the number of fields
// it needs. Fields after those are ignored: new ones may be added at
// the end of a record.
const (
	breadcrumbKind   = "breadcrumb"
	breadcrumbFields = 5
	linkKind         = "link"
	linkFields       = 6
	claimKind        = "claim" // skipped by Read
	claimFields      = 7
	problemKind      = "problem" // skipped by Read
	problemFields    = 4
	claimVerdictKind = "claim-verdict" // written only
	verdictKind      = "verdict"       // written only
)

// Write writes the records of b, the breadcrumb of the file at path:
// its breadcrumb record, then a link record for each of its links and
// a claim record for each of its claims, each in the order written.
func Write(w io.Writer, path string, b crumb.Breadcrumb) error {
	if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", breadcrumbKind, b.ID, b.Type, path, b.Line); err != nil {
		return err
	}
	for _, l := range b.Links {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n", linkKind, b.ID, l.Verb, l.Object, path, l.Line); err != nil {
			return err
		}
	}
	for _, c := range b.Claims {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\n", claimKind, b.ID, c.Kind, c.Target, c.Argument, path, c.Line); err != nil {
			return err
		}
	}
	return nil
}

// oneLine writes a tab or a line ending as a space, so that a message
// stays one field of one line.
var oneLine = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")

// WriteProblem writes the record of a problem: message, about line of
// the file at path, as it is printed to standard error (ADR-0021).
func WriteProblem(w io.Writer, path string, line int, message string) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", problemKind, path, line, oneLine.Replace(message))
	return err
}

// WriteClaimVerdict writes the record of the verdict of the claim
// written at at, in the breadcrumb whose id is id (ADR-0023).
func WriteClaimVerdict(w io.Writer, id string, v crumb.Verdict, at crumb.Place) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", claimVerdictKind, id, v, at.Text, at.Line)
	return err
}

// WriteVerdict writes the record of the verdict of the breadcrumb
// written at at, whose id is id. Its reason, when it has one, is one
// more field at the end (ADR-0022, ADR-0023).
func WriteVerdict(w io.Writer, id string, v crumb.BreadcrumbVerdict, at crumb.Place) error {
	reason := ""
	if v.Reason != "" {
		reason = "\t" + v.Reason
	}
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d%s\n", verdictKind, id, v.Verdict, at.Text, at.Line, reason)
	return err
}

// Read reads records from r until it ends, and returns the set of
// breadcrumbs they describe, each breadcrumb and each link at the
// path and line of its record. A line may end in \n or \r\n. A record
// of a kind Read does not know is skipped. A line that is not a
// record, an empty one included, is an error that gives its number in
// the input, counted from 1.
func Read(r io.Reader) (crumb.Set, error) {
	return readWith(r, add)
}

// ReadAudit reads records as Read does, for bcr audit: it returns the
// breadcrumbs, the claims and the problems of the set, and skips the
// links with the kinds it does not know. A claim record whose target,
// kind or argument breaks a rule of the core is not a record bcr
// extract could print (ADR-0017).
func ReadAudit(r io.Reader) (crumb.Set, error) {
	return readWith(r, addAudit)
}

// readWith reads the records of r into a set with addLine, which returns why
// a line is not a record, or "".
func readWith(r io.Reader, addLine func(*crumb.Set, string) string) (crumb.Set, error) {
	var set crumb.Set
	in := bufio.NewReader(r)
	for n := 1; ; n++ {
		line, err := in.ReadString('\n')
		if err != nil && err != io.EOF {
			return crumb.Set{}, err
		}
		if err == io.EOF && line == "" {
			return set, nil
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if why := addLine(&set, line); why != "" {
			return crumb.Set{}, fmt.Errorf("line %d of the input is not a record: %s", n, why)
		}
	}
}

// add adds the record in line to set when it is a breadcrumb or a
// link, or returns why line is not a record.
func add(set *crumb.Set, line string) string {
	fields, why := split(line, map[string]int{breadcrumbKind: breadcrumbFields, linkKind: linkFields})
	if fields == nil {
		return why
	}
	at, why := place(fields[len(fields)-2], fields[len(fields)-1])
	if why != "" {
		return why
	}
	if fields[0] == breadcrumbKind {
		set.Breadcrumbs = append(set.Breadcrumbs, crumb.SetBreadcrumb{ID: fields[1], At: at})
	} else {
		set.Links = append(set.Links, crumb.SetLink{Object: fields[3], At: at})
	}
	return ""
}

// addAudit adds the record in line to set when it is a breadcrumb, a
// claim or a problem, or returns why line is not a record.
func addAudit(set *crumb.Set, line string) string {
	fields, why := split(line, map[string]int{breadcrumbKind: breadcrumbFields, claimKind: claimFields, problemKind: problemFields})
	if fields == nil {
		return why
	}
	if fields[0] == problemKind {
		at, why := place(fields[1], fields[2])
		if why != "" {
			return why
		}
		set.Problems = append(set.Problems, crumb.SetProblem{At: at, Message: fields[3]})
		return ""
	}
	at, why := place(fields[len(fields)-2], fields[len(fields)-1])
	if why != "" {
		return why
	}
	if fields[0] == breadcrumbKind {
		set.Breadcrumbs = append(set.Breadcrumbs, crumb.SetBreadcrumb{ID: fields[1], At: at})
		return ""
	}
	c := crumb.Claim{Kind: fields[2], Target: fields[3], Argument: fields[4], Line: at.Line}
	if why := c.Check(); why != "" {
		return "its claim " + why
	}
	set.Claims = append(set.Claims, crumb.SetClaim{ID: fields[1], Claim: c, At: at})
	return ""
}

// split returns the fields of the record in line that its kind needs,
// the kind first; need gives their number for each kind to read. It
// returns no fields for a record of another kind, which is skipped,
// and for a line that is not a record, with why it is not.
func split(line string, need map[string]int) (fields []string, why string) {
	fields = strings.Split(line, "\t")
	if fields[0] == "" {
		if line == "" {
			return nil, "it is empty"
		}
		return nil, "it has no kind"
	}
	n, ok := need[fields[0]]
	if !ok {
		return nil, ""
	}
	if len(fields) < n {
		return nil, fmt.Sprintf("a %s record has %d fields, and it has %d", fields[0], n, len(fields))
	}
	fields = fields[:n]
	for _, f := range fields {
		if f == "" {
			return nil, "it has an empty field"
		}
	}
	return fields, ""
}

// place returns the place a record's PATH and LINE fields give, or
// why LINE is not a line.
func place(path, line string) (crumb.Place, string) {
	n := lineNumber(line)
	if n == 0 {
		return crumb.Place{}, fmt.Sprintf("its LINE is %q, not a whole number from 1", line)
	}
	return crumb.Place{Text: path, Line: n}, ""
}

// lineNumber returns the whole number from 1 that s writes in digits,
// or 0 when it writes none.
func lineNumber(s string) int {
	if strings.Trim(s, "0123456789") != "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
