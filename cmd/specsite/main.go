// Command specsite builds the web site of the specification of
// Breadcrumb SDD from the tags of this repository (TD-0031), with the
// audit report of the latest one (TD-0037).
//
// Usage:
//
//	go run ./cmd/specsite [-o DIR]
//
// Run it from the root of a clone with every tag fetched. It writes the
// site into DIR, _site by default. It needs git, and Go to run the bcr
// of the latest tag. It is not part of the toolkit, and no release
// carries it.
package main

import (
	"archive/tar"
	"bytes"
	"errors"
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

	var report *specsite.Report
	if latest := specsite.LatestTag(strings.Fields(tags)); latest != "" {
		html, err := auditReport(latest)
		if err != nil {
			return fmt.Errorf("the audit report of %s: %v", latest, err)
		}
		report = &specsite.Report{Tag: latest, HTML: html}
	}

	files, err := specsite.Build(repo, releases, report)
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

// auditReport is the audit report of the repository at tag, written
// by that tag's own bcr on the files committed at it (TD-0037). A bcr
// that finds something wrong still writes it (RQ-0049).
func auditReport(tag string) ([]byte, error) {
	dir, err := os.MkdirTemp("", "specsite-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	cmd := exec.Command("git", "archive", "--format=tar", tag)
	archive, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := untar(archive, dir); err != nil {
		cmd.Wait()
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git archive %s: %v: %s", tag, err, strings.TrimSpace(errOut.String()))
	}

	out := filepath.Join(dir, "audit-report.html")
	run := exec.Command("go", "run", "./cmd/bcr", "audit-report", "--html", "-o", out)
	run.Dir = dir
	var runErr bytes.Buffer
	run.Stderr = &runErr
	var exit *exec.ExitError
	if err := run.Run(); err != nil && !(errors.As(err, &exit) && exit.ExitCode() == 1) {
		return nil, fmt.Errorf("bcr audit-report: %v: %s", err, strings.TrimSpace(runErr.String()))
	}
	return os.ReadFile(out)
}

// untar writes the files of the tar archive r under dir.
func untar(r io.Reader, dir string) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !filepath.IsLocal(h.Name) {
			return fmt.Errorf("the archive has a path outside it: %s", h.Name)
		}
		p := filepath.Join(dir, filepath.FromSlash(h.Name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
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
