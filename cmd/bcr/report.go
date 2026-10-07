package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/page"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/records"
)

// runReport runs bcr report; args starts with the command. It reads
// the records of the pipe from stdin and writes one page to stdout:
// every breadcrumb, problem and claim, and one diagram for each spec
// (specs/report/spec.md). It does not pass the records on.
func runReport(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, reportUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, reportUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	// The whole input is read before anything is written: a line that
	// is not a record, however late, means no page.
	set, err := records.ReadReport(stdin)
	if err != nil {
		return trouble(stderr, err)
	}
	var buf bytes.Buffer
	if err := page.Write(&buf, set); err != nil {
		return trouble(stderr, err)
	}
	if _, err := stdout.Write(buf.Bytes()); err != nil {
		return trouble(stderr, err)
	}
	return exitOK
}
