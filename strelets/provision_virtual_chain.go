package strelets

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"io/ioutil"
)

type ProvisionVirtualChainInput struct {
	Chain            VirtualChainId
	VchainConfigPath string
	HttpPort         int
	GossipPort       int
	DockerConfig     *DockerImageConfig
	Peers            map[PublicKey]*Peer
}

func (s *strelets) ProvisionVirtualChain(input *ProvisionVirtualChainInput) error {
	v := &vchain{
		id:           input.Chain,
		httpPort:     input.HttpPort,
		gossipPort:   input.GossipPort,
		dockerConfig: input.DockerConfig,
	}
	s.vchains[input.Chain] = v

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		return err
	}

	imageName := v.dockerConfig.FullImageName()
	if v.dockerConfig.Pull {
		pullImage(ctx, cli, imageName)
	}

	containerName := v.getContainerName()
	vchainVolumes := s.prepareVirtualChainConfig(containerName)

	createDir(vchainVolumes.configRoot)
	createDir(vchainVolumes.logs)

	if err := copyNodeConfig(input.VchainConfigPath, vchainVolumes.config); err != nil {
		return err
	}

	if err := ioutil.WriteFile(vchainVolumes.network, getNetworkConfigJSON(s.peers), 0644); err != nil {
		return err
	}

	containerId, err := s.runContainer(ctx, cli, containerName, imageName, v, vchainVolumes)
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}
