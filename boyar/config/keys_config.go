package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type keysConfig struct {
	Address    string `json:"node-address"`
	PrivateKey string `json:"node-private-key"`
}

func (n *nodeConfigurationContainer) readKeysConfig() (cfg *keysConfig, err error) {
	data, err := ioutil.ReadFile(n.keyConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys config: %s", err)
	}

	cfg = &keysConfig{}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		return nil, fmt.Errorf("invalid keys config: %s", err)
	}

	if cfg.Address == "" || cfg.PrivateKey == "" {
		return nil, fmt.Errorf("invalid keys in config: one or both values are empty")
	}

	return
}
