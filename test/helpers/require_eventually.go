package helpers

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"runtime"
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
func RequireEventually(t TestingT, duration time.Duration, f func(t TestingT)) {
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
	if mock.format == "" {
		t.Errorf("test failed after running for %v", duration+time.Since(timeout))
	} else {
		t.Errorf("test failed after running for %v :\n %s", duration+time.Since(timeout), fmt.Sprintf(mock.format, mock.args...))
	}
}
