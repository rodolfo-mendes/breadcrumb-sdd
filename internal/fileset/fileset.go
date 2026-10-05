// Package fileset reads .breadcrumbs, the file in which a repository
// names the files that carry its breadcrumbs (ADR-0014), and finds
// those files. It decides which files are read, not what a breadcrumb
// means: the core does not use it (ADR-0013).
package fileset

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Name is the name of the file, at the root of a repository.
const Name = ".breadcrumbs"

// Problem is a line .breadcrumbs does not allow.
type Problem struct {
	Line    int // counted from 1
	Message string
}

// Set is the files a .breadcrumbs names.
type Set struct {
	patterns []pattern // in the order written
}

// pattern is one line of .breadcrumbs.
type pattern struct {
	text    string   // without its !
	exclude bool     // written with !
	parts   []string // text split at /, or nil when a / may not be between two parts
}

// Parse reads src, the text of a .breadcrumbs. When a line is not
// allowed, it returns every problem, in order of line, and no set.
func Parse(src []byte) (Set, []Problem) {
	var s Set
	var problems []Problem
	text := strings.TrimPrefix(string(src), "\xEF\xBB\xBF")
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pat, exclude := strings.CutPrefix(line, "!")
		if why := refuse(pat); why != "" {
			problems = append(problems, Problem{Line: i + 1, Message: fmt.Sprintf("pattern %q %s", line, why)})
			continue
		}
		p := pattern{text: pat, exclude: exclude}
		if !strings.ContainsAny(pat, `[\`) {
			p.parts = strings.Split(pat, "/")
		}
		s.patterns = append(s.patterns, p)
	}
	if problems != nil {
		return Set{}, problems
	}
	return s, nil
}

// refuse returns why a pattern is not allowed, or "" when it is.
func refuse(text string) string {
	switch {
	case text == "":
		return "names no file"
	case strings.Contains(text, "**"):
		return "uses **; * matches within one part of a path"
	case strings.HasPrefix(text, "/"):
		return "starts with /; a pattern is matched from the root as it is"
	case strings.HasSuffix(text, "/"):
		return "ends with /; a pattern names files, not a directory"
	}
	if _, err := path.Match(text, ""); err != nil {
		return "is malformed"
	}
	return ""
}

// Names reports whether the set names the file at name, its whole path
// from the root, with / between its parts. The last line that matches
// decides; a file no line matches is not named.
func (s Set) Names(name string) bool {
	named := false
	for _, p := range s.patterns {
		if ok, _ := path.Match(p.text, name); ok {
			named = !p.exclude
		}
	}
	return named
}

// Files returns the files under the root of fsys that the set names,
// in order of path compared as bytes, and what could not be read on
// the way. Only regular files are returned, so a symbolic link is not
// followed, and a directory named .git is not entered.
func (s Set) Files(fsys fs.FS) ([]string, []error) {
	var files []string
	var errs []error
	fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			errs = append(errs, err)
		case name == ".":
		case d.IsDir():
			if d.Name() == ".git" || !s.reaches(name) {
				return fs.SkipDir
			}
		case d.Type().IsRegular() && s.Names(name):
			files = append(files, name)
		}
		return nil
	})
	sort.Strings(files)
	return files, errs
}

// reaches reports whether the set could name a file under dir: a part
// of a pattern matches one part of a path, so a pattern reaches only
// the directories its first parts match.
func (s Set) reaches(dir string) bool {
	n := strings.Count(dir, "/") + 1
	for _, p := range s.patterns {
		switch {
		case p.exclude:
		case p.parts == nil:
			return true
		case len(p.parts) > n:
			if ok, _ := path.Match(strings.Join(p.parts[:n], "/"), dir); ok {
				return true
			}
		}
	}
	return false
}
