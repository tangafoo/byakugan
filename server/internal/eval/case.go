package eval

import (
	"byakugan/internal/corpus"
)

// Case struct from the eval gate. Expect are statute-
// qualified (corpus.RelatedSections) because bare section numbers collide across
// acts — DDA 1952 and MOA 1955 both have a s31, and an eval that can't tell
// them apart will happily credit the wrong law.
//
// Forbid is the tripwire list: sections that must NOT appear for this
// question (e.g. a graffiti question surfacing the Dangerous Drugs Act).
// A forbidden hit fails the case even if every expected section was found.
type Case struct {
	ID         string                  `json:"id"`
	Question   string                  `json:"question"`
	Lang       corpus.Lang             `json:"lang"`
	Expect     []corpus.RelatedSection `json:"expect"`
	Forbid     []corpus.RelatedSection `json:"forbid,omitempty"`
	ShouldFind bool                    `json:"should_find"`
}

func (c Case) Valid() bool {
	switch {
	case len(c.Question) == 0:
		return false
	case !c.Lang.Valid():
		return false
	case c.ShouldFind && len(c.Expect) == 0:
		return false
	}

	for _, r := range c.Expect {
		if !r.Valid() {
			return false
		}
	}
	for _, r := range c.Forbid {
		if !r.Valid() {
			return false
		}
	}

	return true
}
