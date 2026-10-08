package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/rules"
)

// version is the release version bcr was built as, set by the release
// workflow with -ldflags "-X main.version=VERSION".
var version string

// released is a release version, MAJOR.MINOR.PATCH.
var released = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// bcrVersion is the release version of this bcr, or "" when it was
// built with none, such as with go run (specs/init/spec.md).
func bcrVersion() string {
	if version != "" {
		return moduleVersion(version)
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		return moduleVersion(info.Main.Version)
	}
	return ""
}

// moduleVersion is the release version in v, such as 0.8.0 for v0.8.0,
// or "" when v is not one.
func moduleVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if !released.MatchString(v) {
		return ""
	}
	return v
}

// The pieces bcr init sets up, each found by its name.
const (
	patternsFile  = ".breadcrumbs"
	workflowFile  = ".github/workflows/breadcrumbs.yml"
	defaultAgents = "AGENTS.md"
	agentsHead    = "## Breadcrumbs"
)

// releaseURL is where the workflow downloads the release of bcr from.
const releaseURL = "https://github.com/rodolfo-mendes/breadcrumb-sdd/releases/download/"

// patterns is the .breadcrumbs bcr init writes: comments only, so that
// the owner chooses the files.
const patterns = `# The files bcr extract reads, one pattern per line, from the root of
# the repository. * matches any characters but /, ? one character but
# /, and [...] one character of a set. A pattern that starts with !
# excludes the files it matches; when several match, the last decides.
# Blank lines and lines that start with # are ignored.
#
# For example, to read the decisions, the specs and the tasks:
#
# docs/adrs/*.md
# specs/*/spec.md
# tasks/*.md
`

// agentsSection points agents to the breadcrumbs.
const agentsSection = agentsHead + `

This repository records the intent of its documents in breadcrumbs:
YAML front matter, under the ` + "`breadcrumb`" + ` key, in the files
` + "`.breadcrumbs`" + ` names. ` + "`bcr`" + ` checks their links and their claims.

- When you add a file ` + "`.breadcrumbs`" + ` names, give it a breadcrumb.
- Before committing, run
  ` + "`set -o pipefail; bcr extract | bcr verify | bcr audit > /dev/null`" + `.
  It prints each problem and each Refuted claim, and should print
  nothing.
- Never edit a claim to turn a Refuted verdict green. Fix the code,
  or propose the change to the document that holds the claim.

To look breadcrumbs up, read their records rather than search the
files:

- One breadcrumb, with its links and claims:
  ` + "`bcr extract | bcr filter --id ID`" + `
- Every breadcrumb of a type: ` + "`bcr extract | bcr filter --type TYPE`" + `
- What links to a breadcrumb: ` + "`bcr extract | bcr filter --object ID`" + `
- A breadcrumb's verdict:
  ` + "`bcr extract | bcr verify | bcr audit | bcr filter --kind verdict --id ID`" + `
`

// workflow checks the breadcrumbs on each change with the release of
// bcr that wrote it.
const workflow = `# Checks the breadcrumbs on each change with bcr {{bcr}}, set up by
# bcr init. The page bcr report writes goes to the job's summary and is
# kept as an artifact of the run, whether the job fails or not.
name: Breadcrumbs

on:
  push:
  pull_request:

permissions:
  contents: read

jobs:
  breadcrumbs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1

      - name: Install bcr {{bcr}}
        run: |
          cd "$RUNNER_TEMP"
          curl -fsSLO "{{release}}v{{bcr}}/bcr_{{bcr}}_linux_amd64.tar.gz"
          curl -fsSLO "{{release}}v{{bcr}}/bcr_{{bcr}}_checksums.txt"
          sha256sum --check --ignore-missing "bcr_{{bcr}}_checksums.txt"
          tar -xzf "bcr_{{bcr}}_linux_amd64.tar.gz"
          echo "$RUNNER_TEMP/bcr_{{bcr}}_linux_amd64" >> "$GITHUB_PATH"

      - name: Verify, audit and report breadcrumbs
        run: |
          set -o pipefail
          bcr extract | bcr verify | bcr audit | bcr report > "$RUNNER_TEMP/report.md"

      - name: Show the report
        if: ${{ !cancelled() }}
        run: |
          if [ -s "$RUNNER_TEMP/report.md" ]; then
            cat "$RUNNER_TEMP/report.md" >> "$GITHUB_STEP_SUMMARY"
          fi

      - name: Keep the report
        if: ${{ !cancelled() }}
        uses: actions/upload-artifact@cf430e030ddbb5b0abf93d22962f4752f3646cd9 # v7.0.2
        with:
          path: ${{ runner.temp }}/report.md
          archive: false
          if-no-files-found: ignore
`

// runInit runs bcr init in the current directory; args starts with the
// command (specs/init/spec.md).
func runInit(args []string, stdout, stderr io.Writer) int {
	return initIn(".", bcrVersion(), args, stdout, stderr)
}

