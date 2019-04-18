package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/log_types"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"strings"
)

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets    strelets.Strelets
	config      config.NodeConfiguration
	configCache config.Cache
	logger      log.Logger
}

type errorContainer struct {
	error error
	id    strelets.VirtualChainId
}

func NewBoyar(strelets strelets.Strelets, cfg config.NodeConfiguration, configCache config.Cache, logger log.Logger) Boyar {
	return &boyar{
		strelets:    strelets,
		config:      cfg,
		configCache: configCache,
		logger:      logger,
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
	var keys []config.HttpReverseProxyCompositeKey

	// TODO move key manipulation to config package
	for _, chain := range b.config.Chains() {
		keys = append(keys, config.HttpReverseProxyCompositeKey{
			Id:         chain.Id,
			HttpPort:   chain.HttpPort,
			GossipPort: chain.HttpPort,
			Disabled:   chain.Disabled,
		})
	}

	data, _ := json.Marshal(keys)
	hash := crypto.CalculateHash(data)

	if hash == b.configCache.Get(config.HTTP_REVERSE_PROXY_HASH) {
		return nil
	}

	// TODO is there a better way to get a loopback interface?
	if err := b.strelets.UpdateReverseProxy(ctx, &strelets.UpdateReverseProxyInput{
		Chains:     b.config.Chains(),
		IP:         helpers.LocalIP(),
		SSLOptions: b.config.SSLOptions(),
	}); err != nil {
		b.logger.Error("failed to apply http proxy configuration", log.Error(err))
		return err
	}

	b.logger.Info("updated http proxy configuration")

	b.configCache.Put(config.HTTP_REVERSE_PROXY_HASH, hash)
	return nil
}

func (b *boyar) removeVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan *errorContainer) {
	go func() {
		input := &strelets.RemoveVirtualChainInput{
			VirtualChain: chain,
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache.Get(chain.Id.String()) {
			errChannel <- nil
			return
		}

		if err := b.strelets.RemoveVirtualChain(ctx, input); err != nil {
			b.logger.Error("failed to remove virtual chain",
				log_types.VirtualChainId(int64(chain.Id)),
				log.Error(err))
			errChannel <- &errorContainer{err, chain.Id}
		} else {
			b.configCache.Put(chain.Id.String(), hash)
			b.logger.Info("removed virtual chain", log_types.VirtualChainId(int64(chain.Id)))
			errChannel <- nil
		}
	}()
}

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan *errorContainer) {
	go func() {
		peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

		input := &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.config.KeyConfigPath(),
			Peers:             peers,
			NodeAddress:       b.config.NodeAddress(),
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache.Get(chain.Id.String()) {
			errChannel <- nil
			return
		}

		if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
			b.logger.Error("failed to apply virtual chain configuration",
				log_types.VirtualChainId(int64(chain.Id)),
				log.Error(err))
			errChannel <- &errorContainer{err, chain.Id}
		} else {
			b.configCache.Put(chain.Id.String(), hash)
			b.logger.Info("updated virtual chain configuration",
				log_types.VirtualChainId(int64(chain.Id)),
				log.String("configuration", string(data)))
			errChannel <- nil
		}
	}()
}
