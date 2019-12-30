package p1000e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const PublickKey = "cfc9e5189223aedce9543be0ef419f89aaa69e8b"
const PrivateKey = "c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8"

func TestE2ERunSingleVirtualChainNodeWithCorrectPortIdAndKey(t *testing.T) {
	vc1 := VChainArgument{Id: 42}
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublickKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, vc1)
		defer cleanup()
		go InProcessBoyar(t, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			metrics := GetVChainMetrics(t, vc1)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			AssertGossipServer(t, vc1)
		})
	})
}

func TestE2ERunMultipleVirtualChainsNodeWithCorrectPortIdAndKey(t *testing.T) {
	vc1 := VChainArgument{Id: 42}
	vc2 := VChainArgument{Id: 45}
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublickKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, vc1, vc2)
		defer cleanup()
		go InProcessBoyar(t, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			metrics := GetVChainMetrics(t, vc1)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			metrics = GetVChainMetrics(t, vc2)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			AssertGossipServer(t, vc1)
			AssertGossipServer(t, vc2)
		})
	})
}
