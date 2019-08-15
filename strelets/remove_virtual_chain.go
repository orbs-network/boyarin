package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type RemoveVirtualChainInput struct {
	VirtualChain *VirtualChain
}

func (s *strelets) RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error {
	serviceName := adapter.GetServiceId(input.VirtualChain.getContainerName())
	return s.orchestrator.RemoveContainer(ctx, serviceName)
}
