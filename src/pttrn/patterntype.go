package pttrn

import "fmt"

type PatternType int

const (
	BackRef   PatternType = iota
	End                   = iota
	Group                 = iota
	Literal               = iota
	OneOrMore             = iota
	Start                 = iota
	Wildcard              = iota
	ZeroOrOne             = iota
)

func (i PatternType) String() string {
	switch i {
	case BackRef:
		return "BackRef"
	case End:
		return "End"
	case Group:
		return "Group"
	case Literal:
		return "Literal"
	case OneOrMore:
		return "OneOrMore"
	case Start:
		return "Start"
	case Wildcard:
		return "Wildcard"
	case ZeroOrOne:
		return "ZeroOrOne"
	default:
		return fmt.Sprintf("PatternType(%d)", i)
	}
}
