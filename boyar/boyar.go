package boyar

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/strelets"
)

type configurationSource struct {
	value configValue
}

type configValue struct {
	Keys   map[string]string
	Chains []*strelets.VirtualChain `json:"chains"`
}

type ConfigurationSource interface {
	Keys() []byte
	Chains() []*strelets.VirtualChain
}

func NewStringConfigurationSource(input string) (ConfigurationSource, error) {
	return parseStringConfig(input)
}

func parseStringConfig(input string) (*configurationSource, error) {
	var value configValue
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return nil, err
	}

	return &configurationSource{
		value: value,
	}, nil
}

func (c *configurationSource) Keys() []byte {
	keys, _ := json.Marshal(c.value.Keys)
	return keys
}

func (c *configurationSource) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}
