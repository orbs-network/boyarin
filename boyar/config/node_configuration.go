package config

import (
	"encoding/json"
	"fmt"
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
	KeyConfig() KeyConfig
	ReloadTimeDelay(maxDelay time.Duration) time.Duration
	EthereumEndpoint() string
	NodeAddress() NodeAddress
	SSLOptions() adapter.SSLOptions
	Services() strelets.Services

	VerifyConfig() error
	Hash() string
}

type MutableNodeConfiguration interface {
	NodeConfiguration

	SetFederationNodes(federationNodes []*strelets.FederationNode) MutableNodeConfiguration
	SetKeyConfigPath(keyConfigPath string) MutableNodeConfiguration
	SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration
	SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration
	SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration
}

type nodeConfiguration struct {
	Chains              []*strelets.VirtualChain    `json:"chains"`
	FederationNodes     []*strelets.FederationNode  `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
	Services            strelets.Services           `json:"services"`
}

type nodeConfigurationContainer struct {
	value            nodeConfiguration
	keyConfigPath    string
	ethereumEndpoint string
	sslOptions       adapter.SSLOptions
}

func (c *nodeConfigurationContainer) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}

func (c *nodeConfigurationContainer) FederationNodes() []*strelets.FederationNode {
	return c.value.FederationNodes
}

func (c *nodeConfigurationContainer) Services() strelets.Services {
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
	if c.Services().SignerOn() {
		value := "http://" + c.Services().Signer.SignerInternalEndpoint()
		c.value.overrideValues("signer-endpoint", value)
	}
}

func (n *nodeConfigurationContainer) KeyConfig() KeyConfig {
	cfg, err := n.readKeysConfig()
	if err != nil {
		fmt.Println("error reading KeysConfig", err)
	}
	return cfg
}
