package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func calcSha256(input []byte) []byte {
	sum := sha256.Sum256(input)
	return sum[:]
}

func calcSha256FromFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return calcSha256(data), nil
}

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
		vChainsChannel := make(chan []VChainArgument)
		defer close(vChainsChannel)

		deps := defaultBoyarDependencies(keys, genesisValidators(NETWORK_KEY_CONFIG))
		deps.binaryUrl = "https://github.com/orbs-network/boyarin/releases/download/v1.4.0/boyar-v1.4.0.bin"
		deps.binarySha256 = "1998cc1f7721acfe1954ab2878cc0ad8062cd6d919cd61fa22401c6750e195fe"

		flags, cleanup := SetupDynamicBoyarDepencenciesForNetwork(t, deps, vChainsChannel)
		flags.TargetPath = targetPath
		flags.AutoUpdate = true
		flags.ShutdownAfterUpdate = true

		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			expectedChecksum, _ := hex.DecodeString(deps.binarySha256)
			require.FileExists(t, targetPath)

			checksum, err := calcSha256FromFile(targetPath)
			require.NoError(t, err)

			require.EqualValues(t, expectedChecksum, checksum)
		})

		return
	})
}
