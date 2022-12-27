package config

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"time"
)

type NodeConfiguration interface {
	OrchestratorOptions() *adapter.OrchestratorOptions
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
	SetOrchestratorOptions(options *adapter.OrchestratorOptions) MutableNodeConfiguration
	SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration
	UpdateDefaultServiceConfig() MutableNodeConfiguration
}

type nodeConfiguration struct {
	OrchestratorOptions *adapter.OrchestratorOptions `json:"orchestrator"`
	Services            Services                     `json:"services"`
}

type nodeConfigurationContainer struct {
	value            nodeConfiguration
	keyConfigPath    string
	ethereumEndpoint string
	sslOptions       adapter.SSLOptions
	withNamespace    bool
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

func (c *nodeConfigurationContainer) OrchestratorOptions() *adapter.OrchestratorOptions {
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

	if c.OrchestratorOptions() == nil {
		return fmt.Errorf("config is missing orchestrator options")
	}

	return nil
}

func (c *nodeConfigurationContainer) EthereumEndpoint() string {
	return c.ethereumEndpoint
}

func (c *nodeConfigurationContainer) SetEthereumEndpoint(ethereumEndpoint string) MutableNodeConfiguration {
	c.ethereumEndpoint = ethereumEndpoint
	return c
}

func (c *nodeConfigurationContainer) SetOrchestratorOptions(options *adapter.OrchestratorOptions) MutableNodeConfiguration {
	c.value.OrchestratorOptions = options
	return c
}

func (c *nodeConfigurationContainer) SetSSLOptions(options adapter.SSLOptions) MutableNodeConfiguration {
	c.sslOptions = options
	return c
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
