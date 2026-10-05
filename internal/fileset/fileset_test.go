package fileset

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// set parses lines as a .breadcrumbs, and fails on any problem.
func set(t *testing.T, lines ...string) Set {
	t.Helper()
	s, ps := Parse([]byte(strings.Join(lines, "\n") + "\n"))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	return s
}

// tree writes each file, named by a path from a temporary directory,
// and returns the directory.
func tree(t *testing.T, names ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range names {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// found returns the files under root that s names, and fails on any
// error.
func found(t *testing.T, s Set, root string) []string {
	t.Helper()
	files, errs := s.Files(os.DirFS(root))
	if errs != nil {
		t.Fatalf("got errors %v", errs)
	}
	return files
}

func TestBlankLinesAndCommentsAreIgnored(t *testing.T) {
	s, ps := Parse([]byte("\xEF\xBB\xBF# the decisions\r\ndocs/adrs/*.md\r\n\r\n  \t\r\n#tasks/*.md\r\n"))
	if ps != nil {
		t.Fatalf("got problems %v", ps)
	}
	if !s.Names("docs/adrs/ADR-0001.md") || s.Names("tasks/PBI-00001.md") {
		t.Errorf("want docs/adrs/ADR-0001.md named and tasks/PBI-00001.md not")
	}
}

func TestAnEmptyFileNamesNothing(t *testing.T) {
	s, ps := Parse(nil)
	if ps != nil || s.Names("a.md") {
		t.Errorf("got problems %v, or a.md named", ps)
	}
}

func TestAPatternIsMatchedAgainstTheWholePath(t *testing.T) {
	for _, tc := range []struct {
		pattern, name string
		want          bool
	}{
		{"tasks/*.md", "tasks/PBI-00001.md", true},
		{"tasks/*.md", "tasks/old/PBI-00001.md", false},
		{"tasks/*.md", "x/tasks/PBI-00001.md", false},
		{"*.md", "README.md", true},
		{"*.md", "docs/bcr.md", false},
		{"specs/*/spec.md", "specs/extract/spec.md", true},
		{"specs/*/spec.md", "specs/spec.md", false},
		{"tasks/PBI-0000?.md", "tasks/PBI-00001.md", true},
		{"tasks/PBI-0000?.md", "tasks/PBI-00010.md", false},
		{"tasks?PBI.md", "tasks/PBI.md", false},
		{"docs/[a-c]*.md", "docs/bcr.md", true},
		{"docs/[a-c]*.md", "docs/spec.md", false},
		{"README.md", "README.md", true},
	} {
		if got := set(t, tc.pattern).Names(tc.name); got != tc.want {
			t.Errorf("%q names %q: %v, want %v", tc.pattern, tc.name, got, tc.want)
		}
	}
}

func TestTheLastLineThatMatchesDecides(t *testing.T) {
	s := set(t, "tasks/*.md", "!tasks/PBI-0000?.md", "tasks/PBI-00001.md")
	for name, want := range map[string]bool{
		"tasks/PBI-00001.md": true,
		"tasks/PBI-00002.md": false,
		"tasks/PBI-00010.md": true,
		"docs/bcr.md":        false,
	} {
		if got := s.Names(name); got != want {
			t.Errorf("names %q: %v, want %v", name, got, want)
		}
	}
}

func TestALineThatIsNotAllowedIsAProblem(t *testing.T) {
	src := strings.Join([]string{
		"docs/adrs/*.md",
		"docs/**/*.md",
		"tasks/",
		"/tasks/*.md",
		"!",
		"# a comment",
		"tasks/[a-.md",
		"!specs/**",
	}, "\n")
	s, ps := Parse([]byte(src))
	var lines []int
	for _, p := range ps {
		lines = append(lines, p.Line)
	}
	if want := []int{2, 3, 4, 5, 7, 8}; !reflect.DeepEqual(lines, want) {
		t.Errorf("problems at lines %v, want %v: %v", lines, want, ps)
	}
	if !reflect.DeepEqual(s, Set{}) {
		t.Errorf("got set %+v with problems, want none", s)
	}
}

func TestFilesAreInOrderOfPathAsBytes(t *testing.T) {
	root := tree(t, "a/x.md", "a-b/x.md", "a/B.md", "a/b.md", "a/b.txt", "c/d/x.md", "x.md")
	got := found(t, set(t, "*/*.md"), root)
	if want := []string{"a-b/x.md", "a/B.md", "a/b.md", "a/x.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestADirectoryIsNotAFile(t *testing.T) {
	root := tree(t, "docs/adrs/x.md", "docs/bcr.md")
	got := found(t, set(t, "docs/*"), root)
	if want := []string{"docs/bcr.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGitIsNotEntered(t *testing.T) {
	root := tree(t, ".git/x.md", "docs/.git/x.md", "docs/x.md")
	got := found(t, set(t, "*/*.md", "docs/[.]git/*.md"), root)
	if want := []string{"docs/x.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSymbolicLinksAreNotFollowed(t *testing.T) {
	root := tree(t, "docs/x.md", "other/y.md")
	if err := os.Symlink(filepath.Join(root, "docs", "x.md"), filepath.Join(root, "docs", "link.md")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	if err := os.Symlink(filepath.Join(root, "other"), filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	got := found(t, set(t, "*/*.md"), root)
	if want := []string{"docs/x.md", "other/y.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestADirectoryNoPatternReachesIsNotEntered(t *testing.T) {
	root := tree(t, "tasks/a.md", "vendor/x/a.md")
	if err := os.Chmod(filepath.Join(root, "vendor"), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(root, "vendor"), 0o755) })
	got := found(t, set(t, "tasks/*.md", "!vendor/*/*.md"), root)
	if want := []string{"tasks/a.md"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestADirectoryThatCannotBeReadIsAnError(t *testing.T) {
	root := tree(t, "tasks/a.md", "docs/a.md")
	if err := os.Chmod(filepath.Join(root, "docs"), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(filepath.Join(root, "docs"), 0o755) })
	if _, err := os.ReadDir(filepath.Join(root, "docs")); err == nil {
		t.Skip("every directory can be read here")
	}
	files, errs := set(t, "*/*.md").Files(os.DirFS(root))
	if want := []string{"tasks/a.md"}; !reflect.DeepEqual(files, want) || len(errs) != 1 {
		t.Errorf("got %v and errors %v, want %v and one error", files, errs, want)
	}
}
