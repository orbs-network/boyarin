package strelets

import (
	"context"
)

type RemoveVirtualChainInput struct {
	VirtualChain *VirtualChain
}

func (s *strelets) RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error {
	containerName := input.VirtualChain.getContainerName()
	return s.docker.RemoveContainer(ctx, containerName)
}
