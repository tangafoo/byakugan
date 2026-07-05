package eval

import (
	"byakugan/internal/corpus"
)

type Case struct {
	ID             string      `json:"id"`
	Question       string      `json:"question"`
	Lang           corpus.Lang `json:"lang"`
	ExpectSections []string    `json:"expect_sections"`
	ShouldFind     bool        `json:"should_find"`
}

func (c Case) Valid() bool {
	switch {
	case len(c.Question) == 0:
		return false
	case !c.Lang.Valid():
		return false
	case c.ShouldFind && len(c.ExpectSections) == 0:
		return false
	default:
		return true
	}
}
