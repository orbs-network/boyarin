package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/boyar/topology"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"sort"
	"sync"
	"time"
)

type Cache struct {
	vChains  *utils.CacheMap
	nginx    *utils.CacheFilter
	services *utils.CacheMap
}

func NewCache() *Cache {
	return &Cache{
		vChains:  utils.NewCacheMap(),
		nginx:    utils.NewCacheFilter(),
		services: utils.NewCacheMap(),
	}
}

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
	ProvisionServices(ctx context.Context) error
}

type boyar struct {
	nginxLock    sync.Mutex
	orchestrator adapter.Orchestrator
	config       config.NodeConfiguration
	cache        *Cache
	logger       log.Logger
}

type errorContainer struct {
	error error
	id    config.VirtualChainId
}

func NewBoyar(orchestrator adapter.Orchestrator, cfg config.NodeConfiguration, cache *Cache, logger log.Logger) Boyar {
	return &boyar{
		orchestrator: orchestrator,
		config:       cfg,
		cache:        cache,
		logger:       logger,
	}
}

func (b *boyar) ProvisionServices(ctx context.Context) error {
	if _, err := b.orchestrator.GetOverlayNetwork(ctx, adapter.SHARED_SIGNER_NETWORK); err != nil {
		return errors.Wrap(err, "failed creating network")
	}

	var errors []error
	for serviceName, service := range b.config.Services().AsMap() {

		if b.cache.services.CheckNewJsonValue(serviceName, service) {
			if service != nil {
				fullServiceName := service.GetContainerName(serviceName)
				imageName := service.DockerConfig.FullImageName()

				if service.Disabled {
					return fmt.Errorf("signer service is disabled")
				}

				if service.DockerConfig.Pull {
					if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
						return fmt.Errorf("could not pull docker image: %s", err)
					}
				}

				serviceConfig := &adapter.ServiceConfig{
					ImageName:     imageName,
					ContainerName: fullServiceName,

					LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
					LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
					ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
					ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,
				}

				jsonConfig, _ := json.Marshal(service.Config)

				var keyPairConfigJSON []byte
				if b.config.Services().NeedsKeys(fullServiceName) {
					keyPairConfigJSON = getKeyConfigJson(b.config, false)
				}

				appConfig := &adapter.AppConfig{
					KeyPair: keyPairConfigJSON,
					Config:  jsonConfig,
				}

				if err := b.orchestrator.RunService(ctx, serviceConfig, appConfig); err == nil {
					b.logger.Info("updated service configuration", log.Service(serviceName))
				} else {
					b.logger.Error("failed to update service configuration", log.Service(serviceName), log.Error(err))
					b.cache.services.Clear(serviceName)
					errors = append(errors, err)
				}
			}
		}
	}

	return utils.AggregateErrors(errors)
}

const (
	PROVISION_VCHAIN_MAX_TRIES       = 5
	PROVISION_VCHAIN_ATTEMPT_TIMEOUT = 30 * time.Second
	PROVISION_VCHAIN_RETRY_INTERVAL  = 3 * time.Second
)

func (b *boyar) getServiceConfig(serviceName string, service *config.Service) *config.UpdateServiceInput {
	var keyPairConfigJSON []byte
	if b.config.Services().NeedsKeys(serviceName) {
		keyPairConfigJSON = getKeyConfigJson(b.config, false)
	}

	return &config.UpdateServiceInput{
		Name:          serviceName,
		Service:       service,
		KeyPairConfig: keyPairConfigJSON,
	}
}

func getVirtualChainConfig(cfg config.NodeConfiguration, chain *config.VirtualChain) *config.VirtualChainCompositeKey {
	peers := buildPeersMap(cfg.FederationNodes(), chain.GossipPort)

	signerOn := cfg.Services().SignerOn()
	keyPairConfig := getKeyConfigJson(cfg, signerOn)

	input := &config.VirtualChainCompositeKey{
		VirtualChain:  chain,
		Peers:         peers,
		NodeAddress:   cfg.NodeAddress(),
		KeyPairConfig: keyPairConfig,
	}
	return input
}

func getKeyConfigJson(config config.NodeConfiguration, addressOnly bool) []byte {
	keyConfig := config.KeyConfig()
	if keyConfig == nil {
		return []byte{}
	}
	return keyConfig.JSON(addressOnly)
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
