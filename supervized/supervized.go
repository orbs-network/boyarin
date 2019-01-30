package supervized

// Copied from orbs-network-go/synchronization/supervized

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime"
	"runtime/debug"
	"strings"
)

// Runs f() in a goroutine; if it panics, logs the error and stack trace to the specified Errorer
func GoOnce(f func()) {
	go func() {
		tryOnce(f)
	}()
}

// Runs f() in a goroutine; if it panics, logs the error and stack trace to the specified Errorer
// If the provided Context isn't closed, re-runs f()
// Returns a channel that is closed when the goroutine has quit due to context ending
func GoForever(f func()) {
	go func() {
		for {
			tryOnce(f)
		}
	}()
}

// this function is needed so that we don't return out of the goroutine when it panics
func tryOnce(f func()) {
	defer recoverPanics()
	f()
}

func recoverPanics() {
	if p := recover(); p != nil {
		e := errors.Errorf("goroutine panicked at [%s]: %v", identifyPanic(), p)
		fmt.Println(fmt.Errorf("recovered panic: %s\n%s", e.Error(), string(debug.Stack())))
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
