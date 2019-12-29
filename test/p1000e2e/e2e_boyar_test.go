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
const VChainId = 42
const VChain2Id = 123

func TestE2ESingleVchainRunsVirtualChainNodeWithCorrectIdAndKey(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublickKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, VChainId)
		defer cleanup()
		go InProcessBoyar(t, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			metrics := GetVChainMetrics(t, VChainId)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			AssertGossipServer(t, VChainId)
		})
	})
}

func TestE2EMultipleVchainRunsVirtualChainNodeWithCorrectIdAndKey(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublickKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, VChainId, VChain2Id)
		defer cleanup()
		go InProcessBoyar(t, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			metrics := GetVChainMetrics(t, VChainId)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			metrics = GetVChainMetrics(t, VChain2Id)
			require.Equal(t, metrics.String("Node.Address"), PublickKey)
			AssertGossipServer(t, VChainId)
			AssertGossipServer(t, VChain2Id)
		})
	})
}
