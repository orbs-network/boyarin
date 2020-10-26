package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
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
		containerName := b.config.NamespacedContainerName(chain.GetContainerName())
		logger := b.logger.WithTags(log_types.VirtualChainId(int64(chain.Id)))

		if chain.Disabled {
			if key := containerName; b.cache.vChains.CheckNewJsonValue(key, removed) {
				if err := b.orchestrator.RemoveService(ctx, containerName); err != nil {
					b.cache.vChains.Clear(key)

					logger.Error("failed to remove virtual chain", log.Error(err))
					errors = append(errors, err)
				} else {
					logger.Info("successfully removed virtual chain", log_types.VirtualChainId(int64(chain.Id)))
				}
			}

			if key := containerName + "-data"; chain.PurgeData && b.cache.vChains.CheckNewJsonValue(key, removed) {
				if err := b.orchestrator.PurgeVirtualChainData(ctx, string(b.config.NodeAddress()), uint32(chain.Id), containerName); err != nil {
					b.cache.vChains.Clear(key)
					logger.Error("failed to purge vchain data", log.Error(err))
					errors = append(errors, err)
				} else {
					logger.Info("successfully purged vchain data")
				}
			}

			continue
		}

		input := getVirtualChainConfig(b.config, chain)
		if key := containerName; b.cache.vChains.CheckNewJsonValue(key, input) {
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
				InternalPort:  chain.InternalPort,
				ExternalPort:  chain.ExternalPort,

				AllowAccessToSigner:     true,
				HTTPProxyNetworkEnabled: true,
				AllowAccessToServices:   true,

				LimitedMemory:  chain.DockerConfig.Resources.Limits.Memory,
				LimitedCPU:     chain.DockerConfig.Resources.Limits.CPUs,
				ReservedMemory: chain.DockerConfig.Resources.Reservations.Memory,
				ReservedCPU:    chain.DockerConfig.Resources.Reservations.CPUs,
			}

			appConfig := &adapter.AppConfig{
				KeyPair: input.KeyPairConfig,
				Network: getNetworkConfigJSON(overrideTopologyPort(b.config.FederationNodes(), chain.ExternalPort)),
				Config:  chain.GetSerializedConfig(),
			}

			if err := b.orchestrator.RunVirtualChain(ctx, serviceConfig, appConfig); err != nil {
				b.cache.vChains.Clear(key)
				logger.Error("failed to apply virtual chain configuration", log.Error(err))
				errors = append(errors, err)
			} else {
				data, _ := json.Marshal(chain)
				logger.Info("updated virtual chain configuration", log.String("configuration", string(data)))
			}
		}
	}

	return utils.AggregateErrors(errors)
}

func getNetworkConfigJSON(nodes []*config.FederationNode) []byte {
	jsonMap := make(map[string]interface{})

	// A workaround for tests because range does not preserve key order over iteration
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Address > nodes[j].Address
	})

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}

func overrideTopologyPort(nodes []*config.FederationNode, gossipPort int) []*config.FederationNode {
	var newTopology []*config.FederationNode

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		newTopology = append(newTopology, &config.FederationNode{
			Port:    port,
			Address: node.Address,
			IP:      node.IP,
		})
	}

	return newTopology
}

func getVirtualChainConfig(cfg config.NodeConfiguration, chain *config.VirtualChain) *config.VirtualChainConfig {
	return &config.VirtualChainConfig{
		VirtualChain:  chain,
		Topology:      overrideTopologyPort(cfg.FederationNodes(), chain.ExternalPort),
		NodeAddress:   cfg.NodeAddress(),
		KeyPairConfig: getKeyConfigJson(cfg, true),
	}
}
