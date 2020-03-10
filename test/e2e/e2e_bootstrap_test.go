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

		defaultFlags, cleanup := SetupBoyarDependencies(t, keys, genesisValidators(NETWORK_KEY_CONFIG), vc1)
		defer cleanup()

		bootstrapConfig := fmt.Sprintf(`
{
  "orchestrator": {
    "DynamicManagementConfig": {
      "Url": "http://localhost:7666/node/management",
      "ReadInterval": "1m",
      "ResetTimeout": "30m"
    }
  },
  "services": {
    "management": {
      "InternalPort": 8080,
      "ExternalPort": 7666,
      "DockerConfig": {
        "Image":  "orbsnetwork/management",
        "Tag":    "latest",
        "Pull":   true
      },
      "Config": {
        "EthereumGenesisContract": "0x2384723487623784638272348",
        "EthereumEndpoint": "http://eth.orbs.com",
		"boyarLegacyBootstrap": "%s"
      }
    }
  }
}
`, defaultFlags.ConfigUrl)

		cfg, err := config.NewStringConfigurationSource(bootstrapConfig, "", defaultFlags.KeyPairConfigPath)
		require.NoError(t, err)

		err = services.Bootstrap(ctx, cfg, logger)
		require.NoError(t, err)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertServiceUp(t, ctx, "cfc9e5-management-service-stack")
		})

		flags := &config.Flags{
			ConfigUrl:         cfg.OrchestratorOptions().DynamicManagementConfig.Url,
			KeyPairConfigPath: defaultFlags.KeyPairConfigPath,
			Timeout:           time.Minute,
			PollingInterval:   500 * time.Millisecond,
		}

		waiter, err = services.Execute(ctx, flags, logger)
		require.NoError(t, err)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
		})

		return
	})
}
