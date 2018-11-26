package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
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
		if err := s.orchestrator.PullImage(ctx, imageName); err != nil {
			return err
		}
	}

	keyPair, err := ioutil.ReadFile(input.KeyPairConfigPath)
	if err != nil {
		return fmt.Errorf("could not read key pair config: %s at %s", err, input.KeyPairConfigPath)
	}

	if runner, err := s.orchestrator.Prepare(ctx, imageName, chain.getContainerName(), chain.HttpPort, chain.GossipPort, &adapter.AppConfig{
		KeyPair: keyPair,
		Network: getNetworkConfigJSON(input.Peers),
	}); err != nil {
		return err
	} else {
		return runner.Run(ctx)
	}
}
