package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/fileset"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/frontmatter"
)

// runExtract runs bcr extract; args starts with the command. For each
// file, in the order given, it prints the records of its breadcrumb,
// or its problems (ADR-0011). Which files to read is the caller's
// choice: their names and paths are not checked. Given no file, it
// reads the files .breadcrumbs names (ADR-0014).
func runExtract(args []string, stdout, stderr io.Writer) int {
	_, operands, err := cli.Parse(args[1:], nil)
	if err != nil {
		return usageError(stderr, extractUsage, err.Error())
	}
	code := exitOK
	if len(operands) == 0 {
		if operands, code = namedFiles(stderr); code != exitOK && operands == nil {
			return code
		}
	}
	for _, name := range operands {
		src, err := os.ReadFile(filepath.FromSlash(name))
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

// namedFiles returns the files the .breadcrumbs of the current
// directory names, each as its path from there with / between its
// parts, in order of path. With no .breadcrumbs, or a problem in it,
// it returns no file and the exit status; when only a directory could
// not be read, the files found and exitTrouble.
func namedFiles(stderr io.Writer) ([]string, int) {
	src, err := os.ReadFile(fileset.Name)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(stderr, "bcr: no FILE given, and no %s in this directory\n", fileset.Name)
		return nil, exitTrouble
	}
	if err != nil {
		fmt.Fprintln(stderr, "bcr:", err)
		return nil, exitTrouble
	}
	set, problems := fileset.Parse(src)
	for _, p := range problems {
		fmt.Fprintf(stderr, "%s:%d: %s\n", fileset.Name, p.Line, p.Message)
	}
	if len(problems) > 0 {
		return nil, exitFound
	}
	files, errs := set.Files(os.DirFS("."))
	for _, err := range errs {
		fmt.Fprintln(stderr, "bcr:", err)
	}
	if files == nil {
		files = []string{}
	}
	if len(errs) > 0 {
		return files, exitTrouble
	}
	return files, exitOK
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
