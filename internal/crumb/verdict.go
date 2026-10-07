package crumb

import (
	"bytes"
	"fmt"
)

// Verdict says whether the repository's files support a claim, or the
// claims of a breadcrumb (ADR-0022).
type Verdict int

const (
	Undecided Verdict = iota // nothing was checked, or something could not be
	Confirmed                // every claim holds
	Refuted                  // a claim does not hold
)

func (v Verdict) String() string {
	switch v {
	case Confirmed:
		return "Confirmed"
	case Refuted:
		return "Refuted"
	}
	return "Undecided"
}

// NoClaims is the reason a breadcrumb with no claim and no problem is
// Undecided.
const NoClaims = "no claims"

// Verdict returns the verdict of c, a claim that passes Check. file
// says whether its target is a regular file, and content is then what
// the file holds. A claim whose target is not a file does not hold
// (ADR-0019, ADR-0023).
func (c Claim) Verdict(file bool, content []byte) Verdict {
	if file && claimKinds[c.Kind].holds(content, c.Argument) {
		return Confirmed
	}
	return Refuted
}

// Lacks says why c is Refuted: what its target lacks, or that it is
// not a file.
func (c Claim) Lacks(file bool) string {
	if !file {
		return fmt.Sprintf("%s is not a file", c.Target)
	}
	return fmt.Sprintf("%s has %s", c.Target, claimKinds[c.Kind].lacks(c.Argument))
}

// hasLine reports whether a has-line claim with text holds in content:
// at least one line, without its line ending (\n or \r\n) and without
// the spaces and tabs at its start and end, is equal to text, byte for
// byte (ADR-0019).
func hasLine(content []byte, text string) bool {
	for _, line := range bytes.Split(content, []byte("\n")) {
		line = bytes.TrimSuffix(line, []byte("\r"))
		if string(bytes.Trim(line, " \t")) == text {
			return true
		}
	}
	return false
}

func lacksLine(text string) string { return fmt.Sprintf("no line %q", text) }

// Judge returns the verdict of a breadcrumb from the verdicts of its
// claims and the number of its problems, each of which counts as one
// Undecided claim (ADR-0022). It is Refuted when a claim is Refuted;
// otherwise Undecided when one is Undecided; otherwise Confirmed when
// it has a claim. With no claim and no problem it is Undecided, and
// reason is NoClaims; reason is otherwise empty.
func Judge(claims []Verdict, problems int) (verdict Verdict, reason string) {
	undecided := problems > 0
	for _, c := range claims {
		switch c {
		case Refuted:
			return Refuted, ""
		case Undecided:
			undecided = true
		}
	}
	switch {
	case undecided:
		return Undecided, ""
	case len(claims) > 0:
		return Confirmed, ""
	}
	return Undecided, NoClaims
}

// Target is the file a claim is about, as it was found: whether it is
// a regular file, and then its content.
type Target struct {
	File    bool
	Content []byte
}

// BreadcrumbVerdict is the verdict of a breadcrumb of a set, with the
// reason Judge gives.
type BreadcrumbVerdict struct {
	Verdict Verdict
	Reason  string
}

// Targets returns the target of each claim of the set, once each, in
// the order first written.
func (s Set) Targets() []string {
	var targets []string
	seen := map[string]bool{}
	for _, c := range s.Claims {
		if !seen[c.Claim.Target] {
			seen[c.Claim.Target] = true
			targets = append(targets, c.Claim.Target)
		}
	}
	return targets
}

// Audit returns the verdict of each claim of the set and of each of
// its breadcrumbs, each in the order of the set. targets holds what
// was found at each of Targets; a target it does not hold is not a
// file. A claim and a problem belong to the breadcrumb written in the
// same text of their place, such as the same path (ADR-0021,
// ADR-0022); one that has no such breadcrumb belongs to none.
func (s Set) Audit(targets map[string]Target) (claims []Verdict, breadcrumbs []BreadcrumbVerdict) {
	of := map[string][]Verdict{} // the verdicts of the claims written in each text
	for _, c := range s.Claims {
		t := targets[c.Claim.Target]
		v := c.Claim.Verdict(t.File, t.Content)
		claims = append(claims, v)
		of[c.At.Text] = append(of[c.At.Text], v)
	}
	problems := map[string]int{}
	for _, p := range s.Problems {
		problems[p.At.Text]++
	}
	for _, b := range s.Breadcrumbs {
		v, reason := Judge(of[b.At.Text], problems[b.At.Text])
		breadcrumbs = append(breadcrumbs, BreadcrumbVerdict{Verdict: v, Reason: reason})
	}
	return claims, breadcrumbs
}
