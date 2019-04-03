package config

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"time"
)

type NodeConfiguration interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
	OrchestratorOptions() adapter.OrchestratorOptions
	KeyConfigPath() string
	ReloadTimeDelay(maxDelay time.Duration) time.Duration
	EthereumEndpoint() string
	NodeAddress() strelets.NodeAddress

	VerifyConfig() error
	Hash() string
}

type MutableNodeConfiguration interface {
	NodeConfiguration

	SetFederationNodes(federationNodes []*strelets.FederationNode) MutableNodeConfiguration
	SetKeyConfigPath(keyConfigPath string) MutableNodeConfiguration
	SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration
	SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration
}

type nodeConfiguration struct {
	Chains              []*strelets.VirtualChain    `json:"chains"`
	FederationNodes     []*strelets.FederationNode  `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
}

type nodeConfigurationContainer struct {
	value            nodeConfiguration
	keyConfigPath    string
	ethereumEndpoint string
}

func (c *nodeConfigurationContainer) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}

func (c *nodeConfigurationContainer) FederationNodes() []*strelets.FederationNode {
	return c.value.FederationNodes
}

func (c *nodeConfigurationContainer) Hash() string {
	data, _ := json.Marshal(c.value)
	return crypto.CalculateHash(data)
}

func (c *nodeConfigurationContainer) KeyConfigPath() string {
	return c.keyConfigPath
}

func (c *nodeConfigurationContainer) OrchestratorOptions() adapter.OrchestratorOptions {
	return c.value.OrchestratorOptions
}

func (c *nodeConfigurationContainer) SetFederationNodes(federationNodes []*strelets.FederationNode) MutableNodeConfiguration {
	c.value.FederationNodes = federationNodes
	return c
}

func (c *nodeConfigurationContainer) SetKeyConfigPath(keyConfigPath string) MutableNodeConfiguration {
	c.keyConfigPath = keyConfigPath
	return c
}

// FIXME should add more checks
func (n *nodeConfigurationContainer) VerifyConfig() error {
	_, err := n.readKeysConfig()
	if err != nil {
		return err
	}

	return nil
}

func (n *nodeConfiguration) overrideValues(ethereumEndpoint string) {
	if ethereumEndpoint != "" {
		for _, chain := range n.Chains {
			chain.Config["ethereum-endpoint"] = ethereumEndpoint
		}
	}
}

func (c *nodeConfigurationContainer) EthereumEndpoint() string {
	return c.ethereumEndpoint
}

func (c *nodeConfigurationContainer) SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration {
	c.ethereumEndpoint = ethereumEndpoint
	c.value.overrideValues(ethereumEndpoint)
	return c
}

func (c *nodeConfigurationContainer) SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration {
	c.value.OrchestratorOptions = options
	return c
}
