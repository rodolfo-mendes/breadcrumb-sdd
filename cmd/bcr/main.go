// Command bcr is the Breadcrumb toolkit.
//
// Usage:
//
//	bcr extract [FILE...]
//	bcr verify
//	bcr audit
//	bcr report
//	bcr filter [-i ID] [-o ID] [-t TYPE] [-k KIND]
//	bcr init [-a FILE]
//
// Run it from the root of a repository. extract prints the breadcrumb
// of each file it is given, or of each file .breadcrumbs names, as
// records, verify each problem in the set of breadcrumbs whose records
// it reads from standard input, audit the verdict of each claim and
// breadcrumb whose records it reads from standard input, and report a
// page of the records it reads from standard input. filter prints the
// records it reads from standard input that match its flags. init sets
// a repository up for the pipe. Its user documentation is docs/bcr.md.
//
// It exits 0 when it finds nothing wrong, 1 when it finds something
// wrong, and 2 when it is used wrongly or cannot run (ADR-0006).
package main

import (
	"fmt"
	"io"
	"os"
)

// The usage line of each command (ADR-0006), and of bcr.
const (
	extractUsage = "usage: bcr extract [FILE...]\n"
	verifyUsage  = "usage: bcr verify\n"
	auditUsage   = "usage: bcr audit\n"
	reportUsage  = "usage: bcr report\n"
	filterUsage  = "usage: bcr filter [-i ID] [-o ID] [-t TYPE] [-k KIND]\n"
	initUsage    = "usage: bcr init [-a FILE]\n"
	usage        = extractUsage + verifyUsage + auditUsage + reportUsage + filterUsage + initUsage
)

// Exit codes (ADR-0006).
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
	case "extract":
		return runExtract(args, stdout, stderr)
	case "verify":
		return runVerify(args, stdin, stdout, stderr)
	case "audit":
		return runAudit(args, stdin, stdout, stderr)
	case "report":
		return runReport(args, stdin, stdout, stderr)
	case "filter":
		return runFilter(args, stdin, stdout, stderr)
	case "init":
		return runInit(args, stdout, stderr)
	}
	return usageError(stderr, usage, fmt.Sprintf("unknown command %q", args[0]))
}

// usageError writes msg and the usage to stderr (ADR-0006).
func usageError(stderr io.Writer, usage, msg string) int {
	fmt.Fprintf(stderr, "bcr: %s\n%s", msg, usage)
	return exitTrouble
}

// trouble writes err to stderr (ADR-0006).
func trouble(stderr io.Writer, err error) int {
	fmt.Fprintln(stderr, "bcr:", err)
	return exitTrouble
}
