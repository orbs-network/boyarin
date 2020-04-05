package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"testing"
	"time"
)

func TestE2EEthereumClient(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupConfigServer(t, keys, func() *string {
			config := `
{
  "network": [],
  "orchestrator": {
    "storage-driver": "local"
  },
  "chains": [],
  "services": {
    "ethereum-client": {
      "InternalHttpPort": 8545,
      "InternalPort": 30303,
      "ExternalPort": 30303,
      "DockerConfig": {
        "Image":  "orbsnetwork/ethereum-light-client-service",
        "Tag":    "latest",
        "Pull":   false
      },
      "Config": {
        "api": "v1"
      }
    }
  }
}
`
			return &config
		})
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, 3*time.Minute, func(t helpers.TestingT) {
			AssertServiceUp(t, ctx, "cfc9e5-ethereum-client")
		})
		return
	})
}
