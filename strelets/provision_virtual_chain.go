package strelets

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"io/ioutil"
	"time"
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

type NodeAddress string

type PeersMap map[NodeAddress]*Peer

const (
	PROVISION_VCHAIN_MAX_TRIES       = 5
	PROVISION_VCHAIN_ATTEMPT_TIMEOUT = 30 * time.Second
	PROVISION_VCHAIN_RETRY_INTERVAL  = 3 * time.Second
)

func (s *strelets) ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error {
	chain := input.VirtualChain
	imageName := chain.DockerConfig.FullImageName()
	if chain.DockerConfig.Pull {
		if err := s.orchestrator.PullImage(ctx, imageName); err != nil {
			return fmt.Errorf("could not pull docker image: %s", err)
		}
	}

	keyPair, err := ioutil.ReadFile(input.KeyPairConfigPath)
	if err != nil {
		return fmt.Errorf("could not read key pair config: %s at %s", err, input.KeyPairConfigPath)
	}

	return Try(ctx, PROVISION_VCHAIN_MAX_TRIES, PROVISION_VCHAIN_ATTEMPT_TIMEOUT, PROVISION_VCHAIN_RETRY_INTERVAL,
		func(ctxWithTimeout context.Context) error {
			serviceConfig := &adapter.ServiceConfig{
				ImageName:     imageName,
				ContainerName: chain.getContainerName(),
				HttpPort:      chain.HttpPort,
				GossipPort:    chain.GossipPort,

				LimitedMemory:  chain.DockerConfig.Resources.Limits.Memory,
				LimitedCPU:     chain.DockerConfig.Resources.Limits.CPUs,
				ReservedMemory: chain.DockerConfig.Resources.Reservations.Memory,
				ReservedCPU:    chain.DockerConfig.Resources.Reservations.CPUs,
			}

			appConfig := &adapter.AppConfig{
				KeyPair: keyPair,
				Network: getNetworkConfigJSON(input.Peers),
				Config:  chain.getSerializedConfig(),
			}

			if runner, err := s.orchestrator.Prepare(ctx, serviceConfig, appConfig); err != nil {
				return err
			} else {
				return runner.Run(ctx)
			}
		})
}
