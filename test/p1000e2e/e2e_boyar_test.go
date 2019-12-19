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

func TestE2ESingleVchainConfigurationMetrics(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		helpers.WithLogging(t, func(h *helpers.LoggingHarness) {
			//logger := h.Logger
			logger := log.GetLogger()
			helpers.InitSwarmEnvironment(t, ctx)
			keys := KeyConfig{
				NodeAddress:    PublickKey,
				NodePrivateKey: PrivateKey,
			}

			go InProcessBoyar(t, logger, keys, VChainId)

			helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
				metrics := GetVChainMetrics(t, VChainId)
				require.Equal(t, metrics.String("Node.Address"), PublickKey)
			})
		})
	})
}
