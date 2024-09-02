package pttrn

import "fmt"

type PatternChar struct {
	PType       PatternType
	Values      string
	Exclude     bool
	References  *PatternChar
	Occurrences int
	Alternates  []*Pattern
	Matched     *Pattern
}

func (p *PatternChar) String() string {
	return fmt.Sprintf("{pType: %v, values: %s, exclude: %v, references: %v, occ: %d, alternates: %v, matched: %v}", p.PType, p.Values, p.Exclude, p.References, p.Occurrences, p.Alternates, p.Matched)
}
