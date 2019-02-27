package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"strings"
	"time"
)

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets    strelets.Strelets
	config      config.NodeConfiguration
	configCache config.BoyarConfigCache
}

func NewBoyar(strelets strelets.Strelets, cfg config.NodeConfiguration, configCache config.BoyarConfigCache) Boyar {
	return &boyar{
		strelets:    strelets,
		config:      cfg,
		configCache: configCache,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	chains := b.config.Chains()

	var errors []error
	errorChannel := make(chan error, len(chains))

	for _, chain := range chains {
		b.provisionVirtualChain(ctx, chain, errorChannel)
	}

	for i := 0; i < len(chains); i++ {
		select {
		case err := <-errorChannel:
			if err != nil {
				errors = append(errors, err)

			}
		case <-ctx.Done():
			errors = append(errors, fmt.Errorf("failed to provision virtual chain %s: %s", chains[i].Id, ctx.Err()))
		}
	}

	return aggregateErrors(errors)
}

func RunOnce(ctx context.Context, cfg config.NodeConfiguration, configCache config.BoyarConfigCache) error {
	orchestrator, err := adapter.NewDockerSwarm(cfg.OrchestratorOptions())
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := NewBoyar(s, cfg, configCache)

	var errors []error

	if err := b.ProvisionVirtualChains(ctx); err != nil {
		errors = append(errors, err)
	}

	if err := b.ProvisionHttpAPIEndpoint(ctx); err != nil {
		errors = append(errors, err)
	}

	return aggregateErrors(errors)
}

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	var keys []config.HttpReverseProxyCompositeKey

	// TODO move key manipulation to config package
	for _, chain := range b.config.Chains() {
		keys = append(keys, config.HttpReverseProxyCompositeKey{
			Id:         chain.Id,
			HttpPort:   chain.HttpPort,
			GossipPort: chain.HttpPort,
		})
	}

	data, _ := json.Marshal(keys)
	hash := crypto.CalculateHash(data)

	if hash == b.configCache[config.HTTP_REVERSE_PROXY_HASH] {
		return nil
	}

	// TODO is there a better way to get a loopback interface?
	if err := b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), helpers.LocalIP()); err != nil {
		return err
	}

	b.configCache[config.HTTP_REVERSE_PROXY_HASH] = hash
	return nil
}

func buildPeersMap(nodes []*strelets.FederationNode, gossipPort int) *strelets.PeersMap {
	peersMap := make(strelets.PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[strelets.NodeAddress(node.Address)] = &strelets.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}

func aggregateErrors(errors []error) error {
	if errors == nil {
		return nil
	}

	var lines []string

	for _, err := range errors {
		lines = append(lines, err.Error())
	}

	return fmt.Errorf(strings.Join(lines, "\n"))
}

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan error) {
	go func() {
		peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

		input := &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.config.KeyConfigPath(),
			Peers:             peers,
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache[chain.Id.String()] {
			errChannel <- nil
			return
		}

		if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
			errChannel <- fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err)
		} else {
			b.configCache[chain.Id.String()] = hash
			fmt.Println(time.Now(), fmt.Sprintf("INFO: updated virtual chain %d with configuration %s", chain.Id, hash))
			fmt.Println(string(data))
			errChannel <- nil
		}
	}()
}
