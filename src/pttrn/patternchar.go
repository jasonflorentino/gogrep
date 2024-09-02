package pttrn

import "fmt"

type PatternChar struct {
	// please don't mutate once pattern is built
	PType      PatternType
	Values     string
	Exclude    bool
	References *PatternChar
	Alternates []*Pattern
	// mutable
	Occurrences int
	Matched     *Pattern
}

func (p *PatternChar) Reset() {
	if p.Occurrences > 0 {
		p.Occurrences = 0
	}
	if p.Matched != nil {
		p.Matched = nil
	}
}

func (p *PatternChar) String() string {
	return fmt.Sprintf("{pType: %v, values: %s, exclude: %v, references: %v, occ: %d, alternates: %v, matched: %v}", p.PType, p.Values, p.Exclude, p.References, p.Occurrences, p.Alternates, p.Matched)
}

// Match that takes into account the pChar's
// `Exclude` status
func (p *PatternChar) XMatch(match bool) bool {
	return (match && !p.Exclude) || (!match && p.Exclude)
}
