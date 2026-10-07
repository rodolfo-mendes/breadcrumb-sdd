// Package target reads the file a claim is about, its target, from
// the working tree (ADR-0023). It is infrastructure: it knows the file
// system, and nothing about claims.
package target

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// Read reads the target at path, a path from the current directory
// with / between its parts. file says whether it is a regular file;
// content is then what it holds. A target that does not exist, or is
// anything else, such as a directory or a symbolic link, is not a
// file, and no error: a symbolic link is not followed. A regular file
// that cannot be read is an error.
func Read(path string) (content []byte, file bool, err error) {
	name := filepath.FromSlash(path)
	info, err := os.Lstat(name)
	switch {
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, syscall.ENOTDIR):
		// ENOTDIR: a part of the path before the last is a file, so
		// nothing can be at the path.
		return nil, false, nil
	case err != nil:
		return nil, false, err
	case !info.Mode().IsRegular():
		return nil, false, nil
	}
	content, err = os.ReadFile(name)
	if err != nil {
		return nil, false, err
	}
	return content, true, nil
}
