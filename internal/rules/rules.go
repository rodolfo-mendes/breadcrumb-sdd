// Package rules reads breadcrumb.rules, the file in which a
// repository declares its shape (ADR-0025). It knows the format of the
// file, not what a rule means: it gives the core its rules, and the
// core does not use it (ADR-0013).
//
//	type	NAME
//	link	NAME	VERB	NAME
//	claims	NAME
package rules

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rodolfo-mendes/breadcrumb-sdd/internal/crumb"
)

// Name is the name of the file, at the root of a repository.
const Name = "breadcrumb.rules"

// The kind of a line is its first field, with the number of fields it
// has, the kind included.
const (
	typeKind   = "type"
	linkKind   = "link"
	claimsKind = "claims"
)

var fieldsOf = map[string]int{typeKind: 2, linkKind: 4, claimsKind: 2}

// Parse reads src, the text of a breadcrumb.rules. When a line is not
// allowed, it returns every problem, in order of line, and no rules.
// What the rules say as a whole is not checked here: the core does
// (crumb.Rules.Check).
func Parse(src []byte) (crumb.Rules, []crumb.Problem) {
	r := crumb.Rules{Name: Name}
	var problems []crumb.Problem
	text := strings.TrimPrefix(string(src), "\xEF\xBB\xBF")
	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.Trim(line, " \t") == "" || strings.HasPrefix(line, "#") {
			continue
		}
		n := i + 1
		fields := strings.Split(line, "\t")
		if why := refuse(line, fields); why != "" {
			problems = append(problems, crumb.Problem{Line: n, Message: why})
			continue
		}
		switch fields[0] {
		case typeKind:
			r.Types = append(r.Types, crumb.RuleType{Name: fields[1], Line: n})
		case linkKind:
			r.Links = append(r.Links, crumb.RuleLink{Subject: fields[1], Verb: fields[2], Object: fields[3], Line: n})
		case claimsKind:
			r.Claims = append(r.Claims, crumb.RuleClaims{Type: fields[1], Line: n})
		}
	}
	if problems != nil {
		return crumb.Rules{}, problems
	}
	return r, nil
}

// refuse returns why line, split into fields at each tab, is not
// allowed, or "" when it is. The kind of a line is the first word of
// its first field, so that a line typed with spaces is told so.
func refuse(line string, fields []string) string {
	kind := ""
	if words := strings.FieldsFunc(fields[0], unicode.IsSpace); len(words) > 0 {
		kind = words[0]
	}
	need, ok := fieldsOf[kind]
	if !ok {
		return fmt.Sprintf("unknown kind of line %q; a line is type, link or claims", kind)
	}
	if len(fields) != need {
		why := fmt.Sprintf("%s has %s, needs %d", kind, count(len(fields)), need)
		// With a tab for each run of spaces, the line would have the
		// fields it needs.
		if strings.Contains(line, " ") && len(strings.FieldsFunc(line, unicode.IsSpace)) == need {
			why += "; its fields are separated by spaces, use a tab"
		}
		return why
	}
	for i, f := range fields {
		switch {
		case f == "":
			return fmt.Sprintf("%s has an empty field %d", kind, i+1)
		case strings.IndexFunc(f, unicode.IsSpace) >= 0:
			return fmt.Sprintf("%s has white space in field %d", kind, i+1)
		}
	}
	return ""
}

// count writes a number of fields.
func count(n int) string {
	if n == 1 {
		return "1 field"
	}
	return fmt.Sprintf("%d fields", n)
}
