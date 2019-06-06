package test

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func getKeyPairConfig() []byte {
	cfg, _ := config.NewKeysConfig("./fixtures/keys.json")
	return cfg.JSON(false)
}

func TestStrelets_ProvisionVirtualChain(t *testing.T) {
	orchestrator, runner, _ := NewOrchestratorAndRunnerMocks()
	runner.FailedAttempts = 0

	s := strelets.NewStrelets(orchestrator)

	peers := make(strelets.PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
		},
		Peers:         &peers,
		KeyPairConfig: getKeyPairConfig(),
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestStrelets_ProvisionVirtualChainWithRetries(t *testing.T) {
	orchestrator, runner, _ := NewOrchestratorAndRunnerMocks()
	runner.FailedAttempts = 2

	s := strelets.NewStrelets(orchestrator)

	peers := make(strelets.PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
		},
		Peers:         &peers,
		KeyPairConfig: getKeyPairConfig(),
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	runner.AssertNumberOfCalls(t, "Run", 3)
}

func TestStrelets_ProvisionVirtualChainWhenDisabled(t *testing.T) {
	orchestrator, runner, _ := NewOrchestratorAndRunnerMocks()
	runner.FailedAttempts = 0

	s := strelets.NewStrelets(orchestrator)

	peers := make(strelets.PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:       strelets.VirtualChainId(1972),
			Disabled: true,
		},
		Peers:         &peers,
		KeyPairConfig: getKeyPairConfig(),
	})

	require.Error(t, err, "virtual chain 1972 is disabled")
	orchestrator.AssertNumberOfCalls(t, "Prepare", 0)
	runner.AssertNumberOfCalls(t, "Run", 0)
}
