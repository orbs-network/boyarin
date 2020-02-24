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

	var errors []error
	for _, chain := range chains {
		containerName := b.config.PrefixedContainerName("chain-" + chain.Id.String())

		if chain.Disabled {
			if b.cache.vChains.CheckNewValue(containerName, removed) {
				serviceName := adapter.GetServiceId(chain.GetContainerName())
				if err := b.orchestrator.ServiceRemove(ctx, serviceName); err != nil {
					b.cache.vChains.Clear(chain.Id.String())

					b.logger.Error("failed to remove virtual chain",
						log_types.VirtualChainId(int64(chain.Id)),
						log.Error(err))
					errors = append(errors, err)
				}
			} else {
				b.logger.Info("removed virtual chain", log_types.VirtualChainId(int64(chain.Id)))

			}
		} else {
			input := getVirtualChainConfig(b.config, chain)

			if b.cache.vChains.CheckNewJsonValue(containerName, input) {
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
					ContainerName: containerName,
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
					KeyPair: input.KeyPairConfig,
					Network: getNetworkConfigJSON(b.config.FederationNodes()),
					Config:  chain.GetSerializedConfig(),
				}

				if err := b.orchestrator.RunVirtualChain(ctx, serviceConfig, appConfig); err != nil {
					b.cache.vChains.Clear(containerName)
					b.logger.Error("failed to apply virtual chain configuration",
						log_types.VirtualChainId(int64(chain.Id)),
						log.Error(err))
					errors = append(errors, err)
				} else {
					data, _ := json.Marshal(chain)
					b.logger.Info("updated virtual chain configuration",
						log_types.VirtualChainId(int64(chain.Id)),
						log.String("configuration", string(data)))
				}
			}
		}
	}

	return utils.AggregateErrors(errors)
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

func buildPeersMap(nodes []*topology.FederationNode, gossipPort int) *config.PeersMap {
	peersMap := make(config.PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[config.NodeAddress(node.Address)] = &config.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}

func getVirtualChainConfig(cfg config.NodeConfiguration, chain *config.VirtualChain) *config.VirtualChainConfig {
	peers := buildPeersMap(cfg.FederationNodes(), chain.GossipPort)
	keyPairConfig := getKeyConfigJson(cfg, true)

	return &config.VirtualChainConfig{
		VirtualChain:  chain,
		Peers:         peers,
		NodeAddress:   cfg.NodeAddress(),
		KeyPairConfig: keyPairConfig,
	}
}
