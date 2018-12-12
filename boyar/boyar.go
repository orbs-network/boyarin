package boyar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/test"
)

type nodeConfiguration struct {
	Chains          []*strelets.VirtualChain   `json:"chains"`
	FederationNodes []*strelets.FederationNode `json:"network"`
}

type NodeConfiguration interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
	Hash() string
}

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
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

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	// TODO is there a better way to get a loopback interface?
	return b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), test.LocalIP())
}

func (n *nodeConfiguration) hash() string {
	bytes, _ := json.Marshal(n)
	checksum := sha256.Sum256(bytes)
	return hex.EncodeToString(checksum[:])
}
