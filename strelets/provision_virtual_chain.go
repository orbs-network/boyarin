package strelets

import (
	"context"
	"fmt"
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
	chain := input.VirtualChain
	imageName := chain.DockerConfig.FullImageName()
	if chain.DockerConfig.Pull {
		if err := s.docker.PullImage(ctx, imageName); err != nil {
			return err
		}
	}

	containerName := chain.getContainerName()
	vchainVolumes := chain.getDockerVolumes(s.root)

	vchainVolumes.createDirs()

	if err := copyFile(input.KeysConfigPath, vchainVolumes.keyPairConfigFile); err != nil {
		return err
	}

	if err := ioutil.WriteFile(vchainVolumes.networkConfigFile, getNetworkConfigJSON(input.Peers), 0644); err != nil {
		return err
	}

	exposedPorts, portBindings := buildDockerNetworkOptions(chain.HttpPort, chain.GossipPort)
	dockerConfig := buildDockerConfig(imageName, exposedPorts, portBindings, vchainVolumes)

	containerId, err := s.docker.RunContainer(ctx, imageName, containerName, dockerConfig)
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}
