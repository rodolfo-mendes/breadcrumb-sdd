package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/records"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/rules"
)

// runVerify runs bcr verify; args starts with the command. It reads
// the records bcr extract prints from stdin, copies them to stdout
// byte for byte (ADR-0016), and prints each problem of the set they
// describe as PATH:LINE: MESSAGE (ADR-0015), and after the copy as a
// record (ADR-0021). When the current directory has a breadcrumb.rules,
// the set is also checked against the shape it declares (ADR-0025).
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
	shape, declared, ruleProblems, err := readRules()
	if err != nil {
		return trouble(stderr, err)
	}
	if _, err := stdout.Write(input); err != nil {
		return trouble(stderr, err)
	}
	problems := append(set.Check(), ruleProblems...)
	if declared {
		problems = append(problems, set.CheckShape(shape)...)
	}
	crumb.SortProblems(problems)
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

// readRules reads the breadcrumb.rules of the current directory.
// declared reports whether it gives a shape to check: with no file,
// or a problem in it, there is none (ADR-0025). A problem in the file
// is returned at its line in breadcrumb.rules.
func readRules() (shape crumb.Rules, declared bool, problems []crumb.SetProblem, err error) {
	src, err := os.ReadFile(rules.Name)
	if errors.Is(err, fs.ErrNotExist) {
		return crumb.Rules{}, false, nil, nil
	}
	if err != nil {
		return crumb.Rules{}, false, nil, err
	}
	shape, found := rules.Parse(src)
	if found == nil {
		found = shape.Check()
	}
	for _, p := range found {
		problems = append(problems, crumb.SetProblem{At: crumb.Place{Text: rules.Name, Line: p.Line}, Message: p.Message})
	}
	return shape, len(found) == 0, problems, nil
}
