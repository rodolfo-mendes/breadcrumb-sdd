package specsite

import (
	"os"
	"strings"
	"testing"
)

const repo = "https://github.com/rodolfo-mendes/breadcrumb-sdd"

func spec(version, body string) string {
	return "# Breadcrumb SDD Specification\n\nVersion " + version + "\n\n" + body
}

// TestTheSpecificationConverts fails when docs/breadcrumb-sdd.md uses
// Markdown the converter cannot read (TD-0032), which would break the
// site at the next release.
func TestTheSpecificationConverts(t *testing.T) {
	md, err := os.ReadFile("../../" + SpecPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Build(repo, []Release{{"v9.9.9", string(md)}}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestBuildKeepsTheLatestTagOfEachVersion(t *testing.T) {
	files, err := Build(repo, []Release{
		{"v0.5.0", spec("0.2.0", "Old wording.\n")},
		{"v0.10.0", spec("0.2.0", "New wording.\n")},
		{"v0.6.0", spec("0.3.0", "Next rules.\n")},
		{"v1.0.0-rc1", spec("9.0.0", "Not a release.\n")},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	if len(files) != 4 {
		t.Fatalf("files %q, want the root, 0.2.0, 0.3.0 and latest", paths)
	}
	if p := string(files["spec/0.2.0/index.html"]); !strings.Contains(p, "New wording.") || !strings.Contains(p, "v0.10.0") {
		t.Error("spec/0.2.0/ is not published from v0.10.0, its latest tag")
	}
	if string(files["spec/latest/index.html"]) != string(files["spec/0.3.0/index.html"]) {
		t.Error("spec/latest/ is not the highest version (RQ-0038)")
	}
	index := string(files["index.html"])
	if strings.Index(index, `href="spec/0.3.0/"`) > strings.Index(index, `href="spec/0.2.0/"`) {
		t.Error("the root does not list the highest version first (RQ-0038)")
	}
}

func TestBuildNeedsAVersionLine(t *testing.T) {
	if _, err := Build(repo, []Release{{"v0.5.0", "# Title\n\nNo version.\n"}}, nil); err == nil {
		t.Error("a specification with no version line was published")
	}
}

func TestLinksLeadToTheFilesAtTheTag(t *testing.T) {
	files, err := Build(repo, []Release{{"v0.5.0", spec("0.2.0",
		"See [TD-0005](../breadcrumbs/TD-0005.md), [bcr(1)](bcr.md), [Go](https://go.dev) and [terms](#terms).\n")}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := string(files["spec/0.2.0/index.html"])
	for _, want := range []string{
		`href="` + repo + `/blob/v0.5.0/breadcrumbs/TD-0005.md"`,
		`href="` + repo + `/blob/v0.5.0/docs/bcr.md"`,
		`href="https://go.dev"`,
		`href="#terms"`,
	} {
		if !strings.Contains(p, want) {
			t.Errorf("the page has no %s (RQ-0039)", want)
		}
	}
}

func TestAPageStandsOnItsOwn(t *testing.T) {
	files, err := Build(repo, []Release{{"v0.5.0", spec("0.2.0", "## The kinds of breadcrumb\n\n### Task\n\n## Task\n")}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := string(files["spec/0.2.0/index.html"])
	for _, want := range []string{`id="the-kinds-of-breadcrumb"`, `id="task"`, `id="task-2"`, "Version 0.2.0", `href="../../"`, repo + "/tree/v0.5.0"} {
		if !strings.Contains(p, want) {
			t.Errorf("the page has no %s (RQ-0040)", want)
		}
	}
	for _, not := range []string{"<script", `src="http`, `href="http://`, `<link`} {
		if strings.Contains(p, not) {
			t.Errorf("the page has %s (RQ-0040)", not)
		}
	}
}

func TestConvert(t *testing.T) {
	md := strings.Join([]string{
		"# A <title>",
		"",
		"A paragraph with `code & more`",
		"on two lines.",
		"",
		"- one",
		"  item",
		"- two",
		"",
		"```markdown",
		"- `a` contains `<b>`",
		"```",
	}, "\n")
	title, body, err := Convert(md, func(s string) string { return s })
	if err != nil {
		t.Fatal(err)
	}
	want := "<p>A paragraph with <code>code &amp; more</code> on two lines.</p>\n" +
		"<ul>\n<li>one item</li>\n<li>two</li>\n</ul>\n" +
		"<pre><code>- `a` contains `&lt;b&gt;`\n</code></pre>\n"
	if title != "A <title>" || body != want {
		t.Errorf("got %q\n%q", title, body)
	}
}

func TestConvertRefusesWhatItCannotRead(t *testing.T) {
	for _, md := range []string{
		"No heading",
		"# T\n\n| a | b |",
		"# T\n\n#### Deep",
		"# T\n\n```\nopen",
	} {
		if _, _, err := Convert(md, func(s string) string { return s }); err == nil {
			t.Errorf("%q converted", md)
		}
	}
}

func TestTheReportPageShowsTheReportInAFrame(t *testing.T) {
	report := []byte("<!doctype html><svg>the graph</svg>")
	files, err := Build(repo, []Release{{"v0.5.0", spec("0.2.0", "Rules.\n")}}, &Report{"v0.7.0", report})
	if err != nil {
		t.Fatal(err)
	}
	if string(files["report/audit-report.html"]) != string(report) {
		t.Error("the report is not published as bcr writes it (TD-0038)")
	}
	p := string(files["report/index.html"])
	for _, want := range []string{`<iframe src="audit-report.html" sandbox`, repo + "/tree/v0.7.0", `href="../"`, `href="../spec/latest/"`, `href="audit-report.html"`} {
		if !strings.Contains(p, want) {
			t.Errorf("the report page has no %s (RQ-0047, RQ-0048)", want)
		}
	}
	if strings.Contains(p, "<script") {
		t.Error("the report page has a script (RQ-0047)")
	}
	if !strings.Contains(string(files["index.html"]), `href="report/"`) {
		t.Error("the root does not link to the report page (RQ-0048)")
	}
}

func TestLatestTag(t *testing.T) {
	if got := LatestTag([]string{"v0.9.0", "v0.10.0", "v1.0.0-rc1", "nightly", "v0.2.0"}); got != "v0.10.0" {
		t.Errorf("LatestTag = %q, want v0.10.0", got)
	}
	if got := LatestTag(nil); got != "" {
		t.Errorf("LatestTag(nil) = %q", got)
	}
}
