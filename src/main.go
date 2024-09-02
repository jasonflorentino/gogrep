package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jasonflorentino/gogrep/src/debug"
	"github.com/jasonflorentino/gogrep/src/idxablstr"
	"github.com/jasonflorentino/gogrep/src/pttrn"
)

type Args struct {
	expr  string
	debug bool
}

func toArgsMap(args []string) Args {
	argsMap := Args{}
	for i := 0; i < len(args); {
		v := args[i]
		switch v {
		case "-E":
			argsMap.expr = args[i+1]
			i += 1
		case "--debug":
			argsMap.debug = true
			debug.DEBUG = true
		}
		i += 1
	}
	debug.Log(fmt.Sprintf("%v", argsMap))
	return argsMap
}

// 1 means no lines were selected, >1 means error
func bail(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(2)
}

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 {
		bail("usage: mygrep -E <pattern>")
	}

	args := toArgsMap(os.Args[1:])

	pattern := args.expr
	if pattern == "" {
		bail("usage: mygrep -E <pattern>")
	}

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if line[len(line)-1] == 10 {
		debug.Log("Stripping new line")
		line = line[:len(line)-1]
	}
	debug.Log(fmt.Sprintf("line: %s", line))
	if err != nil {
		bail(fmt.Sprintf("read input text: %v", err))
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		bail(err.Error())
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
	os.Exit(0)
}

func matchLine(line []byte, pattern string) (bool, error) {
	patternChars, err := pttrn.BuildPattern(idxablstr.FromString(pattern))
	if err != nil {
		bail(err.Error())
	}
	debug.Log(fmt.Sprintf("patternChars: %v", patternChars))
	return rMatch(idxablstr.FromBytes(line), patternChars), nil
}

func rMatch(line idxablstr.IndexableString, pattern *pttrn.Pattern) bool {
	log := debug.LogPrefix("rMatch")
	for i := range line {
		isMatching, halt := rPatternMatch(line[i:], pattern, 0)
		log(fmt.Sprintf("isMatching: %v, halt: %v", isMatching, halt))
		if halt || isMatching {
			return isMatching
		}
	}
	return false
}

// rPatternMatch recursively matches a pattern against the line
// The first return val tells the caller whether or not a match was
// found. And if it wasn't the second return val says whether or not
// to continue trying to match.
func rPatternMatch(line idxablstr.IndexableString, pattern *pttrn.Pattern, patIdx int) (bool, bool) {
	log := debug.LogPrefix("  rPatternMatch")
	debug.Log("\n")
	log(fmt.Sprintf("patIdx: %d", patIdx))
	log(fmt.Sprintf("line: %v, len:%d", line, len(line)))
	log(fmt.Sprintf("pattern: %s, len:%d", pattern.ToString(), len(*pattern)))
	// Check then get the next pattern char
	if len(*pattern) == patIdx {
		// We've reach the end without failing
		return true, true
	}
	var pChar *pttrn.PatternChar
	pChar = (*pattern)[patIdx]
	log(fmt.Sprintf("pChar: %v", pChar))

	// Check then get the next line char
	if len(line) == 0 {
		// Only a match if the next pattern char is
		// the End or if we're allowed zero of the previous char
		return pChar.PType == pttrn.End || pChar.PType == pttrn.ZeroOrOne, true
	}

	switch pChar.PType {
	case pttrn.Start:
		isMatching, _ := rPatternMatch(line, pattern, patIdx+1)
		return isMatching, true
	case pttrn.OneOrMore:
		match := strings.Contains(pChar.Values, line[0])
		log(fmt.Sprintf("match: %v", match))
		if (match && !pChar.Exclude) || (!match && pChar.Exclude) {
			pChar.Occurrences++
			isMatching, halt := rPatternMatch(line[1:], pattern, patIdx)
			return isMatching, halt
		} else {
			if pChar.Occurrences == 0 {
				return false, false
			}
			isMatching, halt := rPatternMatch(line[0:], pattern, patIdx+1)
			return isMatching, halt
		}
	case pttrn.Wildcard:
		isMatching, halt := rPatternMatch(line[1:], pattern, patIdx+1)
		return isMatching, halt
	case pttrn.Literal:
		match := strings.Contains(pChar.Values, line[0])
		log(fmt.Sprintf("match: %v", match))
		if (match && !pChar.Exclude) || (!match && pChar.Exclude) {
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
		if (match && !pChar.Exclude) || (!match && pChar.Exclude) {
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
