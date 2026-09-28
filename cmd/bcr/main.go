// Command bcr is the Breadcrumb SDD toolkit.
//
// Usage:
//
//	bcr audit-report --html
//
// Run it from the root of a repository. It audits the breadcrumbs in
// breadcrumbs/ and writes the report to audit-report.html.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/breadcrumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/report"
)

const usage = "usage: bcr audit-report --html\n"

// reportFile is where the report is written, in the repository root.
const reportFile = "audit-report.html"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "audit-report" || args[1] != "--html" {
		fmt.Fprint(stderr, usage)
		return 2
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "bcr:", err)
		return 1
	}
	if info, err := os.Stat(filepath.Join(root, breadcrumb.Dir)); err != nil || !info.IsDir() {
		fmt.Fprintf(stderr, "bcr: no %s/ directory here; run bcr from the root of a repository\n", breadcrumb.Dir)
		return 1
	}

	g, err := breadcrumb.Load(os.DirFS(root))
	if err != nil {
		fmt.Fprintln(stderr, "bcr:", err)
		return 1
	}
	var buf bytes.Buffer
	if err := report.Write(&buf, g); err != nil {
		fmt.Fprintln(stderr, "bcr:", err)
		return 1
	}
	out := filepath.Join(root, reportFile)
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintln(stderr, "bcr:", err)
		return 1
	}
	fmt.Fprintln(stdout, out)
	return 0
}
