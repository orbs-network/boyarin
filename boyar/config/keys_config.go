package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type KeyConfig interface {
	Address() string
	PrivateKey() string

	JSON(addressOnly bool) []byte
}

type keyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key"`
}

func (n *nodeConfigurationContainer) readKeysConfig() (cfg KeyConfig, err error) {
	return NewKeysConfig(n.keyConfigPath)
}

func NewKeysConfig(path string) (cfg KeyConfig, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys config: %s", err)
	}

	cfg = &keyConfig{}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		return nil, fmt.Errorf("invalid keys config: %s", err)
	}

	if cfg.Address() == "" || cfg.PrivateKey() == "" {
		return nil, fmt.Errorf("invalid keys in config: one or both values are empty")
	}

	return
}

func (k *keyConfig) JSON(addressOnly bool) []byte {
	keys := keyConfig{
		NodeAddress: k.NodeAddress,
	}

	if !addressOnly {
		keys.NodePrivateKey = k.NodePrivateKey
	}

	data, _ := json.Marshal(keys)
	return data
}

func (k *keyConfig) Address() string {
	return k.NodeAddress
}

func (k *keyConfig) PrivateKey() string {
	return k.NodePrivateKey
}
