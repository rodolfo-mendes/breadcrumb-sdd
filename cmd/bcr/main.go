// Command bcr is the Breadcrumb SDD toolkit.
//
// Usage:
//
//	bcr audit-report --html [-o FILE]
//
// Run it from the root of a repository. It audits the breadcrumbs in
// breadcrumbs/ and writes the report to audit-report.html, to FILE, or
// to standard output when FILE is -. Its contract is docs/bcr.md.
//
// It exits 0 when it finds nothing wrong, 1 when a breadcrumb is Refuted
// or has a problem, and 2 when it is used wrongly or cannot run (RQ-0010).
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/breadcrumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/report"
)

const usage = "usage: bcr audit-report --html [-o FILE]\n"

// reportFile is where the report is written without -o, in the
// repository root (RQ-0019).
const reportFile = "audit-report.html"

// Exit codes (RQ-0010).
const (
	exitOK      = 0 // succeeded and found nothing wrong
	exitFound   = 1 // found something wrong in the breadcrumbs
	exitTrouble = 2 // used wrongly, or could not run
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, "no command")
	}
	if args[0] != "audit-report" {
		return usageError(stderr, fmt.Sprintf("unknown command %q", args[0]))
	}
	set, operands, err := cli.Parse(args[1:], []cli.Flag{{Long: "html"}, {Short: 'o', Long: "output", Value: true}})
	if err != nil {
		return usageError(stderr, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	if _, ok := set["html"]; !ok {
		return usageError(stderr, "audit-report needs --html")
	}
	out, ok := set["output"]
	if !ok {
		out = reportFile
	}
	if out == "" {
		return usageError(stderr, "-o needs a file name, or - for standard output")
	}
	return auditReport(out, stdout, stderr)
}

// auditReport writes the report to the file out, or to stdout when out
// is - (RQ-0017, RQ-0018).
func auditReport(out string, stdout, stderr io.Writer) int {
	root, err := os.Getwd()
	if err != nil {
		return trouble(stderr, err)
	}
	if info, err := os.Stat(filepath.Join(root, breadcrumb.Dir)); err != nil || !info.IsDir() {
		return trouble(stderr, fmt.Errorf("no %s/ directory here; run bcr from the root of a repository", breadcrumb.Dir))
	}

	g, err := breadcrumb.Load(os.DirFS(root))
	if err != nil {
		return trouble(stderr, err)
	}
	var buf bytes.Buffer
	if err := report.Write(&buf, g); err != nil {
		return trouble(stderr, err)
	}
	if out == "-" {
		if _, err := stdout.Write(buf.Bytes()); err != nil {
			return trouble(stderr, err)
		}
	} else {
		if !filepath.IsAbs(out) {
			out = filepath.Join(root, out)
		}
		if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
			return trouble(stderr, err)
		}
		fmt.Fprintln(stdout, out)
	}
	if foundWrong(g) {
		return exitFound
	}
	return exitOK
}

// foundWrong reports whether a breadcrumb is Refuted or has a problem.
func foundWrong(g breadcrumb.Graph) bool {
	for _, b := range g.Breadcrumbs {
		if b.Verdict == breadcrumb.Refuted || len(b.Problems) > 0 {
			return true
		}
	}
	return false
}

// usageError writes msg and the usage line to stderr (TD-0020).
func usageError(stderr io.Writer, msg string) int {
	fmt.Fprintf(stderr, "bcr: %s\n%s", msg, usage)
	return exitTrouble
}

// trouble writes err to stderr (TD-0020).
func trouble(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "bcr:", err)
	return exitTrouble
}
