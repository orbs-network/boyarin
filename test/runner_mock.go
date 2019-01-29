package test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
)

// Set FailedAttempts to sabotage Run()
type RunnerMock struct {
	mock.Mock
	attempts       int
	FailedAttempts int
}

func (m *RunnerMock) Run(ctx context.Context) (err error) {
	m.MethodCalled("Run", ctx)

	if m.attempts == m.FailedAttempts {
		return
	}
	m.attempts += 1

	return fmt.Errorf("some error")
}
