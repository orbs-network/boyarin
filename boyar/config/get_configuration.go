package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

func GetConfiguration(flags *Flags) (NodeConfiguration, error) {
	config, err := NewUrlConfigurationSource(flags.ConfigUrl, flags.EthereumEndpoint)
	if err != nil {
		return nil, err
	}

	orchestratorOptions, err := getOrchestratorOptions(flags.OrchestratorOptions)
	if err != nil {
		return nil, err
	}

	config.SetOrchestratorOptions(orchestratorOptions)

	endpoint := config.EthereumEndpoint()
	if endpoint != "" && flags.TopologyContractAddress != "" {
		federationNodes, err := ethereum.GetEthereumTopology(context.Background(), endpoint, flags.TopologyContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to retrive topology from Ethereum: %s", err)
		}
		config.SetFederationNodes(federationNodes)
	}

	config.SetKeyConfigPath(flags.KeyPairConfigPath)
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
