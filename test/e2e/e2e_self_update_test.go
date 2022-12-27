package e2e

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestE2ESelfUpdate(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	os.RemoveAll("./_tmp")
	os.MkdirAll("./_tmp", 0755)

	targetPath := filepath.Join("./_tmp", "boyar.bin")
	err := ioutil.WriteFile(targetPath, []byte("fake binary"), 0755)
	require.NoError(t, err)

	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		deps := defaultBoyarDependencies(keys, genesisValidators(NETWORK_KEY_CONFIG))
		deps.binaryUrl = "https://github.com/orbs-network/boyarin/releases/download/v1.4.0/boyar-v1.4.0.bin"
		deps.binarySha256 = "1998cc1f7721acfe1954ab2878cc0ad8062cd6d919cd61fa22401c6750e195fe"

		flags, cleanup := SetupDynamicBoyarDepencenciesForNetwork(t, deps)
		flags.BoyarBinaryPath = targetPath
		flags.AutoUpdate = true
		flags.ShutdownAfterUpdate = true

		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		return
	})
}
