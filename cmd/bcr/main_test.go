package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestAnythingElseIsAUsageError(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		usage string
	}{
		{nil, usage},
		{[]string{"help"}, usage},
		{[]string{"extract", "-x", "a.md"}, extractUsage},
		// The commands ADR-0024 removed.
		{[]string{"check"}, usage},
		{[]string{"verdict"}, usage},
		{[]string{"links"}, usage},
		{[]string{"new", "PBI", "A title"}, usage},
		{[]string{"spec"}, usage},
		// bcr init came back, with its own usage (specs/init/spec.md).
		{[]string{"init", "-x"}, initUsage},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, nil, &stdout, &stderr); code != 2 {
			t.Errorf("%q: exit %d, want 2", tc.args, code)
		}
		if stdout.Len() > 0 {
			t.Errorf("%q: wrote %q to standard output", tc.args, stdout.String())
		}
		if !strings.HasPrefix(stderr.String(), "bcr: ") || !strings.HasSuffix(stderr.String(), "\n"+tc.usage) {
			t.Errorf("%q: wrote %q to standard error", tc.args, stderr.String())
		}
	}
}
