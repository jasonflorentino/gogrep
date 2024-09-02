package pttrn

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/jasonflorentino/gogrep/src/idxablstr"
	"github.com/jasonflorentino/gogrep/src/lib"
)

var (
	DIGITS    = `0123456789`
	ALPHANUMS = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ%s_` + DIGITS
)

type Pattern []*PatternChar

func (p *Pattern) ToString() string {
	str := ""
	for i, v := range *p {
		str += fmt.Sprintf("%2d: [%p] %v\n", i, &(*p)[i], v)
	}
	return str
}

func (p *Pattern) GroupNumber(num int) *PatternChar {
	if num < 0 {
		return nil
	}
	groupNum := 0
	for i, pChar := range *p {
		if pChar.PType == Group {
			groupNum++
		}
		if groupNum == num {
			return (*p)[i]
		}
	}
	return nil
}

func (p *Pattern) Reset() {
	for _, pc := range *p {
		pc.Reset()
	}
}

func BuildPattern(pattern idxablstr.IndexableString) (*Pattern, error) {
	log := lib.LogPrefix("buildPattern")
	patternChars := Pattern{}
	var prevChar string
	for i := 0; i < len(pattern); i++ {
		log(fmt.Sprintf("i: %d", i))
		char := pattern[i]
		pChar := PatternChar{}
		pChar.PType = Literal
		log(fmt.Sprintf("char: %s", char))
		switch char {
		case `^`:
			pChar.PType = Start
		case `$`:
			if i == 0 {
				return nil, errors.New("bad pattern $ at 0")
			}
			pChar.PType = End
		case `+`:
			if prevChar == "" {
				return nil, errors.New("bad pattern + doesnt succeed rune")
			}
			patternChars[len(patternChars)-1].PType = OneOrMore
			// Doesn't append a new pChar, only modyfies the previous one
			continue
		case `?`:
			if prevChar == "" {
				return nil, errors.New("bad pattern ? doesnt succeed rune")
			}
			patternChars[len(patternChars)-1].PType = ZeroOrOne
			// Doesn't append a new pChar, only modyfies the previous one
			continue
		case `.`:
			pChar.PType = Wildcard
		case `\`:
			i++
			nextChar := pattern[i]
			log(fmt.Sprintf("next char: %v", nextChar))
			switch nextChar {
			case "d":
				pChar.Values = DIGITS
			case `w`:
				pChar.Values = ALPHANUMS
			default:
				// Assume backreference
				groupNum, err := strconv.Atoi(nextChar)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("nextChar %s is not a number", nextChar))
				}
				pChar.PType = BackRef
				group := patternChars.GroupNumber(groupNum)
				if group == nil {
					return nil, errors.New("no previous group, although i think according to spec we should just match anything?")
				}
				pChar.References = group
			}
		case `[`:
			for offset := 1; offset < len(pattern)-i; offset++ {
				nextChar := pattern[i+offset]
				log(fmt.Sprintf("[ offset: %d, nextChar: %v, pattern len: %d", offset, nextChar, len(pattern)))
				switch nextChar {
				case "]":
					i += offset
					break
				case "^":
					if pChar.Values == "" {
						pChar.Exclude = true
					} else {
						pChar.Values += nextChar
					}
				default:
					prevChar = nextChar
					pChar.Values += nextChar
				}
			}
			if pChar.Values == "]" {
				return nil, errors.New(fmt.Sprintf("No chars in bracket pattern %s", pattern[i:]))
			}
		case `]`:
			return nil, errors.New("Closing an unopened bracket")
		case `(`:
			pChar.PType = Group
			pStrs := make([]string, 0)
			pStr := ""
			// Recursively build up strings
			// to call buildPattern on and push each
			// built pattern into the top-level pattern alternates
		capture:
			for offset := 1; offset < len(pattern)-i; offset++ {
				nextChar := pattern[i+offset]
				log(fmt.Sprintf("( offset: %d, nextChar: %v, pattern len: %d", offset, nextChar, len(pattern)))
				switch nextChar {
				case `)`:
					i += offset
					break capture
				case `|`:
					if pStr == "" && len(pStrs) == 0 {
						return nil, errors.New(fmt.Sprintf("no pattern before alternation"))
					}
					pStrs = append(pStrs, pStr)
					pStr = ""
				default:
					pStr += nextChar
				}
			}
			if pStr != "" {
				pStrs = append(pStrs, pStr)
			}
			log(fmt.Sprintf("( pStrs: %v", pStrs))
			for _, str := range pStrs {
				pat, err := BuildPattern(idxablstr.FromString(str))
				if err != nil {
					return nil, err
				}
				pChar.Alternates = append(pChar.Alternates, pat)
			}
		case `)`:
			// Handling an opening paren should always include
			// consumeing the matching closing paren
			return nil, errors.New("Closing an unopened parenthesis")
		default:
			prevChar = char
			pChar.Values += char
		}
		log(fmt.Sprintf("pChar: %v", pChar))
		patternChars = append(patternChars, &pChar)
	}
	return &patternChars, nil
}
