package config

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
)

func GetConfiguration(configUrl string, ethereumEndpoint string, topologyContractAddress string, keyConfigPath string) (NodeConfiguration, error) {
	config, err := NewUrlConfigurationSource(configUrl)
	if err != nil {
		return nil, err
	}

	if ethereumEndpoint != "" && topologyContractAddress != "" {
		federationNodes, err := ethereum.GetEthereumTopology(context.Background(), ethereumEndpoint, topologyContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to retrive topology from Ethereum: %s", err)
		}
		config.SetFederationNodes(federationNodes)
	}

	config.SetKeyConfigPath(keyConfigPath)

	return config, err
}