// initIn runs bcr init in the directory root, as a bcr released as
// release, or with no version when release is "". It finds every piece
// to write before it writes the first.
func initIn(root, release string, args []string, stdout, stderr io.Writer) int {
	set, operands, err := cli.Parse(args[1:], []cli.Flag{
		{Short: 'a', Long: "agents-file", Value: true},
		{Short: 'l', Long: "layout", Value: true},
	})
	if err != nil {
		return usageError(stderr, initUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, initUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	agents, ok := set["agents-file"]
	if !ok {
		agents = defaultAgents
	}
	if !inRepository(agents) {
		return usageError(stderr, initUsage, fmt.Sprintf("the agents file %q is not a path in the repository", agents))
	}
	name, withLayout := set["layout"]
	lay, known := layouts[name]
	if withLayout && !known {
		return usageError(stderr, initUsage, fmt.Sprintf("unknown layout %q; the layouts are: %s", name, strings.Join(layoutNames(), ", ")))
	}
	if release == "" {
		return trouble(stderr, errors.New("this bcr has no release version, so the workflow cannot install it; run a released bcr"))
	}
	fill := strings.NewReplacer("{{bcr}}", release, "{{release}}", releaseURL).Replace

	// The first line of each file that declares what bcr reads names
	// what wrote it (ADR-0027).
	header := "# bcr init, bcr " + release + "\n"
	pieces := []initWrite{{path: patternsFile, text: header + patterns}}
	sections := []initSection{{head: agentsHead, text: agentsSection}}
	if withLayout {
		// A layout's pieces work only together: both files must be
		// absent, so that neither is left from another setup.
		refused := false
		for _, p := range []string{patternsFile, rules.Name} {
			if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(p))); !errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(stderr, "bcr: %s is already there; --layout sets up %s and %s together: delete both to set the layout up again\n", p, patternsFile, rules.Name)
				refused = true
			}
		}
		if refused {
			return exitTrouble
		}
		header = "# bcr init --layout " + name + ", bcr " + release + "\n"
		pieces = []initWrite{
			{path: patternsFile, text: header + lay.patterns},
			{path: rules.Name, text: header + lay.rules},
		}
		pieces = append(pieces, lay.templates...)
		sections = append(sections, lay.sections...)
	}
	pieces = append(pieces,
		initWrite{path: agents, sections: sections},
		initWrite{path: workflowFile, text: fill(workflow)},
	)

	// What to write, found before anything is written.
	var writes []initWrite
	for _, p := range pieces {
		w, there, present, err := plan(root, p)
		if err != nil {
			return trouble(stderr, err)
		}
		for _, head := range present {
			fmt.Fprintf(stderr, "bcr: %s already has a %s section; left as it is\n", p.path, head)
		}
		if there {
			if p.sections == nil {
				fmt.Fprintf(stderr, "bcr: %s is already there; left as it is\n", p.path)
			}
			continue
		}
		writes = append(writes, w)
	}

	for _, w := range writes {
		if err := w.do(root); err != nil {
			return trouble(stderr, err)
		}
		fmt.Fprintln(stdout, w.path)
	}
	return exitOK
}

// inRepository reports whether p is a path from the root of the
// repository that stays in it: not empty, not starting with /, and with
// no part that is "..".
func inRepository(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") || filepath.IsAbs(p) {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(p), "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

// initWrite is a file bcr init creates, or, for the agents file, adds
// sections to the end of.
type initWrite struct {
	path     string
	text     string
	sections []initSection // the agents file's sections, each found by its heading
	append   bool          // add text to the end of an existing file
}

// initSection is a section of the agents file.
type initSection struct {
	head string // its heading, a line of its own
	text string // the section, its heading first
}

// plan finds what to do for p under root: the write to make, or there
// when the piece is already in place. For the agents file, present
// names the heading of each section already in it.
func plan(root string, p initWrite) (w initWrite, there bool, present []string, err error) {
	name := filepath.Join(root, filepath.FromSlash(p.path))
	var texts []string
	for _, s := range p.sections {
		texts = append(texts, s.text)
	}
	info, err := os.Lstat(name)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if p.sections != nil {
			p.text = strings.Join(texts, "\n")
		}
		return p, false, nil, nil
	case err != nil:
		return p, false, nil, err
	case p.sections == nil:
		return p, true, nil, nil
	case !info.Mode().IsRegular():
		return p, false, nil, fmt.Errorf("%s is not a regular file", p.path)
	}
	existing, err := os.ReadFile(name)
	if err != nil {
		return p, false, nil, err
	}
	var missing []string
	for _, s := range p.sections {
		if hasLine(string(existing), s.head) {
			present = append(present, s.head)
		} else {
			missing = append(missing, s.text)
		}
	}
	if len(missing) == 0 {
		return p, true, present, nil
	}
	sep := ""
	if len(existing) > 0 {
		sep = "\n"
		if existing[len(existing)-1] != '\n' {
			sep = "\n\n"
		}
	}
	p.text = sep + strings.Join(missing, "\n")
	p.append = true
	return p, false, present, nil
}

// do writes w under root. A file it creates must not be there.
func (w initWrite) do(root string) error {
	name := filepath.Join(root, filepath.FromSlash(w.path))
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if w.append {
		flags = os.O_WRONLY | os.O_APPEND
	} else if dir := path.Dir(filepath.ToSlash(w.path)); dir != "." {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(name, flags, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(w.text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// hasLine reports whether s has a line that is exactly l, without its
// line ending.
func hasLine(s, l string) bool {
	for _, x := range strings.Split(s, "\n") {
		if strings.TrimSuffix(x, "\r") == l {
			return true
		}
	}
	return false
}
