package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/records"
)

// runVerify runs bcr verify; args starts with the command. It reads
// the records bcr extract prints from stdin, copies them to stdout
// byte for byte (ADR-0016), and prints each problem of the set they
// describe as PATH:LINE: MESSAGE (ADR-0015), and after the copy as a
// record (ADR-0021).
func runVerify(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, verifyUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, verifyUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	// The whole input is read before anything is written: a line that
	// is not a record, however late, means nothing is passed on.
	input, err := io.ReadAll(stdin)
	if err != nil {
		return trouble(stderr, err)
	}
	set, err := records.Read(bytes.NewReader(input))
	if err != nil {
		return trouble(stderr, err)
	}
	if _, err := stdout.Write(input); err != nil {
		return trouble(stderr, err)
	}
	problems := set.Check()
	// A last line with no ending must not share its line with the
	// first record added.
	if len(problems) > 0 && len(input) > 0 && input[len(input)-1] != '\n' {
		if _, err := io.WriteString(stdout, "\n"); err != nil {
			return trouble(stderr, err)
		}
	}
	for _, p := range problems {
		fmt.Fprintf(stderr, "%s:%d: %s\n", p.At.Text, p.At.Line, p.Message)
		if err := records.WriteProblem(stdout, p.At.Text, p.At.Line, p.Message); err != nil {
			return trouble(stderr, err)
		}
	}
	if len(problems) > 0 {
		return exitFound
	}
	return exitOK
}
