package p1000e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"testing"
	"time"
)

const PublicKey = "cfc9e5189223aedce9543be0ef419f89aaa69e8b"
const PrivateKey = "c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8"

func TestE2ERunSingleVirtualChain(t *testing.T) {
	vc1 := VChainArgument{Id: 42}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, vc1)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertServiceUp(t, ctx, "cfc9e5-signer-service-stack")
		})
		return
	})
}

func TestE2ERunMultipleVirtualChains(t *testing.T) {
	vc1 := VChainArgument{Id: 42}
	vc2 := VChainArgument{Id: 45}
	helpers.WithContextAndShutdown(func(ctx context.Context) (waiter govnr.ShutdownWaiter) {
		logger := log.GetLogger()
		helpers.InitSwarmEnvironment(t, ctx)
		keys := KeyConfig{
			NodeAddress:    PublicKey,
			NodePrivateKey: PrivateKey,
		}

		flags, cleanup := SetupBoyarDependencies(t, keys, vc1, vc2)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainUp(t, 80, PublicKey, vc2)
		})
		return
	})
}

func TestE2EAddVirtualChain(t *testing.T) {
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

		flags, cleanup := SetupDynamicBoyarDependencies(t, keys, vChainsChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		logger.Info(fmt.Sprintf("adding vchain %d", vc1.Id))
		vChainsChannel <- []VChainArgument{vc1}
		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
		})

		logger.Info(fmt.Sprintf("adding vchain %d", vc2.Id))
		vChainsChannel <- []VChainArgument{vc1, vc2}
		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainUp(t, 80, PublicKey, vc2)
		})
		return
	})
}

func TestE2EAddAndRemoveVirtualChain(t *testing.T) {
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

		flags, cleanup := SetupDynamicBoyarDependencies(t, keys, vChainsChannel)
		defer cleanup()
		waiter = InProcessBoyar(t, ctx, logger, flags)

		logger.Info(fmt.Sprintf("adding vchain %d", vc1.Id))
		vChainsChannel <- []VChainArgument{vc1}
		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
		})

		logger.Info(fmt.Sprintf("adding vchain %d", vc2.Id))
		vChainsChannel <- []VChainArgument{vc1, vc2}
		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainUp(t, 80, PublicKey, vc2)
		})

		vc2.Disabled = true
		logger.Info(fmt.Sprintf("adding vchain %d", vc2.Id))
		vChainsChannel <- []VChainArgument{vc1, vc2}
		helpers.RequireEventually(t, 20*time.Second, func(t helpers.TestingT) {
			AssertVchainUp(t, 80, PublicKey, vc1)
			AssertVchainDown(t, 80, vc2)
		})
		return
	})
}
