package lib

import "fmt"

var ARGS *Args

type Args struct {
	Expr  string
	Debug bool
	Help  bool
}

func AssignArgs(args []string) {
	a := Args{}
	ARGS = &a
	for i := 0; i < len(args); {
		v := args[i]
		switch v {
		case "-E":
			ARGS.Expr = args[i+1]
			i += 1
		case "--debug":
			ARGS.Debug = true
		case "--help":
			ARGS.Help = true
		}
		i += 1
	}
	Log(fmt.Sprintf("%v", ARGS))
}
