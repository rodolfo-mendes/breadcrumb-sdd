package manpage

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "write docs/bcr.1 again from docs/bcr.md")

// TestTheManPageIsUpToDate fails when docs/bcr.1 is not what docs/bcr.md
// generates (TD-0023).
func TestTheManPageIsUpToDate(t *testing.T) {
	md, err := os.ReadFile("../../docs/bcr.md")
	if err != nil {
		t.Fatal(err)
	}
	want, err := Convert(string(md))
	if err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.WriteFile("../../docs/bcr.1", []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile("../../docs/bcr.1")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Error("docs/bcr.1 is not what docs/bcr.md generates; run go test ./internal/manpage -update")
	}
}

func TestConvert(t *testing.T) {
	md := strings.Join([]string{
		"# bcr(1)",
		"",
		"## Name",
		"",
		"bcr - the toolkit,",
		"on two lines.",
		"",
		"### audit-report",
		"",
		"- `--html`: write",
		"  HTML.",
		"- see [TD-0020](../breadcrumbs/TD-0020.md) and [spec](https://example.com/a-b)",
		"",
		"```",
		"bcr audit-report --html",
		".not a request",
		"```",
		"",
		`.a line with a back\slash`,
	}, "\n")
	want := strings.Join([]string{
		`.TH "BCR" "1" "" "" "Breadcrumb SDD"`,
		`.SH "NAME"`,
		`.PP`,
		`bcr \- the toolkit, on two lines.`,
		`.SS "audit\-report"`,
		`.IP \(bu 2`,
		`\fB\-\-html\fR: write HTML.`,
		`.IP \(bu 2`,
		`see TD\-0020 and spec <https://example.com/a\-b>`,
		`.PP`,
		`.RS 4`,
		`.nf`,
		`bcr audit\-report \-\-html`,
		`\&.not a request`,
		`.fi`,
		`.RE`,
		`.PP`,
		`\&.a line with a back\eslash`,
		``,
	}, "\n")
	got, err := Convert(md)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestConvertRefusesWhatItCannotShow(t *testing.T) {
	for _, tc := range []struct{ md, err string }{
		{"# bcr", "line 1: the first line is not a # name(section) heading"},
		{"# bcr(1)\n\n#### deep", "line 3: only ## and ### headings may follow the first"},
		{"# bcr(1)\n\n| a | b |", "line 3: tables and HTML cannot be converted"},
		{"# bcr(1)\n\n<b>x</b>", "line 3: tables and HTML cannot be converted"},
		{"# bcr(1)\n\n```\nopen", "line 4: a code block is not closed"},
	} {
		if _, err := Convert(tc.md); err == nil || err.Error() != tc.err {
			t.Errorf("%q: got error %v, want %q", tc.md, err, tc.err)
		}
	}
}
