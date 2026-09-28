// Command bcr is the Breadcrumb SDD toolkit.
//
// Usage:
//
//	bcr audit-report --html [-o FILE]
//	bcr check
//	bcr verdict [ID|PATH|-]...
//
// Run it from the root of a repository. audit-report audits the
// breadcrumbs in breadcrumbs/ and writes the report to
// audit-report.html, to FILE, or to standard output when FILE is -.
// check prints each problem in the breadcrumbs, and verdict the verdict
// of each breadcrumb. Its contract is docs/bcr.md.
//
// It exits 0 when it finds nothing wrong, 1 when it finds something
// wrong, and 2 when it is used wrongly or cannot run (RQ-0010).
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

// The usage line of each command (TD-0020), and of bcr.
const (
	auditReportUsage = "usage: bcr audit-report --html [-o FILE]\n"
	checkUsage       = "usage: bcr check\n"
	verdictUsage     = "usage: bcr verdict [ID|PATH|-]...\n"
	usage            = auditReportUsage + checkUsage + verdictUsage
)

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
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, usage, "no command")
	}
	switch args[0] {
	case "audit-report":
		return runAuditReport(args, stdout, stderr)
	case "check":
		return runCheck(args, stdout, stderr)
	case "verdict":
		return runVerdict(args, stdin, stdout, stderr)
	}
	return usageError(stderr, usage, fmt.Sprintf("unknown command %q", args[0]))
}

// runAuditReport runs bcr audit-report; args starts with the command.
func runAuditReport(args []string, stdout, stderr io.Writer) int {
	set, operands, err := cli.Parse(args[1:], []cli.Flag{{Long: "html"}, {Short: 'o', Long: "output", Value: true}})
	if err != nil {
		return usageError(stderr, auditReportUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, auditReportUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	if _, ok := set["html"]; !ok {
		return usageError(stderr, auditReportUsage, "audit-report needs --html")
	}
	out, ok := set["output"]
	if !ok {
		out = reportFile
	}
	if out == "" {
		return usageError(stderr, auditReportUsage, "-o needs a file name, or - for standard output")
	}
	return auditReport(out, stdout, stderr)
}

// runCheck runs bcr check; args starts with the command.
func runCheck(args []string, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, checkUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, checkUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	return check(stdout, stderr)
}

// runVerdict runs bcr verdict; args starts with the command.
func runVerdict(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, verdictUsage, err.Error())
	}
	names, err := cli.IDs(operands, stdin)
	if err != nil {
		return trouble(stderr, err)
	}
	return verdict(names, len(operands) == 0, stdout, stderr)
}

// verdict prints ID<TAB>VERDICT for each breadcrumb named, or for every
// breadcrumb when all is set (RQ-0024, RQ-0025). It exits 1 when one it
// prints is Refuted (RQ-0026), and 2 when a name names no breadcrumb.
func verdict(names []string, all bool, stdout, stderr io.Writer) int {
	g, _, err := load()
	if err != nil {
		return trouble(stderr, err)
	}
	var crumbs []breadcrumb.Breadcrumb
	unknown := false
	if all {
		crumbs = g.Breadcrumbs
	}
	for _, n := range names {
		b, ok := g.Find(n)
		if !ok {
			fmt.Fprintf(stderr, "bcr: %s names no breadcrumb\n", n)
			unknown = true
			continue
		}
		crumbs = append(crumbs, b)
	}
	refuted := false
	for _, b := range crumbs {
		fmt.Fprintf(stdout, "%s\t%s\n", b.ID, b.Verdict)
		refuted = refuted || b.Verdict == breadcrumb.Refuted
	}
	switch {
	case unknown:
		return exitTrouble
	case refuted:
		return exitFound
	}
	return exitOK
}

// load reads the breadcrumbs of the repository in the current directory.
func load() (breadcrumb.Graph, string, error) {
	root, err := os.Getwd()
	if err != nil {
		return breadcrumb.Graph{}, "", err
	}
	if info, err := os.Stat(filepath.Join(root, breadcrumb.Dir)); err != nil || !info.IsDir() {
		return breadcrumb.Graph{}, "", fmt.Errorf("no %s/ directory here; run bcr from the root of a repository", breadcrumb.Dir)
	}
	g, err := breadcrumb.Load(os.DirFS(root))
	return g, root, err
}

// check prints each problem in the breadcrumbs as PATH:LINE: MESSAGE
// (RQ-0020, TD-0020), and warns of each file in breadcrumbs/ that is not
// a breadcrumb (RQ-0023).
func check(stdout, stderr io.Writer) int {
	g, _, err := load()
	if err != nil {
		return trouble(stderr, err)
	}
	for _, o := range g.Others {
		fmt.Fprintf(stderr, "bcr: warning: %s is not a breadcrumb (TD-0013)\n", o)
	}
	found := false
	for _, b := range g.Breadcrumbs {
		for _, p := range b.Problems {
			fmt.Fprintf(stdout, "%s:%d: %s\n", b.Path, p.Line, p.Message)
			found = true
		}
	}
	if found {
		return exitFound // RQ-0021
	}
	return exitOK
}

// auditReport writes the report to the file out, or to stdout when out
// is - (RQ-0017, RQ-0018).
func auditReport(out string, stdout, stderr io.Writer) int {
	g, root, err := load()
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

// usageError writes msg and the usage to stderr (TD-0020).
func usageError(stderr io.Writer, usage, msg string) int {
	fmt.Fprintf(stderr, "bcr: %s\n%s", msg, usage)
	return exitTrouble
}

// trouble writes err to stderr (TD-0020).
func trouble(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "bcr:", err)
	return exitTrouble
}
