package config

import (
	"encoding/json"
)

func parseStringConfig(jsonInput string, ethereumEndpoint string) (*nodeConfigurationContainer, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(jsonInput), &value); err != nil {
		return nil, err
	}

	cfg := &nodeConfigurationContainer{
		value: value,
	}
	cfg.SetEthereumEndpoint(ethereumEndpoint)

	return cfg, nil
}
