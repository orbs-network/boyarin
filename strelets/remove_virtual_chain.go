package strelets

import (
	"context"
	"github.com/docker/docker/client"
)

type RemoveVirtualChainInput struct {
	Chain        VirtualChainId
	DockerConfig *DockerImageConfig
}

func (s *strelets) RemoveVirtualChain(input *RemoveVirtualChainInput) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		return err
	}

	v := &vchain{
		id:           input.Chain,
		dockerConfig: input.DockerConfig,
	}

	containerName := v.getContainerName()
	return s.removeContainer(ctx, cli, containerName)
}
