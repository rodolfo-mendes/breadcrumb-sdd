package main

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/records"
)

// runFilter runs bcr filter; args starts with the command. It reads
// the records of the pipe from stdin and writes to stdout those that
// match every flag given, as they were read, in the order read
// (specs/filter/spec.md). It judges nothing.
func runFilter(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	set, operands, err := cli.ParseAll(args[1:], []cli.Flag{
		{Short: 'i', Long: "id", Value: true},
		{Short: 'o', Long: "object", Value: true},
		{Short: 't', Long: "type", Value: true},
		{Short: 'k', Long: "kind", Value: true},
	})
	if err != nil {
		return usageError(stderr, filterUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, filterUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	if len(set) == 0 {
		return usageError(stderr, filterUsage, "no flag")
	}
	// The whole input is read before anything is written: a line that
	// is not a record, however late, means nothing is printed.
	lines, err := records.ReadLines(stdin)
	if err != nil {
		return trouble(stderr, err)
	}
	// A flag not given sets no condition. A record with no ID, TYPE or
	// OBJECT field has "" there, which no value of a flag selects.
	matches := func(flag, field string) bool {
		values, given := set[flag]
		return !given || field != "" && slices.Contains(values, field)
	}
	var buf bytes.Buffer
	for _, l := range lines {
		if matches("id", l.ID) && matches("object", l.Object) && matches("type", l.Type) && matches("kind", l.Kind) {
			buf.WriteString(l.Text)
		}
	}
	if _, err := stdout.Write(buf.Bytes()); err != nil {
		return trouble(stderr, err)
	}
	return exitOK
}
