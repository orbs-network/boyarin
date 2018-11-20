package boyar

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/strelets"
)

type configurationSource struct {
	values []*vchainConfig
}

type ConfigurationSource interface {
}

func NewStringConfigurationSource(input string) (ConfigurationSource, error) {
	return parseStringConfig(input)
}

type vchainConfig struct {
	keys   map[string]string     `json:"keys"`
	vchain strelets.VirtualChain `json:"vchain"`
}

func parseStringConfig(input string) (*configurationSource, error) {
	var cfgs []*vchainConfig
	if err := json.Unmarshal([]byte(input), &cfgs); err != nil {
		return nil, err
	}

	return &configurationSource{
		values: cfgs,
	}, nil
}
