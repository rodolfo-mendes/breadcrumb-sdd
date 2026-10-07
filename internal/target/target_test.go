package target

import (
	"os"
	"path/filepath"
	"testing"
)

// in writes docs/bcr.md in a temporary directory, and runs from it.
func in(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "bcr.md"), []byte("### verify\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestReadARegularFile(t *testing.T) {
	in(t)
	content, file, err := Read("docs/bcr.md")
	if string(content) != "### verify\n" || !file || err != nil {
		t.Errorf("got %q, %v, %v; want the content, a file, no error", content, file, err)
	}
}

func TestReadWhatIsNotAFile(t *testing.T) {
	in(t)
	if err := os.Symlink("bcr.md", filepath.Join("docs", "link.md")); err != nil {
		t.Skip("no symbolic links here:", err)
	}
	for _, path := range []string{
		"docs/none.md",  // does not exist
		"none/bcr.md",   // nor does its directory
		"docs",          // a directory
		"docs/link.md",  // a symbolic link to a file
		"docs/bcr.md/x", // a path through a file
		"docs/dangling", // nothing at all
	} {
		content, file, err := Read(path)
		if content != nil || file || err != nil {
			t.Errorf("%s: got %q, %v, %v; want no content, not a file, no error", path, content, file, err)
		}
	}
}

func TestReadAFileThatCannotBeRead(t *testing.T) {
	in(t)
	name := filepath.Join("docs", "bcr.md")
	if err := os.Chmod(name, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(name); err == nil {
		t.Skip("this user can read a file with no permissions")
	}
	content, file, err := Read("docs/bcr.md")
	if content != nil || file || err == nil {
		t.Errorf("got %q, %v, %v; want an error", content, file, err)
	}
}
