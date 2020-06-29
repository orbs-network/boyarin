package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestE2ERunFullNetwork(t *testing.T) {
	//helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
	})

	var vcs []VChainArgument
	for i := 0; i < 4; i++ {
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
				httpPort := basePort*2 + 1000*i
				flags, cleanup := SetupBoyarDependenciesForNetwork(t, keys, topology, genesisValidators(NETWORK_KEY_CONFIG), httpPort, vc)
				defer cleanup()

				waiter = InProcessBoyar(t, ctx, logger, flags)

				helpers.RequireEventually(t, 1*time.Minute, func(t helpers.TestingT) {
					AssertVchainUp(t, httpPort, keys.NodeAddress, vc)
				})

				return
			})
		}(i, keys)
	}

	helpers.RequireEventually(t, 3*time.Minute, func(t helpers.TestingT) {
		metrics := GetVChainMetrics(t, basePort*2, vcs[0])
		require.EqualValues(t, 3, metrics.Uint64("BlockStorage.BlockHeight"))
	})

}
