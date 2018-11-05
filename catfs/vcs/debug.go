package vcs

import (
	"fmt"
)

const (
	printDebug = true
)

func debug(args ...interface{}) {
	if printDebug {
		fmt.Println(args...)
	}
}

func debugf(spec string, args ...interface{}) {
	if printDebug {
		fmt.Printf(spec, args...)
	}
}
