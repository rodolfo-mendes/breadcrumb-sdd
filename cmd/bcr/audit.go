package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/records"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/target"
)

// runAudit runs bcr audit; args starts with the command. It reads the
// records bcr extract and bcr verify print from stdin, checks each
// claim against its target in the working tree, copies the records to
// stdout byte for byte, and adds the verdict of each claim and of each
// breadcrumb (ADR-0022, ADR-0023). Each Refuted claim is printed as
// PATH:LINE: MESSAGE.
func runAudit(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, auditUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, auditUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	// The whole input and every target are read before anything is
	// written: a line that is not a record, or a target that cannot
	// be read, means nothing is passed on.
	input, err := io.ReadAll(stdin)
	if err != nil {
		return trouble(stderr, err)
	}
	set, err := records.ReadAudit(bytes.NewReader(input))
	if err != nil {
		return trouble(stderr, err)
	}
	// Each target is read once, so every claim about a file is judged
	// on the same content.
	targets := map[string]crumb.Target{}
	for _, path := range set.Targets() {
		content, file, err := target.Read(path)
		if err != nil {
			return trouble(stderr, err)
		}
		targets[path] = crumb.Target{File: file, Content: content}
	}
	claims, breadcrumbs := set.Audit(targets)

	if _, err := stdout.Write(input); err != nil {
		return trouble(stderr, err)
	}
	// A last line with no ending must not share its line with the
	// first record added.
	if len(claims)+len(breadcrumbs) > 0 && len(input) > 0 && input[len(input)-1] != '\n' {
		if _, err := io.WriteString(stdout, "\n"); err != nil {
			return trouble(stderr, err)
		}
	}
	code := exitOK
	for i, c := range set.Claims {
		if claims[i] == crumb.Refuted {
			fmt.Fprintf(stderr, "%s:%d: claim is Refuted: %s\n", c.At.Text, c.At.Line, c.Claim.Lacks(targets[c.Claim.Target].File))
			code = exitFound
		}
		if err := records.WriteClaimVerdict(stdout, c.ID, claims[i], c.At); err != nil {
			return trouble(stderr, err)
		}
	}
	for i, b := range set.Breadcrumbs {
		if err := records.WriteVerdict(stdout, b.ID, breadcrumbs[i], b.At); err != nil {
			return trouble(stderr, err)
		}
	}
	return code
}
