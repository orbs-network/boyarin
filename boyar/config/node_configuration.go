package config

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"time"
)

type NodeConfiguration interface {
	FederationNodes() []*FederationNode
	Chains() []*VirtualChain
	OrchestratorOptions() adapter.OrchestratorOptions
	KeyConfigPath() string
	KeyConfig() KeyConfig
	ReloadTimeDelay(maxDelay time.Duration) time.Duration
	EthereumEndpoint() string
	NodeAddress() NodeAddress
	SSLOptions() adapter.SSLOptions
	Services() Services

	NamespacedContainerName(name string) string

	VerifyConfig() error
	Hash() string
}

type MutableNodeConfiguration interface {
	NodeConfiguration

	SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration
	SetOrchestratorOptions(options adapter.OrchestratorOptions) MutableNodeConfiguration
	SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration
	UpdateDefaultServiceConfig() MutableNodeConfiguration
}

type nodeConfiguration struct {
	Chains              []*VirtualChain             `json:"chains"`
	FederationNodes     []*FederationNode           `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
	Services            Services                    `json:"services"`
}

type nodeConfigurationContainer struct {
	value            nodeConfiguration
	keyConfigPath    string
	ethereumEndpoint string
	sslOptions       adapter.SSLOptions
	withNamespace    bool
}

func (c *nodeConfigurationContainer) Chains() []*VirtualChain {
	return c.value.Chains
}

func (c *nodeConfigurationContainer) FederationNodes() []*FederationNode {
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
	if signer := c.Services().Signer(); signer != nil { // FIXME this should become mandatory
		value := fmt.Sprintf("http://%s:%d", c.NamespacedContainerName(SIGNER), signer.InternalPort)
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

func (c *nodeConfigurationContainer) NamespacedContainerName(name string) string {
	if c.withNamespace {
		return c.NodeAddress().ShortID() + "-" + name
	}

	return name
}

func (c *nodeConfigurationContainer) UpdateDefaultServiceConfig() MutableNodeConfiguration {
	// this is compatibility layer that provides defaults for the signer and management service
	// the management service can produce any configuration it desires and override these values if necessary

	for serviceName, service := range c.Services() {
		switch serviceName {
		case SIGNER:
			service.InjectNodePrivateKey = true
			service.AllowAccessToSigner = true
			service.AllowAccessToServices = false
		default:
			service.InjectNodePrivateKey = false
			service.AllowAccessToServices = true
		}
	}

	return c
}
