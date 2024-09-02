package lib

import "fmt"

func Log(msg string) {
	if ARGS.Debug {
		fmt.Println(msg)
	}
}

func LogPrefix(pre string) func(string) {
	return func(msg string) {
		Log(fmt.Sprintf("%s: %s", pre, msg))
	}
}
