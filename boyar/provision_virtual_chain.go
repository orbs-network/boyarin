package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/boyar/topology"
	"github.com/orbs-network/boyarin/log_types"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"sort"
)

var removed = &utils.HashedValue{Value: "foo"}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	chains := b.config.Chains()

	for _, chain := range chains {
		if chain.Disabled {
			if err := b.removeVirtualChain(ctx, chain); err != nil {
				b.logger.Error("failed to remove virtual chain",
					log_types.VirtualChainId(int64(chain.Id)),
					log.Error(err))
			} else {
				b.logger.Info("removed virtual chain", log_types.VirtualChainId(int64(chain.Id)))
			}
		} else {
			if err := b.provisionVirtualChain(ctx, chain); err != nil {
				b.logger.Error("failed to apply virtual chain configuration",
					log_types.VirtualChainId(int64(chain.Id)),
					log.Error(err))
			} else {
				data, _ := json.Marshal(chain)
				b.logger.Info("updated virtual chain configuration",
					log_types.VirtualChainId(int64(chain.Id)),
					log.String("configuration", string(data)))
			}
		}
	}

	return nil
}

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *config.VirtualChain) error {
	peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

	signerOn := b.config.Services().SignerOn()
	keyPairConfig := getKeyConfigJson(b.config, signerOn)

	input := &config.VirtualChainCompositeKey{
		VirtualChain:  chain,
		Peers:         peers,
		NodeAddress:   b.config.NodeAddress(),
		KeyPairConfig: keyPairConfig,
	}

	if b.cache.vChains.CheckNewJsonValue(chain.Id.String(), input) {
		imageName := chain.DockerConfig.FullImageName()

		if chain.DockerConfig.Pull {
			if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
				return fmt.Errorf("could not pull docker image: %b", err)
			}
		}

		serviceConfig := &adapter.ServiceConfig{
			Id:            uint32(chain.Id),
			NodeAddress:   string(b.config.NodeAddress()),
			ImageName:     imageName,
			ContainerName: chain.GetContainerName(),
			HttpPort:      chain.HttpPort,
			GossipPort:    chain.GossipPort,

			LimitedMemory:  chain.DockerConfig.Resources.Limits.Memory,
			LimitedCPU:     chain.DockerConfig.Resources.Limits.CPUs,
			ReservedMemory: chain.DockerConfig.Resources.Reservations.Memory,
			ReservedCPU:    chain.DockerConfig.Resources.Reservations.CPUs,

			BlocksVolumeSize: chain.DockerConfig.Volumes.Blocks,
			LogsVolumeSize:   chain.DockerConfig.Volumes.Logs,
		}

		appConfig := &adapter.AppConfig{
			KeyPair: keyPairConfig,
			Network: getNetworkConfigJSON(b.config.FederationNodes()),
			Config:  chain.GetSerializedConfig(),
		}

		// skip keys
		if err := b.orchestrator.RunVirtualChain(ctx, serviceConfig, appConfig); err != nil {
			b.cache.vChains.Clear(chain.Id.String())
			return err
		}
	}

	return nil
}

func (b *boyar) removeVirtualChain(ctx context.Context, chain *config.VirtualChain) error {
	if b.cache.vChains.CheckNewValue(chain.Id.String(), removed) {
		serviceName := adapter.GetServiceId(chain.GetContainerName())
		if err := b.orchestrator.ServiceRemove(ctx, serviceName); err != nil {
			b.cache.vChains.Clear(chain.Id.String())
			return err
		}
	}

	return nil
}

func getNetworkConfigJSON(nodes []*topology.FederationNode) []byte {
	jsonMap := make(map[string]interface{})

	// A workaround for tests because range does not preserve key order over iteration
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Address > nodes[j].Address
	})

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}
