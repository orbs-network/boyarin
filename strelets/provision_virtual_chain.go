package strelets

import (
	"context"
	"fmt"
	"io/ioutil"
)

type ProvisionVirtualChainInput struct {
	VirtualChain      *VirtualChain
	KeyPairConfigPath string
	Peers             *PeersMap
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

	if err := input.prepareVirtualChainConfig(s.root); err != nil {
		return err
	}

	containerId, err := s.docker.RunContainer(ctx, chain.getContainerName(), chain.getContainerConfig(s.root))
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}

func (input *ProvisionVirtualChainInput) prepareVirtualChainConfig(root string) error {
	vchainVolumes := input.VirtualChain.getContainerVolumes(root)
	vchainVolumes.createDirs()

	if err := copyFile(input.KeyPairConfigPath, vchainVolumes.keyPairConfigFile); err != nil {
		return err
	}

	if err := ioutil.WriteFile(vchainVolumes.networkConfigFile, getNetworkConfigJSON(input.Peers), 0644); err != nil {
		return err
	}

	return nil
}
