package config

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

func GetConfiguration(flags *Flags) (NodeConfiguration, error) {
	config, err := NewUrlConfigurationSource(flags.ConfigUrl, flags.EthereumEndpoint, flags.KeyPairConfigPath, flags.WithNamespace)

	if err != nil {
		return nil, err
	}

	config.SetSSLOptions(getSSLOptions(flags))

	if flags.OrchestratorOptions != "" {
		orchestratorOptions, err := getOrchestratorOptions(flags.OrchestratorOptions)
		if err != nil {
			return nil, err
		}

		config.SetOrchestratorOptions(orchestratorOptions)
	}

	if err := config.VerifyConfig(); err != nil {
		return nil, fmt.Errorf("config verification failed: %s", err)
	}

	return config, err
}

func getOrchestratorOptions(options string) (adapter.OrchestratorOptions, error) {
	orchestratorOptions := adapter.OrchestratorOptions{}
	err := json.Unmarshal([]byte(options), &orchestratorOptions)

	if err != nil {
		return orchestratorOptions, fmt.Errorf("could not parse orchestrator options properly: %s", err)
	}

	return orchestratorOptions, err
}

func getSSLOptions(flags *Flags) adapter.SSLOptions {
	return adapter.SSLOptions{
		SSLCertificatePath: flags.SSLCertificatePath,
		SSLPrivateKeyPath:  flags.SSLPrivateKeyPath,
	}
}
