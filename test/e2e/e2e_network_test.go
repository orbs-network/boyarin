package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestE2ERunFullNetwork(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
	})

	const NUM_VCS = 4
	var vcs []VChainArgument
	for i := 0; i < NUM_VCS; i++ {
		vcs = append(vcs, VChainArgument{
			Id:       42,
			BasePort: basePort + 1000*i,
		})
	}

	topology := buildTopology(NETWORK_KEY_CONFIG, 42)

	for i, keys := range NETWORK_KEY_CONFIG {
		go func(i int, keys KeyConfig) {
			helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
				logger := log.GetLogger().WithTags(log.Int("node", i))

				vc := vcs[i]
				httpPort := basePort*2 + 1000*i // very ad hoc
				deps := boyarDependencies{
					keyPair:           keys,
					topology:          topology,
					genesisValidators: genesisValidators(NETWORK_KEY_CONFIG),
					httpPort:          httpPort,
				}

				vChainsChannel := readOnlyChannel(vc)
				flags, cleanup := SetupDynamicBoyarDepencenciesForNetwork(t, deps, helpers.PRODUCTION_DOCKER_REGISTRY_AND_USER, vChainsChannel)
				defer cleanup()

				waiter = InProcessBoyar(t, ctx, logger, flags)

				helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
					AssertVchainUp(t, httpPort, keys.NodeAddress, vc)
				})

				return
			})
		}(i, keys)
	}

	helpers.RequireEventually(t, NUM_VCS*DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
		metrics := GetVChainMetrics(t, basePort*2, vcs[0])
		require.GreaterOrEqual(t, uint64(3), metrics.Uint64("BlockStorage.BlockHeight"))
	})

}
