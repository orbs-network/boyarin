package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestE2EBootstrapWithDefaultConfig(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoostrapDependencies(t, keys)
		defer cleanup()

		err := services.Bootstrap(ctx, flags, logger)
		require.NoError(t, err)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertServiceUp(t, ctx, "cfc9e5-config-service-stack")
		})
		return
	})
}
