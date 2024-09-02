package debug

import "fmt"

var DEBUG bool = false

func Log(msg string) {
	if DEBUG {
		fmt.Println(msg)
	}
}

func LogPrefix(pre string) func(string) {
	return func(msg string) {
		Log(fmt.Sprintf("%s: %s", pre, msg))
	}
}
