package utils

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestTry(t *testing.T) {
	var iPtr *int32
	iPtr = new(int32)

	err := Try(context.Background(), 4, 10*time.Millisecond, 1*time.Millisecond, func(ctxWithTimeout context.Context) error {
		atomic.AddInt32(iPtr, 1)
		if *iPtr != int32(3) {
			return fmt.Errorf("no!")
		}

		return nil
	})

	require.NoError(t, err)
	require.EqualValues(t, 3, *iPtr, "should attempt to execute 3 times")
}

func TestTryWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := Try(ctx, 100, 10*time.Millisecond, 1*time.Millisecond, func(ctxWithTimeout context.Context) error {
		return fmt.Errorf("no!")
	})

	require.EqualError(t, err, "context deadline exceeded after 7 attempts")
}

func TestTryWithError(t *testing.T) {
	var iPtr *int32
	iPtr = new(int32)

	err := Try(context.Background(), 7, 10*time.Millisecond, 1*time.Millisecond, func(ctxWithTimeout context.Context) error {
		atomic.AddInt32(iPtr, 1)
		return fmt.Errorf("no!")
	})

	require.EqualError(t, err, "no!")
	require.EqualValues(t, 7, *iPtr, "should attempt to execute 3 times")
}

func TestTryAllAttemptsFailed(t *testing.T) {
	var iPtr *int32
	iPtr = new(int32)

	err := Try(context.Background(), 4, 10*time.Millisecond, 1*time.Millisecond, func(ctxWithTimeout context.Context) error {
		atomic.AddInt32(iPtr, 1)
		return fmt.Errorf("no!")
	})

	require.EqualError(t, err, "no!")
	require.EqualValues(t, 4, *iPtr, "should attempt to execute 4 times")
}
