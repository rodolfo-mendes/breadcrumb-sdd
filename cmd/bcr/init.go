package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/rodolfo-mendes/breadcrumb-sdd/docs"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/breadcrumb"
	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/cli"
)

// version is the version bcr was released as, set by the release
// workflow with -ldflags "-X main.version=VERSION" (TD-0036).
var version string

// released is a release version, MAJOR.MINOR.PATCH.
var released = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// bcrVersion is the version bcr was released as, or "" when this build
// has none (TD-0036).
func bcrVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		return moduleVersion(info.Main.Version)
	}
	return ""
}

// moduleVersion is the release version in the version Go records for a
// build, such as v0.7.0, or "" for any other build.
func moduleVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if !released.MatchString(v) {
		return ""
	}
	return v
}

// specVersion is the version of the specification this bcr carries out,
// from its third line (RQ-0036).
func specVersion() string {
	lines := strings.SplitN(docs.Spec, "\n", 4)
	if len(lines) < 3 {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(lines[2]), "Version ")
}

// Where bcr init points to.
const (
	siteURL    = "https://rodolfo-mendes.github.io/breadcrumb-sdd/"
	releaseURL = "https://github.com/rodolfo-mendes/breadcrumb-sdd/releases/download/"
)

// The pieces bcr init sets up, each found by its name (TD-0034).
const (
	adoptionFile = "breadcrumbs/TD-0001.md"
	workflowFile = ".github/workflows/breadcrumbs.yml"
	agentsHead   = "## Breadcrumb SDD"
)

// agentsFiles are the files bcr init may give its section (RQ-0046).
var agentsFiles = []string{"AGENTS.md", "CLAUDE.md"}

// adoption is the first breadcrumb of a repository (RQ-0042).
const adoption = `# TD-0001: Adopt Breadcrumb SDD

## Decision

This repository adopts Breadcrumb SDD, following version {{spec}} of
its specification: <{{site}}spec/{{spec}}/>.

## Why

Breadcrumbs record why the repository changes, in a form that can be
checked against its code, so that drift between intent and code shows.

## What it beat
`

// agentsSection points agents to the breadcrumbs (RQ-0043).
const agentsSection = `## Breadcrumb SDD

This repository is developed with Breadcrumb SDD, following version
{{spec}} of its specification: <{{site}}spec/{{spec}}/>.
Its breadcrumbs live in ` + "`breadcrumbs/`" + `.

- Before changing the repository, read the breadcrumbs.
- Record each change as the specification says. Write Technical
  Decisions and Tasks; draft Intakes and Requirements for a person to
  approve.
- Before committing, run ` + "`bcr check`" + ` and ` + "`bcr verdict`" + `; neither should
  fail.
`

// workflow is the check on each change (RQ-0044, TD-0035). It runs the
// Linux x86-64 archive of the release that wrote it.
const workflow = `# Checks the breadcrumbs on each change with bcr {{bcr}}, set up by
# bcr init. See {{site}}
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

      - name: Check the breadcrumbs
        run: bcr check

      - name: Audit the breadcrumbs
        run: bcr verdict
`

// runInit runs bcr init; args starts with the command. It finds what
// it cannot do before it writes anything (RQ-0045).
func runInit(args []string, stdout, stderr io.Writer) int {
	set, operands, err := cli.Parse(args[1:], []cli.Flag{{Short: 'a', Long: "agents-file", Value: true}})
	if err != nil {
		return usageError(stderr, initUsage, err.Error())
	}
	if len(operands) > 0 {
		return usageError(stderr, initUsage, fmt.Sprintf("unexpected operand %q", operands[0]))
	}
	agents, ok := set["agents-file"]
	if !ok {
		agents = agentsFiles[0]
	}
	if agents != agentsFiles[0] && agents != agentsFiles[1] {
		return usageError(stderr, initUsage, fmt.Sprintf("-a is AGENTS.md or CLAUDE.md, not %q", agents))
	}
	bcr, spec := bcrVersion(), specVersion()
	if bcr == "" {
		return trouble(stderr, errors.New("this bcr has no version, so init cannot pin the check to it; install a release of bcr (TD-0036)"))
	}
	fill := strings.NewReplacer("{{bcr}}", bcr, "{{spec}}", spec, "{{site}}", siteURL, "{{release}}", releaseURL).Replace

	// What to write, found before anything is written.
	var writes []initWrite

	if info, err := os.Stat(breadcrumb.Dir); err == nil && !info.IsDir() {
		return trouble(stderr, fmt.Errorf("%s is not a directory", breadcrumb.Dir))
	} else if err == nil {
		g, err := breadcrumb.Load(os.DirFS("."))
		if err != nil {
			return trouble(stderr, err)
		}
		if len(g.Breadcrumbs) > 0 {
			fmt.Fprintf(stderr, "bcr: %s/ already has breadcrumbs; left as it is\n", breadcrumb.Dir)
		} else {
			writes = append(writes, initWrite{path: adoptionFile, text: fill(adoption)})
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		writes = append(writes, initWrite{path: adoptionFile, text: fill(adoption)})
	} else {
		return trouble(stderr, err)
	}

	switch existing, err := os.ReadFile(agents); {
	case errors.Is(err, fs.ErrNotExist):
		writes = append(writes, initWrite{path: agents, text: fill(agentsSection)})
	case err != nil:
		return trouble(stderr, err)
	case hasLine(string(existing), agentsHead):
		fmt.Fprintf(stderr, "bcr: %s already has a %s section; left as it is\n", agents, agentsHead)
	default:
		sep := ""
		if len(existing) > 0 {
			sep = "\n"
			if existing[len(existing)-1] != '\n' {
				sep = "\n\n"
			}
		}
		writes = append(writes, initWrite{path: agents, text: sep + fill(agentsSection), append: true})
	}

	switch _, err := os.Lstat(workflowFile); {
	case errors.Is(err, fs.ErrNotExist):
		writes = append(writes, initWrite{path: workflowFile, text: fill(workflow)})
	case err != nil:
		return trouble(stderr, err)
	default:
		fmt.Fprintf(stderr, "bcr: %s is already there; left as it is\n", workflowFile)
	}

	for _, w := range writes {
		if err := w.do(); err != nil {
			return trouble(stderr, err)
		}
		fmt.Fprintln(stdout, w.path)
	}
	return exitOK
}

// initWrite is a file bcr init creates, or adds text to the end of.
type initWrite struct {
	path   string
	text   string
	append bool
}

// do writes w. A file it creates must not be there (RQ-0045).
func (w initWrite) do() error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if w.append {
		flags = os.O_WRONLY | os.O_APPEND
	} else if err := os.MkdirAll(filepath.Dir(filepath.FromSlash(w.path)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.FromSlash(w.path), flags, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(w.text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// hasLine reports whether s has a line that is exactly l.
func hasLine(s, l string) bool {
	for _, x := range strings.Split(s, "\n") {
		if strings.TrimSuffix(x, "\r") == l {
			return true
		}
	}
	return false
}
