package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStrelets_ProvisionVirtualChain(t *testing.T) {
	orchestrator := &OrchestratorMock{}
	s := strelets.NewStrelets(orchestrator)

	runner := &RunnerMock{FailedAttempts: 0}
	runner.On("Run", mock.Anything).Return(nil)

	orchestrator.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(runner, nil)

	peers := make(strelets.PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
		},
		Peers:             &peers,
		KeyPairConfigPath: "./fixtures/keys.json",
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestStrelets_ProvisionVirtualChainWithRetries(t *testing.T) {
	orchestrator := &OrchestratorMock{}
	s := strelets.NewStrelets(orchestrator)

	runner := &RunnerMock{FailedAttempts: 2}
	runner.On("Run", mock.Anything).Return(nil)

	orchestrator.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(runner, nil)

	peers := make(strelets.PeersMap)
	err := s.ProvisionVirtualChain(context.Background(), &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
		},
		Peers:             &peers,
		KeyPairConfigPath: "./fixtures/keys.json",
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	runner.AssertNumberOfCalls(t, "Run", 3)
}
