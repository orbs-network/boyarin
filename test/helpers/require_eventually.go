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

/**
run the test func with a testing.T -like reporter
func will run eventuallyIterations times at most,  but will not start later than the specified duration
expects the test func to succeed at least once, at which point this function returns immediately, Otherwise, the parent test will fail with the details of the last func failure.
*/
func RequireEventually(t *testing.T, duration time.Duration, f func(t TestingT)) {
	var mock testingT
	ticker := time.NewTicker(duration / eventuallyIterations) //maximum eventuallyIterations iterations
	defer ticker.Stop()
	timeout := time.Now().Add(duration)
	exec := func() {
		c := make(chan struct{})
		go func() {
			defer close(c)
			mock = testingT{}
			f(&mock)
		}()
		<-c
	}
	exec()
	for range ticker.C {
		if !mock.Failed {
			return
		}
		if time.Now().After(timeout) {
			break
		}
		exec()
	}
	t.Logf("failed after running for %v", duration+time.Since(timeout))
	if mock.format == "" {
		t.Fatalf("test failed")
	} else {
		t.Fatalf(mock.format, mock.args...)
	}
}
