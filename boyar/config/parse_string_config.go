package config

import (
	"encoding/json"
)

func parseStringConfig(jsonInput string, ethereumEndpoint string, keyConfigPath string, withNamespace bool) (*nodeConfigurationContainer, error) {
	var value nodeConfiguration
	if err := json.Unmarshal([]byte(jsonInput), &value); err != nil {
		return nil, err
	}

	cfg := &nodeConfigurationContainer{
		value:         value,
		keyConfigPath: keyConfigPath,
		withNamespace: withNamespace,
	}

	cfg.SetEthereumEndpoint(ethereumEndpoint)
	cfg.SetSignerEndpoint()

	return cfg, nil
}
