package p1000e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"testing"
	"time"
)

var NETWORK_KEY_CONFIG = []KeyConfig{
	{
		"a328846cd5b4979d68a8c58a9bdfeee657b34de7",
		"901a1a0bfbe217593062a054e561e708707cb814a123474c25fd567a0fe088f8",
	},
	{
		"d27e2e7398e2582f63d0800330010b3e58952ff6",
		"87a210586f57890ae3642c62ceb58f0f0a54e787891054a5a54c80e1da418253",
	},
	{
		"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9",
		"426308c4d11a6348a62b4fdfb30e2cad70ab039174e2e8ea707895e4c644c4ec",
	},
	{
		"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17",
		"1e404ba4e421cedf58dcc3dddcee656569afc7904e209612f7de93e1ad710300",
	},
}

func TestE2ERunFullNetwork(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
	})

	for i, keys := range NETWORK_KEY_CONFIG[:1] {
		helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
			logger := log.GetLogger().WithTags(log.Int("node", i))

			vc1 := VChainArgument{Id: 42}
			flags, cleanup := SetupBoyarDependencies(t, keys, vc1)
			defer cleanup()
			waiter = InProcessBoyar(t, ctx, logger, flags)

			helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
				AssertVchainUp(t, uint32(80+i), keys.NodeAddress, vc1)
			})

			return
		})
	}
}
