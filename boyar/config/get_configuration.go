package config

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
)

func GetConfiguration(configUrl string, ethereumEndpoint string, topologyContractAddress string, keyConfigPath string) (NodeConfiguration, error) {
	config, err := NewUrlConfigurationSource(configUrl, ethereumEndpoint)
	if err != nil {
		return nil, err
	}

	endpoint := config.EthereumEndpoint()
	if endpoint != "" && topologyContractAddress != "" {
		federationNodes, err := ethereum.GetEthereumTopology(context.Background(), endpoint, topologyContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to retrive topology from Ethereum: %s", err)
		}
		config.SetFederationNodes(federationNodes)
	}

	config.SetKeyConfigPath(keyConfigPath)
	if err := config.VerifyConfig(); err != nil {
		return nil, fmt.Errorf("config verification failed: %s", err)
	}

	return config, err
}
