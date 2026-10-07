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
