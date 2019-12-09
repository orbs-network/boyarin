package helpers

import (
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
	"time"
)

type TestingT interface {
	require.TestingT
}

type testingT struct {
	Failed bool
	format string
	args   []interface{}
}

func (t *testingT) FailNow() {
	t.Failed = true
	runtime.Goexit()
}

func (t *testingT) Errorf(format string, args ...interface{}) {
	t.format = format
	t.args = args
	t.FailNow()
}

func RequireEventually(t *testing.T, timeout time.Duration, f func(t TestingT)) {
	var mock testingT

	for i := 0; i < eventuallyIterations; i++ {
		c := make(chan struct{})
		go func() {
			defer close(c)
			mock = testingT{}
			f(&mock)
		}()
		<-c
		if !mock.Failed {
			return
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	if mock.format == "" {
		t.Fatalf("test failed")
	} else {
		t.Fatalf(mock.format, mock.args...)
	}
}
