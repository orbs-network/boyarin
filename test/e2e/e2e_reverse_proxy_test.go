package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestE2ERunSingleVirtualChainWithSSL(t *testing.T) {
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

		flags.SSLCertificatePath = "./fixtures/cert.pem"
		flags.SSLPrivateKeyPath = "./fixtures/key.pem"

		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)

			url := fmt.Sprintf("https://127.0.0.1:443/vchains/%d", vc1.Id)
			fmt.Println(url)

			client := http.Client{
				Timeout: 2 * time.Second,
			}
			_, err := client.Get(url)
			require.Error(t, err)
			require.Errorf(t, err, "Get %s: x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs", url)
		})
		return
	})
}
