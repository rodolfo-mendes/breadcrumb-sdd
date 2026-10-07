package crumb

import "sort"

// SpecType is the type of the breadcrumbs that get a view.
const SpecType = "spec"

// View is what stands around one spec of a set: the spec, every
// breadcrumb with a link to it, every breadcrumb it has a link to, and
// the links among those. Each breadcrumb is given by its index in the
// set.
type View struct {
	Spec        int
	Breadcrumbs []int      // the spec included, in the order of ByID
	Links       []ViewLink // by From, then To, in the order of Breadcrumbs, then by Verb
}

// ViewLink is a link drawn in a view: From and To are indexes in the
// set, of the breadcrumb the link is written in and of the one it
// points to.
type ViewLink struct {
	From, To int
	Verb     string
}

// ByID returns the index of each breadcrumb of the set, in order of
// id, compared as bytes, and then of place.
func (s Set) ByID() []int {
	order := make([]int, len(s.Breadcrumbs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := s.Breadcrumbs[order[i]], s.Breadcrumbs[order[j]]
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		return before(a.At, b.At)
	})
	return order
}

// Views returns one view for each breadcrumb of the set whose type is
// SpecType, in the order of ByID. A link belongs to the breadcrumb
// written in the same text of its place, such as the same path, as a
// claim does, and points to every breadcrumb that has the id of its
// object. A link that points to no breadcrumb is in no view.
func (s Set) Views() []View {
	order := s.ByID()
	rank := make([]int, len(order))
	for r, i := range order {
		rank[i] = r
	}
	at := map[string][]int{}  // the breadcrumbs written in each text
	ids := map[string][]int{} // the breadcrumbs that have each id
	for i, b := range s.Breadcrumbs {
		at[b.At.Text] = append(at[b.At.Text], i)
		ids[b.ID] = append(ids[b.ID], i)
	}
	var links []ViewLink
	seen := map[ViewLink]bool{}
	for _, l := range s.Links {
		for _, from := range at[l.At.Text] {
			for _, to := range ids[l.Object] {
				if vl := (ViewLink{From: from, To: to, Verb: l.Verb}); !seen[vl] {
					seen[vl] = true
					links = append(links, vl)
				}
			}
		}
	}
	sort.SliceStable(links, func(i, j int) bool {
		a, b := links[i], links[j]
		switch {
		case a.From != b.From:
			return rank[a.From] < rank[b.From]
		case a.To != b.To:
			return rank[a.To] < rank[b.To]
		}
		return a.Verb < b.Verb
	})

	var views []View
	for _, spec := range order {
		if s.Breadcrumbs[spec].Type != SpecType {
			continue
		}
		in := map[int]bool{spec: true}
		for _, l := range links {
			switch spec {
			case l.To:
				in[l.From] = true
			case l.From:
				in[l.To] = true
			}
		}
		v := View{Spec: spec}
		for _, i := range order {
			if in[i] {
				v.Breadcrumbs = append(v.Breadcrumbs, i)
			}
		}
		for _, l := range links {
			if in[l.From] && in[l.To] {
				v.Links = append(v.Links, l)
			}
		}
		views = append(views, v)
	}
	return views
}

// ClaimsOf returns, for each breadcrumb of the set, the index of each
// of its claims, in the order of the set. A claim belongs to the
// breadcrumb written in the same text of its place (ADR-0022).
func (s Set) ClaimsOf() [][]int {
	in := map[string][]int{}
	for i, c := range s.Claims {
		in[c.At.Text] = append(in[c.At.Text], i)
	}
	of := make([][]int, len(s.Breadcrumbs))
	for i, b := range s.Breadcrumbs {
		of[i] = in[b.At.Text]
	}
	return of
}

// GivenVerdict is the verdict given before to a claim or a breadcrumb
// of a set. Audited is false when none was: the set was not audited,
// and Verdict then says nothing.
type GivenVerdict struct {
	Audited bool
	Verdict Verdict
	Reason  string
}

// Given returns the verdict given before to each claim of the set and
// to each of its breadcrumbs, each in the order of the set. A verdict
// belongs to what is written at its place, not to an id (ADR-0023).
// When two were given at one place, the one given last holds.
func (s Set) Given() (claims, breadcrumbs []GivenVerdict) {
	given := func(verdicts []SetVerdict) map[Place]GivenVerdict {
		at := map[Place]GivenVerdict{}
		for _, v := range verdicts {
			at[v.At] = GivenVerdict{Audited: true, Verdict: v.Verdict, Reason: v.Reason}
		}
		return at
	}
	ofClaims, ofBreadcrumbs := given(s.ClaimVerdicts), given(s.Verdicts)
	for _, c := range s.Claims {
		claims = append(claims, ofClaims[c.At])
	}
	for _, b := range s.Breadcrumbs {
		breadcrumbs = append(breadcrumbs, ofBreadcrumbs[b.At])
	}
	return claims, breadcrumbs
}
