package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/frontmatter"
)

// runExtract runs bcr extract; args starts with the command. For each
// file, in the order given, it prints the records of its breadcrumb,
// or its problems (ADR-0011). Which files to read is the caller's
// choice: their names and paths are not checked.
func runExtract(args []string, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, extractUsage, err.Error())
	}
	if len(operands) == 0 {
		return usageError(stderr, extractUsage, "extract needs a FILE")
	}
	code := exitOK
	for _, name := range operands {
		src, err := os.ReadFile(name)
		if err != nil {
			fmt.Fprintln(stderr, "bcr:", err)
			code = exitTrouble
			continue
		}
		if !extract(name, src, stdout, stderr) && code == exitOK {
			code = exitFound
		}
	}
	return code
}

// extract prints the records of the breadcrumb in src, the text of the
// file at path, or its problems as PATH:LINE: MESSAGE. It reports
// whether the file has no problem.
func extract(path string, src []byte, stdout, stderr io.Writer) bool {
	fm, found, problems := frontmatter.Read(src)
	if !found {
		return true
	}
	var b crumb.Breadcrumb
	if problems == nil {
		b, problems = crumb.Read(fm.Line, fm.Properties)
	}
	for _, p := range problems {
		fmt.Fprintf(stderr, "%s:%d: %s\n", path, p.Line, p.Message)
	}
	if len(problems) > 0 {
		return false
	}
	fmt.Fprintf(stdout, "breadcrumb\t%s\t%s\t%s\t%d\n", b.ID, b.Type, path, b.Line)
	for _, l := range b.Links {
		fmt.Fprintf(stdout, "link\t%s\t%s\t%s\t%s\t%d\n", b.ID, l.Verb, l.Object, path, l.Line)
	}
	return true
}
