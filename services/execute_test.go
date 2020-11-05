package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestExecuteWithInvalidFlags(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()

		resetTimeout := 1 * time.Second
		pollingInterval := 2 * time.Second
		_, err := Execute(ctx, &config.Flags{
			ConfigUrl:             "http://localhost/fake-url",
			KeyPairConfigPath:     "../boyar/config/test/fake-key-pair.json",
			PollingInterval:       pollingInterval,
			BootstrapResetTimeout: resetTimeout,
		}, logger)
		require.EqualError(t, err, "invalid configuration: bootstrap reset timeout is less or equal to config polling interval")
	})
}

// FIXME get to the bottom of docker socket issues
func TestExecuteWithInvalidConfig(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		successfulAttempts := 0
		maxSuccessfulAttempts := 3
		server := helpers.CreateHttpServer("/", 0, func(writer http.ResponseWriter, request *http.Request) {
			if successfulAttempts < maxSuccessfulAttempts {
				successfulAttempts++
				writer.Write([]byte(`{"orchestrator":{}}`))
			} else {
				writer.Write([]byte("{}"))
			}
		})
		server.Start()
		defer server.Shutdown()

		logger := log.GetLogger()
		executionCtx := context.Background()

		startTime := time.Now()

		resetTimeout := 1 * time.Second
		pollingInterval := 100 * time.Millisecond
		waiter, err := Execute(ctx, &config.Flags{
			ConfigUrl:             server.Url(),
			KeyPairConfigPath:     "../boyar/config/test/fake-key-pair.json",
			PollingInterval:       pollingInterval,
			BootstrapResetTimeout: resetTimeout,
		}, logger)
		require.NoError(t, err)

		waiter.WaitUntilShutdown(executionCtx)

		require.InDelta(t, resetTimeout, time.Since(startTime).Nanoseconds(), float64(4*pollingInterval.Nanoseconds()))
	})
}
