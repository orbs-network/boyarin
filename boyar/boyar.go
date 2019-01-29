package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
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

type BoyarConfigCache map[strelets.VirtualChainId]string

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context, configCache BoyarConfigCache) (BoyarConfigCache, error)
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets          strelets.Strelets
	config            NodeConfiguration
	keyPairConfigPath string
}

func NewBoyar(strelets strelets.Strelets, config NodeConfiguration, keyPairConfigPath string) Boyar {
	return &boyar{
		strelets:          strelets,
		config:            config,
		keyPairConfigPath: keyPairConfigPath,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context, configCache BoyarConfigCache) (BoyarConfigCache, error) {
	for _, chain := range b.config.Chains() {
		peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

		input := &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.keyPairConfigPath,
			Peers:             peers,
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == configCache[chain.Id] {
			return configCache, nil
		}

		if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
			return nil, fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err)
		}

		configCache[chain.Id] = hash
	}

	return configCache, nil
}

func RunOnce(keyPairConfigPath string, configUrl string, prevConfigHash string) (configHash string, err error) {
	config, err := NewUrlConfigurationSource(configUrl)
	if err != nil {
		return
	}
	configHash = config.Hash()
	if configHash == prevConfigHash {
		return
	}

	orchestrator, err := adapter.NewDockerSwarm(config.OrchestratorOptions())
	if err != nil {
		return
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := NewBoyar(s, config, keyPairConfigPath)

	cache := make(BoyarConfigCache)
	if _, err = b.ProvisionVirtualChains(context.Background(), cache); err != nil {
		return
	}

	if err = b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return
	}

	return
}

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	// TODO is there a better way to get a loopback interface?
	return b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), helpers.LocalIP())
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
