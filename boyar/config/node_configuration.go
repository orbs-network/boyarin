package config

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/topology"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"time"
)

type NodeConfiguration interface {
	FederationNodes() []*topology.FederationNode
	Chains() []*VirtualChain
	OrchestratorOptions() adapter.OrchestratorOptions
	KeyConfigPath() string
	KeyConfig() KeyConfig
	ReloadTimeDelay(maxDelay time.Duration) time.Duration
	EthereumEndpoint() string
	NodeAddress() NodeAddress
	SSLOptions() adapter.SSLOptions
	Services() Services

	PrefixedContainerName(name string) string

	VerifyConfig() error
	Hash() string
}

type MutableNodeConfiguration interface {
	NodeConfiguration

	SetFederationNodes(federationNodes []*topology.FederationNode) MutableNodeConfiguration
	SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration
	SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration
	SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration
}

type nodeConfiguration struct {
	Chains              []*VirtualChain             `json:"chains"`
	FederationNodes     []*topology.FederationNode  `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
	Services            Services                    `json:"services"`
}

type nodeConfigurationContainer struct {
	value            nodeConfiguration
	keyConfigPath    string
	ethereumEndpoint string
	sslOptions       adapter.SSLOptions
}

func (c *nodeConfigurationContainer) Chains() []*VirtualChain {
	return c.value.Chains
}

func (c *nodeConfigurationContainer) FederationNodes() []*topology.FederationNode {
	return c.value.FederationNodes
}

func (c *nodeConfigurationContainer) Services() Services {
	return c.value.Services
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

func (c *nodeConfigurationContainer) SSLOptions() adapter.SSLOptions {
	return c.sslOptions
}

func (c *nodeConfigurationContainer) SetFederationNodes(federationNodes []*topology.FederationNode) MutableNodeConfiguration {
	c.value.FederationNodes = federationNodes
	return c
}

// FIXME should add more checks
func (c *nodeConfigurationContainer) VerifyConfig() error {
	_, err := c.readKeysConfig()
	if err != nil {
		return err
	}

	return nil
}

func (n *nodeConfiguration) overrideValues(key string, value string) {
	if value != "" {
		for _, chain := range n.Chains {
			chain.Config[key] = value
		}
	}
}

func (c *nodeConfigurationContainer) EthereumEndpoint() string {
	return c.ethereumEndpoint
}

func (c *nodeConfigurationContainer) SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration {
	c.ethereumEndpoint = ethereumEndpoint
	c.value.overrideValues("ethereum-endpoint", ethereumEndpoint)
	return c
}

func (c *nodeConfigurationContainer) SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration {
	c.value.OrchestratorOptions = options
	return c
}

func (c *nodeConfigurationContainer) SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration {
	c.sslOptions = options
	return c
}

func (c *nodeConfigurationContainer) SetSignerEndpoint() {
	if signer := c.Services().Signer; signer != nil { // FIXME this should become mandatory
		value := fmt.Sprintf("http://%s:%d", adapter.GetServiceId(c.PrefixedContainerName(SIGNER)), signer.Port)
		c.value.overrideValues("signer-endpoint", value)
	}
}

func (c *nodeConfigurationContainer) KeyConfig() KeyConfig {
	cfg, err := c.readKeysConfig()
	if err != nil {
		panic(fmt.Sprintf("key file %s is missing: %s", c.keyConfigPath, err.Error()))
	}
	return cfg
}

func (c *nodeConfigurationContainer) PrefixedContainerName(name string) string {
	return c.NodeAddress().ShortID() + "-" + name
}
