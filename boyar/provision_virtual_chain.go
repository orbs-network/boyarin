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
	"github.com/pkg/errors"
	"sort"
	"strings"
	"time"
)

const (
	PROVISION_VCHAIN_MAX_TRIES       = 5
	PROVISION_VCHAIN_ATTEMPT_TIMEOUT = 30 * time.Second
	PROVISION_VCHAIN_RETRY_INTERVAL  = 3 * time.Second
)

var removed = &utils.HashedValue{Value: "foo"}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	chains := b.config.Chains()

	var errors []error
	errorChannel := make(chan *errorContainer, len(chains))

	for _, chain := range chains {
		if chain.Disabled {
			b.removeVirtualChain(ctx, chain, errorChannel)
		} else {
			b.provisionVirtualChain(ctx, chain, errorChannel)
		}
	}

	var messages []string

	for i := 0; i < len(chains); i++ {
		select {
		case err := <-errorChannel:
			if err != nil {
				errors = append(errors, err.error)
				messages = append(messages, err.id.String())
			}
		case <-ctx.Done():
			errors = append(errors, ctx.Err())
			messages = append(messages, ctx.Err().Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to provision virtual chain %v", strings.Join(messages, ", "))
	}

	return nil
}

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *config.VirtualChain, errChannel chan *errorContainer) {
	go func() {
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

			if chain.Disabled {
				errChannel <- &errorContainer{errors.New("virtual chain is disabled"), chain.Id}
				return
			}

			if chain.DockerConfig.Pull {
				if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
					errChannel <- &errorContainer{fmt.Errorf("could not pull docker image: %b", err), chain.Id}
					return
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
			err := utils.Try(ctx, PROVISION_VCHAIN_MAX_TRIES, PROVISION_VCHAIN_ATTEMPT_TIMEOUT, PROVISION_VCHAIN_RETRY_INTERVAL,
				func(ctxWithTimeout context.Context) error {

					return b.orchestrator.RunVirtualChain(ctx, serviceConfig, appConfig)
				})
			if err != nil {
				b.cache.vChains.Clear(chain.Id.String())
				b.logger.Error("failed to apply virtual chain configuration",
					log_types.VirtualChainId(int64(chain.Id)),
					log.Error(err))
				errChannel <- &errorContainer{err, chain.Id}
			} else {
				input.KeyPairConfig = nil // Prevents key leak via log
				data, _ := json.Marshal(input)
				b.logger.Info("updated virtual chain configuration",
					log_types.VirtualChainId(int64(chain.Id)),
					log.String("configuration", string(data)))
				errChannel <- nil
			}
		} else {
			errChannel <- nil
		}
	}()
}

func (b *boyar) removeVirtualChain(ctx context.Context, chain *config.VirtualChain, errChannel chan *errorContainer) {
	go func() {
		if b.cache.vChains.CheckNewValue(chain.Id.String(), removed) {
			serviceName := adapter.GetServiceId(chain.GetContainerName())
			if err := b.orchestrator.ServiceRemove(ctx, serviceName); err != nil {
				b.cache.vChains.Clear(chain.Id.String())
				b.logger.Error("failed to remove virtual chain",
					log_types.VirtualChainId(int64(chain.Id)),
					log.Error(err))
				errChannel <- &errorContainer{err, chain.Id}
			} else {
				b.logger.Info("removed virtual chain", log_types.VirtualChainId(int64(chain.Id)))
				errChannel <- nil
			}
		} else {
			errChannel <- nil
		}
	}()
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
