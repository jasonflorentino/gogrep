package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jasonflorentino/gogrep/src/idxablstr"
	"github.com/jasonflorentino/gogrep/src/lib"
	"github.com/jasonflorentino/gogrep/src/pttrn"
)

const usage string = `
jrep: A worse version of grep. Reads from stdin. Use the -E option to specify an expression.
      echo <input> | jrep -E <pattern>

      Option        Description
      -E <pattern>  Expression to use (required)
      --debug       Print debug information
      --help        Display this message

`

// echo 'hello' | go run src/main.go -- -E 'hell'
//
// Exit Codes:
//
//	0 - Successful match
//	1 - No match found
//	2 - Error
func main() {
	if len(os.Args) < 3 {
		fmt.Print(usage)
		bail("Missing args")
	}

	lib.AssignArgs(os.Args[1:])

	if lib.ARGS.Help {
		fmt.Print(usage)
		os.Exit(0)
	}

	if lib.ARGS.Expr == "" {
		fmt.Print(usage)
		bail("No expression.")
	}

	// Get input

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if line[len(line)-1] == 10 {
		lib.Log("Stripping new line")
		line = line[:len(line)-1]
	}
	lib.Log(fmt.Sprintf("line: %s", line))
	if err != nil {
		bail(fmt.Sprintf("read input text: %v", err))
	}

	// Build pattern

	pattern, err := pttrn.BuildPattern(idxablstr.FromString(lib.ARGS.Expr))
	lib.Log(fmt.Sprintf("pattern: %v", pattern))
	if err != nil {
		bail(err.Error())
	}

	// Match

	ok := matchLine(idxablstr.FromBytes(line), pattern)

	if !ok {
		os.Exit(1)
	}

	os.Exit(0)
}

func bail(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(2)
}

func matchLine(line idxablstr.IndexableString, pattern *pttrn.Pattern) bool {
	log := lib.LogPrefix("matchLine")
	for i := range line {
		isMatching, halt := rPatternMatch(line[i:], pattern, 0)
		log(fmt.Sprintf("isMatching: %v, halt: %v", isMatching, halt))
		if halt || isMatching {
			return isMatching
		}
	}
	return false
}

// rPatternMatch recursively matches a pattern against the line.
// Recurses by contnually slicing the head of `line`, and testing it
// against the Pattern Char at `patIdx`. It does this until there are no
// more `line` characters left to match, or end of the Pattern is reached.
//
// The first return val tells the caller whether or not a match was found.
// And if it wasn't the second return val says whether to continue with
// trying to match or to stop because a match is no longer possible.
func rPatternMatch(line idxablstr.IndexableString, pattern *pttrn.Pattern, patIdx int) (bool, bool) {
	log := lib.LogPrefix("  rPatternMatch")
	lib.Log("\n")
	log(fmt.Sprintf("patIdx: %d", patIdx))
	log(fmt.Sprintf("line: %v, len:%d", line, len(line)))
	log(fmt.Sprintf("pattern: %s, len:%d", pattern.ToString(), len(*pattern)))

	if len(*pattern) == patIdx {
		// We've reach the end without failing
		return true, true
	}
	var pChar *pttrn.PatternChar
	pChar = (*pattern)[patIdx]
	log(fmt.Sprintf("pChar: %v", pChar))

	if len(line) == 0 {
		// Only a match if the next pattern char if
		// the End or if we're allowed zero of the previous char
		return pChar.PType == pttrn.End || pChar.PType == pttrn.ZeroOrOne, true
	}

	switch pChar.PType {

	case pttrn.Start:
		isMatching, _ := rPatternMatch(line, pattern, patIdx+1)
		return isMatching, true

	case pttrn.OneOrMore:
		match := strings.Contains(pChar.Values, line[0])
		log(fmt.Sprintf("match: %v, occurrences: %d", match, pChar.Occurrences))
		if pChar.Occurrences == 0 {
			if pChar.XMatch(match) {
				pChar.Occurrences++
				isMatching, halt := rPatternMatch(line[1:], pattern, patIdx)
				return isMatching, halt
			} else {
				return false, false
			}
		} else {
			if patIdx+2 <= len(*pattern) {
				nextCharPattern := (*pattern)[patIdx+1 : patIdx+2]
				nextCharMatch, _ := rPatternMatch(line[0:], &nextCharPattern, 0)
				if nextCharMatch {
					// OK do the rest of it
					isMatching, halt := rPatternMatch(line[0:], pattern, patIdx+1)
					return isMatching, halt
				}
			}
			if pChar.XMatch(match) {
				pChar.Occurrences++
				isMatching, halt := rPatternMatch(line[1:], pattern, patIdx)
				return isMatching, halt
			} else {
				isMatching, halt := rPatternMatch(line[1:], pattern, patIdx+1)
				return isMatching, halt
			}
		}

	case pttrn.Wildcard:
		isMatching, halt := rPatternMatch(line[1:], pattern, patIdx+1)
		return isMatching, halt

	case pttrn.Literal:
		match := strings.Contains(pChar.Values, line[0])
		log(fmt.Sprintf("match: %v", match))
		if pChar.XMatch(match) {
			isMatching, halt := rPatternMatch(line[1:], pattern, patIdx+1)
			return isMatching, halt
		} else {
			return false, false
		}

	case pttrn.Group:
		isMatching, halt := false, false
		matchLen := 0
		for _, altAddr := range pChar.Alternates {
			log(fmt.Sprintf("[altPat] %v", altAddr))
			isMatching, halt = rPatternMatch(line[0:], altAddr, 0)
			log(fmt.Sprintf("[altPat] isMatching: %v, halt: %v", isMatching, halt))
			if isMatching {
				log(fmt.Sprintf("[altPat] matchedAddr: %#v", altAddr))
				pChar.Matched = altAddr
				// BUG: Assumes only char matches and no groups
				matchLen = len(*altAddr)
				break
			}
		}
		log(fmt.Sprintf("isMatching: %v, matchLen: %d", isMatching, matchLen))
		if isMatching {
			isMatching, halt = rPatternMatch(line[0+matchLen:], pattern, patIdx+1)
		} else {
			return false, false
		}
		return isMatching, halt

	case pttrn.BackRef:
		if pChar.References.Matched == nil {
			bail(fmt.Sprintf("Backref expected a matched pattern.\n\n"))
		}
		isMatching, halt := rPatternMatch(line[0:], pChar.References.Matched, 0)
		log(fmt.Sprintf("isMatching: %v, halt: %v", isMatching, halt))
		return isMatching, halt

	case pttrn.ZeroOrOne:
		match := strings.Contains(pChar.Values, line[0])
		log(fmt.Sprintf("match: %v", match))
		if pChar.XMatch(match) {
			pChar.Occurrences++
			isMatching, halt := rPatternMatch(line[1:], pattern, patIdx+1)
			return isMatching, halt
		} else {
			isMatching, halt := rPatternMatch(line[0:], pattern, patIdx+1)
			return isMatching, halt
		}

	default:
		bail(fmt.Sprintf("matching for type %v is not implemented", pChar.PType))
	}
	return false, true
}
