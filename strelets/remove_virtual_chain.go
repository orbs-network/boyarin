package strelets

import (
	"context"
	"github.com/docker/docker/client"
)

type RemoveVirtualChainInput struct {
	VirtualChain *VirtualChain
}

func (s *strelets) RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error {
	cli, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))
	if err != nil {
		return err
	}

	containerName := input.VirtualChain.getContainerName()
	return s.removeContainer(ctx, cli, containerName)
}
