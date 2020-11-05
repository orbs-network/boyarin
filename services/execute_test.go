package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestExecuteWithInvalidConfig(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		executionCtx := context.Background()

		startTime := time.Now()

		resetTimeout := 1 * time.Second
		pollingInterval := 100 * time.Millisecond
		waiter, err := Execute(ctx, &config.Flags{
			ConfigUrl:             "http://localhost/fake-url",
			KeyPairConfigPath:     "../boyar/config/test/fake-key-pair.json",
			PollingInterval:       pollingInterval,
			BootstrapResetTimeout: resetTimeout,
		}, logger)
		require.NoError(t, err)

		waiter.WaitUntilShutdown(executionCtx)

		require.InDelta(t, resetTimeout, time.Since(startTime).Nanoseconds(), float64(2*pollingInterval.Nanoseconds()))
	})
}
