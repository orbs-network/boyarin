package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/config"
	"github.com/orbs-network/boyarin/strelets"
)

type configValue struct {
	Chains          []*strelets.VirtualChain   `json:"chains"`
	FederationNodes []*strelets.FederationNode `json:"network"`
}

type ConfigurationSource interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
}

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
}

type boyar struct {
	strelets          strelets.Strelets
	config            ConfigurationSource
	keyPairConfigPath string
}

func NewBoyar(strelets strelets.Strelets, source ConfigurationSource, keyPairConfigPath string) Boyar {
	return &boyar{
		strelets:          strelets,
		config:            source,
		keyPairConfigPath: keyPairConfigPath,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	for _, chain := range b.config.Chains() {
		peers := config.GetPeersMap(b.config.FederationNodes(), chain.GossipPort)

		if err := b.strelets.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.keyPairConfigPath,
			Peers:             peers,
		}); err != nil {
			return fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err)
		}
	}

	return nil
}
