package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/log_types"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"sort"
	"strings"
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
	nginxLock sync.Mutex
	strelets  strelets.Strelets
	config    config.NodeConfiguration
	cache     *Cache
	logger    log.Logger
}

type errorContainer struct {
	error error
	id    strelets.VirtualChainId
}

func NewBoyar(strelets strelets.Strelets, cfg config.NodeConfiguration, cache *Cache, logger log.Logger) Boyar {
	return &boyar{
		strelets: strelets,
		config:   cfg,
		cache:    cache,
		logger:   logger,
	}
}

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

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	b.nginxLock.Lock()
	defer b.nginxLock.Unlock()
	// TODO is there a better way to get a loopback interface?
	nginxConfig := getNginxCompositeConfig(b.config)

	if b.cache.nginx.CheckNewJsonValue(nginxConfig) {
		sslEnabled := nginxConfig.SSLOptions.SSLCertificatePath != "" && nginxConfig.SSLOptions.SSLPrivateKeyPath != ""

		config := &adapter.ReverseProxyConfig{
			NginxConfig: getNginxConfig(nginxConfig.Chains, nginxConfig.IP, sslEnabled),
		}

		if sslEnabled {
			if sslCertificate, err := ioutil.ReadFile(nginxConfig.SSLOptions.SSLCertificatePath); err != nil {
				return fmt.Errorf("could not read SSL certificate from %s: %s", nginxConfig.SSLOptions.SSLCertificatePath, err)
			} else {
				config.SSLCertificate = sslCertificate
			}

			if sslPrivateKey, err := ioutil.ReadFile(nginxConfig.SSLOptions.SSLPrivateKeyPath); err != nil {
				return fmt.Errorf("could not read SSL private key from %s: %s", nginxConfig.SSLOptions.SSLCertificatePath, err)
			} else {
				config.SSLPrivateKey = sslPrivateKey
			}
		}

		if err := b.strelets.Orchestrator().RunReverseProxy(ctx, config); err != nil {
			b.logger.Error("failed to apply http proxy configuration", log.Error(err))
			b.cache.nginx.Clear()
			return err
		}

		b.logger.Info("updated http proxy configuration")
	}
	return nil
}

type UpdateReverseProxyInput struct {
	Chains []*strelets.VirtualChain
	IP     string

	SSLOptions adapter.SSLOptions
}

func getNginxCompositeConfig(cfg config.NodeConfiguration) *UpdateReverseProxyInput {
	return &UpdateReverseProxyInput{
		Chains:     cfg.Chains(),
		IP:         helpers.LocalIP(),
		SSLOptions: cfg.SSLOptions(),
	}
}

func (b *boyar) ProvisionServices(ctx context.Context) error {
	if _, err := b.strelets.Orchestrator().GetOverlayNetwork(ctx, adapter.SHARED_SIGNER_NETWORK); err != nil {
		return errors.Wrap(err, "failed creating network")
	}

	var errors []error
	for serviceName, service := range b.config.Services().AsMap() {
		if b.cache.services.CheckNewJsonValue(serviceName, service) {
			if service != nil {
				err := b.strelets.UpdateService(ctx, b.getServiceConfig(serviceName, service))
				if err == nil {
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

var removed = &utils.HashedValue{Value: "foo"}

func (b *boyar) removeVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan *errorContainer) {
	go func() {
		if b.cache.vChains.CheckNewValue(chain.Id.String(), removed) {
			serviceName := adapter.GetServiceId(chain.GetContainerName())
			if err := b.strelets.Orchestrator().ServiceRemove(ctx, serviceName); err != nil {
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

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan *errorContainer) {
	go func() {
		input := getVirtualChainConfig(b.config, chain)

		if b.cache.vChains.CheckNewJsonValue(chain.Id.String(), input) {
			// skip keys
			if err := b.provisionSingleVirtualChain(ctx, b.config.NodeAddress(), chain, b.config.KeyConfig().JSON(false)); err != nil {
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

const (
	PROVISION_VCHAIN_MAX_TRIES       = 5
	PROVISION_VCHAIN_ATTEMPT_TIMEOUT = 30 * time.Second
	PROVISION_VCHAIN_RETRY_INTERVAL  = 3 * time.Second
)

func (s *boyar) provisionSingleVirtualChain(ctx context.Context, nodeAddress config.NodeAddress,
	chain *strelets.VirtualChain, keyPairConfig []byte) error {
	imageName := chain.DockerConfig.FullImageName()

	if chain.Disabled {
		return fmt.Errorf("virtual chain %d is disabled", chain.Id)
	}

	if chain.DockerConfig.Pull {
		if err := s.strelets.Orchestrator().PullImage(ctx, imageName); err != nil {
			return fmt.Errorf("could not pull docker image: %s", err)
		}
	}

	return utils.Try(ctx, PROVISION_VCHAIN_MAX_TRIES, PROVISION_VCHAIN_ATTEMPT_TIMEOUT, PROVISION_VCHAIN_RETRY_INTERVAL,
		func(ctxWithTimeout context.Context) error {
			serviceConfig := &adapter.ServiceConfig{
				Id:            uint32(chain.Id),
				NodeAddress:   string(nodeAddress),
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
				Network: getNetworkConfigJSON(s.config.FederationNodes()),
				Config:  chain.GetSerializedConfig(),
			}

			return s.strelets.Orchestrator().RunVirtualChain(ctx, serviceConfig, appConfig)
		})
}

func (b *boyar) getServiceConfig(serviceName string, service *strelets.Service) *strelets.UpdateServiceInput {
	var keyPairConfigJSON []byte
	if b.config.Services().NeedsKeys(serviceName) {
		keyPairConfigJSON = getKeyConfigJson(b.config, false)
	}

	return &strelets.UpdateServiceInput{
		Name:          serviceName,
		Service:       service,
		KeyPairConfig: keyPairConfigJSON,
	}
}

func getVirtualChainConfig(config config.NodeConfiguration, chain *strelets.VirtualChain) *ProvisionVirtualChainInput {
	peers := buildPeersMap(config.FederationNodes(), chain.GossipPort)

	signerOn := config.Services().SignerOn()
	keyPairConfig := getKeyConfigJson(config, signerOn)

	input := &ProvisionVirtualChainInput{
		VirtualChain:  chain,
		Peers:         peers,
		NodeAddress:   config.NodeAddress(),
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

func getNetworkConfigJSON(nodes []*strelets.FederationNode) []byte {
	jsonMap := make(map[string]interface{})

	// A workaround for tests because range does not preserve key order over iteration
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Address > nodes[j].Address
	})

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}

type ProvisionVirtualChainInput struct {
	VirtualChain *strelets.VirtualChain
	Peers        *PeersMap
	NodeAddress  config.NodeAddress

	KeyPairConfig []byte `json:"-"` // Prevents key leak via log
}

type Peer struct {
	IP   string
	Port int
}

type PeersMap map[config.NodeAddress]*Peer
