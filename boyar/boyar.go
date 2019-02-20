package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"time"
)

type nodeConfiguration struct {
	Chains              []*strelets.VirtualChain    `json:"chains"`
	FederationNodes     []*strelets.FederationNode  `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
}

type NodeConfiguration interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
	OrchestratorOptions() adapter.OrchestratorOptions
	Hash() string
}

type BoyarConfigCache map[string]string

type httpReverseProxyCompositeKey struct {
	Id         strelets.VirtualChainId
	HttpPort   int
	GossipPort int
}

const HTTP_REVERSE_PROXY_HASH = "HTTP_REVERSE_PROXY_HASH"

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets          strelets.Strelets
	config            NodeConfiguration
	configCache       BoyarConfigCache
	keyPairConfigPath string
}

func NewBoyar(strelets strelets.Strelets, config NodeConfiguration, configCache BoyarConfigCache, keyPairConfigPath string) Boyar {
	return &boyar{
		strelets:          strelets,
		config:            config,
		configCache:       configCache,
		keyPairConfigPath: keyPairConfigPath,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	for _, chain := range b.config.Chains() {
		peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

		input := &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.keyPairConfigPath,
			Peers:             peers,
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache[chain.Id.String()] {
			continue
		}

		if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
			return fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err)
		}

		b.configCache[chain.Id.String()] = hash
		fmt.Println(time.Now(), fmt.Sprintf("INFO: updated virtual chain %d with configuration %s", chain.Id, hash))
		fmt.Println(string(data))
	}

	return nil
}

func RunOnce(keyPairConfigPath string, configUrl string, configCache BoyarConfigCache) error {
	config, err := NewUrlConfigurationSource(configUrl)
	if err != nil {
		return err
	}

	orchestrator, err := adapter.NewDockerSwarm(config.OrchestratorOptions())
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := NewBoyar(s, config, configCache, keyPairConfigPath)

	if err := b.ProvisionVirtualChains(context.Background()); err != nil {
		return err
	}

	if err := b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return err
	}

	return nil
}

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	var keys []httpReverseProxyCompositeKey

	for _, chain := range b.config.Chains() {
		keys = append(keys, httpReverseProxyCompositeKey{
			Id:         chain.Id,
			HttpPort:   chain.HttpPort,
			GossipPort: chain.HttpPort,
		})
	}

	data, _ := json.Marshal(keys)
	hash := crypto.CalculateHash(data)

	if hash == b.configCache[HTTP_REVERSE_PROXY_HASH] {
		return nil
	}

	// TODO is there a better way to get a loopback interface?
	if err := b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), helpers.LocalIP()); err != nil {
		return err
	}

	b.configCache[HTTP_REVERSE_PROXY_HASH] = hash
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

		peersMap[strelets.NodeAddress(node.Key)] = &strelets.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}
