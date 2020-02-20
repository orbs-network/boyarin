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
	"strings"
	"sync"
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
	nginxConfig := getNginxConfig(b.config)

	if b.cache.nginx.CheckNewJsonValue(nginxConfig) {
		if err := b.strelets.UpdateReverseProxy(ctx, nginxConfig); err != nil {
			b.logger.Error("failed to apply http proxy configuration", log.Error(err))
			b.cache.nginx.Clear()
			return err
		}

		b.logger.Info("updated http proxy configuration")
	}
	return nil
}

func getNginxConfig(cfg config.NodeConfiguration) *strelets.UpdateReverseProxyInput {
	return &strelets.UpdateReverseProxyInput{
		Chains:     cfg.Chains(),
		IP:         helpers.LocalIP(),
		SSLOptions: cfg.SSLOptions(),
	}
}

func (b *boyar) ProvisionServices(ctx context.Context) error {
	if err := b.strelets.ProvisionSharedNetwork(ctx, &strelets.ProvisionSharedNetworkInput{
		Name: adapter.SHARED_SIGNER_NETWORK,
	}); err != nil {
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
			input := &strelets.RemoveVirtualChainInput{
				VirtualChain: chain,
			}
			if err := b.strelets.RemoveVirtualChain(ctx, input); err != nil {
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
			if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
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

func getVirtualChainConfig(config config.NodeConfiguration, chain *strelets.VirtualChain) *strelets.ProvisionVirtualChainInput {
	peers := buildPeersMap(config.FederationNodes(), chain.GossipPort)

	signerOn := config.Services().SignerOn()
	keyPairConfig := getKeyConfigJson(config, signerOn)

	input := &strelets.ProvisionVirtualChainInput{
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
