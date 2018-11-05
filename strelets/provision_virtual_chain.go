package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"io/ioutil"
)

type ProvisionVirtualChainInput struct {
	VirtualChain   *VirtualChain
	KeysConfigPath string
	Peers          *PeersMap
}

type Peer struct {
	IP   string
	Port int
}

type PublicKey string

type PeersMap map[PublicKey]*Peer

func (s *strelets) ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error {
	docker, err := adapter.NewDockerAPI()
	if err != nil {
		return err
	}

	chain := input.VirtualChain
	imageName := chain.DockerConfig.FullImageName()
	if chain.DockerConfig.Pull {
		if err := docker.PullImage(ctx, imageName); err != nil {
			return err
		}
	}

	containerName := chain.getContainerName()
	vchainVolumes := chain.getDockerVolumes(s.root)

	createDir(vchainVolumes.configRoot)
	createDir(vchainVolumes.logs)

	if err := copyFile(input.KeysConfigPath, vchainVolumes.config); err != nil {
		return err
	}

	if err := ioutil.WriteFile(vchainVolumes.network, getNetworkConfigJSON(input.Peers), 0644); err != nil {
		return err
	}

	exposedPorts, portBindings := buildDockerNetworkOptions(chain.HttpPort, chain.GossipPort)
	dockerConfig := buildDockerConfig(imageName, exposedPorts, portBindings, vchainVolumes)

	containerId, err := docker.RunContainer(ctx, imageName, containerName, dockerConfig)
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}
