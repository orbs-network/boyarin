package strelets

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io/ioutil"
)

type ProvisionVirtualChainInput struct {
	VirtualChain           *VirtualChain
	VirtualChainConfigPath string
	Peers                  *PeersMap
}

type Peer struct {
	IP   string
	Port int
}

type PublicKey string

type PeersMap map[PublicKey]*Peer

func (s *strelets) ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error {
	cli, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))
	if err != nil {
		return err
	}

	chain := input.VirtualChain
	imageName := chain.DockerConfig.FullImageName()
	if chain.DockerConfig.Pull {
		pullImage(ctx, cli, imageName)
	}

	containerName := chain.getContainerName()
	vchainVolumes := chain.getDockerVolumes(s.root)

	createDir(vchainVolumes.configRoot)
	createDir(vchainVolumes.logs)

	if err := copyVirtualChainConfig(input.VirtualChainConfigPath, vchainVolumes.config); err != nil {
		return err
	}

	if err := ioutil.WriteFile(vchainVolumes.network, getNetworkConfigJSON(input.Peers), 0644); err != nil {
		return err
	}

	exposedPorts, portBindings := buildDockerNetworkOptions(chain.HttpPort, chain.GossipPort)
	dockerConfig := buildDockerConfig(imageName, exposedPorts, portBindings, vchainVolumes)

	containerId, err := s.runContainer(ctx, cli, containerName, imageName, dockerConfig)
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}
