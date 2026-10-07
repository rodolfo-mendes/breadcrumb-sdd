// Package records writes and reads the records bcr extract prints:
// one line for each breadcrumb, each link, each claim and each
// problem, its fields separated by a tab (ADR-0011, ADR-0021). It is
// the only package that knows this format; the core does not use it
// (ADR-0013).
//
//	breadcrumb	ID	TYPE	PATH	LINE
//	link	ID	VERB	OBJECT	PATH	LINE
//	claim	ID	KIND	TARGET	ARGUMENT	PATH	LINE
//	problem	PATH	LINE	MESSAGE
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
	claimKind        = "claim"   // written, and skipped by Read
	problemKind      = "problem" // written, and skipped by Read
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

// Read reads records from r until it ends, and returns the set of
// breadcrumbs they describe, each breadcrumb and each link at the
// path and line of its record. A line may end in \n or \r\n. A record
// of a kind Read does not know is skipped. A line that is not a
// record, an empty one included, is an error that gives its number in
// the input, counted from 1.
func Read(r io.Reader) (crumb.Set, error) {
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
		if why := add(&set, line); why != "" {
			return crumb.Set{}, fmt.Errorf("line %d of the input is not a record: %s", n, why)
		}
	}
}

// add adds the record in line to set, or returns why line is not a
// record.
func add(set *crumb.Set, line string) string {
	fields := strings.Split(line, "\t")
	var need int
	switch fields[0] {
	case "":
		if line == "" {
			return "it is empty"
		}
		return "it has no kind"
	case breadcrumbKind:
		need = breadcrumbFields
	case linkKind:
		need = linkFields
	default:
		return ""
	}
	if len(fields) < need {
		return fmt.Sprintf("a %s record has %d fields, and it has %d", fields[0], need, len(fields))
	}
	fields = fields[:need]
	for _, f := range fields {
		if f == "" {
			return "it has an empty field"
		}
	}
	at := crumb.Place{Text: fields[need-2]}
	at.Line = lineNumber(fields[need-1])
	if at.Line == 0 {
		return fmt.Sprintf("its LINE is %q, not a whole number from 1", fields[need-1])
	}
	if fields[0] == breadcrumbKind {
		set.Breadcrumbs = append(set.Breadcrumbs, crumb.SetBreadcrumb{ID: fields[1], At: at})
	} else {
		set.Links = append(set.Links, crumb.SetLink{Object: fields[3], At: at})
	}
	return ""
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
