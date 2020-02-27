package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestE2EBootstrapWithDefaultConfig(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	vc1 := VChainArgument{Id: 42}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, genesisValidators(NETWORK_KEY_CONFIG), vc1)
		defer cleanup()

		err := services.Bootstrap(ctx, flags, logger)
		require.NoError(t, err)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertServiceUp(t, ctx, "cfc9e5-config-service-stack")
		})

		newFlags := &config.Flags{
			ConfigUrl:         fmt.Sprintf("http://%s:7666", helpers.LocalIP()),
			KeyPairConfigPath: flags.KeyPairConfigPath,
			Timeout:           time.Minute,
			PollingInterval:   500 * time.Millisecond,
		}
		waiter, err = services.Execute(ctx, newFlags, logger)
		require.NoError(t, err)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
		})

		return
	})
}
