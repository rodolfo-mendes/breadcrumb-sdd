// Command specsite builds the web site of the specification of
// Breadcrumb SDD from the tags of this repository (TD-0031).
//
// Usage:
//
//	go run ./cmd/specsite [-o DIR]
//
// Run it from the root of a clone with every tag fetched. It writes the
// site into DIR, _site by default. It is not part of the toolkit, and
// no release carries it.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/specsite"
)

const usage = "usage: specsite [-o DIR]\n"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	set, operands, err := cli.Parse(args, []cli.Flag{{Short: 'o', Long: "output", Value: true}})
	dir, ok := set["output"]
	if !ok {
		dir = "_site"
	}
	if err != nil || len(operands) > 0 || dir == "" {
		fmt.Fprint(stderr, usage)
		return 2
	}
	if err := build(dir, stdout); err != nil {
		fmt.Fprintf(stderr, "specsite: %v\n", err)
		return 2
	}
	return 0
}

// build writes the site into dir, and prints the path of each file.
func build(dir string, stdout io.Writer) error {
	mod, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}
	module, _, _ := strings.Cut(strings.TrimPrefix(string(mod), "module "), "\n")
	repo := "https://" + strings.TrimSpace(module)

	tags, err := git("tag", "--list", "v*")
	if err != nil {
		return err
	}
	var releases []specsite.Release
	for _, tag := range strings.Fields(tags) {
		// A tag with no specification has no page (RQ-0037).
		if _, err := git("cat-file", "-e", tag+":"+specsite.SpecPath); err != nil {
			continue
		}
		spec, err := git("show", tag+":"+specsite.SpecPath)
		if err != nil {
			return err
		}
		releases = append(releases, specsite.Release{Tag: tag, Spec: spec})
	}

	files, err := specsite.Build(repo, releases)
	if err != nil {
		return err
	}
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		out := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, files[p], 0o644); err != nil {
			return err
		}
		fmt.Fprintln(stdout, out)
	}
	return nil
}

// git runs git with args and returns what it writes to standard output.
func git(args ...string) (string, error) {
	var out, errOut bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}
