package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStrelets_RemoveVirtualChain(t *testing.T) {
	orchestrator := &OrchestratorMock{}
	orchestrator.On("ServiceRemove", mock.Anything, mock.Anything).Return(nil).Once()

	s := strelets.NewStrelets(orchestrator)

	ctx := context.Background()

	err := s.RemoveVirtualChain(ctx, &strelets.RemoveVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
			DockerConfig: strelets.DockerConfig{
				ContainerNamePrefix: "orbs-network",
			},
		},
	})

	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}
