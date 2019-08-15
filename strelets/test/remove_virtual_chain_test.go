package test

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStrelets_RemoveVirtualChain(t *testing.T) {
	orchestrator, runner, _ := NewOrchestratorAndRunnerMocks()
	runner.FailedAttempts = 0

	s := strelets.NewStrelets(orchestrator)

	ctx := context.Background()
	orchestrator.On("RemoveContainer", ctx, "orbs-network-chain-1972-stack").Return(nil);

	err := s.RemoveVirtualChain(ctx, &strelets.RemoveVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id: strelets.VirtualChainId(1972),
			DockerConfig: strelets.DockerConfig{
				ContainerNamePrefix: "orbs-network",
			},
		},
	})

	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "RemoveContainer", 1)

}