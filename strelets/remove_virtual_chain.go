package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type RemoveVirtualChainInput struct {
	VirtualChain *VirtualChain
}

func (s *strelets) RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error {
	docker, err := adapter.NewDockerAPI()
	if err != nil {
		return err
	}

	containerName := input.VirtualChain.getContainerName()
	return docker.RemoveContainer(ctx, containerName)
}
