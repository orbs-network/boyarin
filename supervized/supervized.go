package supervized

// Copied from orbs-network-go/synchronization/supervized

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Runs f() in a goroutine; if it panics, logs the error and stack trace
func GoOnce(f func(bool)) {
	go func() {
		tryOnce(f, true)
	}()
}

// Runs f() in a goroutine; if it panics, logs the error and stack trace
// Returns a channel that is closed when the goroutine has quit due to context ending
func GoForever(f func(bool)) chan interface{} {
	c := make(chan interface{})
	go func() {
		tryOnce(f, true)
		for {
			tryOnce(f, false)
		}
	}()

	return c
}

// this function is needed so that we don't return out of the goroutine when it panics
func tryOnce(f func(bool), first bool) {
	defer recoverPanics()
	f(first)
}

func recoverPanics() {
	if p := recover(); p != nil {
		e := fmt.Errorf("goroutine panicked at [%s]: %v", identifyPanic(), p)
		fmt.Println(e)
		fmt.Println(string(debug.Stack()))
	}
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}

	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}
