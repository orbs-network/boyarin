package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

const PublicKey = "cfc9e5189223aedce9543be0ef419f89aaa69e8b"
const PrivateKey = "c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8"

func TestE2ERunSingleVirtualChain(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	vc1 := VChainArgument{Id: 42}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		vcChannel := readOnlyChannel(vc1)
		flags, cleanup := SetupDynamicBoyarDependencies(t, keys, genesisValidators(NETWORK_KEY_CONFIG), vcChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVolumeExists(t, ctx, "cfc9e5189223aedce9543be0ef419f89aaa69e8b-42-blocks")
			AssertVolumeExists(t, ctx, "cfc9e5-chain-42-logs")
			AssertVolumeExists(t, ctx, "cfc9e5-chain-42-status")
			AssertVolumeExists(t, ctx, "cfc9e5-signer-cache")
			AssertVolumeExists(t, ctx, "cfc9e5-signer-logs")
			AssertVolumeExists(t, ctx, "cfc9e5-signer-status")

			AssertVolumeExists(t, ctx, "boyar-logs")
			AssertVolumeExists(t, ctx, "boyar-status")
		})

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertServiceUp(t, ctx, "cfc9e5-signer")

			AssertVchainStatusExists(t, 80, vc1)
			AssertVchainLogsExist(t, 80, vc1)

			AssertServiceStatusExists(t, 80, "signer")
			AssertServiceLogsExist(t, 80, "signer")
		})

		return
	})
}

func TestE2EAddVirtualChain(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	vc1 := VChainArgument{Id: 42}
	vc2 := VChainArgument{Id: 45}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		vChainsChannel := make(chan []VChainArgument)
		defer close(vChainsChannel)

		flags, cleanup := SetupDynamicBoyarDependencies(t, keys, genesisValidators(NETWORK_KEY_CONFIG), vChainsChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		vChainsChannel <- []VChainArgument{vc1}

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainDown(t, 80, vc2)
		})

		vChainsChannel <- []VChainArgument{vc1, vc2}

		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainUp(t, 80, PublicKey, vc2)
		})
		return
	})
}

func TestE2ERemoveVirtualChain(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	vc1 := VChainArgument{Id: 42}
	vc2 := VChainArgument{Id: 45}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}
		vChainsChannel := make(chan []VChainArgument)
		defer close(vChainsChannel)

		flags, cleanup := SetupDynamicBoyarDependencies(t, keys, genesisValidators(NETWORK_KEY_CONFIG), vChainsChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		logger.Info(fmt.Sprintf("adding vchains %d and %d", vc1.Id, vc2.Id))
		vChainsChannel <- []VChainArgument{vc1, vc2}
		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainUp(t, 80, PublicKey, vc2)
		})

		vc2.Disabled = true
		logger.Info(fmt.Sprintf("removing vchain %d", vc2.Id))
		vChainsChannel <- []VChainArgument{vc1, vc2}
		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainDown(t, 80, vc2)
			AssertServiceDown(t, ctx, "cfc9e5-chain-45")
		})
		return
	})
}

func efsDir(dir string) string {
	return filepath.Join(adapter.DEFAULT_EFS_PATH, dir)
}

func TestE2EPurgeVirtualChain(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	vc1 := VChainArgument{Id: 42}
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
		deps.storageDriver = "local"
		deps.storageMountType = "bind"

		dirs := []string{
			efsDir("cfc9e5189223aedce9543be0ef419f89aaa69e8b-42-blocks"),
			efsDir("cfc9e5-chain-42-status"),
			efsDir("cfc9e5-chain-42-logs"),
			efsDir("cfc9e5-chain-42-cache"),
		}

		for _, dir := range dirs {
			os.RemoveAll(dir)
		}

		flags, cleanup := SetupDynamicBoyarDepencenciesForNetwork(t, deps, vChainsChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		logger.Info(fmt.Sprintf("adding vchain %d", vc1.Id))
		vChainsChannel <- []VChainArgument{vc1}
		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			require.True(t, helpers.VerifyFilesExist(t, dirs...))
		})

		vc1.Disabled = true
		vc1.PurgeData = true

		logger.Info(fmt.Sprintf("disabling and purging vchain %d", vc1.Id))
		vChainsChannel <- []VChainArgument{vc1}
		helpers.RequireEventually(t, DEFAULT_VCHAIN_TIMEOUT*2, func(t helpers.TestingT) {
			AssertVchainDown(t, 80, vc1)
			require.False(t, helpers.VerifyFilesExist(t, dirs...))
		})
		return
	})
}
