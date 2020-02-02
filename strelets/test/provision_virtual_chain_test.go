package test

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func getKeyPairConfig() []byte {
	cfg, _ := config.NewKeysConfig("./fixtures/keys.json")
	return cfg.JSON(false)
}

func TestStrelets_ProvisionVirtualChain(t *testing.T) {
	orchestrator := &OrchestratorMock{}
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
	orchestrator.AssertExpectations(t)
}

func TestStrelets_ProvisionVirtualChainWithRetries(t *testing.T) {
	orchestrator := &OrchestratorMock{}
	// two failures followed by a success
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("some error")).Times(2)
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

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
	orchestrator.AssertExpectations(t)
}

func TestStrelets_ProvisionVirtualChainWhenDisabled(t *testing.T) {
	orchestrator := &OrchestratorMock{}

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
	orchestrator.AssertExpectations(t)
}
